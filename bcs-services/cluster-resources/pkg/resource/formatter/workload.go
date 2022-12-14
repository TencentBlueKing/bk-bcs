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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/timex"
)

// FormatWorkloadRes xxx
func FormatWorkloadRes(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	ret["images"] = parseContainerImages(manifest, "spec.template.spec.containers")
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
		// ?????????????????? Job????????????????????? Job ???????????????????????? Key ??? 0
		if activeJobs, ok := status["active"]; ok {
			ret["active"] = len(activeJobs.([]interface{}))
		}
		// ??????????????????
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
			// ?????? job ????????????
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
		restartCnt += s.(map[string]interface{})["restartCount"].(int64)
	}
	ret["readyCnt"], ret["totalCnt"], ret["restartCnt"] = readyCnt, totalCnt, restartCnt

	podIPSet := set.NewStringSet()
	podIP := mapx.GetStr(manifest, "status.podIP")
	podIPSet.Add(podIP)

	// ????????????????????????
	for _, item := range mapx.GetList(manifest, "status.podIPs") {
		ip := item.(map[string]interface{})["ip"].(string)
		podIPSet.Add(ip)
	}

	// ???????????? ipv4 / ipv6 ??????
	ret["podIPv4"], ret["podIPv6"] = "", ""
	for _, ip := range podIPSet.ToSlice() {
		switch {
		case stringx.IsIPv4(ip):
			ret["podIPv4"] = ip
		case stringx.IsIPv6(ip):
			ret["podIPv6"] = ip
		}
	}
	return ret
}

// ????????????/?????????

// DeployStatusChecker xxx
type DeployStatusChecker struct{}

// IsNormal ??????????????????????????????????????????????????????
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

// IsNormal ?????????????????? status.currentReplicas ????????????????????????????????????????????????????????????????????????????????????????????????
func (c *STSStatusChecker) IsNormal(manifest map[string]interface{}) bool {
	replicas := mapx.GetInt64(manifest, "spec.replicas")
	if curReplicas, err := mapx.GetItems(manifest, "status.currentReplicas"); err == nil {
		if curReplicas.(int64) != replicas {
			return false
		}
	}
	return slice.AllInt64Equal([]int64{
		mapx.GetInt64(manifest, "status.readyReplicas"),
		mapx.GetInt64(manifest, "status.updatedReplicas"),
		replicas,
	})
}

// WorkloadStatusParser ?????????????????? ??????????????? ?????????
type WorkloadStatusParser struct {
	checker  StatusChecker
	manifest map[string]interface{}
}

// Parse xxx
func (p *WorkloadStatusParser) Parse() string {
	// ?????????????????????????????????????????? deletionTimestamp ??????
	if dt := mapx.Get(p.manifest, "metadata.deletionTimestamp", nil); dt != nil {
		return WorkloadStatusDeleting
	}
	// ??????????????????????????????????????????
	if p.checker.IsNormal(p.manifest) {
		return WorkloadStatusNormal
	}
	// ??????????????????????????? generation????????? 1?????????????????????????????????????????????
	if gen := mapx.GetInt64(p.manifest, "metadata.generation"); gen == int64(1) {
		return WorkloadStatusCreating
	}
	return WorkloadStatusUpdating
}

func newDeployStatusParser(manifest map[string]interface{}) *WorkloadStatusParser {
	return &WorkloadStatusParser{&DeployStatusChecker{}, manifest}
}

func newSTSStatusParser(manifest map[string]interface{}) *WorkloadStatusParser {
	return &WorkloadStatusParser{&STSStatusChecker{}, manifest}
}

// parseContainerImages ????????????????????????????????? image ???????????????
func parseContainerImages(manifest map[string]interface{}, paths string) []string {
	images := set.NewStringSet()
	for _, c := range mapx.GetList(manifest, paths) {
		if image, ok := c.(map[string]interface{})["image"]; ok {
			images.Add(image.(string))
		}
	}
	return images.ToSlice()
}

// PodStatusParser Pod ???????????????
type PodStatusParser struct {
	Manifest     map[string]interface{}
	initializing bool
	// Pod ?????????
	totalStatus string
}

// Parse ????????????????????????????????????https://github.com/kubernetes/dashboard/blob/92a8491b99afa2cfb94dbe6f3410cadc42b0dc31/modules/api/pkg/resource/pod/common.go#L40
func (p *PodStatusParser) Parse() string {
	// ?????????????????? PodStatus ???????????? Pod Status???total?????????
	podStatus := LightPodStatus{}
	if err := mapstructure.Decode(p.Manifest["status"], &podStatus); err != nil {
		return "--"
	}

	// 1. ???????????? Pod.Status.Phase
	p.totalStatus = string(podStatus.Phase)

	// 2. ??????????????? Pod.Status.Reason ?????????
	if podStatus.Reason != "" {
		p.totalStatus = podStatus.Reason
	}

	// 3. ?????? Pod ????????????????????????
	p.updateStatusByInitContainerStatuses(&podStatus)
	if !p.initializing {
		p.updateStatusByContainerStatuses(&podStatus)
	}

	// 4. ?????? Pod.Metadata.DeletionTimestamp ????????????
	deletionTimestamp, _ := mapx.GetItems(p.Manifest, "metadata.deletionTimestamp")
	if deletionTimestamp != nil && podStatus.Reason == "NodeLost" {
		p.totalStatus = string(v1.PodUnknown)
	} else if deletionTimestamp != nil {
		p.totalStatus = "Terminating"
	}

	// 5. ?????????????????????????????????????????????????????????????????????
	if len(p.totalStatus) == 0 {
		p.totalStatus = string(v1.PodUnknown)
	}
	return p.totalStatus
}

// updateStatusByInitContainerStatuses ?????? pod.Status.InitContainerStatuses ?????? ?????????
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

// updateStatusByContainerStatuses ?????? pod.Status.ContainerStatuses ?????? ?????????
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
