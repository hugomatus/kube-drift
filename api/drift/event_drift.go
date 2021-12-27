package provider

import (
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"strconv"
	"time"
)

type EventDrift struct {
	Key            string
	EventInfo      DriftMetric
	InvolvedObject DriftMetric
}

type InvolvedObject *DriftMetric

func (r *EventDrift) NewKubeDrift(obj interface{}) interface{} {

	event := obj.(v1.Event)
	info := GetEventInfo(&event)
	o := GetInvolvedObject(&event)

	d := &EventDrift{
		Key:            fmt.Sprintf("/%s/%s/%s/%s", "event", event.Namespace, event.Name, event.UID),
		EventInfo:      *info,
		InvolvedObject: *o,
	}
	return d
}

func (r *EventDrift) GetKey() string {
	if r.Key == "" {
		r.Key = r.EventInfo["key"]
	}
	return r.Key
}

func GetEventInfo(e *v1.Event) *DriftMetric {

	dm := &DriftMetric{}

	dm = &DriftMetric{
		"key":                  fmt.Sprintf("/%s/%s/%s/%s", "event", e.Namespace, e.Name, e.UID),
		"uid":                  string(e.UID),
		"kind":                 e.Kind,
		"name":                 e.Name,
		"namespace":            e.Namespace,
		"reason":               e.Reason,
		"message":              e.Message,
		"count":                strconv.FormatInt(int64(e.Count), 10),
		"type":                 e.Type,
		"action":               e.Action,
		"event_time":           e.EventTime.Format(time.RFC3339),
		"creation_timestamp":   e.CreationTimestamp.Format(time.RFC3339),
		"reporting_controller": e.ReportingController,
		"reporting_instance":   e.ReportingInstance,
		//		"series_count":         strconv.FormatInt(int64(e.Series.Count), 10),
		//"series_last_observed": e.Series.LastObservedTime.Format(time.RFC3339),
	}
	return dm
}

func GetInvolvedObject(e *v1.Event) *DriftMetric {

	o := &DriftMetric{
		"involved_object":                  e.InvolvedObject.Kind,
		"involved_object_name":             e.InvolvedObject.Name,
		"involved_object_namespace":        e.InvolvedObject.Namespace,
		"involved_object_uid":              string(e.InvolvedObject.UID),
		"involved_object_resource_version": e.InvolvedObject.ResourceVersion,
	}
	return o
}

func (d *EventDrift) Marshal() []byte {
	j, err := json.Marshal(d)
	if err != nil {
		fmt.Println(err)
	}
	return j
}
