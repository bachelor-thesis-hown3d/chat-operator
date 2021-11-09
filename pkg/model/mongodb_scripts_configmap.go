package model

import (
	"fmt"
	"reflect"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MongodbScriptsConfigmapCreator struct{}

// Name returns the ressource action of the MongodbAuthSecretCreator
func (c *MongodbScriptsConfigmapCreator) Name() string {
	return "Mongodb Script Configmap"
}
func (c *MongodbScriptsConfigmapCreator) CreateResource(r *chatv1alpha1.Rocket) client.Object {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name + MongodbScriptsConfigmapSuffix,
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		Data: map[string]string{
			"setup.sh": fmt.Sprintf(
				`#!/bin/bash

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

exec /opt/bitnami/scripts/mongodb/entrypoint.sh /opt/bitnami/scripts/mongodb/run.sh`, r.Name+"-mongodb")},
	}
	return cm
}
func (c *MongodbScriptsConfigmapCreator) Update(rocket *chatv1alpha1.Rocket, cur client.Object) (client.Object, bool) {
	update := false
	cm := cur.(*corev1.ConfigMap)
	if !reflect.DeepEqual(cm.Labels, rocket.Labels) {
		cm.Labels = rocket.Labels
		update = true
	}
	return cm, update
}

func (c *MongodbScriptsConfigmapCreator) Selector(r *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      r.Name + MongodbScriptsConfigmapSuffix,
		Namespace: r.Namespace,
	}
}
