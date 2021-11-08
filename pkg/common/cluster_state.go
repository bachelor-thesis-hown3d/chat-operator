package common

import (
	"context"
	"fmt"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/model"
	runtimeClient "sigs.k8s.io/controller-runtime/pkg/client"

	apiErrors "k8s.io/apimachinery/pkg/api/errors"
)

type ClusterStateReader struct {
	// key is Name of the Object, value wether is exists
	state    map[model.ResourceCreator]runtimeClient.Object
	client   runtimeClient.Client
	ctx      context.Context
	instance *chatv1alpha1.Rocket
}

// NewCurrentStateReader creates a new CurrentStateReader with its state attached
// state map wil be initialized with the resourceCreators and nil pointers to the resources
// the Order of the resourceCreators inside the state will determine, in which order the resources will be created
func NewCurrentStateReader(ctx context.Context, client runtimeClient.Client, rocket *chatv1alpha1.Rocket) (*ClusterStateReader, error) {
	reader := &ClusterStateReader{
		client:   client,
		instance: rocket,
		ctx:      ctx,
	}
	mongodbStsCreator := new(model.MongodbStatefulSetCreator)
	reader.state = map[model.ResourceCreator]runtimeClient.Object{
		new(model.ServiceAccountCreator):              nil,
		new(model.MongodbAuthSecretCreator):           nil,
		new(model.MongodbScriptsConfigmapCreator):     nil,
		&model.MongodbServiceCreator{Headless: false}: nil,
		&model.MongodbServiceCreator{Headless: true}:  nil,
		mongodbStsCreator:                             nil,
	}

	ready, err := reader.isStatefulSetReady(mongodbStsCreator, rocket)
	if err != nil {
		return nil, fmt.Errorf("Error determining wether statefulSet %v is ready: %w", mongodbStsCreator.Name(), err)
	}
	if ready {
		rocketState := map[model.ResourceCreator]runtimeClient.Object{
			new(model.RocketAdminSecretCreator): nil,
			new(model.RocketDeploymentCreator):  nil,
			new(model.RocketServiceCreator):     nil,
			new(model.RocketIngressCreator):     nil,
		}
		// merge into state map
		for k, v := range rocketState {
			reader.state[k] = v
		}
	}
	return reader, nil
}
func (c *ClusterStateReader) Read() error {
	for creator := range c.state {
		err := c.readObjectState(creator)
		if err != nil {
			return fmt.Errorf("Error reading object State from creator %v: %w", creator.Name(), err)
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
			// set state of the resource to nil, doesnt exists or no match
			c.state[resourceCreator] = nil
			return nil
		}
		return fmt.Errorf("Error reading resource %v from cluster: %w", resource.GetName(), err)
	}
	// set state of the resource to a pointer to the resource, exists
	c.state[resourceCreator] = resource
	return nil
}
