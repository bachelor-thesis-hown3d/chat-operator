package model

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MongodbStatefulSet(rocket *chatv1alpha1.Rocket) *appsv1.StatefulSet {
	replicas := rocket.Spec.Database.Replicas
	liveness, readiness := mongodbStatefulsetHealthChecks()
	labels := util.MergeLabels(rocket.Labels, mongodbStatefulSetLabels(rocket))
	d := rocket.Spec.Database
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rocket.Name + MongodbStatefulSetSuffix,
			Namespace: rocket.Namespace,
			Labels:    rocket.Labels,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: rocket.Name + MongodbHeadlessServiceSuffix,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: "scripts",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: MongodbConfigmapSelector(rocket).Name,
								},
								// 0 Prefix will assure the number is octal
								DefaultMode: util.CreatePointerInt32(0775),
							},
						},
					}},
					ServiceAccountName: rocket.Name,
					Containers: []corev1.Container{
						{
							Name:    "mongodb",
							Image:   "docker.io/bitnami/mongodb:" + d.Version,
							Command: []string{MongodbScriptPath},
							Ports: []corev1.ContainerPort{
								{
									Name:          "mongodb",
									ContainerPort: 27017,
								},
							},
							Env:       mongodbEnvVars(rocket),
							Resources: corev1.ResourceRequirements{},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      rocket.Name + MongodbVolumeSuffix,
									MountPath: "/bitnami/mongodb",
								},
								{
									Name:      "scripts",
									MountPath: MongodbScriptPath,
									SubPath:   "setup.sh",
								},
							},
							ReadinessProbe: readiness,
							LivenessProbe:  liveness,
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

	if replicas != 0 {
		sts.Spec.Replicas = &replicas
	}

	// Create volumes
	storageSpec := d.StorageSpec
	if storageSpec == nil {
		sts.Spec.Template.Spec.Volumes = append(sts.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: rocket.Name + MongodbVolumeSuffix,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	} else {
		pvcTemplate := VolumeClaimTemplate(*storageSpec)
		if pvcTemplate.Name == "" {
			pvcTemplate.Name = rocket.Name + MongodbVolumeSuffix
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
		Name:      r.Name + MongodbStatefulSetSuffix,
		Namespace: r.Namespace,
	}
}

func mongodbStatefulsetHealthChecks() (liveness, readiness *corev1.Probe) {
	liveness = &corev1.Probe{
		Handler: corev1.Handler{Exec: &corev1.ExecAction{Command: []string{
			"mongo", "--disableImplicitSessions", "--eval", "db.adminCommand('ping')",
		}}},
		InitialDelaySeconds: 30,
	}
	readiness = &corev1.Probe{
		Handler: corev1.Handler{Exec: &corev1.ExecAction{Command: []string{
			"bash", "-ec", MongodbReadinessCommand,
		}}},
		InitialDelaySeconds: 5,
	}
	return
}

func mongodbStatefulSetLabels(r *chatv1alpha1.Rocket) map[string]string {
	return map[string]string{
		"app":       r.Name,
		"component": MongodbComponentName,
	}
}

func mongodbEnvVars(r *chatv1alpha1.Rocket) []corev1.EnvVar {
	authSecretRef := corev1.LocalObjectReference{Name: MongodbAuthSecretSelector(r).Name}
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
			Value: r.Name + MongodbHeadlessServiceSuffix,
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
					Key:                  "user",
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
