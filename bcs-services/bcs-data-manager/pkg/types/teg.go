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

// Package types xxx
package types

var (
	// TegClusterId key for TEG workload
	TegClusterId = "clusterId"
	// TegNamespace key for TEG workload
	TegNamespace = "namespace"
	// TegWorkloadKind key for TEG workload
	TegWorkloadKind = "workload_kind"
	// TegWorkloadName key for TEG workload
	TegWorkloadName = "workload_name"
	// TegMaintainer key for TEG workload
	TegMaintainer = "maintainer"
	// TegBakMaintainer key for TEG workload
	TegBakMaintainer = "bakMaintainer"
	// TegBusinessSetId key for TEG workload
	TegBusinessSetId = "businessSetId"
	// TegBusinessId key for TEG workload
	TegBusinessId = "businessId"
	// TegBusinessModuleId key for TEG workload
	TegBusinessModuleId = "businessModuleId"
	// TegSchedulerStatus key for TEG workload
	TegSchedulerStatus = "schedulerStatus"
	// TegServiceStatus key for TEG workload
	TegServiceStatus = "serviceStatus"

	// TEGWorkloadColumns selected columns for workloads
	TEGWorkloadColumns = []string{
		TegClusterId, TegNamespace, TegWorkloadKind,
		TegWorkloadName, TegMaintainer, TegBakMaintainer,
		TegBusinessSetId, TegBusinessId, TegBusinessModuleId,
		TegSchedulerStatus, TegServiceStatus,
	}

	// TEGWorkloadSortColumns columns for sort in order by
	TEGWorkloadSortColumns = []string{TegClusterId, TegNamespace, TegWorkloadKind, TegWorkloadName}
)

// TEGWorkload type of teg workload
type TEGWorkload struct {
	ClusterId        string `json:"clusterId,omitempty" db:"clusterId"`
	Namespace        string `json:"namespace,omitempty" db:"namespace"`
	WorkloadKind     string `json:"workloadKind,omitempty" bson:"workload_kind" db:"workload_kind"`
	WorkloadName     string `json:"workloadName,omitempty" bson:"workload_name" db:"workload_name"`
	Maintainer       string `json:"maintainer,omitempty" db:"maintainer"`
	BakMaintainer    string `json:"bakMaintainer,omitempty" db:"bakMaintainer"`
	BusinessSetId    int32  `json:"businessSetId,omitempty" db:"businessSetId"`
	BusinessId       int32  `json:"businessId,omitempty" db:"businessId"`
	BusinessModuleId int32  `json:"businessModuleId,omitempty" db:"businessModuleId"`
	SchedulerStatus  int32  `json:"schedulerStatus,omitempty" db:"schedulerStatus"`
	ServiceStatus    int32  `json:"serviceStatus,omitempty" db:"serviceStatus"`
	HpaStatus        int32  `json:"hpaStatus,omitempty" db:"hpaStatus"`
}
