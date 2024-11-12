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

package quota

var (
	// PodPriority pod调度优先级
	PodPriority = "task.bkbcs.tencent.com/priority"
	// PodPriorityHigh pod调度优先级-高
	PodPriorityHigh = "high"
	// PodPriorityLow pod调度优先级-低
	PodPriorityLow = "low"

	// PodTTL pod稳定时长
	PodTTL = "task.bkbcs.tencent.com/ttl"
	// PodTTL1h pod稳定时长-1h
	PodTTL1h = "1h"
	// PodTTL2h pod稳定时长-2h
	PodTTL2h = "2h"
	// PodTTL24h pod稳定时长-24h
	PodTTL24h = "24h"
	// PodTTL3d pod稳定时长-3d
	PodTTL3d = "3d"
	// PodTTL7d pod稳定时长-7d
	PodTTL7d = "7d"

	// PodCpuType CPU类型
	PodCpuType = "task.bkbcs.tencent.com/cpu-type"
	// PodCpuTypeStatic CPU类型-静态
	PodCpuTypeStatic = "static"
	// PodCpuTypeLow CPU类型-低
	PodCpuTypeLow = "low"

	// PodGpuType GPU类型
	PodGpuType = "task.bkbcs.tencent.com/gpu-type"
	// PodGpuTypeL20 GPU类型-L20
	PodGpuTypeL20 = "L20"
)

// Attribute 属性对象
type Attribute map[string]string

// IsEqual 实现一个方法来检查两个 属性是否完全相同
func (m Attribute) IsEqual(other Attribute) bool {
	// 首先检查长度是否相同
	if len(m) != len(other) {
		return false
	}

	// 然后检查每个键值对是否相同
	for key, value := range m {
		if otherValue, ok := other[key]; !ok || otherValue != value {
			return false
		}
	}

	return true
}
