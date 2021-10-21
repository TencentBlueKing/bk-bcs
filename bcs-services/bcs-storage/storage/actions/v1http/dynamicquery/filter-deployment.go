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

// DeploymentFilter deployment filter
type DeploymentFilter struct {
	ClusterId            string `json:"clusterId" filter:"clusterId"`
	Name                 string `json:"name,omitempty" filter:"resourceName"`
	Namespace            string `json:"namespace,omitempty" filter:"namespace"`
	CheckTime            string `json:"checkTime,omitempty" filter:"data.check_time,int64"`
	Status               string `json:"status,omitempty" filter:"data.status"`
	ApplicationName      string `json:"applicationName,omitempty" filter:"data.application.name"`
	ApplicationExtName   string `json:"applicationExtName,omitempty" filter:"data.application_ext.name"`
	CurrRollingOp        string `json:"currRollingOperation,omitempty" filter:"data.curr_rolling_operation"`
	IsInRolling          string `json:"isInRolling,omitempty" filter:"data.is_in_rolling,bool"`
	LastRollingTimeBegin string `json:"lastRollingTimeBegin,omitempty" filter:"data.last_rolling_time,timeL"`
	LastRollingTimeEnd   string `json:"lastRollingTimeEnd,omitempty" filter:"data.last_rolling_time,timeR"`
}

const deploymentNestedTimeLayout = nestedTimeLayout

func (t DeploymentFilter) getCondition() *operator.Condition {
	return qGenerate(t, deploymentNestedTimeLayout)
}
