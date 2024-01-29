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

package manager

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"sync"
	"time"
	"unicode"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

// EndOfTransmission xxx
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
	once          sync.Once
	wsConn        *websocket.Conn
	bindMgr       *ConsoleManager
	resizeMsgChan chan *types.TerminalSize
	inputMsgChan  <-chan wsMessage
	outputMsgChan chan []byte
	hideBanner    bool
}

// NewRemoteStreamConn :
func NewRemoteStreamConn(ctx context.Context, wsConn *websocket.Conn, mgr *ConsoleManager,
	initTerminalSize *types.TerminalSize, hideBanner bool) *RemoteStreamConn {
	conn := &RemoteStreamConn{
		ctx:           ctx,
		wsConn:        wsConn,
		bindMgr:       mgr,
		resizeMsgChan: make(chan *types.TerminalSize, 1), // 放入初始宽高
		outputMsgChan: make(chan []byte),
		hideBanner:    hideBanner,
	}

	// 初始化命令行宽和高
	conn.resizeMsgChan <- initTerminalSize

	return conn
}

// readInputMsg xxx
func (r *RemoteStreamConn) readInputMsg() <-chan wsMessage {
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

// handleResizeMsg : 处理 Resize 数据流
func (r *RemoteStreamConn) handleResizeMsg(msg []byte) (*types.TerminalSize, error) {
	resizeMsg := types.TerminalSize{}

	// 解析Json数据
	err := json.Unmarshal(msg, &resizeMsg)
	if err != nil {
		return nil, err
	}

	return &resizeMsg, nil
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
		resizeMsg, resizeErr := r.handleResizeMsg(decodeMsg)
		if resizeErr != nil {
			return nil, resizeErr
		}

		r.resizeMsgChan <- resizeMsg

		if err = r.bindMgr.HandleResizeMsg(resizeMsg); err != nil {
			return nil, err
		}

		return nil, nil
	}

	inputMsg, err := r.bindMgr.HandleInputMsg(decodeMsg)
	if err != nil {
		return nil, nil
	}

	return inputMsg, nil
}

// Read : executor 回调读取 web 端的输入, 主动断开链接逻辑
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

	r.outputMsgChan <- outputMsg
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

// Close xxx
func (r *RemoteStreamConn) Close() {
	r.once.Do(func() {
		close(r.outputMsgChan)
		close(r.resizeMsgChan)
	})
}

// Run xxx
func (r *RemoteStreamConn) Run(c *gin.Context) error {
	pingInterval := time.NewTicker(10 * time.Second)
	defer pingInterval.Stop()

	guideMessages := helloMessage(c, r.bindMgr.podCtx.Source)
	notSendMsg := true

	for {
		select {
		case <-r.ctx.Done():
			logger.Infof("close %s RemoteStreamConn done", r.bindMgr.podCtx.PodName)
			return nil
		case output, ok := <-r.outputMsgChan:
			if !ok {
				logger.Infof("close %s RemoteStreamConn done by chan", r.bindMgr.podCtx.PodName)
				return nil
			}
			// 收到首个字节才发送 hello 信息
			if notSendMsg && !r.hideBanner {
				r.bindMgr.HandlePostOutputMsg([]byte(guideMessages))

				if err := PreparedGuideMessage(r.ctx, r.wsConn, guideMessages); err != nil {
					return err
				}
				notSendMsg = false
			}

			r.bindMgr.HandlePostOutputMsg(output)

			outputMsg := []byte(base64.StdEncoding.EncodeToString(output))
			if err := r.wsConn.WriteMessage(websocket.TextMessage, outputMsg); err != nil {
				return err
			}
		case <-pingInterval.C: // 定时主动发送 ping
			if err := r.wsConn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return errors.Wrap(err, "ping")
			}
		}
	}
}

// WaitStreamDone : stream 流处理
func (r *RemoteStreamConn) WaitStreamDone(podCtx *types.PodContext) error {
	defer r.Close()

	k8sClient, err := k8sclient.GetK8SClientByClusterId(podCtx.AdminClusterId)
	if err != nil {
		return err
	}

	req := k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podCtx.PodName).
		Namespace(podCtx.Namespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Command:   podCtx.Commands,
		Container: podCtx.ContainerName,
		Stdin:     true,
		Stdout:    true,
		Stderr:    false, // kubectl 默认 stderr 未设置, virtual-kubelet 节点 stderr 和 tty 不能同时为 true
		TTY:       true,
	}, scheme.ParameterCodec)

	k8sConfig := k8sclient.GetK8SConfigByClusterId(podCtx.AdminClusterId)
	executor, err := remotecommand.NewSPDYExecutor(k8sConfig, "POST", req.URL())
	if err != nil {
		logger.Warnf("start remote stream error, err: %s", err)
		return err
	}

	// start reading
	r.inputMsgChan = r.readInputMsg()

	// Stream Copy IO, Wait Here
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             r,
		Stdout:            r,
		Stderr:            r,
		TerminalSizeQueue: r,
		Tty:               true,
	})

	if err != nil {
		logger.Warnf("close %s WaitStreamDone err, %s", podCtx.PodName, err)
		return err
	}

	logger.Infof("close %s WaitStreamDone done", podCtx.PodName)
	return nil
}

// PreparedGuideMessage , 使用 PreparedMessage, gorilla 有缓存, 提高性能
func PreparedGuideMessage(ctx context.Context, ws *websocket.Conn, guideMessages string) error {
	preparedMsg, err := websocket.NewPreparedMessage(websocket.TextMessage, []byte(base64.StdEncoding.EncodeToString(
		[]byte(guideMessages))))
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
	closeMsg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, errMsg.Error())
	deadline := time.Now().Add(time.Second * 5) // 最迟 5 秒
	if err := ws.WriteControl(websocket.CloseMessage, closeMsg, deadline); err != nil {
		logger.Warnf("gracefully close websocket [%s] error: %s", errMsg, err)
		return
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

func helloMessage(c *gin.Context, source string) string {
	var guideMsg []string
	var messages []string

	if source == "mgr" {
		guideMsg = []string{i18n.T(c,
			"支持常用Bash快捷键; Windows下Ctrl-W为关闭窗口快捷键, 请使用Alt-W代替; 使用Alt-Num切换Tab")}
		guideMsg = append(guideMsg, config.G.WebConsole.GuideDocLinks...)
	} else {
		guideMsg = []string{i18n.T(c, "支持常用Bash快捷键; Windows下Ctrl-W为关闭窗口快捷键, 请使用Alt-W代替")}
		guideMsg = append(guideMsg, config.G.WebConsole.GuideDocLinks...)
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
			length++
		}
	}

	return length
}
