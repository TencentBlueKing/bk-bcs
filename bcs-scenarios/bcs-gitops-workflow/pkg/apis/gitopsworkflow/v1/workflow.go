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

// Workflow is the Schema for the workflow API
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PROJECT",type=string,JSONPath=".spec.project"
// +kubebuilder:printcolumn:name="ENGINE",type=string,JSONPath=".spec.engine"
// +kubebuilder:printcolumn:name="STATUS",type=string,JSONPath=".status.phase"
// +kubebuilder:printcolumn:name="PPLINE-NAME",type=string,JSONPath=".spec.name"
// +kubebuilder:printcolumn:name="PIPELINE-ID",type=string,JSONPath=".status.pipelineID"
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec,omitempty"`
	Status WorkflowStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WorkflowList contains a list of Terraform
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	// Items is a list of secret objects.
	Items []Workflow `json:"items"`
}

// WorkflowSpec is the specification details for workflow, users can define workflow
// by fill on the content.
type WorkflowSpec struct {
	// +optional
	Disable bool `json:"disable"`

	// engine defines the underlying pipeline implementation engine
	// +kubebuilder:default:=bkdevops
	// +kubebuilder:validation:Enum=bkdevops
	// +optional
	Engine string `json:"engine,omitempty"`

	// +optional
	DestroyOnDeletion bool `json:"destroyOnDeletion"`

	// the name of pipeline
	// +optional
	Name string `json:"name"`

	// +optional
	Desc string `json:"desc,omitempty"`

	// project defines the workflow belongs to which project
	// +optional
	Project string `json:"project"`

	// params are the global parameters of workflow, which user can custom changes
	// +optional
	Params []Parameter `json:"params,omitempty"`

	// stepTemplates are the template of step, should define all the steps this workflow
	// need in first
	// +optional
	StepTemplates []StepTemplate `json:"stepTemplates,omitempty"`

	// stages define the real execute orchestration checklist in order.
	// +optional
	Stages []Stage `json:"stages,omitempty"`
}

// Parameter is key-value for workflow
type Parameter struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// StepTemplate defines the step template
type StepTemplate struct {
	Name string `json:"name,omitempty"`

	// bkdevops: type:atomCode:version, such-as: marketBuild:bcscmd-new:1.*
	Plugin    string            `json:"plugin,omitempty"`
	Condition map[string]string `json:"condition,omitempty"`
	With      map[string]string `json:"with,omitempty"`
	Timeout   int64             `json:"timeout,omitempty"`
}

// Stage defines the stage
type Stage struct {
	Name      string            `json:"name,omitempty"`
	Disabled  bool              `json:"disabled,omitempty"`
	Timeout   int64             `json:"timeout,omitempty"`
	Condition map[string]string `json:"condition,omitempty"`

	// review configuration
	ReviewUsers       []string `json:"reviewUsers,omitempty"`
	ReviewMessage     string   `json:"reviewMessage,omitempty"`
	ReviewNotifyGroup []string `json:"reviewNotifyGroup,omitempty"`

	Jobs []Job `json:"jobs,omitempty"`
}

// Job defines the job
type Job struct {
	Name      string            `json:"name,omitempty"`
	Disable   bool              `json:"enable,omitempty"`
	Strategy  Strategy          `json:"strategy,omitempty"`
	Condition map[string]string `json:"condition,omitempty"`
	Timeout   int64             `json:"timeout,omitempty"`
	RunsOn    RunsOn            `json:"runsOn,omitempty"`
	Steps     []Step            `json:"steps,omitempty"`
}

// Strategy defines the strategy
type Strategy struct {
	Matrix      map[string][]string `json:"matrix,omitempty"`
	FastKill    bool                `json:"fastKill,omitempty"`
	MaxParallel int                 `json:"maxParallel,omitempty"`
}

// RunsOn defines the runsOn
type RunsOn struct {
	Image   string `json:"image,omitempty"`
	Version string `json:"version,omitempty"`
}

// Step defines the step
type Step struct {
	Name      string            `json:"name,omitempty"`
	Disable   bool              `json:"disable"`
	Template  string            `json:"template,omitempty"`
	Condition map[string]string `json:"condition,omitempty"`
	With      map[string]string `json:"with,omitempty"`
	Timeout   int64             `json:"timeout,omitempty"`
}

const (
	// InitializingStatus Initializing status
	InitializingStatus = "Initializing"
	// ErrorStatus error status
	ErrorStatus = "Error"
	// ReadyStatus ready status
	ReadyStatus = "Ready"
)

// WorkflowStatus defines the status of workflow
type WorkflowStatus struct {
	Phase string `json:"phase,omitempty"`

	PipelineID string `json:"pipelineID,omitempty"`

	Message string `json:"message,omitempty"`

	LastUpdateTime *metav1.Time `json:"lastUpdateTime,omitempty"`
}
