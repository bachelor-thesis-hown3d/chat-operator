package rocket

import (
	"fmt"
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/util"
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

type config struct {
	name            string
	namespace       string
	webserverLabels map[string]string
	commonLabels    map[string]string
	rocket          *chatv1alpha1.Rocket
}

func NewConfig(r *chatv1alpha1.Rocket) *config {
	c := &config{name: "rocket-webserver" + r.Name, namespace: r.Namespace, rocket: r, webserverLabels: map[string]string{"webserver": WebserverName(r.Name)}}
	c.commonLabels = util.MergeLabels(c.webserverLabels, r.Labels)
	return c
}

func WebserverName(name string) string {
	return name + "-database"
}

func (c *config) MakeServiceAccount() *corev1.ServiceAccount {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: c.namespace,
		},
		Secrets: []corev1.ObjectReference{{Name: c.name + "-auth", Namespace: c.namespace}},
	}
	return sa
}

// MakeDeployment returns a Deployment object for data from m.
func (c *config) MakeDeployment(mongoEnv map[string]corev1.EnvVarSource) *appsv1.Deployment {
	replicas := c.rocket.Spec.Replicas
	if replicas == 0 {
		replicas = 1
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: c.namespace,
			Labels:    c.commonLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: c.webserverLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: c.webserverLabels,
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:    &user,
						RunAsGroup:   &group,
						RunAsNonRoot: &boolTrue,
					},
					Containers: []corev1.Container{{
						Image: fmt.Sprintf("rocket.chat:%v", c.rocket.Spec.Version),
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

func (c *config) MakeService() *corev1.Service {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name + "-service",
			Namespace: c.namespace,
			Labels:    c.commonLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{{
				TargetPort: intstr.FromString("http"),
				Port:       80,
				Name:       "http",
			}},
			Selector: c.webserverLabels,
		},
	}
	return service
}
