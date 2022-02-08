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
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"

	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/transport"
)

// responder implements k8s.io/apimachinery/pkg/util/proxy.ErrorResponder
type responder struct{}

// Error implements ErrorResponder Error function
func (r *responder) Error(w http.ResponseWriter, req *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

// lookupWsHandler will lookup websocket dialer in cache
func (f *TunnelProxyDispatcher) lookupWsHandler(clusterID string) (*proxy.UpgradeAwareHandler, bool, error) {
	credentials, found, err := f.model.GetClusterCredential(context.TODO(), clusterID)
	if err != nil {
		return nil, false, err
	}
	if !found {
		return nil, false, nil
	}

	serverAddress := credentials.ServerAddress
	if !strings.HasSuffix(serverAddress, "/") {
		serverAddress = serverAddress + "/"
	}
	u, err := url.Parse(serverAddress)
	if err != nil {
		return nil, false, err
	}

	transport := f.getTransport(clusterID, credentials)
	if transport == nil {
		return nil, false, nil
	}

	responder := &responder{}
	proxyHandler := proxy.NewUpgradeAwareHandler(u, transport, true, false, responder)
	proxyHandler.UseRequestLocation = true

	return proxyHandler, true, nil
}

func (f *TunnelProxyDispatcher) wsTunnelChanged(clusterID string, credentials *types.ClusterCredential) bool {
	wsTunnel := f.wsTunnelStore[clusterID]
	return wsTunnel.serverAddress != credentials.ServerAddress ||
		wsTunnel.caCertData != credentials.CaCertData ||
		wsTunnel.userToken != credentials.UserToken
}

// getTransport generate transport with dialer from tunnel
func (f *TunnelProxyDispatcher) getTransport(clusterID string, credentials *types.ClusterCredential) http.RoundTripper {
	tunnelServer := f.tunnelServer
	if tunnelServer.HasSession(clusterID) {
		f.wsTunnelMutateLock.Lock()
		defer f.wsTunnelMutateLock.Unlock()

		if f.wsTunnelStore[clusterID] != nil && !f.wsTunnelChanged(clusterID, credentials) {
			return f.wsTunnelStore[clusterID].httpTransport
		}

		tp := &http.Transport{
			MaxIdleConnsPerHost: 10,
		}
		if credentials.CaCertData != "" {
			certs := x509.NewCertPool()
			caCrt := []byte(credentials.CaCertData)
			certs.AppendCertsFromPEM(caCrt)
			tp.TLSClientConfig = &tls.Config{
				RootCAs: certs,
			}
		}

		blog.Infof("found sesseion for k8s: %s", clusterID)
		// get dialer from tunnel sessions
		cd := tunnelServer.Dialer(clusterID, 15*time.Second)
		tp.Dial = cd

		if f.wsTunnelStore[clusterID] != nil {
			f.wsTunnelStore[clusterID].httpTransport.CloseIdleConnections()
		}
		f.wsTunnelStore[clusterID] = &WsTunnel{
			httpTransport: tp,
			serverAddress: credentials.ServerAddress,
			userToken:     credentials.UserToken,
			caCertData:    credentials.CaCertData,
		}

		bearerToken := credentials.UserToken
		bearerAuthRoundTripper := transport.NewBearerAuthRoundTripper(bearerToken, tp)

		return bearerAuthRoundTripper
	}

	return nil
}
