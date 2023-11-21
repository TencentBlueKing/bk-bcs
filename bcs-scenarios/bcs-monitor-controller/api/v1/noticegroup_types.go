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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NoticeAction 执行通知
type NoticeAction struct {
	Execute        *NoticeType `json:"execute" yaml:"execute"`
	ExecuteFailed  *NoticeType `json:"execute_failed" yaml:"execute_failed"`
	ExecuteSuccess *NoticeType `json:"execute_success" yaml:"execute_success"`
}

// NoticeAlert 告警通知
type NoticeAlert struct {
	Fatal   *NoticeType `json:"fatal" yaml:"fatal"`
	Remind  *NoticeType `json:"remind" yaml:"remind"`
	Warning *NoticeType `json:"warning" yaml:"warning"`
}

// NoticeType 通知方式
type NoticeType struct {
	Type    []string `json:"type" yaml:"type"`
	ChatIds []string `json:"chatids,omitempty" yaml:"chatids,omitempty"`
}

// NoticeGroupDetail 告警组配置
type NoticeGroupDetail struct {
	Action map[string]*NoticeAction `json:"action,omitempty" yaml:"action,omitempty"` // 执行通知
	Alert  map[string]*NoticeAlert  `json:"alert,omitempty" yaml:"alert,omitempty"`   // 告警通知
	Users  []string                 `json:"users,omitempty" yaml:"users,omitempty"`   // 通知对象
	Name   string                   `json:"name,omitempty" yaml:"name,omitempty"`     // 告警组名称
}

// NoticeGroupSpec defines the desired state of NoticeGroup
type NoticeGroupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	BizID    string `json:"bizID" yaml:"-"`
	BizToken string `json:"bizToken,omitempty" yaml:"-"`
	// if true, controller will ignore resource's change
	IgnoreChange bool `json:"ignoreChange,omitempty"`
	// 是否覆盖同名配置，默认为false
	Override bool `json:"override,omitempty"`

	Scenario string               `json:"scenario,omitempty"`
	Groups   []*NoticeGroupDetail `json:"groups" yaml:"groups"`
}

// NoticeGroupStatus defines the observed state of NoticeGroup
type NoticeGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	SyncStatus SyncStatus `json:"syncStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.syncStatus.state`

// NoticeGroup is the Schema for the noticegroups API
type NoticeGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NoticeGroupSpec   `json:"spec,omitempty"`
	Status NoticeGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NoticeGroupList contains a list of NoticeGroup
type NoticeGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NoticeGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NoticeGroup{}, &NoticeGroupList{})
}
