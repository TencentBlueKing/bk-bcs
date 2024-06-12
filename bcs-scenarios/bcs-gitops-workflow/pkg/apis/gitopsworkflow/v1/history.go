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

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowHistory is the Schema for the workflow API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="NUM",type=integer,JSONPath=".status.historyNum"
// +kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="WORKFLOW-TRIGGER",type=boolean,JSONPath=".spec.triggerByWorkflow"
// +kubebuilder:printcolumn:name="TRIGGER",type=string,JSONPath=".spec.triggerType"
// +kubebuilder:printcolumn:name="HISTORY-ID",type=string,JSONPath=".status.historyID"
type WorkflowHistory struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowHistorySpec   `json:"spec,omitempty"`
	Status WorkflowHistoryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowHistoryList contains a list of Terraform
type WorkflowHistoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of secret objects.
	Items []WorkflowHistory `json:"items"`
}

// WorkflowHistorySpec is the specification details for workflow history
type WorkflowHistorySpec struct {
	TriggerByWorkflow bool        `json:"triggerByWorkflow,omitempty"`
	TriggerType       string      `json:"triggerType,omitempty"`
	Params            []Parameter `json:"params,omitempty"`
}

// HistoryStatus defines the status of history
type HistoryStatus string

const (
	// HistoryRunning running status
	HistoryRunning HistoryStatus = "Running"
	// HistorySuccess succeed status
	HistorySuccess HistoryStatus = "Succeed"
	// HistoryFailed failed status
	HistoryFailed HistoryStatus = "Failed"
	// HistoryError error status
	HistoryError HistoryStatus = "Error"
)

// WorkflowHistoryStatus defines the status of workflow history
type WorkflowHistoryStatus struct {
	Phase      HistoryStatus `json:"phase,omitempty"`
	Message    string        `json:"message,omitempty"`
	HistoryNum int64         `json:"historyNum,omitempty"`
	HistoryID  string        `json:"historyID,omitempty"`

	StartedAt  *metav1.Time `json:"startedAt,omitempty"`
	FinishedAt *metav1.Time `json:"finishedAt,omitempty"`
	CostTime   string       `json:"costTime,omitempty"`
}
