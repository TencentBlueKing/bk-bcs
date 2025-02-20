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
	"fmt"
	"net/http"
	"strconv"
	"time"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	restful "github.com/emicklei/go-restful"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
)

// CheckBindStatus check bind status for ingress
func (h *HttpServerClient) CheckBindStatus(request *restful.Request, response *restful.Response) {
	startTime := time.Now()
	mf := func(status string) {
		metrics.ReportAPIRequestMetric("check_bind_status", "GET", status, startTime)
	}

	ingressName := request.QueryParameter("ingress_name")
	ingressNamespace := request.QueryParameter("ingress_namespace")

	ingress := &networkextensionv1.Ingress{}
	if err := h.Mgr.GetClient().Get(request.Request.Context(), k8stypes.NamespacedName{
		Namespace: ingressNamespace,
		Name:      ingressName,
	}, ingress); err != nil {
		mf(strconv.Itoa(http.StatusInternalServerError))
		_ = response.WriteErrorString(http.StatusInternalServerError,
			fmt.Sprintf("get ingress '%s/%s' failed, err: %s", ingressNamespace, ingressName, err.Error()))
		return
	}

	finish, err := h.IngressLiConverter.CheckIngressUpdateFinish(ingress)
	if err != nil {
		mf(strconv.Itoa(http.StatusInternalServerError))
		_ = response.WriteErrorString(http.StatusInternalServerError,
			fmt.Sprintf("check ingress '%s/%s' update finish failed, err: %s", ingressNamespace, ingressName,
				err.Error()))
		return
	}
	mf(strconv.Itoa(http.StatusOK))
	if !finish {
		_ = response.WriteErrorString(http.StatusInternalServerError, "ingress update not finish")
		return
	}
	_, _ = response.Write(CreateResponseData(nil, "ingress update finished", nil))
	return
}
