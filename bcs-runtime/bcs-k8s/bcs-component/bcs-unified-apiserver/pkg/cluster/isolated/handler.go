/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package isolated

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/rest"
)

// Handler isolated cluster hander
type Handler struct {
	clusterId    string
	proxyHandler *proxy.ProxyHandler
}

// NewHandler create isolated cluster handler
func NewHandler(clusterId string) (*Handler, error) {
	proxyHandler, err := proxy.NewProxyHandler(clusterId)
	if err != nil {
		return nil, err
	}

	return &Handler{
		clusterId:    clusterId,
		proxyHandler: proxyHandler,
	}, nil
}

// Serve 目前是直接透明代理
func (h *Handler) Serve(c *rest.RequestContext) {
	h.proxyHandler.ServeHTTP(c.Writer, c.Request)
}
