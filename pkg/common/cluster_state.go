package common

import (
	"context"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/model"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

type ClusterStateReader struct {
	// key is Name of the Object, value wether is exists
	state    map[model.ResourceCreator]bool
	client   client.Client
	ctx      context.Context
	instance *chatv1alpha1.Rocket
}

func NewCurrentStateReader(ctx context.Context, client client.Client, rocket *chatv1alpha1.Rocket) (*ClusterStateReader, error) {
	reader := &ClusterStateReader{
		client:   client,
		instance: rocket,
		ctx:      ctx,
	}
	mongodbStsCreator := new(model.MongodbStatefulSetCreator)
	reader.state = map[model.ResourceCreator]bool{
		new(model.ServiceAccountCreator):              false,
		new(model.MongodbAuthSecretCreator):           false,
		new(model.MongodbScriptsConfigmapCreator):     false,
		&model.MongodbServiceCreator{Headless: false}: false,
		&model.MongodbServiceCreator{Headless: true}:  false,
		mongodbStsCreator:                             false,
	}

	ready, err := reader.isStatefulSetReady(mongodbStsCreator, rocket)
	if err != nil {
		return nil, err
	}
	if ready {
		rocketActions := map[model.ResourceCreator]bool{
			new(model.RocketAdminSecretCreator): false,
			new(model.RocketDeploymentCreator):  false,
			new(model.RocketServiceCreator):     false,
			new(model.RocketIngressCreator):     false,
		}
		// merge into state map
		for k, v := range rocketActions {
			reader.state[k] = v
		}
	}
	return reader, nil
}
func (c *ClusterStateReader) Read() error {
	for creator := range c.state {
		err := c.readObjectState(creator)
		if err != nil {
			return err
		}
	}
	return nil
}

// returns the new state, will be nil if resource doesnt already exists or doesnt match configuration
func (c *ClusterStateReader) readObjectState(
	resourceCreator model.ResourceCreator,
) error {
	resource := resourceCreator.CreateResource(c.instance)
	selector := resourceCreator.Selector(c.instance)
	err := c.client.Get(c.ctx, selector, resource)

	if err != nil {
		// If the resource is not found
		if apiErrors.IsNotFound(err) {
			// set state of the resource to false, doesnt exists or no match
			c.state[resourceCreator] = false
			return nil
		}
		return err
	}
	// set state of the resource to true, exists ands matches
	c.state[resourceCreator] = true
	return nil
}
