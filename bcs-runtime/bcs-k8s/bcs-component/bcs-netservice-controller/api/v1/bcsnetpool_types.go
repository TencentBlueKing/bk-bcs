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

// BCSNetPoolSpec defines the desired state of BCSNetPool
type BCSNetPoolSpec struct {
	// 网段
	Net string `json:"net"`
	// 网段掩码
	Mask int `json:"mask"`
	// 网段网关
	Gateway string `json:"gateway"`
	// 对应主机列表
	Hosts []string `json:"hosts,omitempty"`
	// 可用的IP
	AvailableIPs []string `json:"availableIPs,omitempty"`
}

// BCSNetPoolStatus defines the observed state of BCSNetPool
type BCSNetPoolStatus struct {
	// Initializing --初始化中，Normal --正常
	Phase      string      `json:"phase,omitempty"`
	UpdateTime metav1.Time `json:"updateTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

// BCSNetPool is the Schema for the bcsnetpools API
type BCSNetPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BCSNetPoolSpec   `json:"spec,omitempty"`
	Status BCSNetPoolStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BCSNetPoolList contains a list of BCSNetPool
type BCSNetPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BCSNetPool `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BCSNetPool{}, &BCSNetPoolList{})
}
