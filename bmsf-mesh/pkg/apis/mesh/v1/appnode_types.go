/*
Copyright (C) 2019 The BlueKing Authors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AppNodeSpec defines the desired state of AppNode
type AppNodeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
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
	// +kubebuilder:validation:MaxLength=100
	// +kubebuilder:validation:MinLength=2
	Name string `json:"name"`
	// protocol for this port
	// +kubebuilder:validation:Enum=tcp,udp,http
	Protocol string `json:"protocol"`
	// node port
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	NodePort int `json:"nodeport"`
	// proxy port if exists
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=-10
	ProxyPort int `json:"proxyport,omitempty"`
}

// AppNodeStatus defines the observed state of AppNode
type AppNodeStatus struct {
	// last updated timestamp
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppNode is the Schema for the appnodes API
// +k8s:openapi-gen=true
type AppNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppNodeSpec   `json:"spec,omitempty"`
	Status AppNodeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// AppNodeList contains a list of AppNode
type AppNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppNode{}, &AppNodeList{})
}
