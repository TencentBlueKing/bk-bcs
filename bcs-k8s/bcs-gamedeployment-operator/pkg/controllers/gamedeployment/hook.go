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
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/util"
	canaryutil "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/util/canary"
	hooksutil "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/util/hook"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	patchtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

const (
	cancelHookRun = `{
		"spec": {
			"terminate": true
		}
	}`
)

// getHookRunsForGameDeployment list all HookRuns owned by a GameDeployment
func (gdc *defaultGameDeploymentControl) getHookRunsForGameDeployment(deploy *v1alpha1.GameDeployment) ([]*v1alpha1.HookRun, error) {
	hookRuns, err := gdc.hookRunLister.HookRuns(deploy.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	ownedByGd := make([]*v1alpha1.HookRun, 0)
	for _, hr := range hookRuns {
		controllerRef := metav1.GetControllerOf(hr)
		if controllerRef != nil && controllerRef.UID == deploy.UID {
			ownedByGd = append(ownedByGd, hr)
		}
	}
	return ownedByGd, nil
}

// reconcileHookRuns reconcile HookRuns
func (gdc *defaultGameDeploymentControl) reconcileHookRuns(canaryCtx *canaryContext) error {
	otherHrs := canaryCtx.OtherHookRuns()

	newCurrentHookRuns := []*v1alpha1.HookRun{}

	if canaryCtx.deploy.Spec.UpdateStrategy.CanaryStrategy != nil && canaryCtx.deploy.Status.Canary.Revision != "" {
		stepHookRun, err := gdc.reconcileStepHookRun(canaryCtx)
		if err != nil {
			return err
		}
		if stepHookRun != nil {
			newCurrentHookRuns = append(newCurrentHookRuns, stepHookRun)
		}
	}

	canaryCtx.SetCurrentHookRuns(newCurrentHookRuns)

	otherHrs, _ = hooksutil.FilterHookRuns(otherHrs, func(ar *v1alpha1.HookRun) bool {
		for _, currentHr := range newCurrentHookRuns {
			if ar.Name == currentHr.Name {
				return false
			}
		}
		return true
	})

	err := gdc.cancelHookRuns(canaryCtx, otherHrs)
	if err != nil {
		return err
	}

	hrsToDelete := hooksutil.FilterHookRunsToDelete(otherHrs, canaryCtx.newStatus.UpdateRevision)
	err = gdc.deleteHookRuns(hrsToDelete)
	if err != nil {
		return err
	}

	return nil
}

// cancelHookRuns terminate HookRuns
func (gdc *defaultGameDeploymentControl) cancelHookRuns(canaryCtx *canaryContext, hookRuns []*v1alpha1.HookRun) error {
	for _, hr := range hookRuns {
		isNotCompleted := hr == nil || !hr.Status.Phase.Completed()
		if hr != nil && !hr.Spec.Terminate && isNotCompleted {
			klog.Infof("canceling the HookRun %s for GameDeployment %s/%s", hr.Name, canaryCtx.deploy.Namespace, canaryCtx.deploy.Name)
			_, err := gdc.client.TkexV1alpha1().HookRuns(hr.Namespace).Patch(hr.Name, patchtypes.MergePatchType, []byte(cancelHookRun))
			if err != nil {
				if k8serrors.IsNotFound(err) {
					klog.Warningf("HookRun %s not found for GameDeployment %s/%s", hr.Name, canaryCtx.deploy.Namespace, canaryCtx.deploy.Name)
					continue
				}
				return err
			}
		}
	}
	return nil
}

// deleteHookRuns delete HookRuns from k8s
func (gdc *defaultGameDeploymentControl) deleteHookRuns(hrs []*v1alpha1.HookRun) error {
	for _, hr := range hrs {
		if hr.DeletionTimestamp != nil {
			continue
		}
		err := gdc.client.TkexV1alpha1().HookRuns(hr.Namespace).Delete(hr.Name, nil)
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

// reconcileStepHookRun reconcile canary step HookRun
func (gdc *defaultGameDeploymentControl) reconcileStepHookRun(canaryCtx *canaryContext) (*v1alpha1.HookRun, error) {
	deploy := canaryCtx.deploy
	currentHrs := canaryCtx.CurrentHookRuns()
	step, index := canaryutil.GetCurrentCanaryStep(deploy)
	currentHr := hooksutil.FilterHookRunsByName(currentHrs, deploy.Status.Canary.CurrentStepHookRun)

	if len(deploy.Status.PauseConditions) > 0 {
		return currentHr, nil
	}

	if step == nil || step.Hook == nil || index == nil {
		err := gdc.cancelHookRuns(canaryCtx, []*v1alpha1.HookRun{currentHr})
		return nil, err
	}
	if currentHr == nil {
		// need to create new HookRun
		revision := canaryCtx.newStatus.UpdateRevision
		stepLabels := hooksutil.StepLabels(*index, revision)
		currentHr, err := gdc.createHookRun(canaryCtx, step.Hook, index, stepLabels)
		if err == nil {
			klog.Infof("Created HookRun %s for step %d of GameDeployment %s/%s", currentHr.Name, *index, deploy.Namespace, deploy.Name)
		}
		return currentHr, err
	}

	switch currentHr.Status.Phase {
	case v1alpha1.HookPhaseInconclusive, v1alpha1.HookPhaseError, v1alpha1.HookPhaseFailed:
		canaryCtx.AddPauseCondition(v1alpha1.PauseReasonStepBasedHook)
	}
	return currentHr, nil
}

// createHookRun create HookRun
func (gdc *defaultGameDeploymentControl) createHookRun(canaryCtx *canaryContext, hookStep *v1alpha1.HookStep, stepIndex *int32, labels map[string]string) (*v1alpha1.HookRun, error) {
	arguments := []v1alpha1.Argument{}
	for _, arg := range hookStep.Args {
		value := arg.Value
		hookArg := v1alpha1.Argument{
			Name:  arg.Name,
			Value: &value,
		}
		arguments = append(arguments, hookArg)
	}

	hr, err := gdc.newHookRunFromGameDeployment(canaryCtx, hookStep, arguments, canaryCtx.newStatus.UpdateRevision, stepIndex, labels)
	if err != nil {
		return nil, err
	}
	hookRunIf := gdc.client.TkexV1alpha1().HookRuns(canaryCtx.deploy.Namespace)
	return hooksutil.CreateWithCollisionCounter(hookRunIf, *hr)
}

// newHookRunFromGameDeployment generate a HookRun from GameDeployment and HookTemplate
func (gdc *defaultGameDeploymentControl) newHookRunFromGameDeployment(canaryCtx *canaryContext, hookStep *v1alpha1.HookStep, args []v1alpha1.Argument, revision string, stepIdx *int32, labels map[string]string) (*v1alpha1.HookRun, error) {
	deploy := canaryCtx.deploy
	template, err := gdc.hookTemplateLister.HookTemplates(deploy.Namespace).Get(hookStep.TemplateName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.Warningf("HookTemplate '%s' not found for GameDeployment %s/%s", hookStep.TemplateName, deploy.Namespace, deploy.Name)
		}
		return nil, err
	}
	nameParts := []string{revision}
	if stepIdx != nil {
		nameParts = append(nameParts, strconv.Itoa(int(*stepIdx)))
	}
	nameParts = append(nameParts, hookStep.TemplateName)
	name := strings.Join(nameParts, "-")

	run, err := hooksutil.NewHookRunFromTemplate(template, args, name, "", deploy.Namespace)
	if err != nil {
		return nil, err
	}
	run.Labels = labels
	run.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(deploy, util.ControllerKind)}
	return run, nil
}
