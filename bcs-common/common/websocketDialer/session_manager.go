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
 */

package websocketDialer

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/Tencent/bk-bcs/bcs-common/common/websocketDialer/metrics"
)

type sessionListener interface {
	sessionAdded(clientKey string, sessionKey int64)
	sessionRemoved(clientKey string, sessionKey int64)
}

type sessionManager struct {
	sync.Mutex
	clients   map[string][]*Session
	peers     map[string][]*Session
	listeners map[sessionListener]bool
}

func newSessionManager() *sessionManager {
	return &sessionManager{
		clients:   map[string][]*Session{},
		peers:     map[string][]*Session{},
		listeners: map[sessionListener]bool{},
	}
}

func toDialer(s *Session, prefix string, deadline time.Duration) Dialer {
	return func(proto, address string) (net.Conn, error) {
		if prefix == "" {
			return s.serverConnect(deadline, proto, address)
		}
		return s.serverConnect(deadline, prefix+"::"+proto, address)
	}
}

func (sm *sessionManager) removeListener(listener sessionListener) {
	sm.Lock()
	defer sm.Unlock()

	delete(sm.listeners, listener)
}

func (sm *sessionManager) addListener(listener sessionListener) {
	sm.Lock()
	defer sm.Unlock()

	sm.listeners[listener] = true

	for k, sessions := range sm.clients {
		for _, session := range sessions {
			listener.sessionAdded(k, session.sessionKey)
		}
	}

	for k, sessions := range sm.peers {
		for _, session := range sessions {
			listener.sessionAdded(k, session.sessionKey)
		}
	}
}

func (sm *sessionManager) getDialer(clientKey string, deadline time.Duration) (Dialer, error) {
	sm.Lock()
	defer sm.Unlock()

	sessions := sm.clients[clientKey]
	if len(sessions) > 0 {
		return toDialer(sessions[0], "", deadline), nil
	}

	for _, sessions := range sm.peers {
		for _, session := range sessions {
			session.Lock()
			keys := session.remoteClientKeys[clientKey]
			session.Unlock()
			if len(keys) > 0 {
				return toDialer(session, clientKey, deadline), nil
			}
		}
	}

	return nil, fmt.Errorf("failed to find Session for client %s", clientKey)
}

func (sm *sessionManager) add(clientKey string, conn *websocket.Conn, peer bool) *Session {
	sessionKey := rand.Int63() // nolint
	session := newSession(sessionKey, clientKey, conn)

	sm.Lock()
	defer sm.Unlock()

	if peer {
		sm.peers[clientKey] = append(sm.peers[clientKey], session)
	} else {
		sm.clients[clientKey] = append(sm.clients[clientKey], session)
	}
	metrics.IncSMTotalAddWS(clientKey, peer)

	for l := range sm.listeners {
		l.sessionAdded(clientKey, session.sessionKey)
	}

	return session
}

func (sm *sessionManager) remove(s *Session) {
	var isPeer bool
	sm.Lock()
	defer sm.Unlock()

	for i, store := range []map[string][]*Session{sm.clients, sm.peers} {
		var newSessions []*Session

		for _, v := range store[s.clientKey] {
			if v.sessionKey == s.sessionKey {
				if i == 0 {
					isPeer = false
				} else {
					isPeer = true
				}
				metrics.IncSMTotalRemoveWS(s.clientKey, isPeer)
				continue
			}
			newSessions = append(newSessions, v)
		}

		if len(newSessions) == 0 {
			delete(store, s.clientKey)
		} else {
			store[s.clientKey] = newSessions
		}
	}

	for l := range sm.listeners {
		l.sessionRemoved(s.clientKey, s.sessionKey)
	}

	s.Close()
}
