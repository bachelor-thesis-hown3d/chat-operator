package controllers

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/common"
	"github.com/hown3d/chat-operator/pkg/model"
)

func (r *RocketReconciler) setDesiredState(clusterState *common.ClusterState, rocket *chatv1alpha1.Rocket) common.DesiredClusterState {
	desired := common.DesiredClusterState{}
	saAction := r.getServiceAccountDesiredState(clusterState, rocket)
	mongodbAuthSecretAction := r.getMongodbAuthSecretDesiredState(clusterState, rocket)
	mongodbScriptsConfigmapAction := r.getMongodbScriptsConfigmapDesiredState(clusterState, rocket)
	mongodbServiceAction := r.getMongodbServiceDesiredState(clusterState, rocket)
	mongodbStatefulSetAction := r.getMongodbStatefulsetDesiredState(clusterState, rocket)
	rocketDeploymentAction := r.getRocketDeploymentDesiredState(clusterState, rocket)
	rocketServiceAction := r.getRocketServiceDesiredState(clusterState, rocket)
	rocketIngressAction := r.getRocketIngressDesiredState(clusterState, rocket)
	desired.AddActions(saAction, mongodbAuthSecretAction, mongodbScriptsConfigmapAction, mongodbServiceAction, mongodbStatefulSetAction, rocketDeploymentAction, rocketServiceAction, rocketIngressAction)
	return desired
}

func (r *RocketReconciler) getMongodbAuthSecretDesiredState(clusterState *common.ClusterState, rocket *chatv1alpha1.Rocket) common.ClusterAction {
	mongodbAuthSecret := model.MongodbSecret(rocket)

	if clusterState.MongodbAuthSecret != nil {
		return nil
	}
	return common.GenericCreateAction{
		Ref: mongodbAuthSecret,
		Msg: "Create Mongodb auth secret",
	}
}
func (r *RocketReconciler) getMongodbScriptsConfigmapDesiredState(clusterState *common.ClusterState, rocket *chatv1alpha1.Rocket) common.ClusterAction {
	mongodbScriptsConfigmap := model.MongodbScriptsConfigmap(rocket)

	if clusterState.MongodbScriptConfigmap != nil {
		return nil
	}
	return common.GenericCreateAction{
		Ref: mongodbScriptsConfigmap,
		Msg: "Create Mongodb scripts configmap",
	}
}
func (r *RocketReconciler) getMongodbStatefulsetDesiredState(clusterState *common.ClusterState, rocket *chatv1alpha1.Rocket) common.ClusterAction {
	mongodbStatefulSet := model.MongodbStatefulSet(rocket)

	if clusterState.MongodbStatefulSet != nil {
		return nil
	}
	return common.GenericCreateAction{
		Ref: mongodbStatefulSet,
		Msg: "Create Mongodb statefulset",
	}
}
func (r *RocketReconciler) getMongodbServiceDesiredState(clusterState *common.ClusterState, rocket *chatv1alpha1.Rocket) common.ClusterAction {
	mongodbService := model.MongodbService(rocket)

	if clusterState.MongodbAuthSecret != nil {
		return nil
	}
	return common.GenericCreateAction{
		Ref: mongodbService,
		Msg: "Create Mongodb service",
	}
}
func (r *RocketReconciler) getRocketDeploymentDesiredState(clusterState *common.ClusterState, rocket *chatv1alpha1.Rocket) common.ClusterAction {
	rocketDeployment := model.RocketDeployment(rocket)

	if clusterState.MongodbAuthSecret != nil {
		return nil
	}
	return common.GenericCreateAction{
		Ref: rocketDeployment,
		Msg: "Create rocket deployment",
	}
}
func (r *RocketReconciler) getRocketServiceDesiredState(clusterState *common.ClusterState, rocket *chatv1alpha1.Rocket) common.ClusterAction {
	rocketService := model.RocketService(rocket)

	if clusterState.MongodbAuthSecret != nil {
		return nil
	}
	return common.GenericCreateAction{
		Ref: rocketService,
		Msg: "Create rocket service",
	}
}
func (r *RocketReconciler) getRocketIngressDesiredState(clusterState *common.ClusterState, rocket *chatv1alpha1.Rocket) common.ClusterAction {
	rocketIngress := model.RocketIngress(rocket)

	if clusterState.MongodbAuthSecret != nil {
		return nil
	}
	return common.GenericCreateAction{
		Ref: rocketIngress,
		Msg: "Create Rocket ingress",
	}
}

func (r *RocketReconciler) getServiceAccountDesiredState(clusterState *common.ClusterState, rocket *chatv1alpha1.Rocket) common.ClusterAction {
	rocketIngress := model.RocketIngress(rocket)
	if clusterState.ServiceAccount != nil {
		return nil
	}
	return common.GenericCreateAction{
		Ref: rocketIngress,
		Msg: "Create Rocket Serviceaccount",
	}
}
