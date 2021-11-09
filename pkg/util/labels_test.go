package util

import (
	"reflect"
	"testing"

	"github.com/hown3d/chat-operator/api/chat.accso.de/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMergeLabels(t *testing.T) {
	type args struct {
		base    map[string]string
		toMerge map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "two nil maps",
			args: args{
				base:    nil,
				toMerge: nil,
			},
		},
		{
			name: "toMerge nil map",
			args: args{
				base:    map[string]string{"test": "test"},
				toMerge: nil,
			},
			want: map[string]string{"test": "test"},
		},
		{
			name: "base nil map",
			args: args{
				toMerge: map[string]string{"test": "test"},
				base:    nil,
			},
			want: map[string]string{"test": "test"},
		},
		{
			name: "2 non empty maps",
			args: args{
				toMerge: map[string]string{"foo": "bar"},
				base:    map[string]string{"detlef": "desoost"},
			},
			want: map[string]string{"foo": "bar", "detlef": "desoost"},
		},
		{
			name: "two equal maps",
			args: args{
				toMerge: map[string]string{"foo": "bar"},
				base:    map[string]string{"foo": "bar"},
			},
			want: map[string]string{"foo": "bar"},
		},
		{
			name: "equal keys, different values",
			args: args{
				base:    map[string]string{"foo": "detlef"},
				toMerge: map[string]string{"foo": "bar"},
			},
			want: map[string]string{"foo": "bar"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeLabels(tt.args.base, tt.args.toMerge); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeLabels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasDefaultLabels(t *testing.T) {
	rocketName := "test"
	type args struct {
		rocket *v1alpha1.Rocket
	}
	tests := []struct {
		name       string
		args       args
		wantExists bool
	}{
		{
			name: "default-labels-exist-test",
			args: args{
				rocket: &v1alpha1.Rocket{ObjectMeta: v1.ObjectMeta{
					Name:   rocketName,
					Labels: DefaultLabels(rocketName),
				}},
			},
			wantExists: true,
		},
		{
			name: "empty-labels-test",
			args: args{
				rocket: &v1alpha1.Rocket{ObjectMeta: v1.ObjectMeta{
					Name: rocketName,
				}},
			},
			wantExists: false,
		},
		{
			name: "subset-default-labels-test",
			args: args{
				rocket: &v1alpha1.Rocket{ObjectMeta: v1.ObjectMeta{
					Name:   rocketName,
					Labels: MergeLabels(map[string]string{"test": "test"}, DefaultLabels(rocketName)),
				}},
			},
			wantExists: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotExists := HasDefaultLabels(tt.args.rocket); gotExists != tt.wantExists {
				t.Errorf("HasDefaultLabels() = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}
