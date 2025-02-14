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

// Package patch xxx
package patch

import (
	"reflect"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/google/go-cmp/cmp"

	v1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/utils"
)

// ThreeWayMergeMonitorRule param current from bkm
// original: 上次变更的rule
// current： 当前bkm平台上的rule
// modified： 本次尝试变更的rule
func ThreeWayMergeMonitorRule(scenario string, original, current, modified []*v1.MonitorRuleDetail) []*v1.
	MonitorRuleDetail {
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
		currentRule, currentOk := currentMap[modifiedRule.Name]
		if !currentOk { // 用户删除了对应策略 or 新建策略
			mergeResult = append(mergeResult, modifiedRule)
			continue
		}

		originalRule, ok := originalMap[modifiedRule.Name]
		if !ok {
			if !currentOk { // 新增策略
				mergeResult = append(mergeResult, modifiedRule)
				continue
			} else if scenario == "bcs-cluster" || scenario == "test" {
				// 特殊适配， 仅针对bcs-cluster生效。 背景： bcs-cluster原本只包括集群相关的监控策略， 后续需要纳管组件相关策略
				// 纳管了其他场景时，需要保留原策略
				originalRule = modifiedRule
			}
		}

		// 策略：
		// 1. 如果用户对策略进行了 ： 开关、修改告警阈值（包括触发周期）  则保留用户的所有变更
		// 2. 如果用户更新了告警策略的通知方式， 则采用用户的告警通知， 其他字段以模板为准
		// 3. 其余情况， 统一以模板为准
		var mergeRule *v1.MonitorRuleDetail
		if cmp.Equal(originalRule, currentRule, cmp.Comparer(compareMonitorRule)) {
			blog.Infof("[%s]mergeRule..", originalRule.Name)
			// 相同说明用户没有在面板上进行修改
			mergeRule = modifiedRule
			// 告警组配置以用户设置为准
			mergeRule.Notice = mergeNoticeGroup(currentRule.Notice, modifiedRule.Notice)
			// label覆盖
			mergeRule.Labels = modifiedRule.Labels
		} else {
			blog.Infof("[%s]changed Rule..", originalRule.Name)
			if !reflect.DeepEqual(originalRule.Detect, currentRule.Detect) {
				// 用户修改了探测条件， 不做变更
				mergeRule = currentRule
			} else {
				// 目前只有可能是用户开关了告警规则， 保留开关， 但是更新内容
				mergeRule = modifiedRule
				mergeRule.Enabled = currentRule.Enabled
			}
			mergeRule.Notice = mergeNoticeGroup(currentRule.Notice, modifiedRule.Notice)
			// label覆盖
			mergeRule.Labels = modifiedRule.Labels
		}

		mergeResult = append(mergeResult, mergeRule)
	}

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
	if !reflect.DeepEqual(mr1.Detect, mr2.Detect) {
		return false
	}

	return true
}

func mergeNoticeGroup(currentRule *v1.Notice, modifiedRule *v1.Notice) *v1.Notice {
	result := currentRule.DeepCopy()
	// 仅merge告警组配置， 其他以用户配置为准

	result.UserGroups = utils.MergeStringList(currentRule.UserGroups, modifiedRule.UserGroups)
	return result
}
