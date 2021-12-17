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
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"

	"github.com/gorilla/websocket"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// ConsoleCopywritingFailed is a response string
var ConsoleCopywritingFailed = []string{
	"#######################################################################\r\n",
	"#                    Welcome to the BCS                      #\r\n",
	"#######################################################################\r\n",
}

//DefaultCommand 默认命令, 可以优先使用bash, 如果没有, 回退到sh
var DefaultCommand = []string{
	"/bin/sh",
	"-c",
	"TERM=xterm-256color; export TERM; [ -x /bin/bash ] && (" +
		"[ -x /usr/bin/script ] && /usr/bin/script -q -c \"/bin/bash\" /dev/null || exec /bin/bash) || exec /bin/sh",
}

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type errMsg struct {
	Msg string `json:"msg,omitempty"`
}

type wsConn struct {
	conn *websocket.Conn
}

func newWsConn(conn *websocket.Conn) *wsConn {
	return &wsConn{
		conn: conn,
	}
}

// TODO ws 读写还未转base64
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

// ResponseJSON response to client
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

// StartExec start a websocket exec
func (m *manager) StartExec(w http.ResponseWriter, r *http.Request, conf *types.WebSocketConfig) {
	blog.Debug(fmt.Sprintf("start exec for container pod %s", conf.PodName))

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

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				if err := ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
					return
				}
			}
		}
	}()

	// 确认pod状态 是Running
	if !m.checkPodStatus(conf.PodName, v1.PodPhase("Running")) {
		blog.Errorf("the current status of pod(%s) is not Running", conf.PodName)
		ResponseJSON(w, http.StatusBadRequest, errMsg{"pod 当前状态不是Running, 请重试！"})
		return
	}

	// 执行连接
	err = m.startExec(newWsConn(ws), conf)
	if err != nil {
		blog.Errorf("start exec failed for pod(%s) : %s", conf.PodName, err.Error())
		ResponseJSON(w, http.StatusBadRequest, errMsg{err.Error()})
		return
	}

	ResponseJSON(w, http.StatusSwitchingProtocols, nil)
}

func (m *manager) startExec(ws io.ReadWriter, conf *types.WebSocketConfig) error {

	req := m.k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(conf.PodName).
		Namespace(NAMESPACE).
		SubResource("exec")

	req.VersionedParams(
		&v1.PodExecOptions{
			Command: DefaultCommand,
			Stdin:   true,
			Stdout:  true,
			Stderr:  true,
			TTY:     true,
		},
		scheme.ParameterCodec,
	)

	executor, err := remotecommand.NewSPDYExecutor(m.k8sConfig, "POST", req.URL())
	if err != nil {
		blog.Errorf("startExec failed for NewSPDYExecutor err: %v", err)
		return err
	}

	// Stream
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  ws,
		Stdout: ws,
		Stderr: ws,
		Tty:    true,
	})
	if err != nil {
		blog.Errorf("startExec failed for Stream err %v:", err)
		return err
	}

	return nil
}

// 确认pod状态
func (m *manager) checkPodStatus(podName string, status v1.PodPhase) bool {
	pod, err := m.k8sClient.CoreV1().Pods(NAMESPACE).Get(podName, metav1.GetOptions{})
	if err != nil {
		return false
	}

	return pod.Status.Phase == status
}
