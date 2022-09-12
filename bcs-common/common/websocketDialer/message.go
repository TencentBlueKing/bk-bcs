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
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Data message type for data
	Data messageType = iota + 1
	// Connect message type for connection
	Connect
	// Error message type for error
	Error
	// AddClient message for adding client
	AddClient
	// RemoveClient message for removing client
	RemoveClient
)

var (
	idCounter int64
)

func init() {
	r := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	idCounter = r.Int63()
}

type messageType int64

type message struct {
	id          int64
	err         error
	connID      int64
	deadline    int64
	messageType messageType
	bytes       []byte
	body        io.Reader
	proto       string
	address     string
}

func nextid() int64 {
	return atomic.AddInt64(&idCounter, 1)
}

func newMessage(connID int64, deadline int64, bytes []byte) *message {
	return &message{
		id:          nextid(),
		connID:      connID,
		deadline:    deadline,
		messageType: Data,
		bytes:       bytes,
	}
}

func newConnect(connID int64, deadline time.Duration, proto, address string) *message {
	return &message{
		id:          nextid(),
		connID:      connID,
		deadline:    deadline.Nanoseconds() / 1000000,
		messageType: Connect,
		bytes:       []byte(fmt.Sprintf("%s/%s", proto, address)),
		proto:       proto,
		address:     address,
	}
}

func newErrorMessage(connID int64, err error) *message {
	return &message{
		id:          nextid(),
		err:         err,
		connID:      connID,
		messageType: Error,
		bytes:       []byte(err.Error()),
	}
}

func newAddClient(client string) *message {
	return &message{
		id:          nextid(),
		messageType: AddClient,
		address:     client,
		bytes:       []byte(client),
	}
}

func newRemoveClient(client string) *message {
	return &message{
		id:          nextid(),
		messageType: RemoveClient,
		address:     client,
		bytes:       []byte(client),
	}
}

func newServerMessage(reader io.Reader) (*message, error) {
	buf := bufio.NewReader(reader)

	id, err := binary.ReadVarint(buf)
	if err != nil {
		return nil, err
	}

	connID, err := binary.ReadVarint(buf)
	if err != nil {
		return nil, err
	}

	mType, err := binary.ReadVarint(buf)
	if err != nil {
		return nil, err
	}

	m := &message{
		id:          id,
		messageType: messageType(mType),
		connID:      connID,
		body:        buf,
	}

	if m.messageType == Data || m.messageType == Connect {
		deadline, err := binary.ReadVarint(buf)
		if err != nil {
			return nil, err
		}
		m.deadline = deadline
	}

	if m.messageType == Connect {
		bytes, err := ioutil.ReadAll(io.LimitReader(buf, 100))
		if err != nil {
			return nil, err
		}
		parts := strings.SplitN(string(bytes), "/", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("failed to parse connect address")
		}
		m.proto = parts[0]
		m.address = parts[1]
		m.bytes = bytes
	} else if m.messageType == AddClient || m.messageType == RemoveClient {
		bytes, err := ioutil.ReadAll(io.LimitReader(buf, 100))
		if err != nil {
			return nil, err
		}
		m.address = string(bytes)
		m.bytes = bytes
	}

	return m, nil
}

// Err xxx
func (m *message) Err() error {
	if m.err != nil {
		return m.err
	}
	bytes, err := ioutil.ReadAll(io.LimitReader(m.body, 100))
	if err != nil {
		return err
	}

	str := string(bytes)
	if str == "EOF" {
		m.err = io.EOF
	} else {
		m.err = errors.New(str)
	}
	return m.err
}

// Bytes xxx
func (m *message) Bytes() []byte {
	return append(m.header(), m.bytes...)
}

func (m *message) header() []byte {
	buf := make([]byte, 24)
	offset := 0
	offset += binary.PutVarint(buf[offset:], m.id)
	offset += binary.PutVarint(buf[offset:], m.connID)
	offset += binary.PutVarint(buf[offset:], int64(m.messageType))
	if m.messageType == Data || m.messageType == Connect {
		offset += binary.PutVarint(buf[offset:], m.deadline)
	}
	return buf[:offset]
}

// Read 用于常见IO
func (m *message) Read(p []byte) (int, error) {
	return m.body.Read(p)
}

// WriteTo xxx
func (m *message) WriteTo(wsConn *wsConn) (int, error) {
	err := wsConn.WriteMessage(websocket.BinaryMessage, m.Bytes())
	return len(m.bytes), err
}

// String 用于打印
func (m *message) String() string {
	switch m.messageType {
	case Data:
		if m.body == nil {
			return fmt.Sprintf("%d DATA         [%d]: %d bytes: %s", m.id, m.connID, len(m.bytes), string(m.bytes))
		}
		return fmt.Sprintf("%d DATA         [%d]: buffered", m.id, m.connID)
	case Error:
		return fmt.Sprintf("%d ERROR        [%d]: %s", m.id, m.connID, m.Err())
	case Connect:
		return fmt.Sprintf("%d CONNECT      [%d]: %s/%s deadline %d", m.id, m.connID, m.proto, m.address, m.deadline)
	case AddClient:
		return fmt.Sprintf("%d ADDCLIENT    [%s]", m.id, m.address)
	case RemoveClient:
		return fmt.Sprintf("%d REMOVECLIENT [%s]", m.id, m.address)
	}
	return fmt.Sprintf("%d UNKNOWN[%d]: %d", m.id, m.connID, m.messageType)
}
