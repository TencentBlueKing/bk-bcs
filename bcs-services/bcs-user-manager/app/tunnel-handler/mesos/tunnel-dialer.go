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

package mesos

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	"math/rand"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/tunnel"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/storages/sqlstore"
)

// LookupWsTransport will lookup websocket dialer in cache and generate transport
func LookupWsTransport(clusterId string) (string, *http.Transport, bool) {
	credentials := sqlstore.GetWsCredentialsByClusterId(clusterId)
	if len(credentials) == 0 {
		return "", nil, false
	}

	rand.Shuffle(len(credentials), func(i, j int) {
		credentials[i], credentials[j] = credentials[j], credentials[i]
	})

	tunnelServer := tunnel.DefaultTunnelServer
	for _, credential := range credentials {
		clientKey := credential.ServerKey
		serverAddress := credential.ServerAddress
		if tunnelServer.HasSession(clientKey) {
			blog.Infof("found sesseion for mesos: %s", clientKey)
			tp := &http.Transport{}
			cd := tunnelServer.Dialer(clientKey, 15*time.Second)
			tp.Dial = cd
			return serverAddress, tp, true
		}
	}
	return "", nil, false
}

// LookupWsDialer will lookup websocket dialer in cache
func LookupWsDialer(clusterId string) (string, websocketDialer.Dialer, bool) {
	credentials := sqlstore.GetWsCredentialsByClusterId(clusterId)
	if len(credentials) == 0 {
		return "", nil, false
	}

	rand.Shuffle(len(credentials), func(i, j int) {
		credentials[i], credentials[j] = credentials[j], credentials[i]
	})

	tunnelServer := tunnel.DefaultTunnelServer
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
