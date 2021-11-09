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
	"fmt"
	"reflect"
	"time"

	"github.com/bachelor-thesis-hown3d/chat-operator/pkg/common"
	"github.com/bachelor-thesis-hown3d/chat-operator/pkg/model"
	"github.com/bachelor-thesis-hown3d/chat-operator/pkg/util"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

const (
	RequeueDelay                  = 30 * time.Second
	RequeueDelayResourcesNotReady = 5 * time.Second
	RequeueDelayError             = 5 * time.Second
)

var (
	controllerLog = ctrl.Log.WithName("controllers").WithName("Rocket")
	debugLog      = controllerLog.V(1)
)

// RocketReconciler reconciles a Rocket object
type RocketReconciler struct {
	client   runtimeClient.Client
	scheme   *runtime.Scheme
	recorder record.EventRecorder
	ctx      context.Context
}

func NewRocketReconciler(client runtimeClient.Client, scheme *runtime.Scheme, recorder record.EventRecorder) *RocketReconciler {
	return &RocketReconciler{
		client:   client,
		scheme:   scheme,
		recorder: recorder,
		ctx:      context.TODO(),
	}
}

//+kubebuilder:rbac:groups=chat.accso.de,resources=rockets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=chat.accso.de,resources=rockets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=chat.accso.de,resources=rockets/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=serviceaccounts;configmaps;secrets;services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Rocket object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
func (r *RocketReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instance := &chatv1alpha1.Rocket{}
	err := r.client.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		// Request Object not found, could have been deleted after reconcile request
		// return and dont requeue
		if errors.IsNotFound(err) {
			debugLog.Info("Rocket Object not found, might have been deleted", "object", req.NamespacedName)
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}
	// set Default Versions if none specified
	if r.setVersionsIfEmpty(instance) {
		err := r.client.Update(ctx, instance)
		if err != nil {
			return r.manageError(ctx, instance, err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// append default labels to rocketchat object
	if !util.HasDefaultLabels(instance) {
		debugLog.Info("No labels found, applying default labels", "object", req.NamespacedName)
		instance.Labels = util.MergeLabels(instance.Labels, util.DefaultLabels(instance.Name))
		// update the rocket instance and requeue
		err := r.client.Update(ctx, instance)
		if err != nil {
			return r.manageError(ctx, instance, err)
		}
		return ctrl.Result{Requeue: true}, nil
	}
	// read current Cluster State
	currentState, err := common.NewCurrentStateReader(ctx, r.client, instance)
	if err != nil {
		return r.manageError(ctx, instance, err)
	}
	err = currentState.Read()
	if err != nil {
		return r.manageError(ctx, instance, err)
	}

	desiredState := common.NewDesiredState(currentState, instance)
	actionRunner := common.NewClusterActionRunner(ctx, r.client, r.scheme, instance)
	err = actionRunner.RunAll(desiredState)
	if err != nil {
		return r.manageError(ctx, instance, err)
	}
	return r.manageSuccess(ctx, instance, currentState)

}

func (r *RocketReconciler) setStatusPods(ctx context.Context, instance *chatv1alpha1.Rocket) error {
	var statusPods []chatv1alpha1.EmbeddedPod
	// Update the rocket status with the pod names.
	// List the pods for this CR's deployment.
	podList := &corev1.PodList{}
	listOpts := []runtimeClient.ListOption{
		runtimeClient.InNamespace(instance.Namespace),
		runtimeClient.MatchingLabels(instance.Labels),
	}
	if err := r.client.List(ctx, podList, listOpts...); err != nil {
		return err
	}

	// Update status.Pods if needed.
	for _, pod := range podList.Items {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.ContainersReady && pod.Status.Phase == corev1.PodRunning {
				statusPods = append(statusPods, chatv1alpha1.EmbeddedPod{Name: pod.Name})
			}
		}
	}
	if !reflect.DeepEqual(statusPods, instance.Status.Pods) {
		debugLog.Info(fmt.Sprintf("Setting instance status pods to %v", statusPods), "object", instance.Name)
		instance.Status.Pods = statusPods
	}
	return nil
}

// updates the versions of the rocket instance in the cluster to the default versions if none is specified
// returns true if a versions had to be updated
func (r *RocketReconciler) setVersionsIfEmpty(instance *chatv1alpha1.Rocket) bool {
	var dbVersionEmpty, webserverVersionEmpty = false, false
	if instance.Spec.Version == "" {
		instance.Spec.Version = model.RocketWebserverDefaultVersion
		webserverVersionEmpty = true
	}
	if instance.Spec.Database.Version == "" {
		instance.Spec.Database.Version = model.MongodbDefaultVersion
		dbVersionEmpty = true
	}
	// return empty result if no patch was needed
	return webserverVersionEmpty || dbVersionEmpty
}

func (r *RocketReconciler) manageError(ctx context.Context, instance *chatv1alpha1.Rocket, issue error) (ctrl.Result, error) {
	controllerLog.Error(issue, "error while conciling", "object", instance.Name)
	r.recorder.Event(instance, "Warning", "ProcessingError", issue.Error())

	instance.Status.Message = issue.Error()
	instance.Status.Ready = false

	err := r.client.Status().Update(ctx, instance)
	if err != nil {
		controllerLog.Error(err, "unable to update status", "object", instance.Name)
	}

	return ctrl.Result{
		RequeueAfter: RequeueDelayError,
		Requeue:      true,
	}, nil
}

func (r *RocketReconciler) manageSuccess(ctx context.Context, instance *chatv1alpha1.Rocket, currentState *common.ClusterStateReader) (ctrl.Result, error) {
	// Check if the resources are ready
	resourcesReady, err := currentState.IsResourcesReady(instance)
	if err != nil {
		return r.manageError(ctx, instance, fmt.Errorf("Error determining wether resources are ready: %w", err))
	}

	instance.Status.Ready = resourcesReady
	instance.Status.Message = "Successfull"
	err = r.setStatusPods(ctx, instance)
	if err != nil {
		return r.manageError(ctx, instance, fmt.Errorf("Error setting pod Status: %w", err))
	}

	// If resources are ready and we have not errored before now, we are in a reconciling phase
	if resourcesReady {
		instance.Status.Phase = chatv1alpha1.PhaseReconciling
	} else {
		instance.Status.Phase = chatv1alpha1.PhaseInitialising
	}

	//ingress := currentState.RocketIngress
	//if ingress != nil {
	//	instance.Status.ExternalURL = ingress.Status.
	//}

	// only update, if there are changes
	err = r.client.Status().Update(ctx, instance)
	if err != nil {
		controllerLog.Error(err, "unable to update status", "object", instance.Name)
		return ctrl.Result{
			RequeueAfter: RequeueDelayError,
			Requeue:      true,
		}, nil
	}

	if resourcesReady {
		controllerLog.Info("desired cluster state met", "object", instance.Name)
		return ctrl.Result{}, nil
	}
	debugLog.Info("desired cluster state met, but not all resources ready yet", "object", instance.Name)
	return ctrl.Result{RequeueAfter: RequeueDelayResourcesNotReady}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RocketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chatv1alpha1.Rocket{}).
		Complete(r)
}
