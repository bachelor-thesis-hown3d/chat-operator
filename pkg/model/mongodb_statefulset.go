package model

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MongodbStatefulSet(r *chatv1alpha1.Rocket) *appsv1.StatefulSet {
	d := r.Spec.Database
	if d.Version == "" {
		d.Version = MongodbDefaultVersion
	}
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name + "-mongodb",
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		Spec: appsv1.StatefulSetSpec{
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: "scripts",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: MongodbConfigmapSelector(r).Name,
								},
								DefaultMode: &MongodbScriptMode,
							},
						},
					}},
					Containers: []corev1.Container{
						{
							Name:    "mongodb",
							Image:   "docker.io/bitnami/mongodb" + d.Version,
							Command: []string{MongodbScriptPath},
							Ports: []corev1.ContainerPort{
								{
									Name:          "mongodb",
									ContainerPort: 27017,
								},
							},
							Env:       mongoEnvVars(r),
							Resources: corev1.ResourceRequirements{},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "datadir",
									MountPath: "/bitnami/mongodb",
								},
								{
									Name:      "scripts",
									MountPath: MongodbScriptPath,
									SubPath:   "setup.sh",
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{Exec: &corev1.ExecAction{Command: []string{
									"bash", "-ec", MongodbReadinessCommand,
								}}},
								InitialDelaySeconds: 5,
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{Exec: &corev1.ExecAction{Command: []string{
									"mongo", "--disableImplicitSessions", "--eval", "\"db.adminCommand('ping')\"",
								}}},
								InitialDelaySeconds: 30,
							},
							SecurityContext: &corev1.SecurityContext{
								RunAsUser:    &MongodbUser,
								RunAsNonRoot: &boolTrue,
							},
						},
					},
				},
			},
		},
	}

	// Create volumes
	storageSpec := d.StorageSpec
	if storageSpec == nil {
		sts.Spec.Template.Spec.Volumes = append(sts.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: r.Name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	} else {
		pvcTemplate := util.MakeVolumeClaimTemplate(*storageSpec)
		if pvcTemplate.Name == "" {
			pvcTemplate.Name = r.Name + "-mongodb"
		}
		if storageSpec.Spec.AccessModes == nil {
			pvcTemplate.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
		} else {
			pvcTemplate.Spec.AccessModes = storageSpec.Spec.AccessModes
		}
		pvcTemplate.Spec.Resources = storageSpec.Spec.Resources
		pvcTemplate.Spec.Selector = storageSpec.Spec.Selector
		sts.Spec.VolumeClaimTemplates = append(sts.Spec.VolumeClaimTemplates, *pvcTemplate)
	}
	return sts
}

func MongodbStatefulSetSelector(r *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      r.Name + "-mongodb",
		Namespace: r.Namespace,
	}
}

func mongoEnvVars(r *chatv1alpha1.Rocket) []corev1.EnvVar {
	authSecretRef := corev1.LocalObjectReference{MongodbSecretSelector(r).Name}
	return []corev1.EnvVar{
		{
			Name: "MY_POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: "MY_POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name:  "K8S_SERVICE_NAME",
			Value: r.Name + "-mongodb-service",
		},
		{
			Name:  "MONGODB_INITIAL_PRIMARY_HOST",
			Value: r.Name + "-mongodb-0.$(K8S_SERVICE_NAME).$(MY_POD_NAMESPACE).svc.cluster.local",
		},
		{
			Name:  "MONGODB_ADVERTISED_HOSTNAME",
			Value: "$(MY_POD_NAME).$(K8S_SERVICE_NAME).$(MY_POD_NAMESPACE).svc.cluster.local",
		},
		{
			Name:  "MONGODB_REPLICA_SET_NAME",
			Value: "rs0",
		},
		{
			Name: "MONGODB_USERNAME",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: authSecretRef,
					Key:                  "username",
				},
			},
		}, {
			Name: "MONGODB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: authSecretRef,
					Key:                  "password",
				},
			},
		},
		{
			Name:  "MONGODB_DATABASE",
			Value: "rocketchat",
		}, {
			Name: "MONGODB_ROOT_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: authSecretRef,
					Key:                  "root-password",
				},
			},
		}, {
			Name: "MONGODB_REPLICA_SET_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: authSecretRef,
					Key:                  "replicaset-key",
				},
			},
		},
		{
			Name:  "MONGODB_ALLOW_EMPTY_PASSWORD",
			Value: "no",
		},
	}
}
