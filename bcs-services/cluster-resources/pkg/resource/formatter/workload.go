/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package formatter

import (
	"fmt"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/mitchellh/mapstructure"
	v1 "k8s.io/api/core/v1"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
)

// FormatWorkloadRes ...
func FormatWorkloadRes(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	ret["images"] = parseContainerImages(manifest, "spec.template.spec.containers")
	return ret
}

// FormatCJ ...
func FormatCJ(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	ret["images"] = parseContainerImages(manifest, "spec.jobTemplate.spec.template.spec.containers")
	ret["active"], ret["lastSchedule"] = 0, "--"
	if status, ok := manifest["status"].(map[string]interface{}); ok {
		// 若有执行中的 Job，则该字段值为 Job 列表长度，否则该 Key 为 0
		if activeJobs, ok := status["active"]; ok {
			ret["active"] = len(activeJobs.([]interface{}))
		}
		// 最后调度时间
		if status["lastScheduleTime"] != nil {
			ret["lastSchedule"] = util.CalcDuration(status["lastScheduleTime"].(string), "")
		}
	}
	return ret
}

// FormatJob ...
func FormatJob(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatWorkloadRes(manifest)
	ret["duration"] = "--"
	if status, ok := manifest["status"].(map[string]interface{}); ok {
		if status["startTime"] != nil && status["completionTime"] != nil {
			// 执行 job 持续时间
			ret["duration"] = util.CalcDuration(status["startTime"].(string), status["completionTime"].(string))
		}
	}
	return ret
}

// FormatPo ...
func FormatPo(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	ret["images"] = parseContainerImages(manifest, "spec.containers")
	parser := PodStatusParser{Manifest: manifest}
	ret["status"] = parser.Parse()
	readyCnt, totalCnt, restartCnt := 0, 0, int64(0)
	if status, ok := manifest["status"].(map[string]interface{}); ok {
		if containerStatuses, ok := status["containerStatuses"]; ok {
			for _, s := range containerStatuses.([]interface{}) {
				if s.(map[string]interface{})["ready"].(bool) {
					readyCnt++
				}
				totalCnt++
				restartCnt += s.(map[string]interface{})["restartCount"].(int64)
			}
		}
	}
	ret["readyCnt"], ret["totalCnt"], ret["restartCnt"] = readyCnt, totalCnt, restartCnt
	return ret
}

// 工具方法/解析器

// 遍历每个容器，收集所有 image 信息并去重
func parseContainerImages(manifest map[string]interface{}, paths string) []string {
	images := set.NewStringSet()
	containers, _ := util.GetItems(manifest, paths)
	for _, c := range containers.([]interface{}) {
		if image, ok := c.(map[string]interface{})["image"]; ok {
			images.Add(image.(string))
		}
	}
	return images.ToSlice()
}

// PodStatusParser Pod 状态解析器
type PodStatusParser struct {
	Manifest     map[string]interface{}
	initializing bool
	// Pod 总状态
	totalStatus string
}

// Parse 状态解析逻辑，参考来源：https://github.com/kubernetes/dashboard/blob/master/src/app/backend/resource/pod/common.go#L40
func (p *PodStatusParser) Parse() string {
	// 构造轻量化的 PodStatus 用于解析 Pod Status（total）字段
	podStatus := LightPodStatus{}
	if err := mapstructure.Decode(p.Manifest["status"], &podStatus); err != nil {
		return "--"
	}

	// 1. 默认使用 Pod.Status.Phase
	p.totalStatus = string(podStatus.Phase)

	// 2. 若有具体的 Pod.Status.Reason 则使用
	if podStatus.Reason != "" {
		p.totalStatus = podStatus.Reason
	}

	// 3. 根据 Pod 容器状态更新状态
	p.updateStatusByInitContainerStatuses(&podStatus)
	if !p.initializing {
		p.updateStatusByContainerStatuses(&podStatus)
	}

	// 4. 根据 Pod.Metadata.DeletionTimestamp 更新状态
	deletionTimestamp, _ := util.GetItems(p.Manifest, "metadata.deletionTimestamp")
	if deletionTimestamp != nil && podStatus.Reason == "NodeLost" {
		p.totalStatus = string(v1.PodUnknown)
	} else if deletionTimestamp != nil {
		p.totalStatus = "Terminating"
	}

	// 5. 若状态未初始化或在转移中丢失，则标记为未知状态
	if len(p.totalStatus) == 0 {
		p.totalStatus = string(v1.PodUnknown)
	}
	return p.totalStatus
}

// 根据 pod.Status.InitContainerStatuses 更新 总状态
func (p *PodStatusParser) updateStatusByInitContainerStatuses(podStatus *LightPodStatus) {
	for i := range podStatus.InitContainerStatuses {
		container := podStatus.InitContainerStatuses[i]
		if container.State.Terminated != nil { // nolint:nestif
			if container.State.Terminated.ExitCode == 0 {
				continue
			}
			p.initializing = true
			if len(container.State.Terminated.Reason) != 0 {
				p.totalStatus = "Init: " + container.State.Terminated.Reason
			} else if container.State.Terminated.Signal != 0 {
				p.totalStatus = fmt.Sprintf("Init: Signal %d", container.State.Terminated.Signal)
			} else {
				p.totalStatus = fmt.Sprintf("Init: ExitCode %d", container.State.Terminated.ExitCode)
			}
		} else {
			p.initializing = true
			if container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 && container.State.Waiting.Reason != "PodInitializing" { // nolint:lll
				p.totalStatus = fmt.Sprintf("Init: %s", container.State.Waiting.Reason)
			} else {
				initContainers, _ := util.GetItems(p.Manifest, "spec.initContainers")
				p.totalStatus = fmt.Sprintf("Init: %d/%d", i, len(initContainers.([]interface{})))
			}
		}
		break
	}
}

// 根据 pod.Status.ContainerStatuses 更新 总状态
func (p *PodStatusParser) updateStatusByContainerStatuses(podStatus *LightPodStatus) { //nolint:cyclop
	var hasRunning = false
	for i := len(podStatus.ContainerStatuses) - 1; i >= 0; i-- {
		container := podStatus.ContainerStatuses[i]
		if container.State.Waiting != nil && container.State.Waiting.Reason != "" {
			p.totalStatus = container.State.Waiting.Reason
		} else if container.State.Terminated != nil {
			if container.State.Terminated.Reason != "" {
				p.totalStatus = container.State.Terminated.Reason
			} else if container.State.Terminated.Signal != 0 {
				p.totalStatus = fmt.Sprintf("Signal: %d", container.State.Terminated.Signal)
			} else {
				p.totalStatus = fmt.Sprintf("ExitCode: %d", container.State.Terminated.ExitCode)
			}
		} else if container.Ready && container.State.Running != nil {
			hasRunning = true
		}
	}
	if p.totalStatus == "Completed" && hasRunning {
		if hasPodReadyCondition(podStatus.Conditions) {
			p.totalStatus = string(v1.PodRunning)
		} else {
			p.totalStatus = "NotReady"
		}
	}
}

func hasPodReadyCondition(conditions []LightPodCondition) bool {
	for _, condition := range conditions {
		if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}
