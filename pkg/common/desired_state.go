package common

import (
	"fmt"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/model"
)

// The desired cluster state is defined by a list of actions that have to be run to
// get from the current state to the desired state
type DesiredClusterState struct {
	actions []ClusterAction
}

func (d *DesiredClusterState) AddActions(actions ...ClusterAction) {
	for _, action := range actions {
		if action != nil {
			d.actions = append(d.actions, action)
		}
	}
}

func CreateDesiredState(clusterState *ClusterStateReader, rocket *chatv1alpha1.Rocket) *DesiredClusterState {
	desired := &DesiredClusterState{}
	for creator, resourceExists := range clusterState.state {
		action := getObjectDesiredState(rocket, resourceExists, creator)
		desired.AddActions(action)
	}
	return desired
}

func getObjectDesiredState(rocket *chatv1alpha1.Rocket, resourceExistsInState bool, creator model.ResourceCreator) ClusterAction {
	resource := creator.CreateResource(rocket)
	if resourceExistsInState {
		return nil
	}
	return GenericCreateAction{
		Object: resource,
		Msg:    fmt.Sprintf("Create %v", creator.Name()),
	}
}
