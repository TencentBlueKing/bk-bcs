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
	"net/http"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

var (
	errFailedAuth       = errors.New("failed authentication")
	errWrongMessageType = errors.New("wrong websocket message type")
)

// Authorizer authorizer function type
type Authorizer func(req *http.Request) (clientKey string, authed bool, err error)

// CleanCredentials clean credential function type
type CleanCredentials func(clientKey string)

// ErrorWriter error writer function type
type ErrorWriter func(rw http.ResponseWriter, req *http.Request, code int, err error)

// DefaultErrorWriter default error writer
func DefaultErrorWriter(rw http.ResponseWriter, req *http.Request, code int, err error) {
	rw.WriteHeader(code)
	rw.Write([]byte(err.Error()))
}

// Server the server for tunnel
type Server struct {
	PeerID           string
	PeerToken        string
	authorizer       Authorizer
	cleanCredentials CleanCredentials
	errorWriter      ErrorWriter
	sessions         *sessionManager
	peers            map[string]peer
	peerLock         sync.Mutex
}

// New create new tunnel server
func New(auth Authorizer, errorWriter ErrorWriter, clean CleanCredentials) *Server {
	return &Server{
		peers:            map[string]peer{},
		authorizer:       auth,
		cleanCredentials: clean,
		errorWriter:      errorWriter,
		sessions:         newSessionManager(),
	}
}

// ServeHTTP handle http request
func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	clientKey, authed, peer, err := s.auth(req)
	if err != nil {
		s.errorWriter(rw, req, 400, err)
		return
	}
	if !authed {
		s.errorWriter(rw, req, 401, errFailedAuth)
		return
	}

	blog.Infof("Handling backend connection request [%s]", clientKey)

	upgrader := websocket.Upgrader{
		HandshakeTimeout: 5 * time.Second,
		CheckOrigin:      func(r *http.Request) bool { return true },
		Error:            s.errorWriter,
	}

	wsConn, err := upgrader.Upgrade(rw, req, nil)
	if err != nil {
		s.errorWriter(rw, req, 400, errors.Wrapf(err, "Error during upgrade for host [%v]", clientKey))
		return
	}

	session := s.sessions.add(clientKey, wsConn, peer)
	defer s.sessions.remove(session)

	// Don't need to associate req.Context() to the Session, it will cancel otherwise
	code, err := session.Serve(context.Background())
	if err != nil {
		// Hijacked so we can't write to the client
		blog.Infof("error in remotedialer server [%d]: %s", code, err.Error())
		// clean credentials from db
		s.cleanCredentials(clientKey)
	}
}

// auth authorize a peer client
func (s *Server) auth(req *http.Request) (clientKey string, authed, peer bool, err error) {
	id := req.Header.Get(ID)
	token := req.Header.Get(Token)
	if id != "" && token != "" {
		// peer authentication
		s.peerLock.Lock()
		p, ok := s.peers[id]
		s.peerLock.Unlock()

		if ok && p.token == token {
			return id, true, true, nil
		}
	}

	id, authed, err = s.authorizer(req)
	return id, authed, false, err
}
