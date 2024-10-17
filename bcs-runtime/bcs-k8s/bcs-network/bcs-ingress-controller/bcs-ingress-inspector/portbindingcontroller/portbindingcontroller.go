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

// Package portbindingcontroller controller for portbinding
package portbindingcontroller

import (
	"context"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	pbctrl "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/portbindingcontroller"
)

// PortBindingReconciler reconciler for bcs port pool
type PortBindingReconciler struct {
	ctx       context.Context
	k8sClient client.Client
	eventer   record.EventRecorder

	nodeBindCache *pbctrl.NodePortBindingCache
}

// NewPortBindingReconciler create PortBindingReconciler
func NewPortBindingReconciler(ctx context.Context, k8sClient client.Client,
	eventer record.EventRecorder, nodeBindCache *pbctrl.NodePortBindingCache) *PortBindingReconciler {
	return &PortBindingReconciler{
		ctx:           ctx,
		k8sClient:     k8sClient,
		eventer:       eventer,
		nodeBindCache: nodeBindCache,
	}
}

// Reconcile reconcile port pool
// portbinding name is same with pod name
// nolint  funlen
func (pbr *PortBindingReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	blog.V(3).Infof("PortBinding %+v triggered", req.NamespacedName)
	portBinding := &networkextensionv1.PortBinding{}
	if err := pbr.k8sClient.Get(pbr.ctx, req.NamespacedName, portBinding); err != nil {
		if k8serrors.IsNotFound(err) {
			portBinding.SetName(req.Name)
			portBinding.SetNamespace(req.Namespace)
			// 未找到PortBinding时， UpdateCache方法会清理对应缓存
			if err1 := pbr.nodeBindCache.UpdateCache(portBinding); err1 != nil {
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 3 * time.Second,
				}, nil
			}
		} else {
			blog.Warnf("get portbinding %v failed, err %s, requeue it", req.NamespacedName, err.Error())
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 3 * time.Second,
			}, nil
		}
	}

	// 仅当PortBinding类型为Node时，才更新NodePortBindingCache
	if portBindingType, ok := portBinding.Labels[networkextensionv1.PortBindingTypeLabelKey]; ok {
		if portBindingType == networkextensionv1.PortBindingTypeNode {
			if err := pbr.nodeBindCache.UpdateCache(portBinding); err != nil {
				blog.Errorf("update node %v cache failed, err %s", req.NamespacedName, err.Error())
				return ctrl.Result{
					Requeue:      true,
					RequeueAfter: 3 * time.Second,
				}, nil
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager set reconciler
func (pbr *PortBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&networkextensionv1.PortBinding{}).
		WithEventFilter(pbr.getPortBindingPredicate()).
		Complete(pbr)
}

func (pbr *PortBindingReconciler) getPortBindingPredicate() predicate.Predicate {
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
				newPoolBinding.Status.Status == oldPoolBinding.Status.Status {
				blog.V(5).Infof("portbinding %+v updated, but spec not change", newPoolBinding)
				return false
			}
			return true
		},
	}
}
