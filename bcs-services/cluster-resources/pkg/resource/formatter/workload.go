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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
)

// FormatWorkloadRes ...
func FormatWorkloadRes(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	ret["images"] = parseContainerImages(manifest, "spec.template.spec.containers")
	return ret
}

// FormatCronJobRes ...
func FormatCronJobRes(manifest map[string]interface{}) map[string]interface{} {
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

// FormatJobRes ...
func FormatJobRes(manifest map[string]interface{}) map[string]interface{} {
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

// FormatPodRes ...
func FormatPodRes(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	ret["images"] = parseContainerImages(manifest, "spec.containers")
	parser := podStatusParser{manifest: manifest}
	ret["status"] = parser.parse()
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
