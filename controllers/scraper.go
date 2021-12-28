package controllers

import (
	"context"
	"encoding/hex"
	"fmt"
	provider "github.com/hugomatus/kube-drift/api/drift"
	"github.com/pkg/errors"
	"hash/fnv"

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

type MetricLabel map[string]int

var metricLabel = MetricLabel{
	//cpu
	"container_cpu_user_seconds_total":   0,
	"container_cpu_system_seconds_total": 0,
	"container_cpu_usage_seconds_total":  0,
	//memory
	"container_memory_cache":           0,
	"container_memory_swap":            0,
	"container_memory_usage_bytes":     0,
	"container_memory_max_usage_bytes": 0,
	//disk
	"container_fs_io_time_seconds_total":          0,
	"container_fs_io_time_weighted_seconds_total": 0,
	"container_fs_writes_bytes_total":             0,
	"container_fs_reads_bytes_total":              0,
	//network
	"container_network_receive_bytes_total":   0,
	"container_network_receive_errors_total":  0,
	"container_network_transmit_bytes_total":  0,
	"container_network_transmit_errors_total": 0,
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
			//request := client.CoreV1().RESTClient().Get().Resource("nodes").Name(node.Name).SubResource("proxy").Suffix("metrics/resource")
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

		_, err := save(storage, data)

		if err != nil {
			err = errors.Wrap(err, "failed to save scraped metrics")
			klog.Error(err)
		}

	}
}

func save(storage *provider.Store, data map[string][]byte) (string, error) {

	var keyPrefix string
	var cnt, total int
	for nodeName, v := range data {
		//key = utils.GetUniqueKey()

		results, err := provider.DecodeResponse(v)
		if err != nil {
			err = errors.Wrap(err, "failed to decode response")
			klog.Error(err)
			return "", err
		}
		for _, result := range results {
			if _, found := metricLabel[string(result.Metric["__name__"])]; found {
				key := GetUniqueKey()
				d, _ := result.MarshalJSON()
				keyPrefix = fmt.Sprintf("/%s/%s/%s/%s/%s/%v", nodeName, string(result.Metric["namespace"]), string(result.Metric["pod"]), result.Metric["__name__"], result.Metric["container"], key)

				err = storage.DB().Put([]byte(keyPrefix), []byte(d), nil)

				if err != nil {
					err = errors.Wrap(err, "failed to save metrics scrape record")
					klog.Error(err)
				}
				cnt++
			}
		}
		klog.Infof(fmt.Sprintf("SubTotal: Node=%s Metric Sample Records=%v", nodeName, cnt))
		total += cnt
		cnt = 0
	}
	klog.Infof(fmt.Sprintf("Total: Metric Sample Records=%v", total))
	return keyPrefix, nil
}

func GetUniqueKey() string {
	h := fnv.New64a()
	// Hash of Timestamp
	h.Write([]byte(time.Now().String()))
	key := hex.EncodeToString(h.Sum(nil))
	return key
}
