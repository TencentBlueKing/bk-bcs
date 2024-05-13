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
 */

// Package hook xxx
package hook

import (
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
)

// FilterHookRuns xxx
func FilterHookRuns(hrs []*hookv1alpha1.HookRun, cond func(hr *hookv1alpha1.HookRun) bool) (
	[]*hookv1alpha1.HookRun, []*hookv1alpha1.HookRun) {
	condTrue := []*hookv1alpha1.HookRun{}
	condFalse := []*hookv1alpha1.HookRun{}
	for _, hr := range hrs {
		if hr == nil {
			continue
		}
		if cond(hr) {
			condTrue = append(condTrue, hr)
		} else {
			condFalse = append(condFalse, hr)
		}
	}
	return condTrue, condFalse
}

// FilterHookRunsByName xxx
func FilterHookRunsByName(hookRuns []*hookv1alpha1.HookRun, name string) *hookv1alpha1.HookRun {
	hookRunsByName, _ := FilterHookRuns(hookRuns, func(hr *hookv1alpha1.HookRun) bool {
		return hr.Name == name
	})
	if len(hookRunsByName) == 1 {
		return hookRunsByName[0]
	}
	return nil
}

// GetCurrentStepHookRun xxx
func GetCurrentStepHookRun(currentHrs []*hookv1alpha1.HookRun) *hookv1alpha1.HookRun {
	for _, hr := range currentHrs {
		hookRunType, ok := hr.Labels[HookRunTypeLabel]
		if ok && hookRunType == HookRunTypeCanaryStepLabel {
			return hr
		}
	}
	return nil
}

// FilterHookRunsToDelete xxx
func FilterHookRunsToDelete(hrs []*hookv1alpha1.HookRun, revision string) []*hookv1alpha1.HookRun {
	hrsToDelete := []*hookv1alpha1.HookRun{}
	for _, hr := range hrs {
		if hr.Labels[WorkloadRevisionUniqueLabel] != revision {
			hrsToDelete = append(hrsToDelete, hr)
		}
	}

	return hrsToDelete
}

// FilterPreDeleteHookRuns xxx
func FilterPreDeleteHookRuns(hrs []*hookv1alpha1.HookRun) []*hookv1alpha1.HookRun {
	preDeleteHookRuns := []*hookv1alpha1.HookRun{}
	for _, hr := range hrs {
		hookRunType, ok := hr.Labels[HookRunTypeLabel]
		if ok && hookRunType == HookRunTypePreDeleteLabel {
			preDeleteHookRuns = append(preDeleteHookRuns, hr)
		}
	}

	return preDeleteHookRuns
}

// FilterPreInplaceHookRuns xxx
func FilterPreInplaceHookRuns(hrs []*hookv1alpha1.HookRun) []*hookv1alpha1.HookRun {
	preInplaceHookRuns := []*hookv1alpha1.HookRun{}
	for _, hr := range hrs {
		hookRunType, ok := hr.Labels[HookRunTypeLabel]
		if ok && hookRunType == HookRunTypePreInplaceLabel {
			preInplaceHookRuns = append(preInplaceHookRuns, hr)
		}
	}

	return preInplaceHookRuns
}
