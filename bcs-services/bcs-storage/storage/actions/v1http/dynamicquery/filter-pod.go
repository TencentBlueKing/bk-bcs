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

package dynamicquery

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
)

// PodFilter filter of pod
type PodFilter struct {
	ClusterId      string `json:"clusterId" filter:"clusterId"`
	Name           string `json:"name,omitempty" filter:"resourceName"`
	Namespace      string `json:"namespace,omitempty" filter:"namespace"`
	HostIp         string `json:"hostIp,omitempty" filter:"data.status.hostIP"`
	PodIp          string `json:"podIp,omitempty" filter:"data.status.podIP"`
	Status         string `json:"status,omitempty" filter:"data.status.phase"`
	StartTimeBegin string `json:"startTimeBegin,omitempty" filter:"data.status.startTime,timeL"`
	StartTimeEnd   string `json:"startTimeEnd,omitempty" filter:"data.status.startTime,timeR"`
}

const podNestedTimeLayout = nestedTimeLayout

func (t PodFilter) getCondition() *operator.Condition {
	return qGenerate(t, podNestedTimeLayout)
}
