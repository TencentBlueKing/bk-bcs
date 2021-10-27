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
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// ServiceFilter filter for service event
type ServiceFilter struct {
	filterName string
	cli        client.Client
}

// NewServiceFilter create service filter
func NewServiceFilter(cli client.Client) *ServiceFilter {
	return &ServiceFilter{
		filterName: "service",
		cli:        cli,
	}
}

func (sf *ServiceFilter) enqueueServiceRelatedIngress(svc *k8scorev1.Service, q workqueue.RateLimitingInterface) {
	ingressList := &networkextensionv1.IngressList{}
	err := sf.cli.List(context.TODO(), ingressList, &client.ListOptions{})
	if err != nil {
		blog.Warnf("list bcs ingresses failed, err %s", err.Error())
		return
	}
	ingresses := findIngressesByService(svc.GetName(), svc.GetNamespace(), ingressList)
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
func (sf *ServiceFilter) Create(e event.CreateEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(sf.filterName, metrics.EventTypeAdd)

	svc, ok := e.Object.(*k8scorev1.Service)
	if !ok {
		blog.Warnf("recv create object is not Service, event %+v", e)
		return
	}
	sf.enqueueServiceRelatedIngress(svc, q)
}

// Update implement EventFilter
func (sf *ServiceFilter) Update(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(sf.filterName, metrics.EventTypeUpdate)

	svc, ok := e.ObjectNew.(*k8scorev1.Service)
	if !ok {
		blog.Warnf("recv update object is not Service, event %+v", e)
		return
	}
	sf.enqueueServiceRelatedIngress(svc, q)
}

// Delete implement EventFilter
func (sf *ServiceFilter) Delete(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(sf.filterName, metrics.EventTypeDelete)

	svc, ok := e.Object.(*k8scorev1.Service)
	if !ok {
		blog.Warnf("recv delete object is not Service, event %+v", e)
		return
	}
	sf.enqueueServiceRelatedIngress(svc, q)
}

// Generic implement EventFilter
func (sf *ServiceFilter) Generic(e event.GenericEvent, q workqueue.RateLimitingInterface) {
	metrics.IncreaseEventCounter(sf.filterName, metrics.EventTypeUnknown)

	if e.Meta == nil {
		blog.Infof("GenericEvent received with no metadata, event %+v", e)
		return
	}
}
