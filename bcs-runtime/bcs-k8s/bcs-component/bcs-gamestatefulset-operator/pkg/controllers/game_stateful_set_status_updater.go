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

package gamestatefulset

import (
	"context"
	"time"

	gstsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	gstsclientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/client/clientset/versioned"
	gstslister "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/client/listers/tkex/v1alpha1"
	canaryutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/util/canary"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	commondiffutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/diff"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/hook"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	patchtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"k8s.io/utils/pointer"
)

// GameStatefulSetStatusUpdaterInterface is an interface used to update the GameStatefulSetStatus associated with a GameStatefulSet.
// For any use other than testing, clients should create an instance using NewRealGameStatefulSetStatusUpdater.
type GameStatefulSetStatusUpdaterInterface interface {
	// UpdateGameStatefulSetStatus sets the set's Status to status. Implementations are required to retry on conflicts,
	// but fail on other errors. If the returned error is nil set's Status has been successfully set to status.
	UpdateGameStatefulSetStatus(set *gstsv1alpha1.GameStatefulSet, canaryCtx *canaryContext) error
}

// NewRealGameStatefulSetStatusUpdater returns a GameStatefulSetStatusUpdaterInterface that updates the Status of a GameStatefulSet,
// using the supplied client and setLister.
func NewRealGameStatefulSetStatusUpdater(
	gstsClient gstsclientset.Interface, setLister gstslister.GameStatefulSetLister, recorder record.EventRecorder) GameStatefulSetStatusUpdaterInterface {
	return &realGameStatefulSetStatusUpdater{gstsClient, setLister, recorder}
}

// realGameStatefulSetStatusUpdater updater implementation
type realGameStatefulSetStatusUpdater struct {
	gstsClient gstsclientset.Interface
	setLister  gstslister.GameStatefulSetLister
	recorder   record.EventRecorder
}

// UpdateGameStatefulSetStatus update gamesatefulset status
func (ssu *realGameStatefulSetStatusUpdater) UpdateGameStatefulSetStatus(
	set *gstsv1alpha1.GameStatefulSet,
	canaryCtx *canaryContext) error {

	currentStep, currentStepIndex := canaryutil.GetCurrentCanaryStep(set)
	canaryCtx.newStatus.Canary.Revision = set.Status.Canary.Revision
	if set.Spec.UpdateStrategy.CanaryStrategy != nil {
		canaryCtx.newStatus.CurrentStepHash = canaryutil.ComputeStepHash(set)
	}
	var stepCount int32
	if set.Spec.UpdateStrategy.CanaryStrategy != nil {
		stepCount = int32(len(set.Spec.UpdateStrategy.CanaryStrategy.Steps))
	}

	// if canary step hash changes, reset current step index
	if canaryutil.CheckStepHashChange(set) {
		canaryCtx.newStatus.CurrentStepIndex = canaryutil.ResetCurrentStepIndex(set)
		if set.Status.Canary.Revision == canaryCtx.newStatus.UpdateRevision {
			if canaryCtx.newStatus.CurrentStepIndex != nil {
				klog.Info("Skipping all steps because already been the update revision")
				canaryCtx.newStatus.CurrentStepIndex = pointer.Int32Ptr(stepCount)
				ssu.recorder.Eventf(set, corev1.EventTypeNormal, "SetStepIndex", "Set Step Index to %d", int(stepCount))
			}
		}
		return ssu.updateStatus(set, canaryCtx.newStatus, pointer.BoolPtr(false))
	}

	// if pod template change, reset current step index
	if canaryutil.CheckRevisionChange(set, canaryCtx.newStatus.UpdateRevision) {
		canaryCtx.newStatus.CurrentStepIndex = canaryutil.ResetCurrentStepIndex(set)
		if set.Status.Canary.Revision == canaryCtx.newStatus.UpdateRevision {
			if canaryCtx.newStatus.CurrentStepIndex != nil {
				klog.Info("Skipping all steps because already been the update revision")
				canaryCtx.newStatus.CurrentStepIndex = pointer.Int32Ptr(stepCount)
				ssu.recorder.Eventf(set, corev1.EventTypeNormal, "SetStepIndex", "Set Step Index to %d", int(stepCount))
			}
		}
		return ssu.updateStatus(set, canaryCtx.newStatus, pointer.BoolPtr(false))
	}

	if set.Status.Canary.Revision == "" {
		if set.Spec.UpdateStrategy.CanaryStrategy == nil {
			return ssu.updateStatus(set, canaryCtx.newStatus, pointer.BoolPtr(false))
		}
		canaryCtx.newStatus.Canary.Revision = canaryCtx.newStatus.UpdateRevision
		if stepCount > 0 {
			if stepCount != *currentStepIndex {
				ssu.recorder.Eventf(set, corev1.EventTypeNormal, "SetStepIndex", "Set Step Index to %d", int(stepCount))
			}
			canaryCtx.newStatus.CurrentStepIndex = &stepCount
		}
		return ssu.updateStatus(set, canaryCtx.newStatus, pointer.BoolPtr(false))
	}

	if stepCount == 0 {
		klog.Info("GameStatefulSet has no steps")
		canaryCtx.newStatus.Canary.Revision = canaryCtx.newStatus.UpdateRevision
		return ssu.updateStatus(set, canaryCtx.newStatus, pointer.BoolPtr(false))
	}

	if *currentStepIndex == stepCount {
		klog.Info("GameStatefulSet has executed every step")
		canaryCtx.newStatus.CurrentStepIndex = &stepCount
		canaryCtx.newStatus.Canary.Revision = canaryCtx.newStatus.UpdateRevision
		return ssu.updateStatus(set, canaryCtx.newStatus, pointer.BoolPtr(false))
	}

	if completeCurrentCanaryStep(set, canaryCtx) {
		*currentStepIndex++
		canaryCtx.newStatus.CurrentStepIndex = currentStepIndex
		canaryCtx.newStatus.Canary.CurrentStepHookRun = ""
		if int(*currentStepIndex) == len(set.Spec.UpdateStrategy.CanaryStrategy.Steps) {
			canaryCtx.newStatus.Canary.Revision = canaryCtx.newStatus.UpdateRevision
		}
		klog.Infof("Incrementing the Current Step Index to %d", *currentStepIndex)
		ssu.recorder.Eventf(set, corev1.EventTypeNormal, "SetStepIndex", "Set Step Index to %d", int(*currentStepIndex))
		return ssu.updateStatus(set, canaryCtx.newStatus, pointer.BoolPtr(false))
	}

	paused := set.Spec.UpdateStrategy.Paused

	pauseCondition := canaryutil.GetPauseCondition(set, hookv1alpha1.PauseReasonCanaryPauseStep)
	if currentStep != nil && currentStep.Pause != nil {
		currentPartition, err := canaryutil.GetCurrentPartition(set)
		if err != nil {
			return err
		}
		if pauseCondition == nil && canaryCtx.newStatus.UpdatedReadyReplicas == *set.Spec.Replicas-int32(currentPartition) {
			canaryCtx.AddPauseCondition(hookv1alpha1.PauseReasonCanaryPauseStep)
		}
	}

	canaryCtx.newStatus.CurrentStepIndex = currentStepIndex
	paused = ssu.calculateConditionStatus(set, canaryCtx)

	return ssu.updateStatus(set, canaryCtx.newStatus, &paused)
}

// completeCurrentCanaryStep checks whether have already complete current canary step
func completeCurrentCanaryStep(set *gstsv1alpha1.GameStatefulSet, canaryCtx *canaryContext) bool {
	currentStep, _ := canaryutil.GetCurrentCanaryStep(set)
	if currentStep == nil {
		return false
	}

	pauseCondition := canaryutil.GetPauseCondition(set, hookv1alpha1.PauseReasonCanaryPauseStep)

	if currentStep.Pause != nil && currentStep.Pause.Duration != nil {
		now := metav1.Now()
		if pauseCondition != nil {
			expiredTime := pauseCondition.StartTime.Add(time.Duration(*currentStep.Pause.Duration) * time.Second)
			if now.After(expiredTime) {
				klog.Info("GameStatefulSet has waited the duration of the pause step")
				return true
			}
		}
	}

	if currentStep.Pause != nil && currentStep.Pause.Duration == nil && pauseCondition != nil && !set.Spec.UpdateStrategy.Paused {
		klog.Info("GameStatefulSet has been unpaused")
		return true
	}

	partition, err := canaryutil.GetCurrentPartition(set)
	if err != nil {
		return false
	}
	if currentStep.Partition != nil && canaryCtx.newStatus.UpdatedReadyReplicas == *set.Spec.Replicas-int32(partition) {
		klog.Info("GameStatefulSet has reached the desired state for the correct partition")
		return true
	}

	currentHrs := canaryCtx.CurrentHookRuns()
	currentStepHr := commonhookutil.GetCurrentStepHookRun(currentHrs)
	hrExistsAndCompleted := currentStepHr != nil && currentStepHr.Status.Phase.Completed()
	if currentStep.Hook != nil && hrExistsAndCompleted && currentStepHr.Status.Phase == hookv1alpha1.HookPhaseSuccessful {
		return true
	}

	pauseConditionByHook := canaryutil.GetPauseCondition(set, hookv1alpha1.PauseReasonStepBasedHook)
	if currentStep.Hook != nil && pauseConditionByHook != nil && !set.Spec.UpdateStrategy.Paused {
		klog.Info("GameStatefulSet has been unpaused")
		return true
	}

	return false
}

// calculateConditionStatus calculate condition of GameStatefulSet, return true if exist pause condition
func (ssu *realGameStatefulSetStatusUpdater) calculateConditionStatus(deploy *gstsv1alpha1.GameStatefulSet, canaryCtx *canaryContext) bool {
	newPauseConditions := []hookv1alpha1.PauseCondition{}
	pauseAlreadyExists := map[hookv1alpha1.PauseReason]bool{}
	for _, cond := range deploy.Status.PauseConditions {
		newPauseConditions = append(newPauseConditions, cond)
		pauseAlreadyExists[cond.Reason] = true
	}
	now := metav1.Now()
	for _, reason := range canaryCtx.pauseReasons {
		if exists := pauseAlreadyExists[reason]; !exists {
			cond := hookv1alpha1.PauseCondition{
				Reason:    reason,
				StartTime: now,
			}
			newPauseConditions = append(newPauseConditions, cond)
		}
	}

	if len(newPauseConditions) == 0 {
		return false
	}
	canaryCtx.newStatus.PauseConditions = newPauseConditions
	return true
}

// updateStatus update status and updateStrategy pause of a GameStatefulSet to k8s
func (ssu *realGameStatefulSetStatusUpdater) updateStatus(set *gstsv1alpha1.GameStatefulSet, newStatus *gstsv1alpha1.GameStatefulSetStatus, newPause *bool) error {
	specCopy := set.Spec.DeepCopy()
	paused := specCopy.UpdateStrategy.Paused
	if newPause != nil {
		paused = *newPause
	}

	specPatch, specModified, err := commondiffutil.CreateTwoWayMergePatch(
		&gstsv1alpha1.GameStatefulSet{
			Spec: gstsv1alpha1.GameStatefulSetSpec{
				UpdateStrategy: gstsv1alpha1.GameStatefulSetUpdateStrategy{
					Paused: set.Spec.UpdateStrategy.Paused,
				},
				PreDeleteUpdateStrategy: gstsv1alpha1.GameStatefulSetPreDeleteUpdateStrategy{
					RetryUnexpectedHooks: set.Spec.PreDeleteUpdateStrategy.RetryUnexpectedHooks,
				},
				PreInplaceUpdateStrategy: gstsv1alpha1.GameStatefulSetPreInplaceUpdateStrategy{
					RetryUnexpectedHooks: set.Spec.PreInplaceUpdateStrategy.RetryUnexpectedHooks,
				},
			},
		},
		&gstsv1alpha1.GameStatefulSet{
			Spec: gstsv1alpha1.GameStatefulSetSpec{
				UpdateStrategy: gstsv1alpha1.GameStatefulSetUpdateStrategy{
					Paused: paused,
				},
				PreDeleteUpdateStrategy: gstsv1alpha1.GameStatefulSetPreDeleteUpdateStrategy{
					RetryUnexpectedHooks: false,
				},
				PreInplaceUpdateStrategy: gstsv1alpha1.GameStatefulSetPreInplaceUpdateStrategy{
					RetryUnexpectedHooks: false,
				},
			},
		}, gstsv1alpha1.GameStatefulSet{})
	if err != nil {
		klog.Errorf("Error constructing app status patch: %v", err)
		return err
	}
	if specModified {
		klog.Infof("GameStatefulSet Spec Patch: %s", specPatch)
		_, err = ssu.gstsClient.TkexV1alpha1().GameStatefulSets(set.Namespace).Patch(context.TODO(),
			set.Name, patchtypes.MergePatchType, specPatch, metav1.PatchOptions{})
		if err != nil {
			klog.Warningf("Error updating GameStatefulSet Spec: %v", err)
			return err
		}
		klog.Info("Patch spec successfully")
	}

	statusPatch, statusModified, err := commondiffutil.CreateTwoWayMergePatch(
		&gstsv1alpha1.GameStatefulSet{
			Status: set.Status,
		},
		&gstsv1alpha1.GameStatefulSet{
			Status: *newStatus,
		}, gstsv1alpha1.GameStatefulSet{})
	if err != nil {
		klog.Errorf("Error constructing app status patch: %v", err)
		return err
	}
	if !statusModified {
		klog.Info("No status changes. Skipping patch")
		return nil
	}
	klog.Infof("Rollout Patch: %s", statusPatch)
	_, err = ssu.gstsClient.TkexV1alpha1().GameStatefulSets(set.Namespace).Patch(context.TODO(),
		set.Name, patchtypes.MergePatchType, statusPatch, metav1.PatchOptions{}, "status")
	if err != nil {
		klog.Warningf("Error updating application: %v", err)
		return err
	}
	klog.Info("Patch status successfully")
	return nil
}

var _ GameStatefulSetStatusUpdaterInterface = &realGameStatefulSetStatusUpdater{}
