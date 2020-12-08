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
	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/util/hook"
)

func FilterCurrentHookRuns(hookRuns []*hookv1alpha1.HookRun, deploy *gdv1alpha1.GameDeployment) ([]*hookv1alpha1.HookRun, []*hookv1alpha1.HookRun) {
	return commonhookutil.FilterHookRuns(hookRuns, func(hr *hookv1alpha1.HookRun) bool {
		if hr.Name == deploy.Status.Canary.CurrentStepHookRun {
			return true
		}
		return false
	})
}
