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

// Package render xxx
package render

import (
	"crypto/md5"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	monitorextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/repo"
)

// Result transfer AppMonitor to Sub CR
type Result struct {
	MonitorRule []*monitorextensionv1.MonitorRule
	NoticeGroup []*monitorextensionv1.NoticeGroup
	Panel       []*monitorextensionv1.Panel
	ConfigMaps  []*v1.ConfigMap
}

// IRender do transfer
type IRender interface {
	Render(appMonitor *monitorextensionv1.AppMonitor) (*Result, error)

	// ReadScenario return related resource of scenario
	ReadScenario(repoKey, scenario string) (*Result, error)
}

// MonitorRender render monitor
type MonitorRender struct {
	// gitRepo *gitRepo
	repoManager *repo.Manager
	decoder     runtime.Decoder
}

// NewMonitorRender return new monitor render
func NewMonitorRender(scheme *runtime.Scheme, cli client.Client, repoManager *repo.Manager,
	opt *option.ControllerOption) (*MonitorRender, error) {
	return &MonitorRender{
		repoManager: repoManager,
		decoder:     serializer.NewCodecFactory(scheme).UniversalDeserializer(),
	}, nil
}

// Render transfer AppMonitor to child crd, i.e. Panel/MonitorRule/NoticeGroup, than fill them with AppMonitor info
func (r *MonitorRender) Render(appMonitor *monitorextensionv1.AppMonitor) (*Result, error) {
	if appMonitor == nil {
		return nil, fmt.Errorf("nil appMonitor")
	}

	rawResult, err := r.ReadScenario(repo.GenRepoKeyFromAppMonitor(appMonitor), appMonitor.Spec.Scenario)
	if err != nil {
		return nil, err
	}
	blog.Infof("load scenario'%s' success", appMonitor.Spec.Scenario)

	renderedResult := &Result{
		MonitorRule: r.renderMonitorRule(appMonitor, rawResult),
		NoticeGroup: r.renderNoticeGroup(appMonitor, rawResult),
		Panel:       r.renderPanel(appMonitor, rawResult),
		ConfigMaps:  r.renderConfigMap(appMonitor, rawResult),
	}

	blog.Infof("transfer scenario'%s' result success", appMonitor.Spec.Scenario)
	return renderedResult, nil
}

// nolint
func (r *MonitorRender) renderMonitorRule(appMonitor *monitorextensionv1.AppMonitor,
	rawResult *Result) []*monitorextensionv1.MonitorRule {
	bizID := appMonitor.Spec.BizId
	bizToken := appMonitor.Spec.BizToken
	scenario := appMonitor.Spec.Scenario
	namespace := appMonitor.GetNamespace()
	ignoreChange := appMonitor.Spec.IgnoreChange
	override := appMonitor.Spec.Override
	conflictHandle := appMonitor.Spec.ConflictHandle
	rawMrs := make([]*monitorextensionv1.MonitorRule, 0)
	renderedMrs := make([]*monitorextensionv1.MonitorRule, 0)
	// render monitor rule
	for _, mr := range rawResult.MonitorRule {
		mr.SetNamespace(namespace)
		mr.SetName(genName(bizID, scenario, mr.GetName()))
		mr.SetLabels(map[string]string{
			monitorextensionv1.LabelKeyForBizID:          bizID,
			monitorextensionv1.LabelKeyForScenarioName:   scenario,
			monitorextensionv1.LabelKeyForAppMonitorName: appMonitor.Name,
			monitorextensionv1.LabelKeyForResourceType:   monitorextensionv1.LabelValueResourceTypeMonitorRule,
		})
		mr.Spec.BizID = bizID
		mr.Spec.BizToken = bizToken
		mr.Spec.Scenario = scenario
		mr.Spec.Override = override
		mr.Spec.ConflictHandle = conflictHandle
		if appMonitor.Spec.RuleEnhance != nil {
			mr.Spec.IgnoreChange = appMonitor.Spec.RuleEnhance.IgnoreChange || ignoreChange
		} else {
			mr.Spec.IgnoreChange = ignoreChange
		}
		mr.Status.SyncStatus.State = monitorextensionv1.SyncStateNeedReSync

		for _, rawRule := range mr.Spec.Rules {
			rawRule.Labels = appendLabels(rawRule.Labels)

			if appMonitor.Spec.RuleEnhance != nil {
				ruleEnhance := appMonitor.Spec.RuleEnhance
				if len(ruleEnhance.NoticeGroupReplace) != 0 {
					rawRule.Notice.UserGroups = ruleEnhance.NoticeGroupReplace
				}
				if len(ruleEnhance.NoticeGroupAppend) != 0 {
					rawRule.Notice.UserGroups = mergeStringList(rawRule.Notice.UserGroups, ruleEnhance.NoticeGroupAppend)
				}
				if ruleEnhance.Trigger != "" {
					rawRule.Detect.Trigger = ruleEnhance.Trigger
				}
				if ruleEnhance.WhereAdd != "" {
					rawRule.WhereAdd(ruleEnhance.WhereAdd)
				}
				if ruleEnhance.WhereOr != "" {
					rawRule.WhereOr(ruleEnhance.WhereOr)
				}

				// 单独指定的Rule优先级更高
				for _, rule := range ruleEnhance.Rules {
					if rawRule.Name != rule.Rule {
						continue
					}

					if rule.Threshold != nil {
						rawRule.Detect.Algorithm = rule.Threshold
					}
					if rule.Trigger != "" {
						rawRule.Detect.Trigger = rule.Trigger
					}
					if len(rule.NoticeGroup) != 0 {
						rawRule.Notice.UserGroups = rule.NoticeGroup
					}
					if rule.WhereAdd != "" {
						rawRule.WhereAdd(rule.WhereAdd)
					}
					if rule.WhereOr != "" {
						rawRule.WhereOr(rule.WhereOr)
					}
				}
			}

		}

		rawMrs = append(rawMrs, mr)
		renderedMrs = append(renderedMrs, mr)
	}

	renderedMrs = append(renderedMrs, r.generateCopyRules(appMonitor, rawMrs)...)

	return renderedMrs
}

func (r *MonitorRender) generateCopyRules(appMonitor *monitorextensionv1.AppMonitor,
	rawMrs []*monitorextensionv1.MonitorRule) []*monitorextensionv1.MonitorRule {
	cpMrList := make([]*monitorextensionv1.MonitorRule, 0)
	if appMonitor.Spec.RuleEnhance != nil {
		for _, cpConfig := range appMonitor.Spec.RuleEnhance.CopyRules {
			for _, mr := range rawMrs {
				cpMr := mr.DeepCopy()

				// NOCC:gas/crypto(误报 未使用于密钥)
				// nolint
				cpMr.SetName(fmt.Sprintf("cp-%x%s", md5.Sum([]byte(cpConfig.NamePrefix+cpConfig.NameSuffix)),
					mr.GetName()))
				for _, rule := range cpMr.Spec.Rules {
					rule.Name = fmt.Sprintf("%s%s%s", cpConfig.NamePrefix, cpConfig.NameSuffix, rule.Name)
					rule.WhereAdd(cpConfig.WhereAdd)
					rule.WhereOr(cpConfig.WhereOr)

					if len(cpConfig.NoticeGroupReplace) != 0 {
						rule.Notice.UserGroups = cpConfig.NoticeGroupReplace
					}
					if len(cpConfig.NoticeGroupAppend) != 0 {
						rule.Notice.UserGroups = mergeStringList(rule.Notice.UserGroups, cpConfig.NoticeGroupAppend)
					}
				}
				cpMrList = append(cpMrList, cpMr)
			}
		}
	}
	return cpMrList
}

func (r *MonitorRender) renderNoticeGroup(appMonitor *monitorextensionv1.AppMonitor,
	rawResult *Result) []*monitorextensionv1.NoticeGroup {
	bizID := appMonitor.Spec.BizId
	bizToken := appMonitor.Spec.BizToken
	scenario := appMonitor.Spec.Scenario
	namespace := appMonitor.GetNamespace()
	ignoreChange := appMonitor.Spec.IgnoreChange
	override := appMonitor.Spec.Override
	// add append notice group
	if appMonitor.Spec.NoticeGroupEnhance != nil {
		ngEnhance := appMonitor.Spec.NoticeGroupEnhance
		ng := monitorextensionv1.NoticeGroup{}
		ng.SetName(GenerateAppendNoticeGroupName)
		ng.Spec.Groups = append(ng.Spec.Groups, ngEnhance.AppendNoticeGroups...)

		rawResult.NoticeGroup = append(rawResult.NoticeGroup, &ng)
	}
	// DOTO transfer noticeGroup
	for _, ng := range rawResult.NoticeGroup {
		ng.SetNamespace(namespace)
		ng.SetName(genName(bizID, scenario, ng.GetName()))
		ng.SetLabels(map[string]string{
			monitorextensionv1.LabelKeyForBizID:          bizID,
			monitorextensionv1.LabelKeyForScenarioName:   scenario,
			monitorextensionv1.LabelKeyForAppMonitorName: appMonitor.Name,
			monitorextensionv1.LabelKeyForResourceType:   monitorextensionv1.LabelValueResourceTypeNoticeGroup,
		})
		ng.Spec.BizID = bizID
		ng.Spec.BizToken = bizToken
		ng.Spec.Scenario = scenario
		ng.Spec.Override = override
		if appMonitor.Spec.NoticeGroupEnhance != nil {
			ng.Spec.IgnoreChange = appMonitor.Spec.NoticeGroupEnhance.IgnoreChange || ignoreChange
		} else {
			ng.Spec.IgnoreChange = ignoreChange
		}
		ng.Status.SyncStatus.State = monitorextensionv1.SyncStateNeedReSync
	}

	return rawResult.NoticeGroup
}

func (r *MonitorRender) renderPanel(appMonitor *monitorextensionv1.AppMonitor,
	rawResult *Result) []*monitorextensionv1.Panel {
	bizID := appMonitor.Spec.BizId
	bizToken := appMonitor.Spec.BizToken
	scenario := appMonitor.Spec.Scenario
	namespace := appMonitor.GetNamespace()
	ignoreChange := appMonitor.Spec.IgnoreChange
	override := appMonitor.Spec.Override
	for _, panel := range rawResult.Panel {
		panel.SetNamespace(namespace)
		panel.SetName(genName(bizID, scenario, panel.GetName()))
		panel.SetLabels(map[string]string{
			monitorextensionv1.LabelKeyForBizID:          bizID,
			monitorextensionv1.LabelKeyForScenarioName:   scenario,
			monitorextensionv1.LabelKeyForAppMonitorName: appMonitor.Name,
			monitorextensionv1.LabelKeyForResourceType:   monitorextensionv1.LabelValueResourceTypePanel,
		})
		panel.Spec.BizID = bizID
		panel.Spec.BizToken = bizToken
		panel.Spec.Scenario = scenario
		panel.Spec.Override = override
		if appMonitor.Spec.DashBoardEnhance != nil {
			panel.Spec.IgnoreChange = appMonitor.Spec.DashBoardEnhance.IgnoreChange || ignoreChange
		} else {
			panel.Spec.IgnoreChange = ignoreChange
		}
		panel.Status.SyncStatus.State = monitorextensionv1.SyncStateNeedReSync

		// DOTO 用户应该可以设置使用哪些board
		if appMonitor.Spec.DashBoardEnhance != nil {
			boardList := make([]monitorextensionv1.DashBoardConfig, 0)
			for _, board := range appMonitor.Spec.DashBoardEnhance.DashBoards {
				find := false
				for _, rawBoard := range panel.Spec.DashBoard {
					if rawBoard.Board == board.Board {
						find = true
						boardList = append(boardList, rawBoard) // nolint never used
						break
					}
				}
				if !find {
					blog.Warnf("unknown board '%s'", board)
				}
			}
		}
	}
	return rawResult.Panel
}

func (r *MonitorRender) renderConfigMap(appMonitor *monitorextensionv1.AppMonitor, rawResult *Result) []*v1.ConfigMap {
	for _, cm := range rawResult.ConfigMaps {
		cm.SetNamespace(appMonitor.GetNamespace())
		cm.SetLabels(map[string]string{
			monitorextensionv1.LabelKeyForScenarioName: appMonitor.Spec.Scenario,
			// monitorextensionv1.LabelKeyForAppMonitorName: appMonitor.Name,
			monitorextensionv1.LabelKeyForResourceType: monitorextensionv1.LabelValueResourceTypeConfigMap,
		})
	}
	return rawResult.ConfigMaps
}

// mergeStringList merge string and remove duplicates
func mergeStringList(arr1, arr2 []string) []string {
	keys := make(map[string]bool)
	var uniqueArray []string

	for _, entry := range arr1 {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			uniqueArray = append(uniqueArray, entry)
		}
	}

	for _, entry := range arr2 {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			uniqueArray = append(uniqueArray, entry)
		}
	}

	return uniqueArray
}
