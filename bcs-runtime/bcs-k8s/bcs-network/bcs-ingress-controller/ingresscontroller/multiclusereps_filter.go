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

package ingresscontroller

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	federationv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/federation/v1"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/ingresscache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

// MultiClusterEpsFilter filter for MultiClusterEndpointSlice event
type MultiClusterEpsFilter struct {
	filterName   string
	cli          client.Client
	ingressCache ingresscache.IngressCache
}

// NewMultiClusterEpsFilter create MultiClusterEndpoints filter
func NewMultiClusterEpsFilter(cli client.Client, ingressCache ingresscache.IngressCache) *MultiClusterEpsFilter {
	return &MultiClusterEpsFilter{
		filterName:   "MultiClusterEndpoints",
		cli:          cli,
		ingressCache: ingressCache,
	}
}

func (m *MultiClusterEpsFilter) enqueueRelatedIngress(eps *federationv1.MultiClusterEndpointSlice,
	q workqueue.RateLimitingInterface) {
	ingressMetas := m.ingressCache.GetRelatedIngressOfService(networkextensionv1.ServiceKindMultiClusterService,
		eps.GetRelatedServiceNameSpace(), eps.GetRelatedServiceName())

	if len(ingressMetas) == 0 {
		return
	}
	for _, meta := range ingressMetas {
		q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      meta.Name,
			Namespace: meta.Namespace,
		}})
	}
}

// Create implement EventFilter
func (m *MultiClusterEpsFilter) Create(e event.CreateEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(m.filterName, metrics.EventTypeAdd)

	mEp, ok := e.Object.(*federationv1.MultiClusterEndpointSlice)
	if !ok {
		blog.Warnf("recv create object is not MultiClusterEndpointSlice, event %+v", e)
		return
	}
	m.enqueueRelatedIngress(mEp, q)
}

// Update implement EventFilter
func (m *MultiClusterEpsFilter) Update(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(m.filterName, metrics.EventTypeUpdate)

	mEpNew, ok := e.ObjectNew.(*federationv1.MultiClusterEndpointSlice)
	if !ok {
		blog.Warnf("recv new object is not MultiClusterEndpointSlice, event %+v", e)
		return
	}
	m.enqueueRelatedIngress(mEpNew, q)
}

// Delete implement EventFilter
func (m *MultiClusterEpsFilter) Delete(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(m.filterName, metrics.EventTypeDelete)

	mEp, ok := e.Object.(*federationv1.MultiClusterEndpointSlice)
	if !ok {
		blog.Warnf("recv delete object is not MultiClusterEndpointSlice, event %+v", e)
		return
	}
	m.enqueueRelatedIngress(mEp, q)
}

// Generic implement EventFilter
func (m *MultiClusterEpsFilter) Generic(e event.GenericEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(m.filterName, metrics.EventTypeUnknown)

	mEp, ok := e.Object.(*federationv1.MultiClusterEndpointSlice)
	if !ok {
		blog.Warnf("recv Generic object is not MultiClusterEndpointSlice, event %+v", e)
		return
	}
	m.enqueueRelatedIngress(mEp, q)
}
