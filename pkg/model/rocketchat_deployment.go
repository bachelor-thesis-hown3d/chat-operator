package model

import (
	"fmt"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RocketDeployment(r *chatv1alpha1.Rocket) *appsv1.Deployment {
	replicas := r.Spec.Replicas
	if replicas >= 0 {
		replicas = 1
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name,
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: rocketDeploymentLabels(r),
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: rocketDeploymentLabels(r),
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:    &RocketUser,
						RunAsGroup:   &RocketGroup,
						RunAsNonRoot: &boolTrue,
					},
					Containers: []corev1.Container{{
						Image: fmt.Sprintf("rocket.chat:%v", r.Spec.Version),
						Name:  "rocket",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 3000,
							Name:          "http",
						}},
						LivenessProbe: &corev1.Probe{
							Handler:             corev1.Handler{HTTPGet: &corev1.HTTPGetAction{Path: "/api/info", Port: intstr.FromString("http")}},
							InitialDelaySeconds: 45,
						},
						ReadinessProbe: &corev1.Probe{
							Handler:             corev1.Handler{HTTPGet: &corev1.HTTPGetAction{Path: "/api/info", Port: intstr.FromString("http")}},
							InitialDelaySeconds: 10,
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      "rocket-data",
							MountPath: "/app/uploads",
						}},
					}},
					Volumes: []corev1.Volume{{
						Name:         "rocket-data",
						VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}},
					}},
				},
			},
		},
	}

	return dep
}
func rocketDeploymentLabels(r *chatv1alpha1.Rocket) map[string]string {
	return map[string]string{
		"app":       r.Name,
		"component": MongodbComponentName,
	}
}

func RocketDeploymentSelector(r *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      r.Name,
		Namespace: r.Namespace,
	}
}
