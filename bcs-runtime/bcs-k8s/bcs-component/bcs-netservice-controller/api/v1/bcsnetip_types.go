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

// BCSNetIPSpec defines the desired state of BCSNetIP
type BCSNetIPSpec struct {
	// 所属网段
	Net string `json:"net"`
	// 网段掩码
	Mask int `json:"mask"`
	// 网段网关
	Gateway string `json:"gateway"`
}

// BCSNetIPStatus defines the observed state of BCSNetIP
type BCSNetIPStatus struct {
	// Active --已使用，Available --可用, Reserved --保留
	Phase string `json:"phase,omitempty"`
	// 对应主机信息
	Host string `json:"host,omitempty"`
	// 是否被用作固定IP
	Fixed bool `json:"fixed,omitempty"`
	// 容器ID
	ContainerID string `json:"containerID,omitempty"`
	// BCSNetIPClaim名称
	IPClaimKey   string      `json:"ipClaimKey,omitempty"`
	PodName      string      `json:"podName,omitempty"`
	PodNamespace string      `json:"podNamespace,omitempty"`
	UpdateTime   metav1.Time `json:"updateTime,omitempty"`
	KeepDuration string      `json:"keepDuration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:scope=Cluster
//+kubebuilder:subresource:status

// BCSNetIP is the Schema for the bcsnetips API
type BCSNetIP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BCSNetIPSpec   `json:"spec,omitempty"`
	Status BCSNetIPStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BCSNetIPList contains a list of BCSNetIP
type BCSNetIPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BCSNetIP `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BCSNetIP{}, &BCSNetIPList{})
}
