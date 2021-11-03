package model

import (
	"fmt"

	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/hown3d/chat-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func MongodbSecret(r *chatv1alpha1.Rocket) *corev1.Secret {
	rootPassword := []byte(util.RandomString(15))
	password := []byte(util.RandomString(15))
	mongodbService := r.Name + "-mongodb-service"
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Name + "-auth",
			Namespace: r.Namespace,
			Labels:    r.Labels,
		},
		Data: map[string][]byte{
			"root-password":  rootPassword,
			"password":       password,
			"user":           []byte("rocketchat"),
			"replicaset-key": []byte(util.RandomString(15)),
			"oplog-uri":      []byte(fmt.Sprintf("mongodb://root:%v@%v:27017/local?replicaSet=rs0&authSource=admin", rootPassword, mongodbService)),
			"uri":            []byte(fmt.Sprintf("mongodb://rocketchat:%v@%v:27017/rocketchat", password, mongodbService)),
		},
	}
	return secret
}
func MongodbSecretSelector(r *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      r.Name + "-mongodb-service",
		Namespace: r.Namespace,
	}
}
