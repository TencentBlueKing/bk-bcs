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
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	hookclientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/clientset/versioned"
	hooklister "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/client/listers/tkex/v1alpha1"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/hook"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

const (
	PodNameArgKey      = "PodName"
	NamespaceArgKey    = "PodNamespace"
	PodIPArgKey        = "PodIP"
	PodImageArgKey     = "PodContainer"
	HostArgKey         = "HostIP"
	DeletingAnnotation = "io.tencent.bcs.dev/game-pod-deleting"
)

type PreDeleteInterface interface {
	CheckDelete(obj PreDeleteHookObjectInterface, pod *v1.Pod, newStatus PreDeleteHookStatusInterface, podNameLabelKey string) (bool, error)
}

type PreDeleteControl struct {
	kubeClient         clientset.Interface
	hookClient         hookclientset.Interface
	recorder           record.EventRecorder
	hookRunLister      hooklister.HookRunLister
	hookTemplateLister hooklister.HookTemplateLister
}

func New(kubeClient clientset.Interface, hookClient hookclientset.Interface, recorder record.EventRecorder,
	hookRunLister hooklister.HookRunLister, hookTemplateLister hooklister.HookTemplateLister) PreDeleteInterface {
	return &PreDeleteControl{kubeClient: kubeClient, hookClient: hookClient, recorder: recorder, hookRunLister: hookRunLister,
		hookTemplateLister: hookTemplateLister}
}

// CheckDelete check whether the pod can be deleted safely
func (p *PreDeleteControl) CheckDelete(obj PreDeleteHookObjectInterface, pod *v1.Pod, newStatus PreDeleteHookStatusInterface, podNameLabelKey string) (bool, error) {
	if pod.Status.Phase != v1.PodRunning {
		return true, nil
	}
	metaObj, ok := obj.(metav1.Object)
	if !ok {
		return false, fmt.Errorf("error decoding object to meta object for checking predelete hook, invalid type")
	}
	runtimeObj, ok1 := obj.(runtime.Object)
	if !ok1 {
		return false, fmt.Errorf("error decoding object to runtime object for checking predelete hook, invalid type")
	}
	objectKind := runtimeObj.GetObjectKind().GroupVersionKind().Kind
	namespace := metaObj.GetNamespace()
	name := metaObj.GetName()

	preDeleteHook := obj.GetPreDeleteHook()
	if preDeleteHook == nil {
		return true, nil
	}

	preDeleteLabels := map[string]string{
		commonhookutil.HookRunTypeLabel:            commonhookutil.HookRunTypePreDeleteLabel,
		commonhookutil.WorkloadRevisionUniqueLabel: pod.Labels[apps.ControllerRevisionHashLabelKey],
		commonhookutil.PodInstanceID:               pod.Labels[podNameLabelKey],
	}

	labelSelector := &metav1.LabelSelector{
		MatchLabels: preDeleteLabels,
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
		preDeleteHookRun, err := p.createHookRun(metaObj, runtimeObj, preDeleteHook, pod, preDeleteLabels, podNameLabelKey)
		if err != nil {
			klog.Warningf("Created PreDelete HookRun failed for pod %s of %s %s/%s, err:%s",
				pod.Name, objectKind, namespace, name, err)
			return false, err
		}

		updatePreDeleteHookCondition(newStatus, pod.Name)
		klog.Infof("Created PreDelete HookRun %s for pod %s of %s %s/%s", preDeleteHookRun.Name, pod.Name, objectKind, namespace, name)

		err = p.injectPodDeletingAnnotation(pod)
		if err != nil {
			return false, err
		}

		return false, nil
	}
	if existHookRuns[0].Status.Phase == hookv1alpha1.HookPhaseSuccessful {
		err := deletePreDeleteHookCondition(newStatus, pod.Name)
		if err != nil {
			klog.Warningf("expected the %s %s/%s exists a PreDeleteHookCondition for pod %s, but got an error: %s",
				objectKind, namespace, name, pod.Name, err.Error())
		}
		return true, nil
	}

	err = resetPreDeleteHookConditionPhase(newStatus, pod.Name, existHookRuns[0].Status.Phase)
	if err != nil {
		klog.Warningf("expected the %s %s/%s exists a PreDeleteHookCondition for pod %s, but got an error: %s",
			objectKind, namespace, name, pod.Name, err.Error())
	}

	return false, nil
}

// createHookRun create a PreDelete HookRun
func (p *PreDeleteControl) createHookRun(metaObj metav1.Object, runtimeObj runtime.Object, preDeleteHook *hookv1alpha1.HookStep,
	pod *v1.Pod, labels map[string]string, podNameLabelKey string) (*hookv1alpha1.HookRun, error) {
	arguments := []hookv1alpha1.Argument{}
	for _, arg := range preDeleteHook.Args {
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
		{
			Name:  HostArgKey,
			Value: &pod.Status.HostIP,
		},
	}
	arguments = append(arguments, podArgs...)

	for i, value := range pod.Spec.Containers {
		tmp := new(string)
		*tmp = value.Name
		podArgs = []hookv1alpha1.Argument{
			{
				Name:  PodImageArgKey + "[" + strconv.Itoa(i) + "]",
				Value: tmp,
			},
		}
		arguments = append(arguments, podArgs...)
	}

	hr, err := p.newHookRunFromHookTemplate(metaObj, runtimeObj, arguments, pod, preDeleteHook, labels, podNameLabelKey)
	if err != nil {
		return nil, err
	}
	hookRunIf := p.hookClient.TkexV1alpha1().HookRuns(pod.Namespace)
	return commonhookutil.CreateWithCollisionCounter(hookRunIf, *hr)
}

// newHookRunFromGameStatefulSet generate a HookRun from HookTemplate
func (p *PreDeleteControl) newHookRunFromHookTemplate(metaObj metav1.Object, runtimeObj runtime.Object, args []hookv1alpha1.Argument,
	pod *v1.Pod, preDeleteHook *hookv1alpha1.HookStep, labels map[string]string, podNameLabelKey string) (*hookv1alpha1.HookRun, error) {
	template, err := p.hookTemplateLister.HookTemplates(pod.Namespace).Get(preDeleteHook.TemplateName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.Warningf("HookTemplate '%s' not found for %s/%s", preDeleteHook.TemplateName, pod.Namespace, pod.Name)
		}
		return nil, err
	}

	nameParts := []string{"predelete", pod.Labels[apps.ControllerRevisionHashLabelKey], pod.Labels[podNameLabelKey], preDeleteHook.TemplateName}
	name := strings.Join(nameParts, "-")

	run, err := commonhookutil.NewHookRunFromTemplate(template, args, name, "", pod.Namespace)
	if err != nil {
		return nil, err
	}
	run.Labels = labels
	run.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(metaObj, runtimeObj.GetObjectKind().GroupVersionKind())}
	return run, nil
}

// delete PreDeleteHookCondition of a pod
func deletePreDeleteHookCondition(status PreDeleteHookStatusInterface, podName string) error {
	var index int
	found := false
	conditions := status.GetPreDeleteHookConditions()
	for i, cond := range conditions {
		if cond.PodName == podName {
			found = true
			index = i
			break
		}
	}
	if !found {
		return fmt.Errorf("no PreDeleteHookCondition to delete")
	}

	newConditions := append(conditions[:index], conditions[index+1:]...)
	status.SetPreDeleteHookConditions(newConditions)
	return nil
}

func updatePreDeleteHookCondition(status PreDeleteHookStatusInterface, podName string) {
	conditions := status.GetPreDeleteHookConditions()
	for i, cond := range conditions {
		if cond.PodName == podName {
			conditions[i].HookPhase = hookv1alpha1.HookPhasePending
			conditions[i].StartTime = metav1.Now()
			status.SetPreDeleteHookConditions(conditions)
			return
		}
	}
	conditions = append(conditions, hookv1alpha1.PreDeleteHookCondition{
		PodName:   podName,
		StartTime: metav1.Now(),
		HookPhase: hookv1alpha1.HookPhasePending,
	})
	status.SetPreDeleteHookConditions(conditions)
}

// reset PreDeleteHookConditionPhase of a pod
func resetPreDeleteHookConditionPhase(status PreDeleteHookStatusInterface, podName string, phase hookv1alpha1.HookPhase) error {
	var index int
	found := false
	conditions := status.GetPreDeleteHookConditions()
	for i, cond := range conditions {
		if cond.PodName == podName {
			found = true
			index = i
			break
		}
	}
	if !found {
		return fmt.Errorf("no PreDeleteHookCondition to reset phase")
	}

	conditions[index].HookPhase = phase
	status.SetPreDeleteHookConditions(conditions)
	return nil
}

// injectPodDeletingAnnotation injects an annotation after creating predelete hook
func (p *PreDeleteControl) injectPodDeletingAnnotation(pod *v1.Pod) error {
	currentAnnotations := pod.ObjectMeta.DeepCopy().Annotations
	if currentAnnotations == nil {
		currentAnnotations = map[string]string{}
	}
	currentAnnotations[DeletingAnnotation] = "true"
	if reflect.DeepEqual(currentAnnotations, pod.Annotations) {
		return nil
	}
	patchData := map[string]interface{}{
		"metadata": map[string]map[string]string{
			"annotations": currentAnnotations,
		},
	}
	playLoadBytes, err := json.Marshal(patchData)
	if err != nil {
		return err
	}
	_, err = p.kubeClient.CoreV1().Pods(pod.Namespace).Patch(context.TODO(), pod.Name,
		types.StrategicMergePatchType, playLoadBytes, metav1.PatchOptions{})
	return err
}
