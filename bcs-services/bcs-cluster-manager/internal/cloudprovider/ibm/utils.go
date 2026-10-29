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

package ibm

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

var (
	cloudName = "ibm"
)

// IBM VPC k8s node label key 定义
const (
	// zone相关label
	IBMZoneTopologyLabel = "topology.kubernetes.io/zone"
	IBMZoneBetaLabel     = "failure-domain.beta.kubernetes.io/zone"
	// region相关label
	IBMRegionTopologyLabel = "topology.kubernetes.io/region"
	IBMRegionBetaLabel     = "failure-domain.beta.kubernetes.io/region"
	// instance-type相关label（IBM专属）
	IBMMachineTypeLabel = "ibm-cloud.kubernetes.io/machine-type"
	// vpc实例ID label（IBM专属）
	IBMVpcInstanceIDLabel = "ibm-cloud.kubernetes.io/vpc-instance-id"
	// worker ID label（IBM专属）
	IBMWorkerIDLabel = "ibm-cloud.kubernetes.io/worker-id"
)

// getLabelConfig 返回IBM VPC的k8s node label配置
// IBM VPC同时支持新版标准label和旧版beta label，按优先级从高到低排列
func getLabelConfig() *cloudprovider.K8sLabelConfig {
	return &cloudprovider.K8sLabelConfig{
		ZoneLabels: []string{
			IBMZoneTopologyLabel, // topology.kubernetes.io/zone（优先）
			IBMZoneBetaLabel,     // failure-domain.beta.kubernetes.io/zone
		},
		RegionLabels: []string{
			IBMRegionTopologyLabel, // topology.kubernetes.io/region（优先）
			IBMRegionBetaLabel,     // failure-domain.beta.kubernetes.io/region
		},
		InstanceTypeLabels: []string{
			IBMMachineTypeLabel, // ibm-cloud.kubernetes.io/machine-type（IBM专属）
		},
	}
}

// parseIBMInstanceID 从k8s node的spec.providerID中解析出IBM实例ID
// IBM VPC providerID格式: ibm:///crn:v1:staging:public:containers-kubernetes:...，优先从label获取vpc-instance-id
// 若providerID为空或无法解析，则返回空字符串（由调用方通过ibm-cloud.kubernetes.io/vpc-instance-id label获取）
func parseIBMInstanceID(providerID string) string {
	if providerID == "" {
		return ""
	}
	// IBM providerID格式不统一，尝试按"/"分割取最后一段
	parts := strings.Split(providerID, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return providerID
}
