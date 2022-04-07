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

package sdk

import (
	"context"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/tunnel"
)

// WebsocketClient describe a simple client as an agent.
// It should connect to the argocd-proxy and keep the session.
type WebsocketClient struct {
	// proxy address for connecting
	proxyAddress string

	// target serverAddress
	serverAddress string

	// the cluster under this client's management
	clusterID string

	ctx    context.Context
	cancel context.CancelFunc

	lastConnectTime time.Time
	reconnectTimes  int
}

// NewWebsocketClient return a new WebsocketClient instance
func NewWebsocketClient(proxyAddress, serverAddress, clusterID string) *WebsocketClient {
	return &WebsocketClient{
		proxyAddress:  proxyAddress,
		serverAddress: serverAddress,
		clusterID:     clusterID,
	}
}

// Start the connection and keep the session
func (wc *WebsocketClient) Start() {
	wc.ctx, wc.cancel = context.WithCancel(context.Background())
	go wc.connect2Proxy(wc.ctx)
}

func (wc *WebsocketClient) connect2Proxy(ctx context.Context) {
	headers := http.Header{}
	headers.Set(tunnel.BcsArgocdManagedClusterID, wc.clusterID)
	headers.Set(tunnel.BcsArgocdManagedServerAddress, wc.serverAddress)

	proxyWS := wc.proxyAddress + "/websocket/connect"

	for {
		select {
		case <-ctx.Done():
			return
		default:
			wc.lastConnectTime = time.Now()
			blog.Infof("try connect to tunnel proxy address %s", proxyWS)
			if err := websocketDialer.ClientConnect(ctx, proxyWS, headers, nil, nil,
				func(proto, address string) bool {
					switch proto {
					case "tcp":
						return true
					case "unix":
						return address == "/var/run/docker.sock"
					}
					return false
				}); err != nil {
				blog.Errorf("client websocket connect failed, %s, %v", proxyWS, err)
			}
			time.Sleep(wc.reconnectTimeout())
		}
	}
}

func (wc *WebsocketClient) reconnectTimeout() time.Duration {
	if time.Now().Sub(wc.lastConnectTime) > time.Second*10 {
		wc.reconnectTimes = 0
	}

	if wc.reconnectTimes < 5 {
		return time.Duration(0)
	}

	wc.reconnectTimes++
	return time.Second * 5
}

// Stop the connection
func (wc *WebsocketClient) Stop() {
	if wc.cancel != nil {
		wc.cancel()
	}
}
