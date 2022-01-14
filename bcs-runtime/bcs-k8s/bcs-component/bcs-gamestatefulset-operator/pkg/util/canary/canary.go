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
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	"hash/fnv"
	"time"

	gstsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/klog"
	hashutil "k8s.io/kubernetes/pkg/util/hash"
	"k8s.io/utils/pointer"
)

// GetCurrentCanaryStep get current canary step
func GetCurrentCanaryStep(set *gstsv1alpha1.GameStatefulSet) (*gstsv1alpha1.CanaryStep, *int32) {
	if set.Spec.UpdateStrategy.CanaryStrategy == nil || len(set.Spec.UpdateStrategy.CanaryStrategy.Steps) == 0 {
		return nil, nil
	}

	currentStepIndex := int32(0)
	if set.Status.CurrentStepIndex != nil {
		currentStepIndex = *set.Status.CurrentStepIndex
	}
	if len(set.Spec.UpdateStrategy.CanaryStrategy.Steps) <= int(currentStepIndex) {
		return nil, &currentStepIndex
	}
	return &set.Spec.UpdateStrategy.CanaryStrategy.Steps[currentStepIndex], &currentStepIndex
}

// GetCurrentPartition get current partition of canary
func GetCurrentPartition(set *gstsv1alpha1.GameStatefulSet) (int, error) {
	currentStep, currentStepIndex := GetCurrentCanaryStep(set)
	if currentStep == nil {
		if (set.Spec.UpdateStrategy.CanaryStrategy == nil || len(set.Spec.UpdateStrategy.CanaryStrategy.Steps) == 0) &&
			set.Spec.UpdateStrategy.RollingUpdate != nil && set.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
			return intstrutil.GetValueFromIntOrPercent(set.Spec.UpdateStrategy.RollingUpdate.Partition,
				int(*set.Spec.Replicas), false)
		}
		return 0, nil
	}

	for i := *currentStepIndex; i >= 0; i-- {
		step := set.Spec.UpdateStrategy.CanaryStrategy.Steps[i]
		if step.Partition != nil {
			return intstrutil.GetValueFromIntOrPercent(step.Partition, int(*set.Spec.Replicas), false)
		}
	}
	return int(*set.Spec.Replicas), nil
}

// CheckStepHashChange detects if there is an change in the canary steps
func CheckStepHashChange(set *gstsv1alpha1.GameStatefulSet) bool {
	if set.Status.CurrentStepHash == "" {
		return false
	}
	return set.Status.CurrentStepHash != ComputeStepHash(set)
}

// CheckRevisionChange detects if there is an change in the pod template
func CheckRevisionChange(set *gstsv1alpha1.GameStatefulSet, revision string) bool {
	if set.Status.UpdateRevision == "" {
		return false
	}
	return set.Status.UpdateRevision != revision
}

// ComputeStepHash generates a hash with GameStatefulSet canary steps
func ComputeStepHash(set *gstsv1alpha1.GameStatefulSet) string {
	deployStepHasher := fnv.New32a()
	if set.Spec.UpdateStrategy.CanaryStrategy != nil {
		hashutil.DeepHashObject(deployStepHasher, set.Spec.UpdateStrategy.CanaryStrategy.Steps)
	}
	return rand.SafeEncodeString(fmt.Sprint(deployStepHasher.Sum32()))
}

// ResetCurrentStepIndex resets the canary step
func ResetCurrentStepIndex(set *gstsv1alpha1.GameStatefulSet) *int32 {
	if set.Spec.UpdateStrategy.CanaryStrategy != nil && len(set.Spec.UpdateStrategy.CanaryStrategy.Steps) > 0 {
		return pointer.Int32Ptr(0)
	}
	return nil
}

// GetPauseCondition get pause condition with a pause reason
func GetPauseCondition(set *gstsv1alpha1.GameStatefulSet, reason hookv1alpha1.PauseReason) *hookv1alpha1.PauseCondition {
	for i := range set.Status.PauseConditions {
		cond := set.Status.PauseConditions[i]
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
