package model

import (
	"fmt"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MongodbScriptsConfigmap(r *chatv1alpha1.Rocket) *corev1.ConfigMap {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name + "mongodb-scripts",
			Namespace: r.Namespace,
			Labels:    r.Labels,
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

exec /opt/bitnami/scripts/mongodb/entrypoint.sh /opt/bitnami/scripts/mongodb/run.sh`, r.Name+"mongodb")},
	}
	return cm
}
func MongodbConfigmapSelector(r *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      r.Name + "-mongodb-scripts",
		Namespace: r.Namespace,
	}
}
