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

package portbindingcontroller

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

// NodeFilter filter for node event
type NodeFilter struct {
	filterName        string
	cli               client.Client
	nodePortBindingNs string
}

// NewNodeFilter create pod filter
func NewNodeFilter(cli client.Client, nodePortBindingNs string) *NodeFilter {
	return &NodeFilter{
		filterName:        "node",
		cli:               cli,
		nodePortBindingNs: nodePortBindingNs,
	}
}

// Create implement EventFilter
func (nf *NodeFilter) Create(e event.CreateEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(nf.filterName, metrics.EventTypeAdd)

	node, ok := e.Object.(*k8scorev1.Node)
	if !ok {
		blog.Warnf("recv create object is not node, event %+v", e)
		return
	}
	if !checkPortPoolAnnotation(node.Annotations) {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      node.GetName(),
		Namespace: nf.nodePortBindingNs,
	}})

	// check if related portBinding created success
	go checkPortBindingCreate(nf.cli, nf.nodePortBindingNs, node.GetName())
}

// Update implement EventFilter
func (nf *NodeFilter) Update(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(nf.filterName, metrics.EventTypeUpdate)

	oldNode, okOld := e.ObjectOld.(*k8scorev1.Node)
	newNode, okNew := e.ObjectNew.(*k8scorev1.Node)
	if !okOld || !okNew {
		blog.Warnf("recv create object is not Node, event %+v", e)
		return
	}

	if !checkPortPoolAnnotation(newNode.Annotations) && !checkPortPoolAnnotation(oldNode.Annotations) {
		return
	}

	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      newNode.GetName(),
		Namespace: nf.nodePortBindingNs,
	}})

	// 如果删除portpool相关annotation，认为用户不再需要绑定端口，会在portBinding reconcile过程中删除相关PortBinding
	if !checkPortPoolAnnotation(newNode.Annotations) && checkPortPoolAnnotation(oldNode.Annotations) {
		go checkPortBindingDelete(nf.cli, nf.nodePortBindingNs, newNode.GetName())
	}
}

// Delete implement EventFilter
func (nf *NodeFilter) Delete(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(nf.filterName, metrics.EventTypeDelete)

	node, ok := e.Object.(*k8scorev1.Node)
	if !ok {
		blog.Warnf("recv delete object is not Node, event %+v", e)
		return
	}
	if !checkPortPoolAnnotation(node.Annotations) {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      node.GetName(),
		Namespace: nf.nodePortBindingNs,
	}})

	go checkPortBindingDelete(nf.cli, nf.nodePortBindingNs, node.GetName())
}

// Generic implement EventFilter
func (nf *NodeFilter) Generic(e event.GenericEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(nf.filterName, metrics.EventTypeUnknown)
	pod, ok := e.Object.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("recv delete object is not Node, event %+v", e)
		return
	}
	if !checkPortPoolAnnotation(pod.Annotations) {
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name:      pod.GetName(),
		Namespace: nf.nodePortBindingNs,
	}})
}
