/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StatusPhase string

var (
	NoPhase           StatusPhase
	PhaseReconciling  StatusPhase = "reconciling"
	PhaseFailing      StatusPhase = "failing"
	PhaseInitialising StatusPhase = "initialising"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type RocketDatabase struct {
	// Version of the Mongodb Containers, matches a Tag from https://hub.docker.com/r/bitnami/mongodb repository
	// +optional
	Version string `json:"version,omitempty"`
	// Replicas of Mongodb Instance
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// StorageSpec embedds a PersistentVolumeClaim Template
	// (+)kubebuilder:validation:EmbeddedResource
	StorageSpec *EmbeddedPersistentVolumeClaim `json:"storageSpec,omitempty"`
}

// RocketAdminSpec contains the email and username of the administrator
type RocketAdminSpec struct {
	// Email is the email of the administrator
	Email string `json:"email,omitempty"`
	// Username is the Username of the administrator
	Username string `json:"username,omitempty"`
}

// EmbeddedPersistentVolumeClaim is an embedded version of k8s.io/api/core/corev1.PersistentVolumeClaim.
// It contains TypeMeta and a reduced ObjectMeta.
type EmbeddedPersistentVolumeClaim struct {
	metav1.TypeMeta `json:",inline"`

	// EmbeddedMetadata contains metadata relevant to an EmbeddedResource.
	EmbeddedObjectMetadata `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the desired characteristics of a volume requested by the user.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims
	// +optional
	Spec corev1.PersistentVolumeClaimSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status represents the current information/status of a persistent volume claim.
	// Read-only.
	// More info: https://kubernetes.io/docs/concepts/storage/persistent-volumes#persistentvolumeclaims
	// +optional
	Status corev1.PersistentVolumeClaimStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// EmbeddedObjectMetadata contains a subset of the fields included in k8s.io/apimachinery/pkg/apis/meta/corev1.ObjectMeta
// Only fields which are relevant to embedded resources are included.
type EmbeddedObjectMetadata struct {
	// Name must be unique within a namespace. Is required when creating resources, although
	// some resources may allow a client to request the generation of an appropriate name
	// automatically. Name is primarily intended for creation idempotence and configuration
	// definition.
	// Cannot be updated.
	// More info: http://kubernetes.io/docs/user-guide/identifiers#names
	// +optional
	Name string `json:"name,omitempty" protobuf:"bytes,1,opt,name=name"`

	// Map of string keys and values that can be used to organize and categorize
	// (scope and select) objects. May match selectors of replication controllers
	// and services.
	// More info: http://kubernetes.io/docs/user-guide/labels
	// +optional
	Labels map[string]string `json:"labels,omitempty" protobuf:"bytes,11,rep,name=labels"`

	// Annotations is an unstructured key value map stored with a resource that may be
	// set by external tools to store and retrieve arbitrary metadata. They are not
	// queryable and should be preserved when modifying objects.
	// More info: http://kubernetes.io/docs/user-guide/annotations
	// +optional
	Annotations map[string]string `json:"annotations,omitempty" protobuf:"bytes,12,rep,name=annotations"`
}

// RocketSpec defines the desired state of Rocket
type RocketSpec struct {
	// Replicas specifies how many Webserver Pods shall be created
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// Version specifies the Rocket.Chat Container Image Version
	// +optional
	Version string `json:"version,omitempty"`
	// AdminSpec specifies the admin username and email
	AdminSpec *RocketAdminSpec `json:"adminSpec,omitempty"`
	// Database contains the specification for the mongodb Database
	Database RocketDatabase `json:"database,omitempty"`
	// Hostname to use for the instance
	IngressSpec RocketIngressSpec `json:"ingressSpec,omitempty"`
}

type RocketIngressSpec struct {
	// Host is the hostname for ingress object
	Host string `json:"host,omitempty"`
	// Annotations to add to the ingress Object
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// RocketStatus defines the observed state of Rocket
type RocketStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Pods are the names of the Rocket.Chat Pods
	// +optional
	Pods []EmbeddedPod `json:"pods,omitempty"`
	// Current phase of the operator.
	Phase StatusPhase `json:"phase,omitempty"`
	// Human-readable message indicating details about current operator phase or error.
	Message string `json:"message,omitempty"`
	// True if all resources are in a ready state and all work is done.
	Ready bool `json:"ready,omitempty"`
	// External URL for accessing Rocket instance from outside the cluster.
	// +optional
	ExternalURL string `json:"externalURL,omitempty"`
}

// EmbeddedPod contains metadata and status of a pod
type EmbeddedPod struct {
	// Name of the Pod
	Name string `json:"name,omitempty"`
}

// Rocket is the Schema for the rockets API
// additional column for `kubectl get` output
//+kubebuilder:printcolumn:name="Webserver Version",type=string,JSONPath=`.spec.version`
//+kubebuilder:printcolumn:name="Database Version",type=string,JSONPath=`.spec.database.version`
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Ready",type=boolean,JSONPath=`.status.ready`
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type Rocket struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RocketSpec   `json:"spec,omitempty"`
	Status RocketStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RocketList contains a list of Rocket
type RocketList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rocket `json:"items,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Rocket{}, &RocketList{})
}
