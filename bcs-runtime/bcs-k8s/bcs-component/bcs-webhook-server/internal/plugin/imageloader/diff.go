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

package imageloader

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func diff(oldMeta, newMeta metav1.ObjectMeta,
	oldContainers, newContainers []corev1.Container) (string, []corev1.Container) {
	revertMeta, updateMeta := metaDiff(oldMeta, newMeta)
	revertCon, updateCon, diffContainers := imageDiff(oldContainers, newContainers)

	updatePatch := []string{strings.Join(updateCon, ",")}
	if len(updateMeta) > 0 {
		updatePatch = append(updatePatch, strings.Join(updateMeta, ","))
	}
	updatePatchStr := fmt.Sprintf("[%s]", strings.Join(updatePatch, ","))
	// set patch to annotations and append annotions patch
	revertPatch := append(revertMeta, revertCon...) // nolint
	revertPatch = append(revertPatch, fmt.Sprintf(
		"{\"op\":\"add\",\"path\":\"/metadata/annotations/%s\",\"value\":\"%s\"}",
		imageUpdateAnno, strings.ReplaceAll(updatePatchStr, "\"", "\\\"")))

	// combine patch string
	patchStr := fmt.Sprintf("[%s]", strings.Join(revertPatch, ","))

	return patchStr, diffContainers
}

func metaDiff(oldMeta, newMeta metav1.ObjectMeta) ([]string, []string) {
	revertLabelPatch, updateLabelPatch := labelDiff(oldMeta, newMeta)
	revertAnnoPatch, updateAnnoPatch := annoDiff(oldMeta, newMeta)

	revertPatch := append(revertLabelPatch, revertAnnoPatch...) // nolint

	updatePatch := append(updateLabelPatch, updateAnnoPatch...) // nolint
	return revertPatch, updatePatch
}

func labelDiff(oldMeta, newMeta metav1.ObjectMeta) ([]string, []string) {
	revertPatch := make([]string, 0, len(newMeta.Labels))
	updatePatch := make([]string, 0, len(newMeta.Labels))
	for key, newValue := range newMeta.Labels {
		if oldValue, ok := oldMeta.Labels[key]; ok && newValue != oldValue {
			revertPatch = append(revertPatch, fmt.Sprintf(
				"{\"op\":\"replace\",\"path\":\"/metadata/labels/%s\",\"value\":\"%s\"}", key, oldValue))
			updatePatch = append(updatePatch, fmt.Sprintf(
				"{\"op\":\"replace\",\"path\":\"/metadata/labels/%s\",\"value\":\"%s\"}", key, newValue))
		} else if !ok {
			revertPatch = append(revertPatch, fmt.Sprintf(
				"{\"op\":\"remove\",\"path\":\"/metadata/labels/%s\"}", key))
			updatePatch = append(updatePatch, fmt.Sprintf(
				"{\"op\":\"add\",\"path\":\"/metadata/labels/%s\",\"value\":\"%s\"}", key, newValue))
		}
	}
	return revertPatch, updatePatch
}

func annoDiff(oldMeta, newMeta metav1.ObjectMeta) ([]string, []string) {
	revertPatch := make([]string, 0, len(newMeta.Annotations))
	updatePatch := make([]string, 0, len(newMeta.Annotations))
	for key, newValue := range newMeta.Annotations {
		if oldValue, ok := oldMeta.Annotations[key]; ok && newValue != oldValue {
			revertPatch = append(revertPatch, fmt.Sprintf(
				"{\"op\":\"replace\",\"path\":\"/metadata/annotations/%s\",\"value\":\"%s\"}", key, oldValue))
			updatePatch = append(updatePatch, fmt.Sprintf(
				"{\"op\":\"replace\",\"path\":\"/metadata/annotations/%s\",\"value\":\"%s\"}", key, newValue))
		} else if !ok {
			revertPatch = append(revertPatch, fmt.Sprintf(
				"{\"op\":\"remove\",\"path\":\"/metadata/annotations/%s\"}", key))
			updatePatch = append(updatePatch, fmt.Sprintf(
				"{\"op\":\"add\",\"path\":\"/metadata/annotations/%s\",\"value\":\"%s\"}", key, newValue))
		}
	}
	return revertPatch, updatePatch
}

func imageDiff(oldContainers, newContainers []corev1.Container) ([]string, []string, []corev1.Container) {
	// for quick index
	oldImageMap := make(map[string]string)
	for _, c := range oldContainers {
		oldImageMap[c.Name] = c.Image
	}
	// container image update and update patch
	revertPatch := make([]string, 0, len(newContainers)+1)
	// patch to annotations, used for trigger real update
	updatePatch := make([]string, 0, len(newContainers))
	retContainers := make([]corev1.Container, 0, len(newContainers))
	for i, c := range newContainers {
		if oi, ok := oldImageMap[c.Name]; ok && oi != c.Image {
			// generate mutate patch
			revertPatch = append(revertPatch, fmt.Sprintf(
				"{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/%d/image\",\"value\":\"%s\"}", i, oi))
			updatePatch = append(updatePatch, fmt.Sprintf(
				"{\"op\":\"replace\",\"path\":\"/spec/template/spec/containers/%d/image\",\"value\":\"%s\"}", i, c.Image))
			// add a image loader container
			retContainers = append(retContainers,
				corev1.Container{
					Name:            c.Name,
					Image:           c.Image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         []string{"echo", "pull " + c.Image}}) // nolint
		}
	}
	return revertPatch, updatePatch, retContainers
}
