/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// +k8s:openapi-gen=true

package generalpodautoscaler

import (
	admregv1b "k8s.io/api/admissionregistration/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:resource:shortName=gpa
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:JSONPath=.spec.minReplicas,name=MinReplicas,type=integer
// +kubebuilder:printcolumn:JSONPath=.spec.maxReplicas,name=MaxReplicas,type=integer
// +kubebuilder:printcolumn:JSONPath=.status.desiredReplicas,name=Desired,type=integer
// +kubebuilder:printcolumn:JSONPath=.status.currentReplicas,name=Current,type=integer
// +kubebuilder:printcolumn:JSONPath=.spec.scaleTargetRef.kind,name=TargetKind,type=string
// +kubebuilder:printcolumn:JSONPath=.spec.scaleTargetRef.name,name=TargetName,type=string
// +kubebuilder:subresource:status
// +k8s:defaulter-gen=true

// GeneralPodAutoscaler is the configuration for a general pod
// autoscaler, which automatically manages the replica count of any resource
// implementing the scale subresource based on the metrics specified.
type GeneralPodAutoscaler struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is the standard object metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// spec is the specification for the behaviour of the autoscaler.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#spec-and-status.
	// +optional
	// +kubebuilder:validation:Required
	Spec GeneralPodAutoscalerSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// status is the current information about the autoscaler.
	// +optional
	Status GeneralPodAutoscalerStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// GeneralPodAutoscalerSpec describes the desired functionality of the GeneralPodAutoscaler.
type GeneralPodAutoscalerSpec struct {
	// DrivenMode is the mode the open autoscaling mode if we do not need scaling according to metrics.
	// including MetricMode, TimeMode, EventMode, WebhookMode
	// +optional
	AutoScalingDrivenMode `json:",inline"`

	// scaleTargetRef points to the target resource to scale, and is used to the pods for which metrics
	// should be collected, as well as to actually change the replica count.
	ScaleTargetRef CrossVersionObjectReference `json:"scaleTargetRef" protobuf:"bytes,1,opt,name=scaleTargetRef"`

	// minReplicas is the lower limit for the number of replicas to which the autoscaler
	// can scale down.  It defaults to 1 pod.  minReplicas is allowed to be 0 if the
	// alpha feature gate GPAScaleToZero is enabled and at least one Object or External
	// metric is configured.  Scaling is active as long as at least one metric value is
	// available.
	// +optional
	// +kubebuilder:default=1
	MinReplicas *int32 `json:"minReplicas,omitempty" protobuf:"varint,2,opt,name=minReplicas"`

	// maxReplicas is the upper limit for the number of replicas to which the autoscaler can scale up.
	// It cannot be less that minReplicas.
	MaxReplicas int32 `json:"maxReplicas" protobuf:"varint,3,opt,name=maxReplicas"`

	// behavior configures the scaling behavior of the target
	// in both Up and Down directions (scaleUp and scaleDown fields respectively).
	// If not set, the default GPAScalingRules for scale up and scale down are used.
	// +optional
	Behavior *GeneralPodAutoscalerBehavior `json:"behavior,omitempty" protobuf:"bytes,4,opt,name=behavior"`
}

// AutoScalingDrivenMode defines the mode to trigger auto scaling
type AutoScalingDrivenMode struct {
	// EventMode is the metric driven mode.
	// +optional
	MetricMode *MetricMode `json:"metric,omitempty" protobuf:"bytes,1,opt,name=metric"`

	// Webhook defines webhook mode the allow us to revive requests to scale.
	// +optional
	WebhookMode *WebhookMode `json:"webhook,omitempty" protobuf:"bytes,2,opt,name=webhook"`

	// Time defines the time driven mode, pod would auto scale to max if time reached
	// +optional
	TimeMode *TimeMode `json:"time,omitempty" protobuf:"bytes,3,opt,name=time"`

	// EventMode is the event driven mode
	// +optional
	EventMode *EventMode `json:"event,omitempty" protobuf:"bytes,4,opt,name=event"`
}

// MetricMode metric mode 指标模式
type MetricMode struct {
	// metrics contains the specifications for which to use to calculate the
	// desired replica count (the maximum replica count across all metrics will
	// be used).  The desired replica count is calculated multiplying the
	// ratio between the target value and the current value by the current
	// number of pods.  Ergo, metrics used must decrease as the pod count is
	// increased, and vice-versa.  See the individual metric source types for
	// more information about how each type of metric must respond.
	// If not set, the default metric will be set to 80% average CPU utilization.
	// +optional
	Metrics []MetricSpec `json:"metrics,omitempty" protobuf:"bytes,1,opt,name=metrics"`
}

// EventMode is the event driven mode
type EventMode struct {
	// Triggers are thr event triggers
	Triggers []ScaleTriggers `json:"triggers"`
}

// ScaleTriggers reference the scaler that will be used
type ScaleTriggers struct {
	// Type are the trigger type
	Type string `json:"type"`
	// Name is the trigger name
	// +optional
	Name string `json:"name,omitempty"`
	// Metadata contains the trigger config
	Metadata map[string]string `json:"metadata"`
}

// WebhookMode allow users to provider a server
type WebhookMode struct {
	*admregv1b.WebhookClientConfig `json:",inline"`
	// Parameters are the webhook parameters
	Parameters map[string]string `json:"parameters,omitempty" protobuf:"bytes,1,opt,name=parameters"`
}

// TimeMode is a mode allows user to define a crontab regular
type TimeMode struct {
	// TimeRanges defines a array that for time driven mode
	TimeRanges []TimeRange `json:"ranges,omitempty" protobuf:"bytes,1,opt,name=ranges"`
}

// TimeRange is a mode allows user to define a crontab regular
type TimeRange struct {
	// Schedule should match crontab format
	Schedule string `json:"schedule,omitempty" protobuf:"bytes,1,opt,name=schedule"`

	// DesiredReplicas is the desired replicas required by timemode,
	DesiredReplicas int32 `json:"desiredReplicas,omitempty" protobuf:"varint,2,opt,name=desiredReplicas"`
}

// CrossVersionObjectReference contains enough information to let you identify the referred resource.
type CrossVersionObjectReference struct {
	// Kind of the referent;
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds"
	Kind string `json:"kind" protobuf:"bytes,1,opt,name=kind"`
	// Name of the referent; More info: http://kubernetes.io/docs/user-guide/identifiers#names
	Name string `json:"name" protobuf:"bytes,2,opt,name=name"`
	// API version of the referent
	// +optional
	APIVersion string `json:"apiVersion,omitempty" protobuf:"bytes,3,opt,name=apiVersion"`
}

// MetricSpec specifies how to scale based on a single metric
// (only `type` and one other matching field should be set at once).
type MetricSpec struct {
	// type is the type of metric source.  It should be one of "Object",
	// "Pods" or "Resource", each mapping to a matching field in the object.
	Type MetricSourceType `json:"type" protobuf:"bytes,1,name=type"`

	// object refers to a metric describing a single kubernetes object
	// (for example, hits-per-second on an Ingress object).
	// +optional
	Object *ObjectMetricSource `json:"object,omitempty" protobuf:"bytes,2,opt,name=object"`
	// pods refers to a metric describing each pod in the current scale target
	// (for example, transactions-processed-per-second).  The values will be
	// averaged together before being compared to the target value.
	// +optional
	Pods *PodsMetricSource `json:"pods,omitempty" protobuf:"bytes,3,opt,name=pods"`
	// resource refers to a resource metric (such as those specified in
	// requests and limits) known to Kubernetes describing each pod in the
	// current scale target (e.g. CPU or memory). Such metrics are built in to
	// Kubernetes, and have special scaling options on top of those available
	// to normal per-pod metrics using the "pods" source.
	// +optional
	Resource *ResourceMetricSource `json:"resource,omitempty" protobuf:"bytes,4,opt,name=resource"`
	// external refers to a global metric that is not associated
	// with any Kubernetes object. It allows autoscaling based on information
	// coming from components running outside of cluster
	// (for example length of queue in cloud messaging service, or
	// QPS from loadbalancer running outside of cluster).
	// +optional
	External *ExternalMetricSource `json:"external,omitempty" protobuf:"bytes,5,opt,name=external"`
	// container resource refers to a resource metric (such as those specified in
	// requests and limits) known to Kubernetes describing a single container in
	// each pod of the current scale target (e.g. CPU or memory). Such metrics are
	// built in to Kubernetes, and have special scaling options on top of those
	// available to normal per-pod metrics using the "pods" source.
	// This is an alpha feature and can be enabled by the HPAContainerMetrics feature flag.
	// +optional
	ContainerResource *ContainerResourceMetricSource `json:"containerResource,omitempty" protobuf:"bytes,6,opt,name=containerResource"`
}

// GeneralPodAutoscalerBehavior configures the scaling behavior of the target
// in both Up and Down directions (scaleUp and scaleDown fields respectively).
type GeneralPodAutoscalerBehavior struct {
	// scaleUp is scaling policy for scaling Up.
	// If not set, the default value is the higher of:
	//   * increase no more than 4 pods per 60 seconds
	//   * double the number of pods per 60 seconds
	// No stabilization is used.
	// +optional
	ScaleUp *GPAScalingRules `json:"scaleUp,omitempty" protobuf:"bytes,1,opt,name=scaleUp"`
	// scaleDown is scaling policy for scaling Down.
	// If not set, the default value is to allow to scale down to minReplicas pods, with a
	// 300 second stabilization window (i.e., the highest recommendation for
	// the last 300sec is used).
	// +optional
	ScaleDown *GPAScalingRules `json:"scaleDown,omitempty" protobuf:"bytes,2,opt,name=scaleDown"`
}

// ScalingPolicySelect is used to specify which policy should be used while scaling in a certain direction
type ScalingPolicySelect string

const (
	// MaxPolicySelect selects the policy with the highest possible change.
	MaxPolicySelect ScalingPolicySelect = "Max"
	// MinPolicySelect selects the policy with the lowest possible change.
	MinPolicySelect ScalingPolicySelect = "Min"
	// DisabledPolicySelect disables the scaling in this direction.
	DisabledPolicySelect ScalingPolicySelect = "Disabled"
)

// GPAScalingRules configures the scaling behavior for one direction.
// These Rules are applied after calculating DesiredReplicas from metrics for the GPA.
// They can limit the scaling velocity by specifying scaling policies.
// They can prevent flapping by specifying the stabilization window, so that the
// number of replicas is not set instantly, instead, the safest value from the stabilization
// window is chosen.
type GPAScalingRules struct {
	// StabilizationWindowSeconds is the number of seconds for which past recommendations should be
	// considered while scaling up or scaling down.
	// StabilizationWindowSeconds must be greater than or equal to zero and less than or equal to 3600 (one hour).
	// If not set, use the default values:
	// - For scale up: 0 (i.e. no stabilization is done).
	// - For scale down: 300 (i.e. the stabilization window is 300 seconds long).
	// +optional
	StabilizationWindowSeconds *int32 `json:"stabilizationWindowSeconds,omitempty" protobuf:"varint,3,opt,name=stabilizationWindowSeconds"`
	// selectPolicy is used to specify which policy should be used.
	// If not set, the default value MaxPolicySelect is used.
	// +optional
	SelectPolicy *ScalingPolicySelect `json:"selectPolicy,omitempty" protobuf:"bytes,1,opt,name=selectPolicy"`
	// policies is a list of potential scaling polices which can be used during scaling.
	// At least one policy must be specified, otherwise the GPAScalingRules will be discarded as invalid
	// +optional
	Policies []GPAScalingPolicy `json:"policies,omitempty" protobuf:"bytes,2,rep,name=policies"`
}

// GPAScalingPolicyType is the type of the policy which could be used while making scaling decisions.
type GPAScalingPolicyType string

const (
	// PodsScalingPolicy is a policy used to specify a change in absolute number of pods.
	PodsScalingPolicy GPAScalingPolicyType = "Pods"
	// PercentScalingPolicy is a policy used to specify a relative amount of change with respect to
	// the current number of pods.
	PercentScalingPolicy GPAScalingPolicyType = "Percent"
)

// GPAScalingPolicy is a single policy which must hold true for a specified past interval.
type GPAScalingPolicy struct {
	// Type is used to specify the scaling policy.
	Type GPAScalingPolicyType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=GPAScalingPolicyType"`
	// Value contains the amount of change which is permitted by the policy.
	// It must be greater than zero
	Value int32 `json:"value" protobuf:"varint,2,opt,name=value"`
	// PeriodSeconds specifies the window of time for which the policy should hold true.
	// PeriodSeconds must be greater than zero and less than or equal to 1800 (30 min).
	PeriodSeconds int32 `json:"periodSeconds" protobuf:"varint,3,opt,name=periodSeconds"`
}

// MetricSourceType indicates the type of metric.
type MetricSourceType string

const (
	// ObjectMetricSourceType is a metric describing a kubernetes object
	// (for example, hits-per-second on an Ingress object).
	ObjectMetricSourceType MetricSourceType = "Object"
	// PodsMetricSourceType is a metric describing each pod in the current scale
	// target (for example, transactions-processed-per-second).  The values
	// will be averaged together before being compared to the target value.
	PodsMetricSourceType MetricSourceType = "Pods"
	// ResourceMetricSourceType is a resource metric known to Kubernetes, as
	// specified in requests and limits, describing each pod in the current
	// scale target (e.g. CPU or memory).  Such metrics are built in to
	// Kubernetes, and have special scaling options on top of those available
	// to normal per-pod metrics (the "pods" source).
	ResourceMetricSourceType MetricSourceType = "Resource"
	// ContainerResourceMetricSourceType is a resource metric known to Kubernetes, as
	// specified in requests and limits, describing a single container in each pod in the current
	// scale target (e.g. CPU or memory).  Such metrics are built in to
	// Kubernetes, and have special scaling options on top of those available
	// to normal per-pod metrics (the "pods" source).
	ContainerResourceMetricSourceType MetricSourceType = "ContainerResource"
	// ExternalMetricSourceType is a global metric that is not associated
	// with any Kubernetes object. It allows autoscaling based on information
	// coming from components running outside of cluster
	// (for example length of queue in cloud messaging service, or
	// QPS from loadbalancer running outside of cluster).
	ExternalMetricSourceType MetricSourceType = "External"
)

// ObjectMetricSource indicates how to scale on a metric describing a
// kubernetes object (for example, hits-per-second on an Ingress object).
type ObjectMetricSource struct {
	DescribedObject CrossVersionObjectReference `json:"describedObject" protobuf:"bytes,1,name=describedObject"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
	// metric identifies the target metric by name and selector
	Metric MetricIdentifier `json:"metric" protobuf:"bytes,3,name=metric"`
}

// PodsMetricSource indicates how to scale on a metric describing each pod in
// the current scale target (for example, transactions-processed-per-second).
// The values will be averaged together before being compared to the target
// value.
type PodsMetricSource struct {
	// metric identifies the target metric by name and selector
	Metric MetricIdentifier `json:"metric" protobuf:"bytes,1,name=metric"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
}

// ResourceMetricSource indicates how to scale on a resource metric known to
// Kubernetes, as specified in requests and limits, describing each pod in the
// current scale target (e.g. CPU or memory).  The values will be averaged
// together before being compared to the target.  Such metrics are built in to
// Kubernetes, and have special scaling options on top of those available to
// normal per-pod metrics using the "pods" source.  Only one "target" type
// should be set.
type ResourceMetricSource struct {
	// name is the name of the resource in question.
	Name v1.ResourceName `json:"name" protobuf:"bytes,1,name=name"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
}

// ContainerResourceMetricSource indicates how to scale on a resource metric known to
// Kubernetes, as specified in requests and limits, describing each pod in the
// current scale target (e.g. CPU or memory).  The values will be averaged
// together before being compared to the target.  Such metrics are built in to
// Kubernetes, and have special scaling options on top of those available to
// normal per-pod metrics using the "pods" source.  Only one "target" type
// should be set.
type ContainerResourceMetricSource struct {
	// name is the name of the resource in question.
	Name v1.ResourceName `json:"name" protobuf:"bytes,1,name=name"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
	// container is the name of the container in the pods of the scaling target
	Container string `json:"container" protobuf:"bytes,3,opt,name=container"`
}

// ExternalMetricSource indicates how to scale on a metric not associated with
// any Kubernetes object (for example length of queue in cloud
// messaging service, or QPS from loadbalancer running outside of cluster).
type ExternalMetricSource struct {
	// metric identifies the target metric by name and selector
	Metric MetricIdentifier `json:"metric" protobuf:"bytes,1,name=metric"`
	// target specifies the target value for the given metric
	Target MetricTarget `json:"target" protobuf:"bytes,2,name=target"`
}

// MetricIdentifier defines the name and optionally selector for a metric
type MetricIdentifier struct {
	// name is the name of the given metric
	Name string `json:"name" protobuf:"bytes,1,name=name"`
	// selector is the string-encoded form of a standard kubernetes label selector for the given metric
	// When set, it is passed as an additional parameter to the metrics server for more specific metrics scoping.
	// When unset, just the metricName will be used to gather metrics.
	// +optional
	Selector *metav1.LabelSelector `json:"selector,omitempty" protobuf:"bytes,2,name=selector"`
}

// MetricTarget defines the target value, average value, or average utilization of a specific metric
type MetricTarget struct {
	// type represents whether the metric type is Utilization, Value, or AverageValue
	Type MetricTargetType `json:"type" protobuf:"bytes,1,name=type"`
	// value is the target value of the metric (as a quantity).
	// +optional
	Value *resource.Quantity `json:"value,omitempty" protobuf:"bytes,2,opt,name=value"`
	// averageValue is the target value of the average of the
	// metric across all relevant pods (as a quantity)
	// +optional
	AverageValue *resource.Quantity `json:"averageValue,omitempty" protobuf:"bytes,3,opt,name=averageValue"`
	// averageUtilization is the target value of the average of the
	// resource metric across all relevant pods, represented as a percentage of
	// the requested value of the resource for the pods.
	// Currently only valid for Resource metric source type
	// +optional
	AverageUtilization *int32 `json:"averageUtilization,omitempty" protobuf:"bytes,4,opt,name=averageUtilization"`
}

// MetricTargetType specifies the type of metric being targeted, and should be either
// "Value", "AverageValue", or "Utilization"
type MetricTargetType string

const (
	// UtilizationMetricType declares a MetricTarget is an AverageUtilization value
	UtilizationMetricType MetricTargetType = "Utilization"
	// ValueMetricType declares a MetricTarget is a raw value
	ValueMetricType MetricTargetType = "Value"
	// AverageValueMetricType declares a MetricTarget is an
	AverageValueMetricType MetricTargetType = "AverageValue"
)

// GeneralPodAutoscalerStatus describes the current status of a general pod autoscaler.
type GeneralPodAutoscalerStatus struct {
	// observedGeneration is the most recent generation observed by this autoscaler.
	// +optional
	ObservedGeneration *int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`

	// lastScaleTime is the last time the GeneralPodAutoscaler scaled the number of pods,
	// used by the autoscaler to control how often the number of pods is changed.
	// +optional
	LastScaleTime *metav1.Time `json:"lastScaleTime,omitempty" protobuf:"bytes,2,opt,name=lastScaleTime"`

	// currentReplicas is current number of replicas of pods managed by this autoscaler,
	// as last seen by the autoscaler.
	CurrentReplicas int32 `json:"currentReplicas" protobuf:"varint,3,opt,name=currentReplicas"`

	// desiredReplicas is the desired number of replicas of pods managed by this autoscaler,
	// as last calculated by the autoscaler.
	DesiredReplicas int32 `json:"desiredReplicas" protobuf:"varint,4,opt,name=desiredReplicas"`

	// currentMetrics is the last read state of the metrics used by this autoscaler.
	// +optional
	CurrentMetrics []MetricStatus `json:"currentMetrics,omitempty" protobuf:"bytes,5,rep,name=currentMetrics"`

	// conditions is the set of conditions required for this autoscaler to scale its target,
	// and indicates whether or not those conditions are met.
	Conditions []GeneralPodAutoscalerCondition `json:"conditions" protobuf:"bytes,6,rep,name=conditions"`

	// LastCronScheduleTime is the schedule time of time mode
	// +optional
	LastCronScheduleTime *metav1.Time `json:"lastCronScheduleTime,omitempty" protobuf:"bytes,7,rep,name=lastCronScheduleTime"`
}

// GeneralPodAutoscalerConditionType are the valid conditions of
// a GeneralPodAutoscaler.
type GeneralPodAutoscalerConditionType string

const (
	// ScalingActive indicates that the GPA controller is able to scale if necessary:
	// it's correctly configured, can fetch the desired metrics, and isn't disabled.
	ScalingActive GeneralPodAutoscalerConditionType = "ScalingActive"
	// AbleToScale indicates a lack of transient issues which prevent scaling from occurring,
	// such as being in a backoff window, or being unable to access/update the target scale.
	AbleToScale GeneralPodAutoscalerConditionType = "AbleToScale"
	// ScalingLimited indicates that the calculated scale based on metrics would be above or
	// below the range for the GPA, and has thus been capped.
	ScalingLimited GeneralPodAutoscalerConditionType = "ScalingLimited"
)

// GeneralPodAutoscalerCondition describes the state of
// a GeneralPodAutoscaler at a certain point.
type GeneralPodAutoscalerCondition struct {
	// type describes the current condition
	Type GeneralPodAutoscalerConditionType `json:"type" protobuf:"bytes,1,name=type"`
	// status is the status of the condition (True, False, Unknown)
	Status v1.ConditionStatus `json:"status" protobuf:"bytes,2,name=status"`
	// lastTransitionTime is the last time the condition transitioned from
	// one status to another
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// reason is the reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// message is a human-readable explanation containing details about
	// the transition
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// MetricStatus describes the last-read state of a single metric.
type MetricStatus struct {
	// type is the type of metric source.  It will be one of "Object",
	// "Pods" or "Resource", each corresponds to a matching field in the object.
	Type MetricSourceType `json:"type" protobuf:"bytes,1,name=type"`

	// object refers to a metric describing a single kubernetes object
	// (for example, hits-per-second on an Ingress object).
	// +optional
	Object *ObjectMetricStatus `json:"object,omitempty" protobuf:"bytes,2,opt,name=object"`
	// pods refers to a metric describing each pod in the current scale target
	// (for example, transactions-processed-per-second).  The values will be
	// averaged together before being compared to the target value.
	// +optional
	Pods *PodsMetricStatus `json:"pods,omitempty" protobuf:"bytes,3,opt,name=pods"`
	// resource refers to a resource metric (such as those specified in
	// requests and limits) known to Kubernetes describing each pod in the
	// current scale target (e.g. CPU or memory). Such metrics are built in to
	// Kubernetes, and have special scaling options on top of those available
	// to normal per-pod metrics using the "pods" source.
	// +optional
	Resource *ResourceMetricStatus `json:"resource,omitempty" protobuf:"bytes,4,opt,name=resource"`
	// external refers to a global metric that is not associated
	// with any Kubernetes object. It allows autoscaling based on information
	// coming from components running outside of cluster
	// (for example length of queue in cloud messaging service, or
	// QPS from loadbalancer running outside of cluster).
	// +optional
	External *ExternalMetricStatus `json:"external,omitempty" protobuf:"bytes,5,opt,name=external"`
	// to normal per-pod metrics using the "pods" source.
	// +optional
	ContainerResource *ContainerResourceMetricStatus `json:"containerResource,omitempty" protobuf:"bytes,6,opt,name=containerResource"`
}

// ObjectMetricStatus indicates the current value of a metric describing a
// kubernetes object (for example, hits-per-second on an Ingress object).
type ObjectMetricStatus struct {
	// metric identifies the target metric by name and selector
	Metric MetricIdentifier `json:"metric" protobuf:"bytes,1,name=metric"`
	// current contains the current value for the given metric
	Current MetricValueStatus `json:"current" protobuf:"bytes,2,name=current"`

	DescribedObject CrossVersionObjectReference `json:"describedObject" protobuf:"bytes,3,name=describedObject"`
}

// PodsMetricStatus indicates the current value of a metric describing each pod in
// the current scale target (for example, transactions-processed-per-second).
type PodsMetricStatus struct {
	// metric identifies the target metric by name and selector
	Metric MetricIdentifier `json:"metric" protobuf:"bytes,1,name=metric"`
	// current contains the current value for the given metric
	Current MetricValueStatus `json:"current" protobuf:"bytes,2,name=current"`
}

// ResourceMetricStatus indicates the current value of a resource metric known to
// Kubernetes, as specified in requests and limits, describing each pod in the
// current scale target (e.g. CPU or memory).  Such metrics are built in to
// Kubernetes, and have special scaling options on top of those available to
// normal per-pod metrics using the "pods" source.
type ResourceMetricStatus struct {
	// Name is the name of the resource in question.
	Name v1.ResourceName `json:"name" protobuf:"bytes,1,name=name"`
	// current contains the current value for the given metric
	Current MetricValueStatus `json:"current" protobuf:"bytes,2,name=current"`
}

// ContainerResourceMetricStatus xxx
// container resource refers to a resource metric (such as those specified in
// requests and limits) known to Kubernetes describing a single container in each pod in the
// current scale target (e.g. CPU or memory). Such metrics are built in to
// Kubernetes, and have special scaling options on top of those available
// to normal per-pod metrics using the "pods" source.
// +optional
type ContainerResourceMetricStatus struct {
	// Name is the name of the resource in question.
	Name v1.ResourceName `json:"name" protobuf:"bytes,1,name=name"`
	// current contains the current value for the given metric
	Current MetricValueStatus `json:"current" protobuf:"bytes,2,name=current"`
	// Container is the name of the container in the pods of the scaling target
	Container string `json:"container" protobuf:"bytes,3,opt,name=container"`
}

// ExternalMetricStatus indicates the current value of a global metric
// not associated with any Kubernetes object.
type ExternalMetricStatus struct {
	// metric identifies the target metric by name and selector
	Metric MetricIdentifier `json:"metric" protobuf:"bytes,1,name=metric"`
	// current contains the current value for the given metric
	Current MetricValueStatus `json:"current" protobuf:"bytes,2,name=current"`
}

// MetricValueStatus holds the current value for a metric
type MetricValueStatus struct {
	// value is the current value of the metric (as a quantity).
	// +optional
	Value *resource.Quantity `json:"value,omitempty" protobuf:"bytes,1,opt,name=value"`
	// averageValue is the current value of the average of the
	// metric across all relevant pods (as a quantity)
	// +optional
	AverageValue *resource.Quantity `json:"averageValue,omitempty" protobuf:"bytes,2,opt,name=averageValue"`
	// currentAverageUtilization is the current value of the average of the
	// resource metric across all relevant pods, represented as a percentage of
	// the requested value of the resource for the pods.
	// +optional
	AverageUtilization *int32 `json:"averageUtilization,omitempty" protobuf:"bytes,3,opt,name=averageUtilization"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:prerelease-lifecycle-gen:introduced=1.12
// +k8s:prerelease-lifecycle-gen:deprecated=1.22

// GeneralPodAutoscalerList is a list of general pod autoscaler objects.
type GeneralPodAutoscalerList struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is the standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// items is the list of general pod autoscaler objects.
	Items []GeneralPodAutoscaler `json:"items" protobuf:"bytes,2,rep,name=items"`
}
