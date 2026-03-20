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

	drv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller/api/v1alpha1"
)

// ActionExecutor defines the interface for executing actions
type ActionExecutor interface {
	// Execute executes an action and returns the action status
	Execute(ctx context.Context, action *drv1alpha1.Action, params map[string]interface{}) (*drv1alpha1.ActionStatus, error)

	// Rollback rolls back an action using the provided action status
	// Returns the rollback action status and error
	Rollback(ctx context.Context, action *drv1alpha1.Action, actionStatus *drv1alpha1.ActionStatus, params map[string]interface{}) (*drv1alpha1.ActionStatus, error)

	// Type returns the action type this executor handles
	Type() string
}

// WorkflowExecutor defines the interface for executing workflows
type WorkflowExecutor interface {
	// ExecuteWorkflow executes a workflow and returns the workflow execution status
	ExecuteWorkflow(ctx context.Context, workflow *drv1alpha1.DRWorkflow, params map[string]interface{}) (*drv1alpha1.WorkflowExecutionStatus, error)

	// RevertWorkflow reverts a workflow using the provided workflow execution status
	// Returns the rollback workflow status and error
	RevertWorkflow(ctx context.Context, workflow *drv1alpha1.DRWorkflow, workflowStatus *drv1alpha1.WorkflowExecutionStatus) (*drv1alpha1.WorkflowExecutionStatus, error)
}

// StageExecutor defines the interface for executing stages
type StageExecutor interface {
	// ExecuteStage executes a stage and returns the stage status
	ExecuteStage(ctx context.Context, plan *drv1alpha1.DRPlan, stage *drv1alpha1.Stage, params map[string]interface{}) (*drv1alpha1.StageStatus, error)

	// RevertStage reverts a stage using the provided stage status
	// Returns the rollback stage status and error
	RevertStage(ctx context.Context, plan *drv1alpha1.DRPlan, stage *drv1alpha1.Stage, stageStatus *drv1alpha1.StageStatus) (*drv1alpha1.StageStatus, error)
}

// PlanExecutor defines the interface for executing plans
type PlanExecutor interface {
	// ExecutePlan executes a DR plan and returns the execution status
	ExecutePlan(ctx context.Context, plan *drv1alpha1.DRPlan, execution *drv1alpha1.DRPlanExecution) error

	// RevertPlan reverts a DR plan using the provided execution status
	RevertPlan(ctx context.Context, plan *drv1alpha1.DRPlan, execution *drv1alpha1.DRPlanExecution) error

	// CancelExecution cancels an ongoing execution
	CancelExecution(ctx context.Context, execution *drv1alpha1.DRPlanExecution) error
}

// ExecutorRegistry manages action executors
type ExecutorRegistry interface {
	// RegisterExecutor registers an action executor
	RegisterExecutor(executor ActionExecutor) error

	// GetExecutor returns an executor for the given action type
	GetExecutor(actionType string) (ActionExecutor, error)

	// ListExecutors returns all registered executors
	ListExecutors() []ActionExecutor
}
