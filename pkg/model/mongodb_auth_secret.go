package model

import (
	"fmt"

	chatv1alpha1 "github.com/bachelor-thesis-hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	"github.com/bachelor-thesis-hown3d/chat-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MongodbAuthSecretCreator struct{}

// Name returns the ressource action of the MongodbAuthSecretCreator
func (c *MongodbAuthSecretCreator) Name() string {
	return "Mongodb Auth Secret"
}

func (c *MongodbAuthSecretCreator) CreateResource(rocket *chatv1alpha1.Rocket) client.Object {
	rootPassword := util.RandomString(25)
	password := util.RandomString(25)
	mongodbService := rocket.Name + MongodbServiceSuffix
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rocket.Name + MongodbAuthSecretSuffix,
			Namespace: rocket.Namespace,
			Labels:    rocket.Labels,
		},
		Data: map[string][]byte{
			"root-password":  []byte(rootPassword),
			"password":       []byte(password),
			"user":           []byte("rocketchat"),
			"replicaset-key": []byte(util.RandomString(25)),
			"oplog-uri":      []byte(fmt.Sprintf("mongodb://root:%v@%v:27017/local?replicaSet=rs0&authSource=admin", rootPassword, mongodbService)),
			"uri":            []byte(fmt.Sprintf("mongodb://rocketchat:%v@%v:27017/rocketchat?replicaSet=rs0&w=majority", password, mongodbService)),
		},
	}
	return secret
}
func (c *MongodbAuthSecretCreator) Selector(rocket *chatv1alpha1.Rocket) client.ObjectKey {
	return client.ObjectKey{
		Name:      rocket.Name + MongodbAuthSecretSuffix,
		Namespace: rocket.Namespace,
	}
}
func (c *MongodbAuthSecretCreator) Update(rocket *chatv1alpha1.Rocket, cur client.Object) (client.Object, bool) {
	// never update auth secret!
	return cur, false
}
