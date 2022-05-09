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
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/transport"
)

type responder struct{}

// Error implements k8s.io/apimachinery/pkg/util/proxy.ErrorResponder
func (r *responder) Error(w http.ResponseWriter, req *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func extractIPAddress(serverAddress string) (*url.URL, error) {
	if !strings.HasSuffix(serverAddress, "/") {
		serverAddress = serverAddress + "/"
	}
	ipAddress, err := url.Parse(serverAddress)
	if err != nil {
		return nil, err
	}
	return ipAddress, nil
}

// makeUpgradeTransport creates a transport that explicitly bypasses HTTP2 support
// for proxy connections that must upgrade.
func makeUpgradeTransport(config *rest.Config, keepalive time.Duration) (proxy.UpgradeRequestRoundTripper, error) {
	transportConfig, err := config.TransportConfig()
	if err != nil {
		return nil, err
	}
	tlsConfig, err := transport.TLSConfigFor(transportConfig)
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: keepalive,
	}
	rt := utilnet.SetOldTransportDefaults(&http.Transport{
		TLSClientConfig: tlsConfig,
		Dial: func(network, addr string) (net.Conn, error) {
			// resolve domain to real apiserver address
			ipAddress, err := extractIPAddress(config.Host)
			if err != nil {
				return nil, err
			}
			return dialer.Dial(network, ipAddress.Host)
		},
	})

	upgrader, err := transport.HTTPWrappersForConfig(transportConfig, proxy.MirrorRequest)
	if err != nil {
		return nil, err
	}
	return proxy.NewUpgradeRequestRoundTripper(rt, upgrader), nil
}

// NewProxyHandlerFromConfig creates a new proxy handler for an kube-apiserver
func NewProxyHandlerFromConfig(config *rest.Config) (*proxy.UpgradeAwareHandler, error) {
	target, err := extractIPAddress(config.Host)
	if err != nil {
		return nil, err
	}
	apiTransport, err := rest.TransportFor(config)
	if err != nil {
		return nil, err
	}
	upgradeTransport, err := makeUpgradeTransport(config, 0)
	if err != nil {
		return nil, err
	}
	apiProxy := proxy.NewUpgradeAwareHandler(target, apiTransport, false, false, &responder{})
	apiProxy.UpgradeTransport = upgradeTransport
	apiProxy.UseRequestLocation = true
	return apiProxy, nil
}
