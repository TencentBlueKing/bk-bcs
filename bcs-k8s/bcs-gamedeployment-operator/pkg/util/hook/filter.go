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
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
)

func FilterCurrentHookRuns(hookRuns []*v1alpha1.HookRun, deploy *v1alpha1.GameDeployment) ([]*v1alpha1.HookRun, []*v1alpha1.HookRun) {
	return FilterHookRuns(hookRuns, func(hr *v1alpha1.HookRun) bool {
		if hr.Name == deploy.Status.Canary.CurrentStepHookRun {
			return true
		}
		return false
	})
}

func FilterHookRuns(hrs []*v1alpha1.HookRun, cond func(hr *v1alpha1.HookRun) bool) ([]*v1alpha1.HookRun, []*v1alpha1.HookRun) {
	condTrue := []*v1alpha1.HookRun{}
	condFalse := []*v1alpha1.HookRun{}
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

func FilterHookRunsByName(hookRuns []*v1alpha1.HookRun, name string) *v1alpha1.HookRun {
	hookRunsByName, _ := FilterHookRuns(hookRuns, func(hr *v1alpha1.HookRun) bool {
		return hr.Name == name
	})
	if len(hookRunsByName) == 1 {
		return hookRunsByName[0]
	}
	return nil
}

func GetCurrentStepHookRun(currentHrs []*v1alpha1.HookRun) *v1alpha1.HookRun {
	for _, hr := range currentHrs {
		hookRunType, ok := hr.Labels[v1alpha1.GameDeploymentTypeLabel]
		if ok && hookRunType == v1alpha1.GameDeploymentTypeStepLabel {
			return hr
		}
	}
	return nil
}

func FilterHookRunsToDelete(hrs []*v1alpha1.HookRun, revision string) []*v1alpha1.HookRun {
	hrsToDelete := []*v1alpha1.HookRun{}
	for _, hr := range hrs {
		if hr.Labels[v1alpha1.DefaultGameDeploymentUniqueLabelKey] != revision {
			hrsToDelete = append(hrsToDelete, hr)
		}
	}

	return hrsToDelete
}

func FilterPreDeleteHookRuns(hrs []*v1alpha1.HookRun) []*v1alpha1.HookRun {
	preDeleteHookRuns := []*v1alpha1.HookRun{}
	for _, hr := range hrs {
		hookRunType, ok := hr.Labels[v1alpha1.GameDeploymentTypeLabel]
		if ok && hookRunType == v1alpha1.GameDeploymentTypePreDeleteLabel {
			preDeleteHookRuns = append(preDeleteHookRuns, hr)
		}
	}

	return preDeleteHookRuns
}
