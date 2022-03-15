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
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/i18n"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// End-of-Transmission character ctrl-d
const EndOfTransmission = "\u0004"

type wsMessage struct {
	msgType int
	msg     []byte
	err     error
}

// RemoteStreamConn 流式处理器
type RemoteStreamConn struct {
	ctx           context.Context
	wsConn        *websocket.Conn
	bindMgr       *ConsoleManager
	resizeMsgChan chan *TerminalSize
	inputMsgChan  <-chan wsMessage
	outputMsgChan chan []byte
	once          sync.Once
}

// NewRemoteStreamConn :
func NewRemoteStreamConn(ctx context.Context, wsConn *websocket.Conn, mgr *ConsoleManager, initTerminalSize *TerminalSize) *RemoteStreamConn {
	conn := &RemoteStreamConn{
		ctx:           ctx,
		wsConn:        wsConn,
		bindMgr:       mgr,
		resizeMsgChan: make(chan *TerminalSize, 1), // 放入初始宽高
		outputMsgChan: make(chan []byte),
	}

	// 初始化命令行宽和高
	conn.resizeMsgChan <- initTerminalSize

	return conn
}

func (r *RemoteStreamConn) ReadInputMsg() <-chan wsMessage {
	inputMsgChan := make(chan wsMessage)
	go func() {
		defer close(inputMsgChan)
		for {
			msgType, msg, err := r.wsConn.ReadMessage()
			inputMsgChan <- wsMessage{
				msgType: msgType,
				msg:     msg,
				err:     err,
			}
			if err != nil {
				break
			}
		}
	}()
	return inputMsgChan
}

// HandleMsg Msg 处理
func (r *RemoteStreamConn) HandleMsg(msgType int, msg []byte) ([]byte, error) {
	// 只处理文本数据
	if msgType != websocket.TextMessage {
		return nil, nil
	}

	// body 解析base64数据
	decodeMsg, err := base64.StdEncoding.DecodeString(string(msg[1:]))
	if err != nil {
		return nil, nil
	}

	// 第一个字符串为 channel
	channel := string(msg[0])
	if channel == ResizeChannel {
		resizeMsg, err := r.bindMgr.HandleResizeMsg(decodeMsg)
		if err != nil {
			return nil, nil
		}

		r.resizeMsgChan <- resizeMsg
		return nil, nil
	}

	inputMsg, err := r.bindMgr.HandleInputMsg(decodeMsg)
	if err != nil {
		return nil, nil
	}
	return inputMsg, nil
}

// Read : executor 回调读取 web 端的输入
func (r *RemoteStreamConn) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return copy(p, EndOfTransmission), r.ctx.Err()

	case m := <-r.inputMsgChan:
		if m.err != nil {
			return copy(p, EndOfTransmission), m.err
		}

		out, err := r.HandleMsg(m.msgType, m.msg)
		if err != nil {
			return copy(p, EndOfTransmission), err
		}
		return copy(p, out), nil
	}
}

// Write : executor 回调向 web 端输出
func (r *RemoteStreamConn) Write(p []byte) (int, error) {
	msg := make([]byte, len(p))
	copy(msg, p)

	outputMsg, err := r.bindMgr.HandleOutputMsg(msg)
	if err != nil {
		return 0, nil
	}

	output := []byte(base64.StdEncoding.EncodeToString(outputMsg))
	r.outputMsgChan <- output
	return len(p), nil
}

// Next : executor回调获取web是否resize
func (r *RemoteStreamConn) Next() *remotecommand.TerminalSize {
	resizeMsg, ok := <-r.resizeMsgChan
	if !ok {
		return nil
	}

	return &remotecommand.TerminalSize{
		Width:  resizeMsg.Cols,
		Height: resizeMsg.Rows,
	}
}

func (r *RemoteStreamConn) Close() {
	r.once.Do(func() {
		close(r.outputMsgChan)
		close(r.resizeMsgChan)
	})
}

func (r *RemoteStreamConn) Run() error {
	pingInterval := time.NewTicker(10 * time.Second)
	defer pingInterval.Stop()

	guideMessages := helloMessage(r.bindMgr.PodCtx.Source)

	PreparedGuideMessage(r.ctx, r.wsConn, guideMessages)

	for {
		select {
		case <-r.ctx.Done():
			logger.Infof("close %s RemoteStreamConn done", r.bindMgr.PodCtx.PodName)
			return r.ctx.Err()
		case output := <-r.outputMsgChan:
			if err := r.wsConn.WriteMessage(websocket.TextMessage, output); err != nil {
				return err
			}
		case <-pingInterval.C: // 定时主动发送 ping
			if err := r.wsConn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return errors.Wrap(err, "ping")
			}
		}
	}
}

// WaitStreamDone: stream 流处理
func (r *RemoteStreamConn) WaitStreamDone(bcsConf *config.BCSConf, podCtx *types.PodContext) error {
	host := fmt.Sprintf("%s/clusters/%s", bcsConf.Host, podCtx.ClusterId)
	k8sConfig := &rest.Config{
		Host:        host,
		BearerToken: bcsConf.Token,
	}
	k8sClient, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return err
	}

	req := k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podCtx.PodName).
		Namespace(podCtx.Namespace).
		SubResource("exec")

	req.VersionedParams(&v1.PodExecOptions{
		Command:   podCtx.Commands,
		Container: podCtx.ContainerName,
		Stdin:     true,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(k8sConfig, "POST", req.URL())
	if err != nil {
		logger.Warnf("start remote stream error, err: %s", err)
		return err
	}

	// start reading
	r.inputMsgChan = r.ReadInputMsg()

	// Stream Copy IO, Wait Here
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             r,
		Stdout:            r,
		Stderr:            r,
		TerminalSizeQueue: r,
		Tty:               true,
	})

	if err != nil {
		logger.Warnf("remote stream %s closed, err: %s", podCtx.PodName, err)
		return err
	}

	logger.Info("remote stream %s closed normal", podCtx.PodName)

	return nil
}

// PreparedGuideMessage, 使用 PreparedMessage, gorilla 有缓存, 提高性能
func PreparedGuideMessage(ctx context.Context, ws *websocket.Conn, guideMessages string) error {
	preparedMsg, err := websocket.NewPreparedMessage(websocket.TextMessage, []byte(base64.StdEncoding.EncodeToString([]byte(guideMessages))))
	if err != nil {
		return err
	}
	if err := ws.WritePreparedMessage(preparedMsg); err != nil {
		return err
	}
	return nil
}

// GracefulCloseWebSocket : 优雅停止 websocket 连接
func GracefulCloseWebSocket(ctx context.Context, ws *websocket.Conn, connected bool, errMsg error) {
	if err := ws.WriteControl(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, errMsg.Error()),
		time.Now().Add(time.Second*5), // 最迟 5 秒
	); err != nil {
		logger.Warnf("gracefully close websocket error, %s", err)
	}

	// 如果没有建立双向连接前, 需要ReadMessage才能正常结束
	if !connected {
		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				return
			}
		}
	}

	<-ctx.Done()
}

func helloMessage(source string) string {

	var guideMsg []string
	var messages []string

	if source == "mgr" {
		guideMsg = []string{
			config.G.WebConsole.GuideDocLink,
			i18n.GetMessage("mgrGuideMessage"),
		}
	} else {
		guideMsg = []string{
			config.G.WebConsole.GuideDocLink,
			i18n.GetMessage("guideMessage"),
		}
	}

	// 两边一个#字符，加一个空格
	var width int
	for _, s := range guideMsg {
		if ZhLength(s)+3 > width {
			width = ZhLength(s) + 3
		}
	}

	messages = append(messages, strings.Repeat("#", width))
	leftSpace := (width - 2 - len(helloBcsMessage)) / 2
	rightSpace := width - 2 - leftSpace - len(helloBcsMessage)
	console := "#" + strings.Repeat(" ", leftSpace) + helloBcsMessage + strings.Repeat(" ", rightSpace) + "#"
	messages = append(messages, console)
	messages = append(messages, strings.Repeat("#", width))

	for _, s := range guideMsg {
		messages = append(messages, "#"+s+strings.Repeat(" ", width-ZhLength(s)-2)+"#")
	}

	messages = append(messages, strings.Repeat("#", width))

	return strings.Join(messages, "\r\n") + "\r\n"
}

// ZhLength 计算中文字符串长度, 中文为2个长度
func ZhLength(str string) int {

	var length int

	for _, i := range str {
		if unicode.Is(unicode.Han, i) {
			length += 2
		} else {
			length += 1
		}
	}

	return length
}
