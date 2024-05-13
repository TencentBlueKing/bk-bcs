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

// Package namespacecontroller add configmap to specified namespace
package namespacecontroller

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/portbindingcontroller"
)

// NamespaceReconciler reconciler for namespace
type NamespaceReconciler struct {
	ctx       context.Context
	k8sClient client.Client

	nodeBindCache *portbindingcontroller.NodePortBindingCache
}

// NewNamespaceReconciler create NamespaceReconciler
func NewNamespaceReconciler(ctx context.Context, k8sClient client.Client,
	nodeBindCache *portbindingcontroller.NodePortBindingCache) *NamespaceReconciler {
	return &NamespaceReconciler{
		ctx:           ctx,
		k8sClient:     k8sClient,
		nodeBindCache: nodeBindCache,
	}
}

// Reconcile reconcile port pool
// portbinding name is same with pod name
func (nr *NamespaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	blog.V(3).Infof("namespace %+v triggered", req.NamespacedName)
	ns := &k8scorev1.Namespace{}
	if err := nr.k8sClient.Get(nr.ctx, req.NamespacedName, ns); err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Infof("not found ns'%s/%s'", req.Namespace, req.Name)
			return ctrl.Result{}, err
		}
		blog.Warnf("get ns failed, err: %v", err)
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}

	configMap := &k8scorev1.ConfigMap{}
	if err := nr.k8sClient.Get(nr.ctx, k8stypes.NamespacedName{Namespace: ns.Name,
		Name: networkextensionv1.NodePortBindingConfigMapName},
		configMap); err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Infof("not found configMap '%s/%s'", req.Namespace, networkextensionv1.NodePortBindingConfigMapName)
			return nr.createNodeConfigMap(ns)
		}
		blog.Warnf("get configMap'%s/%s' failed, err: %v", ns.Name, networkextensionv1.NodePortBindingConfigMapName,
			err)
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}

	if !checkNodeBindLabel(ns) {
		if err := nr.k8sClient.Delete(nr.ctx, configMap); err != nil {
			blog.Warnf("delete configmap '%s/%s' failed, err %v", err)
			return ctrl.Result{
				Requeue:      true,
				RequeueAfter: 3 * time.Second,
			}, nil
		}
	}
	return ctrl.Result{}, nil
}

func (nr *NamespaceReconciler) createNodeConfigMap(namespace *k8scorev1.Namespace) (ctrl.Result, error) {
	if !checkNodeBindLabel(namespace) {
		blog.Infof("namespace %s not have label, do not create configmap")
		return ctrl.Result{}, nil
	}
	configMap := &k8scorev1.ConfigMap{}
	configMap.SetNamespace(namespace.Name)
	configMap.SetName(networkextensionv1.NodePortBindingConfigMapName)

	configMap.Data = nr.nodeBindCache.GetCache()

	if err := nr.k8sClient.Create(nr.ctx, configMap); err != nil {
		blog.Errorf("create config map in namespace[%s] failed, err %s", namespace.Name, err.Error())
		return ctrl.Result{
			Requeue:      true,
			RequeueAfter: 3 * time.Second,
		}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager set reconciler
func (nr *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8scorev1.Namespace{}).
		WithEventFilter(getNamespacePredicate()).
		Complete(nr)
}

func getNamespacePredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			ns, ok := e.Object.(*k8scorev1.Namespace)
			if ok && checkNodeBindLabel(ns) {
				return true
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			newNs, okNew := e.ObjectNew.(*k8scorev1.Namespace)
			oldNs, okOld := e.ObjectOld.(*k8scorev1.Namespace)
			if !okNew || !okOld {
				return true
			}
			if checkNodeBindLabel(oldNs) && !checkNodeBindLabel(newNs) {
				return true
			}
			if checkNodeBindLabel(newNs) && !checkNodeBindLabel(oldNs) {
				return true
			}
			blog.V(3).Infof("namespace '%s' have not configmap label", newNs.Name)
			return false
		},
	}
}
