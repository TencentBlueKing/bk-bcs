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
 *
 */

package k8s

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var DefaultTunnelProxyDispatcher = NewTunnelProxyDispatcher("cluster_id", "sub_path")

// TunnelProxyDispatcher is the handler which dispatch and proxy the incoming requests to external kube-apiserver with websocket tunnel
type TunnelProxyDispatcher struct {
	// ClusterVarName is the path parameter name of cluster_id
	ClusterVarName string
	// SubPathVarName is the path parameter name of sub-path needs to be forwarded
	SubPathVarName string

	wsTunnelStore      map[string]*WsTunnel
	wsTunnelMutateLock sync.RWMutex
}

type ClusterHandlerInstance struct {
	ServerAddress string
	Handler       http.Handler
}

// NewTunnelProxyDispatcher new a default TunnelProxyDispatcher
func NewTunnelProxyDispatcher(clusterVarName, subPathVarName string) *TunnelProxyDispatcher {
	return &TunnelProxyDispatcher{
		ClusterVarName: clusterVarName,
		SubPathVarName: subPathVarName,
		wsTunnelStore:  make(map[string]*WsTunnel),
	}
}

func (f *TunnelProxyDispatcher) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	start := time.Now()

	// first authenticate the request, only admin user be allowed
	auth := utils.Authenticate(req)
	if !auth {
		status := utils.NewUnauthorized("anonymous requests is forbidden")
		utils.WriteKubeAPIError(rw, status)
		return
	}

	vars := mux.Vars(req)
	// Get cluster id
	clusterId := vars[f.ClusterVarName]

	var proxyHandler *ClusterHandlerInstance
	// 先从websocket dialer缓存中查找websocket链
	websocketHandler, found, err := f.lookupWsHandler(clusterId)
	if err != nil {
		blog.Errorf("error when lookup websocket conn: %s", err.Error())
		err := fmt.Errorf("error when lookup websocket conn: %s", err.Error())
		status := utils.NewInternalError(err)
		status.ErrStatus.Reason = "CREATE_TUNNEL_ERROR"
		utils.WriteKubeAPIError(rw, status)
		return
	}
	// if found tunnel, use this tunnel to request to kube-apiserver
	if found {
		blog.Info("found websocket conn for k8s cluster %s", clusterId)
		handlerServer := stripLeaveSlash(f.ExtractPathPrefix(req), websocketHandler)
		proxyHandler = &ClusterHandlerInstance{
			Handler: handlerServer,
		}
		credentials := sqlstore.GetWsCredentials(clusterId)
		bearerToken := "Bearer " + credentials.UserToken
		req.Header.Set("Authorization", bearerToken)

		// set request scheme
		req.URL.Scheme = "https"

		// if webconsole long request, then set the latency before ServerHTTP
		if websocket.IsWebSocketUpgrade(req) {
			metrics.RequestCount.WithLabelValues("k8s_tunnel_request", "websocket").Inc()
			metrics.RequestLatency.WithLabelValues("k8s_tunnel_request", "websocket").Observe(time.Since(start).Seconds())
		}
		proxyHandler.Handler.ServeHTTP(rw, req)
		if !websocket.IsWebSocketUpgrade(req) {
			metrics.RequestCount.WithLabelValues("k8s_tunnel_request", req.Method).Inc()
			metrics.RequestLatency.WithLabelValues("k8s_tunnel_request", req.Method).Observe(time.Since(start).Seconds())
		}

		return
	}

	message := "no cluster can be found using given cluster id"
	status := utils.NewNotFound(utils.ClusterResource, clusterId, message)
	utils.WriteKubeAPIError(rw, status)
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
