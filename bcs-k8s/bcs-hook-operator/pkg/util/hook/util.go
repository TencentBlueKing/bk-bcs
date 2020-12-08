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

package hook

import (
	"github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
)

// hookStatusOrder is a list of completed hook sorted in best to worst condition
var hookStatusOrder = []v1alpha1.HookPhase{
	v1alpha1.HookPhaseSuccessful,
	v1alpha1.HookPhaseRunning,
	v1alpha1.HookPhasePending,
	v1alpha1.HookPhaseInconclusive,
	v1alpha1.HookPhaseError,
	v1alpha1.HookPhaseFailed,
}

func IsTerminating(run *v1alpha1.HookRun) bool {
	if run.Spec.Terminate {
		return true
	}
	for _, res := range run.Status.MetricResults {
		switch res.Phase {
		case v1alpha1.HookPhaseFailed, v1alpha1.HookPhaseError, v1alpha1.HookPhaseInconclusive:
			return true
		}
	}
	return false
}

func GetResult(run *v1alpha1.HookRun, metricName string) *v1alpha1.MetricResult {
	for _, result := range run.Status.MetricResults {
		if result.Name == metricName {
			return &result
		}
	}
	return nil
}

func SetResult(run *v1alpha1.HookRun, result v1alpha1.MetricResult) {
	for i, r := range run.Status.MetricResults {
		if r.Name == result.Name {
			run.Status.MetricResults[i] = result
			return
		}
	}
	run.Status.MetricResults = append(run.Status.MetricResults, result)
}

func MetricCompleted(run *v1alpha1.HookRun, metricName string) bool {
	if result := GetResult(run, metricName); result != nil {
		return result.Phase.Completed()
	}
	return false
}

func LastMeasurement(run *v1alpha1.HookRun, metricName string) *v1alpha1.Measurement {
	if result := GetResult(run, metricName); result != nil {
		totalMeasurements := len(result.Measurements)
		if totalMeasurements == 0 {
			return nil
		}
		return &result.Measurements[totalMeasurements-1]
	}
	return nil
}

func IsWorse(current, new v1alpha1.HookPhase) bool {
	currentIndex := 0
	newIndex := 0
	for i, code := range hookStatusOrder {
		if current == code {
			currentIndex = i
		}
		if new == code {
			newIndex = i
		}
	}
	return newIndex > currentIndex
}
