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

package test

import (
	"fmt"
	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	commonhookutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/util/hook"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientTesting "k8s.io/client-go/testing"
	"k8s.io/kubernetes/pkg/controller/history"
	"reflect"
)

// NewGameDeployment for unit tests.
func NewGameDeployment(replicas int) *gdv1alpha1.GameDeployment {
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

	return &gdv1alpha1.GameDeployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GameDeployment",
			APIVersion: "tkex.tencent.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: v1.NamespaceDefault,
			UID:       "test",
		},
		Spec: gdv1alpha1.GameDeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"foo": "bar"},
			},
			Replicas:       func() *int32 { i := int32(replicas); return &i }(),
			Template:       template,
			UpdateStrategy: gdv1alpha1.GameDeploymentUpdateStrategy{Type: gdv1alpha1.InPlaceGameDeploymentUpdateStrategyType},
			RevisionHistoryLimit: func() *int32 {
				limit := int32(2)
				return &limit
			}(),
		},
	}
}

// NewPod create a new pod for unit tests.
func NewPod(suffix interface{}) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        fmt.Sprintf("foo-%v", suffix),
			Namespace:   v1.NamespaceDefault,
			Labels:      map[string]string{},
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

// NewHookTemplate for unit tests.
func NewHookTemplate() *v1alpha1.HookTemplate {
	return &v1alpha1.HookTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: v1.NamespaceDefault,
		},
		Spec: v1alpha1.HookTemplateSpec{
			Metrics: []v1alpha1.Metric{
				{
					Name: "foo",
				},
			},
		},
	}
}

// NewHookRunFromTemplate create a new hook run from template for unit tests.
func NewHookRunFromTemplate(hookTemplate *v1alpha1.HookTemplate, deploy *gdv1alpha1.GameDeployment) *v1alpha1.HookRun {
	run, _ := commonhookutil.NewHookRunFromTemplate(hookTemplate, nil,
		fmt.Sprintf("predelete---%s", hookTemplate.Name), "", hookTemplate.Namespace)
	run.Labels = map[string]string{
		"hookrun-type":      "pre-delete-step",
		"instance-id":       "",
		"workload-revision": "",
	}
	run.OwnerReferences = []metav1.OwnerReference{*metav1.NewControllerRef(deploy, deploy.GetObjectKind().GroupVersionKind())}
	return run
}

// NewGDControllerRevision create a new gd controller revision for unit tests.
func NewGDControllerRevision(deploy *gdv1alpha1.GameDeployment, revision int64) *v12.ControllerRevision {
	cr, _ := history.NewControllerRevision(deploy,
		schema.GroupVersionKind{Group: v1alpha1.GroupName, Version: v1alpha1.Version, Kind: v1alpha1.Kind},
		nil, runtime.RawExtension{}, revision, func() *int32 { a := int32(1); return &a }())
	return cr
}

// EqualActions check if two actions are equal.
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

// CompareAction compare two action that have same object
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

// FilterActions filter actions by filterFns.
func FilterActions(actions []clientTesting.Action, filterFns ...func(action clientTesting.Action) clientTesting.Action) []clientTesting.Action {
	for i := range actions {
		for _, fn := range filterFns {
			actions[i] = fn(actions[i])
		}
	}
	return actions
}

// FilterCreateAction filter create actions.
func FilterCreateAction(action clientTesting.Action) clientTesting.Action {
	if a, ok := action.(clientTesting.CreateActionImpl); ok {
		return clientTesting.NewCreateAction(a.GetResource(), a.GetNamespace(), nil)
	}
	return action
}

// FilterUpdateAction filter update actions.
func FilterUpdateAction(action clientTesting.Action) clientTesting.Action {
	if a, ok := action.(clientTesting.UpdateActionImpl); ok {
		return clientTesting.NewUpdateAction(a.GetResource(), a.GetNamespace(), nil)
	}
	return action
}

// FilterPatchAction filter patch actions.
func FilterPatchAction(action clientTesting.Action) clientTesting.Action {
	if a, ok := action.(clientTesting.PatchActionImpl); ok {
		return clientTesting.NewPatchAction(a.GetResource(), a.GetNamespace(), a.Name, a.PatchType, nil)
	}
	return action
}
