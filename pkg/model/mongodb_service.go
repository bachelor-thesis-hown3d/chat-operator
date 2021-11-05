package model

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MongodbServiceCreator struct {
	Headless bool
}

// Name returns the ressource action of the MongodbAuthSecretCreator
func (m *MongodbServiceCreator) Name() string {
	return "Mongodb Service"
}
func (c *MongodbServiceCreator) CreateResource(r *chatv1alpha1.Rocket) client.Object {
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
	if c.Headless {
		svc.Name = r.Name + MongodbHeadlessServiceSuffix
		svc.Spec.ClusterIP = "None"
		svc.Spec.PublishNotReadyAddresses = true
	}
	return svc
}

func (c *MongodbServiceCreator) Selector(r *chatv1alpha1.Rocket) client.ObjectKey {
	key := client.ObjectKey{
		Name:      r.Name + MongodbServiceSuffix,
		Namespace: r.Namespace,
	}
	if c.Headless {
		key.Name = r.Name + MongodbHeadlessServiceSuffix
	}
	return key
}
