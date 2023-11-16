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

package proxier

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/transport"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/tunnel"
)

// WsTunnel xxx
type WsTunnel struct {
	httpTransport *http.Transport
	serverAddress string
	userToken     string
	caCertData    string
}

// lookupWsHandler will lookup websocket dialer in cache
func (f *ReverseProxyDispatcher) lookupWsHandler(clusterId string, req *http.Request) ( // nolint
	*proxy.UpgradeAwareHandler, bool, error) {
	credentials := sqlstore.GetWsCredentials(clusterId)
	if credentials == nil {
		return nil, false, nil
	}

	serverAddress := credentials.ServerAddress
	if !strings.HasSuffix(serverAddress, "/") {
		serverAddress += "/"
	}
	u, err := url.Parse(serverAddress)
	if err != nil {
		return nil, false, err
	}

	transport := f.getTransport(clusterId, credentials)
	if transport == nil {
		return nil, false, nil
	}

	responder := &responder{}
	blog.Infof("lookupWsHandler, clusterId: %s, serverAddress: %s", clusterId, serverAddress)
	proxyHandler := proxy.NewUpgradeAwareHandler(u, transport, true, false, responder)
	proxyHandler.UseRequestLocation = true

	return proxyHandler, true, nil
}

func (f *ReverseProxyDispatcher) wsTunnelChanged(clusterId string, credentials *m.WsClusterCredentials) bool {
	wsTunnel := f.wsTunnelStore[clusterId]
	return wsTunnel.serverAddress != credentials.ServerAddress || wsTunnel.caCertData != credentials.CaCertData ||
		wsTunnel.userToken != credentials.UserToken
}

// getTransport generate transport with dialer from tunnel
func (f *ReverseProxyDispatcher) getTransport(clusterId string, credentials *m.WsClusterCredentials) http.RoundTripper {
	tunnelServer := tunnel.DefaultTunnelServer
	if tunnelServer.HasSession(clusterId) {
		f.wsTunnelMutateLock.Lock()
		defer f.wsTunnelMutateLock.Unlock()

		if f.wsTunnelStore[clusterId] != nil && !f.wsTunnelChanged(clusterId, credentials) {
			return f.wsTunnelStore[clusterId].httpTransport
		}

		tp := &http.Transport{
			MaxIdleConnsPerHost: 10,
		}
		if credentials.CaCertData != "" {
			certs := x509.NewCertPool()
			caCrt := []byte(credentials.CaCertData)
			certs.AppendCertsFromPEM(caCrt)
			tp.TLSClientConfig = &tls.Config{ // nolint
				RootCAs: certs,
			}
		}
		cd := tunnelServer.Dialer(clusterId, 15*time.Second)
		tp.Dial = cd // nolint

		if f.wsTunnelStore[clusterId] != nil {
			f.wsTunnelStore[clusterId].httpTransport.CloseIdleConnections()
		}
		f.wsTunnelStore[clusterId] = &WsTunnel{
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
