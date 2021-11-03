package model

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ServiceAccount(r *chatv1alpha1.Rocket) *corev1.ServiceAccount {
	authSecretSelector := MongodbSecretSelector(r)
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		Secrets: []corev1.ObjectReference{{Name: authSecretSelector.Name, Namespace: authSecretSelector.Namespace}},
	}
	return sa
}

func ServiceAccountSelector(r *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      r.Name,
		Namespace: r.Namespace,
	}
}
