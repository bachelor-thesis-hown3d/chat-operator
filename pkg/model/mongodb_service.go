package model

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MongodbService(r *chatv1alpha1.Rocket) *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name + "-mongodb-service",
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				{
					Name:       "mongodb",
					Port:       27017,
					TargetPort: intstr.FromString(MongodbTargetPort),
				},
			},
			Selector: map[string]string{
				"app":       MongodbStatefulSetSelector(r).Name,
				"component": MongodbComponentName,
			},
		},
	}
	return svc
}

func MongodbServiceSelector(r *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      r.Name + "-mongodb-service",
		Namespace: r.Namespace,
	}
}
