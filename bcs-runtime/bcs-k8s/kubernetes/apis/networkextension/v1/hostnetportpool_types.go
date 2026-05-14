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

// HostNetPortPoolSpec defines the desired state of HostNetPortPool
type HostNetPortPoolSpec struct {
	// StartPort is the start port of the port pool
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	StartPort uint32 `json:"startPort"`

	// EndPort is the end port of the port pool
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	EndPort uint32 `json:"endPort"`

	// SegmentLength is the minimum allocated port count for each Pod
	// Controller allocate port count by ceil(PortCount / SegmentLength)
	// e,g. if SegmentLength is 10, and PortCount is 25, Controller will allocate 3 segments for each Pod
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	SegmentLength uint32 `json:"segmentLength"`
}

// HostNetPortPoolStatus defines the observed state of HostNetPortPool
type HostNetPortPoolStatus struct {
	// NodeAllocations is the allocation status of the node host port pool
	NodeAllocations []*NodeHostNetPortPoolStatus `json:"nodeAllocations,omitempty"`
	// Status is the status of the host port pool (Ready, NotReady)
	Status string `json:"status,omitempty"`
}

// NodeHostNetPortPoolStatus defines the observed state of HostNetPortPool on a certain Node
type NodeHostNetPortPoolStatus struct {
	// NodeName is the name of the node
	NodeName string `json:"nodeName"`
	// AllocatedCount is how many segments have been allocated on the Node
	AllocatedCount int `json:"allocatedCount"`
	// TotalSegments is the total segments of the node host port pool
	TotalSegments int `json:"totalSegments"`
}

// +kubebuilder:object:root=true
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.status`

// HostNetPortPool is the Schema for the hostnetportpools API
type HostNetPortPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HostNetPortPoolSpec   `json:"spec,omitempty"`
	Status HostNetPortPoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// HostNetPortPoolList contains a list of HostNetPortPool
type HostNetPortPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HostNetPortPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HostNetPortPool{}, &HostNetPortPoolList{})
}
