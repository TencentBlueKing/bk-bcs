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
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pborman/ansi"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/audit"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/k8sclient"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
)

// EndOfTransmission xxx
// End-of-Transmission character ctrl-d
const EndOfTransmission = "\u0004"

var cmdParser = audit.NewCmdParse()

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
	resizeMsgChan chan *TerminalSize
	inputMsgChan  <-chan wsMessage
	outputMsgChan chan []byte
	hideBanner    bool
}

// NewRemoteStreamConn :
func NewRemoteStreamConn(ctx context.Context, wsConn *websocket.Conn, mgr *ConsoleManager,
	initTerminalSize *TerminalSize, hideBanner bool) *RemoteStreamConn {
	conn := &RemoteStreamConn{
		ctx:           ctx,
		wsConn:        wsConn,
		bindMgr:       mgr,
		resizeMsgChan: make(chan *TerminalSize, 1), // 放入初始宽高
		outputMsgChan: make(chan []byte),
		hideBanner:    hideBanner,
	}

	// 初始化命令行宽和高
	if initTerminalSize != nil {
		conn.resizeMsgChan <- initTerminalSize
	} else {
		// 前端没有指定长宽高, 使用默认值
		conn.resizeMsgChan <- &TerminalSize{
			Rows: DefaultRows,
			Cols: DefaultCols,
		}
	}

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
		resizeMsg, resizeErr := r.bindMgr.HandleResizeMsg(decodeMsg)
		if resizeErr != nil {
			return nil, nil
		}

		r.resizeMsgChan <- resizeMsg
		return nil, nil
	}

	// 打印日志
	if channel == LogChannel {
		inputMsg, _ := r.bindMgr.HandleInputMsg(decodeMsg)
		logger.Infof("UserName=%s  SessionID=%s  Command=%s",
			r.bindMgr.PodCtx.Username, r.bindMgr.PodCtx.SessionId, string(inputMsg))
		return nil, nil
	}

	inputMsg, err := r.bindMgr.HandleInputMsg(decodeMsg)
	if err != nil {
		return nil, nil
	}

	_, ss, _ := ansi.Decode(inputMsg)
	cmdParser.Cmd = ss
	cmdParser.InputSlice = append(cmdParser.InputSlice, ss)
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
	//输入输出映射,用于查找历史命令
	out, ss, e := ansi.Decode(outputMsg)
	if e != nil {
		logger.Error("decode output error:", e)
	}
	//TODO:历史命令问题,可能解析问题导致
	if strings.ReplaceAll(string(ss.Code), "\b", "") == "" {
		rex := regexp.MustCompile("\\x1b\\[\\d+P")
		l := rex.Split(string(out), -1)
		ss.Code = ansi.Name(l[len(l)-1])
	}
	//时序性问题不可避免
	cmdParser.CmdResult[cmdParser.Cmd] = ss

	if cmdParser.Cmd != nil && cmdParser.Cmd.Code == "\r" {
		cmd := audit.ResolveInOut(cmdParser)
		if cmd != "" {
			logger.Infof("UserName=%s  SessionID=%s  Command=%s",
				r.bindMgr.PodCtx.Username, r.bindMgr.PodCtx.SessionId, cmd)
		}
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

	guideMessages := helloMessage(c, r.bindMgr.PodCtx.Source)
	notSendMsg := true

	for {
		select {
		case <-r.ctx.Done():
			logger.Infof("close %s RemoteStreamConn done", r.bindMgr.PodCtx.PodName)
			return nil
		case output, ok := <-r.outputMsgChan:
			if !ok {
				logger.Infof("close %s RemoteStreamConn done by chan", r.bindMgr.PodCtx.PodName)
				return nil
			}
			// 收到首个字节才发送 hello 信息
			if notSendMsg && !r.hideBanner {
				PreparedGuideMessage(r.ctx, r.wsConn, guideMessages)
				notSendMsg = false
			}

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
		guideMsg = []string{i18n.GetMessage(c, "mgrGuideMessage")}
		guideMsg = append(guideMsg, config.G.WebConsole.GuideDocLinks...)
	} else {
		guideMsg = []string{i18n.GetMessage(c, "guideMessage")}
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
