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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clientTesting "k8s.io/client-go/testing"
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
			UID:       types.UID("test"),
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

func NewPod(suffix interface{}) *v1.Pod {
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("foo-%v", suffix),
			Namespace: v1.NamespaceDefault,
			Labels:    map[string]string{},
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

func EqualActions(x, y []clientTesting.Action) bool {
	if len(x) == 0 && len(y) == 0 {
		return true
	}
	return reflect.DeepEqual(x, y)
}
