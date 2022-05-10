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
	"fmt"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/podmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var upgrader = websocket.Upgrader{
	EnableCompression: true,
	CheckOrigin:       func(r *http.Request) bool { return true },
}

// BCSWebSocketHandler WebSocket 连接处理函数
func (s *service) BCSWebSocketHandler(c *gin.Context) {
	// 还未建立 WebSocket 连接, 使用 Json 返回
	errResp := types.APIResponse{
		Code: 400,
		Data: map[string]string{},
	}

	if !websocket.IsWebSocketUpgrade(c.Request) {
		errResp.Message = "invalid websocket connection"
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp)
		return
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		errResp.Message = fmt.Sprintf("upgrade websocket connection error, %s", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp)
		return
	}
	defer ws.Close()

	// 监听 Ctrl-C 信号
	ctx, stop := signal.NotifyContext(c.Request.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	eg, ctx := errgroup.WithContext(ctx)

	// 已经建立 WebSocket 连接, 下面所有的错误返回, 需要使用 GracefulCloseWebSocket
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")
	sessionId := c.Query("session_id")

	rows, _ := strconv.Atoi(c.Query("rows"))
	cols, _ := strconv.Atoi(c.Query("cols"))

	initTerminalSize := &manager.TerminalSize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	}

	connected := false
	store := sessions.NewRedisStore(projectId, clusterId)
	podCtx, err := store.Get(c.Request.Context(), sessionId)
	if err != nil {
		manager.GracefulCloseWebSocket(ctx, ws, connected, errors.Wrap(err, "获取session失败"))
		return
	}

	consoleMgr := manager.NewConsoleManager(ctx, podCtx)
	remoteStreamConn := manager.NewRemoteStreamConn(ctx, ws, consoleMgr, initTerminalSize)
	connected = true

	// kubectl 容器， 需要定时上报心跳
	if podCtx.Mode == types.ClusterExternalMode || podCtx.Mode == types.ClusterInternalMode {
		podCleanUpMgr := podmanager.NewCleanUpManager(ctx)
		consoleMgr.AddMgrFunc(podCleanUpMgr.Heartbeat)
	}

	eg.Go(func() error {
		defer stop()

		// 定时检查任务等
		return consoleMgr.Run()
	})

	eg.Go(func() error {
		defer stop()

		// 定时发送心跳等, 保持连接的活跃
		return remoteStreamConn.Run()
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
	manager.GracefulCloseWebSocket(ctx, ws, connected, errors.New("BCS Console 服务端连接断开，请重新登录"))
}
