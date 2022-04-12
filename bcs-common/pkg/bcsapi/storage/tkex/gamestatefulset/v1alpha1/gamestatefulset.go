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
 *
 */

// +kubebuilder:validation:Optional
package gamestatefulset

import (
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/update/inplaceupdate"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	//ControllerRevisionHashLabelKey controller revision hash tag
	ControllerRevisionHashLabelKey = "controller-revision-hash"
	//GameStatefulSetRevisionLabel controller revision hash tag
	GameStatefulSetRevisionLabel = ControllerRevisionHashLabelKey
	//DeprecatedRollbackTo deprecated label
	DeprecatedRollbackTo = "deprecated.deployment.rollback.to"
	//DeprecatedTemplateGeneration deprecated label
	DeprecatedTemplateGeneration = "deprecated.daemonset.template.generation"
	//GameStatefulSetPodNameLabel pod name reference label
	GameStatefulSetPodNameLabel = "gamestatefulset.kubernetes.io/pod-name"
	//GameStatefulSetPodOrdinal pod ordinal reference label
	GameStatefulSetPodOrdinal = "gamestatefulset.kubernetes.io/pod-ordinal"
)

// GameStatefulSet compatible with original StatefulSet but support in-place update additionally
// +genclient
// +genclient:method=GetScale,verb=get,subresource=scale,result=k8s.io/kubernetes/pkg/apis/autoscaling.Scale
// +genclient:method=UpdateScale,verb=update,subresource=scale,input=k8s.io/kubernetes/pkg/apis/autoscaling.Scale,result=k8s.io/kubernetes/pkg/apis/autoscaling.Scale
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:JSONPath=.spec.replicas,name=Replicas,type=integer
// +kubebuilder:printcolumn:JSONPath=.status.readyReplicas,name=Ready_Replicas,type=integer
// +kubebuilder:printcolumn:JSONPath=.status.currentReplicas,name=Current_Replicas,type=integer
// +kubebuilder:printcolumn:JSONPath=.status.updatedReplicas,name=Updated_Replicas,type=integer
// +kubebuilder:printcolumn:JSONPath=.status.updatedReadyReplicas,name=Updated_Ready_Replicas,type=integer
// +kubebuilder:printcolumn:JSONPath=.metadata.creationTimestamp,name=Age,type=date,description=Age of the gamestatefulset
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:selectorpath=.status.labelSelector,specpath=.spec.replicas,statuspath=.status.replicas
type GameStatefulSet struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,inline" protobuf:"bytes,1,opt,name=metadata"`

	// +kubebuilder:validation:Required
	Spec GameStatefulSetSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// +optional
	Status GameStatefulSetStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// PodManagementPolicyType defines the policy for creating pods under a gamestatefulset.
type PodManagementPolicyType string

const (
	// OrderedReadyPodManagement will create pods in strictly increasing order on
	// scale up and strictly decreasing order on scale down, progressing only when
	// the previous pod is ready or terminated. At most one pod will be changed
	// at any time.
	OrderedReadyPodManagement PodManagementPolicyType = "OrderedReady"
	// ParallelPodManagement will create and delete pods as soon as the stateful set
	// replica count is changed, and will not wait for pods to be ready or complete
	// termination.
	ParallelPodManagement = "Parallel"
)

// GameStatefulSetUpdateStrategy indicates the strategy that the StatefulSet
// controller will use to perform updates. It includes any additional parameters
// necessary to perform the update for the indicated strategy.
type GameStatefulSetUpdateStrategy struct {
	// Type indicates the type of the StatefulSetUpdateStrategy.
	// Default is RollingUpdate.
	// +kubebuilder:default=RollingUpdate
	// +optional
	// +kubebuilder:validation:Enum=RollingUpdate;OnDelete;InplaceUpdate;HotPatchUpdate
	Type GameStatefulSetUpdateStrategyType `json:"type,omitempty" protobuf:"bytes,1,opt,name=type,casttype=StatefulSetStrategyType"`
	// RollingUpdate is used to communicate parameters when Type is RollingUpdateStatefulSetStrategyType.
	// +optional
	RollingUpdate *RollingUpdateStatefulSetStrategy `json:"rollingUpdate,omitempty" protobuf:"bytes,2,opt,name=rollingUpdate"`

	// InPlaceUpdateStrategy contains strategies for in-place update.
	InPlaceUpdateStrategy *inplaceupdate.InPlaceUpdateStrategy `json:"inPlaceUpdateStrategy,omitempty"`
	CanaryStrategy        *CanaryStrategy                      `json:"canary,omitempty"`
	// Paused indicates that the GameStatefulSet is paused.
	// Default value is false
	// +kubebuilder:default=false
	Paused bool `json:"paused,omitempty"`
}

type CanaryStrategy struct {
	// +kubebuilder:validation:Required
	Steps []CanaryStep `json:"steps,omitempty"`
}

type CanaryStep struct {
	Partition *intstr.IntOrString    `json:"partition,omitempty"`
	Pause     *CanaryPause           `json:"pause,omitempty"`
	Hook      *hookv1alpha1.HookStep `json:"hook,omitempty"`
}

type CanaryPause struct {
	// Duration the amount of time to wait before moving to the next step.
	// +optional
	Duration *int32 `json:"duration,omitempty"`
}

// GameStatefulSetUpdateStrategyType is a string enumeration type that enumerates
// all possible update strategies for the StatefulSet controller.
type GameStatefulSetUpdateStrategyType string

const (
	// RollingUpdateGameStatefulSetStrategyType indicates that update will be
	// applied to all Pods in the StatefulSet with respect to the StatefulSet
	// ordering constraints. When a scale operation is performed with this
	// strategy, new Pods will be created from the specification version indicated
	// by the StatefulSet's updateRevision.
	RollingUpdateGameStatefulSetStrategyType = "RollingUpdate"
	// OnDeleteGameStatefulSetStrategyType triggers the legacy behavior. Version
	// tracking and ordered rolling restarts are disabled. Pods are recreated
	// from the StatefulSetSpec when they are manually deleted. When a scale
	// operation is performed with this strategy,specification version indicated
	// by the StatefulSet's currentRevision.
	OnDeleteGameStatefulSetStrategyType = "OnDelete"
	// InplaceUpdateGameStatefulSetStrategyType indicates that update will be
	// applied to all Pods in the StatefulSet with respect to the StatefulSet
	// ordering constraints. When a scale operation is performed with this
	// strategy, new Pods will be created from the specification version indicated
	// by the StatefulSet's updateRevision.
	InplaceUpdateGameStatefulSetStrategyType = "InplaceUpdate"
	// HotPatchGameStatefulSetStrategyType indicates that pods in the GameStatefulSet will be update hot-patch
	HotPatchGameStatefulSetStrategyType = "HotPatchUpdate"
)

// RollingUpdateStatefulSetStrategy is used to communicate parameter for RollingUpdateStatefulSetStrategyType.
type RollingUpdateStatefulSetStrategy struct {
	// Partition indicates the ordinal at which the StatefulSet should be
	// partitioned.
	// Default value is 0.
	// +kubebuilder:default=0
	// +optional
	Partition *intstr.IntOrString `json:"partition,omitempty" protobuf:"varint,1,opt,name=partition"`

	// The maximum number of pods that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding up. This can not be 0.
	// Defaults to 25%
	// +kubebuilder:default="25%"
	// +optional
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty" protobuf:"varint,2,opt,name=maxUnavailable"`

	// The maximum number of pods that can be scheduled above the desired replicas during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding up.
	// Defaults to 0.
	// +kubebuilder:default=0
	// +optional
	MaxSurge *intstr.IntOrString `json:"maxSurge,omitempty" protobuf:"varint,3,opt,name=maxSurge"`
}

// GameStatefulSetSpec A StatefulSetSpec is the specification of a StatefulSet.
type GameStatefulSetSpec struct {
	// replicas is the desired number of replicas of the given Template.
	// These are replicas in the sense that they are instantiations of the
	// same Template, but individual replicas also have a consistent identity.
	// If unspecified, defaults to 1.
	// TODO: Consider a rename of this field.
	// +kubebuilder:default=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	// +kubebuilder:validation:Required
	Selector *metav1.LabelSelector `json:"selector" protobuf:"bytes,2,opt,name=selector"`

	// template is the object that describes the pod that will be created if
	// insufficient replicas are detected. Each pod stamped out by the StatefulSet
	// will fulfill this Template, but have a unique identity from the rest
	// of the StatefulSet.
	// +kubebuilder:validation:Required
	Template core.PodTemplateSpec `json:"template" protobuf:"bytes,3,opt,name=template"`

	// volumeClaimTemplates is a list of claims that pods are allowed to reference.
	// The StatefulSet controller is responsible for mapping network identities to
	// claims in a way that maintains the identity of a pod. Every claim in
	// this list must have at least one matching (by name) volumeMount in one
	// container in the template. A claim in this list takes precedence over
	// any volumes in the template, with the same name.
	// TODO: Define the behavior if a claim already exists with the same name.
	// +optional
	VolumeClaimTemplates []core.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty" protobuf:"bytes,4,rep,name=volumeClaimTemplates"`

	// serviceName is the name of the service that governs this StatefulSet.
	// This service must exist before the StatefulSet, and is responsible for
	// the network identity of the set. Pods get DNS/hostnames that follow the
	// pattern: pod-specific-string.serviceName.default.svc.cluster.local
	// where "pod-specific-string" is managed by the StatefulSet controller.
	// +kubebuilder:validation:Required
	ServiceName string `json:"serviceName" protobuf:"bytes,5,opt,name=serviceName"`

	// podManagementPolicy controls how pods are created during initial scale up,
	// when replacing pods on nodes, or when scaling down. The default policy is
	// `OrderedReady`, where pods are created in increasing order (pod-0, then
	// pod-1, etc) and the controller will wait until each pod is ready before
	// continuing. When scaling down, the pods are removed in the opposite order.
	// The alternative policy is `Parallel` which will create pods in parallel
	// to match the desired scale without waiting, and on scale down will delete
	// all pods at once.
	// +optional
	// +kubebuilder:validation:Enum=OrderedReady;Parallel
	// +kubebuilder:default=OrderedReady
	PodManagementPolicy PodManagementPolicyType `json:"podManagementPolicy,omitempty" protobuf:"bytes,6,opt,name=podManagementPolicy,casttype=PodManagementPolicyType"`

	// updateStrategy indicates the StatefulSetUpdateStrategy that will be
	// employed to update Pods in the StatefulSet when a revision is made to
	// Template.
	UpdateStrategy GameStatefulSetUpdateStrategy `json:"updateStrategy,omitempty" protobuf:"bytes,7,opt,name=updateStrategy"`

	// PreDeleteUpdateStrategy indicates the PreDeleteUpdateStrategy that will be employed to
	// before Delete Or Update Pods
	PreDeleteUpdateStrategy GameStatefulSetPreDeleteUpdateStrategy `json:"preDeleteUpdateStrategy,omitempty"`

	// PreInplaceUpdateStrategy indicates the PreInplaceUpdateStrategy that will be employed to
	// before Delete Or Update Pods
	PreInplaceUpdateStrategy GameStatefulSetPreInplaceUpdateStrategy `json:"preInplaceUpdateStrategy,omitempty"`

	// PostInplaceUpdateStrategy indicates the PostInplaceUpdateStrategy that will be employed to
	// after Delete Or Update Pods
	PostInplaceUpdateStrategy GameDeploymentPostInplaceUpdateStrategy `json:"postInplaceUpdateStrategy,omitempty"`

	// revisionHistoryLimit is the maximum number of revisions that will
	// be maintained in the StatefulSet's revision history. The revision history
	// consists of all revisions not represented by a currently applied
	// StatefulSetSpec version. The default value is 10.
	// +kubebuilder:default=10
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty" protobuf:"varint,8,opt,name=revisionHistoryLimit"`
}

type GameStatefulSetPreDeleteUpdateStrategy struct {
	// +kubebuilder:validation:Required
	Hook                 *hookv1alpha1.HookStep `json:"hook,omitempty"`
	RetryUnexpectedHooks bool                   `json:"retry,omitempty"`
}

type GameStatefulSetPreInplaceUpdateStrategy struct {
	// +kubebuilder:validation:Required
	Hook                 *hookv1alpha1.HookStep `json:"hook,omitempty"`
	RetryUnexpectedHooks bool                   `json:"retry,omitempty"`
}

// GameDeploymentPostInplaceUpdateStrategy defines the structure of PostInplaceUpdateStrategy
type GameDeploymentPostInplaceUpdateStrategy struct {
	// +kubebuilder:validation:Required
	Hook                 *hookv1alpha1.HookStep `json:"hook,omitempty"`
	RetryUnexpectedHooks bool                   `json:"retry,omitempty"`
}

// GameStatefulSetStatus represents the current state of a StatefulSet.
type GameStatefulSetStatus struct {
	// observedGeneration is the most recent generation observed for this StatefulSet. It corresponds to the
	// StatefulSet's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`

	// replicas is the number of Pods created by the StatefulSet controller.
	Replicas int32 `json:"replicas" protobuf:"varint,2,opt,name=replicas"`

	// readyReplicas is the number of Pods created by the StatefulSet controller that have a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas,omitempty" protobuf:"varint,3,opt,name=readyReplicas"`

	// currentReplicas is the number of Pods created by the StatefulSet controller from the StatefulSet version
	// indicated by currentRevision.
	CurrentReplicas int32 `json:"currentReplicas,omitempty" protobuf:"varint,4,opt,name=currentReplicas"`

	// updatedReplicas is the number of Pods created by the StatefulSet controller from the StatefulSet version
	// indicated by updateRevision.
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty" protobuf:"varint,5,opt,name=updatedReplicas"`

	// UpdatedReadyReplicas is the number of Pods created by the StatefulSet controller from the StatefulSet version
	// indicated by updateRevision and have a Ready Condition.
	UpdatedReadyReplicas int32 `json:"updatedReadyReplicas"`

	// currentRevision, if not empty, indicates the version of the StatefulSet used to generate Pods in the
	// sequence [0,currentReplicas).
	CurrentRevision string `json:"currentRevision,omitempty" protobuf:"bytes,6,opt,name=currentRevision"`

	// updateRevision, if not empty, indicates the version of the StatefulSet used to generate Pods in the sequence
	// [replicas-updatedReplicas,replicas)
	UpdateRevision string `json:"updateRevision,omitempty" protobuf:"bytes,7,opt,name=updateRevision"`

	// collisionCount is the count of hash collisions for the StatefulSet. The StatefulSet controller
	// uses this field as a collision avoidance mechanism when it needs to create the name for the
	// newest ControllerRevision.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty" protobuf:"varint,9,opt,name=collisionCount"`

	// Represents the latest available observations of a statefulset's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []GameStatefulSetCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,10,rep,name=conditions"`

	// +optional
	LabelSelector *string `json:"labelSelector" protobuf:"bytes 12,opt,name=labelSelector"`

	PauseConditions           []hookv1alpha1.PauseCondition           `json:"pauseConditions,omitempty"`
	CurrentStepIndex          *int32                                  `json:"currentStepIndex,omitempty"`
	CurrentStepHash           string                                  `json:"currentStepHash,omitempty"`
	Canary                    CanaryStatus                            `json:"canary,omitempty"`
	PreDeleteHookConditions   []hookv1alpha1.PreDeleteHookCondition   `json:"preDeleteHookCondition,omitempty"`
	PreInplaceHookConditions  []hookv1alpha1.PreInplaceHookCondition  `json:"preInplaceHookCondition,omitempty"`
	PostInplaceHookConditions []hookv1alpha1.PostInplaceHookCondition `json:"postInplaceHookCondition,omitempty"`
}

type CanaryStatus struct {
	Revision           string       `json:"revision,omitempty"`
	PauseStartTime     *metav1.Time `json:"pauseStartTime,omitempty"`
	CurrentStepHookRun string       `json:"currentStepHookRun,omitempty"`
}

//GameStatefulSetConditionType condition type for statefulset
type GameStatefulSetConditionType string

// GameStatefulSetCondition describes the state of a statefulset at a certain point.
type GameStatefulSetCondition struct {
	// Type of statefulset condition.
	Type GameStatefulSetConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=StatefulSetConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status core.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,3,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GameStatefulSetList is a collection of StatefulSets.
type GameStatefulSetList struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Items []GameStatefulSet `json:"items" protobuf:"bytes,2,rep,name=items"`
}
