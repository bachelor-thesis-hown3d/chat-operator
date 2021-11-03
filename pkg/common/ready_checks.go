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

func IsDeploymentReady(deployment *appsv1.Deployment) (bool, error) {
	if deployment == nil {
		return false, nil
	}
	// A deployment has an array of conditions
	for _, condition := range deployment.Status.Conditions {
		// One failure condition exists, if this exists, return the Reason
		if condition.Type == appsv1.DeploymentReplicaFailure {
			return false, errors.Errorf(condition.Reason)
			// A successful deployment will have the progressing condition type as true
		} else if condition.Type == appsv1.DeploymentProgressing && condition.Status != "True" {
			return false, nil
		}
	}
	return true, nil
}
