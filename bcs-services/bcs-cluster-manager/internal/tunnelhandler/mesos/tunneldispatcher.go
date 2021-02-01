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

package mesos

import (
	"context"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tunnel"
)

// WsTunnelDispatcher websocket tunnel dispatcher
type WsTunnelDispatcher struct {
	model              store.ClusterManagerModel
	tunnelServer       *websocketDialer.Server
	wsTunnelStore      map[string]map[string]*WsTunnel
	wsTunnelMutateLock sync.RWMutex
}

// WsTunnel websocket tunnel
type WsTunnel struct {
	httpTransport *http.Transport
	serverAddress string
}

// NewWsTunnelDispatcher create websocket tunnel dispatcher
func NewWsTunnelDispatcher(model store.ClusterManagerModel, tunnelServer *websocketDialer.Server) *WsTunnelDispatcher {
	return &WsTunnelDispatcher{
		model:         model,
		tunnelServer:  tunnelServer,
		wsTunnelStore: make(map[string]map[string]*WsTunnel),
	}
}

// LookupWsTransport will lookup websocket dialer in cache and generate transport
func (w *WsTunnelDispatcher) LookupWsTransport(clusterID string) (string, *http.Transport, bool) {
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		"clusterID":    clusterID,
		"clientModule": tunnel.MesosDriverModule,
	})
	credentials, err := w.model.ListClusterCredential(context.TODO(), cond, &storeopt.ListOption{})
	if err != nil {
		blog.Warnf("get clueter %s credential from store failed, err %s", clusterID, err.Error())
		return "", nil, false
	}
	if len(credentials) == 0 {
		blog.Warnf("cluster %s credential not found in store", clusterID)
		return "", nil, false
	}

	rand.Shuffle(len(credentials), func(i, j int) {
		credentials[i], credentials[j] = credentials[j], credentials[i]
	})

	tunnelServer := w.tunnelServer
	for _, credential := range credentials {
		clientKey := credential.ServerKey
		serverAddress := credential.ServerAddress
		if tunnelServer.HasSession(clientKey) {
			blog.Infof("found sesseion: %s", clientKey)
			w.wsTunnelMutateLock.Lock()
			wsTunnel := w.wsTunnelStore[clusterID][clientKey]
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
			if w.wsTunnelStore[clusterID] == nil {
				w.wsTunnelStore[clusterID] = make(map[string]*WsTunnel)
			}
			w.wsTunnelStore[clusterID][clientKey] = &WsTunnel{
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
