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
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/emicklei/go-restful"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

func (h *HttpServerClient) listPortPool(request *restful.Request, response *restful.Response) {
	startTime := time.Now()
	mf := func(status string) {
		metrics.ReportAPIRequestMetric("list_port_pool", "GET", status, startTime)
	}

	poolList := &networkextensionv1.PortPoolList{}
	if err := h.Mgr.GetClient().List(context.Background(), poolList); err != nil {
		blog.Errorf("list port pool failed when collect metrics, err %s", err.Error())
		_, _ = response.Write(CreateResponseData(fmt.Errorf("list port pool failed, err: %s", err.Error()), "", nil))
		mf(strconv.Itoa(http.StatusInternalServerError))
		return
	}
	mf(strconv.Itoa(http.StatusOK))
	data := CreateResponseData(nil, "success", poolList)
	_, _ = response.Write(data)
}

func (h *HttpServerClient) getNodePortBindings(request *restful.Request, response *restful.Response) {
	startTime := time.Now()
	mf := func(status string) {
		metrics.ReportAPIRequestMetric("get_node_port_binding", "GET", status, startTime)
	}
	cache := h.NodePortBindCache.GetCache()
	nodes := strings.TrimSpace(request.QueryParameter("nodes"))
	if nodes == "" {
		_, _ = response.Write(CreateResponseData(nil, "success", cache))
		mf(strconv.Itoa(http.StatusOK))
		return
	}

	nodesArr := strings.Split(nodes, ",")
	newMap := make(map[string]string)
	for i := range nodesArr {
		node := nodesArr[i]
		v, ok := cache[node]
		if ok {
			newMap[node] = v
		}
	}
	mf(strconv.Itoa(http.StatusOK))
	_, _ = response.Write(CreateResponseData(nil, "success", newMap))
}

// 获取客户端 IP 地址的函数
func getClientIP(r *http.Request) string {
	// 尝试从 X-Forwarded-For 头获取 IP 地址
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// 如果有多个 IP 地址，取第一个
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从 X-Real-Ip 头获取 IP 地址
	ip = r.Header.Get("X-Real-Ip")
	if ip != "" {
		return ip
	}

	// 最后从 RemoteAddr 获取 IP 地址
	ip = r.RemoteAddr
	// 去掉端口号
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}
