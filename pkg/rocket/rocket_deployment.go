package rocket

import (
	"fmt"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

var (
	boolTrue = true
	user     = int64(999)
	group    = int64(999)
)

// CreateOrUpdateRocketDeployment returns a Deployment object for data from m.
func CreateOrUpdateRocketDeployment(m *chatv1alpha1.Rocket, mongoEnv map[string]corev1.EnvVarSource) *appsv1.Deployment {
	labels := LabelsForRocket(m.Name)
	replicas := m.Spec.Replicas
	if replicas == 0 {
		replicas = 1
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    m.Labels,
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
						RunAsUser:    &user,
						RunAsGroup:   &group,
						RunAsNonRoot: &boolTrue,
					},
					Containers: []corev1.Container{{
						Image: fmt.Sprintf("rocket.chat:%v", m.Spec.Version),
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

	for key, value := range mongoEnv {
		dep.Spec.Template.Spec.Containers[0].Env = append(dep.Spec.Template.Spec.Containers[0].Env, corev1.EnvVar{Name: key, ValueFrom: &value})
	}

	return dep
}

func CreateOrUpdateRocketService(m *chatv1alpha1.Rocket) *corev1.Service {
	labels := LabelsForRocket(m.Name)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-service",
			Namespace: m.Namespace,
			Labels:    m.Labels,
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

func CreateOrUpdateRocketSecret() {
	panic("Not Implemented")
}

// LabelsForRocket creates a simple set of labels for rocket.
func LabelsForRocket(name string) map[string]string {
	return map[string]string{"rocketchat": name}
}
