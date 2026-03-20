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

// Phase constants for Execution, Action, Workflow, and Stage status
const (
	// PhasePending indicates the resource is waiting to start
	PhasePending = "Pending"

	// PhaseRunning indicates the resource is currently executing
	PhaseRunning = "Running"

	// PhaseSucceeded indicates the resource completed successfully
	PhaseSucceeded = "Succeeded"

	// PhaseFailed indicates the resource encountered an error
	PhaseFailed = "Failed"

	// PhaseSkipped indicates the resource was intentionally skipped
	PhaseSkipped = "Skipped"

	// PhaseCancelled indicates the resource was cancelled before completion
	PhaseCancelled = "Cancelled"

	// PhaseUnknown indicates the resource phase cannot be determined
	PhaseUnknown = "Unknown"
)

// DRPlan Phase constants
const (
	// PlanPhaseReady indicates the plan is validated and ready to execute
	PlanPhaseReady = "Ready"

	// PlanPhaseExecuted indicates the plan has been executed at least once
	PlanPhaseExecuted = "Executed"

	// PlanPhaseInvalid indicates the plan validation failed
	PlanPhaseInvalid = "Invalid"
)

// DRWorkflow Phase constants
const (
	// WorkflowPhaseReady indicates the workflow is validated and ready
	WorkflowPhaseReady = "Ready"

	// WorkflowPhaseInvalid indicates the workflow validation failed
	WorkflowPhaseInvalid = "Invalid"
)

// Operation Type constants for DRPlanExecution
const (
	// OperationTypeExecute indicates an execution operation
	OperationTypeExecute = "Execute"

	// OperationTypeRevert indicates a revert/rollback operation
	OperationTypeRevert = "Revert"
)

// Failure Policy constants
const (
	// FailurePolicyStop stops execution immediately on first failure
	FailurePolicyStop = "Stop"

	// FailurePolicyContinue continues execution despite failures
	FailurePolicyContinue = "Continue"

	// FailurePolicyFailFast stops parallel execution when any task fails
	FailurePolicyFailFast = "FailFast"
)

// Action Type constants
const (
	// ActionTypeHTTP represents an HTTP action
	ActionTypeHTTP = "HTTP"

	// ActionTypeJob represents a Kubernetes Job action
	ActionTypeJob = "Job"

	// ActionTypeLocalization represents a Clusternet Localization action
	ActionTypeLocalization = "Localization"

	// ActionTypeSubscription represents a Clusternet Subscription action
	ActionTypeSubscription = "Subscription"

	// ActionTypeKubernetesResource represents a generic Kubernetes resource action
	ActionTypeKubernetesResource = "KubernetesResource"
)

// Resource Operation constants
const (
	// OperationCreate creates a new resource
	OperationCreate = "Create"

	// OperationApply applies a resource configuration (create or update)
	OperationApply = "Apply"

	// OperationPatch patches an existing resource
	OperationPatch = "Patch"

	// OperationDelete deletes a resource
	OperationDelete = "Delete"
)

// HTTP Method constants
const (
	// HTTPMethodGet represents GET method
	HTTPMethodGet = "GET"

	// HTTPMethodPost represents POST method
	HTTPMethodPost = "POST"

	// HTTPMethodPut represents PUT method
	HTTPMethodPut = "PUT"

	// HTTPMethodPatch represents PATCH method
	HTTPMethodPatch = "PATCH"

	// HTTPMethodDelete represents DELETE method
	HTTPMethodDelete = "DELETE"

	// HTTPMethodHead represents HEAD method
	HTTPMethodHead = "HEAD"

	// HTTPMethodOptions represents OPTIONS method
	HTTPMethodOptions = "OPTIONS"
)

// Executor Type constants
const (
	// ExecutorTypeNative represents the native Go executor
	ExecutorTypeNative = "Native"

	// ExecutorTypeArgo represents the Argo Workflows executor (future extension)
	ExecutorTypeArgo = "Argo"
)

// Override Type constants for Localization
const (
	// OverrideTypeJSONPatch represents JSON Patch (RFC 6902)
	OverrideTypeJSONPatch = "JSONPatch"

	// OverrideTypeMergePatch represents JSON Merge Patch (RFC 7386)
	OverrideTypeMergePatch = "MergePatch"

	// OverrideTypeHelm represents Helm value overrides
	OverrideTypeHelm = "Helm"
)

// Parameter Type constants
const (
	// ParameterTypeString represents a string parameter
	ParameterTypeString = "string"

	// ParameterTypeNumber represents a numeric parameter
	ParameterTypeNumber = "number"

	// ParameterTypeBoolean represents a boolean parameter
	ParameterTypeBoolean = "boolean"
)
