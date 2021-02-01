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
	"net"
	"time"
)

// Dialer dialer for tunnel request
type Dialer func(network, address string) (net.Conn, error)

// HasSession to see if server has session for the client key
func (s *Server) HasSession(clientKey string) bool {
	_, err := s.sessions.getDialer(clientKey, 0)
	return err == nil
}

// Dial do dial
func (s *Server) Dial(clientKey string, deadline time.Duration, proto, address string) (net.Conn, error) {
	d, err := s.sessions.getDialer(clientKey, deadline)
	if err != nil {
		return nil, err
	}

	return d(proto, address)
}

// Dialer get dialer for client key
func (s *Server) Dialer(clientKey string, deadline time.Duration) Dialer {
	return func(proto, address string) (net.Conn, error) {
		return s.Dial(clientKey, deadline, proto, address)
	}
}
