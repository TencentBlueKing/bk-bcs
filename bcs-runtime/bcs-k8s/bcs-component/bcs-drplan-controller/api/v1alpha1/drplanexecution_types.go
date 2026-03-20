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

// DRPlanExecutionSpec defines the desired state of DRPlanExecution
type DRPlanExecutionSpec struct {
	// PlanRef is the associated DRPlan name
	// +kubebuilder:validation:Required
	PlanRef string `json:"planRef"`

	// OperationType is the operation type: Execute, Revert
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Execute;Revert
	OperationType string `json:"operationType"`

	// RevertExecutionRef specifies which execution to revert (required for Revert operation).
	// Must reference an existing DRPlanExecution with operationType=Execute and phase=Succeeded.
	// This ensures precise control over which execution to rollback.
	// +optional
	RevertExecutionRef string `json:"revertExecutionRef,omitempty"`
}

// DRPlanExecutionStatus defines the observed state of DRPlanExecution.
type DRPlanExecutionStatus struct {
	// Phase is the execution phase: Pending, Running, Succeeded, Failed, Cancelled
	// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Cancelled;Unknown
	// +optional
	Phase string `json:"phase,omitempty"`

	// StartTime is the start time
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is the completion time
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// StageStatuses are the status of each stage
	// +optional
	StageStatuses []StageStatus `json:"stageStatuses,omitempty"`

	// Summary is the execution statistics summary
	// +optional
	Summary *ExecutionSummary `json:"summary,omitempty"`

	// Message is the status message
	// +optional
	Message string `json:"message,omitempty"`

	// Conditions represent the current state of the DRPlanExecution resource
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// StageStatus defines the status of a stage execution
type StageStatus struct {
	// Name is the stage name
	Name string `json:"name"`

	// Phase is the stage phase: Pending, Running, Succeeded, Failed, Skipped
	// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Skipped
	Phase string `json:"phase"`

	// Parallel indicates whether this stage executes workflows in parallel
	// +optional
	Parallel bool `json:"parallel,omitempty"`

	// DependsOn lists the stage dependencies
	// +optional
	DependsOn []string `json:"dependsOn,omitempty"`

	// StartTime is the start time
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is the completion time
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Duration is the execution duration
	// +optional
	Duration string `json:"duration,omitempty"`

	// Message is the status message
	// +optional
	Message string `json:"message,omitempty"`

	// WorkflowExecutions are the status of each workflow
	// +optional
	WorkflowExecutions []WorkflowExecutionStatus `json:"workflowExecutions,omitempty"`
}

// WorkflowExecutionStatus defines the status of a workflow execution
type WorkflowExecutionStatus struct {
	// WorkflowRef is the workflow reference
	WorkflowRef ObjectReference `json:"workflowRef"`

	// Phase is the workflow phase: Pending, Running, Succeeded, Failed, Skipped
	// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Skipped
	Phase string `json:"phase"`

	// StartTime is the start time
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is the completion time
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Duration is the execution duration
	// +optional
	Duration string `json:"duration,omitempty"`

	// Progress is the progress information (e.g., "2/5 actions completed")
	// +optional
	Progress string `json:"progress,omitempty"`

	// CurrentAction is the currently executing action name
	// +optional
	CurrentAction string `json:"currentAction,omitempty"`

	// Message is the status message
	// +optional
	Message string `json:"message,omitempty"`

	// ActionStatuses are the detailed action execution statuses
	// +optional
	ActionStatuses []ActionStatus `json:"actionStatuses,omitempty"`
}

// ExecutionSummary defines execution statistics
type ExecutionSummary struct {
	// TotalStages is the total number of stages
	// +optional
	TotalStages int `json:"totalStages,omitempty"`

	// CompletedStages is the number of completed stages
	// +optional
	CompletedStages int `json:"completedStages,omitempty"`

	// RunningStages is the number of running stages
	// +optional
	RunningStages int `json:"runningStages,omitempty"`

	// PendingStages is the number of pending stages
	// +optional
	PendingStages int `json:"pendingStages,omitempty"`

	// FailedStages is the number of failed stages
	// +optional
	FailedStages int `json:"failedStages,omitempty"`

	// TotalWorkflows is the total number of workflows
	// +optional
	TotalWorkflows int `json:"totalWorkflows,omitempty"`

	// CompletedWorkflows is the number of completed workflows
	// +optional
	CompletedWorkflows int `json:"completedWorkflows,omitempty"`

	// RunningWorkflows is the number of running workflows
	// +optional
	RunningWorkflows int `json:"runningWorkflows,omitempty"`

	// PendingWorkflows is the number of pending workflows
	// +optional
	PendingWorkflows int `json:"pendingWorkflows,omitempty"`

	// FailedWorkflows is the number of failed workflows
	// +optional
	FailedWorkflows int `json:"failedWorkflows,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// DRPlanExecution is the Schema for the drplanexecutions API
type DRPlanExecution struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of DRPlanExecution
	// +required
	Spec DRPlanExecutionSpec `json:"spec"`

	// status defines the observed state of DRPlanExecution
	// +optional
	Status DRPlanExecutionStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// DRPlanExecutionList contains a list of DRPlanExecution
type DRPlanExecutionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []DRPlanExecution `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DRPlanExecution{}, &DRPlanExecutionList{})
}
