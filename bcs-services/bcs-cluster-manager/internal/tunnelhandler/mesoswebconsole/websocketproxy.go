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

package mesoswebconsole

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer"
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

	BackendURL *url.URL

	// Upgrader specifies the parameters for upgrading a incoming HTTP
	// connection to a WebSocket connection. If nil, DefaultUpgrader is used.
	Upgrader *websocket.Upgrader

	//  Dialer contains options for connecting to the backend WebSocket server.
	//  If nil, DefaultDialer is used.
	Dialer *websocket.Dialer

	clientTLSConfig *tls.Config
}

// NewWebsocketProxy returns a new Websocket reverse proxy that rewrites the
// URL's to the scheme, host and base path provider in target.
func NewWebsocketProxy(clientTLSConfig *tls.Config, backendURL *url.URL, clusterDialer websocketDialer.Dialer) *WebsocketProxy {

	// DefaultDialer is a dialer with all fields set to the default zero values.
	defaultDialer := websocket.DefaultDialer
	if backendURL.Scheme == "https" {
		defaultDialer.TLSClientConfig = clientTLSConfig
		backendURL.Scheme = "wss"
	} else {
		backendURL.Scheme = "ws"
	}
	defaultDialer.NetDial = clusterDialer

	return &WebsocketProxy{
		BackendURL: backendURL,
		Dialer:     defaultDialer,
		Upgrader:   DefaultUpgrader,
	}
}

// ServeHTTP implements the http.Handler that proxies WebSocket connections.
func (w *WebsocketProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	backendURL := w.BackendURL
	requestHeader := w.setRequestHeader(req)
	if w.Director != nil {
		w.Director(req, requestHeader)
	}
	connBackend, resp, err := w.Dialer.Dial(backendURL.String(), requestHeader)
	if err != nil {
		blog.Errorf("websocketproxy: couldn't dial to remote backend url %s", err)
		message := fmt.Sprintf("errcode: %d, couldn't dial to remote backend url %s",
			common.BcsErrApiWebConsoleFailedCode, err.Error())
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

	// begin to replicate websocket conn data
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

// setRequestHeader pass the origin header to backend request
func (w *WebsocketProxy) setRequestHeader(req *http.Request) http.Header {
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
	// should always add the BCS-ClusterID to request to mesos-driver
	for _, cluster := range req.Header[http.CanonicalHeaderKey("BCS-ClusterID")] {
		requestHeader.Add("BCS-ClusterID", cluster)
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

	return requestHeader
}

// replicateWebsocketConn keep replicating data from source websocket conn to destination websocket conn
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

// copyHeader copy all header from source to destination
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// copyResponse copy body from source ResponseWriter to destination Response
func copyResponse(rw http.ResponseWriter, resp *http.Response) error {
	copyHeader(rw.Header(), resp.Header)
	rw.WriteHeader(resp.StatusCode)
	defer resp.Body.Close()

	_, err := io.Copy(rw, resp.Body)
	return err
}
