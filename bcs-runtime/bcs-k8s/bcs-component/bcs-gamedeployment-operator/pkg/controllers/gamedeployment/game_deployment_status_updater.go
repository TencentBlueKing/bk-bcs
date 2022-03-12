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

package gamedeployment

import (
	"context"
	"time"

	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	gdclientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/clientset/versioned"
	gdlister "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/client/listers/tkex/v1alpha1"
	gdmetrics "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util"
	canaryutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util/canary"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	commondiffutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/diff"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/hook"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	patchtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"k8s.io/utils/pointer"
)

// GameDeploymentStatusUpdaterInterface is an interface used to update the GameDeploymentStatus associated with a GameDeployment.
// For any use other than testing, clients should create an instance using NewRealGameDeploymentStatusUpdater.
type GameDeploymentStatusUpdaterInterface interface {
	// UpdateGameDeploymentStatus sets the set's Status to status. Implementations are required to retry on conflicts,
	// but fail on other errors. If the returned error is nil set's Status has been successfully set to status.
	UpdateGameDeploymentStatus(deploy *gdv1alpha1.GameDeployment, canaryCtx *canaryContext, pods []*v1.Pod) error
}

// NewRealGameDeploymentStatusUpdater returns a GameDeploymentStatusUpdaterInterface that updates the Status of a GameDeployment,
// using the supplied gdClient and setLister.
func NewRealGameDeploymentStatusUpdater(
	tkexClient gdclientset.Interface,
	setLister gdlister.GameDeploymentLister,
	record record.EventRecorder,
	metrics *gdmetrics.Metrics) GameDeploymentStatusUpdaterInterface {
	return &realGameDeploymentStatusUpdater{tkexClient, setLister, record, metrics}
}

type realGameDeploymentStatusUpdater struct {
	gdClient  gdclientset.Interface
	setLister gdlister.GameDeploymentLister
	recorder  record.EventRecorder
	metrics   *gdmetrics.Metrics
}

func (r *realGameDeploymentStatusUpdater) UpdateGameDeploymentStatus(
	deploy *gdv1alpha1.GameDeployment, canaryCtx *canaryContext, pods []*v1.Pod) error {

	r.calculateBaseStatus(deploy, canaryCtx, pods)

	currentStep, currentStepIndex := canaryutil.GetCurrentCanaryStep(deploy)
	canaryCtx.newStatus.Canary.Revision = deploy.Status.Canary.Revision
	if deploy.Spec.UpdateStrategy.CanaryStrategy != nil {
		canaryCtx.newStatus.CurrentStepHash = canaryutil.ComputeStepHash(deploy)
	}
	var stepCount int32
	if deploy.Spec.UpdateStrategy.CanaryStrategy != nil {
		stepCount = int32(len(deploy.Spec.UpdateStrategy.CanaryStrategy.Steps))
	}

	// if canary step hash changes, reset current step index
	if canaryutil.CheckStepHashChange(deploy) {
		canaryCtx.newStatus.CurrentStepIndex = canaryutil.ResetCurrentStepIndex(deploy)
		if deploy.Status.Canary.Revision == canaryCtx.newStatus.UpdateRevision {
			if canaryCtx.newStatus.CurrentStepIndex != nil {
				klog.Info("Skipping all steps because already been the update revision")
				canaryCtx.newStatus.CurrentStepIndex = pointer.Int32Ptr(stepCount)
				r.recorder.Eventf(deploy, corev1.EventTypeNormal, "SetStepIndex", "Set Step Index to %d", int(stepCount))
			}
		}
		return r.updateStatus(deploy, canaryCtx.newStatus, pointer.BoolPtr(false))
	}

	// if pod template change, reset current step index
	if canaryutil.CheckRevisionChange(deploy, canaryCtx.newStatus.UpdateRevision) {
		canaryCtx.newStatus.CurrentStepIndex = canaryutil.ResetCurrentStepIndex(deploy)
		if deploy.Status.Canary.Revision == canaryCtx.newStatus.UpdateRevision {
			if canaryCtx.newStatus.CurrentStepIndex != nil {
				klog.Info("Skipping all steps because already been the update revision")
				canaryCtx.newStatus.CurrentStepIndex = pointer.Int32Ptr(stepCount)
				r.recorder.Eventf(deploy, corev1.EventTypeNormal, "SetStepIndex", "Set Step Index to %d", int(stepCount))
			}
		}
		return r.updateStatus(deploy, canaryCtx.newStatus, pointer.BoolPtr(false))
	}

	if deploy.Status.Canary.Revision == "" {
		if deploy.Spec.UpdateStrategy.CanaryStrategy == nil {
			return r.updateStatus(deploy, canaryCtx.newStatus, pointer.BoolPtr(false))
		}
		canaryCtx.newStatus.Canary.Revision = canaryCtx.newStatus.UpdateRevision
		if stepCount > 0 {
			if stepCount != *currentStepIndex {
				r.recorder.Eventf(deploy, corev1.EventTypeNormal, "SetStepIndex", "Set Step Index to %d", int(stepCount))
			}
			canaryCtx.newStatus.CurrentStepIndex = &stepCount
		}
		return r.updateStatus(deploy, canaryCtx.newStatus, pointer.BoolPtr(false))
	}

	if stepCount == 0 {
		klog.Info("GameDeployment has no steps")
		canaryCtx.newStatus.Canary.Revision = canaryCtx.newStatus.UpdateRevision
		return r.updateStatus(deploy, canaryCtx.newStatus, pointer.BoolPtr(false))
	}

	if *currentStepIndex == stepCount {
		klog.Info("GameDeployment has executed every step")
		canaryCtx.newStatus.CurrentStepIndex = &stepCount
		canaryCtx.newStatus.Canary.Revision = canaryCtx.newStatus.UpdateRevision
		return r.updateStatus(deploy, canaryCtx.newStatus, pointer.BoolPtr(false))
	}

	if completeCurrentCanaryStep(deploy, canaryCtx) {
		*currentStepIndex++
		canaryCtx.newStatus.CurrentStepIndex = currentStepIndex
		canaryCtx.newStatus.Canary.CurrentStepHookRun = ""
		if int(*currentStepIndex) == len(deploy.Spec.UpdateStrategy.CanaryStrategy.Steps) {
			canaryCtx.newStatus.Canary.Revision = canaryCtx.newStatus.UpdateRevision
		}
		klog.Infof("Incrementing the Current Step Index to %d", *currentStepIndex)
		r.recorder.Eventf(deploy, corev1.EventTypeNormal, "SetStepIndex", "Set Step Index to %d", int(*currentStepIndex))
		return r.updateStatus(deploy, canaryCtx.newStatus, pointer.BoolPtr(false))
	}

	paused := deploy.Spec.UpdateStrategy.Paused

	pauseCondition := canaryutil.GetPauseCondition(deploy, hookv1alpha1.PauseReasonCanaryPauseStep)
	if currentStep != nil && currentStep.Pause != nil {
		currentPartition := canaryutil.GetCurrentPartition(deploy)
		if pauseCondition == nil && canaryCtx.newStatus.UpdatedReadyReplicas == *deploy.Spec.Replicas-currentPartition &&
			canaryCtx.newStatus.AvailableReplicas == canaryCtx.newStatus.ReadyReplicas {
			canaryCtx.AddPauseCondition(hookv1alpha1.PauseReasonCanaryPauseStep)
		}
	}

	canaryCtx.newStatus.CurrentStepIndex = currentStepIndex
	paused = r.calculateConditionStatus(deploy, canaryCtx)

	return r.updateStatus(deploy, canaryCtx.newStatus, &paused)
}

// updateStatus update status and updateStrategy pause of a GameDeployment to k8s
func (r *realGameDeploymentStatusUpdater) updateStatus(deploy *gdv1alpha1.GameDeployment, newStatus *gdv1alpha1.GameDeploymentStatus, newPause *bool) error {
	specCopy := deploy.Spec.DeepCopy()
	paused := specCopy.UpdateStrategy.Paused
	if newPause != nil {
		paused = *newPause
	}

	specPatch, specModified, err := commondiffutil.CreateTwoWayMergePatch(
		&gdv1alpha1.GameDeployment{
			Spec: gdv1alpha1.GameDeploymentSpec{
				UpdateStrategy: gdv1alpha1.GameDeploymentUpdateStrategy{
					Paused: deploy.Spec.UpdateStrategy.Paused,
				},
				PreDeleteUpdateStrategy: gdv1alpha1.GameDeploymentPreDeleteUpdateStrategy{
					RetryUnexpectedHooks: deploy.Spec.PreDeleteUpdateStrategy.RetryUnexpectedHooks,
				},
				PreInplaceUpdateStrategy: gdv1alpha1.GameDeploymentPreInplaceUpdateStrategy{
					RetryUnexpectedHooks: deploy.Spec.PreInplaceUpdateStrategy.RetryUnexpectedHooks,
				},
				PostInplaceUpdateStrategy: gdv1alpha1.GameDeploymentPostInplaceUpdateStrategy{
					RetryUnexpectedHooks: deploy.Spec.PostInplaceUpdateStrategy.RetryUnexpectedHooks,
				},
			},
		},
		&gdv1alpha1.GameDeployment{
			Spec: gdv1alpha1.GameDeploymentSpec{
				UpdateStrategy: gdv1alpha1.GameDeploymentUpdateStrategy{
					Paused: paused,
				},
				PreDeleteUpdateStrategy: gdv1alpha1.GameDeploymentPreDeleteUpdateStrategy{
					RetryUnexpectedHooks: false,
				},
				PreInplaceUpdateStrategy: gdv1alpha1.GameDeploymentPreInplaceUpdateStrategy{
					RetryUnexpectedHooks: false,
				},
				PostInplaceUpdateStrategy: gdv1alpha1.GameDeploymentPostInplaceUpdateStrategy{
					RetryUnexpectedHooks: false,
				},
			},
		}, gdv1alpha1.GameDeployment{})
	if err != nil {
		klog.Errorf("Error constructing app status patch: %v", err)
		return err
	}
	if specModified {
		klog.Infof("Rollout Spec Patch: %s", specPatch)
		_, err = r.gdClient.TkexV1alpha1().GameDeployments(deploy.Namespace).Patch(context.TODO(),
			deploy.Name, patchtypes.MergePatchType, specPatch, metav1.PatchOptions{})
		if err != nil {
			klog.Warningf("Error updating GameDeployment Spec: %v", err)
			return err
		}
		klog.Info("Patch spec successfully")
	}

	statusPatch, statusModified, err := commondiffutil.CreateTwoWayMergePatch(
		&gdv1alpha1.GameDeployment{
			Status: deploy.Status,
		},
		&gdv1alpha1.GameDeployment{
			Status: *newStatus,
		}, gdv1alpha1.GameDeployment{})
	if err != nil {
		klog.Errorf("Error constructing app status patch: %v", err)
		return err
	}
	if !statusModified {
		klog.Info("No status changes. Skipping patch")
		return nil
	}
	klog.Infof("Rollout Patch: %s", statusPatch)
	_, err = r.gdClient.TkexV1alpha1().GameDeployments(deploy.Namespace).Patch(context.TODO(),
		deploy.Name, patchtypes.MergePatchType, statusPatch, metav1.PatchOptions{}, "status")
	if err != nil {
		klog.Warningf("Error updating application: %v", err)
		return err
	}
	klog.Info("Patch status successfully")
	r.metrics.CollectRelatedReplicas(util.GetControllerKey(deploy), *deploy.Spec.Replicas, newStatus.ReadyReplicas,
		newStatus.AvailableReplicas, newStatus.UpdatedReplicas, newStatus.UpdatedReadyReplicas)

	return nil
}

func (r *realGameDeploymentStatusUpdater) inconsistentStatus(deploy *gdv1alpha1.GameDeployment, newStatus *gdv1alpha1.GameDeploymentStatus) bool {
	oldStatus := deploy.Status
	return newStatus.ObservedGeneration > oldStatus.ObservedGeneration ||
		newStatus.Replicas != oldStatus.Replicas ||
		newStatus.ReadyReplicas != oldStatus.ReadyReplicas ||
		newStatus.AvailableReplicas != oldStatus.AvailableReplicas ||
		newStatus.UpdatedReadyReplicas != oldStatus.UpdatedReadyReplicas ||
		newStatus.UpdatedReplicas != oldStatus.UpdatedReplicas ||
		newStatus.UpdateRevision != oldStatus.UpdateRevision ||
		newStatus.LabelSelector != oldStatus.LabelSelector
}

// calculateBaseStatus calculate a base status of GameDeployment
func (r *realGameDeploymentStatusUpdater) calculateBaseStatus(deploy *gdv1alpha1.GameDeployment, canaryCtx *canaryContext, pods []*v1.Pod) {
	for _, pod := range pods {
		canaryCtx.newStatus.Replicas++
		if util.IsRunningAndReady(pod) {
			canaryCtx.newStatus.ReadyReplicas++
		}
		if util.IsRunningAndAvailable(pod, deploy.Spec.MinReadySeconds) {
			canaryCtx.newStatus.AvailableReplicas++
		}
		if util.GetPodRevision(pod) == canaryCtx.newStatus.UpdateRevision {
			canaryCtx.newStatus.UpdatedReplicas++
		}
		if util.IsRunningAndReady(pod) && util.GetPodRevision(pod) == canaryCtx.newStatus.UpdateRevision {
			canaryCtx.newStatus.UpdatedReadyReplicas++
		}
	}
}

// calculateConditionStatus calculate condition of GameDeployment, return true if exist pause condition
func (r *realGameDeploymentStatusUpdater) calculateConditionStatus(deploy *gdv1alpha1.GameDeployment, canaryCtx *canaryContext) bool {
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

// completeCurrentCanaryStep checks whether have already complete current canary step
func completeCurrentCanaryStep(deploy *gdv1alpha1.GameDeployment, canaryCtx *canaryContext) bool {
	currentStep, _ := canaryutil.GetCurrentCanaryStep(deploy)
	if currentStep == nil {
		return false
	}

	pauseCondition := canaryutil.GetPauseCondition(deploy, hookv1alpha1.PauseReasonCanaryPauseStep)

	if currentStep.Pause != nil && currentStep.Pause.Duration != nil {
		now := metav1.Now()
		if pauseCondition != nil {
			expiredTime := pauseCondition.StartTime.Add(time.Duration(*currentStep.Pause.Duration) * time.Second)
			if now.After(expiredTime) {
				klog.Info("GameDeployment has waited the duration of the pause step")
				return true
			}
		}
	}

	if currentStep.Pause != nil && currentStep.Pause.Duration == nil && pauseCondition != nil && !deploy.Spec.UpdateStrategy.Paused {
		klog.Info("GameDeployment has been unpaused")
		return true
	}

	if currentStep.Partition != nil && canaryCtx.newStatus.UpdatedReadyReplicas == *deploy.Spec.Replicas-*currentStep.Partition && canaryCtx.newStatus.AvailableReplicas == canaryCtx.newStatus.ReadyReplicas {
		klog.Info("GameDeployment has reached the desired state for the correct partition")
		return true
	}

	currentHrs := canaryCtx.CurrentHookRuns()
	currentStepHr := commonhookutil.GetCurrentStepHookRun(currentHrs)
	hrExistsAndCompleted := currentStepHr != nil && currentStepHr.Status.Phase.Completed()
	if currentStep.Hook != nil && hrExistsAndCompleted && currentStepHr.Status.Phase == hookv1alpha1.HookPhaseSuccessful {
		return true
	}

	pauseConditionByHook := canaryutil.GetPauseCondition(deploy, hookv1alpha1.PauseReasonStepBasedHook)
	if currentStep.Hook != nil && pauseConditionByHook != nil && !deploy.Spec.UpdateStrategy.Paused {
		klog.Info("GameDeployment has been unpaused")
		return true
	}

	return false
}
