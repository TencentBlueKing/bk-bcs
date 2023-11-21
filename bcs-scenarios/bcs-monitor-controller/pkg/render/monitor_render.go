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

package render

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	monitorextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-monitor-controller/pkg/option"
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
	ReadScenario(scenario string) (*Result, error)
}

// MonitorRender render monitor
type MonitorRender struct {
	gitRepo *gitRepo

	decoder runtime.Decoder
}

// NewMonitorRender return new monitor render
func NewMonitorRender(scheme *runtime.Scheme, cli client.Client, opt *option.ControllerOption) (*MonitorRender, error) {
	gr, err := newGitRepo(cli, opt)
	if err != nil {
		return nil, err
	}
	return &MonitorRender{
		gitRepo: gr,
		decoder: serializer.NewCodecFactory(scheme).UniversalDeserializer(),
	}, nil
}

// Render transfer AppMonitor to child crd, i.e. Panel/MonitorRule/NoticeGroup, than fill them with AppMonitor info
func (r *MonitorRender) Render(appMonitor *monitorextensionv1.AppMonitor) (*Result, error) {
	if appMonitor == nil {
		return nil, fmt.Errorf("nil appMonitor")
	}

	rawResult, err := r.ReadScenario(appMonitor.Spec.Scenario)
	if err != nil {
		return nil, err
	}
	blog.Infof("load scenario'%s' success", appMonitor.Spec.Scenario)

	r.renderMonitorRule(appMonitor, rawResult)
	r.renderNoticeGroup(appMonitor, rawResult)
	r.renderPanel(appMonitor, rawResult)
	r.renderConfigMap(appMonitor, rawResult)

	blog.Infof("transfer scenario'%s' result success", appMonitor.Spec.Scenario)
	return rawResult, nil
}

func (r *MonitorRender) renderMonitorRule(appMonitor *monitorextensionv1.AppMonitor, rawResult *Result) {
	bizID := appMonitor.Spec.BizId
	bizToken := appMonitor.Spec.BizToken
	scenario := appMonitor.Spec.Scenario
	namespace := appMonitor.GetNamespace()
	ignoreChange := appMonitor.Spec.IgnoreChange
	override := appMonitor.Spec.Override
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
		if appMonitor.Spec.RuleEnhance != nil {
			mr.Spec.IgnoreChange = appMonitor.Spec.RuleEnhance.IgnoreChange || ignoreChange
		} else {
			mr.Spec.IgnoreChange = ignoreChange
		}
		mr.Status.SyncStatus.State = monitorextensionv1.SyncStateNeedReSync

		if appMonitor.Spec.RuleEnhance != nil {
			ruleEnhance := appMonitor.Spec.RuleEnhance
			for _, rawRule := range mr.Spec.Rules {
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
					for _, query := range rawRule.Query.QueryConfigs {
						if query.Where == "" {
							query.Where = ruleEnhance.WhereAdd
						} else {
							query.Where = fmt.Sprintf("%s and %s", query.Where, ruleEnhance.WhereAdd)
						}
					}
				}
				if ruleEnhance.WhereOr != "" {
					for _, query := range rawRule.Query.QueryConfigs {
						if query.Where == "" {
							query.Where = ruleEnhance.WhereAdd
						} else {
							query.Where = fmt.Sprintf("%s or %s", query.Where, ruleEnhance.WhereAdd)
						}
					}
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
				}
			}
		}

	}
}

func (r *MonitorRender) renderNoticeGroup(appMonitor *monitorextensionv1.AppMonitor, rawResult *Result) {
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
}

func (r *MonitorRender) renderPanel(appMonitor *monitorextensionv1.AppMonitor, rawResult *Result) {
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
						boardList = append(boardList, rawBoard)
						break
					}
				}
				if !find {
					blog.Warnf("unknown board '%s'", board)
				}
			}
		}
	}
}

func (r *MonitorRender) renderConfigMap(appMonitor *monitorextensionv1.AppMonitor, rawResult *Result) {
	for _, cm := range rawResult.ConfigMaps {
		cm.SetNamespace(appMonitor.GetNamespace())
		cm.SetLabels(map[string]string{
			monitorextensionv1.LabelKeyForScenarioName: appMonitor.Spec.Scenario,
			// monitorextensionv1.LabelKeyForAppMonitorName: appMonitor.Name,
			monitorextensionv1.LabelKeyForResourceType: monitorextensionv1.LabelValueResourceTypeConfigMap,
		})
	}
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
