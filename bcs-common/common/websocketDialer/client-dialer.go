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
	"io"
	"net"
	"sync"
	"time"
)

func clientDial(dialer Dialer, conn *connection, message *message) {
	defer conn.Close()

	var (
		netConn net.Conn
		err     error
	)

	if dialer == nil {
		netConn, err = net.DialTimeout(message.proto, message.address, time.Duration(message.deadline)*time.Millisecond)
	} else {
		netConn, err = dialer(message.proto, message.address)
	}

	if err != nil {
		conn.tunnelClose(err)
		return
	}
	defer netConn.Close()

	pipe(conn, netConn)
}

func pipe(client *connection, server net.Conn) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	close := func(err error) error {
		if err == nil {
			err = io.EOF
		}
		client.doTunnelClose(err)
		server.Close()
		return err
	}

	go func() {
		defer wg.Done()
		_, err := io.Copy(server, client)
		close(err)
	}()

	_, err := io.Copy(client, server)
	err = close(err)
	wg.Wait()

	// Write tunnel error after no more I/O is happening, just incase messages get out of order
	client.writeErr(err)
}
