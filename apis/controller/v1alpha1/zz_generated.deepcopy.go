//go:build !ignore_autogenerated

//
// Copyright (c) 2019-2025 Red Hat, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	v1 "k8s.io/api/core/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in Attributes) DeepCopyInto(out *Attributes) {
	{
		in := &in
		*out = make(Attributes, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Attributes.
func (in Attributes) DeepCopy() Attributes {
	if in == nil {
		return nil
	}
	out := new(Attributes)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ConfigmapReference) DeepCopyInto(out *ConfigmapReference) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ConfigmapReference.
func (in *ConfigmapReference) DeepCopy() *ConfigmapReference {
	if in == nil {
		return nil
	}
	out := new(ConfigmapReference)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DevWorkspaceOperatorConfig) DeepCopyInto(out *DevWorkspaceOperatorConfig) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	if in.Config != nil {
		in, out := &in.Config, &out.Config
		*out = new(OperatorConfiguration)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DevWorkspaceOperatorConfig.
func (in *DevWorkspaceOperatorConfig) DeepCopy() *DevWorkspaceOperatorConfig {
	if in == nil {
		return nil
	}
	out := new(DevWorkspaceOperatorConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DevWorkspaceOperatorConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DevWorkspaceOperatorConfigList) DeepCopyInto(out *DevWorkspaceOperatorConfigList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DevWorkspaceOperatorConfig, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DevWorkspaceOperatorConfigList.
func (in *DevWorkspaceOperatorConfigList) DeepCopy() *DevWorkspaceOperatorConfigList {
	if in == nil {
		return nil
	}
	out := new(DevWorkspaceOperatorConfigList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DevWorkspaceOperatorConfigList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DevWorkspaceRouting) DeepCopyInto(out *DevWorkspaceRouting) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DevWorkspaceRouting.
func (in *DevWorkspaceRouting) DeepCopy() *DevWorkspaceRouting {
	if in == nil {
		return nil
	}
	out := new(DevWorkspaceRouting)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DevWorkspaceRouting) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DevWorkspaceRoutingList) DeepCopyInto(out *DevWorkspaceRoutingList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]DevWorkspaceRouting, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DevWorkspaceRoutingList.
func (in *DevWorkspaceRoutingList) DeepCopy() *DevWorkspaceRoutingList {
	if in == nil {
		return nil
	}
	out := new(DevWorkspaceRoutingList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *DevWorkspaceRoutingList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DevWorkspaceRoutingSpec) DeepCopyInto(out *DevWorkspaceRoutingSpec) {
	*out = *in
	if in.Endpoints != nil {
		in, out := &in.Endpoints, &out.Endpoints
		*out = make(map[string]EndpointList, len(*in))
		for key, val := range *in {
			var outVal []Endpoint
			if val == nil {
				(*out)[key] = nil
			} else {
				inVal := (*in)[key]
				in, out := &inVal, &outVal
				*out = make(EndpointList, len(*in))
				for i := range *in {
					(*in)[i].DeepCopyInto(&(*out)[i])
				}
			}
			(*out)[key] = outVal
		}
	}
	if in.PodSelector != nil {
		in, out := &in.PodSelector, &out.PodSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DevWorkspaceRoutingSpec.
func (in *DevWorkspaceRoutingSpec) DeepCopy() *DevWorkspaceRoutingSpec {
	if in == nil {
		return nil
	}
	out := new(DevWorkspaceRoutingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *DevWorkspaceRoutingStatus) DeepCopyInto(out *DevWorkspaceRoutingStatus) {
	*out = *in
	if in.PodAdditions != nil {
		in, out := &in.PodAdditions, &out.PodAdditions
		*out = new(PodAdditions)
		(*in).DeepCopyInto(*out)
	}
	if in.ExposedEndpoints != nil {
		in, out := &in.ExposedEndpoints, &out.ExposedEndpoints
		*out = make(map[string]ExposedEndpointList, len(*in))
		for key, val := range *in {
			var outVal []ExposedEndpoint
			if val == nil {
				(*out)[key] = nil
			} else {
				inVal := (*in)[key]
				in, out := &inVal, &outVal
				*out = make(ExposedEndpointList, len(*in))
				for i := range *in {
					(*in)[i].DeepCopyInto(&(*out)[i])
				}
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new DevWorkspaceRoutingStatus.
func (in *DevWorkspaceRoutingStatus) DeepCopy() *DevWorkspaceRoutingStatus {
	if in == nil {
		return nil
	}
	out := new(DevWorkspaceRoutingStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Endpoint) DeepCopyInto(out *Endpoint) {
	*out = *in
	if in.Attributes != nil {
		in, out := &in.Attributes, &out.Attributes
		*out = make(Attributes, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
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

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Endpoint.
func (in *Endpoint) DeepCopy() *Endpoint {
	if in == nil {
		return nil
	}
	out := new(Endpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in EndpointList) DeepCopyInto(out *EndpointList) {
	{
		in := &in
		*out = make(EndpointList, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EndpointList.
func (in EndpointList) DeepCopy() EndpointList {
	if in == nil {
		return nil
	}
	out := new(EndpointList)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExposedEndpoint) DeepCopyInto(out *ExposedEndpoint) {
	*out = *in
	if in.Attributes != nil {
		in, out := &in.Attributes, &out.Attributes
		*out = make(Attributes, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExposedEndpoint.
func (in *ExposedEndpoint) DeepCopy() *ExposedEndpoint {
	if in == nil {
		return nil
	}
	out := new(ExposedEndpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in ExposedEndpointList) DeepCopyInto(out *ExposedEndpointList) {
	{
		in := &in
		*out = make(ExposedEndpointList, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExposedEndpointList.
func (in ExposedEndpointList) DeepCopy() ExposedEndpointList {
	if in == nil {
		return nil
	}
	out := new(ExposedEndpointList)
	in.DeepCopyInto(out)
	return *out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *KeyNotFoundError) DeepCopyInto(out *KeyNotFoundError) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new KeyNotFoundError.
func (in *KeyNotFoundError) DeepCopy() *KeyNotFoundError {
	if in == nil {
		return nil
	}
	out := new(KeyNotFoundError)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OperatorConfiguration) DeepCopyInto(out *OperatorConfiguration) {
	*out = *in
	if in.Routing != nil {
		in, out := &in.Routing, &out.Routing
		*out = new(RoutingConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Workspace != nil {
		in, out := &in.Workspace, &out.Workspace
		*out = new(WorkspaceConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.Webhook != nil {
		in, out := &in.Webhook, &out.Webhook
		*out = new(WebhookConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.EnableExperimentalFeatures != nil {
		in, out := &in.EnableExperimentalFeatures, &out.EnableExperimentalFeatures
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OperatorConfiguration.
func (in *OperatorConfiguration) DeepCopy() *OperatorConfiguration {
	if in == nil {
		return nil
	}
	out := new(OperatorConfiguration)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PersistentHomeConfig) DeepCopyInto(out *PersistentHomeConfig) {
	*out = *in
	if in.Enabled != nil {
		in, out := &in.Enabled, &out.Enabled
		*out = new(bool)
		**out = **in
	}
	if in.DisableInitContainer != nil {
		in, out := &in.DisableInitContainer, &out.DisableInitContainer
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PersistentHomeConfig.
func (in *PersistentHomeConfig) DeepCopy() *PersistentHomeConfig {
	if in == nil {
		return nil
	}
	out := new(PersistentHomeConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PodAdditions) DeepCopyInto(out *PodAdditions) {
	*out = *in
	if in.Annotations != nil {
		in, out := &in.Annotations, &out.Annotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Containers != nil {
		in, out := &in.Containers, &out.Containers
		*out = make([]v1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.InitContainers != nil {
		in, out := &in.InitContainers, &out.InitContainers
		*out = make([]v1.Container, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Volumes != nil {
		in, out := &in.Volumes, &out.Volumes
		*out = make([]v1.Volume, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.VolumeMounts != nil {
		in, out := &in.VolumeMounts, &out.VolumeMounts
		*out = make([]v1.VolumeMount, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.PullSecrets != nil {
		in, out := &in.PullSecrets, &out.PullSecrets
		*out = make([]v1.LocalObjectReference, len(*in))
		copy(*out, *in)
	}
	if in.ServiceAccountAnnotations != nil {
		in, out := &in.ServiceAccountAnnotations, &out.ServiceAccountAnnotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PodAdditions.
func (in *PodAdditions) DeepCopy() *PodAdditions {
	if in == nil {
		return nil
	}
	out := new(PodAdditions)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ProjectCloneConfig) DeepCopyInto(out *ProjectCloneConfig) {
	*out = *in
	if in.Resources != nil {
		in, out := &in.Resources, &out.Resources
		*out = new(v1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.Env != nil {
		in, out := &in.Env, &out.Env
		*out = make([]v1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ProjectCloneConfig.
func (in *ProjectCloneConfig) DeepCopy() *ProjectCloneConfig {
	if in == nil {
		return nil
	}
	out := new(ProjectCloneConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Proxy) DeepCopyInto(out *Proxy) {
	*out = *in
	if in.HttpProxy != nil {
		in, out := &in.HttpProxy, &out.HttpProxy
		*out = new(string)
		**out = **in
	}
	if in.HttpsProxy != nil {
		in, out := &in.HttpsProxy, &out.HttpsProxy
		*out = new(string)
		**out = **in
	}
	if in.NoProxy != nil {
		in, out := &in.NoProxy, &out.NoProxy
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Proxy.
func (in *Proxy) DeepCopy() *Proxy {
	if in == nil {
		return nil
	}
	out := new(Proxy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RoutingConfig) DeepCopyInto(out *RoutingConfig) {
	*out = *in
	if in.ProxyConfig != nil {
		in, out := &in.ProxyConfig, &out.ProxyConfig
		*out = new(Proxy)
		(*in).DeepCopyInto(*out)
	}
	if in.TLSCertificateConfigmapRef != nil {
		in, out := &in.TLSCertificateConfigmapRef, &out.TLSCertificateConfigmapRef
		*out = new(ConfigmapReference)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RoutingConfig.
func (in *RoutingConfig) DeepCopy() *RoutingConfig {
	if in == nil {
		return nil
	}
	out := new(RoutingConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceAccountConfig) DeepCopyInto(out *ServiceAccountConfig) {
	*out = *in
	if in.DisableCreation != nil {
		in, out := &in.DisableCreation, &out.DisableCreation
		*out = new(bool)
		**out = **in
	}
	if in.ServiceAccountTokens != nil {
		in, out := &in.ServiceAccountTokens, &out.ServiceAccountTokens
		*out = make([]ServiceAccountToken, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceAccountConfig.
func (in *ServiceAccountConfig) DeepCopy() *ServiceAccountConfig {
	if in == nil {
		return nil
	}
	out := new(ServiceAccountConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceAccountToken) DeepCopyInto(out *ServiceAccountToken) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceAccountToken.
func (in *ServiceAccountToken) DeepCopy() *ServiceAccountToken {
	if in == nil {
		return nil
	}
	out := new(ServiceAccountToken)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *StorageSizes) DeepCopyInto(out *StorageSizes) {
	*out = *in
	if in.Common != nil {
		in, out := &in.Common, &out.Common
		x := (*in).DeepCopy()
		*out = &x
	}
	if in.PerWorkspace != nil {
		in, out := &in.PerWorkspace, &out.PerWorkspace
		x := (*in).DeepCopy()
		*out = &x
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StorageSizes.
func (in *StorageSizes) DeepCopy() *StorageSizes {
	if in == nil {
		return nil
	}
	out := new(StorageSizes)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WebhookConfig) DeepCopyInto(out *WebhookConfig) {
	*out = *in
	if in.NodeSelector != nil {
		in, out := &in.NodeSelector, &out.NodeSelector
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Tolerations != nil {
		in, out := &in.Tolerations, &out.Tolerations
		*out = make([]v1.Toleration, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Replicas != nil {
		in, out := &in.Replicas, &out.Replicas
		*out = new(int32)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WebhookConfig.
func (in *WebhookConfig) DeepCopy() *WebhookConfig {
	if in == nil {
		return nil
	}
	out := new(WebhookConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkspaceConfig) DeepCopyInto(out *WorkspaceConfig) {
	*out = *in
	if in.ProjectCloneConfig != nil {
		in, out := &in.ProjectCloneConfig, &out.ProjectCloneConfig
		*out = new(ProjectCloneConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.ServiceAccount != nil {
		in, out := &in.ServiceAccount, &out.ServiceAccount
		*out = new(ServiceAccountConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.StorageClassName != nil {
		in, out := &in.StorageClassName, &out.StorageClassName
		*out = new(string)
		**out = **in
	}
	if in.DefaultStorageSize != nil {
		in, out := &in.DefaultStorageSize, &out.DefaultStorageSize
		*out = new(StorageSizes)
		(*in).DeepCopyInto(*out)
	}
	if in.PersistUserHome != nil {
		in, out := &in.PersistUserHome, &out.PersistUserHome
		*out = new(PersistentHomeConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.IgnoredUnrecoverableEvents != nil {
		in, out := &in.IgnoredUnrecoverableEvents, &out.IgnoredUnrecoverableEvents
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.CleanupOnStop != nil {
		in, out := &in.CleanupOnStop, &out.CleanupOnStop
		*out = new(bool)
		**out = **in
	}
	if in.PodSecurityContext != nil {
		in, out := &in.PodSecurityContext, &out.PodSecurityContext
		*out = new(v1.PodSecurityContext)
		(*in).DeepCopyInto(*out)
	}
	if in.ContainerSecurityContext != nil {
		in, out := &in.ContainerSecurityContext, &out.ContainerSecurityContext
		*out = new(v1.SecurityContext)
		(*in).DeepCopyInto(*out)
	}
	if in.DefaultTemplate != nil {
		in, out := &in.DefaultTemplate, &out.DefaultTemplate
		*out = new(v1alpha2.DevWorkspaceTemplateSpecContent)
		(*in).DeepCopyInto(*out)
	}
	if in.DefaultContainerResources != nil {
		in, out := &in.DefaultContainerResources, &out.DefaultContainerResources
		*out = new(v1.ResourceRequirements)
		(*in).DeepCopyInto(*out)
	}
	if in.PodAnnotations != nil {
		in, out := &in.PodAnnotations, &out.PodAnnotations
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.RuntimeClassName != nil {
		in, out := &in.RuntimeClassName, &out.RuntimeClassName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkspaceConfig.
func (in *WorkspaceConfig) DeepCopy() *WorkspaceConfig {
	if in == nil {
		return nil
	}
	out := new(WorkspaceConfig)
	in.DeepCopyInto(out)
	return out
}
