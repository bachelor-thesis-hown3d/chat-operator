package model

import (
	"reflect"

	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	"github.com/bachelor-thesis-hown3d/chat-operator/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MongodbStatefulSetCreator struct{}

// Name returns the ressource action of the MongodbAuthSecretCreator
func (c *MongodbStatefulSetCreator) Name() string {
	return "Mongodb StatefulSet"
}
func (c *MongodbStatefulSetCreator) Update(rocket *chatv1alpha1.Rocket, cur client.Object) (client.Object, bool) {
	update := false
	sts := cur.(*appsv1.StatefulSet)

	// check labels
	d := rocket.Spec.Database
	if !reflect.DeepEqual(sts.Labels, rocket.Labels) {
		sts.Labels = rocket.Labels
		update = true
	}

	// check image
	curImage := sts.Spec.Template.Spec.Containers[0].Image
	newImage := "docker.io/bitnami/mongodb:" + d.Version
	if curImage != newImage {
		sts.Spec.Template.Spec.Containers[0].Image = newImage
		update = true
	}

	// check replicas
	curReplicas := sts.Spec.Replicas
	if *curReplicas != d.Replicas && d.Replicas > 0 {
		sts.Spec.Replicas = &d.Replicas
		update = true
	}

	// check storageSpec
	copy := sts.DeepCopy()
	createStatefulSetVolumes(rocket, d.StorageSpec, sts)
	if !reflect.DeepEqual(copy.Spec.VolumeClaimTemplates[0].Spec, sts.Spec.VolumeClaimTemplates[0].Spec) {
		update = true
	}

	return sts, update
}

func (c *MongodbStatefulSetCreator) CreateResource(rocket *chatv1alpha1.Rocket) client.Object {
	replicas := rocket.Spec.Database.Replicas
	liveness, readiness := mongodbStatefulsetHealthChecks()
	labels := util.MergeLabels(mongodbStatefulSetLabels(rocket), rocket.Labels)
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
							Env: mongodbEnvVars(rocket),
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									// 50m CPU
									corev1.ResourceCPU: *resource.NewMilliQuantity(50, resource.BinarySI),
									// 1500Mi Memory
									corev1.ResourceMemory: *resource.NewQuantity(1500*1024*1024, resource.BinarySI),
								},
							},
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

	if replicas > 0 {
		sts.Spec.Replicas = &replicas
	}

	// Create volumes
	createStatefulSetVolumes(rocket, d.StorageSpec, sts)

	return sts
}

func createStatefulSetVolumes(rocket *chatv1alpha1.Rocket, claimTemplate *chatv1alpha1.EmbeddedPersistentVolumeClaim, sts *appsv1.StatefulSet) {
	var volumes []corev1.Volume
	selector := new(MongodbScriptsConfigmapCreator).Selector(rocket)

	volumes = append(volumes, corev1.Volume{Name: "scripts",
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: selector.Name,
				},
				// 0 Prefix will assure the number is octal
				DefaultMode: util.CreatePointerInt32(0775),
			},
		},
	})

	var volumeSource corev1.VolumeSource

	if claimTemplate.Name == "" {
		claimTemplate.Name = rocket.Name + MongodbVolumeSuffix
	}
	if claimTemplate.Spec.AccessModes == nil {
		claimTemplate.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce}
	}
	if claimTemplate.Spec.VolumeMode == nil {
		defaultVolumeMode := corev1.PersistentVolumeFilesystem
		claimTemplate.Spec.VolumeMode = &defaultVolumeMode
	}

	pvcTemplate := VolumeClaimTemplate(claimTemplate)
	volumeSource.PersistentVolumeClaim = &corev1.PersistentVolumeClaimVolumeSource{
		ClaimName: rocket.Name + MongodbVolumeSuffix,
	}
	sts.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{*pvcTemplate}

	volumes = append(volumes, corev1.Volume{
		Name:         rocket.Name + MongodbVolumeSuffix,
		VolumeSource: volumeSource,
	})
	sts.Spec.Template.Spec.Volumes = volumes
}

func (c *MongodbStatefulSetCreator) Selector(r *chatv1alpha1.Rocket) client.ObjectKey {
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
	secretCreator := MongodbAuthSecretCreator{}
	authSecretRef := corev1.LocalObjectReference{Name: secretCreator.Selector(r).Name}
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
