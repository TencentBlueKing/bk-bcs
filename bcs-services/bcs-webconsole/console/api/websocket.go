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

package api

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/podmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
)

var upgrader = websocket.Upgrader{
	EnableCompression: true,
	CheckOrigin:       func(r *http.Request) bool { return true },
}

var (
	stopWg    = &sync.WaitGroup{}
	stopCount = atomic.Int64{}
)

// wsQuery websocket 支持的参数
type wsQuery struct {
	Rows       uint16 `form:"rows"`
	Cols       uint16 `form:"cols"`
	HideBanner bool   `form:"hide_banner"`
	Lang       string `form:"lang"` // banner 国际化, 在中间件已经处理，这里只做记录
}

// GetTerminalSize 获取初始宽高
func (q *wsQuery) GetTerminalSize() *types.TerminalSize {
	if q.Rows > 0 && q.Cols > 0 {
		return &types.TerminalSize{
			Rows: q.Rows,
			Cols: q.Cols,
		}
	}

	return types.DefaultTerminalSize()
}

// BCSWebSocketHandler WebSocket 连接处理函数
// NOCC:golint/fnsize(设计如此:)
func (s *service) BCSWebSocketHandler(c *gin.Context) { // nolint
	// 还未建立 WebSocket 连接, 使用 Json 返回
	if !websocket.IsWebSocketUpgrade(c.Request) {
		rest.APIError(c, "invalid websocket connection")
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		rest.APIError(c, errors.Wrap(err, "upgrade websocket connection").Error())
		return
	}
	defer ws.Close() // nolint

	// 已经建立 WebSocket 连接, 下面所有的错误返回, 需要使用 GracefulCloseWebSocket 返回
	eg, ctx := errgroup.WithContext(c.Request.Context())
	connected := false

	query := &wsQuery{}
	if e := c.BindQuery(query); e != nil {
		manager.GracefulCloseWebSocket(ctx, ws, connected, errors.Wrap(e, i18n.T(c, "参数不合法")))
		return
	}

	sessionId := route.GetSessionId(c)
	podCtx, err := sessions.NewStore().WebSocketScope().Get(ctx, sessionId)
	if err != nil {
		manager.GracefulCloseWebSocket(ctx, ws, connected, errors.Wrap(err, i18n.T(c, "session不合法")))
		return
	}
	// 赋值session id
	podCtx.SessionId = sessionId

	terminalSize := query.GetTerminalSize()
	consoleMgr, err := manager.NewConsoleManager(ctx, podCtx, terminalSize)
	if err != nil {
		manager.GracefulCloseWebSocket(ctx, ws, connected, errors.Wrap(err, i18n.T(c, "初始化session失败")))
		return
	}

	remoteStreamConn := manager.NewRemoteStreamConn(ctx, ws, consoleMgr, terminalSize, query.HideBanner)
	connected = true

	// kubectl 容器， 需要定时上报心跳
	if podCtx.Mode == types.ClusterExternalMode || podCtx.Mode == types.ClusterInternalMode {
		podCleanUpMgr := podmanager.NewCleanUpManager(ctx)
		consoleMgr.AddMgrFunc(podCleanUpMgr.Heartbeat)
	}

	eg.Go(func() error {
		// 定时检查任务
		// 命令行审计
		// terminal recorder
		return consoleMgr.Run(c)
	})

	eg.Go(func() error {
		// 定时发送心跳等, 保持连接的活跃
		return remoteStreamConn.Run(c)
	})

	eg.Go(func() error {
		// 关闭需要主动发送 Ctrl-D 命令
		return remoteStreamConn.WaitStreamDone(c, podCtx)
	})

	stopWg.Add(1)
	stopCount.Add(1)
	defer stopWg.Done()
	defer stopCount.Add(-1)

	eg.Go(func() error {
		select {
		case <-ctx.Done():
			return nil
		case <-s.opts.StopSignCtx.Done():
			return errors.New(i18n.T(c, "BCS Console 服务端连接断开，请重新登录"))
		}
	})

	// 封装一个独立函数, 统计耗时
	if err := func() error {
		start := time.Now()

		// 单独统计 ws metrics
		metrics.CollectWsConnection(podCtx.Username, podCtx.ClusterId, podCtx.Namespace, podCtx.PodName, podCtx.ContainerName)
		metrics.CollectWsConnectionOnline(1)

		defer func() {
			// 过滤掉 ws 长链接时间
			wsConnDuration := time.Since(start)
			metrics.SetRequestIgnoreDuration(c, wsConnDuration)

			metrics.CollectWsConnectionOnline(-1)
		}()

		return eg.Wait()
	}(); err != nil {
		manager.GracefulCloseWebSocket(ctx, ws, connected, err)
		return
	}

	// 正常退出, 如使用 Exit 命令主动退出返回提示
	manager.GracefulCloseWebSocket(ctx, ws, connected, errors.New(i18n.T(c, "BCS Console 服务端连接断开，请重新登录")))
}

// WaitWebsocketClose wait all conn close
func WaitWebsocketClose(timeout time.Duration) {
	st := time.Now()
	count := stopCount.Load()

	c := make(chan struct{})
	go func() {
		defer close(c)
		stopWg.Wait()
	}()

	select {
	case <-c:
		logger.Infof("all websocket connections closed, count=%d, duration=%s", count, time.Since(st))
		return // completed normally
	case <-time.After(timeout):
		logger.Warnf("websocket connections close timeout, just ignore, count=%d, duration=%s", count, time.Since(st))
		return // timed out
	}
}
