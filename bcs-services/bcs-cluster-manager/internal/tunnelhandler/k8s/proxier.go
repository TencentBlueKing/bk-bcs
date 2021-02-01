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

package k8s

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// WsTunnel is http tunnel destination
type WsTunnel struct {
	httpTransport *http.Transport
	serverAddress string
	userToken     string
	caCertData    string
}

// TunnelProxyDispatcher is the handler which dispatch and proxy the incoming requests to external kube-apiserver with
// websocket tunnel
type TunnelProxyDispatcher struct {
	// ClusterVarName is the path parameter name of cluster_id
	ClusterVarName string
	// SubPathVarName is the path parameter name of sub-path needs to be forwarded
	SubPathVarName string

	model store.ClusterManagerModel

	tunnelServer *websocketDialer.Server

	// cache for http tunnel info
	wsTunnelStore      map[string]*WsTunnel
	wsTunnelMutateLock sync.RWMutex
}

// ClusterHandlerInstance is http handler instance of certain cluster
type ClusterHandlerInstance struct {
	ServerAddress string
	Handler       http.Handler
}

// NewTunnelProxyDispatcher create a TunnelProxyDispatcher
func NewTunnelProxyDispatcher(
	clusterVarName, subPathVarName string,
	model store.ClusterManagerModel, tunnelServer *websocketDialer.Server) *TunnelProxyDispatcher {
	return &TunnelProxyDispatcher{
		ClusterVarName: clusterVarName,
		SubPathVarName: subPathVarName,
		model:          model,
		tunnelServer:   tunnelServer,
		wsTunnelStore:  make(map[string]*WsTunnel),
	}
}

// ServeHTTP implements http.Handler
func (f *TunnelProxyDispatcher) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	blog.V(3).Infof("xreq %s, host %s, url %s, src %s",
		utils.GetXRequestIDFromHTTPRequest(req), req.Host, req.URL, req.RemoteAddr)
	start := time.Now()
	vars := mux.Vars(req)
	// Get cluster id
	clusterID := vars[f.ClusterVarName]

	var proxyHandler *ClusterHandlerInstance
	// 先从websocket dialer缓存中查找websocket链
	websocketHandler, found, err := f.lookupWsHandler(clusterID)
	if err != nil {
		blog.Errorf("error when lookup websocket conn, err %s", err.Error())
		status := common.NewInternalError(
			fmt.Errorf("error when lookup websocket conn, err %s", err.Error()))
		status.ErrStatus.Reason = common.ErrorStatusCreateTunnel
		common.WriteKubeAPIError(rw, status)
		return
	}
	// if found tunnel, use this tunnel to request to kube-apiserver
	if found {
		blog.Info("found websocket conn for k8s cluster %s", clusterID)
		handlerServer := stripLeaveSlash(f.ExtractPathPrefix(req), websocketHandler)
		proxyHandler = &ClusterHandlerInstance{
			Handler: handlerServer,
		}
		credentials, found, err := f.model.GetClusterCredential(context.TODO(), clusterID)
		if err != nil {
			blog.Errorf("error when get cluster %s credential, err %s", clusterID, err.Error())
			status := common.NewInternalError(
				fmt.Errorf("error when get cluster %s credential, err %s", clusterID, err.Error()))
			status.ErrStatus.Reason = common.ErrorStatusCreateTunnel
			common.WriteKubeAPIError(rw, status)
			return
		}
		if !found {
			blog.Errorf("cluster %s credential not found", clusterID)
			status := common.NewInternalError(
				fmt.Errorf("cluster %s credential not found", clusterID))
			status.ErrStatus.Reason = common.ErrorStatusCreateTunnel
			common.WriteKubeAPIError(rw, status)
			return
		}
		bearerToken := "Bearer " + credentials.UserToken
		req.Header.Set("Authorization", bearerToken)

		// set request scheme
		req.URL.Scheme = "https"

		// if webconsole long request, then set the latency before ServerHTTP
		if websocket.IsWebSocketUpgrade(req) {
			metrics.ReportAPIRequestMetric("k8s_tunnel_request", "websocket", "", start)
		}
		proxyHandler.Handler.ServeHTTP(rw, req)
		if !websocket.IsWebSocketUpgrade(req) {
			metrics.ReportAPIRequestMetric("k8s_tunnel_request", req.Method, "", start)
		}
		return
	}
	status := common.NewNotFoundError(common.GroupResourceCluster, clusterID,
		"no cluster session can be found using given cluster id")
	common.WriteKubeAPIError(rw, status)
	return
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

// ExtractPathPrefix extracts the path prefix which needs to be stripped when the request is forwarded to the reverse
// proxy handler.
func (f *TunnelProxyDispatcher) ExtractPathPrefix(req *http.Request) string {
	subPath := mux.Vars(req)[f.SubPathVarName]
	fullPath := req.URL.Path
	// We need to strip the prefix string before the request can be forward to apiserver, so we will walk over the full
	// request path backwards, everything before the `sub_path` will be the prefix we need to strip
	return fullPath[:len(fullPath)-len(subPath)]
}
