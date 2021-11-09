package common

import (
	"fmt"
	"reflect"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/model"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

func (c *ClusterStateReader) isStatefulSetReady(creator model.ResourceCreator, rocket *chatv1alpha1.Rocket) (bool, error) {
	// Check statefulset is ready
	sts := &appsv1.StatefulSet{}
	copy := sts.DeepCopy()
	selector := creator.Selector(rocket)
	err := c.client.Get(c.ctx, selector, sts)
	if err != nil && !apiErrors.IsNotFound(err) {
		return false, err
	}
	if reflect.DeepEqual(sts, copy) {
		return false, nil
	}
	// Check the correct number of replicas match and are ready
	numOfReplicasMatch := *sts.Spec.Replicas == sts.Status.Replicas
	allReplicasReady := sts.Status.Replicas == sts.Status.ReadyReplicas
	revisionsMatch := sts.Status.CurrentRevision == sts.Status.UpdateRevision

	return numOfReplicasMatch && allReplicasReady && revisionsMatch, nil
}

// isDeploymentReady checks if a deployment is ready.
// The function checks wether the ReadyReplicas match the wanted Replicas and no replicaFailure condition exists
func (c *ClusterStateReader) isDeploymentReady(creator model.ResourceCreator, rocket *chatv1alpha1.Rocket) (bool, error) {
	// Check Rocket Deployment is ready
	dep := &appsv1.Deployment{}
	selector := creator.Selector(rocket)
	err := c.client.Get(c.ctx, selector, dep)
	if err != nil {
		return false, err
	}

	if dep == nil {
		return false, nil
	}
	// if the desired Replica doesnt match the ReadyReplicas in Status, deployment isn't ready
	numOfReplicasMatch := *dep.Spec.Replicas == dep.Status.Replicas
	allReplicasReady := dep.Status.Replicas == dep.Status.ReadyReplicas
	if !numOfReplicasMatch || !allReplicasReady {
		return false, nil
	}
	// A deployment has an array of conditions
	for _, condition := range dep.Status.Conditions {
		// One failure condition exists, if this exists, return the Reason
		if condition.Type == appsv1.DeploymentReplicaFailure {
			return false, errors.Errorf(condition.Reason)
		}
	}
	return true, nil
}

// IsResourcesReady checks if the mongodb StatefulSet and the Rocketchat Deployment are ready
func (c *ClusterStateReader) IsResourcesReady(rocket *chatv1alpha1.Rocket) (bool, error) {
	var mongodbStatefulSetReady, rocketchatDeploymentReady bool
	var err error
	for creator := range c.state {
		if val, ok := creator.(*model.MongodbStatefulSetCreator); ok {
			mongodbStatefulSetReady, err = c.isStatefulSetReady(val, rocket)
			if err != nil {
				return false, fmt.Errorf("Error determining if statefulSet is ready: %w", err)
			}
		}
		if val, ok := creator.(*model.RocketDeploymentCreator); ok {
			rocketchatDeploymentReady, err = c.isDeploymentReady(val, rocket)
			if err != nil {
				return false, fmt.Errorf("Error determining if deployment is ready: %w", err)
			}
		}
	}

	return mongodbStatefulSetReady && rocketchatDeploymentReady, nil
}
