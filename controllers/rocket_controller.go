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
	"github.com/hown3d/chat-operator/pkg/mongo"
	"github.com/hown3d/chat-operator/pkg/util"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	rocketUtil "github.com/hown3d/chat-operator/pkg/rocket"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// RocketReconciler reconciles a Rocket object
type RocketReconciler struct {
	runtimeClient.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=chat.accso.de,resources=rockets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=chat.accso.de,resources=rockets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=chat.accso.de,resources=rockets/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Rocket object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *RocketReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	// fetch rocket instance
	rocket := &chatv1alpha1.Rocket{}
	err := r.Get(ctx, req.NamespacedName, rocket)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// append default labels to rocketchat object
	if rocket.Labels == nil {
		rocket.Labels = util.DefaultLabels(rocket.Name)
	} else {
		rocket.Labels = util.MergeLabels(rocket.Labels, util.DefaultLabels(rocket.Name))
	}

	// Check if the mongodb already exists, if not create a new one.
	result, err := r.createMongo(ctx, rocket)
	if err != nil {
		return result, err
	}
	if result.Requeue {
		// err is always nil here
		return result, err
	}

	//
	// Check if the deployment already exists, if not create a new deployment.
	result, err = r.createWebserver(ctx, rocket, mongo.DatabaseName(rocket.Name))
	if err != nil {
		return result, err
	}
	if result.Requeue {
		// err is always nil here
		return result, err
	}

	// Update the rocket status with the pod names.
	// List the pods for this CR's deployment.
	podList := &corev1.PodList{}
	listOpts := []runtimeClient.ListOption{
		runtimeClient.InNamespace(rocket.Namespace),
		runtimeClient.MatchingLabels(rocket.Labels),
	}
	if err = r.List(ctx, podList, listOpts...); err != nil {
		return ctrl.Result{}, err
	}
	//
	// Update status.Nodes if needed.
	podNames := getPodNames(podList.Items)
	if !reflect.DeepEqual(podNames, rocket.Status.Pods) {
		rocket.Status.Pods = podNames
		if err := r.Status().Update(ctx, rocket); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func getPodNames(pods []corev1.Pod) (podNames []string) {
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}

func (r *RocketReconciler) createMongo(ctx context.Context, obj *chatv1alpha1.Rocket) (ctrl.Result, error) {
	conf := mongo.NewConfig(obj)

	secret := conf.MakeSecret()
	result, err := r.createOrUpdateResource(ctx, obj, secret)
	if err != nil {
		return result, err
	}
	if result.Requeue {
		// err is always nil here
		return result, err
	}
	cm := conf.MakeScriptsConfigmap()
	result, err = r.createOrUpdateResource(ctx, obj, cm)
	if err != nil {
		return result, err
	}
	if result.Requeue {
		// err is always nil here
		return result, err
	}

	service := conf.MakeStatefulSetService()
	result, err = r.createOrUpdateResource(ctx, obj, service)
	if err != nil {
		return result, err
	}
	if result.Requeue {
		// err is always nil here
		return result, err
	}

	sts := conf.MakeStatefulSet(obj.Spec.Database, cm.Name, service.Name)
	result, err = r.createOrUpdateResource(ctx, obj, sts)
	if err != nil {
		return result, err
	}
	if result.Requeue {
		// err is always nil here
		return result, err
	}

	return ctrl.Result{}, err
}

func (r *RocketReconciler) createWebserver(ctx context.Context, obj *chatv1alpha1.Rocket, mongoName string) (ctrl.Result, error) {
	mongoEnv := map[string]corev1.EnvVarSource{
		"MONGO_URL": {
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: mongoName + "-auth"},
				Key:                  "uri",
			},
		},
		"MONGO_OPLOG_URL": {
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: mongoName + "-auth"},
				Key:                  "oplog-uri",
			},
		},
	}
	conf := rocketUtil.NewConfig(obj)

	serviceaccount := conf.MakeServiceAccount()
	result, err := r.createOrUpdateResource(ctx, obj, serviceaccount)
	if err != nil {
		return result, err
	}
	if result.Requeue {
		// err is always nil here
		return result, err
	}

	deployment := conf.MakeDeployment(mongoEnv)
	result, err = r.createOrUpdateResource(ctx, obj, deployment)
	if err != nil {
		return result, err
	}
	if result.Requeue {
		// err is always nil here
		return result, err
	}

	service := conf.MakeService()
	result, err = r.createOrUpdateResource(ctx, obj, service)
	if err != nil {
		return result, err
	}
	if result.Requeue {
		// err is always nil here
		return result, err
	}

	return ctrl.Result{}, err
}

func (r *RocketReconciler) checkIfResourceExists(ctx context.Context, resource runtimeClient.Object) (bool, error) {
	var found runtimeClient.Object
	err := r.Get(ctx, types.NamespacedName{Name: resource.GetName(), Namespace: resource.GetNamespace()}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return false, nil
		}
		// error is not wanted, throw it
		return true, fmt.Errorf("Error getting resource %v: %w", resource.GetName(), err)
	}
	// resource exists
	return true, nil
}

// Check if the deployment already exists, if not create it.
func (r *RocketReconciler) createOrUpdateResource(ctx context.Context, obj *chatv1alpha1.Rocket, resource runtimeClient.Object) (ctrl.Result, error) {
	exists, err := r.checkIfResourceExists(ctx, resource)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error checking if resource %v exists: %w", resource.GetName(), err)
	}
	// create
	if !exists {
		// Set rocket instance as the owner of the resource
		// NOTE: calling SetControllerReference, and setting owner references in
		// general, is important as it allows deleted objects to be garbage collected.
		err := controllerutil.SetControllerReference(obj, resource, r.Scheme)
		if err != nil {
			return ctrl.Result{}, fmt.Errorf("error setting owner reference for %v to rocket %v: %w", obj.Name, resource.GetName(), err)
		}
		// create Resource and requeue if no error is found
		if err = r.Create(ctx, resource); err != nil {
			return ctrl.Result{}, fmt.Errorf("error creating resource %v: %w", resource.GetName(), err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RocketReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&chatv1alpha1.Rocket{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
