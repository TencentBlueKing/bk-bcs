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

const (
	// PortPoolBindingLabelKeyFromat label key prefix for port pool
	PortPoolBindingLabelKeyFromat = "portpool.%s.%s"
	// PortPoolBindingAnnotationKeyKeepDuration annotation key for keep duration of port pool binding
	PortPoolBindingAnnotationKeyKeepDuration = "keepduration.portbinding.bkbcs.tencent.com"
)

// PortBindingItem defines the port binding item
type PortBindingItem struct {
	PoolName              string                    `json:"poolName"`
	PoolNamespace         string                    `json:"poolNamespace"`
	LoadBalancerIDs       []string                  `json:"loadBalancerIDs,omitempty"`
	ListenerAttribute     *IngressListenerAttribute `json:"listenerAttribute,omitempty"`
	PoolItemLoadBalancers []*IngressLoadBalancer    `json:"poolItemLoadBalancers,omitempty"`
	PoolItemName          string                    `json:"poolItemName"`
	Protocol              string                    `json:"protocol"`
	StartPort             int                       `json:"startPort"`
	EndPort               int                       `json:"endPort"`
	RsStartPort           int                       `json:"rsStartPort"`
	// +optional
	HostPort bool `json:"hostPort,omitempty"`
}

// GetKey get port pool item key
func (pbi *PortBindingItem) GetKey() string {
	return GetPoolItemKey(pbi.PoolItemName, pbi.LoadBalancerIDs)
}

// PortBindingSpec defines the desired state of PortBinding
type PortBindingSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	PortBindingList []*PortBindingItem `json:"portBindingList,omitempty"`
}

// PortBindingStatusItem port binding item status
type PortBindingStatusItem struct {
	PoolName      string `json:"portPoolName"`
	PoolNamespace string `json:"portPoolNamespace"`
	PoolItemName  string `json:"poolItemName"`
	StartPort     int    `json:"startPort"`
	EndPort       int    `json:"endPort"`
	// Status is single pool item status
	Status string `json:"status"`
}

// PortBindingStatus defines the observed state of PortBinding
type PortBindingStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// 整体Pod绑定的状态, NotReady, PartialReady, Ready
	Status                string                   `json:"status"`
	UpdateTime            string                   `json:"updateTime"`
	PortBindingStatusList []*PortBindingStatusItem `json:"portPoolBindStatusList,omitempty"`
}

// +kubebuilder:object:root=true
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:subresource:status

// PortBinding is the Schema for the portbindings API
type PortBinding struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PortBindingSpec   `json:"spec,omitempty"`
	Status PortBindingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// PortBindingList contains a list of PortBinding
type PortBindingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PortBinding `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PortBinding{}, &PortBindingList{})
}
