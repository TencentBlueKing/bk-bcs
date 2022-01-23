/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either expostss or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package postinplace

import (
	"context"
	"fmt"
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
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog"
)

const (
	PodNameArgKey   = "PodName"
	NamespaceArgKey = "PodNamespace"
	PodIPArgKey     = "PodIP"
	PodImageArgKey  = "PodContainer"
	HostArgKey      = "HostIP"
)

type PostInplaceInterface interface {
	CreatePostInplaceHook(obj PostInplaceHookObjectInterface,
		pod *v1.Pod,
		newStatus PostInplaceHookStatusInterface,
		podNameLabelKey string) (bool, error)

	UpdatePostInplaceHook(obj PostInplaceHookObjectInterface,
		pod *v1.Pod,
		newStatus PostInplaceHookStatusInterface,
		podNameLabelKey string) error
}

type PostInplaceControl struct {
	kubeClient         clientset.Interface
	hookClient         hookclientset.Interface
	recorder           record.EventRecorder
	hookRunLister      hooklister.HookRunLister
	hookTemplateLister hooklister.HookTemplateLister
}

func New(kubeClient clientset.Interface, hookClient hookclientset.Interface, recorder record.EventRecorder,
	hookRunLister hooklister.HookRunLister, hookTemplateLister hooklister.HookTemplateLister) PostInplaceInterface {
	return &PostInplaceControl{kubeClient: kubeClient, hookClient: hookClient, recorder: recorder,
		hookRunLister: hookRunLister, hookTemplateLister: hookTemplateLister}
}

// CreatePostInplaceHook create post hook, returns whether creates a hook this time and error
func (p *PostInplaceControl) CreatePostInplaceHook(obj PostInplaceHookObjectInterface, pod *v1.Pod,
	newStatus PostInplaceHookStatusInterface, podNameLabelKey string) (bool, error) {
	metaObj, ok := obj.(metav1.Object)
	if !ok {
		return false, fmt.Errorf(
			"error decoding object to meta object for checking postinplace hook, invalid type")
	}
	runtimeObj, ok1 := obj.(runtime.Object)
	if !ok1 {
		return false, fmt.Errorf(
			"error decoding object to runtime object for checking postinplace hook, invalid type")
	}
	objectKind := runtimeObj.GetObjectKind().GroupVersionKind().Kind
	namespace := metaObj.GetNamespace()
	name := metaObj.GetName()

	postInplaceHook := obj.GetPostInplaceHook()
	if postInplaceHook == nil {
		return false, nil
	}

	postInplaceLabels := map[string]string{
		commonhookutil.HookRunTypeLabel:            commonhookutil.HookRunTypePostInplaceLabel,
		commonhookutil.WorkloadRevisionUniqueLabel: pod.Labels[apps.ControllerRevisionHashLabelKey],
		commonhookutil.PodInstanceID:               pod.Labels[podNameLabelKey],
	}

	labelSelector := &metav1.LabelSelector{
		MatchLabels: postInplaceLabels,
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
		postInplaceHookRun, err := p.createHookRun(metaObj, runtimeObj,
			postInplaceHook, pod, postInplaceLabels, podNameLabelKey)
		if err != nil {
			return false, err
		}

		updatePostInplaceHookCondition(newStatus, pod.Name)
		klog.Infof("Created PostInplace HookRun %s for pod %s of %s %s/%s",
			postInplaceHookRun.Name, pod.Name, objectKind, namespace, name)
		return true, nil
	}
	return false, nil
}

// UpdatePostInplaceHook resyncs post inplace hook condition
func (p *PostInplaceControl) UpdatePostInplaceHook(obj PostInplaceHookObjectInterface, pod *v1.Pod,
	newStatus PostInplaceHookStatusInterface, podNameLabelKey string) error {
	metaObj, ok := obj.(metav1.Object)
	if !ok {
		return fmt.Errorf(
			"error decoding object to meta object for checking postinplace hook, invalid type")
	}
	runtimeObj, ok1 := obj.(runtime.Object)
	if !ok1 {
		return fmt.Errorf(
			"error decoding object to runtime object for checking postinplace hook, invalid type")
	}
	objectKind := runtimeObj.GetObjectKind().GroupVersionKind().Kind
	namespace := metaObj.GetNamespace()
	name := metaObj.GetName()

	postInplaceHook := obj.GetPostInplaceHook()
	if postInplaceHook == nil {
		return nil
	}

	postInplaceLabels := map[string]string{
		commonhookutil.HookRunTypeLabel:            commonhookutil.HookRunTypePostInplaceLabel,
		commonhookutil.WorkloadRevisionUniqueLabel: pod.Labels[apps.ControllerRevisionHashLabelKey],
		commonhookutil.PodInstanceID:               pod.Labels[podNameLabelKey],
	}

	labelSelector := &metav1.LabelSelector{
		MatchLabels: postInplaceLabels,
	}
	selector, err := metav1.LabelSelectorAsSelector(labelSelector)
	if err != nil {
		return fmt.Errorf("invalid label selector: %s", err.Error())
	}
	existHookRuns, err := p.hookRunLister.HookRuns(namespace).List(selector)
	if err != nil {
		return err
	}
	if len(existHookRuns) == 0 {
		return nil
	}
	if existHookRuns[0].Status.Phase == hookv1alpha1.HookPhaseSuccessful {
		err := deletePostInplaceHookCondition(newStatus, pod.Name)
		if err != nil {
			klog.Warningf("expected the %s %s/%s exists a PostInplaceHookCondition for pod %s, but got an error: %s",
				objectKind, namespace, name, pod.Name, err.Error())
		}
		err = p.deleteHookRun(existHookRuns[0])
		if err != nil {
			klog.Warningf("delete the post inplace hook %s/%s for pod %s, but got an error: %s",
				existHookRuns[0].Namespace, existHookRuns[0].Name, pod.Name, err.Error())
		}
		return nil
	}

	err = resetPostInplaceHookConditionPhase(newStatus, pod.Name, existHookRuns[0].Status.Phase)
	if err != nil {
		klog.Warningf("expected the %s %s/%s exists a PostInplaceHookCondition for pod %s, but got an error: %s",
			objectKind, namespace, name, pod.Name, err.Error())
	}

	return nil
}

// createHookRun create a PostInplace HookRun
func (p *PostInplaceControl) createHookRun(metaObj metav1.Object, runtimeObj runtime.Object,
	postInplaceHook *hookv1alpha1.HookStep, pod *v1.Pod, labels map[string]string,
	podNameLabelKey string) (*hookv1alpha1.HookRun, error) {
	arguments := []hookv1alpha1.Argument{}
	for _, arg := range postInplaceHook.Args {
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
		imageArgs := []hookv1alpha1.Argument{
			{
				Name:  PodImageArgKey + "[" + strconv.Itoa(i) + "]",
				Value: tmp,
			},
		}
		arguments = append(arguments, imageArgs...)
	}

	hr, err := p.newHookRunFromHookTemplate(metaObj, runtimeObj, arguments, pod, postInplaceHook, labels, podNameLabelKey)
	if err != nil {
		return nil, err
	}
	hookRunIf := p.hookClient.TkexV1alpha1().HookRuns(pod.Namespace)
	return commonhookutil.CreateWithCollisionCounter(hookRunIf, *hr)
}

// newHookRunFromGameStatefulSet generate a HookRun from HookTemplate
func (p *PostInplaceControl) newHookRunFromHookTemplate(metaObj metav1.Object,
	runtimeObj runtime.Object, args []hookv1alpha1.Argument,
	pod *v1.Pod, postInplaceHook *hookv1alpha1.HookStep, labels map[string]string,
	podNameLabelKey string) (*hookv1alpha1.HookRun, error) {
	template, err := p.hookTemplateLister.HookTemplates(pod.Namespace).Get(postInplaceHook.TemplateName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			klog.Warningf("HookTemplate '%s' not found for %s/%s",
				postInplaceHook.TemplateName, pod.Namespace, pod.Name)
		}
		return nil, err
	}

	nameParts := []string{"postinplace", pod.Labels[apps.ControllerRevisionHashLabelKey],
		pod.Labels[podNameLabelKey], postInplaceHook.TemplateName}
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

// delete PostInplaceHookCondition of a pod
func deletePostInplaceHookCondition(status PostInplaceHookStatusInterface, podName string) error {
	var index int
	found := false
	conditions := status.GetPostInplaceHookConditions()
	for i, cond := range conditions {
		if cond.PodName == podName {
			found = true
			index = i
			break
		}
	}
	if !found {
		return fmt.Errorf("no PostInplaceHookCondition to delete")
	}

	newConditions := append(conditions[:index], conditions[index+1:]...)
	status.SetPostInplaceHookConditions(newConditions)
	return nil
}

func updatePostInplaceHookCondition(status PostInplaceHookStatusInterface, podName string) {
	conditions := status.GetPostInplaceHookConditions()
	for i, cond := range conditions {
		if cond.PodName == podName {
			conditions[i].HookPhase = hookv1alpha1.HookPhasePending
			conditions[i].StartTime = metav1.Now()
			status.SetPostInplaceHookConditions(conditions)
			return
		}
	}
	conditions = append(conditions, hookv1alpha1.PostInplaceHookCondition{
		PodName:   podName,
		StartTime: metav1.Now(),
		HookPhase: hookv1alpha1.HookPhasePending,
	})
	status.SetPostInplaceHookConditions(conditions)
}

// reset PostInplaceHookConditionPhase of a pod
func resetPostInplaceHookConditionPhase(status PostInplaceHookStatusInterface, podName string,
	phase hookv1alpha1.HookPhase) error {
	var index int
	found := false
	conditions := status.GetPostInplaceHookConditions()
	for i, cond := range conditions {
		if cond.PodName == podName {
			found = true
			index = i
			break
		}
	}
	if !found {
		return fmt.Errorf("no PostInplaceHookCondition to reset phase")
	}

	cond := hookv1alpha1.PostInplaceHookCondition{
		PodName:   podName,
		StartTime: conditions[index].StartTime,
		HookPhase: phase,
	}
	conditions = append(conditions, cond)
	conditions = append(conditions[:index], conditions[index+1:]...)
	status.SetPostInplaceHookConditions(conditions)
	return nil
}

func (p *PostInplaceControl) deleteHookRun(hr *hookv1alpha1.HookRun) error {
	if hr.DeletionTimestamp != nil {
		return nil
	}
	err := p.hookClient.TkexV1alpha1().HookRuns(hr.Namespace).Delete(context.TODO(), hr.Name, metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return err
	}
	return nil
}
