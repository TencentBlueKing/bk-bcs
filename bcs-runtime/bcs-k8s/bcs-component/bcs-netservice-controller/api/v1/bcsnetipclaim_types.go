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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BCSNetIPClaimSpec defines the desired state of BCSNetIPClaim
type BCSNetIPClaimSpec struct {
	// BCSNetIPName sets the name for BCSNetIP will be bounded with this claim
	BCSNetIPName string `json:"bcsNetIPName,omitempty"`
	// ExpiredDuration defines expired duration for this claim after claimed IP is released
	ExpiredDuration string `json:"expiredDuration,omitempty"`
}

// BCSNetIPClaimStatus defines the observed state of BCSNetIPClaim
type BCSNetIPClaimStatus struct {
	// BCSNetIPName is name for BCSNetIP bounded with this claim
	BoundedIP string `json:"boundedIP"`
	// Phase represents the state of this claim
	Phase string `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BCSNetIPClaim is the Schema for the bcsnetipclaims API
type BCSNetIPClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BCSNetIPClaimSpec   `json:"spec,omitempty"`
	Status BCSNetIPClaimStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BCSNetIPClaimList contains a list of BCSNetIPClaim
type BCSNetIPClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BCSNetIPClaim `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BCSNetIPClaim{}, &BCSNetIPClaimList{})
}
