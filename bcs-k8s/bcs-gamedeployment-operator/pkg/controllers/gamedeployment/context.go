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
	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	hooksutil "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/util/hook"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/util/hook"

	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type canaryContext struct {
	deploy       *gdv1alpha1.GameDeployment
	newStatus    *gdv1alpha1.GameDeploymentStatus
	currentHrs   []*hookv1alpha1.HookRun
	otherHrs     []*hookv1alpha1.HookRun
	pauseReasons []hookv1alpha1.PauseReason
	pause        bool
}

func newCanaryCtx(deploy *gdv1alpha1.GameDeployment, hrList []*hookv1alpha1.HookRun, updateRevision *apps.ControllerRevision,
	collisionCount int32, selector labels.Selector) *canaryContext {

	currentHrs, otherHrs := hooksutil.FilterCurrentHookRuns(hrList, deploy)
	canaryHrs := []*hookv1alpha1.HookRun{}
	for _, hr := range otherHrs {
		hookRunType, ok := hr.Labels[commonhookutil.HookRunTypeLabel]
		if ok && hookRunType == commonhookutil.HookRunTypeCanaryStepLabel {
			canaryHrs = append(canaryHrs, hr)
		}
	}
	newStatus := gdv1alpha1.GameDeploymentStatus{
		ObservedGeneration: deploy.Generation,
		UpdateRevision:     updateRevision.Name,
		CollisionCount:     new(int32),
		LabelSelector:      selector.String(),
	}
	*newStatus.CollisionCount = collisionCount
	copyStatus := deploy.Status.DeepCopy()
	newStatus.PreDeleteHookConditions = copyStatus.PreDeleteHookConditions

	return &canaryContext{
		deploy:     deploy,
		newStatus:  &newStatus,
		currentHrs: currentHrs,
		otherHrs:   canaryHrs,
	}
}

func (cCtx *canaryContext) CurrentHookRuns() []*hookv1alpha1.HookRun {
	return cCtx.currentHrs
}

func (cCtx *canaryContext) OtherHookRuns() []*hookv1alpha1.HookRun {
	return cCtx.otherHrs
}

func (cCtx *canaryContext) SetCurrentHookRuns(ars []*hookv1alpha1.HookRun) {
	cCtx.currentHrs = ars
	currStepAr := commonhookutil.GetCurrentStepHookRun(ars)
	if currStepAr != nil {
		cCtx.newStatus.Canary.CurrentStepHookRun = currStepAr.Name
	}

}

func (cCtx *canaryContext) GetPauseCondition(reason hookv1alpha1.PauseReason) *hookv1alpha1.PauseCondition {
	for _, cond := range cCtx.deploy.Status.PauseConditions {
		if cond.Reason == reason {
			return &cond
		}
	}
	return nil
}

func (cCtx *canaryContext) AddPauseCondition(reason hookv1alpha1.PauseReason) {
	cCtx.pauseReasons = append(cCtx.pauseReasons, reason)
}

func (cCtx *canaryContext) HasAddPause() bool {
	return len(cCtx.pauseReasons) > 0
}
