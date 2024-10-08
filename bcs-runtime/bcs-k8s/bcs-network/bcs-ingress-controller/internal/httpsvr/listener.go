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

package httpsvr

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/emicklei/go-restful"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	k8sapitypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func (h *HttpServerClient) listListener(request *restful.Request, response *restful.Response) {
	startTime := time.Now()
	mf := func(status string) {
		metrics.ReportAPIRequestMetric("list_listeners", "GET", status, startTime)
	}
	var data []byte
	switch request.PathParameter("condition") {
	case "ingress":
		ingress := &networkextensionv1.Ingress{}
		if err := h.Mgr.GetClient().Get(context.Background(), k8sapitypes.NamespacedName{
			Name:      request.PathParameter("name"),
			Namespace: request.PathParameter("namespace"),
		}, ingress); err != nil {
			if k8serrors.IsNotFound(err) {
				blog.Infof("ingress %s/%s not found", request.PathParameter("namespace"), request.PathParameter("name"))
				// NOCC:ineffassign/assign(误报)
				mf(strconv.Itoa(http.StatusInternalServerError))
				data = CreateResponseData(err, "failed", nil)
				break
			}
		}
		existedListenerList := &networkextensionv1.ListenerList{}
		selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(k8slabels.Set(map[string]string{
			ingress.Name: networkextensionv1.LabelValueForIngressName,
			networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
		})))
		err = h.Mgr.GetClient().List(context.TODO(), existedListenerList, &client.ListOptions{
			Namespace:     ingress.Namespace,
			LabelSelector: selector})
		if err != nil {
			blog.Errorf("list listeners filter by ingress %s failed, err %s",
				request.PathParameter("name"), err.Error())
			mf(strconv.Itoa(http.StatusInternalServerError))
			// NOCC:ineffassign/assign(误报)
			data = CreateResponseData(err, "failed", nil)
			break
		}
		data = CreateResponseData(nil, "success", existedListenerList)
		mf(strconv.Itoa(http.StatusOK))
	case "portpool":
		portPool := &networkextensionv1.PortPool{}
		if err := h.Mgr.GetClient().Get(context.Background(), k8sapitypes.NamespacedName{
			Name:      request.PathParameter("name"),
			Namespace: request.PathParameter("namespace"),
		}, portPool); err != nil {
			if k8serrors.IsNotFound(err) {
				blog.Infof("portpool %s/%s not found", request.PathParameter("namespace"), request.PathParameter("name"))
				mf(strconv.Itoa(http.StatusInternalServerError))
				// NOCC:ineffassign/assign(误报)
				data = CreateResponseData(err, "failed", nil)
				break
			}
		}
		result := make(map[string]*networkextensionv1.ListenerList, 0)
		for i := range portPool.Spec.PoolItems {
			existedListenerList := &networkextensionv1.ListenerList{}
			selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(k8slabels.Set(map[string]string{
				portPool.Spec.PoolItems[i].ItemName:             networkextensionv1.LabelValueForPortPoolItemName,
				networkextensionv1.LabelKeyForIsSegmentListener: networkextensionv1.LabelValueFalse,
			})))
			err = h.Mgr.GetClient().List(context.TODO(), existedListenerList, &client.ListOptions{
				Namespace:     request.PathParameter("namespace"),
				LabelSelector: selector})
			if err != nil {
				blog.Errorf("list listeners filter by port pool item %s failed, err %s",
					portPool.Spec.PoolItems[i].ItemName, err.Error())
				mf(strconv.Itoa(http.StatusInternalServerError))
				// NOCC:ineffassign/assign(误报)
				data = CreateResponseData(err, "failed", nil)
				break
			}
			result[portPool.Spec.PoolItems[i].ItemName] = existedListenerList
		}
		data = CreateResponseData(nil, "success", result)
		mf(strconv.Itoa(http.StatusOK))
	}

	_, _ = response.Write(data)
}
