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

package util

import (
	"fmt"
	"hash/fnv"
	"k8s.io/utils/pointer"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	"k8s.io/apimachinery/pkg/util/rand"
	hashutil "k8s.io/kubernetes/pkg/util/hash"
)

func GetCurrentCanaryStep(deploy *v1alpha1.GameDeployment) (*v1alpha1.CanaryStep, *int32) {
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

func GetCurrentPartition(deploy *v1alpha1.GameDeployment) int32 {
	currentStep, currentStepIndex := GetCurrentCanaryStep(deploy)
	if currentStep == nil {
		if deploy.Spec.UpdateStrategy.Partition != nil {
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
	return 0
}

func CheckStepHashChange(deploy *v1alpha1.GameDeployment) bool {
	if deploy.Status.CurrentStepHash == "" {
		return false
	}
	return deploy.Status.CurrentStepHash != ComputeStepHash(deploy)
}

func CheckRevisionChange(deploy *v1alpha1.GameDeployment, revision string) bool {
	if deploy.Status.UpdateRevision == "" {
		return false
	}
	return deploy.Status.UpdateRevision != revision
}

func ComputeStepHash(deploy *v1alpha1.GameDeployment) string {
	deployStepHasher := fnv.New32a()
	if deploy.Spec.UpdateStrategy.CanaryStrategy != nil {
		hashutil.DeepHashObject(deployStepHasher, deploy.Spec.UpdateStrategy.CanaryStrategy.Steps)
	}
	return rand.SafeEncodeString(fmt.Sprint(deployStepHasher.Sum32()))
}

func ResetCurrentStepIndex(deploy *v1alpha1.GameDeployment) *int32 {
	if deploy.Spec.UpdateStrategy.CanaryStrategy != nil && len(deploy.Spec.UpdateStrategy.CanaryStrategy.Steps) > 0 {
		return pointer.Int32Ptr(0)
	}
	return nil
}
