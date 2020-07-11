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
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer/metrics"
	"github.com/gorilla/websocket"
)

var (
	Token = "BCS-API-Tunnel-Token"
	ID    = "BCS-API-Tunnel-ID"
)

func (s *Server) AddPeer(url, id, token string) {
	if s.PeerID == "" || s.PeerToken == "" {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	peer := peer{
		url:    url,
		id:     id,
		token:  token,
		cancel: cancel,
	}

	blog.Infof("Adding peer %s, %s", url, id)

	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	if p, ok := s.peers[id]; ok {
		if p.equals(peer) {
			return
		}
		p.cancel()
	}

	s.peers[id] = peer
	go peer.start(ctx, s)
}

func (s *Server) RemovePeer(id string) {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	if p, ok := s.peers[id]; ok {
		blog.Infof("Removing peer %s", id)
		p.cancel()
	}
	delete(s.peers, id)
}

type peer struct {
	url, id, token string
	cancel         func()
}

func (p peer) equals(other peer) bool {
	return p.url == other.url &&
		p.id == other.id &&
		p.token == other.token
}

func (p *peer) start(ctx context.Context, s *Server) {
	headers := http.Header{
		ID:    {s.PeerID},
		Token: {s.PeerToken},
	}

	dialer := &websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		HandshakeTimeout: HandshakeTimeOut,
	}

outer:
	for {
		select {
		case <-ctx.Done():
			break outer
		default:
		}

		metrics.IncSMTotalAddPeerAttempt(p.id)
		// wait for peer server to be ok
		time.Sleep(2 * time.Second)
		ws, _, err := dialer.Dial(p.url, headers)
		if err != nil {
			blog.Errorf("Failed to connect to peer %s [local ID=%s]: %s", p.url, s.PeerID, err.Error())
			time.Sleep(5 * time.Second)
			continue
		}
		metrics.IncSMTotalPeerConnected(p.id)

		session := NewClientSession(func(string, string) bool { return true }, ws)
		session.dialer = func(network, address string) (net.Conn, error) {
			parts := strings.SplitN(network, "::", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid clientKey/proto: %s", network)
			}
			return s.Dial(parts[0], 15*time.Second, parts[1], address)
		}

		s.sessions.addListener(session)
		_, err = session.Serve(context.Background())
		s.sessions.removeListener(session)
		session.Close()

		if err != nil {
			blog.Errorf("Failed to serve peer connection %s: %s", p.id, err.Error())
		}

		ws.Close()
		time.Sleep(5 * time.Second)
	}
}
