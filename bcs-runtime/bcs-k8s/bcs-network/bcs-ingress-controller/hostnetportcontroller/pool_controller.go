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

// Package hostnetportcontroller implements the HostNetPortPool allocator controllers.
package hostnetportcontroller

import (
	"context"
	"encoding/json"
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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/hostnetportpoolcache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

// HostNetPortPoolReconciler reconciles HostNetPortPool CR lifecycle: creation, update, deletion.
type HostNetPortPoolReconciler struct {
	ctx         context.Context
	client      client.Client
	cache       *hostnetportpoolcache.HostNetPortPoolCache
	eventer     record.EventRecorder
	cacheSynced bool
}

// NewHostNetPortPoolReconciler creates a new HostNetPortPool reconciler.
func NewHostNetPortPoolReconciler(
	ctx context.Context,
	cli client.Client,
	cache *hostnetportpoolcache.HostNetPortPoolCache,
	eventer record.EventRecorder,
) *HostNetPortPoolReconciler {
	return &HostNetPortPoolReconciler{
		ctx:     ctx,
		client:  cli,
		cache:   cache,
		eventer: eventer,
	}
}

// SetupWithManager registers the controller for HostNetPortPool CRD.
// It also watches a channel fed by cache mutation events (allocate/release)
// so that status sync is event-driven rather than polling-based.
func (r *HostNetPortPoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	statusCh := make(chan event.GenericEvent, 64)
	go r.forwardCacheNotifications(statusCh)

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkextensionv1.HostNetPortPool{}).
		Watches(&source.Channel{Source: statusCh},
			&handler.EnqueueRequestsFromMapFunc{
				ToRequests: handler.ToRequestsFunc(
					func(obj handler.MapObject) []reconcile.Request {
						return []reconcile.Request{{
							NamespacedName: types.NamespacedName{
								Namespace: obj.Meta.GetNamespace(),
								Name:      obj.Meta.GetName(),
							},
						}}
					}),
			}).
		Complete(r)
}

// forwardCacheNotifications converts cache PoolChangedEvents into controller-runtime
// GenericEvents so the pool reconciler is triggered on every allocation change.
func (r *HostNetPortPoolReconciler) forwardCacheNotifications(ch chan<- event.GenericEvent) {
	for {
		select {
		case <-r.ctx.Done():
			return
		case evt, ok := <-r.cache.NotifyCh():
			if !ok {
				return
			}
			pool := &networkextensionv1.HostNetPortPool{}
			pool.SetName(evt.PoolName)
			pool.SetNamespace(evt.PoolNamespace)
			ch <- event.GenericEvent{
				Meta:   pool,
				Object: pool,
			}
		}
	}
}

// Reconcile handles HostNetPortPool CR events.
func (r *HostNetPortPoolReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	blog.V(3).Infof("hostnetport pool reconcile: %s/%s", req.Namespace, req.Name)

	if !r.cacheSynced {
		if err := r.initCache(); err != nil {
			blog.Errorf("hostnetport pool initCache failed: %v", err)
			return ctrl.Result{RequeueAfter: 5 * time.Second}, err
		}
		r.cacheSynced = true
		r.cache.MarkSynced()
	}

	pool := &networkextensionv1.HostNetPortPool{}
	if err := r.client.Get(r.ctx, req.NamespacedName, pool); err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Infof("hostnetport: pool %s/%s not found, skipping", req.Namespace, req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	return r.reconcilePool(pool)
}

// initCache rebuilds the in-memory cache from existing CRs and Pods.
func (r *HostNetPortPoolReconciler) initCache() error {
	startTime := time.Now()
	blog.Infof("hostnetport: starting cache rebuild")

	poolList := &networkextensionv1.HostNetPortPoolList{}
	if err := r.client.List(r.ctx, poolList); err != nil {
		return fmt.Errorf("list HostNetPortPool failed: %w", err)
	}
	for i := range poolList.Items {
		r.cache.AddPool(&poolList.Items[i])
	}

	podList := &k8scorev1.PodList{}
	if err := r.client.List(r.ctx, podList); err != nil {
		return fmt.Errorf("list pods failed: %w", err)
	}

	recovered := 0
	for i := range podList.Items {
		pod := &podList.Items[i]
		if _, ok := pod.Annotations[constant.AnnotationForHostNetPortPool]; !ok {
			continue
		}
		if pod.Status.Phase == k8scorev1.PodFailed || pod.Status.Phase == k8scorev1.PodSucceeded {
			continue
		}
		resultStr, ok := pod.Annotations[constant.AnnotationForHostNetPortPoolBindingResult]
		if !ok {
			continue
		}

		var result hostnetportpoolcache.HostNetPortPoolBindingResult
		if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
			blog.Warnf("hostnetport: failed to parse binding result for pod %s/%s: %v",
				pod.Namespace, pod.Name, err)
			continue
		}

		poolKey := fmt.Sprintf("%s/%s", result.PoolNamespace, result.PoolName)
		podKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
		r.cache.RebuildFromPod(poolKey, result.NodeName, podKey, result.StartPort, result.EndPort)
		recovered++
	}

	// Cross-check with existing nodes: remove allocators for nodes that
	// were deleted while the controller was down.
	nodeList := &k8scorev1.NodeList{}
	if err := r.client.List(r.ctx, nodeList); err != nil {
		return fmt.Errorf("list nodes failed: %w", err)
	}
	validNodes := make(map[string]struct{}, len(nodeList.Items))
	for i := range nodeList.Items {
		validNodes[nodeList.Items[i].Name] = struct{}{}
	}
	staleCount := r.cache.CleanupStaleNodes(validNodes)
	if staleCount > 0 {
		blog.Warnf("hostnetport: removed %d stale node allocator(s) during cache rebuild", staleCount)
	}

	metrics.ReportHostNetCacheRebuildRecovered(recovered)
	blog.Infof("hostnetport: cache rebuild completed in %v, recovered %d pods",
		time.Since(startTime), recovered)
	return nil
}

// reconcilePool handles HostNetPortPool CR lifecycle: creation, update, and deletion.
func (r *HostNetPortPoolReconciler) reconcilePool(
	pool *networkextensionv1.HostNetPortPool) (ctrl.Result, error) {

	poolKey := fmt.Sprintf("%s/%s", pool.Namespace, pool.Name)

	// Handle deletion
	if pool.DeletionTimestamp != nil {
		return r.handlePoolDeletion(pool, poolKey)
	}

	// Add finalizer if missing
	if !controllerutil.ContainsFinalizer(pool, constant.FinalizerNameHostNetPortPool) {
		controllerutil.AddFinalizer(pool, constant.FinalizerNameHostNetPortPool)
		if err := r.client.Update(r.ctx, pool); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Detect pending shrink before applying update
	cachedStart, cachedEnd, cacheExists := r.cache.GetPoolRange(poolKey)
	isShrinkPending := cacheExists &&
		(int(pool.Spec.EndPort) < cachedEnd || int(pool.Spec.StartPort) > cachedStart)

	// Sync cache with pool spec
	conflicts := r.cache.UpdatePool(pool)
	if len(conflicts) > 0 {
		for _, c := range conflicts {
			r.eventer.Eventf(pool, k8scorev1.EventTypeWarning, "PoolShrinkConflict",
				"Cannot shrink port range: segment in use: node=%s ports=%d-%d pod=%s",
				c.NodeName, c.StartPort, c.EndPort, c.PodKey)
			metrics.IncreaseHostNetPoolShrinkConflict(pool.Name, pool.Namespace, c.NodeName)
		}
		blog.Warnf("hostnetport: pool %s shrink blocked by %d conflict(s)", poolKey, len(conflicts))
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	if isShrinkPending {
		r.eventer.Eventf(pool, k8scorev1.EventTypeNormal, "PoolShrinkResolved",
			"Port range change applied successfully: now %d-%d (was %d-%d)",
			pool.Spec.StartPort, pool.Spec.EndPort, cachedStart, cachedEnd)
		blog.Infof("hostnetport: pool %s shrink resolved, range now %d-%d",
			poolKey, pool.Spec.StartPort, pool.Spec.EndPort)
	}

	nodeAllocs := r.cache.GetNodeAllocations(poolKey)
	if !nodeAllocationsEqual(pool.Status.NodeAllocations, nodeAllocs) ||
		pool.Status.Status != "Ready" {

		statusPatch, err := json.Marshal(map[string]interface{}{
			"status": map[string]interface{}{
				"nodeAllocations": nodeAllocs,
				"status":          "Ready",
			},
		})
		if err != nil {
			blog.Errorf("hostnetport: marshal status patch for pool %s failed: %v", poolKey, err)
			return ctrl.Result{RequeueAfter: 10 * time.Second}, err
		}
		if err := r.client.Status().Patch(r.ctx, pool,
			client.RawPatch(types.MergePatchType, statusPatch)); err != nil {
			blog.Warnf("hostnetport: failed to patch pool %s status: %v", poolKey, err)
			return ctrl.Result{RequeueAfter: 10 * time.Second}, err
		}
		blog.Infof("hostnetport: updated pool %s status, %d node allocation(s)",
			poolKey, len(nodeAllocs))
	}

	for _, na := range nodeAllocs {
		metrics.ReportHostNetSegmentMetrics(pool.Name, pool.Namespace, na.NodeName,
			na.AllocatedCount, na.TotalSegments)
	}

	return ctrl.Result{}, nil
}

// handlePoolDeletion cleans up cache and removes the finalizer when a pool is being deleted.
// If any segment is still in use, deletion is deferred and the request is requeued.
func (r *HostNetPortPoolReconciler) handlePoolDeletion(
	pool *networkextensionv1.HostNetPortPool, poolKey string) (ctrl.Result, error) {

	segments := r.cache.GetAllocatedSegments()
	for _, seg := range segments {
		if seg.PoolKey == poolKey {
			blog.Warnf("hostnetport: cannot remove finalizer from pool %s, segment %d-%d still in use by %s",
				poolKey, seg.StartPort, seg.EndPort, seg.PodKey)
			return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
		}
	}

	r.cache.RemovePool(poolKey)
	controllerutil.RemoveFinalizer(pool, constant.FinalizerNameHostNetPortPool)
	if err := r.client.Update(r.ctx, pool); err != nil {
		return ctrl.Result{}, err
	}
	blog.Infof("hostnetport: removed finalizer and cache for pool %s", poolKey)
	return ctrl.Result{}, nil
}

func nodeAllocationsEqual(
	a []*networkextensionv1.NodeHostNetPortPoolStatus,
	b []*networkextensionv1.NodeHostNetPortPoolStatus) bool {

	if len(a) == 0 && len(b) == 0 {
		return true
	}
	if len(a) != len(b) {
		return false
	}
	aMap := make(map[string]*networkextensionv1.NodeHostNetPortPoolStatus, len(a))
	for _, n := range a {
		if n != nil {
			aMap[n.NodeName] = n
		}
	}
	for _, n := range b {
		if n == nil {
			continue
		}
		old, ok := aMap[n.NodeName]
		if !ok || old.AllocatedCount != n.AllocatedCount || old.TotalSegments != n.TotalSegments {
			return false
		}
	}
	return true
}

