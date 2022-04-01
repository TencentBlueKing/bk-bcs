/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ArgocdPluginSpec defines the desired state of ArgocdPlugin
type ArgocdPluginSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	NickName string              `json:"nickName,omitempty" protobuf:"bytes,1,opt,name=nickName"`
	Project  string              `json:"project,omitempty" protobuf:"bytes,2,opt,name=project"`
	Type     string              `json:"type,omitempty" protobuf:"bytes,3,opt,name=type"`
	Service  ArgocdPluginService `json:"service,omitempty" protobuf:"bytes,4,opt,name=service"`
	Image    string              `json:"image,omitempty" protobuf:"bytes,5,opt,name=image"`
}

// ArgocdPluginService defines the service information of the plugins when type is "service"
type ArgocdPluginService struct {
	Protocol string            `json:"protocol,omitempty" protobuf:"bytes,1,opt,name=protocol"`
	Address  string            `json:"address,omitempty" protobuf:"bytes,2,opt,name=address"`
	Headers  map[string]string `json:"headers,omitempty" protobuf:"bytes,3,opt,name=headers"`
}

// ArgocdPluginStatus defines the observed state of ArgocdPlugin
type ArgocdPluginStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+genclient
//+genclient:noStatus
//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ArgocdPlugin is the Schema for the argocdplugins API
type ArgocdPlugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec   ArgocdPluginSpec   `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
	Status ArgocdPluginStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

//+k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
//+kubebuilder:object:root=true

// ArgocdPluginList contains a list of ArgocdPlugin
type ArgocdPluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []ArgocdPlugin `json:"items" protobuf:"bytes,2,rep,name=items"`
}
