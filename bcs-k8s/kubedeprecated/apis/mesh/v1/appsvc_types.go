/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppSvcSpec defines the desired state of AppSvc
type AppSvcSpec struct {
	Selector map[string]string `json:"selector"`
	// service type, ClusterIP, Intergration or Empty
	Type string `json:"type,omitempty"`
	// service version
	Version string `json:"version,omitempty"`
	// frontend represents service ip address, use for proxy or intergate
	Frontend []string `json:"frontend,omitempty"`
	// domain alias
	Alias string `json:"alias,omitempty"`
	// use for wan export
	WANIP        []string      `json:"wanip,omitempty"`
	ServicePorts []ServicePort `json:"ports"` //BcsService.Ports
}

//ServicePort port definition for application
type ServicePort struct {
	// name for service port
	// +kubebuilder:validation:MaxLength=100
	// +kubebuilder:validation:MinLength=3
	Name string `json:"name"`
	// protocol for service port
	Protocol string `json:"protocol"`
	// domain value for http proxy
	// +kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:MinLength=3
	Domain string `json:"domain,omitempty"`
	// http url path
	Path string `json:"path,omitempty"`
	// service port for all AppNode, ServicePort.Name == AppNode.Ports[i].Name
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	ServicePort int `json:"serviceport"`
	// proxy port for this Service Port if exist
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=0
	ProxyPort int `json:"proxyport,omitempty"`
}

// AppSvcStatus defines the observed state of AppSvc
type AppSvcStatus struct {
	// last updated timestamp
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// AppSvc is the Schema for the appsvcs API
type AppSvc struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppSvcSpec   `json:"spec,omitempty"`
	Status AppSvcStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// AppSvcList contains a list of AppSvc
type AppSvcList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppSvc `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppSvc{}, &AppSvcList{})
}
