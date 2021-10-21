/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package portbindingcontroller

import (
	"context"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	ingresscommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	bcsnetcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8sapitypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// PortBindingReconciler reconciler for bcs port pool
type PortBindingReconciler struct {
	cleanInterval time.Duration
	ctx           context.Context
	k8sClient     client.Client
	poolCache     *portpoolcache.Cache
}

// NewPortBindingReconciler create PortBindingReconciler
func NewPortBindingReconciler(
	ctx context.Context, cleanInterval time.Duration,
	k8sClient client.Client, poolCache *portpoolcache.Cache) *PortBindingReconciler {
	return &PortBindingReconciler{
		ctx:           ctx,
		cleanInterval: cleanInterval,
		k8sClient:     k8sClient,
		poolCache:     poolCache,
	}
}

// Reconcile reconcile port pool
func (pbr *PortBindingReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	portBinding := &networkextensionv1.PortBinding{}
	if err := pbr.k8sClient.Get(pbr.ctx, req.NamespacedName, portBinding); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	blog.V(3).Infof("PortBinding %+v triggered", req.NamespacedName)
	pod := &k8scorev1.Pod{}
	if err := pbr.k8sClient.Get(pbr.ctx, k8sapitypes.NamespacedName{
		Name:      portBinding.GetName(),
		Namespace: portBinding.GetNamespace(),
	}, pod); err != nil {
		if k8serrors.IsNotFound(err) {
			return pbr.cleanPortBinding(portBinding)
		}
		blog.Warnf("get pod %s/%s failed, err %s",
			portBinding.GetName(), portBinding.GetNamespace(), err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	if portBinding.DeletionTimestamp != nil {
		return pbr.cleanPortBinding(portBinding)
	}
	if len(pod.Status.PodIP) == 0 {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 300 * time.Millisecond,
		}, nil
	}
	pbhandler := newPortBindingHandler(pbr.ctx, pbr.k8sClient)
	retry, err := pbhandler.ensurePortBinding(pod, portBinding)
	if err != nil {
		blog.Warnf("ensure port binding %s/%s failed, err %s",
			portBinding.GetName(), portBinding.GetNamespace(), err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	if retry {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	blog.Infof("ensure port binding %s/%s successfully", portBinding.GetName(), portBinding.GetNamespace())
	return ctrl.Result{}, nil
}

func (pbr *PortBindingReconciler) cleanPortBinding(portBinding *networkextensionv1.PortBinding) (ctrl.Result, error) {
	if portBinding.Status.Status == constant.PortBindingStatusCleaned {
		expired, err := isPortBindingExpired(portBinding)
		if !expired && err == nil {
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: pbr.cleanInterval,
			}, nil
		}
		if err != nil {
			blog.Warnf("check port binding expire time failed, err %s", err.Error())
		}
		if portBinding.DeletionTimestamp != nil {
			portBinding.Finalizers = bcsnetcommon.RemoveString(
				portBinding.Finalizers, constant.FinalizerNameBcsIngressController)
			if err := pbr.k8sClient.Update(pbr.ctx, portBinding, &client.UpdateOptions{}); err != nil {
				blog.Warnf("remote finalizer from port binding %s/%s failed, err %s",
					portBinding.GetName(), portBinding.GetNamespace(), err.Error())
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 3 * time.Second,
				}, nil
			}
			pbr.poolCache.Lock()
			defer pbr.poolCache.Unlock()
			for _, portBindingItem := range portBinding.Spec.PortBindingList {
				poolKey := ingresscommon.GetNamespacedNameKey(portBindingItem.PoolName, portBindingItem.PoolNamespace)
				blog.Infof("release portbinding %s %s %s %d %d from cache",
					poolKey, portBindingItem.GetKey(), portBindingItem.Protocol,
					portBindingItem.StartPort, portBindingItem.EndPort)
				pbr.poolCache.ReleasePortBinding(
					poolKey, portBindingItem.GetKey(), portBindingItem.Protocol,
					portBindingItem.StartPort, portBindingItem.EndPort)
			}

		} else {
			if err := pbr.k8sClient.Delete(pbr.ctx, portBinding, &client.DeleteOptions{}); err != nil {
				blog.Warnf("delete port binding %s/%s from api failed, err %s",
					portBinding.GetName(), portBinding.GetNamespace(), err.Error())
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 3 * time.Second,
				}, nil
			}
		}

		return ctrl.Result{}, nil
	}
	pbhandler := newPortBindingHandler(pbr.ctx, pbr.k8sClient)
	retry, err := pbhandler.cleanPortBinding(portBinding)
	if err != nil {
		blog.Warnf("delete port binding %s/%s failed, err %s",
			portBinding.GetName(), portBinding.GetNamespace(), err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	if retry {
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}
	return ctrl.Result{}, nil
}

// SetupWithManager set reconciler
func (pbr *PortBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&networkextensionv1.PortBinding{}).
		Watches(&source.Kind{Type: &k8scorev1.Pod{}}, NewPodFilter(mgr.GetClient())).
		WithEventFilter(getPortBindingPredicate()).
		Complete(pbr)
}

func getPortBindingPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			newPoolBinding, okNew := e.ObjectNew.(*networkextensionv1.PortBinding)
			oldPoolBinding, okOld := e.ObjectOld.(*networkextensionv1.PortBinding)
			if !okNew || !okOld {
				return true
			}
			if reflect.DeepEqual(newPoolBinding.Spec, oldPoolBinding.Spec) &&
				reflect.DeepEqual(newPoolBinding.Status.PortBindingStatusList,
					oldPoolBinding.Status.PortBindingStatusList) &&
				newPoolBinding.Status.Status == oldPoolBinding.Status.Status &&
				reflect.DeepEqual(newPoolBinding.ObjectMeta, oldPoolBinding.ObjectMeta) {
				blog.V(5).Infof("portbinding %+v updated, but spec not change", newPoolBinding)
				return false
			}
			return true
		},
	}
}
