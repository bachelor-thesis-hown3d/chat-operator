package common

import (
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
)

func IsStatefulSetReady(statefulSet *appsv1.StatefulSet) (bool, error) {
	if statefulSet == nil {
		return false, nil
	}
	// Check the correct number of replicas match and are ready
	numOfReplicasMatch := *statefulSet.Spec.Replicas == statefulSet.Status.Replicas
	allReplicasReady := statefulSet.Status.Replicas == statefulSet.Status.ReadyReplicas
	revisionsMatch := statefulSet.Status.CurrentRevision == statefulSet.Status.UpdateRevision

	return numOfReplicasMatch && allReplicasReady && revisionsMatch, nil
}

// IsDeploymentReady checks if a deployment is ready.
// The function checks wether the ReadyReplicas match the wanted Replicas and no replicaFailure condition exists
func IsDeploymentReady(deployment *appsv1.Deployment) (bool, error) {
	if deployment == nil {
		return false, nil
	}
	// if the desired Replica doesnt match the ReadyReplicas in Status, deployment isn't ready
	numOfReplicasMatch := *deployment.Spec.Replicas == deployment.Status.Replicas
	allReplicasReady := deployment.Status.Replicas == deployment.Status.ReadyReplicas
	if !numOfReplicasMatch || !allReplicasReady {
		return false, nil
	}
	// A deployment has an array of conditions
	for _, condition := range deployment.Status.Conditions {
		// One failure condition exists, if this exists, return the Reason
		if condition.Type == appsv1.DeploymentReplicaFailure {
			return false, errors.Errorf(condition.Reason)
		} 
	}
	return true, nil
}
