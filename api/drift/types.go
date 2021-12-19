package provider

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"time"
)

type KubeDrift struct {
	Key             string            `json:"key"`
	UUID            string            `json:"uuid"`
	ResourceVersion string            `json:"resourceVersion"`
	Name            string            `json:"name"`
	GenerationName  string            `json:"generationName"`
	Namespace       string            `json:"namespace"`
	Kind            string            `json:"kind"`
	CreationTime    time.Time         `json:"creationTime"`
	Labels          map[string]string `json:"labels"`
	Annotations     map[string]string `json:"annotations"`
	Status          interface{}       `json:"status"`
	drift           interface{}
}

func (p *KubeDrift) GetKey() string {
	return p.UUID
}

func (p *KubeDrift) New(drift interface{}) {

	switch v := drift.(type) {
	case corev1.Pod:
		o := (drift).(corev1.Pod)
		p.UUID = string(o.ObjectMeta.UID)
		p.ResourceVersion = o.ObjectMeta.ResourceVersion
		p.Name = o.ObjectMeta.Name
		p.GenerationName = o.ObjectMeta.GenerateName
		p.Namespace = o.ObjectMeta.Namespace
		p.Kind = o.Kind
		p.Labels = o.ObjectMeta.Labels
		p.Annotations = o.ObjectMeta.Annotations
		p.Status = o.Status
		//p.Spec = drift.Spec

		p.drift = drift
	case corev1.Node:
		o := (drift).(corev1.Node)
		p.UUID = string(o.ObjectMeta.UID)
		p.ResourceVersion = o.ObjectMeta.ResourceVersion
		p.Name = o.ObjectMeta.Name
		p.GenerationName = o.ObjectMeta.GenerateName
		p.Namespace = o.ObjectMeta.Namespace
		p.Kind = o.Kind
		p.Labels = o.ObjectMeta.Labels
		p.Annotations = o.ObjectMeta.Annotations
		p.Status = o.Status
		//p.Spec = drift.Spec
		p.drift = drift
	case appsv1.Deployment:
		o := (drift).(appsv1.Deployment)
		p.UUID = string(o.ObjectMeta.UID)
		p.ResourceVersion = o.ObjectMeta.ResourceVersion
		p.Name = o.ObjectMeta.Name
		p.GenerationName = o.ObjectMeta.GenerateName
		p.Namespace = o.ObjectMeta.Namespace
		p.Kind = o.Kind
		p.Labels = o.ObjectMeta.Labels
		p.Annotations = o.ObjectMeta.Annotations
		p.Status = o.Status
		//p.Spec = drift.Spec
		p.drift = drift
	default:
		fmt.Printf("I don't know about type %T!\n", v)
	}

}

func (p *KubeDrift) GetDrift() interface{} {
	return p.drift
}

func (p *KubeDrift) Marshal() string {
	j, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err)
	}
	return string(j)
}

func Marshal(e corev1.Event) string {
	j, err := json.Marshal(e)
	if err != nil {
		fmt.Println(err)
	}
	return string(j)
}

func (p *KubeDrift) serialize(obj *KubeDrift) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(obj)
	if err != nil {
		fmt.Printf("error encoding object: %v", err)
	}
	return buf.Bytes()
}

func (p *KubeDrift) deserialize(data []byte) KubeDrift {
	obj := KubeDrift{}
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&obj)
	if err != nil {
		fmt.Printf("error decoding object: %v", err)
	}
	return obj
}
