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

package core

import (
	"reflect"

	apiv1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/autoscaler/cluster-autoscaler/processors/status"
	"k8s.io/autoscaler/cluster-autoscaler/utils/drain"
)

type podEquivalenceGroup struct {
	pods             []*apiv1.Pod
	schedulingErrors map[string]status.Reasons
	schedulable      bool
}

// buildPodEquivalenceGroups prepares pod groups with equivalent scheduling properties.
func buildPodEquivalenceGroups(pods []*apiv1.Pod) []*podEquivalenceGroup {
	podEquivalenceGroups := []*podEquivalenceGroup{}
	for _, pods := range groupPodsBySchedulingProperties(pods) {
		podEquivalenceGroups = append(podEquivalenceGroups, &podEquivalenceGroup{
			pods:             pods,
			schedulingErrors: map[string]status.Reasons{},
			schedulable:      false,
		})
	}
	return podEquivalenceGroups
}

type equivalenceGroupId int
type equivalenceGroup struct {
	id           equivalenceGroupId
	representant *apiv1.Pod
}

// groupPodsBySchedulingProperties groups pods based on scheduling properties. Group ID is meaningless.
func groupPodsBySchedulingProperties(pods []*apiv1.Pod) map[equivalenceGroupId][]*apiv1.Pod {
	podEquivalenceGroups := map[equivalenceGroupId][]*apiv1.Pod{}
	equivalenceGroupsByController := make(map[types.UID][]equivalenceGroup)

	var nextGroupId equivalenceGroupId
	for _, pod := range pods {
		controllerRef := drain.ControllerRef(pod)
		if controllerRef == nil {
			podEquivalenceGroups[nextGroupId] = []*apiv1.Pod{pod}
			nextGroupId++
			continue
		}

		matchingFound := false
		for _, g := range equivalenceGroupsByController[controllerRef.UID] {
			if reflect.DeepEqual(pod.Labels, g.representant.Labels) &&
				apiequality.Semantic.DeepEqual(pod.Spec, g.representant.Spec) {
				matchingFound = true
				podEquivalenceGroups[g.id] = append(podEquivalenceGroups[g.id], pod)
				break
			}
		}

		if !matchingFound {
			newGroup := equivalenceGroup{
				id:           nextGroupId,
				representant: pod,
			}
			equivalenceGroupsByController[controllerRef.UID] = append(equivalenceGroupsByController[controllerRef.UID], newGroup)
			podEquivalenceGroups[newGroup.id] = append(podEquivalenceGroups[newGroup.id], pod)
			nextGroupId++
		}
	}

	return podEquivalenceGroups
}
