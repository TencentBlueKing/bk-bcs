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

package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/tunnel"

	"github.com/gorilla/mux"
	"k8s.io/apimachinery/pkg/util/proxy"
)

// NewWsTunnelDispatcher return a new WsTunnelDispatcher
func NewWsTunnelDispatcher(
	subPathVarName string,
	opt *options.ProxyOptions,
	serverCallBack *tunnel.WsTunnelServerCallback) *WsTunnelDispatcher {
	return &WsTunnelDispatcher{
		subPathVarName: subPathVarName,
		opt:            opt,
		serverCallBack: serverCallBack,
		tunnelServer:   serverCallBack.GetTunnelServer(),
		wsTunnels:      make(map[string]*WsTunnel),
	}
}

// WsTunnelDispatcher describe the dispatcher for routing to tunnels
type WsTunnelDispatcher struct {
	// subPathVarName is the path parameter name of sub-path needs to be forwarded
	subPathVarName string

	opt *options.ProxyOptions

	tunnelServer   *websocketDialer.Server
	serverCallBack *tunnel.WsTunnelServerCallback

	wsTunnelMutex sync.Mutex
	wsTunnels     map[string]*WsTunnel
}

// ServeHTTP implements http.Handler
func (w *WsTunnelDispatcher) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	clusterID := w.opt.Tunnel.ManagedClusterID

	handler, err := w.LookupWsHandler(clusterID)
	if err != nil {
		blog.Errorf("look up websocket handler for cluster %s failed, %v", clusterID, err)
		_, _ = rw.Write([]byte("error: " + err.Error()))
		return
	}

	handlerServer := stripLeaveSlash(w.ExtractPathPrefix(req), handler)
	handlerServer.ServeHTTP(rw, req)
	return
}

// ExtractPathPrefix extracts the path prefix which needs to be stripped when the request is forwarded to the reverse
// proxy handler.
func (w *WsTunnelDispatcher) ExtractPathPrefix(req *http.Request) string {
	subPath := mux.Vars(req)[w.subPathVarName]
	fullPath := req.URL.Path

	// We need to strip the prefix string before the request can be forward to apiserver, so we will walk over the full
	// request path backwards, everything before the `sub_path` will be the prefix we need to strip
	return fullPath[:len(fullPath)-len(subPath)]
}

// LookupWsHandler according to given clusterID, get the target cluster's ws-handler and error
func (w *WsTunnelDispatcher) LookupWsHandler(clusterID string) (*proxy.UpgradeAwareHandler, error) {
	clusterInfo, err := w.serverCallBack.GetClusterInfo(clusterID)
	if err != nil {
		blog.Errorf("get cluster info for %s failed, %v", clusterID, err)
		return nil, err
	}

	tp, err := w.getTransport(clusterID)
	if err != nil {
		blog.Errorf("get transport for clusterID %s failed, %v", clusterID, err)
		return nil, err
	}

	serverAddress := clusterInfo.ServerAddress
	if !strings.HasSuffix(serverAddress, "/") {
		serverAddress = serverAddress + "/"
	}
	u, err := url.Parse(serverAddress)
	if err != nil {
		blog.Errorf("parse server address url for clusterID %s failed, %v", clusterID, err)
		return nil, err
	}

	responder := &responder{}
	proxyHandler := proxy.NewUpgradeAwareHandler(u, tp, true, false, responder)
	proxyHandler.UseRequestLocation = true

	return proxyHandler, nil
}

func (w *WsTunnelDispatcher) getTransport(clusterID string) (http.RoundTripper, error) {
	if !w.tunnelServer.HasSession(clusterID) {
		return nil, fmt.Errorf("session %s not found in tunnel servers", clusterID)
	}

	clusterInfo, err := w.serverCallBack.GetClusterInfo(clusterID)
	if err != nil {
		blog.Errorf("get cluster info for %s failed, %v", clusterID, err)
		return nil, err
	}

	w.wsTunnelMutex.Lock()
	defer w.wsTunnelMutex.Unlock()

	wsTunnel, ok := w.wsTunnels[clusterID]
	if ok && !w.serverAddressChanged(wsTunnel.serverAddress, clusterInfo.ServerAddress) {
		return wsTunnel.httpTransport, nil
	}

	dialer := w.tunnelServer.Dialer(clusterID, 15*time.Second)
	transport := &http.Transport{
		MaxIdleConnsPerHost: 10,
		DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			return dialer(network, addr)
		},
	}

	// if tunnel exist, close the old one
	if ok {
		wsTunnel.httpTransport.CloseIdleConnections()
	}
	w.wsTunnels[clusterID] = &WsTunnel{
		httpTransport: transport,
		serverAddress: clusterInfo.ServerAddress,
	}

	return transport, nil
}

func (w *WsTunnelDispatcher) getClientKey(clusterID string) string {
	return clusterID
}

func (w *WsTunnelDispatcher) serverAddressChanged(oldAddress, newAddress string) bool {
	return oldAddress != newAddress
}

// WsTunnel describe the websocket tunnel
type WsTunnel struct {
	httpTransport *http.Transport
	serverAddress string
}

// responder implements k8s.io/apimachinery/pkg/util/proxy.ErrorResponder
type responder struct{}

// Error implements ErrorResponder Error function
func (r *responder) Error(w http.ResponseWriter, _ *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// like http.StripPrefix, but always leaves an initial slash. (so that our
// regexps will work.)
func stripLeaveSlash(prefix string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		blog.Debug(fmt.Sprintf("begin proxy for: %s", req.URL.Path))
		p := strings.TrimPrefix(req.URL.Path, prefix)
		if len(p) >= len(req.URL.Path) {
			http.NotFound(w, req)
			return
		}
		if len(p) > 0 && p[:1] != "/" {
			p = "/" + p
		}
		req.URL.Path = p
		h.ServeHTTP(w, req)
	})
}
