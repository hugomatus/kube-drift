package scraper

import (
	"context"
	data "github.com/hugomatus/kube-drift/api/store"
	"github.com/hugomatus/kube-drift/client"
	"github.com/pkg/errors"
	"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appLog "k8s.io/klog/v2"
	"time"
)

type Scraper struct {
	MetricsClient *client.MetricsClient
	Store         *data.Store
	Frequency     time.Duration
	Endpoint      string
}

func (s *Scraper) Run(ctx context.Context) {
	// First scrape run to seed the store
	s.scrape(ctx)

	// Scrape @ every r (metric resolution)
	ticker := time.NewTicker(s.Frequency)
	quit := make(chan struct{})

	for {
		select {
		case <-quit:
			ticker.Stop()
			return

		case <-ticker.C:
			s.scrape(ctx)
		}
	}
}

// Scrape each node in the cluster for stats/summary
func (s *Scraper) scrape(ctx context.Context) {

	ctx, cancelTimeout := context.WithTimeout(ctx, s.Frequency)
	defer cancelTimeout()
	scrapeTimeout := time.Duration(float64(s.Frequency) * 0.90)

	nodeList, err := s.MetricsClient.Clientset.CoreV1().Nodes().List(context.TODO(), v1.ListOptions{})
	nodes := nodeList.Items

	if err != nil {
		err := errors.Wrap(err, "failed to get nodes list to scrape\n")
		appLog.Error(err)
	}

	q := make(chan map[string][]*model.Sample, len(nodes))
	defer close(q)

	for _, n := range nodes {

		go func(node corev1.Node) {
			ctx_, cancelTimeout := context.WithTimeout(ctx, scrapeTimeout)
			defer cancelTimeout()
			resp, err := s.MetricsClient.GetMetrics(ctx_, node)
			if err != nil {
				err = errors.Wrap(err, "failed to scrape metrics")
				appLog.Error(err)
			}

			q <- map[string][]*model.Sample{
				node.Name: resp,
			}
			appLog.Infof("scraped metrics for node %s", node.Name)
		}(n)
	}

	for range nodes {
		data := <-q
		if data == nil {
			continue
		}

		_, err := s.Store.SaveMetrics(data)

		if err != nil {
			err = errors.Wrap(err, "failed to save scraped metrics")
			appLog.Error(err)
		}
	}
}
