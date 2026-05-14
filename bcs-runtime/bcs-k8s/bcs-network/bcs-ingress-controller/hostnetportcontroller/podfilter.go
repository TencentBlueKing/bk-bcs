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

package hostnetportcontroller

import (
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

const hostnetPodFilterName = "hostnet_pod"

// HostNetPodFilter filters pod events for HostNetPortPool annotation.
type HostNetPodFilter struct{}

// NewHostNetPodFilter creates a new pod filter.
func NewHostNetPodFilter() *HostNetPodFilter {
	return &HostNetPodFilter{}
}

// Create handles pod creation events.
func (f *HostNetPodFilter) Create(e event.CreateEvent, q workqueue.RateLimitingInterface) {
	pod, ok := e.Object.(*k8scorev1.Pod)
	if !ok {
		return
	}
	if !hasHostNetPortPoolAnnotation(pod.Annotations) {
		return
	}
	metrics.IncreaseEventCounter(hostnetPodFilterName, metrics.EventTypeAdd)
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name: pod.Name, Namespace: pod.Namespace,
	}})
}

// Update handles pod update events.
func (f *HostNetPodFilter) Update(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	oldPod, okOld := e.ObjectOld.(*k8scorev1.Pod)
	newPod, okNew := e.ObjectNew.(*k8scorev1.Pod)
	if !okOld || !okNew {
		return
	}
	if !hasHostNetPortPoolAnnotation(newPod.Annotations) && !hasHostNetPortPoolAnnotation(oldPod.Annotations) {
		return
	}
	metrics.IncreaseEventCounter(hostnetPodFilterName, metrics.EventTypeUpdate)
	if !checkHostNetPortPodNeedReconcile(oldPod, newPod) {
		blog.V(5).Infof("hostnetport: ignoring pod %s/%s update (no relevant change)", newPod.Namespace, newPod.Name)
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name: newPod.Name, Namespace: newPod.Namespace,
	}})
}

// Delete handles pod deletion events.
func (f *HostNetPodFilter) Delete(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	pod, ok := e.Object.(*k8scorev1.Pod)
	if !ok {
		return
	}
	if !hasHostNetPortPoolAnnotation(pod.Annotations) {
		return
	}
	metrics.IncreaseEventCounter(hostnetPodFilterName, metrics.EventTypeDelete)
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name: pod.Name, Namespace: pod.Namespace,
	}})
}

// Generic handles generic events.
func (f *HostNetPodFilter) Generic(e event.GenericEvent, q workqueue.RateLimitingInterface) {
	pod, ok := e.Object.(*k8scorev1.Pod)
	if !ok {
		return
	}
	if !hasHostNetPortPoolAnnotation(pod.Annotations) {
		return
	}
	metrics.IncreaseEventCounter(hostnetPodFilterName, metrics.EventTypeUnknown)
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name: pod.Name, Namespace: pod.Namespace,
	}})
}

func hasHostNetPortPoolAnnotation(annotations map[string]string) bool {
	_, ok := annotations[constant.AnnotationForHostNetPortPool]
	return ok
}

func checkHostNetPortPodNeedReconcile(oldPod, newPod *k8scorev1.Pod) bool {
	if oldPod == nil || newPod == nil {
		return true
	}
	if oldPod.Spec.NodeName != newPod.Spec.NodeName {
		return true
	}
	if oldPod.Status.Phase != newPod.Status.Phase {
		return true
	}
	// Reconcile skips pods with DeletionTimestamp set, so no need to enqueue
	if newPod.DeletionTimestamp != nil {
		return false
	}
	// Check hostnetportpool annotation changes
	for _, key := range []string{
		constant.AnnotationForHostNetPortPool,
		constant.AnnotationForHostNetPortPoolNamespace,
		constant.AnnotationForHostNetPortPoolPortCount,
		constant.AnnotationForHostNetPortPoolBindingResult,
		constant.AnnotationForHostNetPortPoolBindingStatus,
	} {
		old := getAnnotation(oldPod.Annotations, key)
		new := getAnnotation(newPod.Annotations, key)
		if old != new {
			return true
		}
	}
	return false
}

func getAnnotation(annotations map[string]string, key string) string {
	if annotations == nil {
		return ""
	}
	return annotations[key]
}

// isHostNetPortPoolAnnotation checks if a key belongs to hostnetportpool namespace (unused but kept for reference).
func isHostNetPortPoolAnnotation(key string) bool {
	return strings.HasPrefix(key, "hostnetportpool.")
}
