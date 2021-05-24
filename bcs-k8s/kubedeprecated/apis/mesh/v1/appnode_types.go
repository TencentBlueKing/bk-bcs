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

// AppNodeSpec defines the desired state of AppNode
type AppNodeSpec struct {
	// node key, pod instance name/taskgroup name
	// +kubebuilder:validation:MaxLength=200
	// +kubebuilder:validation:MinLength=1
	Index string `json:"index"`
	// node version, like v1, v1.1, v12.01.1, come from env[BCS_DISCOVERY_VERSION]
	Version string `json:"version,omitempty"`
	// node weight, it's a Relative value
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=0
	Weight uint `json:"weight,omitempty"`
	// app node network mode
	Network string `json:"network,omitempty"`
	// node ip address
	NodeIP string `json:"nodeIP"`
	// proxy ip address for this node
	ProxyIP string `json:"proxyIP,omitempty"`
	// port info for container
	Ports []NodePort `json:"ports,omitempty"`
}

//NodePort port info for one node of service
type NodePort struct {
	// name for port, must equal to one service port
	Name string `json:"name"`
	// protocol for this port
	Protocol string `json:"protocol"`
	// node port
	NodePort int `json:"nodeport"`
	// proxy port if exists
	ProxyPort int `json:"proxyport,omitempty"`
}

// AppNodeStatus defines the observed state of AppNode
type AppNodeStatus struct {
	// last updated timestamp
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// AppNode is the Schema for the appnodes API
type AppNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppNodeSpec   `json:"spec,omitempty"`
	Status AppNodeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// AppNodeList contains a list of AppNode
type AppNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppNode{}, &AppNodeList{})
}
