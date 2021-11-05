package common

import (
	"context"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/model"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	metaErrors "k8s.io/apimachinery/pkg/api/meta"
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

type ClusterState struct {
	MongodbAuthSecret      *corev1.Secret
	MongodbScriptConfigmap *corev1.ConfigMap
	MongodbStatefulSet     *appsv1.StatefulSet
	MongodbService         *corev1.Service
	MongodbHeadlessService *corev1.Service
	RocketDeployment       *appsv1.Deployment
	RocketService          *corev1.Service
	RocketIngress          *networkingv1.Ingress
	ServiceAccount         *corev1.ServiceAccount
}

func (c *ClusterState) Read(context context.Context, r *chatv1alpha1.Rocket, controllerClient client.Client) error {
	err := c.readServiceAccount(context, r, controllerClient)
	if err != nil {
		return err
	}
	err = c.readMongodbAuthSecret(context, r, controllerClient)
	if err != nil {
		return err
	}
	err = c.readMongodbScriptsConfigmap(context, r, controllerClient)
	if err != nil {
		return err
	}
	err = c.readMongodbHeadlessService(context, r, controllerClient)
	if err != nil {
		return err
	}
	err = c.readMongodbService(context, r, controllerClient)
	if err != nil {
		return err
	}
	err = c.readMongodbStatefulSet(context, r, controllerClient)
	if err != nil {
		return err
	}
	err = c.readRocketDeployment(context, r, controllerClient)
	if err != nil {
		return err
	}
	err = c.readRocketIngress(context, r, controllerClient)
	if err != nil {
		return err
	}
	err = c.readRocketService(context, r, controllerClient)
	if err != nil {
		return err
	}

	// Read other things
	return nil
}

func (c *ClusterState) readMongodbAuthSecret(context context.Context, r *chatv1alpha1.Rocket, controllerClient client.Client) error {
	mongodbSecret := model.AuthSecret(r)
	mongodbSecretSelector := model.AuthSecretSelector(r)

	err := controllerClient.Get(context, mongodbSecretSelector, mongodbSecret)

	if err != nil {
		// If the resource type doesn't exist on the cluster or does exist but is not found
		if metaErrors.IsNoMatchError(err) || apiErrors.IsNotFound(err) {
			c.MongodbAuthSecret = nil
		} else {
			return err
		}
	} else {
		c.MongodbAuthSecret = mongodbSecret.DeepCopy()
	}
	return nil
}

func (c *ClusterState) readMongodbService(context context.Context, r *chatv1alpha1.Rocket, controllerClient client.Client) error {
	mongodbService := model.MongodbService(r, false)
	mongodbServiceSelector := model.MongodbServiceSelector(r, false)

	err := controllerClient.Get(context, mongodbServiceSelector, mongodbService)

	if err != nil {
		// If the resource type doesn't exist on the cluster or does exist but is not found
		if metaErrors.IsNoMatchError(err) || apiErrors.IsNotFound(err) {
			c.MongodbService = nil
		} else {
			return err
		}
	} else {
		c.MongodbService = mongodbService.DeepCopy()
	}
	return nil
}

func (c *ClusterState) readMongodbHeadlessService(context context.Context, r *chatv1alpha1.Rocket, controllerClient client.Client) error {
	mongodbHeadlessService := model.MongodbService(r, true)
	mongodbServiceSelector := model.MongodbServiceSelector(r, true)

	err := controllerClient.Get(context, mongodbServiceSelector, mongodbHeadlessService)

	if err != nil {
		// If the resource type doesn't exist on the cluster or does exist but is not found
		if metaErrors.IsNoMatchError(err) || apiErrors.IsNotFound(err) {
			c.MongodbHeadlessService = nil
		} else {
			return err
		}
	} else {
		c.MongodbHeadlessService = mongodbHeadlessService.DeepCopy()
	}
	return nil
}

func (c *ClusterState) readServiceAccount(context context.Context, r *chatv1alpha1.Rocket, controllerClient client.Client) error {
	sa := model.ServiceAccount(r)
	saSelector := model.ServiceAccountSelector(r)

	err := controllerClient.Get(context, saSelector, sa)

	if err != nil {
		// If the resource type doesn't exist on the cluster or does exist but is not found
		if metaErrors.IsNoMatchError(err) || apiErrors.IsNotFound(err) {
			c.ServiceAccount = nil
		} else {
			return err
		}
	} else {
		c.ServiceAccount = sa.DeepCopy()
	}
	return nil
}

func (c *ClusterState) readMongodbStatefulSet(context context.Context, r *chatv1alpha1.Rocket, controllerClient client.Client) error {
	mongodbStatefulSet := model.MongodbStatefulSet(r)
	mongodbStatefulSetSelector := model.MongodbStatefulSetSelector(r)

	err := controllerClient.Get(context, mongodbStatefulSetSelector, mongodbStatefulSet)

	if err != nil {
		// If the resource type doesn't exist on the cluster or does exist but is not found
		if metaErrors.IsNoMatchError(err) || apiErrors.IsNotFound(err) {
			c.MongodbStatefulSet = nil
		} else {
			return err
		}
	} else {
		c.MongodbStatefulSet = mongodbStatefulSet.DeepCopy()
	}
	return nil
}

func (c *ClusterState) readMongodbScriptsConfigmap(context context.Context, r *chatv1alpha1.Rocket, controllerClient client.Client) error {
	mongodbScriptsConfigmap := model.MongodbScriptsConfigmap(r)
	mongodbConfigmapSelector := model.MongodbConfigmapSelector(r)

	err := controllerClient.Get(context, mongodbConfigmapSelector, mongodbScriptsConfigmap)

	if err != nil {
		// If the resource type doesn't exist on the cluster or does exist but is not found
		if metaErrors.IsNoMatchError(err) || apiErrors.IsNotFound(err) {
			c.MongodbScriptConfigmap = nil
		} else {
			return err
		}
	} else {
		c.MongodbScriptConfigmap = mongodbScriptsConfigmap.DeepCopy()
	}
	return nil
}

func (c *ClusterState) readRocketDeployment(context context.Context, r *chatv1alpha1.Rocket, controllerClient client.Client) error {
	rocketDeployment := model.RocketDeployment(r)
	rocketDeploymentSelector := model.RocketDeploymentSelector(r)

	err := controllerClient.Get(context, rocketDeploymentSelector, rocketDeployment)

	if err != nil {
		// If the resource type doesn't exist on the cluster or does exist but is not found
		if metaErrors.IsNoMatchError(err) || apiErrors.IsNotFound(err) {
			c.RocketDeployment = nil
		} else {
			return err
		}
	} else {
		c.RocketDeployment = rocketDeployment.DeepCopy()
	}
	return nil
}

func (c *ClusterState) readRocketIngress(context context.Context, r *chatv1alpha1.Rocket, controllerClient client.Client) error {
	rocketIngress := model.RocketIngress(r)
	rocketIngressSelector := model.RocketIngressSelector(r)

	err := controllerClient.Get(context, rocketIngressSelector, rocketIngress)

	if err != nil {
		// If the resource type doesn't exist on the cluster or does exist but is not found
		if metaErrors.IsNoMatchError(err) || apiErrors.IsNotFound(err) {
			c.RocketIngress = nil
		} else {
			return err
		}
	} else {
		c.RocketIngress = rocketIngress.DeepCopy()
	}
	return nil
}

func (c *ClusterState) readRocketService(context context.Context, r *chatv1alpha1.Rocket, controllerClient client.Client) error {
	rocketService := model.RocketService(r)
	rocketIngressSelector := model.RocketServiceSelector(r)

	err := controllerClient.Get(context, rocketIngressSelector, rocketService)

	if err != nil {
		// If the resource type doesn't exist on the cluster or does exist but is not found
		if metaErrors.IsNoMatchError(err) || apiErrors.IsNotFound(err) {
			c.RocketService = nil
		} else {
			return err
		}
	} else {
		c.RocketService = rocketService.DeepCopy()
	}
	return nil
}

func (i *ClusterState) IsResourcesReady(r *chatv1alpha1.Rocket) (bool, error) {
	// Check mongodb statefulset is ready
	mongodbStatefulsetReady, _ := IsStatefulSetReady(i.MongodbStatefulSet)

	// Check rocketchat deployment is ready
	rocketchatDeploymentReady, err := IsDeploymentReady(i.RocketDeployment)
	if err != nil {
		return false, err
	}

	return mongodbStatefulsetReady && rocketchatDeploymentReady, nil
}
