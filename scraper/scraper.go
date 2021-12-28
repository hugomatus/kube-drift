package scraper

import (
	"context"
	"encoding/hex"
	"fmt"
	data "github.com/hugomatus/kube-drift/api/store"
	"github.com/pkg/errors"
	"hash/fnv"

	//"github.com/prometheus/common/model"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
	appLog "k8s.io/klog/v2"
	"sync"
	"time"
)

type scraper struct {
	nodeLister    v1listers.NodeLister
	scrapeTimeout time.Duration
	buffers       sync.Pool
}

type MetricLabels map[string]int

var metricLabel = MetricLabels{
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
			scrape(c, s)
		}
	}
}

// scrape each node in the cluster for stats/summary
func scrape(c *kubernetes.Clientset, s *data.Store) {

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

		_, err := save(s, d)

		if err != nil {
			err = errors.Wrap(err, "failed to save scraped metrics")
			appLog.Error(err)
		}

	}
}

func save(s *data.Store, d map[string][]byte) (string, error) {

	var prefix string
	var cnt int

	//for each node key: nodeName, value: []byte
	for n, v := range d {
		resp, err := data.DecodeResponse(v)
		if err != nil {
			err = errors.Wrap(err, "failed to decode response")
			appLog.Error(err)
			return "", err
		}
		for _, sample := range resp {
			if _, found := metricLabel[string(sample.Metric["__name__"])]; found {
				key := GetUniqueKey()
				d, _ := sample.MarshalJSON()
				prefix = fmt.Sprintf("/%s/%s/%s/%s/%s/%v", n, string(sample.Metric["namespace"]), string(sample.Metric["pod"]), sample.Metric["__name__"], sample.Metric["container"], key)

				err = s.DB().Put([]byte(prefix), []byte(d), nil)
				if err != nil {
					err = errors.Wrap(err, "failed to save metrics scrape record")
					appLog.Error(err)
				}
				cnt++
			}
		}
	}
	appLog.Infof(fmt.Sprintf("Total: Metric Sample Records=%v", cnt))
	return prefix, nil
}

func GetUniqueKey() string {
	h := fnv.New64a()
	// Hash of Timestamp
	h.Write([]byte(time.Now().String()))
	key := hex.EncodeToString(h.Sum(nil))
	return key
}
