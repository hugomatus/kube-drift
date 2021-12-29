package scraper

import (
	"context"
	data "github.com/hugomatus/kube-drift/api/store"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appLog "k8s.io/klog/v2"
	"time"
)

func Start(c *kubernetes.Clientset, r time.Duration, s *data.Store) {

	//Scrape @ every r (metric resolution)
	ticker := time.NewTicker(r)
	quit := make(chan struct{})

	for {
		select {
		case <-quit:
			ticker.Stop()
			return

		case <-ticker.C:
			Scrape(c, s)
		}
	}
}

// scrape each node in the cluster for stats/summary
func Scrape(c *kubernetes.Clientset, s *data.Store) {

	nodeList, err := c.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	nodes := nodeList.Items

	if err != nil {
		err := errors.Wrap(err, "failed to get nodes list to scrape\n")
		appLog.Error(err)
	}

	q := make(chan map[string][]byte, len(nodes))
	defer close(q)

	for _, n := range nodes {

		go func(node corev1.Node) {

			req := c.CoreV1().RESTClient().Get().Resource("nodes").Name(node.Name).SubResource("proxy").Suffix("metrics/cadvisor")
			resp, err := req.DoRaw(context.Background())

			if err != nil {
				err = errors.Wrap(err, "failed to scrape metrics")
				appLog.Error(err)
			}

			q <- map[string][]byte{
				node.Name: resp,
			}
		}(n)
	}

	for range nodes {
		d := <-q
		if d == nil {
			continue
		}

		_, err := s.SaveMetricSamples(d)

		if err != nil {
			err = errors.Wrap(err, "failed to save scraped metrics")
			appLog.Error(err)
		}
	}
}
