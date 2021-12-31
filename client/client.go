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

type Client struct {
	Client *kubernetes.Clientset
}

func (c *Client) GetMetrics(node corev1.Node, endpoint string) ([]*model.Sample, error) {

	req := c.Client.CoreV1().RESTClient().Get().Resource("nodes").Name(node.Name).SubResource("proxy").Suffix(endpoint)

	resp, err := req.DoRaw(context.Background())
	if err != nil {
		appLog.Errorf("Error getting metrics: %v", err)
		return nil, err
	}

	resp_, err := c.decodeResponse(resp)
	if err != nil {
		appLog.Errorf("Error decoding response: %v", err)
		return nil, err
	}
	return resp_, nil
}

// DecodeResponse decodes the response from the prometheus samples
func (c *Client) decodeResponse(d []byte) ([]*model.Sample, error) {

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
