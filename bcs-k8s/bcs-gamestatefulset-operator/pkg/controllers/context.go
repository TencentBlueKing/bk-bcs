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

package gamestatefulset

import (
	gstsv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/util"
	hooksutil "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/util/hook"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/util/hook"
	apps "k8s.io/api/apps/v1"
)

type canaryContext struct {
	set          *gstsv1alpha1.GameStatefulSet
	newStatus    *gstsv1alpha1.GameStatefulSetStatus
	currentHrs   []*hookv1alpha1.HookRun
	otherHrs     []*hookv1alpha1.HookRun
	pauseReasons []hookv1alpha1.PauseReason
}

func newCanaryCtx(set *gstsv1alpha1.GameStatefulSet, hrList []*hookv1alpha1.HookRun, currentRevision,
	updateRevision *apps.ControllerRevision, collisionCount int32) *canaryContext {

	currentHrs, otherHrs := hooksutil.FilterCurrentHookRuns(hrList, set)
	canaryHrs := []*hookv1alpha1.HookRun{}
	for _, hr := range otherHrs {
		hookRunType, ok := hr.Labels[commonhookutil.HookRunTypeLabel]
		if ok && hookRunType == commonhookutil.HookRunTypeCanaryStepLabel {
			canaryHrs = append(canaryHrs, hr)
		}
	}

	// set the generation, and revisions in the returned status
	newStatus := gstsv1alpha1.GameStatefulSetStatus{}
	newStatus.ObservedGeneration = set.Generation
	newStatus.CurrentRevision = currentRevision.Name
	newStatus.UpdateRevision = updateRevision.Name
	newStatus.CollisionCount = new(int32)
	*newStatus.CollisionCount = collisionCount
	copyStatus := set.Status.DeepCopy()
	newStatus.PreDeleteHookConditions = copyStatus.PreDeleteHookConditions

	// add Status.labelSelector
	util.ToLabelString(set.Spec.Selector)
	newStatus.LabelSelector = new(string)
	*newStatus.LabelSelector = util.ToLabelString(set.Spec.Selector)

	return &canaryContext{
		set:        set,
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

func (cCtx *canaryContext) AddPauseCondition(reason hookv1alpha1.PauseReason) {
	cCtx.pauseReasons = append(cCtx.pauseReasons, reason)
}

func (cCtx *canaryContext) HasAddPause() bool {
	return len(cCtx.pauseReasons) > 0
}
