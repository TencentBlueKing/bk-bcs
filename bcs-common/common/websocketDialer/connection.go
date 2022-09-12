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
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer/metrics"
)

type connection struct {
	sync.Mutex

	ctx           context.Context
	cancel        func()
	err           error
	writeDeadline time.Time
	buf           chan []byte
	readBuf       []byte
	addr          addr
	session       *Session
	connID        int64
}

func newConnection(connID int64, session *Session, proto, address string) *connection {
	c := &connection{
		addr: addr{
			proto:   proto,
			address: address,
		},
		connID:  connID,
		session: session,
		buf:     make(chan []byte, 1024),
	}
	metrics.IncSMTotalAddConnectionsForWS(session.clientKey, proto, address)
	return c
}

func (c *connection) tunnelClose(err error) {
	metrics.IncSMTotalRemoveConnectionsForWS(c.session.clientKey, c.addr.Network(), c.addr.String())
	c.writeErr(err)
	c.doTunnelClose(err)
}

func (c *connection) doTunnelClose(err error) {
	c.Lock()
	defer c.Unlock()

	if c.err != nil {
		return
	}

	c.err = err
	if c.err == nil {
		c.err = io.ErrClosedPipe
	}

	close(c.buf)
}

func (c *connection) tunnelWriter() io.Writer {
	return chanWriter{conn: c, C: c.buf}
}

// Close xxx
func (c *connection) Close() error {
	c.session.closeConnection(c.connID, io.EOF)
	return nil
}

func (c *connection) copyData(b []byte) int {
	n := copy(b, c.readBuf)
	c.readBuf = c.readBuf[n:]
	return n
}

// Read 用于常见IO
func (c *connection) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	n := c.copyData(b)
	if n > 0 {
		metrics.AddSMTotalReceiveBytesOnWS(c.session.clientKey, float64(n))
		return n, nil
	}

	next, ok := <-c.buf
	if !ok {
		err := io.EOF
		c.Lock()
		if c.err != nil {
			err = c.err
		}
		c.Unlock()
		return 0, err
	}

	c.readBuf = next
	n = c.copyData(b)
	metrics.AddSMTotalReceiveBytesOnWS(c.session.clientKey, float64(n))
	return n, nil
}

// Write 用于常见IO
func (c *connection) Write(b []byte) (int, error) {
	c.Lock()
	if c.err != nil {
		defer c.Unlock()
		return 0, c.err
	}
	c.Unlock()

	deadline := int64(0)
	if !c.writeDeadline.IsZero() {
		deadline = c.writeDeadline.Sub(time.Now()).Nanoseconds() / 1000000
	}
	msg := newMessage(c.connID, deadline, b)
	metrics.AddSMTotalTransmitBytesOnWS(c.session.clientKey, float64(len(msg.Bytes())))
	return c.session.writeMessage(msg)
}

func (c *connection) writeErr(err error) {
	if err != nil {
		msg := newErrorMessage(c.connID, err)
		metrics.AddSMTotalTransmitErrorBytesOnWS(c.session.clientKey, float64(len(msg.Bytes())))
		c.session.writeMessage(msg)
	}
}

// LocalAddr xxx
func (c *connection) LocalAddr() net.Addr {
	return c.addr
}

// RemoteAddr xxx
func (c *connection) RemoteAddr() net.Addr {
	return c.addr
}

// SetDeadline xxx
func (c *connection) SetDeadline(t time.Time) error {
	if err := c.SetReadDeadline(t); err != nil {
		return err
	}
	return c.SetWriteDeadline(t)
}

// SetReadDeadline xxx
func (c *connection) SetReadDeadline(t time.Time) error {
	return nil
}

// SetWriteDeadline xxx
func (c *connection) SetWriteDeadline(t time.Time) error {
	c.writeDeadline = t
	return nil
}

type addr struct {
	proto   string
	address string
}

// Network xxx
func (a addr) Network() string {
	return a.proto
}

// String 用于打印
func (a addr) String() string {
	return a.address
}

type chanWriter struct {
	conn *connection
	C    chan []byte
}

// Write 用于常见IO
func (c chanWriter) Write(buf []byte) (int, error) {
	c.conn.Lock()
	defer c.conn.Unlock()

	if c.conn.err != nil {
		return 0, c.conn.err
	}

	newBuf := make([]byte, len(buf))
	copy(newBuf, buf)
	buf = newBuf

	select {
	// must copy the buffer
	case c.C <- buf:
		return len(buf), nil
	default:
		select {
		case c.C <- buf:
			return len(buf), nil
		case <-time.After(15 * time.Second):
			return 0, errors.New("backed up reader")
		}
	}
}
