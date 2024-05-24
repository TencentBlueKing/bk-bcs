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

package cluster

import (
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "k8s.io/api/core/v1"
)

// Node for kubernetes node
type Node struct {
	Name        string
	IP          string
	Status      string
	Labels      map[string]string
	Annotations map[string]string
}

// IsOptionalForScaleDown check node status
func (n *Node) IsOptionalForScaleDown(taskLabel, drainDelayLabel, taskID, drainDelay string) (bool, bool, int) {
	if n.Status != string(v1.ConditionTrue) {
		blog.Infof("node %s status %s", n.Name, n.Status)
		return false, false, 0
	}
	if n.Labels[taskLabel] != "" && n.Labels[taskLabel] != taskID {
		blog.Infof("node %s task id %s", n.IP, n.Labels[taskLabel])
		return false, false, 0
	}
	if n.Labels[drainDelayLabel] == "" || n.Labels[drainDelayLabel] == drainDelay {
		blog.Infof("node %s drain-delay %s", n.IP, n.Labels[drainDelayLabel])
		return true, false, 0
	}
	drainHourStr := strings.Split(drainDelay, "h")[0]
	drainHour, _ := strconv.Atoi(drainHourStr)
	nodeDrainHourStr := strings.Split(n.Labels[drainDelayLabel], "h")[0]
	nodeDrainHour, _ := strconv.Atoi(nodeDrainHourStr)
	if nodeDrainHour < drainHour {
		return false, true, nodeDrainHour
	}
	return false, false, 0
}
