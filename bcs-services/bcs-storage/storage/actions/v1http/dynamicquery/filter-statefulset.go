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

// StatefulSetFilter filter fo statefulset
type StatefulSetFilter struct {
	ClusterId       string `json:"clusterId" filter:"clusterId"`
	Name            string `json:"name,omitempty" filter:"resourceName"`
	Namespace       string `json:"namespace,omitempty" filter:"namespace"`
	CreateTimeBegin string `json:"createTimeBegin,omitempty" filter:"data.metadata.creationTimestamp,timeL"`
	CreateTimeEnd   string `json:"createTimeEnd,omitempty" filter:"data.metadata.creationTimestamp,timeR"`
	ResourceVersion string `json:"resourceVersion,omitempty" filter:"data.metadata.resourceVersion"`
	Uid             string `json:"uid,omitempty" filter:"data.metadata.uid"`
	Generation      string `json:"generation,omitempty" filter:"data.metadata.generation,int"`
	PodMPolicy      string `json:"podManagementPolicy,omitempty" filter:"data.spec.podManagementPolicy"`
	StrategyType    string `json:"updateStrategyType,omitempty" filter:"data.spec.updateStrategy.type"`
	ServiceName     string `json:"serviceName,omitempty" filter:"data.spec.serviceName"`
}

const statefulSetNestedTimeLayout = nestedTimeLayout

func (t StatefulSetFilter) getCondition() *operator.Condition {
	return qGenerate(t, statefulSetNestedTimeLayout)
}
