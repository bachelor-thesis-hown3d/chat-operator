package model

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RocketServiceCreator struct{}

// Name returns the ressource action of the RocketServiceCreator
func (m *RocketServiceCreator) Name() string {
	return "Rocket Service"
}
func (c *RocketServiceCreator) CreateResource(r *chatv1alpha1.Rocket) client.Object {
	labels := util.MergeLabels(r.Labels, mongodbStatefulSetLabels(r))
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name + RocketWebserverServiceSuffix,
			Namespace: r.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{{
				TargetPort: intstr.FromString("http"),
				Port:       80,
				Name:       "http",
			}},
			Selector: labels,
		},
	}
	return service
}

func (c *RocketServiceCreator) Selector(r *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      r.Name + RocketWebserverServiceSuffix,
		Namespace: r.Namespace,
	}
}
