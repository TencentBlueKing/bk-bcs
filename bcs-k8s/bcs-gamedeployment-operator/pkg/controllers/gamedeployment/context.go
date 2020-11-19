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
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	hooksutil "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/util/hook"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type canaryContext struct {
	deploy       *v1alpha1.GameDeployment
	newStatus    *v1alpha1.GameDeploymentStatus
	currentHrs   []*v1alpha1.HookRun
	otherHrs     []*v1alpha1.HookRun
	pauseReasons []v1alpha1.PauseReason
	pause        bool
}

func newCanaryCtx(deploy *v1alpha1.GameDeployment, hrList []*v1alpha1.HookRun, updateRevision *apps.ControllerRevision,
	collisionCount int32, selector labels.Selector) *canaryContext {

	currentHrs, otherHrs := hooksutil.FilterCurrentHookRuns(hrList, deploy)
	newStatus := v1alpha1.GameDeploymentStatus{
		ObservedGeneration: deploy.Generation,
		UpdateRevision:     updateRevision.Name,
		CollisionCount:     new(int32),
		LabelSelector:      selector.String(),
	}
	*newStatus.CollisionCount = collisionCount

	return &canaryContext{
		deploy:     deploy,
		newStatus:  &newStatus,
		currentHrs: currentHrs,
		otherHrs:   otherHrs,
	}
}

func (cCtx *canaryContext) CurrentHookRuns() []*v1alpha1.HookRun {
	return cCtx.currentHrs
}

func (cCtx *canaryContext) OtherHookRuns() []*v1alpha1.HookRun {
	return cCtx.otherHrs
}

func (cCtx *canaryContext) SetCurrentHookRuns(ars []*v1alpha1.HookRun) {
	cCtx.currentHrs = ars
	currStepAr := hooksutil.GetCurrentStepHookRun(ars)
	if currStepAr != nil {
		cCtx.newStatus.Canary.CurrentStepHookRun = currStepAr.Name
	}

}

func (cCtx *canaryContext) GetPauseCondition(reason v1alpha1.PauseReason) *v1alpha1.PauseCondition {
	for _, cond := range cCtx.deploy.Status.PauseConditions {
		if cond.Reason == reason {
			return &cond
		}
	}
	return nil
}

func (cCtx *canaryContext) AddPauseCondition(reason v1alpha1.PauseReason) {
	cCtx.pauseReasons = append(cCtx.pauseReasons, reason)
}

func (cCtx *canaryContext) HasAddPause() bool {
	return len(cCtx.pauseReasons) > 0
}
