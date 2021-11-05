package common

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/model"
)

func CreateDesiredState(clusterState *ClusterState, rocket *chatv1alpha1.Rocket) *DesiredClusterState {
	desired := &DesiredClusterState{}
	saAction := getServiceAccountDesiredState(clusterState, rocket)
	mongodbAuthSecretAction := getMongodbAuthSecretDesiredState(clusterState, rocket)
	mongodbScriptsConfigmapAction := getMongodbScriptsConfigmapDesiredState(clusterState, rocket)
	mongodbServiceAction := getMongodbServiceDesiredState(clusterState, rocket)
	mongodbHeadlessServiceAction := getMongodbHeadlessServiceDesiredState(clusterState, rocket)
	mongodbStatefulSetAction := getMongodbStatefulsetDesiredState(clusterState, rocket)
	desired.AddActions(
		saAction,
		mongodbAuthSecretAction,
		mongodbScriptsConfigmapAction,
		mongodbHeadlessServiceAction,
		mongodbServiceAction,
		mongodbStatefulSetAction,
	)
	// only add deployment, when mongodb is ready
	if ready, _ := IsStatefulSetReady(clusterState.MongodbStatefulSet); ready {
		rocketAdminSecretAction := getRocketAdminSecretDesiredState(clusterState, rocket)
		rocketDeploymentAction := getRocketDeploymentDesiredState(clusterState, rocket)
		rocketServiceAction := getRocketServiceDesiredState(clusterState, rocket)
		rocketIngressAction := getRocketIngressDesiredState(clusterState, rocket)
		desired.AddActions(
			rocketAdminSecretAction,
			rocketDeploymentAction,
			rocketServiceAction,
			rocketIngressAction,
		)
	}
	return desired
}
func getRocketAdminSecretDesiredState(clusterState *ClusterState, rocket *chatv1alpha1.Rocket) ClusterAction {
	rocketAdminSecret := model.RocketAdminSecret(rocket)

	if clusterState.RocketAdminSecret != nil {
		return nil
	}
	return GenericCreateAction{
		Object: rocketAdminSecret,
		Msg:    "Create Rocket admin secret",
	}
}
func getMongodbAuthSecretDesiredState(clusterState *ClusterState, rocket *chatv1alpha1.Rocket) ClusterAction {
	mongodbAuthSecret := model.MongodbAuthSecret(rocket)

	if clusterState.MongodbAuthSecret != nil {
		return nil
	}
	return GenericCreateAction{
		Object: mongodbAuthSecret,
		Msg:    "Create Mongodb auth secret",
	}
}
func getMongodbScriptsConfigmapDesiredState(clusterState *ClusterState, rocket *chatv1alpha1.Rocket) ClusterAction {
	mongodbScriptsConfigmap := model.MongodbScriptsConfigmap(rocket)

	if clusterState.MongodbScriptConfigmap != nil {
		return nil
	}
	return GenericCreateAction{
		Object: mongodbScriptsConfigmap,
		Msg:    "Create Mongodb scripts configmap",
	}
}
func getMongodbStatefulsetDesiredState(clusterState *ClusterState, rocket *chatv1alpha1.Rocket) ClusterAction {
	mongodbStatefulSet := model.MongodbStatefulSet(rocket)

	if clusterState.MongodbStatefulSet != nil {
		return nil
	}
	return GenericCreateAction{
		Object: mongodbStatefulSet,
		Msg:    "Create Mongodb statefulset",
	}
}
func getMongodbServiceDesiredState(clusterState *ClusterState, rocket *chatv1alpha1.Rocket) ClusterAction {
	mongodbService := model.MongodbService(rocket, false)

	if clusterState.MongodbService != nil {
		return nil
	}
	return GenericCreateAction{
		Object: mongodbService,
		Msg:    "Create Mongodb service",
	}
}
func getMongodbHeadlessServiceDesiredState(clusterState *ClusterState, rocket *chatv1alpha1.Rocket) ClusterAction {
	mongodbHeadlessService := model.MongodbService(rocket, true)

	if clusterState.MongodbHeadlessService != nil {
		return nil
	}
	return GenericCreateAction{
		Object: mongodbHeadlessService,
		Msg:    "Create Mongodb headless service",
	}
}
func getRocketDeploymentDesiredState(clusterState *ClusterState, rocket *chatv1alpha1.Rocket) ClusterAction {
	rocketDeployment := model.RocketDeployment(rocket)

	if clusterState.RocketDeployment != nil {
		return nil
	}
	return GenericCreateAction{
		Object: rocketDeployment,
		Msg:    "Create rocket deployment",
	}
}
func getRocketServiceDesiredState(clusterState *ClusterState, rocket *chatv1alpha1.Rocket) ClusterAction {
	rocketService := model.RocketService(rocket)

	if clusterState.RocketService != nil {
		return nil
	}
	return GenericCreateAction{
		Object: rocketService,
		Msg:    "Create rocket service",
	}
}
func getRocketIngressDesiredState(clusterState *ClusterState, rocket *chatv1alpha1.Rocket) ClusterAction {
	rocketIngress := model.RocketIngress(rocket)

	if clusterState.RocketIngress != nil {
		return nil
	}
	return GenericCreateAction{
		Object: rocketIngress,
		Msg:    "Create Rocket ingress",
	}
}

func getServiceAccountDesiredState(clusterState *ClusterState, rocket *chatv1alpha1.Rocket) ClusterAction {
	sa := model.ServiceAccount(rocket)
	if clusterState.ServiceAccount != nil {
		return nil
	}
	return GenericCreateAction{
		Object: sa,
		Msg:    "Create Rocket Serviceaccount",
	}
}
