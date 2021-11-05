package model

import (
	"fmt"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func RocketDeployment(rocket *chatv1alpha1.Rocket) *appsv1.Deployment {
	labels := util.MergeLabels(rocket.Labels, RocketDeploymentLabels(rocket))
	replicas := rocket.Spec.Replicas
	if replicas >= 0 {
		replicas = 1
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rocket.Name + RocketWebserverDeploymentSuffix,
			Namespace: rocket.Namespace,
			Labels:    rocket.Labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:    util.CreatePointerInt64(999),
						RunAsGroup:   util.CreatePointerInt64(999),
						RunAsNonRoot: &boolTrue,
					},
					ServiceAccountName: rocket.Name,
					Containers: []corev1.Container{{
						Image: fmt.Sprintf("rocket.chat:%v", rocket.Spec.Version),
						Name:  "rocket",
						Ports: []corev1.ContainerPort{{
							ContainerPort: 3000,
							Name:          "http",
						}},
						Env: rocketDeploymentEnvVars(rocket),
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
func RocketDeploymentLabels(rocket *chatv1alpha1.Rocket) map[string]string {
	return map[string]string{
		"app":       rocket.Name,
		"component": RocketWebserverComponentName,
	}
}

func rocketDeploymentEnvVars(rocket *chatv1alpha1.Rocket) []corev1.EnvVar {
	authSecretReference := corev1.LocalObjectReference{Name: MongodbAuthSecretSelector(rocket).Name}
	return []corev1.EnvVar{
		{
			Name: "MONGO_OPLOG_URL",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: authSecretReference,
					Key:                  "oplog-uri",
				},
			},
		},
		{
			Name: "MONGO_URL",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: authSecretReference,
					Key:                  "uri",
				},
			},
		},
		{
			Name:  "OVERWRITE_SETTING_Show_Setup_Wizard",
			Value: "completed",
		},
		// skips the inital setup wizard
		{
			Name:  "OVERWRITE_SETTING_Show_Setup_Wizard",
			Value: "completed",
		},
		{
			Name: "INSTANCE_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
	}
}

func RocketDeploymentSelector(rocket *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      rocket.Name + RocketWebserverDeploymentSuffix,
		Namespace: rocket.Namespace,
	}
}
