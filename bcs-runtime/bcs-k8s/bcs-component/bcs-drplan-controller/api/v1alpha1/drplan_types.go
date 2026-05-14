/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2023 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DRPlanSpec defines the desired state of DRPlan
type DRPlanSpec struct {
	// Description is the plan description
	// +optional
	Description string `json:"description,omitempty"`

	// Stages are the stage orchestration list
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Stages []Stage `json:"stages"`

	// GlobalParams are global parameters passed to all workflows
	// +optional
	GlobalParams []Parameter `json:"globalParams,omitempty"`

	// FailurePolicy defines how to handle failures
	// +kubebuilder:validation:Enum=Stop;Continue
	// +kubebuilder:default=Stop
	// +optional
	FailurePolicy string `json:"failurePolicy,omitempty"`
}

// Stage defines a stage in the plan orchestration
type Stage struct {
	// Name is the unique stage name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Description is the stage description
	// +optional
	Description string `json:"description,omitempty"`

	// DependsOn lists stage names this stage depends on
	// +optional
	DependsOn []string `json:"dependsOn,omitempty"`

	// Parallel indicates whether to execute all workflows in this stage in parallel
	// +kubebuilder:default=false
	// +optional
	Parallel bool `json:"parallel,omitempty"`

	// Workflows are the workflow reference list
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems=1
	Workflows []WorkflowReference `json:"workflows"`

	// FailurePolicy is the stage-level failure policy (overrides global)
	// +kubebuilder:validation:Enum=Stop;Continue;FailFast
	// +optional
	FailurePolicy string `json:"failurePolicy,omitempty"`
}

// WorkflowReference defines a workflow reference with parameters
type WorkflowReference struct {
	// WorkflowRef is the workflow reference (name + namespace)
	// +kubebuilder:validation:Required
	WorkflowRef ObjectReference `json:"workflowRef"`

	// Params are parameters that override globalParams
	// +optional
	Params []Parameter `json:"params,omitempty"`
}

// DRPlanStatus defines the observed state of DRPlan.
type DRPlanStatus struct {
	// Phase is the plan phase: Ready, Executed, Invalid
	// +kubebuilder:validation:Enum=Ready;Executed;Invalid
	// +optional
	Phase string `json:"phase,omitempty"`

	// Conditions represent the current state of the DRPlan resource
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// LastExecutionTime is the last execution time
	// +optional
	LastExecutionTime *metav1.Time `json:"lastExecutionTime,omitempty"`

	// LastExecutionRef is the last execution record name
	// +optional
	LastExecutionRef string `json:"lastExecutionRef,omitempty"`

	// CurrentExecution is the currently running execution reference (for concurrency control)
	// +optional
	CurrentExecution *ObjectReference `json:"currentExecution,omitempty"`

	// LastProcessedTrigger is DEPRECATED and no longer used (annotation trigger removed).
	// Kept for backward compatibility, will be removed in future versions.
	// +optional
	LastProcessedTrigger string `json:"lastProcessedTrigger,omitempty"`

	// ExecutionHistory keeps a list of recent execution references (max 10, newest first)
	// +optional
	// +kubebuilder:validation:MaxItems=10
	ExecutionHistory []ExecutionRecord `json:"executionHistory,omitempty"`

	// ObservedGeneration reflects the generation observed by the controller
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// ExecutionRecord records a historical execution reference
type ExecutionRecord struct {
	// Name is the DRPlanExecution name
	// +required
	Name string `json:"name"`

	// Namespace is the DRPlanExecution namespace
	// +required
	Namespace string `json:"namespace"`

	// OperationType is the operation type: Execute, Revert
	// +required
	OperationType string `json:"operationType"`

	// Phase is the execution phase: Pending, Running, Succeeded, Failed
	// +optional
	Phase string `json:"phase,omitempty"`

	// StartTime is the execution start time
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is the execution completion time
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DRPlan is the Schema for the drplans API
type DRPlan struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of DRPlan
	// +required
	Spec DRPlanSpec `json:"spec"`

	// status defines the observed state of DRPlan
	// +optional
	Status DRPlanStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// DRPlanList contains a list of DRPlan
type DRPlanList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []DRPlan `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DRPlan{}, &DRPlanList{})
}
