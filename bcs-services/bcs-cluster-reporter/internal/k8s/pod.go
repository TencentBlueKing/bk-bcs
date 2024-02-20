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

package k8s

import (
	"context"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetPods get namespace pods
func GetPods(clientSet *kubernetes.Clientset, namespace string, opts v1.ListOptions, nameRe string) ([]corev1.Pod,
	error) {
	ctx := context.Background()
	podList, err := clientSet.CoreV1().Pods(namespace).List(ctx, opts)
	if err != nil {
		return nil, err
	}

	if nameRe == "" {
		return podList.Items, nil
	}
	re, _ := regexp.Compile(nameRe)
	result := make([]corev1.Pod, 0, 0)
	for _, pod := range podList.Items {
		if re.MatchString(pod.Name) || strings.Contains(pod.Name, nameRe) {
			result = append(result, pod)
		}
	}
	return result, nil
}
