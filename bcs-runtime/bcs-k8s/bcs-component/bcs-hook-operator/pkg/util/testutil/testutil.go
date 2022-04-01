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
	hookv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/bcs-hook/apis/tkex/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NewHookRun creates a new HookRun object.
func NewHookRun(name string) *hookv1alpha1.HookRun {
	return &hookv1alpha1.HookRun{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   corev1.NamespaceDefault,
			Name:        name,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Spec: hookv1alpha1.HookRunSpec{Metrics: []hookv1alpha1.Metric{}},
	}
}
