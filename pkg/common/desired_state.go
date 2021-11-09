package common

import (
	"fmt"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/model"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// The desired cluster state is defined by a list of actions that have to be run to
// get from the current state to the desired state
type desiredClusterState struct {
	actions []ClusterAction
}

// NewDesiredState creates a new DesiredState regarding the clusterState
func NewDesiredState(clusterState *ClusterStateReader, rocket *chatv1alpha1.Rocket) *desiredClusterState {
	desired := &desiredClusterState{}
	for creator, resource := range clusterState.state {
		action := getObjectDesiredState(rocket, resource, creator)
		if action != nil {
			desired.actions = append(desired.actions, action)
		}
	}
	return desired
}

func getObjectDesiredState(rocket *chatv1alpha1.Rocket, resourceInState client.Object, creator model.ResourceCreator) ClusterAction {
	resource := creator.CreateResource(rocket)
	// resourceInState is nil, doesnt exist
	if resourceInState == nil {
		return GenericCreateAction{
			Object: resource,
			Msg:    fmt.Sprintf("Create %v", creator.Name()),
		}
	}
	newResource, needsUpdate := creator.Update(rocket, resourceInState)
	if needsUpdate {
		return GenericUpdateAction{
			Object: newResource,
			Msg:    fmt.Sprintf("Update %v", creator.Name()),
		}
	}
	return nil
}
