/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MultiClusterResourceQuotaTotalQuotaSpec spec for MultiClusterResourceQuotaTotalQuota
type MultiClusterResourceQuotaTotalQuotaSpec struct {
	Hard corev1.ResourceList `json:"hard,omitempty" protobuf:"bytes,1,rep,name=hard,casttype=ResourceList,castkey=ResourceName"`
}

// MultiClusterResourceQuotaTotalQuotaStatus status for MultiClusterResourceQuotaTotalQuota
type MultiClusterResourceQuotaTotalQuotaStatus struct {
	Hard corev1.ResourceList `json:"hard,omitempty" protobuf:"bytes,1,rep,name=hard,casttype=ResourceList,castkey=ResourceName"`

	Used corev1.ResourceList `json:"used,omitempty" protobuf:"bytes,2,rep,name=used,casttype=ResourceList,castkey=ResourceName"`
}

// MultiClusterResourceQuotaSpec defines the desired state of MultiClusterResourceQuota
type MultiClusterResourceQuotaSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	TotalQuota MultiClusterResourceQuotaTotalQuotaSpec `json:"total_quota,omitempty"`

	ScopeSelector   *corev1.ScopeSelector `json:"scopeSelector,omitempty" protobuf:"bytes,1,opt,name=scopeSelector"`
	TaskSelector    map[string]string     `json:"taskSelector,omitempty" protobuf:"bytes,2,opt,name=taskSelector"`
	ClusterSelector *ClusterSelector      `json:"clusterSelector,omitempty" protobuf:"bytes,3,opt,name=clusterSelector"`
}

// ClusterSelector cluster selectors of MultiClusterResourceQuota
type ClusterSelector struct {
	MatchExpressions []ClusterSelectorRequirement `json:"matchExpressions,omitempty" protobuf:"bytes,1,rep,name=matchExpressions"`
}

// ClusterSelectorRequirement cluster selector requirement
type ClusterSelectorRequirement struct {
	// The ID of the bcs clsuter that the selector applies to.
	Key string `json:"key" protobuf:"bytes,1,opt,name=key"`
	// Represents a scope's relationship to a set of values.
	// Valid operators are In, NotIn.
	Operator ClusterSelectorOperator `json:"operator" protobuf:"bytes,2,opt,name=operator,casttype=ClusterSelectorOperator"`
	// An array of string values. If the operator is In or NotIn,
	// the values array must be non-empty.
	Values []string `json:"values,omitempty" protobuf:"bytes,3,rep,name=values"`
}

// ClusterSelectorOperator clusterselector operator definition
type ClusterSelectorOperator string

const (
	ClusterSelectorOpIn    ClusterSelectorOperator = "In"
	ClusterSelectorOpNotIn ClusterSelectorOperator = "NotIn"
)

// MultiClusterResourceQuotaStatus defines the observed state of MultiClusterResourceQuota
type MultiClusterResourceQuotaStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	TotalQuota MultiClusterResourceQuotaTotalQuotaStatus `json:"total_quota,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// MultiClusterResourceQuota is the Schema for the multiclusterresourcequota API
type MultiClusterResourceQuota struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MultiClusterResourceQuotaSpec   `json:"spec,omitempty"`
	Status MultiClusterResourceQuotaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MultiClusterResourceQuotaList contains a list of MultiClusterResourceQuota
type MultiClusterResourceQuotaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MultiClusterResourceQuota `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MultiClusterResourceQuota{}, &MultiClusterResourceQuotaList{})
}
