package client

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	appLog "k8s.io/klog/v2"
	"path/filepath"
)

type Config struct {
	KubeConfig string
	Client     *kubernetes.Clientset
}

// Init initializes the Config
func (c *Config) Init(inCluster bool) error {

	if inCluster {
		c.KubeConfig = ""
	} else {
		if home := homedir.HomeDir(); home != "" {
			c.KubeConfig = filepath.Join(home, ".kube", "config")
		}
	}

	c.setClient()

	return nil
}

// getClient initializes the kubernetes client
func (c *Config) setClient() {
	cfg, err := clientcmd.BuildConfigFromFlags("", c.KubeConfig)
	if err != nil {
		appLog.Fatalf("Unable to generate a client cfg: %s", err)
	}

	appLog.Infof("Kubernetes host: %s", cfg.Host)

	// create k8 client
	c.Client, err = kubernetes.NewForConfig(cfg)

	if err != nil {
		appLog.Fatalf("Unable to generate a client: %s", err)
	}
}
