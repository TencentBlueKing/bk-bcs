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

package formatter

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/TencentBlueKing/gopkg/collection/set"
	"github.com/mitchellh/mapstructure"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/timex"
)

// FormatWorkloadRes xxx
func FormatWorkloadRes(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	ret["images"] = parseContainerImages(manifest, "spec.template.spec.containers")
	ret["resources"] = parseContainersResources(manifest, "spec.template.spec.containers")
	return ret
}

// FormatControllerRevisionRes xxx
func FormatControllerRevisionRes(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	ret["images"] = parseContainerImages(manifest, "data.spec.template.spec.containers")
	ret["resources"] = parseContainersResources(manifest, "data.spec.template.spec.containers")
	return ret
}

// FormatDeploy xxx
func FormatDeploy(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatWorkloadRes(manifest)
	ret["status"] = newDeployStatusParser(manifest).Parse()
	return ret
}

// FormatRS xxx
func FormatRS(manifest map[string]interface{}) map[string]interface{} {
	return FormatWorkloadRes(manifest)
}

// FormatSTS xxx
func FormatSTS(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatWorkloadRes(manifest)
	ret["status"] = newSTSStatusParser(manifest).Parse()
	return ret
}

// FormatCJ xxx
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
			ret["lastSchedule"] = timex.CalcDuration(status["lastScheduleTime"].(string), "")
		}
	}
	return ret
}

// FormatJob xxx
func FormatJob(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatWorkloadRes(manifest)
	ret["duration"] = "--"
	if status, ok := manifest["status"].(map[string]interface{}); ok {
		if status["startTime"] != nil && status["completionTime"] != nil {
			// 执行 job 持续时间
			ret["duration"] = timex.CalcDuration(status["startTime"].(string), status["completionTime"].(string))
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
	for _, s := range mapx.GetList(manifest, "status.containerStatuses") {
		if s.(map[string]interface{})["ready"].(bool) {
			readyCnt++
		}
		totalCnt++
		restartCnt += mapx.GetInt64(s.(map[string]interface{}), "restartCount")
	}
	ret["readyCnt"], ret["totalCnt"], ret["restartCnt"] = readyCnt, totalCnt, restartCnt

	podIPSet := set.NewStringSet()
	podIP := mapx.GetStr(manifest, "status.podIP")
	podIPSet.Add(podIP)

	// 双栈集群特有字段
	for _, item := range mapx.GetList(manifest, "status.podIPs") {
		ip := item.(map[string]interface{})["ip"].(string)
		podIPSet.Add(ip)
	}

	// 获取Readiness相关内容
	ret["readinessGates"] = parseReadinessGates(manifest)

	// 同时兼容 ipv4 / ipv6 集群
	ret["podIPv4"], ret["podIPv6"] = "", ""
	for _, ip := range podIPSet.ToSlice() {
		switch {
		case stringx.IsIPv4(ip):
			ret["podIPv4"] = ip
		case stringx.IsIPv6(ip):
			ret["podIPv6"] = ip
		}
	}
	ret["resources"] = parseContainersResources(manifest, "spec.containers")
	return ret
}

// 解析容器资源
func parseContainersResources(manifest map[string]interface{}, path string) (res map[string]interface{}) {
	containers := mapx.GetList(manifest, path)
	containerArray := []v1.Container{}
	marshal, err := json.Marshal(containers)
	if err != nil {
		log.Error(context.TODO(), "JSON marshaling error:: %s", err)
		return
	}
	err = json.Unmarshal(marshal, &containerArray)
	if err != nil {
		log.Error(context.TODO(), "JSON unmarshaling error:: %s", err)
		return
	}
	// 累加 CPU 和内存配置
	var totalLimCPU, totalReqCPU, totalReqMemory, totalLimMemory resource.Quantity
	for _, container := range containerArray {
		reqCPU := container.Resources.Requests[v1.ResourceCPU]
		reqMemory := container.Resources.Requests[v1.ResourceMemory]
		limCPU := container.Resources.Limits[v1.ResourceCPU]
		limMemory := container.Resources.Limits[v1.ResourceMemory]
		totalReqCPU.Add(reqCPU)
		totalReqMemory.Add(reqMemory)
		totalLimCPU.Add(limCPU)
		totalLimMemory.Add(limMemory)
	}
	res = map[string]interface{}{
		"limits":   map[string]interface{}{"cpu": totalLimCPU.String(), "memory": totalLimMemory.String()},
		"requests": map[string]interface{}{"cpu": totalReqCPU.String(), "memory": totalReqMemory.String()},
	}
	return res
}

// 解析资源parseReadinessGates相关内容
func parseReadinessGates(manifest map[string]interface{}) (resp map[string]interface{}) {
	// 存放模板设置的readinessGates相关内容， conditionType value为key值
	readinessGates := make(map[string]interface{}, 0)
	// 获取模板设置的readinessGates相关conditionType
	for _, item := range mapx.GetList(manifest, "spec.readinessGates") {
		if value, ok := item.(map[string]interface{}); ok {
			conditionTypeKey := mapx.GetStr(value, "conditionType")
			readinessGates[conditionTypeKey] = "<none>"
		}
	}

	// 获取status readinessGates相关内容
	for _, item := range mapx.GetList(manifest, "status.conditions") {
		if value, ok := item.(map[string]interface{}); ok {
			typeValue := mapx.GetStr(value, "type")
			// 如果Conditions的key值有内容，则赋值给readinessGates
			if _, okTypeValue := readinessGates[typeValue]; okTypeValue {
				readinessGates[typeValue] = mapx.GetStr(value, "status")
			}
		}
	}
	return readinessGates
}

// 工具方法/解析器

// DeployStatusChecker xxx
type DeployStatusChecker struct{}

// IsNormal 检查逻辑：检查以下四个字段值是否相等
func (c *DeployStatusChecker) IsNormal(manifest map[string]interface{}) bool {
	return slice.AllInt64Equal([]int64{
		mapx.GetInt64(manifest, "status.availableReplicas"),
		mapx.GetInt64(manifest, "status.readyReplicas"),
		mapx.GetInt64(manifest, "status.updatedReplicas"),
		mapx.GetInt64(manifest, "spec.replicas"),
	})
}

// STSStatusChecker xxx
type STSStatusChecker struct{}

// IsNormal 检查逻辑：若 status.currentReplicas 存在，则检查与其他几项是否相等，若不存在，则检查剩余几项是否相等
func (c *STSStatusChecker) IsNormal(manifest map[string]interface{}) bool {
	replicas := mapx.GetInt64(manifest, "spec.replicas")
	if curReplicas := mapx.GetInt64(manifest, "status.currentReplicas"); curReplicas != replicas {
		return false
	}
	return slice.AllInt64Equal([]int64{
		mapx.GetInt64(manifest, "status.readyReplicas"),
		mapx.GetInt64(manifest, "status.updatedReplicas"),
		replicas,
	})
}

// WorkloadStatusParser 工作负载资源 自定义状态 解析器
type WorkloadStatusParser struct {
	checker  StatusChecker
	manifest map[string]interface{}
}

// Parse xxx
func (p *WorkloadStatusParser) Parse() string {
	// 删除中优先级最高，判断根据是 deletionTimestamp 存在
	if dt := mapx.Get(p.manifest, "metadata.deletionTimestamp", nil); dt != nil {
		return WorkloadStatusDeleting
	}
	// 不同资源类型正常判断条件不同
	if p.checker.IsNormal(p.manifest) {
		return WorkloadStatusNormal
	}
	// 若非正常情况，检查 generation，若为 1（第一个版本），则状态为创建中
	if gen := mapx.GetInt64(p.manifest, "metadata.generation"); gen == int64(1) {
		return WorkloadStatusCreating
	}
	// 如果包含重启标识，则状态为重启中
	restartPath := []string{"spec", "template", "metadata", "annotations", WorkloadRestartAnnotationKey}
	versionPath := []string{"spec", "template", "metadata", "annotations", WorkloadRestartVersionAnnotationKey}
	if mapx.GetStr(p.manifest, restartPath) != "" &&
		mapx.GetStr(p.manifest, versionPath) == strconv.Itoa(int(mapx.GetInt64(p.manifest, "metadata.generation"))) {
		return WorkloadStatusRestarting
	}
	return WorkloadStatusUpdating
}

func newDeployStatusParser(manifest map[string]interface{}) *WorkloadStatusParser {
	return &WorkloadStatusParser{&DeployStatusChecker{}, manifest}
}

func newSTSStatusParser(manifest map[string]interface{}) *WorkloadStatusParser {
	return &WorkloadStatusParser{&STSStatusChecker{}, manifest}
}

// parseContainerImages 遍历每个容器，收集所有 image 信息并去重
func parseContainerImages(manifest map[string]interface{}, paths string) []string {
	images := set.NewStringSet()
	for _, c := range mapx.GetList(manifest, paths) {
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

// Parse 状态解析逻辑
// nolint
// 参考来源：https://github.com/kubernetes/dashboard/blob/92a8491b99afa2cfb94dbe6f3410cadc42b0dc31/modules/api/pkg/resource/pod/common.go#L40
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
	deletionTimestamp, _ := mapx.GetItems(p.Manifest, "metadata.deletionTimestamp")
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

// updateStatusByInitContainerStatuses 根据 pod.Status.InitContainerStatuses 更新 总状态
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
			if container.State.Waiting != nil && len(container.State.Waiting.Reason) > 0 &&
				container.State.Waiting.Reason != "PodInitializing" {
				p.totalStatus = fmt.Sprintf("Init: %s", container.State.Waiting.Reason)
			} else {
				initContainers := mapx.GetList(p.Manifest, "spec.initContainers")
				p.totalStatus = fmt.Sprintf("Init: %d/%d", i, len(initContainers))
			}
		}
		break
	}
}

// updateStatusByContainerStatuses 根据 pod.Status.ContainerStatuses 更新 总状态
func (p *PodStatusParser) updateStatusByContainerStatuses(podStatus *LightPodStatus) { //nolint:cyclop
	hasRunning := false
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
