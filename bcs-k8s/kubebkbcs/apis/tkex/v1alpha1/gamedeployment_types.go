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

package v1alpha1

import (
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// InPlaceUpdateReady must be added into template.spec.readinessGates when pod podUpdatePolicy
	// is InPlaceIfPossible or InPlaceOnly. The condition in podStatus will be updated to False before in-place
	// updating and updated to True after the update is finished. This ensures pod to remain at NotReady state while
	// in-place update is happening.
	InPlaceUpdateReady core.PodConditionType = "InPlaceUpdateReady"

	// InPlaceUpdateStateKey records the state of inplace-update.
	// The value of annotation is InPlaceUpdateState.
	InPlaceUpdateStateKey string = "inplace-update-state"

	// InPlaceUpdateGraceKey records the spec that Pod should be updated when
	// grace period ends.
	InPlaceUpdateGraceKey string = "inplace-update-grace"
)

// InPlaceUpdateState records latest inplace-update state, including old statuses of containers.
type InPlaceUpdateState struct {
	// Revision is the updated revision hash.
	Revision string `json:"revision"`

	// UpdateTimestamp is the time when the in-place update happens.
	UpdateTimestamp metav1.Time `json:"updateTimestamp"`

	// LastContainerStatuses records the before-in-place-update container statuses. It is a map from ContainerName
	// to InPlaceUpdateContainerStatus
	LastContainerStatuses map[string]InPlaceUpdateContainerStatus `json:"lastContainerStatuses"`
}

// InPlaceUpdateContainerStatus records the statuses of the container that are mainly used
// to determine whether the InPlaceUpdate is completed.
type InPlaceUpdateContainerStatus struct {
	ImageID string `json:"imageID,omitempty"`
}

// InPlaceUpdateStrategy defines the strategies for in-place update.
type InPlaceUpdateStrategy struct {
	// GracePeriodSeconds is the timespan between set Pod status to not-ready and update images in Pod spec
	// when in-place update a Pod.
	GracePeriodSeconds int32 `json:"gracePeriodSeconds,omitempty"`
}

const (
	// GameDeploymentInstanceID is a unique id for Pods and PVCs.
	// Each pod and the pvcs it owns have the same instance-id.
	GameDeploymentInstanceID = "tkex.bkbcs.tencent.com/gamedeployment-instance-id"

	// DefaultGameDeploymentMaxUnavailable is the default value of maxUnavailable for GameDeployment update strategy.
	DefaultGameDeploymentMaxUnavailable = "20%"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GameDeploymentSpec defines the desired state of GameDeployment
type GameDeploymentSpec struct {
	// replicas is the desired number of replicas of the given Template.
	// These are replicas in the sense that they are instantiations of the
	// same Template, but individual replicas also have a consistent identity.
	// If unspecified, defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	// +kubebuilder:validation:Required
	Selector *metav1.LabelSelector `json:"selector" protobuf:"bytes,2,opt,name=selector"`

	// template is the object that describes the pod that will be created
	// +kubebuilder:validation:Required
	Template core.PodTemplateSpec `json:"template" protobuf:"bytes,3,opt,name=template"`

	// ScaleStrategy indicates the ScaleStrategy that will be employed to
	// create and delete Pods in the GameDeployment.
	ScaleStrategy GameDeploymentScaleStrategy `json:"scaleStrategy,omitempty"`

	// UpdateStrategy indicates the UpdateStrategy that will be employed to
	// update Pods in the GameDeployment when a revision is made to Template.
	UpdateStrategy GameDeploymentUpdateStrategy `json:"updateStrategy,omitempty"`

	// PreDeleteUpdateStrategy indicates the PreDeleteUpdateStrategy that will be employed to
	// before Delete Or Update Pods
	PreDeleteUpdateStrategy GameDeploymentPreDeleteUpdateStrategy `json:"preDeleteUpdateStrategy,omitempty"`

	// PreInplaceUpdateStrategy indicates the PreInplaceUpdateStrategy that will be employed to
	// before Delete Or Update Pods
	PreInplaceUpdateStrategy GameDeploymentPreInplaceUpdateStrategy `json:"preInplaceUpdateStrategy,omitempty"`

	// RevisionHistoryLimit is the maximum number of revisions that will
	// be maintained in the GameDeployment's revision history. The revision history
	// consists of all revisions not represented by a currently applied
	// GameDeploymentSpec version. The default value is 10.
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty"`

	// Minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing, for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	MinReadySeconds int32 `json:"minReadySeconds,omitempty"`
}

type GameDeploymentPreDeleteUpdateStrategy struct {
	Hook                 *HookStep `json:"hook,omitempty"`
	RetryUnexpectedHooks bool                   `json:"retry,omitempty"`
}

type GameDeploymentPreInplaceUpdateStrategy struct {
	Hook                 *HookStep `json:"hook,omitempty"`
	RetryUnexpectedHooks bool                   `json:"retry,omitempty"`
}

type GameDeploymentScaleStrategy struct {
	// PodsToDelete is the names of Pod should be deleted.
	// Note that this list will be truncated for non-existing pod names.
	PodsToDelete []string `json:"podsToDelete,omitempty"`
}

type GameDeploymentUpdateStrategy struct {
	// Type indicates the type of the GameDeploymentUpdateStrategy.
	// Default is ReCreate.
	// +kubebuilder:validation:Enum=RollingUpdate;InplaceUpdate;HotPatchUpdate
	Type           GameDeploymentUpdateStrategyType `json:"type,omitempty"`
	CanaryStrategy *CanaryStrategy                  `json:"canary,omitempty"`
	// Partition is the desired number of pods in old revisions. It means when partition
	// is set during pods updating, (replicas - partition) number of pods will be updated.
	// Default value is 0.
	Partition *int32 `json:"partition,omitempty"`
	// The maximum number of pods that can be unavailable during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding up by default.
	// When maxSurge > 0, absolute number is calculated from percentage by rounding down.
	// Defaults to 20%.
	MaxUnavailable *intstr.IntOrString `json:"maxUnavailable,omitempty"`
	// The maximum number of pods that can be scheduled above the desired replicas during the update.
	// Value can be an absolute number (ex: 5) or a percentage of desired pods (ex: 10%).
	// Absolute number is calculated from percentage by rounding up.
	// Defaults to 0.
	MaxSurge *intstr.IntOrString `json:"maxSurge,omitempty"`
	// Paused indicates that the GameDeployment is paused.
	// Default value is false
	Paused bool `json:"paused,omitempty"`
	// InPlaceUpdateStrategy contains strategies for in-place update.
	InPlaceUpdateStrategy *InPlaceUpdateStrategy `json:"inPlaceUpdateStrategy,omitempty"`
}

type CanaryStrategy struct {
	// +kubebuilder:validation:Required
	Steps []CanaryStep `json:"steps,omitempty"`
}

type CanaryStep struct {
	Partition *int32                 `json:"partition,omitempty"`
	Pause     *CanaryPause           `json:"pause,omitempty"`
	Hook      *HookStep `json:"hook,omitempty"`
}

type CanaryPause struct {
	// Duration the amount of time to wait before moving to the next step.
	// +optional
	Duration *int32 `json:"duration,omitempty"`
}

// GameDeploymentUpdateStrategyType defines strategies for pods in-place update.
type GameDeploymentUpdateStrategyType string

const (
	// RollingGameDeploymentUpdateStrategyType indicates that we always delete Pod and create new Pod
	// during Pod update, which is the default behavior.
	RollingGameDeploymentUpdateStrategyType GameDeploymentUpdateStrategyType = "RollingUpdate"

	// InPlaceGameDeploymentUpdateStrategyType indicates that we will in-place update Pod instead of
	// recreating pod. Currently we only allow image update for pod spec. Any other changes to the pod spec will be
	// rejected by kube-apiserver
	InPlaceGameDeploymentUpdateStrategyType GameDeploymentUpdateStrategyType = "InplaceUpdate"

	// HotPatchGameDeploymentUpdateStrategyType indicates that we will hot patch container image with pod being active.
	// Currently we only allow image update for pod spec. Any other changes to the pod spec will be
	// rejected by kube-apiserver
	HotPatchGameDeploymentUpdateStrategyType GameDeploymentUpdateStrategyType = "HotPatchUpdate"
)

// GameDeploymentStatus defines the observed state of GameDeployment
type GameDeploymentStatus struct {
	// ObservedGeneration is the most recent generation observed for this GameDeployment. It corresponds to the
	// GameDeployment's generation, which is updated on mutation by the API Server.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Replicas is the number of Pods created by the GameDeployment controller.
	Replicas int32 `json:"replicas"`

	// ReadyReplicas is the number of Pods created by the GameDeployment controller that have a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas"`

	// AvailableReplicas is the number of Pods created by the GameDeployment controller that have a Ready Condition for at least minReadySeconds.
	AvailableReplicas int32 `json:"availableReplicas"`

	// UpdatedReplicas is the number of Pods created by the GameDeployment controller from the GameDeployment version
	// indicated by updateRevision.
	UpdatedReplicas int32 `json:"updatedReplicas"`

	// UpdatedReadyReplicas is the number of Pods created by the GameDeployment controller from the GameDeployment version
	// indicated by updateRevision and have a Ready Condition.
	UpdatedReadyReplicas int32 `json:"updatedReadyReplicas"`

	// UpdateRevision, if not empty, indicates the latest revision of the GameDeployment.
	UpdateRevision string `json:"updateRevision,omitempty"`

	// CollisionCount is the count of hash collisions for the GameDeployment. The GameDeployment controller
	// uses this field as a collision avoidance mechanism when it needs to create the name for the
	// newest ControllerRevision.
	CollisionCount *int32 `json:"collisionCount,omitempty"`

	// Conditions represents the latest available observations of a GameDeployment's current state.
	Conditions []GameDeploymentCondition `json:"conditions,omitempty"`

	PauseConditions []PauseCondition `json:"pauseConditions,omitempty"`

	// LabelSelector is label selectors for query over pods that should match the replica count used by HPA.
	LabelSelector string `json:"labelSelector,omitempty"`

	CurrentStepIndex        *int32                                `json:"currentStepIndex,omitempty"`
	CurrentStepHash         string                                `json:"currentStepHash,omitempty"`
	Canary                  CanaryStatus                          `json:"canary,omitempty"`
	PreDeleteHookConditions []PreDeleteHookCondition `json:"preDeleteHookCondition,omitempty"`
	PreInplaceHookConditions []PreInplaceHookCondition `json:"preInplaceHookCondition,omitempty"`
}

type CanaryStatus struct {
	Revision           string       `json:"revision,omitempty"`
	PauseStartTime     *metav1.Time `json:"pauseStartTime,omitempty"`
	CurrentStepHookRun string       `json:"currentStepHookRun,omitempty"`
}

// GameDeploymentConditionType is type for GameDeployment conditions.
type GameDeploymentConditionType string

const (
	// GameDeploymentConditionFailedScale indicates GameDeployment controller failed to create or delete pods/pvc.
	GameDeploymentConditionFailedScale GameDeploymentConditionType = "FailedScale"
	// GameDeploymentConditionFailedUpdate indicates GameDeployment controller failed to update pods.
	GameDeploymentConditionFailedUpdate GameDeploymentConditionType = "FailedUpdate"
)

// GameDeploymentCondition describes the state of a GameDeployment at a certain point.
type GameDeploymentCondition struct {
	// Type of GameDeployment condition.
	Type GameDeploymentConditionType `json:"type"`
	// Status of the condition, one of True, False, Unknown.
	Status core.ConditionStatus `json:"status"`
	// Last time the condition transitioned from one status to another.
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}

// GameDeployment is the Schema for the gamedeployments API
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:printcolumn:JSONPath=.spec.replicas,name=DESIRED,type=integer,description=The desired number of pods.
// +kubebuilder:printcolumn:JSONPath=.status.updatedReplicas,name=UPDATED,type=integer,description=The number of pods updated.
// +kubebuilder:printcolumn:JSONPath=.status.updatedReadyReplicas,name=UPDATED_READY,type=integer,description=The number of pods updated and ready.
// +kubebuilder:printcolumn:JSONPath=.status.readyReplicas,name=READY,type=integer,description=The number of pods ready.
// +kubebuilder:printcolumn:JSONPath=.status.replicas,name=TOTAL,type=integer,description=The number of currently all pods.
// +kubebuilder:printcolumn:JSONPath=.metadata.creationTimestamp,name=Age,type=date,description=CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC.
// +kubebuilder:subresource:status
// +kubebuilder:subresource:scale:selectorpath=.status.labelSelector,specpath=.spec.replicas,statuspath=.status.replicas
type GameDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:Required
	Spec   GameDeploymentSpec   `json:"spec,omitempty"`
	Status GameDeploymentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true

// GameDeploymentList contains a list of GameDeployment
type GameDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GameDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GameDeployment{}, &GameDeploymentList{})
}
