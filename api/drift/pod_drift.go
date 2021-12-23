package provider

import (
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strconv"
)

type DriftMetric map[string]string

type DriftPod struct {
	Name          string
	Namespace     string
	PodInfo       DriftMetric
	PodConditions DriftMetric
	PodContainers []*DriftMetric
	ResourceList  []*DriftMetric
}

func GetPodInfo(p *v1.Pod) *DriftMetric {
	createdBy := metav1.GetControllerOf(p)
	createdByKind := "none"
	createdByName := "none"
	if createdBy != nil {
		if createdBy.Kind != "" {
			createdByKind = createdBy.Kind
		}
		if createdBy.Name != "" {
			createdByName = createdBy.Name
		}
	}

	dm := &DriftMetric{
		"uid":             string(p.UID),
		"name":            p.Name,
		"namespace":       p.Namespace,
		"created_by_kind": createdByKind,
		"created_by_name": createdByName,
		"host_ip":         p.Status.HostIP,
		"pod_ip":          p.Status.PodIP,
		"phase":           string(p.Status.Phase),
		"node_name":       p.Spec.NodeName,
		"priority":        strconv.FormatInt(int64(*p.Spec.Priority), 10),
		"qos_class":       string(p.Status.QOSClass),
	}

	return dm
}

func (d *DriftPod) Marshal() string {
	j, err := json.Marshal(d)
	if err != nil {
		fmt.Println(err)
	}
	return string(j)
}

func GetPodConditions(p *v1.Pod) *DriftMetric {
	dm := &DriftMetric{}

	for _, c := range p.Status.Conditions {
		//TODO: get the last one only
		//if c.Status == v1.ConditionTrue {
		dm = &DriftMetric{
			"condition": string(c.Type),
			"status":    string(c.Status),
		}
	}
	return dm
}

func GetContainerStatus(p *v1.Pod) []*DriftMetric {
	dm := []*DriftMetric{}

	for i, cs := range p.Status.ContainerStatuses {

		state := ""

		if cs.State.Running != nil {
			state = "running"
		} else if cs.State.Waiting != nil {
			state = "waiting"
		} else if cs.State.Terminated != nil {
			state = "terminated"
		} else if cs.State.Terminated != nil {
			state = "unknown"
		}

		if cs.State.Running != nil {
			dm = append(dm, &DriftMetric{
				"container_name":          cs.Name,
				"container_image":         cs.Image,
				"container_image_spec":    p.Spec.Containers[i].Image,
				"container_image_id":      cs.ImageID,
				"container_id":            cs.ContainerID,
				"container_state":         state,
				"container_ready":         strconv.FormatBool(cs.Ready),
				"container_restart_count": strconv.FormatInt(int64(cs.RestartCount), 10),
			})
		}
	}

	return dm
}

func GetContainers(p *v1.Pod) []*DriftMetric {
	dm := []*DriftMetric{}
	for _, cs := range p.Status.ContainerStatuses {
		dm = append(dm, &DriftMetric{
			"container_name":          cs.Name,
			"container_image":         cs.Image,
			"container_image_spec":    cs.ImageID,
			"container_id":            cs.ContainerID,
			"container_state":         string(cs.State.Waiting.Reason),
			"container_ready":         strconv.FormatBool(cs.Ready),
			"container_restart_count": strconv.FormatInt(int64(cs.RestartCount), 10),
		})
	}
	return dm
}

func GetResourceList(p *v1.Pod) []*DriftMetric {
	dm := []*DriftMetric{}

	for _, c := range p.Spec.Containers {
		resourceList := c.Resources.Limits
		for resourceName, val := range resourceList {
			dm = getResources(p, resourceName, dm, c, val)
		}
	}
	return dm
}

func getResources(p *v1.Pod, resourceName v1.ResourceName, dm []*DriftMetric, c v1.Container, val resource.Quantity) []*DriftMetric {
	switch resourceName {
	case v1.ResourceCPU:
		dm = append(dm, &DriftMetric{
			"container_name": c.Name,
			"node_name":      p.Spec.NodeName,
			"resource_name":  string(resourceName),
			"resource_value": val.String(), //float64(val.MilliValue()) / 1000
		})
	case v1.ResourceMemory:
		dm = append(dm, &DriftMetric{
			"container_name": c.Name,
			"node_name":      p.Spec.NodeName,
			"resource_name":  string(resourceName),
			"resource_value": val.String(),
		})
	default:

	}
	return dm
}

func GetResourceRequests(p *v1.Pod) []*DriftMetric {
	dm := []*DriftMetric{}
	for _, c := range p.Spec.Containers {
		requests := c.Resources.Requests
		dm = append(dm, getResourceList(requests)...)
	}
	return dm
}

func getResourceList(resourceList v1.ResourceList) []*DriftMetric {
	var dm []*DriftMetric
	for resourceName, val := range resourceList {
		switch resourceName {
		case v1.ResourceCPU:
			dm = append(dm, &DriftMetric{
				"resource_name":  string(resourceName),
				"resource_value": val.String(), //float64(val.MilliValue()) / 1000
			})
		case v1.ResourceMemory:
			dm = append(dm, &DriftMetric{
				"resource_name":  string(resourceName),
				"resource_value": val.String(),
			})
		default:

		}
	}
	return dm
}
