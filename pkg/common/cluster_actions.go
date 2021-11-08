package common

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

var actionLogger = ctrl.Log.WithName("actions").WithName("Rocket")

type ClusterAction interface {
	Run(runner *ClusterActionRunner) (string, error)
}

type ClusterActionRunner struct {
	client  runtimeClient.Client
	context context.Context
	scheme  *runtime.Scheme
	parent  runtimeClient.Object
	log     *logr.Logger
}

// Create an action runner to run kubernetes actions
func NewClusterActionRunner(context context.Context, client runtimeClient.Client, scheme *runtime.Scheme, rocket runtimeClient.Object) *ClusterActionRunner {
	return &ClusterActionRunner{
		client:  client,
		context: context,
		scheme:  scheme,
		parent:  rocket,
	}
}

func (runner *ClusterActionRunner) RunAll(desiredState *desiredClusterState) error {
	for index, action := range desiredState.actions {
		msg, err := action.Run(runner)
		if err != nil {
			actionLogger.Info(fmt.Sprintf("(%5d) %10s %s : %s", index, "FAILED", msg, err))
			return err
		}
		actionLogger.Info(fmt.Sprintf("(%5d) %10s %s", index, "SUCCESS", msg), "object", runner.parent.GetName())
	}

	return nil
}

func (runner *ClusterActionRunner) Create(obj runtimeClient.Object) error {
	err := controllerutil.SetControllerReference(runner.parent.(metav1.Object), obj.(metav1.Object), runner.scheme)
	if err != nil {
		return fmt.Errorf("Error setting controller owner reference on resource %v to owner %v: %w", obj.GetName(), runner.parent.GetName(), err)
	}

	err = runner.client.Create(runner.context, obj)
	if err != nil {
		return fmt.Errorf("Error creating resource %v: %w", obj.GetName(), err)
	}

	return nil
}

func (runner *ClusterActionRunner) Update(obj runtimeClient.Object) error {
	err := controllerutil.SetControllerReference(runner.parent.(metav1.Object), obj.(metav1.Object), runner.scheme)
	if err != nil {
		return err
	}

	return runner.client.Update(runner.context, obj)
}

// An action to create generic kubernetes resources
// (resources that don't require special treatment)
type GenericCreateAction struct {
	runtimeClient.Object
	Msg string
}

// An action to update generic kubernetes resources
// (resources that don't require special treatment)
type GenericUpdateAction struct {
	runtimeClient.Object
	Msg string
}

func (action GenericCreateAction) Run(runner *ClusterActionRunner) (string, error) {
	return action.Msg, runner.Create(action.Object)
}

func (action GenericUpdateAction) Run(runner *ClusterActionRunner) (string, error) {
	return action.Msg, runner.Update(action.Object)
}
