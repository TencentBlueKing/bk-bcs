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
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	htplib "net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-client/pkg/types"
	"github.com/pkg/errors"

	"crypto/tls"

	"github.com/Tencent/bk-bcs/bcs-common/common/http"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	"github.com/gorilla/websocket"
)

// ApiRequester http interface for bcs-client
type ApiRequester interface {
	Do(uri, method string, data []byte, header ...*http.HeaderSet) ([]byte, error)
	DoForResponse(uri, method string, data []byte, header ...*http.HeaderSet) (*httpclient.HttpRespone, error)
	DoWebsocket(uri string, header ...*http.HeaderSet) (types.HijackedResponse, error)
	PostHijacked(ctx context.Context, uri string, header ...*http.HeaderSet) (types.HijackedResponse, error)
}

//NewApiRequester api request
func NewApiRequester(clientSSL *tls.Config, bcsToken string) ApiRequester {
	return &bcsApiRequester{
		clientSSL: clientSSL,
		bcsToken:  bcsToken,
	}
}

// BcsApiRequester is the way to request to all bcs-api uri
type bcsApiRequester struct {
	clientSSL *tls.Config
	bcsToken  string
}

// Do do http request
func (b *bcsApiRequester) Do(uri, method string, data []byte, header ...*http.HeaderSet) ([]byte, error) {
	httpCli := httpclient.NewHttpClient()
	httpCli.SetHeader("Content-Type", "application/json")
	httpCli.SetHeader("Accept", "application/json")
	if b.bcsToken != "" {
		httpCli.SetHeader("Authorization", "Bearer "+b.bcsToken)
	}
	//httpCli.SetHeader("X-Bcs-User-Token", b.bcsToken)

	if header != nil {
		httpCli.SetBatchHeader(header)
	}

	if b.clientSSL != nil {
		httpCli.SetTlsVerityConfig(b.clientSSL)
	}
	//changed by DeveloperJim in 2020-04-27 for handling http error code
	response, err := httpCli.RequestEx(uri, method, nil, data)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != htplib.StatusOK {
		return nil, fmt.Errorf("%s", response.Status)
	}
	return response.Reply, nil
}

// DoForResponse get response after request
func (b *bcsApiRequester) DoForResponse(uri, method string, data []byte, header ...*http.HeaderSet) (*httpclient.HttpRespone, error) {
	httpCli := httpclient.NewHttpClient()
	httpCli.SetHeader("Content-Type", "application/json")
	httpCli.SetHeader("Accept", "application/json")
	if b.bcsToken != "" {
		httpCli.SetHeader("Authorization", "Bearer "+b.bcsToken)
	}

	if header != nil {
		httpCli.SetBatchHeader(header)
	}

	if b.clientSSL != nil {
		httpCli.SetTlsVerityConfig(b.clientSSL)
	}

	return httpCli.RequestEx(uri, method, nil, data)
}

// DoWebsocket websocket request
func (b *bcsApiRequester) DoWebsocket(uri string, header ...*http.HeaderSet) (types.HijackedResponse, error) {
	var hijackedResp types.HijackedResponse

	u, err := url.Parse(uri)
	if err != nil {
		return hijackedResp, err
	}
	if u.Scheme == "http" {
		u.Scheme = "ws"
	} else if u.Scheme == "https" {
		u.Scheme = "wss"
	}

	wsHeader := htplib.Header{}
	if b.bcsToken != "" {
		wsHeader.Set("Authorization", "Bearer "+b.bcsToken)
	}
	if header != nil {
		for _, h := range header {
			wsHeader.Set(h.Key, h.Value)
		}
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), wsHeader)

	if err != nil {
		return hijackedResp, fmt.Errorf("unable to dial to backend websocket server: %s", err.Error())
	}

	ws := types.NewWsConn(conn)

	//return types.HijackedResponse{Conn: conn.UnderlyingConn(), Reader: bufio.NewReader(conn.UnderlyingConn())}, err
	return types.HijackedResponse{Ws: ws}, err
}

// PostHijacked post hijack for websocket
func (b *bcsApiRequester) PostHijacked(ctx context.Context, uri string, header ...*http.HeaderSet) (types.HijackedResponse, error) {
	req, err := htplib.NewRequest(htplib.MethodGet, uri, nil)
	if err != nil {
		return types.HijackedResponse{}, err
	}
	if header != nil {
		for _, h := range header {
			req.Header.Set(h.Key, h.Value)
		}
	}
	if b.bcsToken != "" {
		req.Header.Set("Authorization", "Bearer "+b.bcsToken)
	}
	req.Header.Set("Content-Type", "application/json")

	conn, err := b.setupHijackConn(ctx, uri, req, "websocket")
	if err != nil {
		return types.HijackedResponse{}, err
	}

	return types.HijackedResponse{Conn: conn, Reader: bufio.NewReader(conn)}, err
}

func (b *bcsApiRequester) setupHijackConn(ctx context.Context, uri string, req *htplib.Request, proto string) (net.Conn, error) {
	challengeKey, err := generateChallengeKey()
	if err != nil {
		return nil, err
	}

	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", proto)
	req.Header["Sec-WebSocket-Key"] = []string{challengeKey}
	req.Header["Sec-WebSocket-Version"] = []string{"13"}

	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	dialer := b.Dialer(u.Host)
	conn, err := dialer(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "cannot connect to the Docker daemon. Is 'docker daemon' running on this host?")
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	clientconn := httputil.NewClientConn(conn, nil)
	defer clientconn.Close()

	// Server hijacks the connection, error 'connection closed' expected
	resp, err := clientconn.Do(req)

	//nolint:staticcheck // ignore SA1019 for connecting to old (pre go1.8) daemons
	if err != httputil.ErrPersistEOF {
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != htplib.StatusSwitchingProtocols {
			resp.Body.Close()
			return nil, fmt.Errorf("unable to upgrade to %s, received %d", proto, resp.StatusCode)
		}
	}

	c, br := clientconn.Hijack()
	if br.Buffered() > 0 {
		// If there is buffered content, wrap the connection.  We return an
		// object that implements CloseWrite iff the underlying connection
		// implements it.
		if _, ok := c.(types.CloseWriter); ok {
			c = &hijackedConnCloseWriter{&hijackedConn{c, br}}
		} else {
			c = &hijackedConn{c, br}
		}
	} else {
		br.Reset(nil)
	}

	return c, nil
}

func (b *bcsApiRequester) Dialer(addr string) func(context.Context) (net.Conn, error) {
	return func(ctx context.Context) (net.Conn, error) {
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 0 * time.Second,
		}
		return dialer.DialContext(ctx, "tcp", addr)
	}
}

type hijackedConnCloseWriter struct {
	*hijackedConn
}

type hijackedConn struct {
	net.Conn
	r *bufio.Reader
}

func (c *hijackedConn) Read(b []byte) (int, error) {
	return c.r.Read(b)
}

func (c *hijackedConnCloseWriter) CloseWrite() error {
	conn := c.Conn.(types.CloseWriter)
	return conn.CloseWrite()
}

func generateChallengeKey() (string, error) {
	p := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, p); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(p), nil
}
