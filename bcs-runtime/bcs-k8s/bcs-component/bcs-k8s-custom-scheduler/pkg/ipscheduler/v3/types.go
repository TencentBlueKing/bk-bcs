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

package v3

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	// BCSNetIPClaimBoundedStatus for BCSNetIPClaim Bound status
	BCSNetIPClaimBoundedStatus = "Bound"
	// BCSNetIPClaimPendingStatus for BCSNetIPClaim Pending status
	BCSNetIPClaimPendingStatus = "Pending"
	// BCSNetIPClaimExpiredStatus for BCSNetIPClaim Expired status
	BCSNetIPClaimExpiredStatus = "Expired"

	// BCSNetIPActiveStatus for BCSNetIP Active status
	BCSNetIPActiveStatus = "Active"
	// BCSNetIPAvailableStatus for BCSNetIP Available status
	BCSNetIPAvailableStatus = "Available"
	// BCSNetIPReservedStatus for BCSNetIP Reserved status
	BCSNetIPReservedStatus = "Reserved"

	// PodLabelKeyForPool pod label key for pool
	PodLabelKeyForPool = "pool"
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

// BCSNetIPClaim is the Schema for the bcsnetipclaims API
type BCSNetIPClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BCSNetIPClaimSpec   `json:"spec,omitempty"`
	Status BCSNetIPClaimStatus `json:"status,omitempty"`
}

// BCSNetIPClaimList contains a list of BCSNetIPClaim
type BCSNetIPClaimList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BCSNetIPClaim `json:"items"`
}

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

// BCSNetIP is the Schema for the bcsnetips API
type BCSNetIP struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BCSNetIPSpec   `json:"spec,omitempty"`
	Status BCSNetIPStatus `json:"status,omitempty"`
}

// BCSNetIPList contains a list of BCSNetIP
type BCSNetIPList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BCSNetIP `json:"items"`
}

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

// BCSNetPool is the Schema for the bcsnetpools API
type BCSNetPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BCSNetPoolSpec   `json:"spec,omitempty"`
	Status BCSNetPoolStatus `json:"status,omitempty"`
}

// BCSNetPoolList contains a list of BCSNetPool
type BCSNetPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BCSNetPool `json:"items"`
}
