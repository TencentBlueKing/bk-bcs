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

// Package nodecontroller node controller
package nodecontroller

import (
	"context"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloudnode"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/nodecache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/option"
)

const (
	requeueAfter = 5 * time.Second
)

// NodeReconciler reconciler for node
type NodeReconciler struct {
	ctx       context.Context
	k8sClient client.Client
	eventer   record.EventRecorder
	opts      *option.ControllerOption

	nodeCache  *nodecache.NodeCache
	nodeClient cloudnode.NodeClient
}

// NewNodeReconciler return new Node Reconciler
func NewNodeReconciler(ctx context.Context, k8sclient client.Client,
	eventer record.EventRecorder, opts *option.ControllerOption, cache *nodecache.NodeCache,
	nodeClient cloudnode.NodeClient) *NodeReconciler {
	return &NodeReconciler{
		ctx:       ctx,
		k8sClient: k8sclient,
		eventer:   eventer,
		opts:      opts,

		nodeCache:  cache,
		nodeClient: nodeClient,
	}
}

// Reconcile reconcile node
func (r *NodeReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	blog.Infof("Node %+v triggered", req.NamespacedName)
	node := &corev1.Node{}
	if err := r.k8sClient.Get(r.ctx, req.NamespacedName, node); err != nil {
		if k8serrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}

		blog.Errorf("get node %+v failed, err: %s", req.NamespacedName, err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
	}

	externalIpList, err := r.nodeClient.GetNodeExternalIpList(node)
	if err != nil {
		blog.Warnf("get node[%s] external ip list failed, err: %s", node.GetName(), err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter}, err
	}

	blog.Infof("update node %s ip list: %s", node.Name, externalIpList)
	r.nodeCache.SetNodeIps(*node, externalIpList)

	return ctrl.Result{}, nil
}

// SetupWithManager set reconciler
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		WithEventFilter(r.getNodePredicate()).
		Complete(r)
}

func (r *NodeReconciler) getNodePredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(createEvent event.CreateEvent) bool {
			return r.opts.NodeInfoExporterOpen
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			if !r.opts.NodeInfoExporterOpen {
				return false
			}
			newNode, okNew := event.ObjectNew.(*corev1.Node)
			oldNode, okOld := event.ObjectOld.(*corev1.Node)
			if !okNew || !okOld {
				blog.Warnf("unknown event, new:%+v, old: %+v", event.ObjectNew, event.ObjectOld)
				return true
			}

			if !reflect.DeepEqual(newNode.Status.Addresses, oldNode.Status.Addresses) {
				return true
			}

			return false
		},
		DeleteFunc: func(deleteEvent event.DeleteEvent) bool {
			return r.opts.NodeInfoExporterOpen
		},
		GenericFunc: func(genericEvent event.GenericEvent) bool {
			return r.opts.NodeInfoExporterOpen
		},
	}
}
