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

package proxy

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/clientutil"
)

type responder struct{}

// Error implements k8s.io/apimachinery/pkg/util/proxy.ErrorResponder
func (r *responder) Error(w http.ResponseWriter, req *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// makeTarget 提取连接地址
func makeTarget(serverAddress string) (*url.URL, error) {
	if !strings.HasSuffix(serverAddress, "/") {
		serverAddress = serverAddress + "/"
	}
	target, err := url.Parse(serverAddress)
	if err != nil {
		return nil, err
	}
	return target, nil
}

// makeUpgradeTransport creates a transport for proxy connections that must upgrade.
// reference implementation https://github.com/kubernetes/kubectl/blob/master/pkg/proxy/proxy_server.go#L153 and remove tlsConfig
func makeUpgradeTransport(config *rest.Config) (proxy.UpgradeRequestRoundTripper, error) {
	transportConfig, err := config.TransportConfig()
	if err != nil {
		return nil, err
	}

	// 添加 BearerToken 等鉴权
	upgrader, err := transport.HTTPWrappersForConfig(transportConfig, proxy.MirrorRequest)
	if err != nil {
		return nil, err
	}

	// config.Transport is don't matter, only use upgrader
	return proxy.NewUpgradeRequestRoundTripper(config.Transport, upgrader), nil
}

// makeUpgradeAwareHandler creates a new proxy handler for an kube-apiserver
func makeUpgradeAwareHandler(config *rest.Config) (*proxy.UpgradeAwareHandler, error) {
	target, err := makeTarget(config.Host)
	if err != nil {
		return nil, err
	}
	apiTransport, err := rest.TransportFor(config)
	if err != nil {
		return nil, err
	}

	upgradeTransport, err := makeUpgradeTransport(config)
	if err != nil {
		return nil, err
	}

	apiProxy := proxy.NewUpgradeAwareHandler(target, apiTransport, false, false, &responder{})
	apiProxy.UpgradeTransport = upgradeTransport
	apiProxy.UseRequestLocation = true
	apiProxy.AppendLocationPath = true
	return apiProxy, nil
}

// ProxyHandler 代理请求
type ProxyHandler struct {
	handler *proxy.UpgradeAwareHandler
	config  *rest.Config
}

// NewProxyHandler
func NewProxyHandler(clusterId string) (*ProxyHandler, error) {
	kubeConf, err := clientutil.GetKubeConfByClusterId(clusterId)
	if err != nil {
		return nil, errors.Wrapf(err, "build %s proxy handler", clusterId)
	}

	proxyHandler, err := makeUpgradeAwareHandler(kubeConf)
	if err != nil {
		return nil, errors.Wrapf(err, "build %s proxy handler from config %s", clusterId, kubeConf)
	}

	handler := &ProxyHandler{
		config:  kubeConf,
		handler: proxyHandler,
	}
	return handler, nil
}

// ServeHTTP
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// exec 需要 Upgrade
	if req.Header.Get("X-Stream-Protocol-Version") != "" {
		h.handler.UpgradeRequired = true
	}

	// 代理请求处理
	h.handler.ServeHTTP(w, req)
}
