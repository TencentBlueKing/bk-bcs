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

package controllers

import (
	"encoding/json"
	"time"

	tkexv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	tkexclientset "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/client/clientset/versioned"
	gamedeploylister "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/client/listers/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/util"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	patchtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
	"k8s.io/utils/pointer"
)

// GameDeploymentStatusUpdaterInterface is an interface used to update the GameDeploymentStatus associated with a GameDeployment.
// For any use other than testing, clients should create an instance using NewRealGameDeploymentStatusUpdater.
type GameDeploymentStatusUpdaterInterface interface {
	// UpdateGameDeploymentStatus sets the set's Status to status. Implementations are required to retry on conflicts,
	// but fail on other errors. If the returned error is nil set's Status has been successfully set to status.
	UpdateGameDeploymentStatus(deploy *tkexv1alpha1.GameDeployment, newStatus *tkexv1alpha1.GameDeploymentStatus, pods []*v1.Pod) error
}

// NewRealGameDeploymentStatusUpdater returns a GameDeploymentStatusUpdaterInterface that updates the Status of a GameDeployment,
// using the supplied client and setLister.
func NewRealGameDeploymentStatusUpdater(
	tkexClient tkexclientset.Interface,
	setLister gamedeploylister.GameDeploymentLister,
	record record.EventRecorder) GameDeploymentStatusUpdaterInterface {
	return &realGameDeploymentStatusUpdater{tkexClient, setLister, record}
}

type realGameDeploymentStatusUpdater struct {
	tkexClient tkexclientset.Interface
	setLister  gamedeploylister.GameDeploymentLister
	recorder   record.EventRecorder
}

func (r *realGameDeploymentStatusUpdater) UpdateGameDeploymentStatus(
	deploy *tkexv1alpha1.GameDeployment,
	newStatus *tkexv1alpha1.GameDeploymentStatus,
	pods []*v1.Pod) error {
	r.calculateStatus(deploy, newStatus, pods)
	currentStep, currentStepIndex := util.GetCurrentCanaryStep(deploy)
	newStatus.Canary.Revision = deploy.Status.Canary.Revision
	if deploy.Spec.UpdateStrategy.CanaryStrategy != nil {
		newStatus.CurrentStepHash = util.ComputeStepHash(deploy)
	}
	var stepCount int32
	if deploy.Spec.UpdateStrategy.CanaryStrategy != nil {
		stepCount = int32(len(deploy.Spec.UpdateStrategy.CanaryStrategy.Steps))
	}

	if util.CheckStepHashChange(deploy) {
		newStatus.CurrentStepIndex = util.ResetCurrentStepIndex(deploy)
		if deploy.Status.Canary.Revision == newStatus.UpdateRevision {
			if newStatus.CurrentStepIndex != nil {
				klog.Info("Skipping all steps because already been the update revision")
				newStatus.CurrentStepIndex = pointer.Int32Ptr(stepCount)
				r.recorder.Eventf(deploy, corev1.EventTypeNormal, "SetStepIndex", "Set Step Index to %d", int(stepCount))
			}
		}
		return r.updateStatus(deploy, newStatus, pointer.BoolPtr(false))
	}

	if util.CheckRevisionChange(deploy, newStatus.UpdateRevision) {
		newStatus.CurrentStepIndex = util.ResetCurrentStepIndex(deploy)
		if deploy.Status.Canary.Revision == newStatus.UpdateRevision {
			if newStatus.CurrentStepIndex != nil {
				klog.Info("Skipping all steps because already been the update revision")
				newStatus.CurrentStepIndex = pointer.Int32Ptr(stepCount)
				r.recorder.Eventf(deploy, corev1.EventTypeNormal, "SetStepIndex", "Set Step Index to %d", int(stepCount))
			}
		}
		return r.updateStatus(deploy, newStatus, pointer.BoolPtr(false))
	}

	if deploy.Status.Canary.Revision == "" {
		if deploy.Spec.UpdateStrategy.CanaryStrategy == nil {
			return r.updateStatus(deploy, newStatus, pointer.BoolPtr(false))
		}
		newStatus.Canary.Revision = newStatus.UpdateRevision
		if stepCount > 0 {
			if stepCount != *currentStepIndex {
				r.recorder.Eventf(deploy, corev1.EventTypeNormal, "SetStepIndex", "Set Step Index to %d", int(stepCount))
			}
			newStatus.CurrentStepIndex = &stepCount
		}
		return r.updateStatus(deploy, newStatus, pointer.BoolPtr(false))
	}

	if stepCount == 0 {
		klog.Info("GameDeployment has no steps")
		newStatus.Canary.Revision = newStatus.UpdateRevision
		return r.updateStatus(deploy, newStatus, pointer.BoolPtr(false))
	}

	if *currentStepIndex == stepCount {
		klog.Info("GameDeployment has executed every step")
		newStatus.CurrentStepIndex = &stepCount
		newStatus.Canary.Revision = newStatus.UpdateRevision
		return r.updateStatus(deploy, newStatus, pointer.BoolPtr(false))
	}

	if completeCurrentCanaryStep(deploy, newStatus) {
		*currentStepIndex++
		newStatus.CurrentStepIndex = currentStepIndex
		if int(*currentStepIndex) == len(deploy.Spec.UpdateStrategy.CanaryStrategy.Steps) {
			newStatus.Canary.Revision = newStatus.UpdateRevision
		}
		klog.Infof("Incrementing the Current Step Index to %d", *currentStepIndex)
		r.recorder.Eventf(deploy, corev1.EventTypeNormal, "SetStepIndex", "Set Step Index to %d", int(*currentStepIndex))
		return r.updateStatus(deploy, newStatus, pointer.BoolPtr(false))
	}

	pauseStartTime := deploy.Status.Canary.PauseStartTime
	paused := deploy.Spec.UpdateStrategy.Paused
	if currentStep != nil && currentStep.Pause != nil {
		now := metav1.Now()
		currentPartition := util.GetCurrentPartition(deploy)
		if deploy.Status.Canary.PauseStartTime == nil && newStatus.UpdatedReadyReplicas == *deploy.Spec.Replicas-currentPartition &&
			newStatus.AvailableReplicas == newStatus.ReadyReplicas {
			klog.Infof("Setting PauseStartTime to %s", now.UTC().Format(time.RFC3339))
			pauseStartTime = &now
			paused = true
		}
	}

	newStatus.Canary.PauseStartTime = pauseStartTime
	newStatus.CurrentStepIndex = currentStepIndex
	return r.updateStatus(deploy, newStatus, &paused)
}

func (r *realGameDeploymentStatusUpdater) updateStatus(deploy *tkexv1alpha1.GameDeployment, newStatus *tkexv1alpha1.GameDeploymentStatus, newPause *bool) error {
	specCopy := deploy.Spec.DeepCopy()
	paused := specCopy.UpdateStrategy.Paused
	if newPause != nil {
		paused = *newPause
	}

	specPatch, specModified, err := CreateTwoWayMergePatch(
		&tkexv1alpha1.GameDeployment{
			Spec: tkexv1alpha1.GameDeploymentSpec{
				UpdateStrategy: tkexv1alpha1.GameDeploymentUpdateStrategy{
					Paused: deploy.Spec.UpdateStrategy.Paused,
				},
			},
		},
		&tkexv1alpha1.GameDeployment{
			Spec: tkexv1alpha1.GameDeploymentSpec{
				UpdateStrategy: tkexv1alpha1.GameDeploymentUpdateStrategy{
					Paused: paused,
				},
			},
		}, tkexv1alpha1.GameDeployment{})
	if err != nil {
		klog.Errorf("Error constructing app status patch: %v", err)
		return err
	}
	if specModified {
		klog.Infof("Rollout Spec Patch: %s", specPatch)
		_, err = r.tkexClient.TkexV1alpha1().GameDeployments(deploy.Namespace).Patch(deploy.Name, patchtypes.MergePatchType, specPatch)
		if err != nil {
			klog.Warningf("Error updating GameDeployment Spec: %v", err)
			return err
		}
		klog.Info("Patch spec successfully")
	}

	statusPatch, statusModified, err := CreateTwoWayMergePatch(
		&tkexv1alpha1.GameDeployment{
			Status: deploy.Status,
		},
		&tkexv1alpha1.GameDeployment{
			Status: *newStatus,
		}, tkexv1alpha1.GameDeployment{})
	if err != nil {
		klog.Errorf("Error constructing app status patch: %v", err)
		return err
	}
	if !statusModified {
		klog.Info("No status changes. Skipping patch")
		return nil
	}
	klog.Infof("Rollout Patch: %s", statusPatch)
	_, err = r.tkexClient.TkexV1alpha1().GameDeployments(deploy.Namespace).Patch(deploy.Name, patchtypes.MergePatchType, statusPatch, "status")
	if err != nil {
		klog.Warningf("Error updating application: %v", err)
		return err
	}
	klog.Info("Patch status successfully")
	return nil
}

func CreateTwoWayMergePatch(orig, new, dataStruct interface{}) ([]byte, bool, error) {
	origBytes, err := json.Marshal(orig)
	if err != nil {
		return nil, false, err
	}
	newBytes, err := json.Marshal(new)
	if err != nil {
		return nil, false, err
	}
	patch, err := strategicpatch.CreateTwoWayMergePatch(origBytes, newBytes, dataStruct)
	if err != nil {
		return nil, false, err
	}
	return patch, string(patch) != "{}", nil
}

func (r *realGameDeploymentStatusUpdater) inconsistentStatus(deploy *tkexv1alpha1.GameDeployment, newStatus *tkexv1alpha1.GameDeploymentStatus) bool {
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

func (r *realGameDeploymentStatusUpdater) calculateStatus(deploy *tkexv1alpha1.GameDeployment, newStatus *tkexv1alpha1.GameDeploymentStatus, pods []*v1.Pod) {
	for _, pod := range pods {
		newStatus.Replicas++
		if util.IsRunningAndReady(pod) {
			newStatus.ReadyReplicas++
		}
		if util.IsRunningAndAvailable(pod, deploy.Spec.MinReadySeconds) {
			newStatus.AvailableReplicas++
		}
		if util.GetPodRevision(pod) == newStatus.UpdateRevision {
			newStatus.UpdatedReplicas++
		}
		if util.IsRunningAndReady(pod) && util.GetPodRevision(pod) == newStatus.UpdateRevision {
			newStatus.UpdatedReadyReplicas++
		}
	}
}

func completeCurrentCanaryStep(deploy *tkexv1alpha1.GameDeployment, newStatus *tkexv1alpha1.GameDeploymentStatus) bool {
	currentStep, _ := util.GetCurrentCanaryStep(deploy)
	if currentStep == nil {
		return false
	}

	if currentStep.Pause != nil && currentStep.Pause.Duration != nil {
		now := metav1.Now()
		if deploy.Status.Canary.PauseStartTime != nil {
			expiredTime := deploy.Status.Canary.PauseStartTime.Add(time.Duration(*currentStep.Pause.Duration) * time.Second)
			if now.After(expiredTime) {
				klog.Info("GameDeployment has waited the duration of the pause step")
				return true
			}
		}
	}

	if currentStep.Pause != nil && currentStep.Pause.Duration == nil && deploy.Status.Canary.PauseStartTime != nil && !deploy.Spec.UpdateStrategy.Paused {
		klog.Info("GameDeployment has been unpaused")
		return true
	}

	if currentStep.Partition != nil && newStatus.UpdatedReadyReplicas == *deploy.Spec.Replicas-*currentStep.Partition && newStatus.AvailableReplicas == newStatus.ReadyReplicas {
		klog.Info("GameDeployment has reached the desired state for the correct partition")
		return true
	}

	return false
}
