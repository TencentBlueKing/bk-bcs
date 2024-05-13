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

// Package util xxx
package util

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// GetPodRevision returns revision hash of this pod.
func GetPodRevision(pod metav1.Object) string {
	return pod.GetLabels()[appsv1.ControllerRevisionHashLabelKey]
}

// GetPodsRevisions return revision hash set of these pods.
func GetPodsRevisions(pods []*v1.Pod) sets.String {
	revisions := sets.NewString()
	for _, p := range pods {
		revisions.Insert(GetPodRevision(p))
	}
	return revisions
}
