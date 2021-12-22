package provider

import (
	"encoding/json"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"strconv"
)

type InfoBit struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type Container struct {
	Image           string
	Status          Status     `json:"status,omitempty"`
	ResourceRequest []*InfoBit `json:"resource_request,omitempty"`
	ResourceLimit   []*InfoBit `json:"resource_limit,omitempty"`
}

type Status struct {
	Info  []*InfoBit `json:"info,omitempty"`
	State []*InfoBit `json:"state,omitempty"`
}

type PodDrift struct {
	UID             string       `json:"uid,omitempty"`
	Phase           string       `json:"phase,omitempty"`
	Name            string       `json:"name,omitempty"`
	Kind            string       `json:"kind,omitempty"`
	Namespace       string       `json:"namespace,omitempty"`
	ResourceVersion string       `json:"resource_version,omitempty"`
	QOSClass        string       `json:"qos_class,omitempty"`
	NodeName        string       `json:"node_name,omitempty"`
	HostIP          string       `json:"host_ip,omitempty"`
	IP              string       `json:"ip,omitempty"`
	MetaInfo        []*InfoBit   `json:"meta_info,omitempty"`
	Labels          interface{}  `json:"labels,omitempty"`
	Annotations     interface{}  `json:"annotations,omitempty"`
	ContainerInfo   []*Container `json:"container_info,omitempty"`
	Conditions      []*Condition `json:"conditions,omitempty"`
}

type Condition []*InfoBit

type MetaInfo struct {
	Name            string
	Kind            string
	Namespace       string
	ResourceVersion string
}

func NewPodDrift(o *corev1.Pod) *PodDrift {

	drift := &PodDrift{
		UID:             string(o.ObjectMeta.UID),
		Name:            o.ObjectMeta.Name,
		Kind:            o.Kind,
		Namespace:       o.ObjectMeta.Namespace,
		ResourceVersion: o.ObjectMeta.ResourceVersion,
		Labels:          o.ObjectMeta.Labels,
		Annotations:     o.ObjectMeta.Annotations,
		ContainerInfo:   []*Container{},
		Conditions:      []*Condition{},
	}

	//drift.ContainerInfo = make([]*Container, len(o.Spec.Containers))
	drift.ContainerInfo = []*Container{}

	//containers
	for _, c := range o.Spec.Containers {

		container := GetContainerInfo(&c)

		//container status
		for _, cs := range o.Status.ContainerStatuses {

			if c.Image == cs.Image {
				data := container.GetStatus(o.Status)
				if len(data) > 0 {
					container.Status.Info = data
				}
				data = container.GetState(o.Status)
				if len(data) > 0 {
					container.Status.State = data
				}
			}
		}

		drift.ContainerInfo = append(drift.ContainerInfo, container)
	}

	//conditions
	drift.Conditions = GetConditions(&o.Status)
	drift.Phase = string(o.Status.Phase)
	drift.IP = o.Status.PodIP
	drift.QOSClass = string(o.Status.QOSClass)
	drift.NodeName = o.Spec.NodeName
	drift.HostIP = o.Status.HostIP

	return drift
}

func GetConditions(o *corev1.PodStatus) []*Condition {
	cond := []*Condition{}
	cond = append(cond, &Condition{

		{
			Name:  "Status",
			Value: string(o.Conditions[len(o.Conditions)-1].Status),
		},
		{
			Name:  "Type",
			Value: string(o.Conditions[len(o.Conditions)-1].Type),
		},
	})
	return cond
}

func GetContainerInfo(o *corev1.Container) *Container {

	container := &Container{
		Image:           o.Image,
		Status:          Status{},
		ResourceRequest: []*InfoBit{},
		ResourceLimit:   []*InfoBit{},
	}

	for key, value := range o.Resources.Requests {
		container.ResourceRequest = append(container.ResourceRequest, &InfoBit{
			Name:  string(key),
			Value: strconv.FormatFloat(float64(value.Value()), 'f', -1, 64),
		})
	}

	for key, value := range o.Resources.Limits {
		container.ResourceLimit = append(container.ResourceLimit, &InfoBit{
			Name:  string(key),
			Value: strconv.FormatFloat(float64(value.Value()), 'f', -1, 64),
		})
	}

	return container
}

func (c *Container) GetStatus(s corev1.PodStatus) []*InfoBit {
	info := []*InfoBit{}
	for _, cs := range s.ContainerStatuses {
		if cs.Image == c.Image {

			info = append(info, &InfoBit{
				Name:  "Ready",
				Value: strconv.FormatBool(cs.Ready),
			})

			info = append(info, &InfoBit{
				Name:  "RestartCount",
				Value: strconv.Itoa(int(cs.RestartCount)),
			})

			info = append(info, &InfoBit{
				Name:  "Started",
				Value: strconv.FormatBool(*cs.Started),
			})
			break
		}
	}
	return info
}

func (c *Container) GetState(cs corev1.PodStatus) []*InfoBit {
	info := []*InfoBit{}

	for _, cs := range cs.ContainerStatuses {
		if cs.Image == c.Image {

			if cs.State.Running != nil {
				info = append(info, &InfoBit{
					Name:  "Running",
					Value: "true",
				})
			}

			if cs.State.Terminated != nil {
				info = append(info, &InfoBit{
					Name:  "Terminated",
					Value: "true",
				})
			}

			if cs.State.Waiting != nil {
				info = append(info, &InfoBit{
					Name:  "Waiting",
					Value: "true",
				})
			}
		}
	}
	return info
}

func (p *PodDrift) Marshal() string {
	j, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err)
	}
	return string(j)
}

func (p *PodDrift) Unmarshal(j string) {
	err := json.Unmarshal([]byte(j), p)
	if err != nil {
		fmt.Println(err)
	}
}
