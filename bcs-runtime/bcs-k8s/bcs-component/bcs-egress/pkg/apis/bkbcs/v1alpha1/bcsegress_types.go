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
 */

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ControllerRef reference for egress controller
type ControllerRef struct {
	// +kubebuilder:default=bcs-system
	Namespace string `json:"namespace"`
	// +kubebuilder:default=egress-controller
	Name string `json:"name"`
}

// HTTP http egress definition
type HTTP struct {
	// Name for http management
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// Host for domain, use for acl
	// +kubebuilder:validation:MinLength=4
	Host string `json:"host"`
	// Destination port for remote host
	// +kubebuilder:default=80
	DestPort uint `json:"destport"`
}

// TCP tcp egress definition
type TCP struct {
	// name for tcp management
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// Domain for destination, domain first
	Domain string `json:"domain"`
	// iplist(split by comma)
	IPs string `json:"ips"`
	// source & dest port use for tcp network flow control
	// +kubebuilder:validation:Mininm=1024
	SourcePort uint `json:"sourceport"`
	// +kubebuilder:validation:Mininm=1
	DestPort uint `json:"destport"`
	// algorithm for specified IP list
	// +kubebuilder:validation:Enum=roundrobin;least_conn;hash
	// +kubebuilder:default=roundrobin
	Algorithm string `json:"algorithm"`
}

// BCSEgressSpec defines the desired state of BCSEgress
type BCSEgressSpec struct {
	// Controller can be empty, we use egress-controller.bcs-system for default
	Controller ControllerRef `json:"controller"`
	HTTPS      []HTTP        `json:"https"`
	TCPS       []TCP         `json:"tcps"`
}

const (
	// EgressStatePending pending state when egress created
	EgressStatePending = "Pending"
	// EgressStateError when rule error or controller error
	EgressStateError = "Error"
	// EgressStateUnknown unknown when controller lost
	EgressStateUnknown = "Unknown"
	// EgressStateSynced all working correct
	EgressStateSynced = "Synced"
)

// BCSEgressStatus defines the observed state of BCSEgress
type BCSEgressStatus struct {
	// State reference EgressState above
	// +kubebuilder:default=Pending
	State       string `json:"state"`
	HTTPActives uint   `json:"httpActives"`
	TCPActives  uint   `json:"tcpActives"`
	// Reason when some error happened
	Reason string `json:"reason"`
	// all egress sync timestamp
	SyncedAt metav1.Time `json:"syncedat"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BCSEgress is the Schema for the bcsegresses API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=bcsegresses,scope=Namespaced
// +kubebuilder:printcolumn:name="Namespace",type="string",JSONPath=".metadata.namespace"
// +kubebuilder:printcolumn:name="Name",type="string",JSONPath=".metadata.name"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="HTTPActives",type="uint",JSONPath=".status.httpActives"
// +kubebuilder:printcolumn:name="TCPActives",type="uint",JSONPath=".status.tcpActives"
// +kubebuilder:printcolumn:name="SyncedAt",type="date",JSONPath=".status.syncedat"
type BCSEgress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BCSEgressSpec   `json:"spec,omitempty"`
	Status BCSEgressStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BCSEgressList contains a list of BCSEgress
type BCSEgressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BCSEgress `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BCSEgress{}, &BCSEgressList{})
}
