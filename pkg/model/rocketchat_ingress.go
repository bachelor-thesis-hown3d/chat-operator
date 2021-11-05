package model

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RocketIngressCreator struct{}

// Name returns the ressource action of the RocketIngressCreator
func (m *RocketIngressCreator) Name() string {
	return "Rocket Ingress"
}
func (c *RocketIngressCreator) CreateResource(r *chatv1alpha1.Rocket) client.Object {
	serviceSelector := new(RocketServiceCreator).Selector(r)
	ingressPathType := networkingv1.PathTypeImplementationSpecific

	return &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					// TODO: Fix Host!
					Host: "test-host",
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &ingressPathType,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: serviceSelector.Name,
											Port: networkingv1.ServiceBackendPort{
												Name: "http",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (c *RocketIngressCreator) Selector(r *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      r.Name,
		Namespace: r.Namespace,
	}
}
