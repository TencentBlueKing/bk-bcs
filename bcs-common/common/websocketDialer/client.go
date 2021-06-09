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

package websocketDialer

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gorilla/websocket"
)

// ConnectAuthorizer is function form for authorize connect
type ConnectAuthorizer func(proto, address string) bool

// ClientConnect do client connection
func ClientConnect(
	ctx context.Context, wsURL string, headers http.Header,
	tlsConfig *tls.Config, dialer *websocket.Dialer, auth ConnectAuthorizer) error {
	if err := connectToProxy(ctx, wsURL, headers, tlsConfig, auth, dialer); err != nil {
		time.Sleep(time.Duration(5) * time.Second)
		return err
	}

	return nil
}

func connectToProxy(
	rootCtx context.Context, proxyURL string, headers http.Header,
	tlsConfig *tls.Config, auth ConnectAuthorizer, dialer *websocket.Dialer) error {
	blog.Infof("connecting to proxy %s", proxyURL)

	if dialer == nil {
		dialer = &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: HandshakeTimeOut,
			TLSClientConfig:  tlsConfig}
	}
	ws, resp, err := dialer.Dial(proxyURL, headers)
	if err != nil {
		if resp == nil {
			blog.Errorf("Failed to connect to proxy, empty dialer response")
		} else {
			rb, err2 := ioutil.ReadAll(resp.Body)
			if err2 != nil {
				blog.Errorf(
					"Failed to connect to proxy. Response status: %v - %v. Couldn't read response body (err: %v)",
					resp.StatusCode, resp.Status, err2)
			} else {
				blog.Errorf(
					"Failed to connect to proxy. Response status: %v - %v. Response body: %s. Error: %s",
					resp.StatusCode, resp.Status, rb, err.Error())
			}
		}
		blog.Errorf("Failed to connect, err %s", err.Error())
		return err
	}
	defer ws.Close()

	result := make(chan error, 2)

	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	session := NewClientSession(auth, ws)
	defer session.Close()

	go func() {
		_, err = session.Serve(ctx)
		result <- err
	}()

	select {
	case <-ctx.Done():
		blog.Infof("proxy %s done", proxyURL)
		return nil
	case err := <-result:
		return err
	}
}
