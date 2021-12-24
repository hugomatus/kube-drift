/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	provider "github.com/hugomatus/kube-drift/api/drift"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PodReconciler reconciles a Pod object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	store  *provider.Store
}

//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=pods/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Pod object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	var pod corev1.Pod
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	klog.Infof("Reconciling Pod %s Phase: %s\n", req.NamespacedName, pod.Status.Phase)

	drift := r.NewPodDrift(pod)

	err := r.store.Save(drift.Key, drift.Marshal())
	if err != nil {
		klog.Errorf("Failed to save event drift: with key %s\n%v", drift.Key,err)
	}
	return ctrl.Result{}, nil
}

func (r *PodReconciler) NewPodDrift(pod corev1.Pod) *provider.PodDrift {
	info := provider.GetPodInfo(&pod)
	cond := provider.GetPodConditions(&pod)
	status := provider.GetContainerStatus(&pod)
	resourceRequest := provider.GetResourceRequests(&pod)
	resourceLimit := provider.GetResourceLimits(&pod)
	labels := provider.GetPodLabels(&pod)
	annotations := provider.GetPodAnnotations(&pod)
	vols := provider.GetPodVolumes(&pod)

	d := &provider.PodDrift{
		PodInfo:             *info,
		PodConditions:       *cond,
		PodContainers:       status,
		PodResourceRequests: resourceRequest,
		PodResourceLimits:   resourceLimit,
		PodLabels:           *labels,
		PodAnnotations:      *annotations,
		PodVolumes:          vols,
	}

	d.Key = d.PodInfo["key"]
	return d
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager, store *provider.Store) error {
	r.store = store
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(r)
}
