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

// Package hotpatchupdate xxx
package hotpatchupdate

import (
	"encoding/json"
	"fmt"

	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/common/update"
)

// Interface for managing pods in-place update.
type Interface interface {
	Update(pod *v1.Pod, oldRevision, newRevision *apps.ControllerRevision) error
}

type realControl struct {
	adp         update.Adapter
	revisionKey string

	now func() metav1.Time
}

// NewForTypedClient xxx
func NewForTypedClient(c clientset.Interface, revisionKey string) Interface {
	return &realControl{adp: &update.AdapterTypedClient{Client: c}, revisionKey: revisionKey, now: metav1.Now}
}

func (c *realControl) Update(pod *v1.Pod, oldRevision, newRevision *apps.ControllerRevision) error {
	// 1. calculate inplace update spec
	spec := update.CalculateInPlaceUpdateSpec(oldRevision, newRevision)

	if spec == nil {
		return fmt.Errorf("find Pod %s update strategy is HotPatch,"+
			"but the diff not only contains replace operation of spec.containers[x].image", pod)
	}

	return c.updatePodHotPatch(pod, spec)
}

func (c *realControl) updatePodHotPatch(pod *v1.Pod, spec *update.UpdateSpec) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		clone, err := c.adp.GetPod(pod.Namespace, pod.Name)
		if err != nil {
			return err
		}

		// update new revision
		if c.revisionKey != "" {
			clone.Labels[c.revisionKey] = spec.Revision
		}
		if clone.Annotations == nil {
			clone.Annotations = map[string]string{}
		}

		// record old containerStatuses
		hotPatchUpdateState := HotPatchUpdateState{
			Revision:              spec.Revision,
			UpdateTimestamp:       c.now(),
			LastContainerStatuses: make(map[string]HotPatchUpdateContainerStatus, len(spec.ContainerImages)),
		}
		for _, c := range clone.Status.ContainerStatuses {
			if _, ok := spec.ContainerImages[c.Name]; ok {
				hotPatchUpdateState.LastContainerStatuses[c.Name] = HotPatchUpdateContainerStatus{
					ImageID: c.ImageID,
				}
			}
		}
		hotPatchUpdateStateJSON, _ := json.Marshal(hotPatchUpdateState)
		clone.Annotations[HotPatchUpdateStateKey] = string(hotPatchUpdateStateJSON)

		if clone, err = update.PatchUpdateSpecToPod(clone, spec); err != nil {
			return err
		}

		_, ok := clone.Annotations[PodHotpatchContainerKey]
		if !ok {
			clone.Annotations[PodHotpatchContainerKey] = "true"
		}

		return c.adp.UpdatePod(clone)
	})
}
