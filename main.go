/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/hugomatus/kube-drift/api"
	store "github.com/hugomatus/kube-drift/api/store"
	"github.com/hugomatus/kube-drift/scraper"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"k8s.io/component-base/logs"
	appLog "k8s.io/klog/v2"
	"net/http"
	"os"
	"path/filepath"
	"time"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/hugomatus/kube-drift/controllers"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func main() {
	logs.InitLogs()
	defer logs.FlushLogs()

	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var metricResolution time.Duration
	var dbStoragePath string

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")

	flag.DurationVar(&metricResolution, "metric-resolution", 60*time.Minute, "The resolution at which metrics-scraper will poll metrics.")
	flag.StringVar(&dbStoragePath, "db-storage-path", "/tmp/kube-drift", "What path to use for storage.")

	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	//create datastore
	s := &store.Store{}
	s.Init(dbStoragePath)
	defer s.Close()

	// API Server
	go func(s *store.Store) {
		setupLog.Info("Run API Server::ListenAndServe on port 8001")
		r := mux.NewRouter()
		api.Manager(r, s)
		// Bind to a port and pass our router in
		err := http.ListenAndServe(":8001", handlers.CombinedLoggingHandler(os.Stdout, r))

		if err != nil {
			setupLog.Error(err, "Error starting server")
		}
	}(s)

	// Metrics Scraper
	go func(store *store.Store) {
		setupLog.Info("Run Scraper::ListenAndServe on port 8001")

		z := scraper.Scraper{
			Client:    getKubernetesClient(),
			Store:     s,
			Frequency: metricResolution,
			Endpoint:  "metrics/cadvisor",
		}

		// Run the scraper and scrape @ metricResolution
		z.Run()
	}(s)

	// Kubernetes Operator setup
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "7aa6c727.kubedrift.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder
	if err = (&controllers.PodReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr, s); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Pod")
		os.Exit(1)
	}
	if err = (&controllers.EventReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr, s); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Event")
		os.Exit(1)
	}
	if err = (&controllers.NodeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr, s); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Node")
		os.Exit(1)
	}

	if err = (&controllers.DeploymentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr, s); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Deployment")
		os.Exit(1)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}

}

func getKubernetesClient() *kubernetes.Clientset {

	var c string
	if home := homedir.HomeDir(); home != "" {
		c = filepath.Join(home, ".kube", "config")
	} else {
		c = ""
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", c)
	if err != nil {
		appLog.Fatalf("Unable to generate a client cfg: %s", err)
	}
	appLog.Infof("Kubernetes host: %s", cfg.Host)

	// create k8 client
	client, err := kubernetes.NewForConfig(cfg)

	if err != nil {
		appLog.Fatalf("Unable to generate a client: %s", err)
	}

	return client
}
