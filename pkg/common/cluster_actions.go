package common

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterAction interface {
	Run(runner *ClusterActionRunner) (string, error)
}

type ClusterActionRunner struct {
	client  runtimeClient.Client
	context context.Context
	scheme  *runtime.Scheme
	rocket  runtime.Object
	log     logr.Logger
}

// Create an action runner to run kubernetes actions
func NewClusterActionRunner(context context.Context, client runtimeClient.Client, scheme *runtime.Scheme, rocket runtime.Object, log *logr.Logger) *ClusterActionRunner {
	return &ClusterActionRunner{
		client:  client,
		context: context,
		scheme:  scheme,
		rocket:  rocket,
	}
}

func (i *ClusterActionRunner) RunAll(desiredState DesiredClusterState) error {
	for index, action := range desiredState.actions {
		msg, err := action.Run(i)
		if err != nil {
			i.log.Info(fmt.Sprintf("(%5d) %10s %s : %s", index, "FAILED", msg, err))
			return err
		}
		i.log.Info(fmt.Sprintf("(%5d) %10s %s", index, "SUCCESS", msg))
	}

	return nil
}

func (i *ClusterActionRunner) Create(obj runtime.Object) error {
	err := controllerutil.SetControllerReference(i.rocket.(metav1.Object), obj.(metav1.Object), i.scheme)
	if err != nil {
		return err
	}

	err = i.client.Create(i.context, obj)
	if err != nil {
		return err
	}

	return nil
}

func (i *ClusterActionRunner) Update(obj runtime.Object) error {
	err := controllerutil.SetControllerReference(i.rocket.(metav1.Object), obj.(metav1.Object), i.scheme)
	if err != nil {
		return err
	}

	return i.client.Update(i.context, obj)
}

// An action to create generic kubernetes resources
// (resources that don't require special treatment)
type GenericCreateAction struct {
	Ref runtime.Object
	Msg string
}

// An action to update generic kubernetes resources
// (resources that don't require special treatment)
type GenericUpdateAction struct {
	Ref runtime.Object
	Msg string
}

func (i GenericCreateAction) Run(runner *ClusterActionRunner) (string, error) {
	return i.Msg, runner.Create(i.Ref)
}

func (i GenericUpdateAction) Run(runner *ClusterActionRunner) (string, error) {
	return i.Msg, runner.Update(i.Ref)
}
