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

// Package ingresscontroller xxx
package ingresscontroller

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/ingresscache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

// EndpointsFilter filter for endpoints event
type EndpointsFilter struct {
	filterName   string
	cli          client.Client
	ingressCache ingresscache.IngressCache
}

// NewEndpointsFilter create endpoints filter
func NewEndpointsFilter(cli client.Client, endpointsCache *ingresscache.Cache) *EndpointsFilter {
	return &EndpointsFilter{
		filterName:   "endpoints",
		cli:          cli,
		ingressCache: endpointsCache,
	}
}

func (ef *EndpointsFilter) enqueueEndpointsRelatedIngress(eps *k8scorev1.Endpoints, q workqueue.RateLimitingInterface) {
	ingressMetas := ef.ingressCache.GetRelatedIngressOfService(networkextensionv1.ServiceKindNativeService,
		eps.GetNamespace(), eps.GetName())
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
func (ef *EndpointsFilter) Create(e event.CreateEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(ef.filterName, metrics.EventTypeAdd)

	eps, ok := e.Object.(*k8scorev1.Endpoints)
	if !ok {
		blog.Warnf("recv create object is not Endpoints, event %+v", e)
		return
	}
	ef.enqueueEndpointsRelatedIngress(eps, q)
}

// Update implement EventFilter
func (ef *EndpointsFilter) Update(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(ef.filterName, metrics.EventTypeUpdate)

	eps, ok := e.ObjectNew.(*k8scorev1.Endpoints)
	if !ok {
		blog.Warnf("recv update object is not Endpoints, event %+v", e)
		return
	}
	ef.enqueueEndpointsRelatedIngress(eps, q)
}

// Delete implement EventFilter
func (ef *EndpointsFilter) Delete(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(ef.filterName, metrics.EventTypeDelete)

	eps, ok := e.Object.(*k8scorev1.Endpoints)
	if !ok {
		blog.Warnf("recv delete object is not Endpoints, event %+v", e)
		return
	}
	ef.enqueueEndpointsRelatedIngress(eps, q)
}

// Generic implement EventFilter
func (ef *EndpointsFilter) Generic(e event.GenericEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(ef.filterName, metrics.EventTypeUnknown)

	eps, ok := e.Object.(*k8scorev1.Endpoints)
	if !ok {
		blog.Warnf("recv generic object is not Endpoints, event %+v", e)
		return
	}
	ef.enqueueEndpointsRelatedIngress(eps, q)
}
