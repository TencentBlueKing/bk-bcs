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
	"context"
	"fmt"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// DefaultStageExecutor implements StageExecutor interface
type DefaultStageExecutor struct {
	client           client.Client
	workflowExecutor WorkflowExecutor
}

// NewStageExecutor creates a new DefaultStageExecutor
func NewStageExecutor(client client.Client, workflowExecutor WorkflowExecutor) *DefaultStageExecutor {
	return &DefaultStageExecutor{
		client:           client,
		workflowExecutor: workflowExecutor,
	}
}

// ExecuteStage executes a stage with support for parallel execution and dependencies
func (e *DefaultStageExecutor) ExecuteStage(ctx context.Context, _ *drv1alpha1.DRPlan, stage *drv1alpha1.Stage, params map[string]interface{}) (*drv1alpha1.StageStatus, error) {
	klog.Infof("Executing stage: %s (parallel=%v, workflows=%d)", stage.Name, stage.Parallel, len(stage.Workflows))

	stageStatus := &drv1alpha1.StageStatus{
		Name:               stage.Name,
		Phase:              "Running",
		Parallel:           stage.Parallel,
		DependsOn:          stage.DependsOn,
		StartTime:          &metav1.Time{Time: time.Now()},
		WorkflowExecutions: make([]drv1alpha1.WorkflowExecutionStatus, 0, len(stage.Workflows)),
	}

	var err error
	if stage.Parallel {
		err = e.executeParallel(ctx, stage, params, stageStatus)
	} else {
		err = e.executeSequential(ctx, stage, params, stageStatus)
	}

	// Update completion time and duration
	stageStatus.CompletionTime = &metav1.Time{Time: time.Now()}
	stageStatus.Duration = stageStatus.CompletionTime.Sub(stageStatus.StartTime.Time).String()

	// Determine final phase
	if err != nil {
		stageStatus.Phase = drv1alpha1.PhaseFailed
		stageStatus.Message = err.Error()
		klog.Errorf("Stage %s failed: %v", stage.Name, err)
	} else {
		// Check if all workflows succeeded
		allSucceeded := true
		for _, ws := range stageStatus.WorkflowExecutions {
			if ws.Phase != drv1alpha1.PhaseSucceeded {
				allSucceeded = false
				break
			}
		}
		if allSucceeded {
			stageStatus.Phase = drv1alpha1.PhaseSucceeded
			klog.Infof("Stage %s succeeded", stage.Name)
		} else {
			stageStatus.Phase = drv1alpha1.PhaseFailed
			stageStatus.Message = "One or more workflows failed"
			klog.Warningf("Stage %s completed with failures", stage.Name)
		}
	}

	return stageStatus, err
}

// executeSequential executes workflows sequentially
func (e *DefaultStageExecutor) executeSequential(ctx context.Context, stage *drv1alpha1.Stage, params map[string]interface{}, stageStatus *drv1alpha1.StageStatus) error {
	klog.V(4).Infof("Executing stage %s sequentially", stage.Name)

	for i, wfRef := range stage.Workflows {
		klog.Infof("Executing workflow %d/%d: %s/%s", i+1, len(stage.Workflows), wfRef.WorkflowRef.Namespace, wfRef.WorkflowRef.Name)

		// Get workflow
		workflow, err := e.getWorkflow(ctx, wfRef.WorkflowRef)
		if err != nil {
			return fmt.Errorf("failed to get workflow %s/%s: %w", wfRef.WorkflowRef.Namespace, wfRef.WorkflowRef.Name, err)
		}

		// Merge parameters: workflow default -> plan globalParams (value) -> stage params (value)
		workflowParams := e.mergeParams(workflow, params, wfRef.Params)

		// Execute workflow
		workflowStatus, err := e.workflowExecutor.ExecuteWorkflow(ctx, workflow, workflowParams)
		if err != nil {
			workflowStatus.Phase = drv1alpha1.PhaseFailed
			workflowStatus.Message = err.Error()
			stageStatus.WorkflowExecutions = append(stageStatus.WorkflowExecutions, *workflowStatus)
			return fmt.Errorf("workflow %s/%s failed: %w", wfRef.WorkflowRef.Namespace, wfRef.WorkflowRef.Name, err)
		}

		stageStatus.WorkflowExecutions = append(stageStatus.WorkflowExecutions, *workflowStatus)

		// Check failure policy
		if workflowStatus.Phase == drv1alpha1.PhaseFailed {
			return fmt.Errorf("workflow %s/%s failed", wfRef.WorkflowRef.Namespace, wfRef.WorkflowRef.Name)
		}
	}

	return nil
}

// executeParallel executes workflows in parallel with FailFast strategy
func (e *DefaultStageExecutor) executeParallel(ctx context.Context, stage *drv1alpha1.Stage, params map[string]interface{}, stageStatus *drv1alpha1.StageStatus) error {
	klog.V(4).Infof("Executing stage %s in parallel", stage.Name)

	var (
		wg                sync.WaitGroup
		mu                sync.Mutex
		firstErr          error
		cancelCtx, cancel = context.WithCancel(ctx)
	)
	defer cancel()

	workflowStatuses := make([]*drv1alpha1.WorkflowExecutionStatus, len(stage.Workflows))

	for i, wfRef := range stage.Workflows {
		wg.Add(1)
		go func(idx int, ref drv1alpha1.WorkflowReference) {
			defer wg.Done()

			// Check if already cancelled
			select {
			case <-cancelCtx.Done():
				klog.V(4).Infof("Workflow %s/%s skipped due to cancellation", ref.WorkflowRef.Namespace, ref.WorkflowRef.Name)
				workflowStatuses[idx] = &drv1alpha1.WorkflowExecutionStatus{
					WorkflowRef: ref.WorkflowRef,
					Phase:       "Skipped",
					Message:     "Cancelled due to parallel failure",
				}
				return
			default:
			}

			klog.Infof("Executing workflow (parallel): %s/%s", ref.WorkflowRef.Namespace, ref.WorkflowRef.Name)

			// Get workflow
			workflow, err := e.getWorkflow(cancelCtx, ref.WorkflowRef)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("failed to get workflow %s/%s: %w", ref.WorkflowRef.Namespace, ref.WorkflowRef.Name, err)
					cancel() // FailFast: cancel other workflows
				}
				mu.Unlock()
				workflowStatuses[idx] = &drv1alpha1.WorkflowExecutionStatus{
					WorkflowRef: ref.WorkflowRef,
					Phase:       "Failed",
					Message:     err.Error(),
				}
				return
			}

			// Merge parameters: workflow default -> plan globalParams (value) -> stage params (value)
			workflowParams := e.mergeParams(workflow, params, ref.Params)

			// Execute workflow
			workflowStatus, err := e.workflowExecutor.ExecuteWorkflow(cancelCtx, workflow, workflowParams)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
					cancel() // FailFast: cancel other workflows
				}
				mu.Unlock()
				workflowStatus.Phase = drv1alpha1.PhaseFailed
				workflowStatus.Message = err.Error()
			}

			workflowStatuses[idx] = workflowStatus
		}(i, wfRef)
	}

	// Wait for all workflows to complete
	wg.Wait()

	// Collect workflow statuses
	for _, ws := range workflowStatuses {
		if ws != nil {
			stageStatus.WorkflowExecutions = append(stageStatus.WorkflowExecutions, *ws)
		}
	}

	return firstErr
}

// RevertStage reverts a stage by reverting workflows in reverse order
func (e *DefaultStageExecutor) RevertStage(ctx context.Context, _ *drv1alpha1.DRPlan, stage *drv1alpha1.Stage, stageStatus *drv1alpha1.StageStatus) (*drv1alpha1.StageStatus, error) {
	klog.Infof("Reverting stage: %s", stage.Name)

	// Create rollback stage status object
	rollbackStageStatus := &drv1alpha1.StageStatus{
		Name:               stageStatus.Name,
		Phase:              "Running",
		Parallel:           stageStatus.Parallel,
		DependsOn:          stageStatus.DependsOn,
		StartTime:          &metav1.Time{Time: time.Now()},
		WorkflowExecutions: []drv1alpha1.WorkflowExecutionStatus{},
	}

	// Revert workflows in reverse order
	succeededCount := 0
	for i := len(stageStatus.WorkflowExecutions) - 1; i >= 0; i-- {
		wfStatus := stageStatus.WorkflowExecutions[i]
		if wfStatus.Phase != drv1alpha1.PhaseSucceeded {
			klog.V(4).Infof("Skipping revert for workflow %s/%s (phase=%s)", wfStatus.WorkflowRef.Namespace, wfStatus.WorkflowRef.Name, wfStatus.Phase)
			// Record skipped workflow in rollback status
			skippedStatus := drv1alpha1.WorkflowExecutionStatus{
				WorkflowRef: wfStatus.WorkflowRef,
				Phase:       "Skipped",
				Message:     fmt.Sprintf("Original workflow phase was %s, skipped rollback", wfStatus.Phase),
			}
			rollbackStageStatus.WorkflowExecutions = append([]drv1alpha1.WorkflowExecutionStatus{skippedStatus}, rollbackStageStatus.WorkflowExecutions...)
			continue
		}

		// Get workflow
		workflow, err := e.getWorkflow(ctx, wfStatus.WorkflowRef)
		if err != nil {
			klog.Warningf("Failed to get workflow %s/%s for revert: %v", wfStatus.WorkflowRef.Namespace, wfStatus.WorkflowRef.Name, err)
			skippedStatus := drv1alpha1.WorkflowExecutionStatus{
				WorkflowRef: wfStatus.WorkflowRef,
				Phase:       "Skipped",
				Message:     fmt.Sprintf("Failed to get workflow: %v", err),
			}
			rollbackStageStatus.WorkflowExecutions = append([]drv1alpha1.WorkflowExecutionStatus{skippedStatus}, rollbackStageStatus.WorkflowExecutions...)
			continue
		}

		// Revert workflow and get rollback status
		rollbackWorkflowStatus, err := e.workflowExecutor.RevertWorkflow(ctx, workflow, &wfStatus)
		if err != nil {
			klog.Errorf("Failed to revert workflow %s/%s: %v", wfStatus.WorkflowRef.Namespace, wfStatus.WorkflowRef.Name, err)
			// Add failed workflow status to rollback stage status
			rollbackStageStatus.WorkflowExecutions = append([]drv1alpha1.WorkflowExecutionStatus{*rollbackWorkflowStatus}, rollbackStageStatus.WorkflowExecutions...)
			rollbackStageStatus.Phase = drv1alpha1.PhaseFailed
			rollbackStageStatus.Message = fmt.Sprintf("Failed to revert workflow %s/%s", wfStatus.WorkflowRef.Namespace, wfStatus.WorkflowRef.Name)
			rollbackStageStatus.CompletionTime = &metav1.Time{Time: time.Now()}
			rollbackStageStatus.Duration = rollbackStageStatus.CompletionTime.Sub(rollbackStageStatus.StartTime.Time).String()
			return rollbackStageStatus, fmt.Errorf("failed to revert workflow %s/%s: %w", wfStatus.WorkflowRef.Namespace, wfStatus.WorkflowRef.Name, err)
		}

		// Add successful workflow status to rollback stage status (prepend to maintain reverse order)
		rollbackStageStatus.WorkflowExecutions = append([]drv1alpha1.WorkflowExecutionStatus{*rollbackWorkflowStatus}, rollbackStageStatus.WorkflowExecutions...)
		if rollbackWorkflowStatus.Phase == drv1alpha1.PhaseSucceeded {
			succeededCount++
		}
		klog.Infof("Workflow %s/%s reverted successfully", wfStatus.WorkflowRef.Namespace, wfStatus.WorkflowRef.Name)
	}

	// Finalize rollback stage status
	rollbackStageStatus.Phase = drv1alpha1.PhaseSucceeded
	rollbackStageStatus.Message = fmt.Sprintf("Stage reverted successfully: %d workflow(s) rolled back", succeededCount)
	rollbackStageStatus.CompletionTime = &metav1.Time{Time: time.Now()}
	rollbackStageStatus.Duration = rollbackStageStatus.CompletionTime.Sub(rollbackStageStatus.StartTime.Time).String()

	klog.Infof("Stage %s reverted successfully", stage.Name)
	return rollbackStageStatus, nil
}

// getWorkflow retrieves a workflow from the API server
func (e *DefaultStageExecutor) getWorkflow(ctx context.Context, ref drv1alpha1.ObjectReference) (*drv1alpha1.DRWorkflow, error) {
	workflow := &drv1alpha1.DRWorkflow{}
	namespace := ref.Namespace
	if namespace == "" {
		namespace = "default"
	}

	key := client.ObjectKey{
		Name:      ref.Name,
		Namespace: namespace,
	}

	if err := e.client.Get(ctx, key, workflow); err != nil {
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	return workflow, nil
}

// mergeParams merges workflow defaults, plan global params, and stage params.
// Plan only uses value (not default). Priority: workflow default -> globalParams -> stage params (value).
func (e *DefaultStageExecutor) mergeParams(workflow *drv1alpha1.DRWorkflow, globalParams map[string]interface{}, stageParams []drv1alpha1.Parameter) map[string]interface{} {
	result := make(map[string]interface{})

	// 1. Base: workflow parameter defaults
	for _, p := range workflow.Spec.Parameters {
		if p.Default != "" {
			result[p.Name] = p.Default
		}
	}

	// 2. Override with plan global params (value only)
	for k, v := range globalParams {
		result[k] = v
	}

	// 3. Override with stage params (plan uses value only)
	for _, param := range stageParams {
		if param.Value != "" {
			result[param.Name] = param.Value
		}
	}

	return result
}
