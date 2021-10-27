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

package scale

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"

	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	canaryutil "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-gamedeployment-operator/pkg/util/canary"

	v1 "k8s.io/api/core/v1"
	intstrutil "k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog"
	kubecontroller "k8s.io/kubernetes/pkg/controller"
	"k8s.io/utils/integer"
)

func getPodsToDelete(deploy *gdv1alpha1.GameDeployment, pods []*v1.Pod) []*v1.Pod {
	var podsToDelete []*v1.Pod
	s := sets.NewString(deploy.Spec.ScaleStrategy.PodsToDelete...)
	for _, p := range pods {
		if s.Has(p.Name) {
			podsToDelete = append(podsToDelete, p)
		}
	}
	return podsToDelete
}

// Generate available IDs, keep all ids different from existing ids
func genAvailableIDs(num int, pods []*v1.Pod) sets.String {
	existingIDs := sets.NewString()

	for _, pod := range pods {
		if id := pod.Labels[gdv1alpha1.GameDeploymentInstanceID]; len(id) > 0 {
			existingIDs.Insert(id)
		}
	}

	retIDs := sets.NewString()
	for i := 0; i < num; i++ {
		id := genInstanceID(existingIDs)
		retIDs.Insert(id)
	}

	return retIDs
}

func genInstanceID(existingIDs sets.String) string {
	var id string
	for {
		id = rand.String(LengthOfInstanceID)
		if !existingIDs.Has(id) {
			break
		}
	}
	return id
}

func calculateDiffs(deploy *gdv1alpha1.GameDeployment, revConsistent bool, totalPods int, notUpdatedPods int) (totalDiff int, currentRevDiff int) {
	var maxSurge int

	if !revConsistent {
		currentPartition := canaryutil.GetCurrentPartition(deploy)
		if currentPartition != 0 {
			currentRevDiff = notUpdatedPods - integer.IntMin(int(currentPartition), int(*deploy.Spec.Replicas))
		}
		//if deploy.Spec.UpdateStrategy.Partition != nil {
		//	currentRevDiff = notUpdatedPods - integer.IntMin(int(*deploy.Spec.UpdateStrategy.Partition), int(*deploy.Spec.Replicas))
		//}

		// Use maxSurge only if partition has not satisfied
		if currentRevDiff > 0 {
			if deploy.Spec.UpdateStrategy.MaxSurge != nil {
				maxSurge, _ = intstrutil.GetValueFromIntOrPercent(deploy.Spec.UpdateStrategy.MaxSurge, int(*deploy.Spec.Replicas), true)
				maxSurge = integer.IntMin(maxSurge, currentRevDiff)
			}
		}
	}
	totalDiff = totalPods - int(*deploy.Spec.Replicas) - maxSurge

	if totalDiff != 0 && maxSurge > 0 {
		klog.V(3).Infof("GameDeployment scale diff(%d),currentRevDiff(%d) with maxSurge %d", totalDiff, currentRevDiff, maxSurge)
	}
	return
}

func choosePodsToDelete(totalDiff int, currentRevDiff int, notUpdatedPods, updatedPods []*v1.Pod, sortMethod string) []*v1.Pod {
	choose := func(pods []*v1.Pod, diff int) []*v1.Pod {
		// No need to sort pods if we are about to delete all of them.
		if diff < len(pods) {
			// Sort the pods in the order such that not-ready < ready, unscheduled
			// < scheduled, and pending < running. This ensures that we delete pods
			// in the earlier stages whenever possible.
			//TODO (by bryanhe) consider some pods maybe crashed or status changed, then the pods order to be PreDeleteHook maybe
			// change, maybe we should use a simple alphabetical sort
			sort.Sort(kubecontroller.ActivePods(pods))
			// sort the pods with deletion cost
			switch sortMethod {
			case CostSortMethodDescend:
				sort.Slice(pods, func(i, j int) bool {
					costA := getDeletionCostFromPodAnnotations(pods[i], sortMethod)
					costB := getDeletionCostFromPodAnnotations(pods[j], sortMethod)
					return costA > costB
				})
			default:
				sort.Slice(pods, func(i, j int) bool {
					costA := getDeletionCostFromPodAnnotations(pods[i], sortMethod)
					costB := getDeletionCostFromPodAnnotations(pods[j], sortMethod)
					return costA < costB
				})
			}

		} else if diff > len(pods) {
			klog.Warningf("Diff > len(pods) in choosePodsToDelete func which is not expected.")
			return pods
		}
		return pods[:diff]
	}

	var podsToDelete []*v1.Pod
	if currentRevDiff >= totalDiff {
		podsToDelete = choose(notUpdatedPods, totalDiff)
	} else if currentRevDiff > 0 {
		podsToDelete = choose(notUpdatedPods, currentRevDiff)
		podsToDelete = append(podsToDelete, choose(updatedPods, totalDiff-currentRevDiff)...)
	} else {
		podsToDelete = choose(updatedPods, totalDiff)
	}

	return podsToDelete
}

func getDeletionCostSortMethod(deploy *gdv1alpha1.GameDeployment) string {
	method := deploy.Annotations[DeletionCostSortMethod]
	if len(method) == 0 {
		method = CostSortMethodAscend
	}
	return method
}

func getDeletionCostFromPodAnnotations(pod *v1.Pod, method string) float64 {
	var edgeCase float64
	switch method {
	case CostSortMethodDescend:
		edgeCase = -math.MaxFloat64
	default:
		edgeCase = math.MaxFloat64
	}
	costAnnotation := pod.Annotations[DeletionCost]
	if len(costAnnotation) == 0 {
		return edgeCase
	}
	costValue, err := strconv.ParseFloat(costAnnotation, 64)
	if err != nil {
		return edgeCase
	}
	return costValue
}

func validateGameDeploymentPodIndex(deploy *gdv1alpha1.GameDeployment) (bool, int, int, error) {
	if deploy == nil {
		return false, 0, 0, nil
	}

	ok := gameDeploymentIndexFeature(deploy)
	if !ok {
		return false, 0, 0, nil
	}

	start, end, err := getDeploymentIndexRange(deploy)
	if err != nil {
		return false, 0, 0, err
	}

	if start < 0 || end < 0 || start >= end {
		return false, 0, 0, fmt.Errorf("gamedeployment %s invalid index range", deploy.Name)
	}

	if *deploy.Spec.Replicas > int32(end-start) {
		return false, 0, 0, fmt.Errorf("deploy %s scale replicas gt available indexs", deploy.GetName())
	}

	return true, start, end, nil
}

func gameDeploymentIndexFeature(deploy *gdv1alpha1.GameDeployment) bool {
	value, ok := deploy.Annotations[gdv1alpha1.GameDeploymentIndexOn]
	if ok && value == "true" {
		return true
	}

	return false
}

func getDeploymentIndexRange(deploy *gdv1alpha1.GameDeployment) (int, int, error) {
	indexRange := &gdv1alpha1.GameDeploymentPodIndexRange{}

	value, ok := deploy.Annotations[gdv1alpha1.GameDeploymentIndexRange]
	if ok {
		err := json.Unmarshal([]byte(value), indexRange)
		if err != nil {
			return 0, 0, err
		}

		return indexRange.PodStartIndex, indexRange.PodEndIndex, nil
	}

	return 0, 0, fmt.Errorf("gamedeployment %s inject index on, get index-range failed", deploy.Name)
}

// Generate available index IDs, keep it unique
func genAvailableIndex(inject bool, start, end int, pods []*v1.Pod) []int {
	needIDs := make([]int, 0)
	if !inject {
		return needIDs
	}

	existIDs := getExistPodsIndex(pods)
	for i := start; i < end; i++ {
		_, ok := existIDs[i]
		if !ok {
			needIDs = append(needIDs, i)
		}
	}

	sort.Ints(needIDs)
	return needIDs
}

func getExistPodsIndex(pods []*v1.Pod) map[int]struct{} {
	idIndex := make([]string, 0)
	for _, pod := range pods {
		if id := pod.Annotations[gdv1alpha1.GameDeploymentIndexID]; len(id) > 0 {
			idIndex = append(idIndex, id)
		}
	}

	existIDs := make(map[int]struct{}, 0)
	for _, id := range idIndex {
		n, err := strconv.Atoi(id)
		if err == nil {
			existIDs[n] = struct{}{}
		}
	}

	return existIDs
}
