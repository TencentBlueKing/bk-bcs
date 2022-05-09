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
 */

package testutil

import (
	"fmt"
	gstsv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/hook"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientTesting "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/controller/history"
	"reflect"
)

// NewGameStatefulSet creates a new GameStatefulSet object.
func NewGameStatefulSet(replicas int) *gstsv1alpha1.GameStatefulSet {
	name := "foo"
	template := v1.PodTemplateSpec{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "nginx",
					Image: "nginx",
				},
			},
		},
	}
	template.Labels = map[string]string{"foo": "bar"}

	return &gstsv1alpha1.GameStatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GameStatefulSet",
			APIVersion: "tkex.tencent.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: v1.NamespaceDefault,
			UID:       "test",
		},
		Spec: gstsv1alpha1.GameStatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"foo": "bar"},
			},
			Replicas: func() *int32 { i := int32(replicas); return &i }(),
			Template: template,
			UpdateStrategy: gstsv1alpha1.GameStatefulSetUpdateStrategy{
				Type: gstsv1alpha1.InplaceUpdateGameStatefulSetStrategyType,
			},
			RevisionHistoryLimit: func() *int32 {
				limit := int32(2)
				return &limit
			}(),
		},
	}
}

// NewPod creates a new Pod object.
func NewPod(suffix interface{}) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:         fmt.Sprintf("foo-%v", suffix),
			GenerateName: "foo-",
			Namespace:    v1.NamespaceDefault,
			Labels: map[string]string{
				"foo":                                    "bar",
				gstsv1alpha1.GameStatefulSetPodNameLabel: fmt.Sprintf("foo-%v", suffix),
				gstsv1alpha1.GameStatefulSetPodOrdinal:   fmt.Sprintf("%v", suffix),
			},
			Annotations: map[string]string{},
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "nginx",
					Image: "nginx",
				},
			},
		},
	}
}

// NewHookTemplate creates a new HookTemplate object.
func NewHookTemplate() *hookv1alpha1.HookTemplate {
	return &hookv1alpha1.HookTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: v1.NamespaceDefault,
		},
		Spec: hookv1alpha1.HookTemplateSpec{
			Metrics: []hookv1alpha1.Metric{
				{
					Name: "foo",
				},
			},
		},
	}
}

// NewHookRunFromTemplate creates a new HookRun object.
func NewHookRunFromTemplate(hookTemplate *hookv1alpha1.HookTemplate, sts *gstsv1alpha1.GameStatefulSet) *hookv1alpha1.HookRun {
	run, _ := commonhookutil.NewHookRunFromTemplate(hookTemplate, nil,
		fmt.Sprintf("predelete---%s", hookTemplate.Name), "", hookTemplate.Namespace)
	run.Labels = map[string]string{
		"hookrun-type":      "pre-delete-step",
		"instance-id":       "",
		"workload-revision": "",
	}
	run.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(sts, sts.GetObjectKind().GroupVersionKind())}
	return run
}

// NewControllerRevision creates a new ControllerRevision object.
func NewControllerRevision(sts *gstsv1alpha1.GameStatefulSet, revision int64) *appsv1.ControllerRevision {
	cr, _ := history.NewControllerRevision(sts,
		schema.GroupVersionKind{Group: gstsv1alpha1.GroupName, Version: gstsv1alpha1.Version, Kind: gstsv1alpha1.Kind},
		nil, runtime.RawExtension{}, revision, func() *int32 { a := int32(1); return &a }())
	cr.Namespace = sts.GetNamespace()
	return cr
}

// EqualActions compares two sets of actions.
func EqualActions(x, y []clientTesting.Action) bool {
	if len(x) != len(y) {
		return false
	}

	for i := range x {
		if !CompareAction(x[i], y[i]) {
			return false
		}
	}
	return true
}

// CompareAction compares two actions.
func CompareAction(x, y clientTesting.Action) bool {
	// for create action
	a, ok := x.(clientTesting.CreateActionImpl)
	if !ok {
		return reflect.DeepEqual(x, y)
	}
	b, ok := y.(clientTesting.CreateActionImpl)
	if !ok {
		return reflect.DeepEqual(x, y)
	}
	poda, ok := a.GetObject().(*v1.Pod)
	if !ok {
		return reflect.DeepEqual(x, y)
	}
	podb, ok := b.GetObject().(*v1.Pod)
	if !ok {
		return reflect.DeepEqual(x, y)
	}
	return poda.String() == podb.String()
}

// FilterActions filters actions by the specified verb.
func FilterActions(actions []clientTesting.Action, filterFns ...func(action clientTesting.Action) clientTesting.Action) []clientTesting.Action {
	for i := range actions {
		for _, fn := range filterFns {
			actions[i] = fn(actions[i])
		}
	}
	return actions
}

// FilterCreateAction filters create actions.
func FilterCreateAction(action clientTesting.Action) clientTesting.Action {
	if a, ok := action.(clientTesting.CreateActionImpl); ok {
		return clientTesting.NewCreateAction(a.GetResource(), a.GetNamespace(), nil)
	}
	return action
}

// FilterUpdateAction filters update actions.
func FilterUpdateAction(action clientTesting.Action) clientTesting.Action {
	if a, ok := action.(clientTesting.UpdateActionImpl); ok {
		return clientTesting.NewUpdateAction(a.GetResource(), a.GetNamespace(), nil)
	}
	return action
}

// FilterPatchAction filters patch actions.
func FilterPatchAction(action clientTesting.Action) clientTesting.Action {
	if a, ok := action.(clientTesting.PatchActionImpl); ok {
		return clientTesting.NewPatchAction(a.GetResource(), a.GetNamespace(), a.Name, a.PatchType, nil)
	}
	return action
}

// FilterOwnerRefer filters owner refer actions.
func FilterOwnerRefer(action clientTesting.Action) clientTesting.Action {
	if a, ok := action.(clientTesting.CreateActionImpl); ok {
		if pod, ok := a.GetObject().(*v1.Pod); ok {
			pod.OwnerReferences = nil
			return clientTesting.NewCreateAction(a.GetResource(), a.GetNamespace(), pod)
		}
		return a
	}
	return action
}

// FilterGameStatefulSetStatusTime filters GameStatefulSet status time.
func FilterGameStatefulSetStatusTime(status *gstsv1alpha1.GameStatefulSetStatus) {
	if status == nil {
		return
	}
	for i := range status.PreInplaceHookConditions {
		status.PreInplaceHookConditions[i].StartTime = metav1.Time{}
	}
	for i := range status.PreInplaceHookConditions {
		status.PreInplaceHookConditions[i].StartTime = metav1.Time{}
	}
	for i := range status.PostInplaceHookConditions {
		status.PostInplaceHookConditions[i].StartTime = metav1.Time{}
	}
}
