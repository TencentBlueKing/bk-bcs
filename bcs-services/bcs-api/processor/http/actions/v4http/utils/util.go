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

package utils

import (
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/pkg/storages/sqlstore"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-api/tunnel"
)

type WsTunnelDispatcher struct {
	wsTunnelStore      map[string]map[string]*WsTunnel
	wsTunnelMutateLock sync.RWMutex
}

type WsTunnel struct {
	httpTransport *http.Transport
	serverAddress string
}

func NewWsTunnelDispatcher() *WsTunnelDispatcher {
	return &WsTunnelDispatcher{
		wsTunnelStore: make(map[string]map[string]*WsTunnel),
	}
}

var DefaultWsTunnelDispatcher = NewWsTunnelDispatcher()

// LookupWsHandler will lookup websocket dialer in cache
func (w *WsTunnelDispatcher) LookupWsHandler(clusterId string) (string, *http.Transport, bool) {
	cluster := sqlstore.GetClusterByBCSInfo("", clusterId)
	if cluster == nil {
		return "", nil, false
	}
	credentials := sqlstore.GetWsCredentialsByClusterId(cluster.ID)
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
			w.wsTunnelMutateLock.Lock()
			wsTunnel := w.wsTunnelStore[clusterId][clientKey]
			if wsTunnel != nil && !w.serverAddressChanged(wsTunnel.serverAddress, serverAddress) {
				w.wsTunnelMutateLock.Unlock()
				return serverAddress, wsTunnel.httpTransport, true
			}
			tp := &http.Transport{
				MaxIdleConnsPerHost: 10,
			}
			cd := tunnelServer.Dialer(clientKey, 15*time.Second)
			tp.Dial = cd
			if wsTunnel != nil {
				wsTunnel.httpTransport.CloseIdleConnections()
			}
			if w.wsTunnelStore[clusterId] == nil {
				w.wsTunnelStore[clusterId] = make(map[string]*WsTunnel)
			}
			w.wsTunnelStore[clusterId][clientKey] = &WsTunnel{
				httpTransport: tp,
				serverAddress: serverAddress,
			}
			w.wsTunnelMutateLock.Unlock()
			return serverAddress, tp, true
		}
	}
	return "", nil, false
}

func (w *WsTunnelDispatcher) serverAddressChanged(oldAddress, newAddress string) bool {
	return oldAddress != newAddress
}
