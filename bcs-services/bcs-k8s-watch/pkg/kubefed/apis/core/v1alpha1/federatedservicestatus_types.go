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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FederatedServiceClusterStatus is the observed status of the resource for a named cluster
type FederatedServiceClusterStatus struct {
	ClusterName string               `json:"clusterName"`
	Status      corev1.ServiceStatus `json:"status"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FederatedServiceStatus
// +k8s:openapi-gen=true
// +kubebuilder:resource:path=federatedservicestatuses
type FederatedServiceStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +optional
	ClusterStatus []FederatedServiceClusterStatus `json:"clusterStatus,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FederatedServiceStatusList contains a list of FederatedServiceStatus
type FederatedServiceStatusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FederatedServiceStatus `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FederatedServiceStatus{}, &FederatedServiceStatusList{})
}
