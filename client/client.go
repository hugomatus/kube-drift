package client

import (
	"bytes"
	"context"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/model"
	"io"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	appLog "k8s.io/klog/v2"
	"strings"
)

type Client interface {
	Init()
	GetMetrics(node corev1.Node, endpoint string) ([]*model.Sample, error)
	DecodeResponse(d []byte) ([]*model.Sample, error)
}

type MetricsClient struct {
	config    *Config
	Clientset *kubernetes.Clientset
	Endpoint  string
}

func (c *MetricsClient) Init(inCluster bool, endpoint string) {
	c.Endpoint = endpoint
	c.config = &Config{}
	c.config.Init(inCluster)
	c.Clientset = c.config.Client
}

func (c *MetricsClient) GetMetrics(ctx context.Context, node corev1.Node) ([]*model.Sample, error) {

	req := c.Clientset.CoreV1().RESTClient().Get().Resource("nodes").Name(node.Name).SubResource("proxy").Suffix(c.Endpoint)

	resp, err := req.DoRaw(ctx)
	if err != nil {
		appLog.Errorf("Error getting metrics: %v", err)
		return nil, err
	}

	resp_, err := c.DecodeResponse(resp)
	if err != nil {
		appLog.Errorf("Error decoding response: %v", err)
		return nil, err
	}
	return resp_, nil
}

// DecodeResponse decodes the response from the prometheus samples
func (c *MetricsClient) DecodeResponse(d []byte) ([]*model.Sample, error) {

	ioReaderData := strings.NewReader(string(d))
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(ioReaderData)

	if err != nil {
		return nil, err
	}
	dec := expfmt.NewDecoder(buf, expfmt.FmtText)
	decoder := expfmt.SampleDecoder{
		Dec:  dec,
		Opts: &expfmt.DecodeOptions{},
	}

	var samples []*model.Sample
	for {
		var v model.Vector
		if err := decoder.Decode(&v); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		samples = append(samples, v...)
	}
	return samples, nil
}
