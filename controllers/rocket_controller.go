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
	"time"

	"github.com/hown3d/chat-operator/pkg/common"
	"github.com/hown3d/chat-operator/pkg/util"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
)

const (
	RequeueDelay      = 30 * time.Second
	RequeueDelayError = 5 * time.Second
)

var controllerLog = ctrl.Log.WithName("controllers").WithName("Rocket")

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
//+kubebuilder:rbac:groups=core,resources=serviceaccounts;configmaps;secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking,resources=ingress,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Rocket object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
func (r *RocketReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.Log.WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	reqLogger.Info("Reconciling...")
	// fetch rocket instance
	instance := &chatv1alpha1.Rocket{}
	err := r.client.Get(r.ctx, req.NamespacedName, instance)
	if err != nil {
		// Request Object not found, could have been deleted after reconcile request
		// return and dont requeue
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}
	// append default labels to rocketchat object
	if instance.Labels == nil {
		instance.Labels = util.DefaultLabels(instance.Name)
	} else {
		instance.Labels = util.MergeLabels(instance.Labels, util.DefaultLabels(instance.Name))
	}
	// read current Cluster State
	currentState := &common.ClusterState{}
	err = currentState.Read(r.ctx, instance, r.client)
	if err != nil {
		return r.manageError(instance, err)
	}

	desiredState := r.setDesiredState(currentState, instance)
	actionRunner := common.NewClusterActionRunner(r.ctx, r.client, r.scheme, instance, &controllerLog)
	err = actionRunner.RunAll(desiredState)
	if err != nil {
		return r.manageError(instance, err)
	}
	return r.manageSuccess(instance, currentState)

}

func (r *RocketReconciler) updatePodNames(instance *chatv1alpha1.Rocket) error {
	var podNames []string
	// Update the rocket status with the pod names.
	// List the pods for this CR's deployment.
	podList := &corev1.PodList{}
	listOpts := []runtimeClient.ListOption{
		runtimeClient.InNamespace(instance.Namespace),
		runtimeClient.MatchingLabels(instance.Labels),
	}
	if err := r.client.List(r.ctx, podList, listOpts...); err != nil {
		return err
	}

	// Update status.Pods if needed.
	for _, pod := range podList.Items {
		podNames = append(podNames, pod.Name)
	}
	instance.Status.Pods = podNames
	return nil
}

func (r *RocketReconciler) manageError(instance *chatv1alpha1.Rocket, issue error) (ctrl.Result, error) {
	r.recorder.Event(instance, "Warning", "ProcessingError", issue.Error())

	instance.Status.Message = issue.Error()
	instance.Status.Ready = false

	err := r.client.Status().Update(r.ctx, instance)
	if err != nil {
		controllerLog.Error(err, "unable to update status of rocketchat %v", instance.Name)
	}

	return ctrl.Result{
		RequeueAfter: RequeueDelayError,
		Requeue:      true,
	}, nil
}

func (r *RocketReconciler) manageSuccess(instance *chatv1alpha1.Rocket, currentState *common.ClusterState) (ctrl.Result, error) {
	// Check if the resources are ready
	resourcesReady, err := currentState.IsResourcesReady(instance)
	if err != nil {
		return r.manageError(instance, err)
	}

	instance.Status.Ready = resourcesReady
	instance.Status.Message = "Sucessfull"
	err = r.updatePodNames(instance)
	if err != nil {
		return r.manageError(instance, err)
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

	err = r.client.Status().Update(r.ctx, instance)
	if err != nil {
		controllerLog.Error(err, "unable to update status")
		return ctrl.Result{
			RequeueAfter: RequeueDelayError,
			Requeue:      true,
		}, nil
	}

	controllerLog.Info("desired cluster state met")
	return ctrl.Result{RequeueAfter: RequeueDelay}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RocketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chatv1alpha1.Rocket{}).
		Complete(r)
}
