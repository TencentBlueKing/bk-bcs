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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/hostnetportpoolcache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

// HostNetPodReconciler reconciles Pods that request hostNetwork port allocation.
type HostNetPodReconciler struct {
	ctx     context.Context
	client  client.Client
	cache   *hostnetportpoolcache.HostNetPortPoolCache
	eventer record.EventRecorder
}

// NewHostNetPodReconciler creates a new Pod reconciler.
func NewHostNetPodReconciler(
	ctx context.Context,
	cli client.Client,
	cache *hostnetportpoolcache.HostNetPortPoolCache,
	eventer record.EventRecorder,
) *HostNetPodReconciler {
	return &HostNetPodReconciler{
		ctx:     ctx,
		client:  cli,
		cache:   cache,
		eventer: eventer,
	}
}

// SetupWithManager registers the controller for Pod events filtered by HostNetPodFilter.
func (r *HostNetPodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkextensionv1.HostNetPortPool{}).
		Watches(&source.Kind{Type: &k8scorev1.Pod{}}, NewHostNetPodFilter()).
		Complete(r)
}

// Reconcile handles Pod port allocation events.
func (r *HostNetPodReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	blog.V(3).Infof("hostnetport pod reconcile: %s/%s", req.Namespace, req.Name)

	if !r.cache.IsSynced() {
		blog.V(4).Infof("hostnetport: cache not yet synced, deferring pod %s/%s", req.Namespace, req.Name)
		return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
	}

	pod := &k8scorev1.Pod{}
	if err := r.client.Get(r.ctx, req.NamespacedName, pod); err != nil {
		if k8serrors.IsNotFound(err) {
			r.handlePodDeletion(req.Namespace, req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	return r.reconcilePod(pod)
}

// handlePodDeletion releases port segments when a pod is deleted.
func (r *HostNetPodReconciler) handlePodDeletion(namespace, name string) {
	podKey := fmt.Sprintf("%s/%s", namespace, name)
	affected := r.cache.ReleaseByPodKey(podKey)
	metrics.CleanHostNetAllocateFailedMetric(name, namespace)
	blog.Infof("hostnetport: released segments for deleted pod %s, affected %d pool(s)",
		podKey, len(affected))
}

// reconcilePod implements the pod port allocation flow.
func (r *HostNetPodReconciler) reconcilePod(pod *k8scorev1.Pod) (ctrl.Result, error) {
	poolName, ok := pod.Annotations[constant.AnnotationForHostNetPortPool]
	if !ok {
		return ctrl.Result{}, nil
	}

	if pod.Status.Phase == k8scorev1.PodFailed || pod.Status.Phase == k8scorev1.PodSucceeded {
		r.handlePodDeletion(pod.Namespace, pod.Name)
		return ctrl.Result{}, nil
	}

	if pod.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	if pod.Spec.NodeName == "" {
		blog.V(4).Infof("hostnetport: pod %s/%s not yet scheduled, skipping", pod.Namespace, pod.Name)
		return ctrl.Result{}, nil
	}

	// Idempotent check via cache instead of Pod annotation to avoid stale informer reads
	podKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
	if r.cache.IsPodAllocated(podKey) {
		return ctrl.Result{}, nil
	}

	poolNamespace := pod.Namespace
	if ns, ok := pod.Annotations[constant.AnnotationForHostNetPortPoolNamespace]; ok && ns != "" {
		poolNamespace = ns
	}
	poolKey := fmt.Sprintf("%s/%s", poolNamespace, poolName)

	poolObj := &networkextensionv1.HostNetPortPool{}
	if err := r.client.Get(r.ctx, types.NamespacedName{
		Namespace: poolNamespace, Name: poolName,
	}, poolObj); err != nil {
		if k8serrors.IsNotFound(err) {
			r.patchPodBindingAnnotation(pod, "Failed")
			r.eventer.Eventf(pod, k8scorev1.EventTypeWarning, "AllocateHostNetPortFailed",
				"HostNetPortPool %s not found", poolKey)
			metrics.ReportHostNetAllocateFailed(pod.Name, pod.Namespace, true)
			metrics.IncreaseHostNetAllocateFailedTotal(poolName, pod.Spec.NodeName)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	segmentsNeeded, err := r.calculateSegmentsNeeded(pod, poolObj)
	if err != nil {
		r.patchPodBindingAnnotation(pod, "Failed")
		r.eventer.Eventf(pod, k8scorev1.EventTypeWarning, "AllocateHostNetPortFailed",
			"Invalid portcount annotation: %v", err)
		metrics.ReportHostNetAllocateFailed(pod.Name, pod.Namespace, true)
		blog.Warnf("hostnetport: pod %s invalid portcount: %v", podKey, err)
		return ctrl.Result{}, nil
	}

	startPort, endPort, allocErr := r.cache.AllocateContiguous(poolKey, pod.Spec.NodeName, podKey, segmentsNeeded)
	if allocErr != nil {
		if errors.Is(allocErr, hostnetportpoolcache.ErrPoolNotInCache) {
			blog.Warnf("hostnetport: pool %s not yet in cache for pod %s, will retry: %v",
				poolKey, podKey, allocErr)
			return ctrl.Result{RequeueAfter: 3 * time.Second}, nil
		}
		r.patchPodBindingAnnotation(pod, "Failed")
		r.eventer.Eventf(pod, k8scorev1.EventTypeWarning, "AllocateHostNetPortFailed",
			"No available segment on node %s from pool %s: %v", pod.Spec.NodeName, poolKey, allocErr)
		metrics.ReportHostNetAllocateFailed(pod.Name, pod.Namespace, true)
		metrics.IncreaseHostNetAllocateFailedTotal(poolName, pod.Spec.NodeName)
		blog.Warnf("hostnetport: allocate failed for pod %s: %v", podKey, allocErr)
		return ctrl.Result{}, nil
	}

	bindingResult := hostnetportpoolcache.HostNetPortPoolBindingResult{
		PoolName:      poolName,
		PoolNamespace: poolNamespace,
		NodeName:      pod.Spec.NodeName,
		StartPort:     startPort,
		EndPort:       endPort,
		SegmentLength: int(poolObj.Spec.SegmentLength),
	}

	if err := r.patchPodBindingResult(pod, &bindingResult); err != nil {
		r.cache.Release(poolKey, pod.Spec.NodeName, startPort, endPort)
		blog.Errorf("hostnetport: patch annotation failed for pod %s, rolled back: %v", podKey, err)
		return ctrl.Result{RequeueAfter: 5 * time.Second}, err
	}

	r.eventer.Eventf(pod, k8scorev1.EventTypeNormal, "HostNetPortAllocated",
		"Allocated port segment %d-%d on node %s", startPort, endPort, pod.Spec.NodeName)
	metrics.ReportHostNetAllocateFailed(pod.Name, pod.Namespace, false)
	blog.Infof("hostnetport: allocated %d-%d on node %s for pod %s",
		startPort, endPort, pod.Spec.NodeName, podKey)

	return ctrl.Result{}, nil
}

// calculateSegmentsNeeded parses the portcount annotation and converts to segment count.
// Returns an error if portcount is syntactically invalid or non-positive.
func (r *HostNetPodReconciler) calculateSegmentsNeeded(
	pod *k8scorev1.Pod, pool *networkextensionv1.HostNetPortPool) (int, error) {

	portCountStr, ok := pod.Annotations[constant.AnnotationForHostNetPortPoolPortCount]
	if !ok {
		return 1, nil
	}

	var portCount int
	if _, err := fmt.Sscanf(portCountStr, "%d", &portCount); err != nil {
		return 0, fmt.Errorf("pod %s/%s has invalid portcount %q: %w",
			pod.Namespace, pod.Name, portCountStr, err)
	}
	if portCount <= 0 {
		return 0, fmt.Errorf("pod %s/%s has non-positive portcount %d",
			pod.Namespace, pod.Name, portCount)
	}

	segLen := int(pool.Spec.SegmentLength)
	return (portCount + segLen - 1) / segLen, nil
}

func (r *HostNetPodReconciler) patchPodBindingResult(
	pod *k8scorev1.Pod, result *hostnetportpoolcache.HostNetPortPoolBindingResult) error {

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("marshal binding result: %w", err)
	}

	patch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{
				constant.AnnotationForHostNetPortPoolBindingResult: string(resultJSON),
				constant.AnnotationForHostNetPortPoolBindingStatus: "Ready",
			},
		},
	}
	patchData, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("marshal patch: %w", err)
	}

	return r.client.Patch(r.ctx, pod, client.RawPatch(types.MergePatchType, patchData))
}

// patchPodBindingAnnotation patches the Pod's binding status annotation.
func (r *HostNetPodReconciler) patchPodBindingAnnotation(pod *k8scorev1.Pod, status string) {
	patch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": map[string]string{
				constant.AnnotationForHostNetPortPoolBindingStatus: status,
			},
		},
	}
	patchData, err := json.Marshal(patch)
	if err != nil {
		blog.Warnf("hostnetport: marshal status patch for pod %s/%s failed: %v",
			pod.Namespace, pod.Name, err)
		return
	}

	if err := r.client.Patch(r.ctx, pod, client.RawPatch(types.MergePatchType, patchData)); err != nil {
		blog.Warnf("hostnetport: failed to patch status annotation for pod %s/%s: %v",
			pod.Namespace, pod.Name, err)
	}
}
