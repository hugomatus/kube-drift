package controllers

import (
	"context"
	"fmt"
	provider "github.com/hugomatus/kube-drift/api/drift"
	"github.com/hugomatus/kube-drift/utils"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"time"
)

func ScrapeStats(kubeconfig *string, metricResolution time.Duration, metricDuration time.Duration, storage *provider.Store) {

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		klog.Fatalf("Unable to generate a client config: %s", err)
	}

	klog.Infof("Kubernetes host: %s", config.Host)

	// create k8 clientset
	clientsetCorev1, err := kubernetes.NewForConfig(config)

	if err != nil {
		klog.Fatalf("Unable to generate a clientset: %s", err)
	}

	/*go func() {
		r := mux.NewRouter()
		api.Manager(r, storage)
		// Bind to a port and pass our router in
		klog.Fatal(http.ListenAndServe(":8001", handlers.CombinedLoggingHandler(os.Stdout, r)))
	}()*/

	//Scrape @ every metricResolution
	ticker := time.NewTicker(metricResolution)
	quit := make(chan struct{})

	for {
		select {
		case <-quit:
			ticker.Stop()
			return

		case <-ticker.C:
			klog.Infof("Status: Start Scraping...")
			scrape(clientsetCorev1, storage)
			klog.Infof("Status: Done Scraping...")
		}
	}
}

// scrape each node in the cluster for stats/summary
func scrape(client *kubernetes.Clientset, storage *provider.Store) {

	nodeList, err := client.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	nodes := nodeList.Items

	if err != nil {
		err := errors.Wrap(err, "failed to get nodes list to scrape\n")
		klog.Error(err)
	}

	responseChannel := make(chan map[string][]byte, len(nodes))
	defer close(responseChannel)

	for _, node := range nodes {

		go func(node corev1.Node) {
			klog.Infof("Scraping Node:  %s", node.Name)

			//kubectl get --raw /api/v1/nodes/cygnus/proxy/stats/summary -v 10
			request := client.CoreV1().RESTClient().Get().Resource("nodes").Name(node.Name).SubResource("proxy").Suffix("stats/summary")
			response, err := request.DoRaw(context.Background())

			if err != nil {
				err = errors.Wrap(err, "failed to scrape metrics")
				klog.Error(err)
			}

			responseChannel <- map[string][]byte{
				node.Name: response,
			}
		}(node)
	}

	for range nodes {
		data := <-responseChannel
		if data == nil {
			continue
		}

		key, err := save(storage, data)
		klog.Infof("Saved Sample using key %s", key)

		if err != nil {
			err = errors.Wrap(err, "failed to scrape metrics")
			klog.Error(err)
		}

	}
}

func save(storage *provider.Store, data map[string][]byte) (string, error) {
	key := utils.GetUniqueKey()
	klog.Infof("Generated Unique Key %s", key)

	for nodeName, v := range data {
		err := storage.DB().Put([]byte(fmt.Sprintf("%s-%s", nodeName, key)), []byte(v), nil)

		if err != nil {
			err = errors.Wrap(err, "failed to save metrics scrape record")
			klog.Error(err)
		}
	}
	return key, nil
}
