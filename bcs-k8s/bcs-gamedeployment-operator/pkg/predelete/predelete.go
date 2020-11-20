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

package predelete

import (
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	tkexclientset "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/client/clientset/versioned"
	gamedeploylister "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/client/listers/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/util"
	hooksutil "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/util/hook"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

const (
	PodNameArgKey   = "PodName"
	NamespaceArgKey = "PodNamespace"
	PodIPArgKey     = "PodIP"
)

type PreDeleteInterface interface {
	CheckDelete(deploy *v1alpha1.GameDeployment, pod *v1.Pod, newStatus *v1alpha1.GameDeploymentStatus) (bool, error)
}

type GameDeploymentPreDeleteControl struct {
	kubeClient         clientset.Interface
	tkexClient         tkexclientset.Interface
	recorder           record.EventRecorder
	hookRunLister      gamedeploylister.HookRunLister
	hookTemplateLister gamedeploylister.HookTemplateLister
}

func New(kubeClient clientset.Interface, tkexClient tkexclientset.Interface, recorder record.EventRecorder,
	hookRunLister gamedeploylister.HookRunLister, hookTemplateLister gamedeploylister.HookTemplateLister) PreDeleteInterface {
	return &GameDeploymentPreDeleteControl{kubeClient: kubeClient, tkexClient: tkexClient, recorder: recorder, hookRunLister: hookRunLister,
		hookTemplateLister: hookTemplateLister}
}

// CheckDelete check whether the pod can be deleted safely
func (p *GameDeploymentPreDeleteControl) CheckDelete(deploy *v1alpha1.GameDeployment, pod *v1.Pod, newStatus *v1alpha1.GameDeploymentStatus) (bool, error) {
	if deploy.Spec.PreDeleteUpdateStrategy.Hook == nil {
		return true, nil
	}

	preDeleteLabels := map[string]string{
		v1alpha1.GameDeploymentTypeLabel:             v1alpha1.GameDeploymentTypePreDeleteLabel,
		v1alpha1.GameDeploymentPodControllerRevision: pod.Labels[apps.ControllerRevisionHashLabelKey],
		v1alpha1.GameDeploymentPodInstanceID:         pod.Labels[v1alpha1.GameDeploymentInstanceID],
	}

	labelSelector := &metav1.LabelSelector{
		MatchLabels: preDeleteLabels,
	}
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return false, fmt.Errorf("invalid label selector: %s", err.Error())
	}
	existHookRuns, err := p.hookRunLister.HookRuns(deploy.Namespace).List(selector)
	if err != nil {
		return false, err
	}
	if len(existHookRuns) == 0 {
		preDeleteHookRun, err := p.createHookRun(deploy, pod, preDeleteLabels)
		if err != nil {
			return false, err
		}

		updatePreDeleteHookCondition(newStatus, pod.Name)
		klog.Infof("Created PreDelete HookRun %s for pod %s of GameDeployment %s/%s", preDeleteHookRun.Name, pod.Name, deploy.Namespace, deploy.Name)
		return false, nil
	}
	if existHookRuns[0].Status.Phase == v1alpha1.HookPhaseSuccessful {
		err := deletePreDeleteHookCondition(newStatus, pod.Name)
		if err != nil {
			klog.Warningf("expected the GameDeployment %s/%s exists a PreDeleteHookCondition for pod %s, but got an error: %s",
				deploy.Namespace, deploy.Name, pod.Name, err.Error())
		}
		return true, nil
	}

	err = resetPreDeleteHookConditionPhase(newStatus, pod.Name, existHookRuns[0].Status.Phase)
	if err != nil {
		klog.Warningf("expected the GameDeployment %s/%s exists a PreDeleteHookCondition for pod %s, but got an error: %s",
			deploy.Namespace, deploy.Name, pod.Name, err.Error())
	}

	return false, nil
}

// createHookRun create a PreDelete HookRun
func (p *GameDeploymentPreDeleteControl) createHookRun(deploy *v1alpha1.GameDeployment, pod *v1.Pod, labels map[string]string) (*v1alpha1.HookRun, error) {
	preDeleteHook := deploy.Spec.PreDeleteUpdateStrategy.Hook
	arguments := []v1alpha1.Argument{}
	for _, arg := range preDeleteHook.Args {
		value := arg.Value
		hookArg := v1alpha1.Argument{
			Name:  arg.Name,
			Value: &value,
		}
		arguments = append(arguments, hookArg)
	}
	// add PodName and PodNamespace args
	podArgs := []v1alpha1.Argument{
		{
			Name:  PodNameArgKey,
			Value: &pod.Name,
		},
		{
			Name:  NamespaceArgKey,
			Value: &deploy.Namespace,
		},
		{
			Name:  PodIPArgKey,
			Value: &pod.Status.PodIP,
		},
	}
	arguments = append(arguments, podArgs...)

	hr, err := p.newHookRunFromGameDeployment(deploy, arguments, pod, preDeleteHook, labels)
	if err != nil {
		return nil, err
	}
	hookRunIf := p.tkexClient.TkexV1alpha1().HookRuns(deploy.Namespace)
	return hooksutil.CreateWithCollisionCounter(hookRunIf, *hr)
}

// newHookRunFromGameDeployment generate a HookRun from GameDeployment and HookTemplate
func (p *GameDeploymentPreDeleteControl) newHookRunFromGameDeployment(deploy *v1alpha1.GameDeployment, args []v1alpha1.Argument,
	pod *v1.Pod, preDeleteHook *v1alpha1.HookStep, labels map[string]string) (*v1alpha1.HookRun, error) {
	template, err := p.hookTemplateLister.HookTemplates(deploy.Namespace).Get(preDeleteHook.TemplateName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.Warningf("HookTemplate '%s' not found for GameDeployment %s/%s", preDeleteHook.TemplateName, deploy.Namespace, deploy.Name)
		}
		return nil, err
	}

	nameParts := []string{pod.Labels[apps.ControllerRevisionHashLabelKey], pod.Labels[v1alpha1.GameDeploymentInstanceID], preDeleteHook.TemplateName}
	name := strings.Join(nameParts, "-")

	run, err := hooksutil.NewHookRunFromTemplate(template, args, name, "", deploy.Namespace)
	if err != nil {
		return nil, err
	}
	run.Labels = labels
	run.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(deploy, util.ControllerKind)}
	return run, nil
}

// delete PreDeleteHookCondition of a pod
func deletePreDeleteHookCondition(status *v1alpha1.GameDeploymentStatus, podName string) error {
	var index int
	found := false
	for i, cond := range status.PreDeleteHookConditions {
		if cond.PodName == podName {
			found = true
			index = i
			break
		}
	}
	if !found {
		return fmt.Errorf("no PreDeleteHookCondition to delete")
	}

	status.PreDeleteHookConditions = append(status.PreDeleteHookConditions[:index], status.PreDeleteHookConditions[index+1:]...)
	return nil
}

func updatePreDeleteHookCondition(status *v1alpha1.GameDeploymentStatus, podName string) {
	for i, cond := range status.PreDeleteHookConditions {
		if cond.PodName == podName {
			status.PreDeleteHookConditions[i].HookPhase = v1alpha1.HookPhasePending
			status.PreDeleteHookConditions[i].StartTime = metav1.Now()
			return
		}
	}
	status.PreDeleteHookConditions = append(status.PreDeleteHookConditions, v1alpha1.PreDeleteHookCondition{
		PodName:   podName,
		StartTime: metav1.Now(),
		HookPhase: v1alpha1.HookPhasePending,
	})
}

// reset PreDeleteHookConditionPhase of a pod
func resetPreDeleteHookConditionPhase(status *v1alpha1.GameDeploymentStatus, podName string, phase v1alpha1.HookPhase) error {
	var index int
	found := false
	for i, cond := range status.PreDeleteHookConditions {
		if cond.PodName == podName {
			found = true
			index = i
			break
		}
	}
	if !found {
		return fmt.Errorf("no PreDeleteHookCondition to reset phase")
	}

	status.PreDeleteHookConditions[index].HookPhase = phase
	return nil
}
