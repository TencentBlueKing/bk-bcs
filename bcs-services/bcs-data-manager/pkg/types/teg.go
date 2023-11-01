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

// TEGWorkload type of teg workload
type TEGWorkload struct {
	ClusterId        string `json:"clusterId,omitempty"`
	Namespace        string `json:"namespace,omitempty"`
	WorkloadKind     string `json:"workloadKind,omitempty" bson:"workload_kind"`
	WorkloadName     string `json:"workloadName,omitempty" bson:"workload_name"`
	Maintainer       string `json:"maintainer,omitempty"`
	BakMaintainer    string `json:"bakMaintainer,omitempty"`
	BusinessSetId    int64  `json:"businessSetId,omitempty"`
	BusinessId       int64  `json:"businessId,omitempty"`
	BusinessModuleId int64  `json:"businessModuleId,omitempty"`
	SchedulerStatus  int64  `json:"schedulerStatus,omitempty"`
	ServiceStatus    int64  `json:"serviceStatus,omitempty"`
	HpaStatus        int64  `json:"hpaStatus,omitempty"`
}
