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

// Package tunnel xxx
package tunnel

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/metric"
)

// ClientOptions for websocket client
type ClientOptions struct {
	// Context for exit control
	Context context.Context
	// remote gitops proxy config
	ProxyAddress string
	ProxyToken   string
	TLSConfig    *tls.Config
	// local http backend
	LocalEndpoint string
	// local ClusterID information for tunnel identity
	ClusterID    string
	ClusterToken string
	ConnectURL   string // connect url likes /bcsapi/v4/gitopsproxy/websocket/connect
}

// Validate option
func (opt *ClientOptions) Validate() error {
	if opt.Context == nil {
		return fmt.Errorf("lost option context")
	}
	if len(opt.ClusterID) == 0 {
		return fmt.Errorf("lost ClusterID for management")
	}
	if len(opt.ProxyAddress) == 0 {
		return fmt.Errorf("lost remote proxy address")
	}
	if len(opt.LocalEndpoint) == 0 {
		return fmt.Errorf("lost location information")
	}
	if opt.TLSConfig == nil {
		return fmt.Errorf("lost tls configuration")
	}
	return nil
}

// Client describe a simple client as an agent.
// It should connect to the gitops-proxy and keep the session.
type Client struct {
	option        *ClientOptions
	lastConnected time.Time
	reconnectCnt  int
}

// NewClient return a new Client instance
func NewClient(opt *ClientOptions) *Client {
	return &Client{
		option: opt,
	}
}

// Init essential elements
func (c *Client) Init() error {
	if c.option == nil {
		return fmt.Errorf("lost client options")
	}
	if err := c.option.Validate(); err != nil {
		return err
	}
	return nil
}

// Start the connection and keep the session, non-blocking
func (c *Client) Start() {
	go c.connect2Proxy()
}

func (c *Client) connect2Proxy() {
	headers := http.Header{}
	headers.Set(websocketDialer.ID, c.option.ClusterID)
	headers.Set(websocketDialer.Token, c.option.ClusterToken)
	headers.Set(common.HeaderServerAddressKey, c.option.LocalEndpoint)
	if len(c.option.ProxyToken) != 0 {
		headers.Set("Authorization", fmt.Sprintf("Bearer %s", c.option.ProxyToken))
	}

	proxyWS := fmt.Sprintf("%s%s", c.option.ProxyAddress, c.option.ConnectURL)
	blog.Infof("tunnel client ready to connect proxy %s", proxyWS)
	for {
		select {
		case <-c.option.Context.Done():
			blog.Infof("tunnel client is ready to close")
			return
		default:
			c.lastConnected = time.Now()
			blog.Infof("try connect to tunnel proxy address %s in loop", proxyWS)
			metric.ManagerTunnelConnectStatus.WithLabelValues().Set(0)
			if err := websocketDialer.ClientConnect(
				c.option.Context,
				proxyWS,
				headers,
				c.option.TLSConfig,
				nil,
				func(proto, address string) bool {
					return proto == "tcp"
				}); err != nil {
				blog.Errorf("client websocket connect %s failed, %s", proxyWS, err.Error())
			}
			metric.ManagerTunnelConnectStatus.WithLabelValues().Set(1)
			metric.ManagerTunnelConnectNum.WithLabelValues().Inc()
			time.Sleep(c.backoff())
		}
	}
}

// backoff strategy for reconnection
func (c *Client) backoff() time.Duration {
	if time.Since(c.lastConnected) > time.Second*10 {
		c.reconnectCnt = 0
	}
	if c.reconnectCnt < 5 {
		return time.Duration(0)
	}
	c.reconnectCnt++
	return time.Second * 5
}
