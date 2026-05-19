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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/hostnetportpoolcache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

// HostNetNodeReconciler reconciles Node deletion events to clean up cache and metrics.
type HostNetNodeReconciler struct {
	ctx    context.Context
	client client.Client
	cache  *hostnetportpoolcache.HostNetPortPoolCache
}

// NewHostNetNodeReconciler creates a new Node reconciler.
func NewHostNetNodeReconciler(
	ctx context.Context,
	cli client.Client,
	cache *hostnetportpoolcache.HostNetPortPoolCache,
) *HostNetNodeReconciler {
	return &HostNetNodeReconciler{
		ctx:    ctx,
		client: cli,
		cache:  cache,
	}
}

// SetupWithManager registers the controller for Node events.
// WithEventFilter restricts the controller to only process Delete events,
// since this controller's sole responsibility is garbage-collecting cache entries
// when a Node is removed. Create/Update events are no-ops and filtered out.
func (r *HostNetNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8scorev1.Node{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc:  func(event.CreateEvent) bool { return false },
			UpdateFunc:  func(event.UpdateEvent) bool { return false },
			GenericFunc: func(event.GenericEvent) bool { return false },
			DeleteFunc:  func(event.DeleteEvent) bool { return true },
		}).
		Complete(r)
}

// Reconcile handles Node events. Only deletion (NotFound) triggers cache cleanup.
// This controller does NOT perform port allocation for Nodes. Port allocation uses
// a lazy design: segments are allocated on-demand by HostNetPodReconciler when a Pod
// is scheduled onto a Node (see pod_controller.go). The Node controller's sole
// responsibility is to garbage-collect cache entries and metrics when a Node is removed.
func (r *HostNetNodeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	blog.V(3).Infof("hostnetport node reconcile: %s", req.Name)

	if !r.cache.IsSynced() {
		blog.V(4).Infof("hostnetport: cache not yet synced, deferring node %s", req.Name)
		return ctrl.Result{RequeueAfter: 1 * time.Second}, nil
	}

	node := &k8scorev1.Node{}
	if err := r.client.Get(r.ctx, req.NamespacedName, node); err != nil {
		if k8serrors.IsNotFound(err) {
			affected := r.cache.CleanupNode(req.Name)
			for _, p := range affected {
				metrics.CleanHostNetSegmentMetrics(p.PoolName, p.PoolNamespace, req.Name)
			}
			blog.Infof("hostnetport: cleaned up node %s from cache, affected %d pool(s)",
				req.Name, len(affected))
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
