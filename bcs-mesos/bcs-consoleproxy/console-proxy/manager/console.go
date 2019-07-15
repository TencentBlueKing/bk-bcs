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

package manager

import (
	"encoding/json"
	"io"
	"net/http"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-mesos/bcs-consoleproxy/console-proxy/types"
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"github.com/gorilla/websocket"
)

type errMsg struct {
	Msg string `json:"msg,omitempty"`
}

type wsConn struct {
	conn *websocket.Conn
}

func (c *wsConn) Read(p []byte) (n int, err error) {
	_, rc, err := c.conn.NextReader()
	if err != nil {
		return 0, err
	}
	return rc.Read(p)
}

func (c *wsConn) Write(p []byte) (n int, err error) {
	wc, err := c.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return 0, err
	}
	defer wc.Close()
	return wc.Write(p)
}

func ResponseJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (m *manager) StartExec(w http.ResponseWriter, r *http.Request, conf *types.WebSocketConfig) {

	blog.Debug(fmt.Sprintf("start exec for container exec_id %s", conf.ExecId))
	upgrader := websocket.Upgrader{
		EnableCompression: true,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	if !websocket.IsWebSocketUpgrade(r) {
		ResponseJSON(w, http.StatusBadRequest, nil)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		ResponseJSON(w, http.StatusBadRequest, errMsg{err.Error()})
		return
	}
	defer ws.Close()

	ws.SetCloseHandler(nil)
	ws.SetPingHandler(nil)

	err = m.startExec(&wsConn{ws}, conf)
	if err != nil {
		blog.Info(err.Error())
		ResponseJSON(w, http.StatusBadRequest, errMsg{err.Error()})
		return
	}

	ResponseJSON(w, http.StatusSwitchingProtocols, nil)
	return
}

func (m *manager) CreateExec(w http.ResponseWriter, r *http.Request, conf *types.WebSocketConfig) {
	blog.Debug(fmt.Sprintf("start create exec for container %s", conf.ContainerId))
	// 创建连接
	exec, err := m.dockerClient.CreateExec(docker.CreateExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          m.conf.Tty,
		Env:          nil,
		Cmd:          conf.Cmd,
		Container:    conf.ContainerId,
		User:         conf.User,
		Privileged:   m.conf.Privilege,
	})

	if err != nil {
		ResponseJSON(w, http.StatusBadRequest, errMsg{err.Error()})
		return
	}

	ResponseJSON(w, http.StatusOK, exec)
	return
}

func (m *manager) startExec(ws io.ReadWriter, conf *types.WebSocketConfig) error {
	fmt.Println("start exec")
	// 执行连接
	err := m.dockerClient.StartExec(conf.ExecId, docker.StartExecOptions{
		InputStream:  ws,
		OutputStream: ws,
		ErrorStream:  ws,
		Detach:       false,
		Tty:          m.conf.Tty,
		RawTerminal:  true,
	})

	if err != nil {
		return err
	}

	return nil
}

func (m *manager) ResizeExec(w http.ResponseWriter, r *http.Request, conf *types.WebSocketConfig) {
	blog.Debug(fmt.Sprintf("start resize for container exec_id %s", conf.ExecId))
	err := m.dockerClient.ResizeExecTTY(conf.ExecId, conf.Height, conf.Width)
	if err != nil {
		ResponseJSON(w, http.StatusBadRequest, errMsg{err.Error()})
		return
	}

	ResponseJSON(w, http.StatusOK, nil)
	return
}
