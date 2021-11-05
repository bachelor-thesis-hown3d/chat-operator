package model

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RocketAdminSecret(r *chatv1alpha1.Rocket) *corev1.Secret {
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
func RocketAdminSecretSelector(r *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      r.Name + RocketAdminSecretSuffix,
		Namespace: r.Namespace,
	}
}
