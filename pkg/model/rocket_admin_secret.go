package model

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RocketAdminSecretCreator struct{}

// Name returns the ressource action of the RocketAdminSecretCreator
func (c *RocketAdminSecretCreator) Name() string {
	return "Rocket Admin Secret"
}

func (c *RocketAdminSecretCreator) Update(rocket *chatv1alpha1.Rocket, cur client.Object) (client.Object, bool) {
	// dont update secret
	return cur, false
}

func (c *RocketAdminSecretCreator) CreateResource(r *chatv1alpha1.Rocket) client.Object {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name + RocketAdminSecretSuffix,
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		Data: map[string][]byte{
			"admin-password": []byte(util.RandomString(25)),
		},
	}
	return secret
}
func (c *RocketAdminSecretCreator) Selector(r *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      r.Name + RocketAdminSecretSuffix,
		Namespace: r.Namespace,
	}
}
