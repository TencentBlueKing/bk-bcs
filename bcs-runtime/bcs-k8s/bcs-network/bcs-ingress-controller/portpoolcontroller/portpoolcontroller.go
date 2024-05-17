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

// Package portpoolcontroller controller for portpool
package portpoolcontroller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	gocache "github.com/patrickmn/go-cache"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	ingresscommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	pkgcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
)

// PortPoolReconciler reconciler for bcs port pool
type PortPoolReconciler struct {
	opts        *option.ControllerOption
	ctx         context.Context
	k8sClient   client.Client
	eventer     record.EventRecorder
	lbClient    cloud.LoadBalance
	poolCache   *portpoolcache.Cache
	lbIDCache   *gocache.Cache
	lbNameCache *gocache.Cache
	isCacheSync bool
}

// NewPortPoolReconciler create port pool reconciler
func NewPortPoolReconciler(
	ctx context.Context,
	opts *option.ControllerOption,
	lbClient cloud.LoadBalance,
	k8sClient client.Client,
	eventer record.EventRecorder,
	poolCache *portpoolcache.Cache, lbIDCache *gocache.Cache, lbNameCache *gocache.Cache) *PortPoolReconciler {
	return &PortPoolReconciler{
		ctx:         ctx,
		opts:        opts,
		lbClient:    lbClient,
		k8sClient:   k8sClient,
		eventer:     eventer,
		poolCache:   poolCache,
		lbIDCache:   lbIDCache,
		lbNameCache: lbNameCache,
	}
}

// Reconcile reconcile port pool
func (ppr *PortPoolReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	blog.V(3).Infof("PortPool %+v triggered", req.NamespacedName)
	if !ppr.isCacheSync {
		if err := ppr.initPortPoolCache(); err != nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 1 * time.Second,
			}, nil
		}
		ppr.isCacheSync = true
	}
	portPool := &networkextensionv1.PortPool{}
	if err := ppr.k8sClient.Get(ppr.ctx, req.NamespacedName, portPool); err != nil {
		if k8serrors.IsNotFound(err) {
			// we will add finalizer for port pool, so when get delete event, do nothing
			return ctrl.Result{}, nil
		}
	}

	handler := newPortPoolHandler(req.NamespacedName.Namespace, ppr.opts.Region,
		ppr.lbClient, ppr.k8sClient, ppr.poolCache, ppr.lbIDCache, ppr.lbNameCache)
	if portPool.DeletionTimestamp != nil {
		retry, err := handler.deletePortPool(portPool)
		if err != nil {
			blog.Warnf("delete port pool '%s/%s' failed, err: %s", portPool.GetNamespace(), portPool.GetName(),
				err.Error())
			metrics.IncreaseFailMetric(metrics.ObjectPortPool, metrics.EventTypeDelete)
			ppr.recordListenerEvent(portPool, k8scorev1.EventTypeWarning, "delete port pool failed", err.Error())
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, nil
		}
		if retry {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 5 * time.Second,
			}, nil
		}
		return ctrl.Result{}, nil
	}

	if !pkgcommon.ContainsString(portPool.Finalizers, constant.FinalizerNameBcsIngressController) {
		portPool.Finalizers = append(portPool.Finalizers, constant.FinalizerNameBcsIngressController)
		if err := ppr.k8sClient.Update(context.Background(), portPool, &client.UpdateOptions{}); err != nil {
			blog.Warnf("add finalizer for port pool %s/%s failed, err %s",
				portPool.GetNamespace(), portPool.GetName(), err.Error())
		}
		return ctrl.Result{}, nil
	}

	// do update
	retry, err := handler.ensurePortPool(portPool)
	if err != nil {
		blog.Errorf("ensure portPool '%s/%s' failed, err: %s", portPool.GetNamespace(), portPool.GetName(),
			err.Error())
		metrics.IncreaseFailMetric(metrics.ObjectPortPool, metrics.EventTypeUnknown)
		ppr.recordListenerEvent(portPool, k8scorev1.EventTypeWarning, "update port pool failed", err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
		}, nil
	}
	if retry {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 5 * time.Second,
		}, nil
	}

	return ctrl.Result{Requeue: true,
		RequeueAfter: 20 * time.Minute}, nil
}

// when the first reconcile comes, there are all the data in the cache of controller manager
func (ppr *PortPoolReconciler) initPortPoolCache() error {
	ppr.poolCache.Lock()
	defer ppr.poolCache.Unlock()
	poolList := &networkextensionv1.PortPoolList{}
	if err := ppr.k8sClient.List(context.Background(), poolList); err != nil {
		return fmt.Errorf("list port pool failed when do init cache; err %s", err.Error())
	}
	for _, pool := range poolList.Items {
		// when bcs-ingress-controller start, need to recover cache
		poolKey := ingresscommon.GetNamespacedNameKey(pool.GetName(), pool.GetNamespace())
		for _, itemStatus := range pool.Status.PoolItemStatuses {
			if !ppr.poolCache.IsItemExisted(poolKey, itemStatus.GetKey()) {
				// 适配旧版本升级（1.28.0-alpha.55以下）
				if len(itemStatus.Protocol) == 0 {
					var protocol []string
					for _, item := range pool.Spec.PoolItems {
						if item.ItemName == itemStatus.ItemName {
							protocol = ingresscommon.GetPortPoolItemProtocols(item.Protocol)
							break
						}
					}
					if len(protocol) == 0 {
						itemStatus.Protocol = []string{constant.PortPoolPortProtocolTCP, constant.PortPoolPortProtocolUDP}
					} else {
						itemStatus.Protocol = protocol
					}
				}
				if err := ppr.poolCache.AddPortPoolItem(poolKey, pool.GetAllocatePolicy(), itemStatus); err != nil {
					blog.Warnf("failed to add port pool %s item %v to cache, err %s",
						poolKey, itemStatus, err.Error())
				} else {
					blog.Infof("add port pool %s item %v to cache", poolKey, itemStatus)
				}
			}
		}
	}
	bindingItemList := &networkextensionv1.PortBindingList{}
	if err := ppr.k8sClient.List(context.Background(), bindingItemList); err != nil {
		return fmt.Errorf("list port binding list failed, err %s", err.Error())
	}
	for _, portBinding := range bindingItemList.Items {
		for _, bindingItem := range portBinding.Spec.PortBindingList {
			ppr.poolCache.SetPortBindingUsed(bindingItem.PoolName+"/"+bindingItem.PoolNamespace,
				bindingItem.GetKey(), bindingItem.Protocol, bindingItem.StartPort, bindingItem.EndPort)
		}
	}
	return nil
}

// SetupWithManager set reconciler
func (ppr *PortPoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkextensionv1.PortPool{}).
		WithEventFilter(getPortPoolPredicate()).
		WithOptions(controller.Options{MaxConcurrentReconciles: 20}).
		Complete(ppr)
}

func (ppr *PortPoolReconciler) recordListenerEvent(pool *networkextensionv1.PortPool, eType, reason, msg string) {
	if ppr.eventer == nil {
		return
	}
	ppr.eventer.Event(pool, eType, reason, msg)
}

func getPortPoolPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newPool, okNew := e.ObjectNew.(*networkextensionv1.PortPool)
			oldPool, okOld := e.ObjectOld.(*networkextensionv1.PortPool)
			if !okNew || !okOld {
				return true
			}
			if reflect.DeepEqual(newPool.Spec, oldPool.Spec) &&
				reflect.DeepEqual(newPool.DeletionTimestamp, oldPool.DeletionTimestamp) &&
				reflect.DeepEqual(newPool.Finalizers, oldPool.Finalizers) {
				blog.V(5).Infof("port pool %+v updated, but spec not change", newPool)
				return false
			}
			return true
		},
	}
}
