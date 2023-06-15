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

package api

import (
	"context"
	"net/http"
	"time"

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

// wsQuery websocket 支持的参数
type wsQuery struct {
	Rows       uint16 `form:"rows"`
	Cols       uint16 `form:"cols"`
	HideBanner bool   `form:"hide_banner"`
	Lang       string `form:"lang"` // banner 国际化, 在中间件已经处理，这里只做记录
}

// GetTerminalSize 获取初始宽高
func (q *wsQuery) GetTerminalSize() *manager.TerminalSize {
	if q.Rows > 0 && q.Cols > 0 {
		return &manager.TerminalSize{
			Rows: q.Rows,
			Cols: q.Cols,
		}
	}
	return nil
}

// BCSWebSocketHandler WebSocket 连接处理函数
func (s *service) BCSWebSocketHandler(c *gin.Context) {
	// 还未建立 WebSocket 连接, 使用 Json 返回
	if !websocket.IsWebSocketUpgrade(c.Request) {
		rest.AbortWithBadRequestError(c, errors.New("invalid websocket connection"))
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		rest.AbortWithBadRequestError(c, errors.Wrap(err, "upgrade websocket connection"))
		return
	}
	defer ws.Close()

	// 已经建立 WebSocket 连接, 下面所有的错误返回, 需要使用 GracefulCloseWebSocket 返回
	ctx, stop := context.WithCancel(c.Request.Context())
	defer stop()

	connected := false

	query := &wsQuery{}
	if e := c.BindQuery(query); e != nil {
		manager.GracefulCloseWebSocket(ctx, ws, connected, errors.Wrap(e, i18n.GetMessage(c, "参数不合法")))
		return
	}

	sessionId := route.GetSessionId(c)
	podCtx, err := sessions.NewStore().WebSocketScope().Get(ctx, sessionId)
	if err != nil {
		manager.GracefulCloseWebSocket(ctx, ws, connected, errors.Wrap(err, i18n.GetMessage(c, "session不合法")))
		return
	}
	// 赋值session id
	podCtx.SessionId = sessionId
	consoleMgr := manager.NewConsoleManager(ctx, podCtx)
	remoteStreamConn := manager.NewRemoteStreamConn(ctx, ws, consoleMgr, query.GetTerminalSize(), query.HideBanner)
	connected = true

	// kubectl 容器， 需要定时上报心跳
	if podCtx.Mode == types.ClusterExternalMode || podCtx.Mode == types.ClusterInternalMode {
		podCleanUpMgr := podmanager.NewCleanUpManager(ctx)
		consoleMgr.AddMgrFunc(podCleanUpMgr.Heartbeat)
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer stop()

		// 定时检查任务等
		return consoleMgr.Run(c)
	})

	eg.Go(func() error {
		defer stop()

		// 定时发送心跳等, 保持连接的活跃
		return remoteStreamConn.Run(c)
	})

	eg.Go(func() error {
		defer stop()

		// 关闭需要主动发送 Ctrl-D 命令
		return remoteStreamConn.WaitStreamDone(podCtx)
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
	manager.GracefulCloseWebSocket(ctx, ws, connected, errors.New(i18n.GetMessage(c, "BCS Console 服务端连接断开，请重新登录")))
}
