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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DashBoardConfig 告警面板配置
type DashBoardConfig struct {
	Board  string `json:"board" yaml:"board"`
	Render bool   `json:"render,omitempty" yaml:"render"`

	// +kubebuilder:validation:OneOf
	// +optional
	ConfigMap string `json:"configMap,omitempty" yaml:"configMap"`
	// +optional
	ConfigMapNs string `json:"configMapNs,omitempty" yaml:"configMapNs"`
	// +kubebuilder:validation:OneOf
	// +optional
	Url string `json:"url,omitempty" yaml:"url"`
}

// PanelSpec defines the desired state of Panel
type PanelSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	BizID    string `json:"bizID" yaml:"-"`
	BizToken string `json:"bizToken,omitempty" yaml:"-"`
	// if true, controller will ignore resource's change
	IgnoreChange bool `json:"ignoreChange,omitempty"`
	// 是否覆盖同名配置，默认为false
	Override bool `json:"override,omitempty"`

	Scenario  string            `json:"scenario,omitempty"`
	DashBoard []DashBoardConfig `json:"dashBoard,omitempty" yaml:"dashBoard"`
}

// PanelStatus defines the observed state of Panel
type PanelStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	SyncStatus SyncStatus `json:"syncStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="status",type=string,JSONPath=`.status.syncStatus.state`

// Panel is the Schema for the panels API
type Panel struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PanelSpec   `json:"spec,omitempty"`
	Status PanelStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PanelList contains a list of Panel
type PanelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Panel `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Panel{}, &PanelList{})
}
