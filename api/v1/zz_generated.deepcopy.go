//go:build !ignore_autogenerated

/*
MIT License

Copyright (c) 2023 Dev Ops Squad

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatabaseAccount) DeepCopyInto(out *DatabaseAccount) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatabaseAccount.
func (in *DatabaseAccount) DeepCopy() *DatabaseAccount {
	if in == nil {
		return nil
	}
	out := new(DatabaseAccount)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DatabaseAccount) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatabaseAccountControllerConfig) DeepCopyInto(out *DatabaseAccountControllerConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.ControllerManagerConfigurationSpec.DeepCopyInto(&out.ControllerManagerConfigurationSpec)
	out.Debug = in.Debug
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatabaseAccountControllerConfig.
func (in *DatabaseAccountControllerConfig) DeepCopy() *DatabaseAccountControllerConfig {
	if in == nil {
		return nil
	}
	out := new(DatabaseAccountControllerConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DatabaseAccountControllerConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatabaseAccountControllerConfigDebug) DeepCopyInto(out *DatabaseAccountControllerConfigDebug) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatabaseAccountControllerConfigDebug.
func (in *DatabaseAccountControllerConfigDebug) DeepCopy() *DatabaseAccountControllerConfigDebug {
	if in == nil {
		return nil
	}
	out := new(DatabaseAccountControllerConfigDebug)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatabaseAccountControllerConfigList) DeepCopyInto(out *DatabaseAccountControllerConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DatabaseAccountControllerConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatabaseAccountControllerConfigList.
func (in *DatabaseAccountControllerConfigList) DeepCopy() *DatabaseAccountControllerConfigList {
	if in == nil {
		return nil
	}
	out := new(DatabaseAccountControllerConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DatabaseAccountControllerConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatabaseAccountList) DeepCopyInto(out *DatabaseAccountList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DatabaseAccount, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatabaseAccountList.
func (in *DatabaseAccountList) DeepCopy() *DatabaseAccountList {
	if in == nil {
		return nil
	}
	out := new(DatabaseAccountList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DatabaseAccountList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatabaseAccountSpec) DeepCopyInto(out *DatabaseAccountSpec) {
	*out = *in
	in.SecretTemplate.DeepCopyInto(&out.SecretTemplate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatabaseAccountSpec.
func (in *DatabaseAccountSpec) DeepCopy() *DatabaseAccountSpec {
	if in == nil {
		return nil
	}
	out := new(DatabaseAccountSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatabaseAccountSpecSecretTemplate) DeepCopyInto(out *DatabaseAccountSpecSecretTemplate) {
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

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatabaseAccountSpecSecretTemplate.
func (in *DatabaseAccountSpecSecretTemplate) DeepCopy() *DatabaseAccountSpecSecretTemplate {
	if in == nil {
		return nil
	}
	out := new(DatabaseAccountSpecSecretTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DatabaseAccountStatus) DeepCopyInto(out *DatabaseAccountStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DatabaseAccountStatus.
func (in *DatabaseAccountStatus) DeepCopy() *DatabaseAccountStatus {
	if in == nil {
		return nil
	}
	out := new(DatabaseAccountStatus)
	in.DeepCopyInto(out)
	return out
}
