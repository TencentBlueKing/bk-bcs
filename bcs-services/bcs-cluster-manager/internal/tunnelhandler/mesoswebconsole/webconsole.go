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

package mesoswebconsole

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bhttp "github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/gorilla/websocket"
)

// WebconsoleProxy proxy for web console
type WebconsoleProxy struct {
	dialerServer    *websocketDialer.Server
	clientTLSConfig *tls.Config
	model           store.ClusterManagerModel
	// Backend returns the backend URL which the proxy uses to reverse proxy
	Backend func(*http.Request) (*url.URL, websocketDialer.Dialer, error)
}

// NewWebconsoleProxy create a webconsole proxy
func NewWebconsoleProxy(
	clientTLSConfig *tls.Config, model store.ClusterManagerModel,
	dialerServer *websocketDialer.Server) *WebconsoleProxy {
	proxy := &WebconsoleProxy{
		dialerServer:    dialerServer,
		clientTLSConfig: clientTLSConfig,
		model:           model,
	}
	proxy.Backend = func(req *http.Request) (*url.URL, websocketDialer.Dialer, error) {
		cluster := req.Header.Get("BCS-ClusterID")
		if cluster == "" {
			blog.Error("handler url read header BCS-ClusterID is empty")
			err1 := bhttp.InternalError(common.BcsErrCommHttpParametersFailed,
				"http header BCS-ClusterID can't be empty")
			return nil, nil, err1
		}

		// find whether exist a cluster tunnel dialer in sessions
		serverAddr, clusterDialer, found := proxy.lookupWsDialer(cluster)
		if found {
			tunnelURL, err := url.Parse(serverAddr)
			if err != nil {
				return nil, nil, fmt.Errorf("error when parse server address: %s", err.Error())
			}
			originURL := req.URL
			originURL.Host = tunnelURL.Host
			originURL.Scheme = tunnelURL.Scheme
			return originURL, clusterDialer, nil
		}
		return nil, nil, fmt.Errorf("no tunnel could be found for cluster %s", cluster)
	}
	return proxy
}

// ServeHTTP handle webconsole request
func (w *WebconsoleProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	start := time.Now()

	backendURL, clusterDialer, err := w.Backend(req)
	if err != nil {
		http.Error(rw, err.Error(), http.StatusBadRequest)
		return
	}
	// if websocket request, handle it with websocket proxy
	if websocket.IsWebSocketUpgrade(req) {
		websocketProxy := NewWebsocketProxy(w.clientTLSConfig, backendURL, clusterDialer)
		websocketProxy.ServeHTTP(rw, req)
		return
	}

	// if ordinary request, handle it with http proxy
	httpProxy := NewHTTPReverseProxy(w.clientTLSConfig, backendURL, clusterDialer)
	httpProxy.ServeHTTP(rw, req)
	metrics.ReportAPIRequestMetric("mesos_webconsole", req.Method, metrics.LibCallStatusOK, start)
	return
}

// lookup websocket dialer in cache
func (w *WebconsoleProxy) lookupWsDialer(clusterID string) (string, websocketDialer.Dialer, bool) {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"clusterID": clusterID,
	})
	credentials, err := w.model.ListClusterCredential(context.TODO(), cond, &storeopt.ListOption{})
	if err != nil {
		blog.Warnf("get clueter %s credential from store failed, err %s", clusterID, err.Error())
		return "", nil, false
	}
	if len(credentials) == 0 {
		return "", nil, false
	}

	rand.Shuffle(len(credentials), func(i, j int) {
		credentials[i], credentials[j] = credentials[j], credentials[i]
	})

	tunnelServer := w.dialerServer
	for _, credential := range credentials {
		clientKey := credential.ServerKey
		serverAddress := credential.ServerAddress
		if tunnelServer.HasSession(clientKey) {
			blog.Infof("found sesseion: %s", clientKey)
			clusterDialer := tunnelServer.Dialer(clientKey, 15*time.Second)
			return serverAddress, clusterDialer, true
		}
	}
	return "", nil, false
}
