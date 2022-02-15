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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/i18n"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// ConsoleCopywritingFailed is a response string
var ConsoleCopywritingFailed = []string{
	"###########################################################################################\r\n",
	"#                                 Welcome To BCS Console                                  #\r\n",
	"###########################################################################################\r\n",
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

	// InputLineBreaker 输入分行标识
	InputLineBreaker = "\r"
	// OutputLineBreaker 输出分行标识
	OutputLineBreaker = "\r\n"

	// AnsiEscape bash 颜色标识
	AnsiEscape = "r\"\\x1B\\[[0-?]*[ -/]*[@-~]\""

	queueName = "bcs_web_console_record"
	tags      = "bcs-web-console"

	StdinChannel  = "0"
	StdoutChannel = "1"
	StderrChannel = "2"
	ErrorChannel  = "3"
	ResizeChannel = "4"

	// 审计上报、ws连接监测时间间隔
	recordInterval = 10
)

type errMsg struct {
	Msg string `json:"msg,omitempty"`
}

// WsMessage websocket消息
type WsMessage struct {
	MessageType int
	Data        types.XtermMessage
}

// ssh流式处理器
type streamHandler struct {
	wsConn      *wsConn
	resizeEvent chan remotecommand.TerminalSize
}

type wsConn struct {
	conn          *websocket.Conn
	inChan        chan *WsMessage // 读取队列
	outChan       chan *WsMessage // 发送队列
	mutex         sync.Mutex      // 避免重复关闭管道
	isClosed      bool
	closeChan     chan struct{} // 关闭通知
	ConnTime      time.Time     // 连接时间
	LastInputTime time.Time
	PodName       string //
	ConfigMapName string
	Username      string //
	SessionID     string
	Project       string
	Cluster       string
	InputRecord   string // 输入
	OutputRecord  string // 输出
	Context       interface{}
}

func genWsConn(conn *websocket.Conn, conf types.WebSocketConfig) *wsConn {
	configMapName := getConfigMapName(conf.ClusterID, conf.ProjectsID)
	podName := getPodName(conf.ClusterID, conf.ProjectsID)
	return &wsConn{
		conn:          conn,
		inChan:        make(chan *WsMessage, 1000),
		outChan:       make(chan *WsMessage, 1000),
		isClosed:      false,
		closeChan:     make(chan struct{}),
		ConnTime:      time.Now(),
		LastInputTime: time.Now(),
		PodName:       podName,
		ConfigMapName: configMapName,
		Username:      conf.User,
		SessionID:     conf.SessionID,
		Project:       conf.ProjectsID,
		Cluster:       conf.ClusterID,
	}
}

// 读取协程
func (c *wsConn) wsReadLoop() {
	defer c.wsClose()
	for {
		// 读一条message
		msgType, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		// 解析base64数据
		dataDec, err := base64.StdEncoding.DecodeString(string(data[1:]))
		if err != nil {
			continue
		}

		// 解析数据
		wsMessage := WsMessage{
			MessageType: msgType,
		}
		xtermMsg := types.XtermMessage{}
		if string(data[0]) == ResizeChannel {
			wsMessage.Data.MsgType = "resize"
			err = json.Unmarshal(dataDec, &xtermMsg)
			if err != nil {
				continue
			}
			wsMessage.Data.Rows = xtermMsg.Rows
			wsMessage.Data.Cols = xtermMsg.Cols
		} else {
			wsMessage.Data.MsgType = "input"
			wsMessage.Data.Input = string(dataDec)
			// 把输入存起来
			c.InputRecord += string(dataDec)
		}

		// 更新ws时间
		c.LastInputTime = time.Now()

		// 放入请求队列
		c.inChan <- &wsMessage
	}
}

// 发送协程
func (c *wsConn) wsWriteLoop() {
	// 服务端返回给页面的数据
	for {
		select {
		// 取一个应答
		case msg := <-c.outChan:
			// 写给web websocket
			output := base64.StdEncoding.EncodeToString([]byte(msg.Data.Output))
			if err := c.conn.WriteMessage(msg.MessageType, []byte(output)); err != nil {
				break
			}
			c.OutputRecord += msg.Data.Output
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
	case c.outChan <- &WsMessage{
		MessageType: messageType,
		Data: types.XtermMessage{
			Output: string(data),
		},
	}:

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

func (c *wsConn) periodicTick(period time.Duration) {

	go wait.NonSlidingUntil(c.tickTimeout, period*time.Second, c.closeChan)
}

// 主动停止掉session
func (c *wsConn) tickTimeout() {
	nowTime := time.Now()
	idleTime := nowTime.Sub(c.LastInputTime).Seconds()
	if idleTime > TickTimeout {
		msg := i18n.GetMessage("BCS Console 使用已经超过{}小时，请重新登录",
			map[string]string{"time": strconv.Itoa(TickTimeout / 60)})
		blog.Info("tick timeout, close session %s, idle time, %.2f", c.PodName, idleTime)
		c.inChan <- &WsMessage{
			MessageType: websocket.TextMessage,
			Data:        types.XtermMessage{Output: msg},
		}
		c.wsClose()
		return
	}
	loginTime := nowTime.Sub(c.ConnTime).Seconds()
	if loginTime > LoginTimeout {
		msg := i18n.GetMessage("BCS Console 使用已经超过{}小时，请重新登录",
			map[string]string{"time": strconv.Itoa(LoginTimeout / 60)})
		blog.Info("tick timeout, close session %s, login time, %.2f", c.PodName, loginTime)
		c.wsClose()
		c.inChan <- &WsMessage{
			MessageType: websocket.TextMessage,
			Data:        types.XtermMessage{Output: msg},
		}
		c.wsClose()
		return
	}

}

// ResponseJSON response to client
// Deprecated : 这个方法将被废弃，改用c.Json()
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
	if msg.Data.MsgType == "reset" {
		// 放到channel里，等remotecommand executor调用Next取走
		handler.resizeEvent <- remotecommand.TerminalSize{Width: msg.Data.Cols, Height: msg.Data.Rows}
	}

	size = len(msg.Data.Input)
	copy(p, msg.Data.Input)
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
func (m *manager) StartExec(c *gin.Context, conf *types.WebSocketConfig) {
	blog.Debug(fmt.Sprintf("start exec for container pod %s", conf.PodName))

	upgrader := websocket.Upgrader{
		EnableCompression: true,
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	if !websocket.IsWebSocketUpgrade(c.Request) {
		msg := i18n.GetMessage("连接已经断开")
		utils.APIError(c, msg)
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		msg := i18n.GetMessage("连接已经断开")
		utils.APIError(c, msg)
		return
	}

	wsConn := genWsConn(ws, *conf)

	defer wsConn.wsClose()

	// 页面读入输入 协程
	go wsConn.wsReadLoop()
	// 服务端返回数据 协程
	go wsConn.wsWriteLoop()

	// 记录pod心跳
	go m.heartbeat(time.Duration(1), wsConn.closeChan, conf.PodName)
	// 获取输入输出数据，定期上报
	go m.startRecord(time.Duration(recordInterval), wsConn.closeChan, wsConn)
	// ws 超时监测
	go wsConn.periodicTick(time.Duration(recordInterval))

	for _, i := range ConsoleCopywritingFailed {
		err := ws.WriteMessage(websocket.TextMessage, []byte(base64.StdEncoding.EncodeToString([]byte(i))))
		if err != nil {
			msg := i18n.GetMessage("连接已经断开")
			utils.APIError(c, msg)
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
		msg := i18n.GetMessage("连接已经断开")
		utils.APIError(c, msg)
		return
	}

	msg := i18n.GetMessage("连接已经断开")
	utils.APIError(c, msg)
}

// 记录pod心跳
// 定时上报存活, 清理时需要使用
func (m *manager) heartbeat(period time.Duration, stopCh <-chan struct{}, podName string) {

	go wait.NonSlidingUntil(func() {
		timeNow := time.Unix(time.Now().Unix(), 0).Format("20060102150405")
		timeNowFloat, _ := strconv.ParseFloat(timeNow, 64)
		m.redisClient.ZAdd(context.Background(), WebConsoleHeartbeatKey, &redis.Z{Member: podName, Score: timeNowFloat})
	}, period*time.Second, stopCh)

}

// 提交数据
func (m *manager) emit(data types.AuditRecord) {
	dataByte, _ := json.Marshal(data)
	m.redisClient.RPush(context.Background(), queueName, dataByte)
}

// 审计
func (m *manager) startRecord(period time.Duration, stopCh <-chan struct{}, wsObj *wsConn) {

	go wait.NonSlidingUntil(func() {
		if data := wsObj.periodicRecord(); data != nil {
			m.emit(*data)
		}
	}, period*time.Second, stopCh)
}

// CleanUserPod 单个集群清理
func (m *manager) CleanUserPod() {

	// TODO 根据不同的集群进行删除

	alivePods := m.getActiveUserPod()
	alivePodsMap := make(map[string]string)
	for _, pod := range alivePods {
		alivePodsMap[pod] = pod
	}

	podList, err := m.k8sClient.CoreV1().Pods(Namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return
	}

	m.cleanUserPodByCluster(podList, alivePodsMap)

}

// 清理用户下的相关集群pod
func (m *manager) cleanUserPodByCluster(podList *v1.PodList, alivePods map[string]string) {

	// 过期时间
	timeDiff, _ := time.ParseDuration("-" + strconv.FormatInt(UserPodExpireTime, 10) + "s")
	minExpireTime := time.Now().Add(timeDiff) // 在此时间之前的都算作过期

	for _, pod := range podList.Items {
		if pod.Status.Phase == "Pending" {
			continue
		}

		// 小于一个周期的pod不清理
		podCreateTimeStr, _ := pod.ObjectMeta.Labels[LabelWebConsoleCreateTimestamp]
		podCreateTime, _ := time.Parse("20060102150405", podCreateTimeStr)
		if minExpireTime.After(podCreateTime) {
			blog.Info("pod %s exist time %s > %s, just ignore", pod.Name, podCreateTimeStr, minExpireTime)
			continue
		}

		// 有心跳上报的不清理
		if _, ok := alivePods[pod.Name]; ok {
			continue
		}

		// 删除pod
		err := m.k8sClient.CoreV1().Pods(Namespace).Delete(context.Background(), pod.Name, metav1.DeleteOptions{})
		if err != nil {
			blog.Errorf("delete pod(%s) failed, err: %v", pod.Name, err)
			continue
		}
		blog.Info("delete pod %s", pod.Name)

		// 删除configMap
		for _, volume := range pod.Spec.Volumes {
			if volume.ConfigMap != nil {
				if volume.ConfigMap != nil {
					err = m.k8sClient.CoreV1().ConfigMaps(Namespace).Delete(context.Background(),
						volume.ConfigMap.LocalObjectReference.Name, metav1.DeleteOptions{})
					if err != nil {
						blog.Errorf("delete configmap %s failed ,err : %v", volume.ConfigMap.LocalObjectReference.Name,
							err)
					}
					blog.Info("delete configmap %s", volume.ConfigMap.LocalObjectReference.Name)
				}

			}
		}

	}

}

// 获取存活节点
func (m *manager) getActiveUserPod() []string {

	now := time.Now()
	timeDiff, _ := time.ParseDuration("-" + strconv.FormatInt(UserPodExpireTime, 10) + "s")
	start := now.Add(timeDiff)
	startTime := start.Format("20060102150405")
	// 删除掉过期数据
	m.redisClient.ZRemRangeByScore(context.Background(), WebConsoleHeartbeatKey, "-inf", startTime)

	// 获取存活的pod
	activatedPods := m.redisClient.ZRange(context.Background(), WebConsoleHeartbeatKey, 0, -1).Val()

	return activatedPods
}

// 周期上报操作记录
func (c *wsConn) periodicRecord() *types.AuditRecord {

	inputRecord := c.flushInputRecord()
	outputRecord := c.flushOutputRecord()

	// 如果输入输出为空则取消此次上报
	if len(inputRecord) == 0 && len(outputRecord) == 0 {
		return nil
	}

	data := types.AuditRecord{
		InputRecord:  inputRecord,
		OutputRecord: outputRecord,
		SessionID:    c.SessionID,
		Context:      nil,
		ProjectID:    c.Project,
		ClusterID:    c.Cluster,
		UserPodName:  c.PodName,
		Username:     c.Username,
	}

	return &data

}

// 获取输入记录
func (c *wsConn) flushInputRecord() string {

	if c.InputRecord == "" {
		return ""
	}

	lineMsg := strings.Split(c.InputRecord, InputLineBreaker)
	var record, cmd string
	for _, s := range lineMsg {
		cmd = cleanBashEscape(s)
		if cmd == "" {
			continue
		}
		record += "\r\n" + fmt.Sprintf("%s: %s",
			time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05.06"), cmd)
		blog.Debug(record)
	}

	c.InputRecord = ""

	return record
}

// 去除bash转义字符
func cleanBashEscape(text string) string {
	// 删除转移字符
	re, err := regexp.Compile(AnsiEscape)
	if err != nil {
		return ""
	}
	text = re.ReplaceAllString(text, "")

	return text
}

// 获取输出记录
func (c *wsConn) flushOutputRecord() string {
	if c.OutputRecord == "" {
		return ""
	}

	lineMsg := strings.Split(c.OutputRecord, OutputLineBreaker)
	var record, cmd string
	for _, s := range lineMsg {
		cmd = cleanBashEscape(s)
		if cmd == "" {
			continue
		}
		record += "\r\n" + fmt.Sprintf("%s: %s",
			time.Unix(time.Now().Unix(), 0).Format("2006-01-02 15:04:05.06"), cmd)
		blog.Debug(record)
	}
	c.OutputRecord = ""

	return record
}

func (m *manager) startExec(ws *wsConn, conf *types.WebSocketConfig) error {

	req := m.k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(conf.PodName).
		Namespace(Namespace).
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
