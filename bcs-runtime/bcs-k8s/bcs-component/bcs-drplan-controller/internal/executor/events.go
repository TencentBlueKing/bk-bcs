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

package executor

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// Event types
const (
	EventTypeNormal  = corev1.EventTypeNormal
	EventTypeWarning = corev1.EventTypeWarning
)

// Event reasons for DRPlanExecution
const (
	// Execution events
	EventReasonExecutionStarted   = "ExecutionStarted"
	EventReasonExecutionSucceeded = "ExecutionSucceeded"
	EventReasonExecutionFailed    = "ExecutionFailed"
	EventReasonExecutionCancelled = "ExecutionCancelled"

	// Stage events
	EventReasonStageStarted   = "StageStarted"
	EventReasonStageSucceeded = "StageSucceeded"
	EventReasonStageFailed    = "StageFailed"

	// Workflow events
	EventReasonWorkflowStarted   = "WorkflowStarted"
	EventReasonWorkflowSucceeded = "WorkflowSucceeded"
	EventReasonWorkflowFailed    = "WorkflowFailed"

	// Action events
	EventReasonActionStarted   = "ActionStarted"
	EventReasonActionSucceeded = "ActionSucceeded"
	EventReasonActionFailed    = "ActionFailed"
	EventReasonActionRetrying  = "ActionRetrying"

	// Revert events
	EventReasonRevertStarted   = "RevertStarted"
	EventReasonRevertSucceeded = "RevertSucceeded"
	EventReasonRevertFailed    = "RevertFailed"
)

// EventRecorder wraps Kubernetes event recorder with DR-specific methods
type EventRecorder struct {
	recorder record.EventRecorder
}

// NewEventRecorder creates a new event recorder
func NewEventRecorder(recorder record.EventRecorder) *EventRecorder {
	return &EventRecorder{recorder: recorder}
}

// ExecutionStarted records an execution started event
func (r *EventRecorder) ExecutionStarted(execution *drv1alpha1.DRPlanExecution, planName string) {
	msg := fmt.Sprintf("Started executing plan %s (operation=%s)", planName, execution.Spec.OperationType)
	r.recorder.Event(execution, EventTypeNormal, EventReasonExecutionStarted, msg)
	klog.Infof("Event: %s - %s", EventReasonExecutionStarted, msg)
}

// ExecutionSucceeded records an execution succeeded event
func (r *EventRecorder) ExecutionSucceeded(execution *drv1alpha1.DRPlanExecution, summary string) {
	msg := fmt.Sprintf("Execution succeeded: %s", summary)
	r.recorder.Event(execution, EventTypeNormal, EventReasonExecutionSucceeded, msg)
	klog.Infof("Event: %s - %s", EventReasonExecutionSucceeded, msg)
}

// ExecutionFailed records an execution failed event
func (r *EventRecorder) ExecutionFailed(execution *drv1alpha1.DRPlanExecution, reason string) {
	msg := fmt.Sprintf("Execution failed: %s", reason)
	r.recorder.Event(execution, EventTypeWarning, EventReasonExecutionFailed, msg)
	klog.Warningf("Event: %s - %s", EventReasonExecutionFailed, msg)
}

// ExecutionCancelled records an execution cancelled event
func (r *EventRecorder) ExecutionCancelled(execution *drv1alpha1.DRPlanExecution) {
	msg := "Execution cancelled by user"
	r.recorder.Event(execution, EventTypeWarning, EventReasonExecutionCancelled, msg)
	klog.Infof("Event: %s - %s", EventReasonExecutionCancelled, msg)
}

// StageStarted records a stage started event
func (r *EventRecorder) StageStarted(execution *drv1alpha1.DRPlanExecution, stageName string, parallel bool) {
	mode := "sequential"
	if parallel {
		mode = "parallel"
	}
	msg := fmt.Sprintf("Started executing stage %s (mode=%s)", stageName, mode)
	r.recorder.Event(execution, EventTypeNormal, EventReasonStageStarted, msg)
	klog.Infof("Event: %s - %s", EventReasonStageStarted, msg)
}

// StageSucceeded records a stage succeeded event
func (r *EventRecorder) StageSucceeded(execution *drv1alpha1.DRPlanExecution, stageName string, duration string) {
	msg := fmt.Sprintf("Stage %s succeeded (duration=%s)", stageName, duration)
	r.recorder.Event(execution, EventTypeNormal, EventReasonStageSucceeded, msg)
	klog.Infof("Event: %s - %s", EventReasonStageSucceeded, msg)
}

// StageFailed records a stage failed event
func (r *EventRecorder) StageFailed(execution *drv1alpha1.DRPlanExecution, stageName string, reason string) {
	msg := fmt.Sprintf("Stage %s failed: %s", stageName, reason)
	r.recorder.Event(execution, EventTypeWarning, EventReasonStageFailed, msg)
	klog.Warningf("Event: %s - %s", EventReasonStageFailed, msg)
}

// WorkflowStarted records a workflow started event
func (r *EventRecorder) WorkflowStarted(execution *drv1alpha1.DRPlanExecution, workflowName string) {
	msg := fmt.Sprintf("Started executing workflow %s", workflowName)
	r.recorder.Event(execution, EventTypeNormal, EventReasonWorkflowStarted, msg)
	klog.V(4).Infof("Event: %s - %s", EventReasonWorkflowStarted, msg)
}

// WorkflowSucceeded records a workflow succeeded event
func (r *EventRecorder) WorkflowSucceeded(execution *drv1alpha1.DRPlanExecution, workflowName string, duration string) {
	msg := fmt.Sprintf("Workflow %s succeeded (duration=%s)", workflowName, duration)
	r.recorder.Event(execution, EventTypeNormal, EventReasonWorkflowSucceeded, msg)
	klog.V(4).Infof("Event: %s - %s", EventReasonWorkflowSucceeded, msg)
}

// WorkflowFailed records a workflow failed event
func (r *EventRecorder) WorkflowFailed(execution *drv1alpha1.DRPlanExecution, workflowName string, reason string) {
	msg := fmt.Sprintf("Workflow %s failed: %s", workflowName, reason)
	r.recorder.Event(execution, EventTypeWarning, EventReasonWorkflowFailed, msg)
	klog.Warningf("Event: %s - %s", EventReasonWorkflowFailed, msg)
}

// ActionStarted records an action started event
func (r *EventRecorder) ActionStarted(execution *drv1alpha1.DRPlanExecution, actionName, actionType string) {
	msg := fmt.Sprintf("Started executing action %s (type=%s)", actionName, actionType)
	r.recorder.Event(execution, EventTypeNormal, EventReasonActionStarted, msg)
	klog.V(4).Infof("Event: %s - %s", EventReasonActionStarted, msg)
}

// ActionSucceeded records an action succeeded event
func (r *EventRecorder) ActionSucceeded(execution *drv1alpha1.DRPlanExecution, actionName string) {
	msg := fmt.Sprintf("Action %s succeeded", actionName)
	r.recorder.Event(execution, EventTypeNormal, EventReasonActionSucceeded, msg)
	klog.V(4).Infof("Event: %s - %s", EventReasonActionSucceeded, msg)
}

// ActionFailed records an action failed event
func (r *EventRecorder) ActionFailed(execution *drv1alpha1.DRPlanExecution, actionName string, reason string) {
	msg := fmt.Sprintf("Action %s failed: %s", actionName, reason)
	r.recorder.Event(execution, EventTypeWarning, EventReasonActionFailed, msg)
	klog.Warningf("Event: %s - %s", EventReasonActionFailed, msg)
}

// ActionRetrying records an action retrying event
func (r *EventRecorder) ActionRetrying(execution *drv1alpha1.DRPlanExecution, actionName string, attempt int32) {
	msg := fmt.Sprintf("Action %s retrying (attempt=%d)", actionName, attempt)
	r.recorder.Event(execution, EventTypeNormal, EventReasonActionRetrying, msg)
	klog.V(4).Infof("Event: %s - %s", EventReasonActionRetrying, msg)
}

// RevertStarted records a revert started event
func (r *EventRecorder) RevertStarted(execution *drv1alpha1.DRPlanExecution, planName string) {
	msg := fmt.Sprintf("Started reverting plan %s", planName)
	r.recorder.Event(execution, EventTypeNormal, EventReasonRevertStarted, msg)
	klog.Infof("Event: %s - %s", EventReasonRevertStarted, msg)
}

// RevertSucceeded records a revert succeeded event
func (r *EventRecorder) RevertSucceeded(execution *drv1alpha1.DRPlanExecution) {
	msg := "Revert succeeded"
	r.recorder.Event(execution, EventTypeNormal, EventReasonRevertSucceeded, msg)
	klog.Infof("Event: %s - %s", EventReasonRevertSucceeded, msg)
}

// RevertFailed records a revert failed event
func (r *EventRecorder) RevertFailed(execution *drv1alpha1.DRPlanExecution, reason string) {
	msg := fmt.Sprintf("Revert failed: %s", reason)
	r.recorder.Event(execution, EventTypeWarning, EventReasonRevertFailed, msg)
	klog.Warningf("Event: %s - %s", EventReasonRevertFailed, msg)
}

// RecordEvent records a generic event
func (r *EventRecorder) RecordEvent(object runtime.Object, eventType, reason, message string) {
	r.recorder.Event(object, eventType, reason, message)
	klog.V(4).Infof("Event: %s [%s] - %s", reason, eventType, message)
}
