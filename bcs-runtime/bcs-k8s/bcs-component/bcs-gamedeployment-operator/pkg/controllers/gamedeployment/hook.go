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
	"context"
	"os"
	"strconv"
	"strings"

	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util"
	canaryutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util/canary"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/hook"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	patchtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
)

// getHookRunsForGameDeployment list all HookRuns owned by a GameDeployment
func (gdc *defaultGameDeploymentControl) getHookRunsForGameDeployment(deploy *gdv1alpha1.GameDeployment) ([]*hookv1alpha1.HookRun, error) {
	hookRuns, err := gdc.hookRunLister.HookRuns(deploy.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	ownedByGd := make([]*hookv1alpha1.HookRun, 0)
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

	newCurrentHookRuns := []*hookv1alpha1.HookRun{}

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

	otherHrs, _ = commonhookutil.FilterHookRuns(otherHrs, func(ar *hookv1alpha1.HookRun) bool {
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

	hrsToDelete := commonhookutil.FilterHookRunsToDelete(otherHrs, canaryCtx.newStatus.UpdateRevision)
	err = gdc.deleteHookRuns(hrsToDelete)
	if err != nil {
		return err
	}

	return nil
}

// cancelHookRuns terminate HookRuns
func (gdc *defaultGameDeploymentControl) cancelHookRuns(canaryCtx *canaryContext, hookRuns []*hookv1alpha1.HookRun) error {
	for _, hr := range hookRuns {
		isNotCompleted := hr == nil || !hr.Status.Phase.Completed()
		if hr != nil && !hr.Spec.Terminate && isNotCompleted {
			klog.Infof("canceling the HookRun %s for GameDeployment %s/%s", hr.Name, canaryCtx.deploy.Namespace, canaryCtx.deploy.Name)
			_, err := gdc.hookClient.TkexV1alpha1().HookRuns(hr.Namespace).Patch(
				context.TODO(), hr.Name, patchtypes.MergePatchType, []byte(commonhookutil.CancelHookRun), metav1.PatchOptions{})
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
func (gdc *defaultGameDeploymentControl) deleteHookRuns(hrs []*hookv1alpha1.HookRun) error {
	for _, hr := range hrs {
		if hr.DeletionTimestamp != nil {
			continue
		}
		err := gdc.hookClient.TkexV1alpha1().HookRuns(hr.Namespace).Delete(context.TODO(), hr.Name, metav1.DeleteOptions{})
		if err != nil && !k8serrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

// reconcileStepHookRun reconcile canary step HookRun
func (gdc *defaultGameDeploymentControl) reconcileStepHookRun(canaryCtx *canaryContext) (*hookv1alpha1.HookRun, error) {
	deploy := canaryCtx.deploy
	currentHrs := canaryCtx.CurrentHookRuns()
	step, index := canaryutil.GetCurrentCanaryStep(deploy)
	currentHr := commonhookutil.FilterHookRunsByName(currentHrs, deploy.Status.Canary.CurrentStepHookRun)

	if len(deploy.Status.PauseConditions) > 0 {
		return currentHr, nil
	}

	if step == nil || step.Hook == nil || index == nil {
		err := gdc.cancelHookRuns(canaryCtx, []*hookv1alpha1.HookRun{currentHr})
		return nil, err
	}
	if currentHr == nil {
		// need to create new HookRun
		revision := canaryCtx.newStatus.UpdateRevision
		stepLabels := commonhookutil.StepLabels(*index, revision)
		currentHr, err := gdc.createHookRun(canaryCtx, step.Hook, index, stepLabels)
		if err == nil {
			klog.Infof("Created HookRun %s for step %d of GameDeployment %s/%s", currentHr.Name, *index, deploy.Namespace, deploy.Name)
		}
		return currentHr, err
	}

	switch currentHr.Status.Phase {
	case hookv1alpha1.HookPhaseInconclusive, hookv1alpha1.HookPhaseError, hookv1alpha1.HookPhaseFailed:
		canaryCtx.AddPauseCondition(hookv1alpha1.PauseReasonStepBasedHook)
	}
	return currentHr, nil
}

// createHookRun create HookRun
func (gdc *defaultGameDeploymentControl) createHookRun(canaryCtx *canaryContext, hookStep *hookv1alpha1.HookStep, stepIndex *int32,
	labels map[string]string) (*hookv1alpha1.HookRun, error) {
	arguments := []hookv1alpha1.Argument{}
	for _, arg := range hookStep.Args {
		value := arg.Value
		hookArg := hookv1alpha1.Argument{
			Name:  arg.Name,
			Value: &value,
		}
		arguments = append(arguments, hookArg)
	}
	hostIP := os.Getenv("HOST_IP")
	hostArgs := hookv1alpha1.Argument{
		Name:  "HostIP",
		Value: &hostIP,
	}
	arguments = append(arguments, hostArgs)

	hr, err := gdc.newHookRunFromGameDeployment(canaryCtx, hookStep, arguments, canaryCtx.newStatus.UpdateRevision, stepIndex, labels)
	if err != nil {
		return nil, err
	}
	hookRunIf := gdc.hookClient.TkexV1alpha1().HookRuns(canaryCtx.deploy.Namespace)
	return commonhookutil.CreateWithCollisionCounter(hookRunIf, *hr)
}

// newHookRunFromGameDeployment generate a HookRun from GameDeployment and HookTemplate
func (gdc *defaultGameDeploymentControl) newHookRunFromGameDeployment(canaryCtx *canaryContext, hookStep *hookv1alpha1.HookStep,
	args []hookv1alpha1.Argument, revision string, stepIdx *int32, labels map[string]string) (*hookv1alpha1.HookRun, error) {
	deploy := canaryCtx.deploy
	template, err := gdc.hookTemplateLister.HookTemplates(deploy.Namespace).Get(hookStep.TemplateName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.Warningf("HookTemplate '%s' not found for GameDeployment %s/%s", hookStep.TemplateName, deploy.Namespace, deploy.Name)
		}
		return nil, err
	}
	nameParts := []string{"canary", revision}
	if stepIdx != nil {
		nameParts = append(nameParts, strconv.Itoa(int(*stepIdx)))
	}
	nameParts = append(nameParts, hookStep.TemplateName)
	name := strings.Join(nameParts, "-")

	run, err := commonhookutil.NewHookRunFromTemplate(template, args, name, "", deploy.Namespace)
	if err != nil {
		return nil, err
	}
	run.Labels = labels
	run.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(deploy, util.ControllerKind)}
	return run, nil
}
