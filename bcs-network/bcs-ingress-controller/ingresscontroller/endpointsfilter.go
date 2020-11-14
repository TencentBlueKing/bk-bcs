/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ingresscontroller

import (
	"context"

	k8scorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/metrics"
)

// EndpointsFilter filter for endpoints event
type EndpointsFilter struct {
	filterName string
	cli        client.Client
}

// NewEndpointsFilter create endpoints filter
func NewEndpointsFilter(cli client.Client) *EndpointsFilter {
	return &EndpointsFilter{
		filterName: "endpoints",
		cli:        cli,
	}
}

func (ef *EndpointsFilter) enqueueEndpointsRelatedIngress(eps *k8scorev1.Endpoints, q workqueue.RateLimitingInterface) {
	ingressList := &networkextensionv1.IngressList{}
	err := ef.cli.List(context.TODO(), ingressList, &client.ListOptions{})
	if err != nil {
		blog.Warnf("list bcs ingresses failed, err %s", err.Error())
		return
	}
	ingresses := findIngressesByService(eps.GetName(), eps.GetNamespace(), ingressList)
	if len(ingresses) == 0 {
		return
	}
	for _, ingress := range ingresses {
		q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      ingress.GetName(),
			Namespace: ingress.GetNamespace(),
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
