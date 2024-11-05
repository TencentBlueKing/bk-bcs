/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package systemappcheck xxx
package systemappcheck

import (
	"fmt"
	"testing"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

func TestGetResourceGaugeVecSet(t *testing.T) {
	obj := &v1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind: "deploy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-deployment",
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "my-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "my-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "my-container",
							Image: "my-image",
						},
					},
				},
			},
		},
	}

	robj := obj.DeepCopyObject()
	objMap, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(robj)
	fmt.Println(objMap)

	unstr := unstructured.Unstructured{Object: objMap}

	unstr.Object = objMap
	unstr.SetUnstructuredContent(objMap)

	fmt.Println(unstr)
	fmt.Println(unstr.GetName())
	fmt.Println(unstr.GetObjectKind().GroupVersionKind().Kind)
	fmt.Println(unstr.GetObjectKind())
}
