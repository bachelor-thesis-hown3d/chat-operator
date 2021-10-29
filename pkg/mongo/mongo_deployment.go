package mongo

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
	replicas         = int32(1)
	defaultVersion   = "4.4.6-debian-10-r29"
	scriptPath       = "/scripts/setup.sh"
	scriptMode       = int32(0755)
	user             = int64(1001)
	boolTrue         = true
	readinessCommand = `
# Run the proper check depending on the version
[[ $(mongo --version | grep "MongoDB shell") =~ ([0-9]+\.[0-9]+\.[0-9]+) ]] && VERSION=${BASH_REMATCH[1]}
. /opt/bitnami/scripts/libversion.sh
VERSION_MAJOR="$(get_sematic_version "$VERSION" 1)"
VERSION_MINOR="$(get_sematic_version "$VERSION" 2)"
VERSION_PATCH="$(get_sematic_version "$VERSION" 3)"
if [[ "$VERSION_MAJOR" -ge 4 ]] && [[ "$VERSION_MINOR" -ge 4 ]] && [[ "$VERSION_PATCH" -ge 2 ]]; then
    mongo --disableImplicitSessions $TLS_OPTIONS --eval 'db.hello().isWritablePrimary || db.hello().secondary' | grep -q 'true'
else
    mongo --disableImplicitSessions $TLS_OPTIONS --eval 'db.isMaster().ismaster || db.isMaster().secondary' | grep -q 'true'
fi`
)

type config struct {
	name           string
	namespace      string
	databaseLabels map[string]string
	commonLabels   map[string]string
}

func NewConfig(r *chatv1alpha1.Rocket) *config {
	c := &config{
		namespace: r.Namespace,
	}
	c.name = DatabaseName(r.Name)
	c.databaseLabels = map[string]string{"database": c.name + "-database"}
	c.commonLabels = util.MergeLabels(r.Labels, c.databaseLabels)
	return c
}

func DatabaseName(name string) string {
	return name + "-database"
}

func (c *config) MakeStatefulSetService() *corev1.Service {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:   c.name,
			Labels: c.commonLabels,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
			Ports: []corev1.ServicePort{
				{
					Name:       "mongodb",
					Port:       27017,
					TargetPort: intstr.FromString("mongodb"),
				},
			},
			Selector: c.databaseLabels,
		},
	}

	return svc
}

func (c *config) MakeStatefulSet(d *chatv1alpha1.RocketDatabase, scriptsConfigMapName string, serviceName string) *appsv1.StatefulSet {
	if d.Version == "" {
		d.Version = defaultVersion
	}
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name,
			Namespace: c.namespace,
			Labels:    c.commonLabels,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{{
						Name: "scripts",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: scriptsConfigMapName,
								},
								DefaultMode: &scriptMode,
							},
						},
					}},
					Containers: []corev1.Container{
						{
							Name:    "mongodb",
							Image:   "docker.io/bitnami/mongodb" + d.Version,
							Command: []string{scriptPath},
							Ports: []corev1.ContainerPort{
								{
									Name:          "mongodb",
									ContainerPort: 27017,
								},
							},
							Env:       mongoEnvVars(c.name, serviceName),
							Resources: corev1.ResourceRequirements{},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "datadir",
									MountPath: "/bitnami/mongodb",
								},
								{
									Name:      "scripts",
									MountPath: scriptPath,
									SubPath:   "setup.sh",
								},
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{Exec: &corev1.ExecAction{Command: []string{
									"bash", "-ec", readinessCommand,
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
								RunAsUser:    &user,
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
			Name: c.name,
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		})
	} else {
		pvcTemplate := util.MakeVolumeClaimTemplate(*storageSpec)
		if pvcTemplate.Name == "" {
			pvcTemplate.Name = c.name
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

func (c *config) MakeScriptsConfigmap() *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name + "-scripts",
			Namespace: c.namespace,
		},
		Data: map[string]string{
			"setup.sh": fmt.Sprintf(`
#!/bin/bash

. /opt/bitnami/scripts/mongodb-env.sh

echo "Advertised Hostname: $MONGODB_ADVERTISED_HOSTNAME"

if [[ "$MY_POD_NAME" = "%v-0" ]]; then
    echo "Pod name matches initial primary pod name, configuring node as a primary"
    export MONGODB_REPLICA_SET_MODE="primary"
else
    echo "Pod name doesn't match initial primary pod name, configuring node as a secondary"
    export MONGODB_REPLICA_SET_MODE="secondary"
    export MONGODB_INITIAL_PRIMARY_ROOT_PASSWORD="$MONGODB_ROOT_PASSWORD"
    export MONGODB_INITIAL_PRIMARY_PORT_NUMBER="$MONGODB_PORT_NUMBER"
    export MONGODB_ROOT_PASSWORD="" MONGODB_USERNAME="" MONGODB_DATABASE="" MONGODB_PASSWORD=""
    export MONGODB_ROOT_PASSWORD_FILE="" MONGODB_USERNAME_FILE="" MONGODB_DATABASE_FILE="" MONGODB_PASSWORD_FILE=""
fi

exec /opt/bitnami/scripts/mongodb/entrypoint.sh /opt/bitnami/scripts/mongodb/run.sh`, c.name)},
	}
	return cm
}

func (c *config) MakeSecret() *corev1.Secret {
	rootPassword := []byte(util.RandomString(15))
	password := []byte(util.RandomString(15))
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c.name + "-auth",
			Namespace: c.namespace,
			Labels:    c.commonLabels,
		},
		Data: map[string][]byte{
			"root-password":  rootPassword,
			"password":       password,
			"user":           []byte("rocketchat"),
			"replicaset-key": []byte(util.RandomString(15)),
			"oplog-uri":      []byte(fmt.Sprintf("mongodb://root:%v@%v:27017/local?replicaSet=rs0&authSource=admin", rootPassword, c.name+"-service")),
			"uri":            []byte(fmt.Sprintf("mongodb://rocketchat:%v@%v:27017/rocketchat", password, c.name+"-service")),
		},
	}
	return secret
}

func mongoEnvVars(databaseName string, serviceName string) []corev1.EnvVar {
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
			Value: serviceName,
		},
		{
			Name:  "MONGODB_INITIAL_PRIMARY_HOST",
			Value: databaseName + "-0.$(K8S_SERVICE_NAME).$(MY_POD_NAMESPACE).svc.cluster.local",
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
					LocalObjectReference: corev1.LocalObjectReference{Name: databaseName + "-auth"},
					Key:                  "username",
				},
			},
		}, {
			Name: "MONGODB_PASSWORD",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: databaseName + "-auth"},
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
					LocalObjectReference: corev1.LocalObjectReference{Name: databaseName + "-auth"},
					Key:                  "root-password",
				},
			},
		}, {
			Name: "MONGODB_REPLICA_SET_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: databaseName + "-auth"},
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
