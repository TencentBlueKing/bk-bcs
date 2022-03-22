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
	"fmt"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PortPoolItem item of port pool
type PortPoolItem struct {
	// +kubebuilder:validation:MaxLength=128
	// +kubebuilder:validation:MinLength=1
	ItemName        string   `json:"itemName"`
	LoadBalancerIDs []string `json:"loadBalancerIDs"`
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	StartPort uint32 `json:"startPort"`
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:validation:Minimum=1
	EndPort       uint32 `json:"endPort"`
	SegmentLength uint32 `json:"segmentLength,omitempty"`
}

// GetKey get port pool item key
func (ppi *PortPoolItem) GetKey() string {
	tmpIDs := make([]string, len(ppi.LoadBalancerIDs))
	copy(tmpIDs, ppi.LoadBalancerIDs)
	sort.Strings(tmpIDs)
	return strings.Join(tmpIDs, ",")
}

// Validate do validation
func (ppi *PortPoolItem) Validate() error {
	if ppi == nil {
		return fmt.Errorf("port pool item cannot be empty")
	}
	if len(ppi.LoadBalancerIDs) == 0 {
		return fmt.Errorf("loadBalancerIDs cannot be empty")
	}
	if ppi.EndPort != 0 && ppi.EndPort <= ppi.StartPort {
		return fmt.Errorf("if endPort is not zero, it should be bigger than startPort")
	}
	return nil
}

// PortPoolSpec defines the desired state of PortPool
type PortPoolSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	PoolItems         []*PortPoolItem           `json:"poolItems"`
	ListenerAttribute *IngressListenerAttribute `json:"listenerAttribute,omitempty"`
}

// PortPoolItemStatus status of a port pool item
type PortPoolItemStatus struct {
	ItemName              string                 `json:"itemName"`
	LoadBalancerIDs       []string               `json:"loadBalancerIDs,omitempty"`
	StartPort             uint32                 `json:"startPort"`
	EndPort               uint32                 `json:"endPort"`
	SegmentLength         uint32                 `json:"segmentLength"`
	PoolItemLoadBalancers []*IngressLoadBalancer `json:"poolItemLoadBalancers,omitempty"`
	Status                string                 `json:"status"`
	Message               string                 `json:"message"`
}

// GetKey get port pool item key
func (ppis *PortPoolItemStatus) GetKey() string {
	tmpIDs := make([]string, len(ppis.LoadBalancerIDs))
	copy(tmpIDs, ppis.LoadBalancerIDs)
	sort.Strings(tmpIDs)
	return strings.Join(tmpIDs, ",")
}

// PortPoolStatus defines the observed state of PortPool
type PortPoolStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	PoolItemStatuses []*PortPoolItemStatus `json:"poolItems,omitempty"`
}

// +kubebuilder:object:root=true
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status

// PortPool is the Schema for the portpools API
type PortPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PortPoolSpec   `json:"spec,omitempty"`
	Status PortPoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PortPoolList contains a list of PortPool
type PortPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PortPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PortPool{}, &PortPoolList{})
}
