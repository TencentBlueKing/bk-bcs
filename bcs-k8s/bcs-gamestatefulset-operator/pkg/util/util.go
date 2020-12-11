/*
 * Tencent is pleased to support the open source community by making TKE
 * available.
 *
 * Copyright (C) 2018 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License"); you may not use this
 * file except in compliance with the License. You may obtain a copy of the
 * License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
 * WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
 * License for the specific language governing permissions and limitations under
 * the License.
 */

package util

import (
	gstsv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamestatefulset-operator/pkg/apis/tkex/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

var (
	// controllerKind contains the schema.GroupVersionKind for this controller type.
	ControllerKind = gstsv1alpha1.SchemeGroupVersion.WithKind("GameStatefulSet")
)

// UpdatePodField Fields that can be modified
// - spec.containers[*].imag
func UpdatePodField(updateSet *gstsv1alpha1.GameStatefulSet, updateRevision string, pod *v1.Pod) *v1.Pod {

	// make a deep copy, do not mutate the shared cache
	newPod := pod.DeepCopy()

	// find the container, then update it's image
	for _, updatedContainer := range updateSet.Spec.Template.Spec.Containers {
		for j, newContainer := range newPod.Spec.Containers {
			if updatedContainer.Name == newContainer.Name && updatedContainer.Image != newContainer.Image {
				newPod.Spec.Containers[j].Image = updatedContainer.Image
			}
		}
	}

	// update Pod revision label
	if newPod.Labels == nil {
		newPod.Labels = make(map[string]string)
	}
	newPod.Labels[gstsv1alpha1.GameStatefulSetRevisionLabel] = updateRevision

	return newPod
}
