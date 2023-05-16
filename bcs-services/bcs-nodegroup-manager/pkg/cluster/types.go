/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cluster

import (
	v1 "k8s.io/api/core/v1"
)

// Node for kubernetes node
type Node struct {
	Name   string
	IP     string
	Status string
	Labels map[string]string
}

// IsOptionalForScaleDown check node status
func (n *Node) IsOptionalForScaleDown(labelKey, taskid string) bool {
	if n.Status != string(v1.ConditionTrue) {
		return false
	}
	if n.Labels[labelKey] == "" || n.Labels[labelKey] == taskid {
		return true
	}
	return false
}
