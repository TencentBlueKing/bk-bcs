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

package types

import (
	"bufio"
	"crypto/tls"
	"github.com/gorilla/websocket"
	"net"
)

type ClientOptions struct {
	BcsApiAddress string
	BcsToken      string
	ClientSSL     *tls.Config
}

type WsConn struct {
	Conn *websocket.Conn
}

func NewWsConn(conn *websocket.Conn) *WsConn {
	return &WsConn{
		Conn: conn,
	}
}

func (c *WsConn) Read(p []byte) (n int, err error) {
	_, rc, err := c.Conn.NextReader()
	if err != nil {
		return 0, err
	}
	return rc.Read(p)
}

func (c *WsConn) Write(p []byte) (n int, err error) {
	wc, err := c.Conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return 0, err
	}
	defer wc.Close()
	return wc.Write(p)
}

// HijackedResponse holds connection information for a hijacked request.
type HijackedResponse struct {
	Ws     *WsConn
	Conn   net.Conn
	Reader *bufio.Reader
}

// Close closes the hijacked connection and reader.
func (h *HijackedResponse) Close() {
	//h.Conn.Close()
	h.Ws.Conn.Close()
}

// CloseWriter is an interface that implements structs
// that close input streams to prevent from writing.
type CloseWriter interface {
	CloseWrite() error
}

// CloseWrite closes a readWriter for writing.
func (h *HijackedResponse) CloseWrite() error {
	//if conn, ok := h.Conn.(CloseWriter); ok {
	//	return conn.CloseWrite()
	//}
	if conn, ok := h.Ws.Conn.UnderlyingConn().(CloseWriter); ok {
		return conn.CloseWrite()
	}
	return nil
}
