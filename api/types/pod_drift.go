package types

import (
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appLog "k8s.io/klog/v2"
	"strconv"
)

type KubeDrift interface {
	GetKey() string
	NewKubeDrift(obj interface{}) *KubeDrift
}

type DriftMetric map[string]string

func (r *PodDrift) GetKey() string {
	if r.Key == "" {
		r.Key = r.PodInfo["key"]
	}
	return r.Key
}

type PodDrift struct {
	Key                 string         `json:"key"`
	PodInfo             DriftMetric    `json:"pod_info"`
	PodLabels           DriftMetric    `json:"pod_labels"`
	PodAnnotations      DriftMetric    `json:"pod_annotations"`
	PodConditions       DriftMetric    `json:"pod_conditions"`
	PodContainers       []*DriftMetric `json:"pod_containers"`
	PodResourceRequests []*DriftMetric `json:"resource_requests"`
	PodResourceLimits   []*DriftMetric `json:"resource_limits"`
	PodVolumes          []*DriftMetric `json:"pod_volumes"`
}

func (r *PodDrift) NewKubeDrift(obj interface{}) interface{} {
	pod := obj.(v1.Pod)
	info := GetPodInfo(&pod)
	cond := GetPodConditions(&pod)
	status := GetContainerStatus(&pod)
	resourceRequest := GetResourceRequests(&pod)
	resourceLimit := GetResourceLimits(&pod)
	labels := GetPodLabels(&pod)
	annotations := GetPodAnnotations(&pod)
	vols := GetPodVolumes(&pod)

	d := PodDrift{
		Key:                 fmt.Sprintf("/%s/%s/%s/%s", "pod", pod.Namespace, pod.Name, pod.UID),
		PodInfo:             *info,
		PodConditions:       *cond,
		PodContainers:       status,
		PodResourceRequests: resourceRequest,
		PodResourceLimits:   resourceLimit,
		PodLabels:           *labels,
		PodAnnotations:      *annotations,
		PodVolumes:          vols,
	}
	return d
}

// GetPodVolumes returns the PersistentVolumeClaim(s) of the pod
func GetPodVolumes(p *v1.Pod) []*DriftMetric {
	for _, v := range p.Spec.Volumes {
		if v.VolumeSource.PersistentVolumeClaim != nil {
			return []*DriftMetric{
				{
					"vol_name":       v.Name,
					"vol_claim_name": v.VolumeSource.PersistentVolumeClaim.ClaimName,
				},
			}
		}
	}
	return []*DriftMetric{}
}

// GetPodInfo returns the pod info
func GetPodInfo(p *v1.Pod) *DriftMetric {
	dm := &DriftMetric{}
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

	dm = &DriftMetric{
		"key":             fmt.Sprintf("/%s/%s/%s/%s", "pod", p.Namespace, p.Name, p.UID),
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

// GetPodLabels returns the pod labels
func GetPodLabels(p *v1.Pod) *DriftMetric {
	dm := DriftMetric{}
	for k, v := range p.Labels {
		dm[k] = v
	}
	return &dm
}

// GetPodAnnotations returns the pod annotations
func GetPodAnnotations(p *v1.Pod) *DriftMetric {
	dm := DriftMetric{}
	for k, v := range p.Annotations {
		dm[k] = v
	}
	return &dm
}

// GetPodConditions returns the pod conditions
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

// GetContainerStatus returns the container status
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

// GetContainers returns the pod containers
func GetContainers(p *v1.Pod) []*DriftMetric {
	dm := []*DriftMetric{}
	for _, cs := range p.Status.ContainerStatuses {
		dm = append(dm, &DriftMetric{
			"container_name":          cs.Name,
			"container_image":         cs.Image,
			"container_image_spec":    cs.ImageID,
			"container_id":            cs.ContainerID,
			"container_state":         cs.State.Waiting.Reason,
			"container_ready":         strconv.FormatBool(cs.Ready),
			"container_restart_count": strconv.FormatInt(int64(cs.RestartCount), 10),
		})
	}
	return dm
}

// GetResourceLimits returns the pod resource list
func GetResourceLimits(p *v1.Pod) []*DriftMetric {
	dm := []*DriftMetric{}
	for _, c := range p.Spec.Containers {
		resourceList := c.Resources.Limits
		for resourceName, val := range resourceList {
			dm = getResources(p, resourceName, dm, c, val)
		}
	}
	return dm
}

func GetResourceRequests(p *v1.Pod) []*DriftMetric {
	dm := []*DriftMetric{}
	for _, c := range p.Spec.Containers {
		resourceList := c.Resources.Requests
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

// Marshal PodDrift to json
func (d *PodDrift) Marshal() []byte {
	j, err := json.Marshal(d)
	if err != nil {
		appLog.Error(err)
	}
	return j
}
