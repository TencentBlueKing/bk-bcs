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
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"

	"github.com/gorilla/websocket"
	v1 "k8s.io/api/core/v1"
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

// WsMessage websocket消息
type WsMessage struct {
	MessageType int
	Data        []byte
}

type wsConn struct {
	conn      *websocket.Conn
	inChan    chan *WsMessage // 读取队列
	outChan   chan *WsMessage // 发送队列
	mutex     sync.Mutex      // 避免重复关闭管道
	isClosed  bool
	closeChan chan byte // 关闭通知
}

func (c *wsConn) Read(p []byte) (n int, err error) {
	_, rc, err := c.conn.NextReader()
	if err != nil {
		return 0, err
	}
	return rc.Read(p)
}

// 读取协程
func (c *wsConn) wsReadLoop() {
	for {
		// 读一条message
		msgType, data, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		// 放入请求队列
		c.inChan <- &WsMessage{
			msgType,
			data,
		}
	}
}

// 发送协程
func (c *wsConn) wsWriteLoop() {
	// 服务端返回给页面的数据
	for {
		select {
		// 取一个应答
		case msg := <-c.outChan:
			// 写给web  websocket

			if err := c.conn.WriteMessage(msg.MessageType, msg.Data); err != nil {
				break
			}
		case <-c.closeChan:
			c.wsClose()
		}
	}
}

// 关闭连接
func (c *wsConn) wsClose() {
	c.conn.Close()
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if !c.isClosed {
		c.isClosed = true
		close(c.closeChan)
	}
}

// 发送返回消息到协程
func (c *wsConn) wsWrite(messageType int, data []byte) (err error) {

	select {
	case c.outChan <- &WsMessage{messageType, data}:

	case <-c.closeChan:
		err = errors.New("WsWrite websocket closed")
		break
	}
	return
}

func (c *wsConn) WsRead() (msg *WsMessage, err error) {

	select {
	case msg = <-c.inChan:
		return
	case <-c.closeChan:
		err = errors.New("WsRead websocket closed")
		break
	}
	return

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

// ssh流式处理器
type streamHandler struct {
	wsConn      *wsConn
	resizeEvent chan remotecommand.TerminalSize
}

//Next executor回调获取web是否resize
func (handler *streamHandler) Next() (size *remotecommand.TerminalSize) {
	ret := <-handler.resizeEvent
	size = &ret
	return
}

// executor回调读取web端的输入
func (handler *streamHandler) Read(p []byte) (size int, err error) {

	// 读web发来的输入
	msg, err := handler.wsConn.WsRead()
	if err != nil {
		handler.wsConn.wsClose()
		return
	}

	xtermMsg := types.XtermMessage{}
	switch string(msg.Data[0]) {
	case "0":
		xtermMsg.MsgType = "input"
		xtermMsg.Input = string(msg.Data[1:])
		break
	case "4":
		xtermMsg.MsgType = "resize"
		err = json.Unmarshal(msg.Data[1:], &xtermMsg)
		if err != nil {
			return 0, err
		}

		// 放到channel里，等remotecommand executor调用Next取走
		handler.resizeEvent <- remotecommand.TerminalSize{Width: xtermMsg.Cols, Height: xtermMsg.Rows}
	default:
		return 0, nil
	}

	size = len(xtermMsg.Input)
	copy(p, xtermMsg.Input)
	return
}

// executor回调向web端输出
func (handler *streamHandler) Write(p []byte) (size int, err error) {
	// 产生副本
	copyData := make([]byte, len(p))
	copy(copyData, p)
	size = len(p)
	err = handler.wsConn.wsWrite(websocket.TextMessage, copyData)
	return
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

	wsConn := &wsConn{
		conn:      ws,
		inChan:    make(chan *WsMessage, 1000),
		outChan:   make(chan *WsMessage, 1000),
		closeChan: make(chan byte),
		isClosed:  false,
	}

	// 页面读入输入 协程
	go wsConn.wsReadLoop()
	// 服务端返回数据 协程
	go wsConn.wsWriteLoop()

	for _, i := range ConsoleCopywritingFailed {
		err := ws.WriteMessage(websocket.TextMessage, []byte(i))
		if err != nil {
			ResponseJSON(w, http.StatusInternalServerError, errMsg{err.Error()})
			return
		}
	}

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

	// 执行连接
	err = m.startExec(wsConn, conf)
	if err != nil {
		blog.Errorf("start exec failed for pod(%s) : %s", conf.PodName, err.Error())
		ResponseJSON(w, http.StatusBadRequest, errMsg{err.Error()})
		return
	}

	ResponseJSON(w, http.StatusSwitchingProtocols, nil)
}

func (m *manager) startExec(ws *wsConn, conf *types.WebSocketConfig) error {

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
	handler := &streamHandler{wsConn: ws, resizeEvent: make(chan remotecommand.TerminalSize)}
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             handler,
		Stdout:            handler,
		Stderr:            handler,
		TerminalSizeQueue: handler,
		Tty:               true,
	})
	if err != nil {
		blog.Errorf("startExec failed for Stream err %v:", err)
		return err
	}

	return nil
}
