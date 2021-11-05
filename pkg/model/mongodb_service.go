package model

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MongodbService(r *chatv1alpha1.Rocket, headless bool) *corev1.Service {
	labels := util.MergeLabels(r.Labels, mongodbStatefulSetLabels(r))
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name + MongodbServiceSuffix,
			Namespace: r.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name:       "mongodb",
					Port:       27017,
					TargetPort: intstr.FromString(MongodbTargetPort),
				},
			},
			Selector: labels,
		},
	}
	if headless {
		svc.Name = r.Name + MongodbHeadlessServiceSuffix
		svc.Spec.ClusterIP = "None"
		svc.Spec.PublishNotReadyAddresses = true
	}
	return svc
}

func MongodbServiceSelector(r *chatv1alpha1.Rocket, headless bool) client.ObjectKey {
	key := client.ObjectKey{
		Name:      r.Name + MongodbServiceSuffix,
		Namespace: r.Namespace,
	}
	if headless {
		key.Name = r.Name + MongodbHeadlessServiceSuffix
	}
	return key
}
