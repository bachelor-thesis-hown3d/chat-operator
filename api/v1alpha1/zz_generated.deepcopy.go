//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EmbeddedObjectMetadata) DeepCopyInto(out *EmbeddedObjectMetadata) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EmbeddedObjectMetadata.
func (in *EmbeddedObjectMetadata) DeepCopy() *EmbeddedObjectMetadata {
	if in == nil {
		return nil
	}
	out := new(EmbeddedObjectMetadata)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EmbeddedPersistentVolumeClaim) DeepCopyInto(out *EmbeddedPersistentVolumeClaim) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.EmbeddedObjectMetadata.DeepCopyInto(&out.EmbeddedObjectMetadata)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EmbeddedPersistentVolumeClaim.
func (in *EmbeddedPersistentVolumeClaim) DeepCopy() *EmbeddedPersistentVolumeClaim {
	if in == nil {
		return nil
	}
	out := new(EmbeddedPersistentVolumeClaim)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EmbeddedPod) DeepCopyInto(out *EmbeddedPod) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EmbeddedPod.
func (in *EmbeddedPod) DeepCopy() *EmbeddedPod {
	if in == nil {
		return nil
	}
	out := new(EmbeddedPod)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Rocket) DeepCopyInto(out *Rocket) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Rocket.
func (in *Rocket) DeepCopy() *Rocket {
	if in == nil {
		return nil
	}
	out := new(Rocket)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Rocket) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RocketAdminSpec) DeepCopyInto(out *RocketAdminSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RocketAdminSpec.
func (in *RocketAdminSpec) DeepCopy() *RocketAdminSpec {
	if in == nil {
		return nil
	}
	out := new(RocketAdminSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RocketDatabase) DeepCopyInto(out *RocketDatabase) {
	*out = *in
	if in.StorageSpec != nil {
		in, out := &in.StorageSpec, &out.StorageSpec
		*out = new(EmbeddedPersistentVolumeClaim)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RocketDatabase.
func (in *RocketDatabase) DeepCopy() *RocketDatabase {
	if in == nil {
		return nil
	}
	out := new(RocketDatabase)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RocketList) DeepCopyInto(out *RocketList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Rocket, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RocketList.
func (in *RocketList) DeepCopy() *RocketList {
	if in == nil {
		return nil
	}
	out := new(RocketList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RocketList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RocketSpec) DeepCopyInto(out *RocketSpec) {
	*out = *in
	if in.AdminSpec != nil {
		in, out := &in.AdminSpec, &out.AdminSpec
		*out = new(RocketAdminSpec)
		**out = **in
	}
	in.Database.DeepCopyInto(&out.Database)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RocketSpec.
func (in *RocketSpec) DeepCopy() *RocketSpec {
	if in == nil {
		return nil
	}
	out := new(RocketSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RocketStatus) DeepCopyInto(out *RocketStatus) {
	*out = *in
	if in.Pods != nil {
		in, out := &in.Pods, &out.Pods
		*out = make([]EmbeddedPod, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RocketStatus.
func (in *RocketStatus) DeepCopy() *RocketStatus {
	if in == nil {
		return nil
	}
	out := new(RocketStatus)
	in.DeepCopyInto(out)
	return out
}
