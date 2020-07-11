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

package webconsole

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-driver/mesosdriver/config"

	"github.com/gorilla/websocket"
)

var (
	// DefaultUpgrader specifies the parameters for upgrading an HTTP
	// connection to a WebSocket connection.
	DefaultUpgrader = &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

// WebsocketProxy is an HTTP Handler that takes an incoming WebSocket
// connection and proxies it to another server.
type WebsocketProxy struct {
	// Director, if non-nil, is a function that may copy additional request
	// headers from the incoming WebSocket connection into the output headers
	// which will be forwarded to another server.
	Director func(incoming *http.Request, out http.Header)

	BackendUrl *url.URL

	// Upgrader specifies the parameters for upgrading a incoming HTTP
	// connection to a WebSocket connection. If nil, DefaultUpgrader is used.
	Upgrader *websocket.Upgrader

	//  Dialer contains options for connecting to the backend WebSocket server.
	//  If nil, DefaultDialer is used.
	Dialer *websocket.Dialer
}

// NewWebsocketProxy returns a new Websocket reverse proxy that rewrites the
// URL's to the scheme, host and base path provider in target.
func NewWebsocketProxy(certConfig *config.CertConfig, backendUrl *url.URL) *WebsocketProxy {
	if certConfig.IsSSL {
		backendUrl.Scheme = "wss"
	} else {
		backendUrl.Scheme = "ws"
	}

	// DefaultDialer is a dialer with all fields set to the default zero values.
	defaultDialer := websocket.DefaultDialer
	if certConfig.IsSSL {
		cliTls, err := ssl.ClientTslConfVerity(certConfig.CAFile, certConfig.CertFile, certConfig.KeyFile, certConfig.CertPasswd)
		if err != nil {
			blog.Errorf("set client tls config error %s", err.Error())
			return nil
		}

		defaultDialer.TLSClientConfig = cliTls
		blog.Infof("set client tls config success")
	}

	return &WebsocketProxy{
		BackendUrl: backendUrl,
		Dialer:     defaultDialer,
		Upgrader:   DefaultUpgrader,
	}
}

// ServeHTTP implements the http.Handler that proxies WebSocket connections.
func (w *WebsocketProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	backendURL := w.BackendUrl
	// Pass headers from the incoming request to the dialer to forward them to
	// the final destinations.
	requestHeader := http.Header{}
	if origin := req.Header.Get("Origin"); origin != "" {
		requestHeader.Add("Origin", origin)
	}
	for _, prot := range req.Header[http.CanonicalHeaderKey("Sec-WebSocket-Protocol")] {
		requestHeader.Add("Sec-WebSocket-Protocol", prot)
	}
	for _, cookie := range req.Header[http.CanonicalHeaderKey("Cookie")] {
		requestHeader.Add("Cookie", cookie)
	}
	if req.Host != "" {
		requestHeader.Set("Host", req.Host)
	}

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if prior, ok := req.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		requestHeader.Set("X-Forwarded-For", clientIP)
	}

	requestHeader.Set("X-Forwarded-Proto", "http")
	if req.TLS != nil {
		requestHeader.Set("X-Forwarded-Proto", "https")
	}

	if w.Director != nil {
		w.Director(req, requestHeader)
	}

	connBackend, resp, err := w.Dialer.Dial(backendURL.String(), requestHeader)
	if err != nil {
		blog.Errorf("websocketproxy: couldn't dial to remote backend url %s", err)
		message := fmt.Sprintf("errcode: %d, couldn't dial to remote backend url %s", common.BcsErrApiWebConsoleFailedCode, err.Error())
		if resp != nil {
			if err := copyResponse(rw, resp); err != nil {
				blog.Errorf("websocketproxy: couldn't write response after failed remote backend handshake: %s", err)
			}
		} else {
			http.Error(rw, message, http.StatusServiceUnavailable)
		}
		return
	}
	defer connBackend.Close()

	// Only pass those headers to the upgrader.
	upgradeHeader := http.Header{}
	if hdr := resp.Header.Get("Sec-Websocket-Protocol"); hdr != "" {
		upgradeHeader.Set("Sec-Websocket-Protocol", hdr)
	}
	if hdr := resp.Header.Get("Set-Cookie"); hdr != "" {
		upgradeHeader.Set("Set-Cookie", hdr)
	}

	w.Upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	// Now upgrade the existing incoming request to a WebSocket connection.
	// Also pass the header that we gathered from the Dial handshake.
	connPub, err := w.Upgrader.Upgrade(rw, req, upgradeHeader)
	if err != nil {
		blog.Errorf("websocketproxy: couldn't upgrade %s", err)
		return
	}
	defer connPub.Close()

	errClient := make(chan error, 1)
	errBackend := make(chan error, 1)

	go replicateWebsocketConn(connPub, connBackend, errClient)
	go replicateWebsocketConn(connBackend, connPub, errBackend)

	var message string
	select {
	case err = <-errClient:
		message = "websocketproxy: Error when copying from backend to client: %v"
	case err = <-errBackend:
		message = "websocketproxy: Error when copying from client to backend: %v"

	}
	if e, ok := err.(*websocket.CloseError); !ok || e.Code == websocket.CloseAbnormalClosure {
		blog.Errorf(message, err)
	}
}

func replicateWebsocketConn(dst, src *websocket.Conn, errc chan error) {
	for {
		msgType, msg, err1 := src.ReadMessage()
		if err1 != nil {
			m := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("%v", err1))
			if e, ok := err1.(*websocket.CloseError); ok {
				if e.Code != websocket.CloseNoStatusReceived {
					m = websocket.FormatCloseMessage(e.Code, e.Text)
				}
			}
			errc <- err1
			dst.WriteMessage(websocket.CloseMessage, m)
			break
		}
		err := dst.WriteMessage(msgType, msg)
		if err != nil {
			errc <- err
			break
		}
	}
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func copyResponse(rw http.ResponseWriter, resp *http.Response) error {
	copyHeader(rw.Header(), resp.Header)
	rw.WriteHeader(resp.StatusCode)
	defer resp.Body.Close()

	_, err := io.Copy(rw, resp.Body)
	return err
}
