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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusternetapps "github.com/clusternet/clusternet/pkg/apis/apps/v1alpha1"
)

// Parameter defines a parameter with name, type, and value/default
type Parameter struct {
	// Name is the parameter name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Type is the parameter type: string (default), number, boolean
	// +kubebuilder:validation:Enum=string;number;boolean
	// +kubebuilder:default=string
	// +optional
	Type string `json:"type,omitempty"`

	// Required indicates if the parameter is required
	// +optional
	Required bool `json:"required,omitempty"`

	// Value is the explicit value to use (e.g. when passing params in DRPlan stage). Takes precedence over Default when both set.
	// Mutually exclusive with ValueFrom.
	// +optional
	Value string `json:"value,omitempty"`

	// Default is the default value (used when Value is not set, e.g. in workflow/plan parameter definitions)
	// +optional
	Default string `json:"default,omitempty"`

	// Description is the parameter description
	// +optional
	Description string `json:"description,omitempty"`

	// ValueFrom dynamically resolves a parameter value from a Kubernetes resource.
	// Mutually exclusive with Value.
	// +optional
	ValueFrom *ParameterValueFrom `json:"valueFrom,omitempty"`
}

// ParameterValueFrom specifies the source of a dynamic parameter value.
type ParameterValueFrom struct {
	// ManifestRef resolves a value from a Kubernetes resource field via JSONPath.
	// +optional
	ManifestRef *ManifestRef `json:"manifestRef,omitempty"`
}

// ManifestRef describes a Kubernetes resource and a JSONPath expression used to extract a parameter value.
type ManifestRef struct {
	// APIVersion is the resource API version (e.g. "batch/v1")
	// +kubebuilder:validation:Required
	APIVersion string `json:"apiVersion"`

	// Kind is the resource kind (e.g. "Job")
	// +kubebuilder:validation:Required
	Kind string `json:"kind"`

	// Namespace is the resource namespace. Supports $(params.xxx) placeholders.
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Name is the exact resource name. Mutually exclusive with LabelSelector.
	// Supports $(params.xxx) placeholders.
	// +optional
	Name string `json:"name,omitempty"`

	// LabelSelector is a label selector string (e.g. "app=foo,env=prod").
	// Mutually exclusive with Name. Supports $(params.xxx) placeholders.
	// +optional
	LabelSelector string `json:"labelSelector,omitempty"`

	// JSONPath is the JSONPath expression to extract the value (e.g. "{.metadata.name}")
	// +kubebuilder:validation:Required
	JSONPath string `json:"jsonPath"`

	// Select specifies which resource to pick when multiple matches are found.
	// Last (default): pick by latest creationTimestamp; First: pick by earliest;
	// Single: require exactly 1 match; Any: pick arbitrarily.
	// +kubebuilder:validation:Enum=Last;First;Single;Any
	// +kubebuilder:default=Last
	// +optional
	Select string `json:"select,omitempty"`
}

// Action defines a single action in a workflow
type Action struct {
	// Name is the unique action name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Type is the action type
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=HTTP;Job;Localization;Subscription;KubernetesResource
	Type string `json:"type"`

	// HTTP configuration (required when type=HTTP)
	// +optional
	HTTP *HTTPAction `json:"http,omitempty"`

	// Job configuration (required when type=Job)
	// +optional
	Job *JobAction `json:"job,omitempty"`

	// Localization configuration (required when type=Localization)
	// +optional
	Localization *LocalizationAction `json:"localization,omitempty"`

	// Subscription configuration (required when type=Subscription)
	// +optional
	Subscription *SubscriptionAction `json:"subscription,omitempty"`

	// KubernetesResource configuration (required when type=KubernetesResource)
	// +optional
	Resource *KubernetesResourceAction `json:"resource,omitempty"`

	// Timeout for the action (default: 5m)
	// +kubebuilder:default="5m"
	// +optional
	Timeout string `json:"timeout,omitempty"`

	// WaitReady indicates whether to wait until the target resource is ready.
	// When false (default), action succeeds once create/apply request is accepted.
	// When true, the executor polls the resource status until ready or timeout.
	// Currently only effective for Subscription actions (child cluster direct query via SocketProxy).
	// +kubebuilder:default=false
	// +optional
	WaitReady bool `json:"waitReady,omitempty"`

	// RetryPolicy defines retry behavior
	// +optional
	RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty"`

	// Rollback defines custom rollback action
	// +optional
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Rollback *Action `json:"rollback,omitempty"`

	// DependsOn lists action names this action depends on (reserved for DAG, Phase 2)
	// +optional
	DependsOn []string `json:"dependsOn,omitempty"`

	// When is a condition expression (reserved for conditional execution, Phase 2)
	// +optional
	When string `json:"when,omitempty"`

	// HookType identifies the originating Helm hook lifecycle when this action was generated from a hook.
	// It allows executors to distinguish pre/post install/upgrade/delete/rollback hooks from normal actions.
	// +optional
	HookType string `json:"hookType,omitempty"`

	// ClusterExecutionMode controls how this action is executed across binding clusters.
	// Empty or "Global" (default): executes as a single aggregate action; waitReady waits for ALL clusters.
	// "PerCluster": at runtime, splits into per-cluster child Subscriptions for independent progression.
	// Only effective for Subscription type actions with waitReady=true.
	// +kubebuilder:validation:Enum="";Global;PerCluster
	// +kubebuilder:default=""
	// +optional
	ClusterExecutionMode string `json:"clusterExecutionMode,omitempty"`

	// HookCleanup controls Helm-like hook cleanup timing for hook actions.
	// It is primarily used by drplan-gen generated hook Subscription actions.
	// When nil, executors keep the current default behavior and do not apply extra cleanup.
	// +optional
	HookCleanup *HookCleanupPolicy `json:"hookCleanup,omitempty"`
}

// HookCleanupPolicy defines when a hook resource should be automatically deleted.
type HookCleanupPolicy struct {
	// BeforeCreate deletes any existing hook resource before creating a new one.
	// Mirrors Helm's before-hook-creation policy.
	// +optional
	BeforeCreate bool `json:"beforeCreate,omitempty"`

	// OnSuccess deletes the hook resource after the action succeeds.
	// Mirrors Helm's hook-succeeded policy.
	// +optional
	OnSuccess bool `json:"onSuccess,omitempty"`

	// OnFailure deletes the hook resource after the action fails.
	// Mirrors Helm's hook-failed policy.
	// +optional
	OnFailure bool `json:"onFailure,omitempty"`
}

// HTTPAction defines HTTP request configuration
type HTTPAction struct {
	// URL is the request URL (supports parameter placeholders)
	// +kubebuilder:validation:Required
	URL string `json:"url"`

	// Method is the HTTP method (default: GET)
	// +kubebuilder:validation:Enum=GET;POST;PUT;PATCH;DELETE;HEAD;OPTIONS
	// +kubebuilder:default=GET
	// +optional
	Method string `json:"method,omitempty"`

	// Headers are request headers
	// +optional
	Headers map[string]string `json:"headers,omitempty"`

	// Body is the request body (supports parameter placeholders)
	// +optional
	Body string `json:"body,omitempty"`

	// SuccessCodes are HTTP status codes considered successful (default: 200-299)
	// +optional
	SuccessCodes []int `json:"successCodes,omitempty"`

	// InsecureSkipVerify skips TLS certificate verification
	// +optional
	InsecureSkipVerify bool `json:"insecureSkipVerify,omitempty"`
}

// JobAction defines Kubernetes Job configuration
type JobAction struct {
	// Namespace is the Job namespace (default: default)
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Template is the Job template spec
	// +kubebuilder:validation:Required
	Template JobTemplateSpec `json:"template"`

	// TTLSecondsAfterFinished limits the lifetime of a Job after it finishes
	// +optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`
}

// JobTemplateSpec wraps batchv1.JobSpec for validation
type JobTemplateSpec struct {
	// Standard Kubernetes Job spec
	// +kubebuilder:validation:Required
	Spec batchv1.JobSpec `json:"spec"`
}

// LocalizationAction defines Clusternet Localization operation.
// Name/Namespace 为 CR 的 metadata；Spec 直接引用 clusternet LocalizationSpec，与上游保持一致、避免缺失字段。
type LocalizationAction struct {
	// Operation is the operation type: Create (default), Patch, Delete
	// +kubebuilder:validation:Enum=Create;Patch;Delete
	// +kubebuilder:default=Create
	// +optional
	Operation string `json:"operation,omitempty"`

	// Name is the Localization CR name (supports placeholders)
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace is the Localization CR namespace (ManagedCluster namespace, supports placeholders)
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// Spec is the spec of the Localization CR, same as apps.clusternet.io LocalizationSpec.
	// +optional
	Spec *clusternetapps.LocalizationSpec `json:"spec,omitempty"`
}

// LocalizationOverride defines Clusternet Localization override
type LocalizationOverride struct {
	// Name is the override name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Type is the override type
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=JSONPatch;MergePatch;Helm
	Type string `json:"type"`

	// Value is the override content (YAML/JSON format)
	// +kubebuilder:validation:Required
	Value string `json:"value"`

	// OverrideChart indicates whether to override HelmChart CR (only effective when type=Helm)
	// +optional
	OverrideChart bool `json:"overrideChart,omitempty"`
}

// Feed defines a resource reference
type Feed struct {
	// APIVersion is the resource API version
	// +kubebuilder:validation:Required
	APIVersion string `json:"apiVersion"`

	// Kind is the resource kind
	// +kubebuilder:validation:Required
	Kind string `json:"kind"`

	// Name is the resource name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace is the resource namespace
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// SubscriptionAction defines Clusternet Subscription operation.
// Name/Namespace 为 CR 的 metadata；Spec 直接引用 clusternet SubscriptionSpec，与上游保持一致、避免缺失字段。
type SubscriptionAction struct {
	// Operation is the operation type: Create (default), Apply, Patch, Delete, Replace.
	// Apply uses Server-Side Apply (idempotent create-or-update).
	// Replace deletes the existing resource (if any) and waits for deletion to complete
	// before creating a new one.
	// +kubebuilder:validation:Enum=Create;Apply;Patch;Delete;Replace
	// +kubebuilder:default=Create
	// +optional
	Operation string `json:"operation,omitempty"`

	// Name is the Subscription name (supports placeholders)
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace is the Subscription namespace
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// Spec is the spec of the Subscription CR, same as apps.clusternet.io SubscriptionSpec.
	// +optional
	Spec *clusternetapps.SubscriptionSpec `json:"spec,omitempty"`
}

// Subscriber defines a subscription target
type Subscriber struct {
	// ClusterAffinity is the cluster affinity selector
	// +optional
	ClusterAffinity *ClusterAffinity `json:"clusterAffinity,omitempty"`

	// Weight for Dividing scheduling strategy
	// +optional
	Weight int32 `json:"weight,omitempty"`
}

// ClusterAffinity defines cluster selection
type ClusterAffinity struct {
	// MatchExpressions is a list of label selector requirements
	// +optional
	MatchExpressions []metav1.LabelSelectorRequirement `json:"matchExpressions,omitempty"`
}

// KubernetesResourceAction defines generic Kubernetes resource operation
type KubernetesResourceAction struct {
	// Operation is the operation type: Create (default), Apply, Patch, Delete
	// +kubebuilder:validation:Enum=Create;Apply;Patch;Delete
	// +kubebuilder:default=Create
	// +optional
	Operation string `json:"operation,omitempty"`

	// Manifest is the Kubernetes resource manifest (YAML format, supports parameter placeholders)
	// +kubebuilder:validation:Required
	Manifest string `json:"manifest"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	// Limit is the maximum retry count (default: 3)
	// +kubebuilder:default=3
	// +optional
	Limit int32 `json:"limit,omitempty"`

	// Interval is the retry interval (default: 5s)
	// +kubebuilder:default="5s"
	// +optional
	Interval string `json:"interval,omitempty"`

	// BackoffMultiplier is the backoff multiplier as a string (default: "2.0")
	// +kubebuilder:default="2.0"
	// +optional
	BackoffMultiplier string `json:"backoffMultiplier,omitempty"`
}

// ExecutorConfig defines executor configuration (reserved for extension)
type ExecutorConfig struct {
	// Type is the executor type: Native (default), Argo
	// +kubebuilder:validation:Enum=Native;Argo
	// +kubebuilder:default=Native
	// +optional
	Type string `json:"type,omitempty"`

	// ArgoOptions is the Argo engine configuration (effective when type=Argo)
	// +optional
	ArgoOptions *ArgoOptions `json:"argoOptions,omitempty"`
}

// ArgoOptions defines Argo Workflow configuration (reserved for extension)
type ArgoOptions struct {
	// Namespace is the Argo Workflow namespace (default: same as DRWorkflow)
	// +optional
	Namespace string `json:"namespace,omitempty"`

	// ServiceAccountName is the ServiceAccount used by Argo Workflow
	// +optional
	ServiceAccountName string `json:"serviceAccountName,omitempty"`

	// TTLStrategy is the workflow retention strategy
	// +optional
	TTLStrategy *TTLStrategy `json:"ttlStrategy,omitempty"`
}

// TTLStrategy defines workflow retention strategy (reserved for extension)
type TTLStrategy struct {
	// SecondsAfterCompletion is the retention time after completion
	// +optional
	SecondsAfterCompletion *int32 `json:"secondsAfterCompletion,omitempty"`

	// SecondsAfterSuccess is the retention time after success
	// +optional
	SecondsAfterSuccess *int32 `json:"secondsAfterSuccess,omitempty"`

	// SecondsAfterFailure is the retention time after failure
	// +optional
	SecondsAfterFailure *int32 `json:"secondsAfterFailure,omitempty"`
}

// ActionStatus defines the status of an action execution
type ActionStatus struct {
	// Name is the action name
	Name string `json:"name"`

	// Phase is the action phase: Pending, Running, Succeeded, Failed, Skipped
	// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Skipped
	Phase string `json:"phase"`

	// StartTime is the start time
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is the completion time
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// RetryCount is the number of retries
	// +optional
	RetryCount int32 `json:"retryCount,omitempty"`

	// Message is the status message or error information
	// +optional
	Message string `json:"message,omitempty"`

	// Outputs are the action outputs (used for rollback)
	// +optional
	Outputs *ActionOutputs `json:"outputs,omitempty"`

	// ClusterStatuses records per-cluster execution state when ClusterExecutionMode=PerCluster.
	// Empty for Global actions (backward compatible).
	// +optional
	ClusterStatuses []ClusterActionStatus `json:"clusterStatuses,omitempty"`
}

// ClusterActionStatus tracks execution state of a single cluster within a PerCluster action.
type ClusterActionStatus struct {
	// Cluster is the binding cluster identifier (format: "namespace/name")
	Cluster string `json:"cluster"`

	// ClusterID is the ManagedCluster spec.clusterId
	ClusterID string `json:"clusterID"`

	// Phase is the per-cluster action phase: Pending, Running, Succeeded, Failed
	// +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed;Skipped
	Phase string `json:"phase"`

	// StartTime is when this cluster's execution started
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime is when this cluster's execution completed
	// +optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Message provides additional information
	// +optional
	Message string `json:"message,omitempty"`
}

// ActionOutputs defines action execution outputs
type ActionOutputs struct {
	// JobRef is the created Job reference
	// +optional
	JobRef *corev1.ObjectReference `json:"jobRef,omitempty"`

	// LocalizationRef is the created Localization reference
	// +optional
	LocalizationRef *corev1.ObjectReference `json:"localizationRef,omitempty"`

	// SubscriptionRef is the created Subscription reference
	// +optional
	SubscriptionRef *corev1.ObjectReference `json:"subscriptionRef,omitempty"`

	// SubscriptionRefs are the created Subscription references for PerCluster actions.
	// +optional
	SubscriptionRefs []corev1.ObjectReference `json:"subscriptionRefs,omitempty"`

	// ResourceRef is the created generic K8s resource reference (KubernetesResource)
	// +optional
	ResourceRef *corev1.ObjectReference `json:"resourceRef,omitempty"`

	// HTTPResponse is the HTTP response summary
	// +optional
	HTTPResponse *HTTPResponse `json:"httpResponse,omitempty"`
}

// HTTPResponse defines HTTP response summary
type HTTPResponse struct {
	// StatusCode is the response status code
	StatusCode int `json:"statusCode"`

	// Body is the response body (truncated)
	// +optional
	Body string `json:"body,omitempty"`
}

// ObjectReference defines a reference to another object
type ObjectReference struct {
	// Name is the object name
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace is the object namespace
	// +optional
	Namespace string `json:"namespace,omitempty"`
}
