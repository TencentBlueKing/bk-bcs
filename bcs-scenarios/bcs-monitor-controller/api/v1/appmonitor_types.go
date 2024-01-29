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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Rule 告警规则
type Rule struct {
	Rule        string     `json:"rule" yaml:"rule,omitempty"`
	Threshold   *Algorithm `json:"threshold,omitempty" yaml:"threshold,omitempty"`
	NoticeGroup []string   `json:"noticeGroup,omitempty" yaml:"noticeGroup,omitempty"`
	Trigger     string     `json:"trigger,omitempty" yaml:"trigger,omitempty"` // 触发配置，如 1/5/6表示5个周期内满足1次则告警，连续6个周期内不满足条件则表示恢复
}

// NoticeGroupConfig 告警组配置
type NoticeGroupConfig struct {
	Group     string   `json:"group,omitempty" yaml:"group,omitempty"`
	Member    []string `json:"member,omitempty" yaml:"member,omitempty"`
	VoiceWhen []string `json:"voice_when,omitempty" yaml:"voiceWhen,omitempty"`
	Robot     *Robot   `json:"robot,omitempty" yaml:"robot,omitempty"`
}

// Robot 机器人提示
type Robot struct {
	ChatID    string   `json:"chatID,omitempty" yaml:"chatID,omitempty"`
	RobotWhen []string `json:"robotWhen,omitempty" yaml:"robotWhen,omitempty"`
}

// DashBoard 监控面板名称
type DashBoard struct {
	Board string `json:"board,omitempty" yaml:"board,omitempty"`
}

// RuleEnhance 告警规则增强能力
type RuleEnhance struct {
	Rules []Rule `json:"rules,omitempty" yaml:"rules,omitempty"`

	NoticeGroupAppend  []string `json:"noticeGroupAppend,omitempty" yaml:"noticeGroupAppend,omitempty"`
	NoticeGroupReplace []string `json:"noticeGroupReplace,omitempty" yaml:"noticeGroupReplace,omitempty"`
	Trigger            string   `json:"trigger,omitempty" yaml:"trigger,omitempty"`
	WhereAdd           string   `json:"whereAdd,omitempty" yaml:"whereAdd,omitempty"`
	WhereOr            string   `json:"whereOr,omitempty" yaml:"whereOr,omitempty"`
	// if true, controller will ignore resource's change
	IgnoreChange bool `json:"ignoreChange,omitempty" yaml:"ignoreChange,omitempty"`
}

// NoticeGroupEnhance 告警组增强能力
type NoticeGroupEnhance struct {
	AppendNoticeGroups []*NoticeGroupDetail `json:"appendNoticeGroups,omitempty" yaml:"appendNoticeGroups,omitempty"`
	// if true, controller will ignore resource's change
	IgnoreChange bool `json:"ignoreChange,omitempty" yaml:"ignoreChange,omitempty"`
}

// DashBoardEnhance 监控面板增强能力
type DashBoardEnhance struct {
	DashBoards []DashBoard `json:"dashboards,omitempty" yaml:"dashboards,omitempty"`
	// if true, controller will ignore resource's change
	IgnoreChange bool `json:"ignoreChange,omitempty" yaml:"ignoreChange,omitempty"`
}

// RepoRef 允许用户自定义场景仓库
type RepoRef struct {
	URL            string `json:"url" yaml:"url"`
	TargetRevision string `json:"targetRevision,omitempty" yaml:"targetRevision,omitempty"`

	// no used
	UserName string `json:"userName,omitempty" yaml:"userName,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
}

// AppMonitorSpec defines the desired state of AppMonitor
type AppMonitorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Scenario string `json:"scenario" yaml:"scenario,omitempty"`
	BizId    string `json:"bizID" yaml:"bizId,omitempty"`
	BizToken string `json:"bizToken,omitempty" yaml:"bizToken,omitempty"`
	// if true, controller will ignore resource's change
	IgnoreChange bool `json:"ignoreChange,omitempty" yaml:"ignoreChange,omitempty"`
	// 是否覆盖同名配置，默认为false
	Override bool `json:"override,omitempty" yaml:"override,omitempty"`

	// if set, import Repo from argo
	RepoRef            *RepoRef            `json:"repoRef,omitempty" yaml:"repoRef,omitempty"`
	Labels             map[string]string   `json:"labels,omitempty" yaml:"labels,omitempty"`
	RuleEnhance        *RuleEnhance        `json:"ruleEnhance,omitempty" yaml:"ruleEnhance,omitempty"`
	NoticeGroupEnhance *NoticeGroupEnhance `json:"noticeGroupEnhance,omitempty" yaml:"noticeGroupEnhance,omitempty"`
	DashBoardEnhance   *DashBoardEnhance   `json:"dashBoardEnhance,omitempty" yaml:"dashBoardEnhance,omitempty"`
}

// AppMonitorStatus defines the observed state of AppMonitor
type AppMonitorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SyncStatus SyncStatus `json:"syncStatus"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.syncStatus.state`

// AppMonitor is the Schema for the appmonitors API
type AppMonitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AppMonitorSpec   `json:"spec,omitempty"`
	Status AppMonitorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AppMonitorList contains a list of AppMonitor
type AppMonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppMonitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AppMonitor{}, &AppMonitorList{})
}
