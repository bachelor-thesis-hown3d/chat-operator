package model

import (
	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResourceCreator .
type ResourceCreator interface {
	// Name returns the ressource action of the Creator
	Name() string
	CreateResource(rocket *chatv1alpha1.Rocket) client.Object
	Selector(rocket *chatv1alpha1.Rocket) client.ObjectKey
	// Checks if a update is needed, returns true if so
	Update(rocket *chatv1alpha1.Rocket, cur client.Object) (client.Object, bool)
}
