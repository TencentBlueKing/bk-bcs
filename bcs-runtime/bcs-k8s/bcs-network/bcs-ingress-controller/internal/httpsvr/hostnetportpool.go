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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/hostnetportpoolcache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// writeResponse writes data to the HTTP response. Write errors are logged but
// not propagated because they typically indicate the client has disconnected,
// and the go-restful handler signature does not support returning an error.
func writeResponse(response *restful.Response, data []byte) {
	if _, err := response.Write(data); err != nil {
		blog.Warnf("hostnetportpool api: write response failed: %v", err)
	}
}

func (h *HttpServerClient) getHostNetPortPoolBindingResult(
	request *restful.Request, response *restful.Response) {

	startTime := time.Now()
	mf := func(status string) {
		metrics.ReportAPIRequestMetric("get_hostnet_binding_result", "GET", status, startTime)
	}

	podName := request.QueryParameter("podName")
	podNamespace := request.QueryParameter("podNamespace")

	if podName == "" || podNamespace == "" {
		mf(strconv.Itoa(http.StatusBadRequest))
		writeResponse(response, CreateResponseData(
			fmt.Errorf("empty parameter: both podName and podNamespace are required"), "", nil))
		return
	}

	pod := &k8scorev1.Pod{}
	if err := h.Mgr.GetClient().Get(context.Background(), types.NamespacedName{
		Namespace: podNamespace, Name: podName,
	}, pod); err != nil {
		if k8serrors.IsNotFound(err) {
			mf(strconv.Itoa(http.StatusNotFound))
			writeResponse(response, CreateResponseDataWithCode(
				http.StatusNotFound,
				fmt.Errorf("pod %s/%s not found", podNamespace, podName)))
			return
		}
		blog.Errorf("hostnetportpool api: get pod %s/%s failed: %v", podNamespace, podName, err)
		mf(strconv.Itoa(http.StatusInternalServerError))
		writeResponse(response, CreateResponseData(
			fmt.Errorf("get pod failed: %v", err), "", nil))
		return
	}

	if _, ok := pod.Annotations[constant.AnnotationForHostNetPortPool]; !ok {
		mf(strconv.Itoa(http.StatusNotFound))
		writeResponse(response, CreateResponseDataWithCode(
			http.StatusNotFound,
			fmt.Errorf("pod %s/%s exists but not using HostNetPortPool", podNamespace, podName)))
		return
	}

	status := pod.Annotations[constant.AnnotationForHostNetPortPoolBindingStatus]
	if status == "" {
		status = "NotReady"
	}

	type bindingResultResponse struct {
		Status string      `json:"status"`
		Result interface{} `json:"result"`
	}

	resp := &bindingResultResponse{Status: status}

	if status == "Ready" {
		resultStr := pod.Annotations[constant.AnnotationForHostNetPortPoolBindingResult]
		if resultStr != "" {
			var result hostnetportpoolcache.HostNetPortPoolBindingResult
			if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
				blog.Errorf("hostnetportpool api: failed to parse result for pod %s/%s: %v",
					podNamespace, podName, err)
				mf(strconv.Itoa(http.StatusInternalServerError))
				writeResponse(response, CreateResponseData(
					fmt.Errorf("corrupted binding result annotation: %v", err), "", nil))
				return
			}
			resp.Result = &result
		}
	}

	mf(strconv.Itoa(http.StatusOK))
	writeResponse(response, CreateResponseData(nil, "", resp))
}
