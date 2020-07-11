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

package proxier

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"net/url"
	"strings"
	"time"

	m "github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/models"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/tunnel"
	"k8s.io/apimachinery/pkg/util/proxy"
	"k8s.io/client-go/transport"
)

// lookupWsHandler will lookup websocket dialer in cache
func lookupWsHandler(clusterId string, req *http.Request) (*proxy.UpgradeAwareHandler, bool, error) {
	credentials := sqlstore.GetWsCredentials(clusterId)
	if credentials == nil {
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

	transport := getTransport(clusterId, credentials)
	if transport == nil {
		return nil, false, nil
	}

	responder := &responder{}
	proxyHandler := proxy.NewUpgradeAwareHandler(u, transport, true, false, responder)
	proxyHandler.UseRequestLocation = true

	return proxyHandler, true, nil
}

func getTransport(clusterId string, credentials *m.WsClusterCredentials) http.RoundTripper {
	tp := &http.Transport{}
	if credentials.CaCertData != "" {
		certs := x509.NewCertPool()
		caCrt := []byte(credentials.CaCertData)
		certs.AppendCertsFromPEM(caCrt)
		tp.TLSClientConfig = &tls.Config{
			RootCAs: certs,
		}
	}

	tunnelServer := tunnel.DefaultTunnelServer
	if tunnelServer.HasSession(clusterId) {
		cd := tunnelServer.Dialer(clusterId, 15*time.Second)
		tp.Dial = cd
		bearerToken := credentials.UserToken
		bearerAuthRoundTripper := transport.NewBearerAuthRoundTripper(bearerToken, tp)

		return bearerAuthRoundTripper
	}

	return nil
}
