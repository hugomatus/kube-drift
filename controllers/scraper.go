package controllers

import (
	"context"
	"fmt"
	provider "github.com/hugomatus/kube-drift/api/drift"
	"github.com/hugomatus/kube-drift/utils"
	"github.com/pkg/errors"
	//"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
	"sync"
	"time"
)

type scraper struct {
	nodeLister    v1listers.NodeLister
	scrapeTimeout time.Duration
	buffers       sync.Pool
}

func ScrapeStats(clientsetCorev1 *kubernetes.Clientset, metricResolution time.Duration, storage *provider.Store) {

	//Scrape @ every metricResolution
	ticker := time.NewTicker(metricResolution)
	quit := make(chan struct{})

	for {
		select {
		case <-quit:
			ticker.Stop()
			return

		case <-ticker.C:
			scrape(clientsetCorev1, storage)
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

			//kubectl get --raw /api/v1/nodes/cygnus/proxy/stats/summary -v 10
			//request := client.CoreV1().RESTClient().Get().Resource("nodes").Name(node.Name).SubResource("proxy").Suffix("stats/summary")
			request := client.CoreV1().RESTClient().Get().Resource("nodes").Name(node.Name).SubResource("proxy").Suffix("metrics/cadvisor")
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
		klog.Infof("Metric Sample: saved using key %s", key)

		if err != nil {
			err = errors.Wrap(err, "failed to scrape metrics")
			klog.Error(err)
		}

	}
}

func save(storage *provider.Store, data map[string][]byte) (string, error) {
	var key string

	for nodeName, v := range data {
		key = utils.GetUniqueKey()
		results, err := provider.DecodeResponse(v)
		if err != nil {
			err = errors.Wrap(err, "failed to decode response")
			klog.Error(err)
			return "", err
		}
		for _, result := range results {

			d, _ := result.MarshalJSON()

			keyPrefix := fmt.Sprintf("%s-%s-%s-%s-%s", nodeName, result.Metric["__name__"], string(result.Metric["namespace"]), string(result.Metric["pod"]), key)
			klog.Infof(fmt.Sprintf("Saving Metric Samples:%s", keyPrefix))
			err = storage.DB().Put([]byte(keyPrefix), []byte(d), nil)

			if err != nil {
				err = errors.Wrap(err, "failed to save metrics scrape record")
				klog.Error(err)
			}
		}

	}
	return key, nil
}
