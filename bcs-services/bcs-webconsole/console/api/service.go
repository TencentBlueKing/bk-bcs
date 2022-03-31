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
	"net/url"
	"os/signal"
	"path"
	"strconv"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/podmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

func (s service) RegisterRoute(router gin.IRoutes) {
	api := router.Use(route.APIAuthRequired())

	permAPI := api.Use(route.PermissionRequired())

	// 用户登入态鉴权, session鉴权
	permAPI.GET("/api/projects/:projectId/clusters/:clusterId/session/", s.CreateWebConsoleSession)
	permAPI.GET("/api/projects/:projectId/clusters/", s.ListClusters)

	api.GET("/api/open_session/", s.CreateOpenSession)

	// 蓝鲸API网关鉴权 & App鉴权
	api.POST("/api/projects/:projectId/clusters/:clusterId/open_session/", s.CreateOpenWebConsoleSession)

	// websocket协议, session鉴权
	api.GET("/ws/projects/:projectId/clusters/:clusterId/", s.BCSWebSocketHandler)
}

func (s *service) ListClusters(c *gin.Context) {
	projectId := c.Param("projectId")
	clusters, err := bcs.ListClusters(c.Request.Context(), config.G.BCS, projectId)
	if err != nil {
		APIError(c, i18n.GetMessage(err.Error()))
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
	source := c.Query("source")
	lang := c.Query("lang")

	authCtx, err := route.GetAuthContext(c)
	if err != nil {
		APIError(c, i18n.GetMessage(err.Error()))
		return
	}

	var mode types.WebConsoleMode
	adminClusterId := clusterId
	if containerId != "" {
		mode = types.ContainerDirectMode
	} else {
		adminClusterId, mode = podmanager.TranslateConsoleMode(clusterId, config.G.WebConsole.Mode)
	}

	startupMgr, err := podmanager.NewStartupManager(c.Request.Context(), mode, adminClusterId)
	if err != nil {
		msg := i18n.GetMessage("k8s客户端初始化失败{}", err)
		APIError(c, msg)
		return
	}

	podCtx := &types.PodContext{
		ProjectId:      projectId,
		Username:       authCtx.Username,
		AdminClusterId: adminClusterId,
		ClusterId:      clusterId,
		Source:         source,
		Mode:           mode,
	}

	if containerId != "" {
		resp, err := startupMgr.GetK8sContextByContainerID(containerId)
		if err != nil {
			msg := i18n.GetMessage("container_id不正确，请检查参数", err)
			APIError(c, msg)
			return
		}
		podCtx.Namespace = resp.Namespace
		podCtx.PodName = resp.PodName
		podCtx.ContainerName = resp.ContainerName
		podCtx.Commands = manager.DefaultCommand
	} else {
		namespace := podmanager.GetNamespace()
		imageTag, err := podmanager.GetKubectldVersion(clusterId)
		if err != nil {
			msg := i18n.GetMessage("k8s集群版本获取失败{}", err)
			APIError(c, msg)
			return
		}

		image := config.G.WebConsole.KubectldImage + ":" + imageTag

		podName, err := startupMgr.WaitPodUp(clusterId, namespace, image, authCtx.Username)
		if err != nil {
			msg := i18n.GetMessage("申请pod资源失败{err}", err)
			APIError(c, msg)
			return
		}
		podCtx.Namespace = namespace
		podCtx.PodName = podName
		podCtx.ContainerName = podmanager.KubectlContainerName
		// 进入 kubectld pod， 固定使用bash
		podCtx.Commands = []string{"/bin/bash"}
	}

	store := sessions.NewRedisStore(projectId, clusterId)
	sessionId, err := store.Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", err)
		APIError(c, msg)
		return
	}

	query := url.Values{}
	query.Set("session_id", sessionId)

	if lang != "" {
		query.Set("lang", lang)
	}

	wsUrl := path.Join(s.opts.RoutePrefix, fmt.Sprintf("/ws/projects/%s/clusters/%s/?%s",
		projectId, clusterId, query.Encode()))

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

func (s *service) CreateOpenSession(c *gin.Context) {
	sessionId := c.Query("session_id")

	store := sessions.NewRedisStore("open-session", "open-session")
	podCtx, err := store.Get(c.Request.Context(), sessionId)
	if err != nil {
		msg := i18n.GetMessage("sessin_id不正确", err)
		APIError(c, msg)
		return
	}

	newStore := sessions.NewRedisStore(podCtx.ProjectId, podCtx.ClusterId)
	NewSessionId, err := newStore.Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", err)
		APIError(c, msg)
		return
	}

	wsUrl := path.Join(s.opts.RoutePrefix, fmt.Sprintf("/ws/projects/%s/clusters/%s/?session_id=%s",
		podCtx.ProjectId, podCtx.ClusterId, NewSessionId))

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

	// kubectl 容器， 需要定时上报心跳
	if podCtx.Mode == types.ClusterExternalMode || podCtx.Mode == types.ClusterInternalMode {
		podCleanUpMgr := podmanager.NewCleanUpManager(ctx)
		consoleMgr.AddMgrFunc(podCleanUpMgr.Heartbeat)
	}

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
		defer logger.Infof("Close %s WaitStreamDone done", podCtx.PodName)

		// 远端错误, 一般是远端 Pod 被关闭或者使用 Exit 命令主动退出
		// 关闭需要主动发送 Ctrl-D 命令
		bcsConf := podmanager.GetBCSConfByClusterId(podCtx.AdminClusterId)
		return remoteStreamConn.WaitStreamDone(bcsConf, podCtx)
	})

	if err := eg.Wait(); err != nil {
		manager.GracefulCloseWebSocket(ctx, ws, connected, errors.Wrap(err, "Handle websocket"))
		return
	}

	manager.GracefulCloseWebSocket(ctx, ws, connected, nil)
}

type OpenSession struct {
	ContainerId   string `json:"container_id"`
	Operator      string `json:"operator" binding:"required"`
	Command       string `json:"command"`
	Namespace     string `json:"namespace"`
	PodName       string `json:"pod_name"`
	ContainerName string `json:"container_name"`
}

func (s *service) CreateOpenWebConsoleSession(c *gin.Context) {
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")

	var openSession OpenSession

	err := c.BindJSON(&openSession)
	if err != nil {
		msg := i18n.GetMessage("请求参数错误")
		APIError(c, msg)
		return
	}
	commands := manager.DefaultCommand
	if openSession.Command == "" {
		commands = []string{}
	}

	mode := types.ContainerDirectMode

	startupMgr, err := podmanager.NewStartupManager(c.Request.Context(), mode, clusterId)
	if err != nil {
		msg := i18n.GetMessage("k8s客户端初始化失败{}", map[string]string{"err": err.Error()})
		APIError(c, msg)
		return
	}

	podCtx := &types.PodContext{
		ProjectId:      projectId,
		AdminClusterId: clusterId,
		ClusterId:      clusterId,
		Mode:           mode,
		Username:       "",
		Commands:       commands,
	}

	// 优先使用containerID
	if openSession.ContainerId != "" {
		resp, err := startupMgr.GetK8sContextByContainerID(openSession.ContainerId)
		if err != nil {
			msg := i18n.GetMessage("container_id不正确，请检查参数", err)
			APIError(c, msg)
			return
		}
		podCtx.Namespace = resp.Namespace
		podCtx.PodName = resp.PodName
		podCtx.ContainerName = resp.ContainerName
		podCtx.Commands = manager.DefaultCommand
	} else if openSession.Namespace != "" && openSession.PodName != "" && openSession.ContainerName != "" {
		podCtx = &types.PodContext{
			Namespace: openSession.Namespace,
		}
	} else {
		msg := i18n.GetMessage("container_id或namespace/pod_name/container_name不能同时为空")
		APIError(c, msg)
		return
	}

	store := sessions.NewRedisStore("open-session", "open-session")

	sessionId, err := store.Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", err)
		APIError(c, msg)
		return
	}

	webConsoleUrl := path.Join(s.opts.RoutePrefix, "/") + "/"

	query := url.Values{}
	query.Set("session_id", sessionId)
	query.Set("container_name", podCtx.ContainerName)

	webConsoleUrl = fmt.Sprintf("%s%s?%s", config.G.Web.Host, webConsoleUrl, query.Encode())

	respData := types.APIResponse{
		Data: map[string]string{
			"session_id":      sessionId,
			"web_console_url": webConsoleUrl,
		},
		Code:      types.NoError,
		Message:   i18n.GetMessage("获取session成功"),
		RequestID: uuid.New().String(),
	}

	c.JSON(http.StatusOK, respData)
}

// APIError 简易的错误返回
func APIError(c *gin.Context, msg string) {
	data := types.APIResponse{
		Code:      types.ApiErrorCode,
		Message:   msg,
		RequestID: uuid.New().String(),
	}

	c.AbortWithStatusJSON(http.StatusOK, data)
}
