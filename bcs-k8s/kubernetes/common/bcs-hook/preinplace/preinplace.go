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

package preinplace

import (
	"fmt"
	"strconv"
	"strings"

	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	hookclientset "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned"
	hooklister "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/bcs-hook/client/listers/tkex/v1alpha1"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/common/util/hook"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

const (
	PodNameArgKey   = "PodName"
	NamespaceArgKey = "PodNamespace"
	PodIPArgKey     = "PodIP"
	PodImageArgKey  = "PodContainer"
)

type PreInplaceInterface interface {
	CheckInplace(obj PreInplaceHookObjectInterface,
		pod *v1.Pod,
		newStatus PreInplaceHookStatusInterface,
		podNameLabelKey string) (bool, error)
}

type PreInplaceControl struct {
	kubeClient         clientset.Interface
	hookClient         hookclientset.Interface
	recorder           record.EventRecorder
	hookRunLister      hooklister.HookRunLister
	hookTemplateLister hooklister.HookTemplateLister
}

func New(kubeClient clientset.Interface, hookClient hookclientset.Interface, recorder record.EventRecorder,
	hookRunLister hooklister.HookRunLister, hookTemplateLister hooklister.HookTemplateLister) PreInplaceInterface {
	return &PreInplaceControl{kubeClient: kubeClient, hookClient: hookClient, recorder: recorder,
		hookRunLister: hookRunLister, hookTemplateLister: hookTemplateLister}
}

// CheckInplace check whether the pod can be deleted safely
func (p *PreInplaceControl) CheckInplace(obj PreInplaceHookObjectInterface, pod *v1.Pod,
	newStatus PreInplaceHookStatusInterface, podNameLabelKey string) (bool, error) {
	metaObj, ok := obj.(metav1.Object)
	if !ok {
		return false, fmt.Errorf(
			"error decoding object to meta object for checking preinplace hook, invalid type")
	}
	runtimeObj, ok1 := obj.(runtime.Object)
	if !ok1 {
		return false, fmt.Errorf(
			"error decoding object to runtime object for checking preinplace hook, invalid type")
	}
	objectKind := runtimeObj.GetObjectKind().GroupVersionKind().Kind
	namespace := metaObj.GetNamespace()
	name := metaObj.GetName()

	preInplaceHook := obj.GetPreInplaceHook()
	if preInplaceHook == nil {
		return true, nil
	}

	preInplaceLabels := map[string]string{
		commonhookutil.HookRunTypeLabel:            commonhookutil.HookRunTypePreInplaceLabel,
		commonhookutil.WorkloadRevisionUniqueLabel: pod.Labels[apps.ControllerRevisionHashLabelKey],
		commonhookutil.PodInstanceID:               pod.Labels[podNameLabelKey],
	}

	labelSelector := &metav1.LabelSelector{
		MatchLabels: preInplaceLabels,
	}
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return false, fmt.Errorf("invalid label selector: %s", err.Error())
	}
	existHookRuns, err := p.hookRunLister.HookRuns(namespace).List(selector)
	if err != nil {
		return false, err
	}
	if len(existHookRuns) == 0 {
		preInplaceHookRun, err := p.createHookRun(metaObj, runtimeObj,
			preInplaceHook, pod, preInplaceLabels, podNameLabelKey)
		if err != nil {
			return false, err
		}

		updatePreInplaceHookCondition(newStatus, pod.Name)
		klog.Infof("Created PreInplace HookRun %s for pod %s of %s %s/%s",
			preInplaceHookRun.Name, pod.Name, objectKind, namespace, name)
		return false, nil
	}
	if existHookRuns[0].Status.Phase == hookv1alpha1.HookPhaseSuccessful {
		err := deletePreInplaceHookCondition(newStatus, pod.Name)
		if err != nil {
			klog.Warningf("expected the %s %s/%s exists a PreInplaceHookCondition for pod %s, but got an error: %s",
				objectKind, namespace, name, pod.Name, err.Error())
		}
		return true, nil
	}

	err = resetPreInplaceHookConditionPhase(newStatus, pod.Name, existHookRuns[0].Status.Phase)
	if err != nil {
		klog.Warningf("expected the %s %s/%s exists a PreInplaceHookCondition for pod %s, but got an error: %s",
			objectKind, namespace, name, pod.Name, err.Error())
	}

	return false, nil
}

// createHookRun create a PreInplace HookRun
func (p *PreInplaceControl) createHookRun(metaObj metav1.Object, runtimeObj runtime.Object,
	preInplaceHook *hookv1alpha1.HookStep, pod *v1.Pod, labels map[string]string,
	podNameLabelKey string) (*hookv1alpha1.HookRun, error) {
	arguments := []hookv1alpha1.Argument{}
	for _, arg := range preInplaceHook.Args {
		value := arg.Value
		hookArg := hookv1alpha1.Argument{
			Name:  arg.Name,
			Value: &value,
		}
		arguments = append(arguments, hookArg)
	}
	// add PodName and PodNamespace args
	podArgs := []hookv1alpha1.Argument{
		{
			Name:  PodNameArgKey,
			Value: &pod.Name,
		},
		{
			Name:  NamespaceArgKey,
			Value: &pod.Namespace,
		},
		{
			Name:  PodIPArgKey,
			Value: &pod.Status.PodIP,
		},
	}
	arguments = append(arguments, podArgs...)

	for i, value := range pod.Spec.Containers {
		imageArgs := []hookv1alpha1.Argument{
			{
				Name: PodImageArgKey + "[" + strconv.Itoa(i) + "]",
				Value: &value.Name,
			},
		}
		arguments = append(arguments, imageArgs...)
	}

	hr, err := p.newHookRunFromHookTemplate(metaObj, runtimeObj, arguments, pod, preInplaceHook, labels, podNameLabelKey)
	if err != nil {
		return nil, err
	}
	hookRunIf := p.hookClient.TkexV1alpha1().HookRuns(pod.Namespace)
	return commonhookutil.CreateWithCollisionCounter(hookRunIf, *hr)
}

// newHookRunFromGameStatefulSet generate a HookRun from HookTemplate
func (p *PreInplaceControl) newHookRunFromHookTemplate(metaObj metav1.Object,
	runtimeObj runtime.Object, args []hookv1alpha1.Argument,
	pod *v1.Pod, preInplaceHook *hookv1alpha1.HookStep, labels map[string]string,
	podNameLabelKey string) (*hookv1alpha1.HookRun, error) {
	template, err := p.hookTemplateLister.HookTemplates(pod.Namespace).Get(preInplaceHook.TemplateName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.Warningf("HookTemplate '%s' not found for %s/%s",
				preInplaceHook.TemplateName, pod.Namespace, pod.Name)
		}
		return nil, err
	}

	nameParts := []string{"preinplace", pod.Labels[apps.ControllerRevisionHashLabelKey],
		pod.Labels[podNameLabelKey], preInplaceHook.TemplateName}
	name := strings.Join(nameParts, "-")

	run, err := commonhookutil.NewHookRunFromTemplate(template, args, name, "", pod.Namespace)
	if err != nil {
		return nil, err
	}
	run.Labels = labels
	run.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(metaObj,
		runtimeObj.GetObjectKind().GroupVersionKind())}
	return run, nil
}

// delete PreInplaceHookCondition of a pod
func deletePreInplaceHookCondition(status PreInplaceHookStatusInterface, podName string) error {
	var index int
	found := false
	conditions := status.GetPreInplaceHookConditions()
	for i, cond := range conditions {
		if cond.PodName == podName {
			found = true
			index = i
			break
		}
	}
	if !found {
		return fmt.Errorf("no PreInplaceHookCondition to delete")
	}

	newConditions := append(conditions[:index], conditions[index+1:]...)
	status.SetPreInplaceHookConditions(newConditions)
	return nil
}

func updatePreInplaceHookCondition(status PreInplaceHookStatusInterface, podName string) {
	conditions := status.GetPreInplaceHookConditions()
	for i, cond := range conditions {
		if cond.PodName == podName {
			conditions[i].HookPhase = hookv1alpha1.HookPhasePending
			conditions[i].StartTime = metav1.Now()
			status.SetPreInplaceHookConditions(conditions)
			return
		}
	}
	conditions = append(conditions, hookv1alpha1.PreInplaceHookCondition{
		PodName:   podName,
		StartTime: metav1.Now(),
		HookPhase: hookv1alpha1.HookPhasePending,
	})
	status.SetPreInplaceHookConditions(conditions)
}

// reset PreInplaceHookConditionPhase of a pod
func resetPreInplaceHookConditionPhase(status PreInplaceHookStatusInterface, podName string,
	phase hookv1alpha1.HookPhase) error {
	var index int
	found := false
	conditions := status.GetPreInplaceHookConditions()
	for i, cond := range conditions {
		if cond.PodName == podName {
			found = true
			index = i
			break
		}
	}
	if !found {
		return fmt.Errorf("no PreInplaceHookCondition to reset phase")
	}

	conditions[index].HookPhase = phase
	status.SetPreInplaceHookConditions(conditions)
	return nil
}
