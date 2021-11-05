package model

import (
	"github.com/hown3d/chat-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResourceCreator .
type ResourceCreator interface {
	// Name returns the ressource action of the Creator
	Name() string
	CreateResource(rocket *v1alpha1.Rocket) client.Object
	Selector(rocket *v1alpha1.Rocket) client.ObjectKey
}
