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

package util

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	gdv1alpha1 "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/apis/tkex/v1alpha1"
	gdlister "github.com/Tencent/bk-bcs/bcs-k8s/bcs-gamedeployment-operator/pkg/client/listers/tkex/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	podutil "k8s.io/kubernetes/pkg/api/v1/pod"
	"k8s.io/utils/integer"
)

// ControllerKind contains the schema.GroupVersionKind for this controller type.
var ControllerKind = gdv1alpha1.SchemeGroupVersion.WithKind("GameDeployment")

// GetControllerKey return key of GameDeployment.
func GetControllerKey(cs *gdv1alpha1.GameDeployment) string {
	return types.NamespacedName{Namespace: cs.Namespace, Name: cs.Name}.String()
}

// GetPodRevision returns revision hash of this pod.
func GetPodRevision(pod metav1.Object) string {
	return pod.GetLabels()[appsv1.ControllerRevisionHashLabelKey]
}

// IsRunningAndAvailable returns true if pod is in the PodRunning Phase, if it is available.
func IsRunningAndAvailable(pod *v1.Pod, minReadySeconds int32) bool {
	return pod.Status.Phase == v1.PodRunning && podutil.IsPodAvailable(pod, minReadySeconds, metav1.Now())
}

// GetPodsRevisions return revision hash set of these pods.
func GetPodsRevisions(pods []*v1.Pod) sets.String {
	revisions := sets.NewString()
	for _, p := range pods {
		revisions.Insert(GetPodRevision(p))
	}
	return revisions
}

// NextRevision finds the next valid revision number based on revisions. If the length of revisions
// is 0 this is 1. Otherwise, it is 1 greater than the largest revision's Revision. This method
// assumes that revisions has been sorted by Revision.
func NextRevision(revisions []*appsv1.ControllerRevision) int64 {
	count := len(revisions)
	if count <= 0 {
		return 1
	}
	return revisions[count-1].Revision + 1
}

// IsRunningAndReady returns true if pod is in the PodRunning Phase, if it is ready.
func IsRunningAndReady(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodRunning && podutil.IsPodReady(pod)
}

// SplitPodsByRevision returns Pods matched and unmatched the given revision
func SplitPodsByRevision(pods []*v1.Pod, rev string) (matched, unmatched []*v1.Pod) {
	for _, p := range pods {
		if GetPodRevision(p) == rev {
			matched = append(matched, p)
		} else {
			unmatched = append(unmatched, p)
		}
	}
	return
}

// DoItSlowly tries to call the provided function a total of 'count' times,
// starting slow to check for errors, then speeding up if calls succeed.
//
// It groups the calls into batches, starting with a group of initialBatchSize.
// Within each batch, it may call the function multiple times concurrently.
//
// If a whole batch succeeds, the next batch may get exponentially larger.
// If there are any failures in a batch, all remaining batches are skipped
// after waiting for the current batch to complete.
//
// It returns the number of successful calls to the function.
func DoItSlowly(count int, initialBatchSize int, fn func() error) (int, error) {
	remaining := count
	successes := 0
	for batchSize := integer.IntMin(remaining, initialBatchSize); batchSize > 0; batchSize = integer.IntMin(2*batchSize, remaining) {
		errCh := make(chan error, batchSize)
		var wg sync.WaitGroup
		wg.Add(batchSize)
		for i := 0; i < batchSize; i++ {
			go func() {
				defer wg.Done()
				if err := fn(); err != nil {
					errCh <- err
				}
			}()
		}
		wg.Wait()
		curSuccesses := batchSize - len(errCh)
		successes += curSuccesses
		if len(errCh) > 0 {
			return successes, <-errCh
		}
		remaining -= batchSize
	}
	return successes, nil
}

// DumpJSON returns the JSON encoding
func DumpJSON(o interface{}) string {
	j, _ := json.Marshal(o)
	return string(j)
}

// GetPodGameDeployments returns a list of GameDeployments that potentially match a pod.
// Only the one specified in the Pod's ControllerRef will actually manage it.
// Returns an error only if no matching GameDeployment are found.
func GetPodGameDeployments(pod *v1.Pod, gdcLister gdlister.GameDeploymentLister) ([]*gdv1alpha1.GameDeployment, error) {
	var selector labels.Selector
	var ps *gdv1alpha1.GameDeployment

	if len(pod.Labels) == 0 {
		return nil, fmt.Errorf("no GameDeployment found for pod %v because it has no labels", pod.Name)
	}

	list, err := gdcLister.GameDeployments(pod.Namespace).List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var psList []*gdv1alpha1.GameDeployment
	for i := range list {
		ps = list[i]
		if ps.Namespace != pod.Namespace {
			continue
		}
		selector, err = metav1.LabelSelectorAsSelector(ps.Spec.Selector)
		if err != nil {
			return nil, fmt.Errorf("invalid selector: %v", err)
		}

		// If a GameDeployment with a nil or empty selector creeps in, it should match nothing, not everything.
		if selector.Empty() || !selector.Matches(labels.Set(pod.Labels)) {
			continue
		}
		psList = append(psList, ps)
	}

	if len(psList) == 0 {
		return nil, fmt.Errorf("could not find GameDeployment for pod %s in namespace %s with labels: %v", pod.Name, pod.Namespace, pod.Labels)
	}

	return psList, nil
}

type AlphabetSortPods []*v1.Pod

func (s AlphabetSortPods) Len() int      { return len(s) }
func (s AlphabetSortPods) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s AlphabetSortPods) Less(i, j int) bool {
	if strings.Compare(s[i].Name, s[j].Name) > 0 {
		return false
	}
	return true
}
