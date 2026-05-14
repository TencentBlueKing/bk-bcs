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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

const hostnetNodeFilterName = "hostnet_node"

// HostNetNodeFilter filters node events. Only Delete events trigger cache cleanup.
type HostNetNodeFilter struct{}

// NewHostNetNodeFilter creates a new node filter.
func NewHostNetNodeFilter() *HostNetNodeFilter {
	return &HostNetNodeFilter{}
}

// Create is a no-op; new nodes are handled via lazy allocation.
func (f *HostNetNodeFilter) Create(_ event.CreateEvent, _ workqueue.RateLimitingInterface) {}

// Update is a no-op; node attribute changes are irrelevant to port allocation.
func (f *HostNetNodeFilter) Update(_ event.UpdateEvent, _ workqueue.RateLimitingInterface) {}

// Delete enqueues a reconcile request with the node name.
func (f *HostNetNodeFilter) Delete(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	node, ok := e.Object.(*k8scorev1.Node)
	if !ok {
		blog.Warnf("hostnetport nodefilter: recv delete object is not Node, event %+v", e)
		return
	}
	metrics.IncreaseEventCounter(hostnetNodeFilterName, metrics.EventTypeDelete)
	if node.Name == "" {
		blog.Warnf("hostnetport nodefilter: node has empty name, skip delete event")
		return
	}
	q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
		Name: node.Name,
	}})
}

// Generic is a no-op.
func (f *HostNetNodeFilter) Generic(_ event.GenericEvent, _ workqueue.RateLimitingInterface) {}
