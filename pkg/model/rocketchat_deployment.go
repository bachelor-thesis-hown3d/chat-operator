package model

import (
	"fmt"
	"reflect"

	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	"github.com/bachelor-thesis-hown3d/chat-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type RocketDeploymentCreator struct{}

// Name returns the ressource action of the RocketDeploymentCreator
func (c *RocketDeploymentCreator) Name() string {
	return "Rocket Deployment"
}
func (c *RocketDeploymentCreator) Update(rocket *chatv1alpha1.Rocket, cur client.Object) (client.Object, bool) {
	update := false
	dep := cur.(*appsv1.Deployment)

	// check labels
	if !reflect.DeepEqual(dep.Labels, rocket.Labels) {
		dep.Labels = rocket.Labels
		update = true
	}

	// check replicas
	curReplicas := dep.Spec.Replicas
	if *curReplicas != rocket.Spec.Replicas && rocket.Spec.Replicas > 0 {
		dep.Spec.Replicas = &rocket.Spec.Replicas
		update = true
	}

	// check image
	// check image
	curImage := dep.Spec.Template.Spec.Containers[0].Image
	newImage := "rocketchat/rocket.chat:" + rocket.Spec.Version
	if curImage != newImage {
		dep.Spec.Template.Spec.Containers[0].Image = newImage
		update = true
	}

	return dep, update
}

func (c *RocketDeploymentCreator) CreateResource(rocket *chatv1alpha1.Rocket) client.Object {
	labels := util.MergeLabels(rocketDeploymentLabels(rocket), rocket.Labels)
	replicas := rocket.Spec.Replicas

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rocket.Name + RocketWebserverDeploymentSuffix,
			Namespace: rocket.Namespace,
			Labels:    rocket.Labels,
		},
		Spec: appsv1.DeploymentSpec{
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
	if replicas > 0 {
		dep.Spec.Replicas = &replicas
	}

	return dep
}
func rocketDeploymentLabels(rocket *chatv1alpha1.Rocket) map[string]string {
	return map[string]string{
		"app":       rocket.Name,
		"component": RocketWebserverComponentName,
	}
}

func rocketDeploymentEnvVars(rocket *chatv1alpha1.Rocket) []corev1.EnvVar {
	authSecretCreator := new(MongodbAuthSecretCreator)
	adminSecretCreator := new(RocketAdminSecretCreator)
	authSecretReference := corev1.LocalObjectReference{Name: authSecretCreator.Selector(rocket).Name}
	adminSecretReference := corev1.LocalObjectReference{Name: adminSecretCreator.Selector(rocket).Name}
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
		// skips the inital setup wizard
		{
			Name:  "OVERWRITE_SETTING_Show_Setup_Wizard",
			Value: "completed",
		},
		{
			Name:  "ADMIN_USERNAME",
			Value: rocket.Spec.AdminSpec.Username,
		},
		{
			Name:  "ADMIN_EMAIL",
			Value: rocket.Spec.AdminSpec.Email,
		},
		{
			Name: "ADMIN_EMAIL",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: adminSecretReference,
					Key:                  "admin-password",
				},
			},
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

func (c *RocketDeploymentCreator) Selector(rocket *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      rocket.Name + RocketWebserverDeploymentSuffix,
		Namespace: rocket.Namespace,
	}
}
