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

package patch

import (
	"github.com/google/go-cmp/cmp"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/utils"
)

// ThreeWayMergeMonitorRule param current from bkm
// original: 上次变更的rule
// current： 当前bkm平台上的rule
// modified： 本次尝试变更的rule
func ThreeWayMergeMonitorRule(original, current, modified []*v1.MonitorRuleDetail) []*v1.MonitorRuleDetail {
	mergeResult := make([]*v1.MonitorRuleDetail, 0)

	originalMap := make(map[string]*v1.MonitorRuleDetail)
	currentMap := make(map[string]*v1.MonitorRuleDetail)
	modifiedMap := make(map[string]*v1.MonitorRuleDetail)

	for _, rule := range original {
		originalMap[rule.Name] = rule
	}
	for _, rule := range current {
		currentMap[rule.Name] = rule
	}
	for _, rule := range modifiedMap {
		modifiedMap[rule.Name] = rule
	}

	for _, modifiedRule := range modified {
		originalRule, ok := originalMap[modifiedRule.Name]
		if !ok { // 新增策略
			mergeResult = append(mergeResult, modifiedRule)
			continue
		}

		currentRule, ok := currentMap[modifiedRule.Name]
		if !ok { // 用户删除了对应策略 or 新建策略
			mergeResult = append(mergeResult, modifiedRule)
			continue
		}

		// 策略：
		// 1. 如果用户对策略进行了 ： 开关、修改告警阈值（包括触发周期）  则保留用户的所有变更
		// 2. 如果用户更新了告警策略的通知方式， 则采用用户的告警通知， 其他字段以模板为准
		// 3. 其余情况， 统一以模板为准
		var mergeRule *v1.MonitorRuleDetail
		if cmp.Equal(originalRule, currentRule, cmp.Comparer(compareMonitorRule)) {
			blog.Infof("mergeRule..")
			// 相同说明用户没有在面板上进行修改
			mergeRule = modifiedRule
			// 告警组配置以用户设置为准
			mergeRule.Notice = currentRule.Notice
		} else {
			blog.Infof("changed Rule..")
			// 用户进行了修改， 不做变更
			mergeRule = currentRule
		}

		mergeResult = append(mergeResult, mergeRule)
	}

	blog.Infof("original rule: %s, \ncurrent rule: %s, \nmodified rule: %s, \nmerged rule: %s",
		utils.ToJsonString(original), utils.ToJsonString(current), utils.ToJsonString(modified), utils.ToJsonString(mergeResult))
	return mergeResult
}

// compareMonitorRule return true if equal
func compareMonitorRule(mr1, mr2 *v1.MonitorRuleDetail) bool {
	if mr1 == nil || mr2 == nil {
		return false
	}

	if mr1.Name != mr2.Name {
		return false
	}

	// compare enable
	if mr1.IsEnabled() != mr2.IsEnabled() {
		return false
	}

	// compare detect
	if !cmp.Equal(mr1.Detect, mr2.Detect) {
		return false
	}

	return true
}
