package model

import (
	"reflect"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/chat.accso.de/v1alpha1"
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
func (c *MongodbServiceCreator) CreateResource(rocket *chatv1alpha1.Rocket) client.Object {
	labels := util.MergeLabels(mongodbStatefulSetLabels(rocket), rocket.Labels)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rocket.Name + MongodbServiceSuffix,
			Namespace: rocket.Namespace,
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
		svc.Name = rocket.Name + MongodbHeadlessServiceSuffix
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

func (c *MongodbServiceCreator) Update(rocket *chatv1alpha1.Rocket, cur client.Object) (client.Object, bool) {
	service := cur.(*corev1.Service)
	labels := util.MergeLabels(mongodbStatefulSetLabels(rocket), rocket.Labels)
	if reflect.DeepEqual(service.Labels, labels) {
		return cur, false
	}
	service.Labels = labels
	service.Spec.Selector = labels
	return service, true
}
