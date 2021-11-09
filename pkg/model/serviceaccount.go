package model

import (
	"reflect"

	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceAccountCreator struct{}

// Name returns the ressource action of the RocketServiceCreator
func (m *ServiceAccountCreator) Name() string {
	return "Service Account"
}

func (c *ServiceAccountCreator) Update(rocket *chatv1alpha1.Rocket, cur client.Object) (client.Object, bool) {
	update := false

	serviceAccount := cur.(*corev1.ServiceAccount)
	// check labels
	if !reflect.DeepEqual(serviceAccount.Labels, rocket.Labels) {
		serviceAccount.Labels = rocket.Labels
		update = true
	}

	return serviceAccount, update
}

func (c *ServiceAccountCreator) CreateResource(r *chatv1alpha1.Rocket) client.Object {
	secretCreator := new(MongodbAuthSecretCreator)
	secretSelector := secretCreator.Selector(r)
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		Secrets: []corev1.ObjectReference{{Name: secretSelector.Name, Namespace: secretSelector.Namespace}},
	}
	return sa
}

func (c *ServiceAccountCreator) Selector(r *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      r.Name,
		Namespace: r.Namespace,
	}
}
