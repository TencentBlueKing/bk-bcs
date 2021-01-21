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

// TaskGroupFilter filter of taskgroup
type TaskGroupFilter struct {
	ClusterId           string `json:"clusterId" filter:"clusterId"`
	Name                string `json:"name,omitempty" filter:"resourceName"`
	Namespace           string `json:"namespace,omitempty" filter:"namespace"`
	RcName              string `json:"rcName,omitempty" filter:"data.rcname"`
	Status              string `json:"status,omitempty" filter:"data.status"`
	LastStatus          string `json:"lastStatus,omitempty" filter:"data.lastStatus"`
	HostIp              string `json:"hostIp,omitempty" filter:"data.hostIP"`
	HostName            string `json:"hostName,omitempty" filter:"data.hostName"`
	PodIp               string `json:"podIp,omitempty" filter:"data.podIP"`
	CreateTimeBegin     string `json:"createTimeBegin,omitempty" filter:"data.metadata.creationTimestamp,timeL"`
	CreateTimeEnd       string `json:"createTimeEnd,omitempty" filter:"data.metadata.creationTimestamp,timeR"`
	StartTimeBegin      string `json:"startTimeBegin,omitempty" filter:"data.startTime,timeL"`
	StartTimeEnd        string `json:"startTimeEnd,omitempty" filter:"data.startTime,timeR"`
	LastUpdateTimeBegin string `json:"lastUpdateTimeBegin,omitempty" filter:"data.lastUpdateTime,timeL"`
	LastUpdateTimeEnd   string `json:"lastUpdateTimeEnd,omitempty" filter:"data.lastUpdateTime,timeR"`
}

const taskGroupNestedTimeLayout = nestedTimeLayout

func (t TaskGroupFilter) getCondition() *operator.Condition {
	return qGenerate(t, taskGroupNestedTimeLayout)
}
