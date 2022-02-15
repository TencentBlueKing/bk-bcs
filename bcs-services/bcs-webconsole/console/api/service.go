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
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
	"github.com/google/uuid"
	"go-micro.dev/v4/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

var upgrader = websocket.Upgrader{
	EnableCompression: true,
	CheckOrigin:       func(r *http.Request) bool { return true },
}

type service struct {
	opts *route.Options
}

func NewRouteRegistrar(opts *route.Options) route.Registrar {
	return service{opts: opts}
}

// 	router.Use(route.Localize())
func (s service) RegisterRoute(router gin.IRoutes) {
	router.Use(route.AuthRequired()).
		GET("/api/projects/:projectId/clusters/:clusterId/session/", s.CreateWebConsoleSession).
		GET("/ws/projects/:projectId/clusters/:clusterId/", s.BCSWebSocketHandler).
		GET("/api/projects/:projectId/clusters/", s.ListClusters).
		POST("/web_console", s.CreateOpenWebConsoleSession).
		GET(filepath.Join(s.opts.RoutePrefix, "/api/projects/:projectId/clusters/:clusterId/session")+"/", s.CreateWebConsoleSession).
		GET(filepath.Join(s.opts.RoutePrefix, "/ws/projects/:projectId/clusters/:clusterId")+"/", s.BCSWebSocketHandler).
		GET(filepath.Join(s.opts.RoutePrefix, "/api/projects/:projectId/clusters/"), s.ListClusters).
		POST(filepath.Join(s.opts.RoutePrefix, "/web_console/"), s.CreateOpenWebConsoleSession)
}

func (s *service) ListClusters(c *gin.Context) {
	projectId := c.Param("projectId")
	clusters, err := bcs.ListClusters(c.Request.Context(), projectId)
	if err != nil {
		utils.APIError(c, i18n.GetMessage(err.Error()))
		return
	}
	data := types.APIResponse{
		Data:      clusters,
		Code:      types.NoError,
		Message:   i18n.GetMessage("获取集群成功"),
		RequestID: uuid.New().String(),
	}
	c.JSON(http.StatusOK, data)
}

func (s *service) CreateWebConsoleSession(c *gin.Context) {
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")
	containerId := c.Query("container_id")
	username := ""

	if config.G.Base.Env == config.DevEnv {
		username = c.Query("username")
		if username == "" {
			utils.APIError(c, i18n.GetMessage("username 不能为空"))
			return
		}
	} else {
		utils.APIError(c, i18n.GetMessage("prod username 不能为空"))
		return
	}

	startupMgr, err := manager.NewPodStartupManager(c.Request.Context(), clusterId)
	if err != nil {
		msg := i18n.GetMessage("k8s客户端初始化失败{}", err)
		utils.APIError(c, msg)
		return
	}

	podCtx := &types.PodContext{
		ProjectId: projectId,
		Username:  username,
		ClusterId: clusterId,
		Mode:      config.G.WebConsole.Mode,
	}

	if containerId != "" {
		resp, err := startupMgr.GetK8sContextByContainerID(containerId)
		if err != nil {
			msg := i18n.GetMessage("container_id不正确，请检查参数", err)
			utils.APIError(c, msg)
			return
		}
		podCtx.Namespace = resp.Namespace
		podCtx.PodName = resp.PodName
		podCtx.ContainerName = resp.ContainerName
		podCtx.Commands = manager.DefaultCommand
	} else {
		namespace := manager.GetNamespace()
		podName, err := startupMgr.WaitPodUp(namespace, username)
		if err != nil {
			msg := i18n.GetMessage("申请pod资源失败{}", err)
			utils.APIError(c, msg)
			return
		}
		podCtx.Namespace = namespace
		podCtx.PodName = podName
		podCtx.ContainerName = manager.KubectlContainerName
		// 进入 kubectld pod， 固定使用bash
		podCtx.Commands = []string{"/bin/bash"}
	}

	store := sessions.NewRedisStore(projectId, clusterId)
	sessionId, err := store.Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", err)
		utils.APIError(c, msg)
		return
	}

	wsUrl := filepath.Join(s.opts.RoutePrefix, fmt.Sprintf("/ws/projects/%s/clusters/%s/?session_id=%s",
		projectId, clusterId, sessionId))

	data := types.APIResponse{
		Data: map[string]string{
			"session_id": sessionId,
			"ws_url":     wsUrl,
		},
		Code:      types.NoError,
		Message:   i18n.GetMessage("获取session成功"),
		RequestID: uuid.New().String(),
	}
	c.JSON(http.StatusOK, data)
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

	eg.Go(func() error {
		// 定时检查任务等
		return consoleMgr.Run()
	})

	eg.Go(func() error {
		// 定时发送心跳等, 保持连接的活跃
		return remoteStreamConn.Run()
	})

	eg.Go(func() error {
		defer remoteStreamConn.Close()
		defer logger.Info("Close WaitSteamDone done")

		// 远端错误, 一般是远端 Pod 被关闭或者使用 Exit 命令主动退出
		// 关闭需要主动发送 Ctrl-D 命令
		return remoteStreamConn.WaitSteamDone(podCtx)
	})

	if err := eg.Wait(); err != nil {
		manager.GracefulCloseWebSocket(ctx, ws, connected, errors.Wrap(err, "Handle websocket"))
		return
	}

	manager.GracefulCloseWebSocket(ctx, ws, connected, nil)
}

func (s *service) CreateOpenWebConsoleSession(c *gin.Context) {

	projectId := c.Query("project_id")
	clusterId := c.Query("cluster_id")

	var containerName string

	startupMgr, err := manager.NewPodStartupManager(c.Request.Context(), clusterId)
	if err != nil {
		msg := i18n.GetMessage("k8s客户端初始化失败{}", map[string]string{"err": err.Error()})
		utils.APIError(c, msg)
		return
	}

	// 优先使用containerID
	containerID, ok := c.GetPostForm("container_id")
	if ok {

		container, err := startupMgr.GetK8sContextByContainerID(containerID)
		if err != nil {
			blog.Info("container_id is incorrect, err : %v", err)
			msg := i18n.GetMessage("container_id不正确，请检查参数")
			utils.APIError(c, msg)
			return
		}

		containerName = container.ContainerName

	} else {

		podName, _ := c.GetPostForm("pod_name")
		containerName, _ := c.GetPostForm("container_name")
		namespace, _ := c.GetPostForm("namespace")

		// 其他使用namespace, pod, container
		if namespace == "" || podName == "" || containerName == "" {
			msg := i18n.GetMessage("container_id或namespace/pod_name/container_name不能同时为空")
			utils.APIError(c, msg)
			return
		}
	}

	store := sessions.NewRedisStore(projectId, clusterId)

	podCtx := &types.PodContext{
		ProjectId: projectId,
		Username:  "",
		ClusterId: clusterId,
		Namespace: "",
		PodName:   "",
		Mode:      config.G.WebConsole.Mode,
	}

	sessionId, err := store.Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", err)
		utils.APIError(c, msg)
		return
	}

	wsUrl := filepath.Join(s.opts.RoutePrefix, fmt.Sprintf("/web_console/?session_id=%s&container_name=%s",
		sessionId, containerName))

	respData := types.APIResponse{
		Data: map[string]string{
			"session_id": sessionId,
			"ws_url":     wsUrl,
		},
		Code:      types.NoError,
		Message:   i18n.GetMessage("获取session成功"),
		RequestID: uuid.New().String(),
	}

	c.JSON(http.StatusOK, respData)
}
