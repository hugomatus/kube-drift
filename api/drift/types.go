package provider

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	"time"
)

type Event struct {
	InvolvedObject      v1.ObjectReference  `json:"involvedObject" protobuf:"bytes,2,opt,name=involvedObject"`
	Reason              string              `json:"reason,omitempty" protobuf:"bytes,3,opt,name=reason"`
	Message             string              `json:"message,omitempty" protobuf:"bytes,4,opt,name=message"`
	Source              v1.EventSource      `json:"source,omitempty" protobuf:"bytes,5,opt,name=source"`
	FirstTimestamp      metav1.Time         `json:"firstTimestamp,omitempty" protobuf:"bytes,6,opt,name=firstTimestamp"`
	LastTimestamp       metav1.Time         `json:"lastTimestamp,omitempty" protobuf:"bytes,7,opt,name=lastTimestamp"`
	Count               int32               `json:"count,omitempty" protobuf:"varint,8,opt,name=count"`
	Type                string              `json:"type,omitempty" protobuf:"bytes,9,opt,name=type"`
	EventTime           metav1.MicroTime    `json:"eventTime,omitempty" protobuf:"bytes,10,opt,name=eventTime"`
	Series              *v1.EventSeries     `json:"series,omitempty" protobuf:"bytes,11,opt,name=series"`
	Action              string              `json:"action,omitempty" protobuf:"bytes,12,opt,name=action"`
	Related             *v1.ObjectReference `json:"related,omitempty" protobuf:"bytes,13,opt,name=related"`
	ReportingController string              `json:"reportingComponent" protobuf:"bytes,14,opt,name=reportingComponent"`
	ReportingInstance   string              `json:"reportingInstance" protobuf:"bytes,15,opt,name=reportingInstance"`
}

type MetaData struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Kind              string            `json:"kind"`
	UUID              string            `json:"uuid"`
	ResourceVersion   string            `json:"resourceVersion"`
	CreationTimestamp metav1.Time       `json:"creationTimestamp"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
	GenerationName    string            `json:"generationName"`
	CreationTime      time.Time         `json:"creationTime"`
	CreatedByKind     string            `json:"createdByKind"`
	CreatedByName     string            `json:"createdByName"`
}

type ObjectMeta struct {
	Name                       string                  `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`
	GenerateName               string                  `json:"generateName,omitempty" protobuf:"bytes,2,opt,name=generateName"`
	Namespace                  string                  `json:"namespace,omitempty" protobuf:"bytes,3,opt,name=namespace"`
	UID                        types.UID               `json:"uid,omitempty" protobuf:"bytes,5,opt,name=uid,casttype=k8s.io/kubernetes/pkg/types.UID"`
	ResourceVersion            string                  `json:"resourceVersion,omitempty" protobuf:"bytes,6,opt,name=resourceVersion"`
	Generation                 int64                   `json:"generation,omitempty" protobuf:"varint,7,opt,name=generation"`
	CreationTimestamp          metav1.Time             `json:"creationTimestamp,omitempty" protobuf:"bytes,8,opt,name=creationTimestamp"`
	DeletionTimestamp          *metav1.Time            `json:"deletionTimestamp,omitempty" protobuf:"bytes,9,opt,name=deletionTimestamp"`
	DeletionGracePeriodSeconds *int64                  `json:"deletionGracePeriodSeconds,omitempty" protobuf:"varint,10,opt,name=deletionGracePeriodSeconds"`
	Labels                     map[string]string       `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`
	Annotations                map[string]string       `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`
	OwnerReferences            []metav1.OwnerReference `json:"ownerReferences,omitempty" patchStrategy:"merge" patchMergeKey:"uid" protobuf:"bytes,13,rep,name=ownerReferences"`
	Finalizers                 []string                `json:"finalizers,omitempty" patchStrategy:"merge" protobuf:"bytes,14,rep,name=finalizers"`
	ClusterName                string                  `json:"clusterName,omitempty" protobuf:"bytes,15,opt,name=clusterName"`
}

/*type DriftMetric struct {
	LabelKeys   []string
	LabelValues []string
	Value       float64
}

func (p *DriftMetric) Marshal() string {
	j, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err)
	}
	return string(j)
}*/

type Drift struct {
	key  string
	Type string `json:"type"`
	//EventType string      `json:"eventType"`
	MetaData ObjectMeta  `json:"metaData"`
	Status   interface{} `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
	//ResourceRequirement      []v1.ResourceRequirements `json:"resourceRequest,omitempty" protobuf:"bytes,4,opt,name=resourceRequest"`
	//SpecContainers           []v1.Container            `json:"specContainers,omitempty" protobuf:"bytes,5,opt,name=specContainers"`
	Event interface{} `json:"event,omitempty" protobuf:"bytes,3,opt,name=event"`
	/*ContainerInfo            []*Container `json:"containerInfo,omitempty" protobuf:"bytes,3,opt,name=event"`
	ContainerStatus          []*InfoBit   `json:"containerStatus,omitempty" protobuf:"bytes,3,opt,name=event"`
	ContainerResourceRequest []*InfoBit   `json:"containerResources,omitempty" protobuf:"bytes,3,opt,name=event"`*/
}

func (p *Drift) SetKey() {

	if p.MetaData.Namespace == "" {
		p.MetaData.Namespace = "none"
	}

	key := fmt.Sprintf("/%s/%s/%s/%s", p.Type, p.MetaData.Namespace, p.MetaData.Name, p.MetaData.UID)
	p.key = key
}

func (p *Drift) GetKey() string {
	if p.key == "" {
		p.SetKey()
	}
	return p.key
}

func New(drift interface{}, eventType string) *Drift {
	p := &Drift{}
	switch v := drift.(type) {
	case v1.Pod:
		klog.Infof("Processing type %T!\n", v)
		o := (drift).(v1.Pod)
		p.newPod(eventType, o)
	case v1.Node:
		klog.Infof("Processing type %T!\n", v)
		o := (drift).(v1.Node)
		p.newNode(eventType, o)
	case v1.Event:
		klog.Infof("Processing type %T!\n", v)
		o := (drift).(v1.Event)
		p.newEvent(eventType, o)
	case appsv1.Deployment:
		klog.Infof("Processing type %T!\n", v)
		o := (drift).(appsv1.Deployment)
		p.newDeployment(eventType, o)
	default:
		klog.Infof("I don't know about type %T!\n", v)
	}
	return p
}

func (p *Drift) newDeployment(eventType string, o appsv1.Deployment) {
	p.Type = "deployment"
	//p.EventType = eventType
	p.MetaData = ObjectMeta{
		Name:                       o.ObjectMeta.Name,
		GenerateName:               o.ObjectMeta.GenerateName,
		Namespace:                  o.ObjectMeta.Namespace,
		UID:                        o.ObjectMeta.UID,
		ResourceVersion:            o.ObjectMeta.ResourceVersion,
		Generation:                 o.ObjectMeta.Generation,
		CreationTimestamp:          o.ObjectMeta.CreationTimestamp,
		DeletionTimestamp:          o.ObjectMeta.DeletionTimestamp,
		DeletionGracePeriodSeconds: o.ObjectMeta.DeletionGracePeriodSeconds,
		Labels:                     o.ObjectMeta.Labels,
		Annotations:                o.ObjectMeta.Annotations,
		OwnerReferences:            o.ObjectMeta.OwnerReferences,
		Finalizers:                 o.ObjectMeta.Finalizers,
		ClusterName:                o.ObjectMeta.ClusterName,
	}
	p.Status = o.Status
	p.SetKey()
}

func (p *Drift) newEvent(eventType string, o v1.Event) {
	p.Type = "event"
	//p.EventType = eventType
	p.newEventDetails(o)
	p.MetaData = ObjectMeta{
		Name:                       o.ObjectMeta.Name,
		GenerateName:               o.ObjectMeta.GenerateName,
		Namespace:                  o.ObjectMeta.Namespace,
		UID:                        o.ObjectMeta.UID,
		ResourceVersion:            o.ObjectMeta.ResourceVersion,
		Generation:                 o.ObjectMeta.Generation,
		CreationTimestamp:          o.ObjectMeta.CreationTimestamp,
		DeletionTimestamp:          o.ObjectMeta.DeletionTimestamp,
		DeletionGracePeriodSeconds: o.ObjectMeta.DeletionGracePeriodSeconds,
		Labels:                     o.ObjectMeta.Labels,
		Annotations:                o.ObjectMeta.Annotations,
		OwnerReferences:            o.ObjectMeta.OwnerReferences,
		Finalizers:                 o.ObjectMeta.Finalizers,
		ClusterName:                o.ObjectMeta.ClusterName,
	}
	p.SetKey()
}

func (p *Drift) newEventDetails(o v1.Event) {
	p.Event = Event{
		InvolvedObject:      o.InvolvedObject,
		Reason:              o.Reason,
		Message:             o.Message,
		Source:              o.Source,
		FirstTimestamp:      o.FirstTimestamp,
		LastTimestamp:       o.LastTimestamp,
		Count:               o.Count,
		Type:                o.Type,
		EventTime:           o.EventTime,
		Series:              o.Series,
		Action:              o.Action,
		Related:             o.Related,
		ReportingController: o.ReportingController,
		ReportingInstance:   o.ReportingInstance,
	}
}

func (p *Drift) newNode(eventType string, o v1.Node) {
	p.Type = "node"
	//p.EventType = eventType
	p.MetaData = ObjectMeta{
		Name:                       o.ObjectMeta.Name,
		GenerateName:               o.ObjectMeta.GenerateName,
		Namespace:                  o.ObjectMeta.Namespace,
		UID:                        o.ObjectMeta.UID,
		ResourceVersion:            o.ObjectMeta.ResourceVersion,
		Generation:                 o.ObjectMeta.Generation,
		CreationTimestamp:          o.ObjectMeta.CreationTimestamp,
		DeletionTimestamp:          o.ObjectMeta.DeletionTimestamp,
		DeletionGracePeriodSeconds: o.ObjectMeta.DeletionGracePeriodSeconds,
		Labels:                     o.ObjectMeta.Labels,
		Annotations:                o.ObjectMeta.Annotations,
		OwnerReferences:            o.ObjectMeta.OwnerReferences,
		Finalizers:                 o.ObjectMeta.Finalizers,
		ClusterName:                o.ObjectMeta.ClusterName,
	}
	p.Status = o.Status
	p.SetKey()
}

func (p *Drift) newPod(eventType string, o v1.Pod) {
	p.Type = "pod"
	//p.EventType = eventType
	p.MetaData = ObjectMeta{
		Name:                       o.ObjectMeta.Name,
		GenerateName:               o.ObjectMeta.GenerateName,
		Namespace:                  o.ObjectMeta.Namespace,
		UID:                        o.ObjectMeta.UID,
		ResourceVersion:            o.ObjectMeta.ResourceVersion,
		Generation:                 o.ObjectMeta.Generation,
		CreationTimestamp:          o.ObjectMeta.CreationTimestamp,
		DeletionTimestamp:          o.ObjectMeta.DeletionTimestamp,
		DeletionGracePeriodSeconds: o.ObjectMeta.DeletionGracePeriodSeconds,
		Labels:                     o.ObjectMeta.Labels,
		Annotations:                o.ObjectMeta.Annotations,
		OwnerReferences:            o.ObjectMeta.OwnerReferences,
		Finalizers:                 o.ObjectMeta.Finalizers,
		ClusterName:                o.ObjectMeta.ClusterName,
	}
	p.Status = o.Status

	/*rq := make([]v1.Container, len(o.Spec.Containers))

	for _, c := range o.Spec.Containers {

		rq = append(rq, c)
	}

	p.SpecContainers = rq*/
	p.SetKey()
}

func (p *Drift) Marshal() string {
	j, err := json.Marshal(p)
	if err != nil {
		fmt.Println(err)
	}
	return string(j)
}

func Marshal(e v1.Event) string {
	j, err := json.Marshal(e)
	if err != nil {
		fmt.Println(err)
	}
	return string(j)
}

func (p *Drift) Serialize() []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		fmt.Printf("error encoding object: %v", err)
	}
	return buf.Bytes()
}

func Deserialize(data []byte) Drift {
	obj := Drift{}
	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&obj)
	if err != nil {
		fmt.Printf("error decoding object: %v", err)
	}
	return obj
}
