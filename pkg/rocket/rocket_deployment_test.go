package rocket

import (
	chatv1alpha1 "github.com/hown3d/chat-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var (
	rocket = &chatv1alpha1.Rocket{
		ObjectMeta: metav1.ObjectMeta{Name: "test-name", Namespace: "test-namespace"},
		Spec:       &chatv1alpha1.RocketSpec{},
		Status:     &chatv1alpha1.RocketStatus{},
	}
	c = NewConfig(rocket)
)

func TestCreateOrUpdateRocketDeployment_MongoEnvVars(t *testing.T) {

	type args struct {
		mongoEnv map[string]corev1.EnvVarSource
	}
	tests := []struct {
		name string
		args args
		want []corev1.EnvVar
	}{
		{name: "empty-mongo-env", args: args{mongoEnv: map[string]corev1.EnvVarSource{}}, want: []corev1.EnvVar(nil)},
		{name: "default-mongo-env", args: args{mongoEnv: map[string]corev1.EnvVarSource{"MONGO_USERNAME": {
			SecretKeyRef: &corev1.SecretKeySelector{Key: "test-secret"},
		}}},
			want: []corev1.EnvVar{
				{
					Name:      "MONGO_USERNAME",
					ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{Key: "test-secret"}},
				},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.MakeDeployment(tt.args.mongoEnv)
			assert.Equal(t, tt.want, got.Spec.Template.Spec.Containers[0].Env)
		})
	}
}

func TestCreateOrUpdateRocketSecret(t *testing.T) {
	tests := []struct {
		name string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}

func TestCreateOrUpdateRocketService_Names(t *testing.T) {
	tests := []struct {
		name string
		want *corev1.Service
	}{
		{name: "test-names",
			want: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{Name: "test-name-service", Namespace: "test-namespace"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.MakeService()
			assert.Equal(t, tt.want.ObjectMeta.Name, got.ObjectMeta.Name)
			assert.Equal(t, tt.want.ObjectMeta.Namespace, got.ObjectMeta.Namespace)
		})
	}
}
