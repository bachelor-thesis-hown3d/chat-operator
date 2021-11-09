package model

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func VolumeClaimTemplate(t *chatv1alpha1.EmbeddedPersistentVolumeClaim) *corev1.PersistentVolumeClaim {
	pvc := corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: t.APIVersion,
			Kind:       t.Kind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        t.Name,
			Labels:      t.Labels,
			Annotations: t.Annotations,
		},
		Spec:   t.Spec,
		Status: t.Status,
	}
	return &pvc
}
