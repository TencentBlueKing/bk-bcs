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

package canary

import (
	"fmt"
	"hash/fnv"
	"time"

	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"

	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog"
	hashutil "k8s.io/kubernetes/pkg/util/hash"
	"k8s.io/utils/pointer"
)

// GetCurrentCanaryStep get current canary step
func GetCurrentCanaryStep(deploy *gdv1alpha1.GameDeployment) (*gdv1alpha1.CanaryStep, *int32) {
	if deploy.Spec.UpdateStrategy.CanaryStrategy == nil || len(deploy.Spec.UpdateStrategy.CanaryStrategy.Steps) == 0 {
		return nil, nil
	}

	currentStepIndex := int32(0)
	if deploy.Status.CurrentStepIndex != nil {
		currentStepIndex = *deploy.Status.CurrentStepIndex
	}
	if len(deploy.Spec.UpdateStrategy.CanaryStrategy.Steps) <= int(currentStepIndex) {
		return nil, &currentStepIndex
	}
	return &deploy.Spec.UpdateStrategy.CanaryStrategy.Steps[currentStepIndex], &currentStepIndex
}

// GetCurrentPartition get current partition of canary
func GetCurrentPartition(deploy *gdv1alpha1.GameDeployment) int32 {
	currentStep, currentStepIndex := GetCurrentCanaryStep(deploy)
	if currentStep == nil {
		if (deploy.Spec.UpdateStrategy.CanaryStrategy == nil || len(deploy.Spec.UpdateStrategy.CanaryStrategy.Steps) == 0) &&
			deploy.Spec.UpdateStrategy.Partition != nil {
			return *deploy.Spec.UpdateStrategy.Partition
		}
		return 0
	}

	for i := *currentStepIndex; i >= 0; i-- {
		step := deploy.Spec.UpdateStrategy.CanaryStrategy.Steps[i]
		if step.Partition != nil {
			return *step.Partition
		}
	}
	return *deploy.Spec.Replicas
}

// CheckStepHashChange detects if there is an change in the canary steps
func CheckStepHashChange(deploy *gdv1alpha1.GameDeployment) bool {
	if deploy.Status.CurrentStepHash == "" {
		return false
	}
	return deploy.Status.CurrentStepHash != ComputeStepHash(deploy)
}

// CheckRevisionChange detects if there is an change in the pod template
func CheckRevisionChange(deploy *gdv1alpha1.GameDeployment, revision string) bool {
	if deploy.Status.UpdateRevision == "" {
		return false
	}
	return deploy.Status.UpdateRevision != revision
}

// ComputeStepHash generates a hash with GameDeployment canary steps
func ComputeStepHash(deploy *gdv1alpha1.GameDeployment) string {
	deployStepHasher := fnv.New32a()
	if deploy.Spec.UpdateStrategy.CanaryStrategy != nil {
		hashutil.DeepHashObject(deployStepHasher, deploy.Spec.UpdateStrategy.CanaryStrategy.Steps)
	}
	return rand.SafeEncodeString(fmt.Sprint(deployStepHasher.Sum32()))
}

// ResetCurrentStepIndex resets the canary step
func ResetCurrentStepIndex(deploy *gdv1alpha1.GameDeployment) *int32 {
	if deploy.Spec.UpdateStrategy.CanaryStrategy != nil && len(deploy.Spec.UpdateStrategy.CanaryStrategy.Steps) > 0 {
		return pointer.Int32Ptr(0)
	}
	return nil
}

// GetPauseCondition get pause condition with a pause reason
func GetPauseCondition(deploy *gdv1alpha1.GameDeployment, reason hookv1alpha1.PauseReason) *hookv1alpha1.PauseCondition {
	for i := range deploy.Status.PauseConditions {
		cond := deploy.Status.PauseConditions[i]
		if cond.Reason == reason {
			return &cond
		}
	}
	return nil
}

// GetMinDuration get min duration from two durations
func GetMinDuration(duration1, duration2 time.Duration) time.Duration {
	if duration1 < 0 || duration2 < 0 {
		klog.Warning("Invalid requeue duration, a requeue duration must greater than 0")
	}
	if duration1 > 0 && duration2 == 0 {
		return duration1
	}
	if duration1 == 0 && duration2 > 0 {
		return duration2
	}

	requeueDuration := duration1
	if duration2 < requeueDuration {
		requeueDuration = duration2
	}
	return requeueDuration
}
