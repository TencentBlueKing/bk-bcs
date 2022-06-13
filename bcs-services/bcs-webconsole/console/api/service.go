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
	"path"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/podmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/rest"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"

	"github.com/gin-gonic/gin"
	"github.com/google/shlex"
	"github.com/pkg/errors"
)

type service struct {
	opts *route.Options
}

// NewRouteRegistrar
func NewRouteRegistrar(opts *route.Options) route.Registrar {
	return service{opts: opts}
}

// RegisterRoute
func (s service) RegisterRoute(router gin.IRoutes) {
	api := router.Use(route.APIAuthRequired())

	// 用户登入态鉴权, session鉴权
	api.GET("/api/projects/:projectId/clusters/:clusterId/session/",
		metrics.RequestCollect("CreateWebConsoleSession"), route.PermissionRequired(), s.CreateWebConsoleSession)
	api.GET("/api/projects/:projectId/clusters/",
		metrics.RequestCollect("ListClusters"), s.ListClusters)

	// 蓝鲸API网关鉴权 & App鉴权
	api.GET("/api/portal/sessions/:sessionId/",
		metrics.RequestCollect("CreatePortalSession"), s.CreatePortalSession)
	api.POST("/api/portal/projects/:projectId/clusters/:clusterId/container/",
		metrics.RequestCollect("CreateContainerPortalSession"), route.CredentialRequired(), s.CreateContainerPortalSession)
	api.POST("/api/portal/projects/:projectId/clusters/:clusterId/cluster/",
		metrics.RequestCollect("CreateClusterPortalSession"), route.CredentialRequired(), s.CreateClusterPortalSession)

	// websocket协议, session鉴权
	api.GET("/ws/sessions/:sessionId/", metrics.RequestCollect("BCSWebSocket"), s.BCSWebSocketHandler)
}

// ListClusters 集群列表
func (s *service) ListClusters(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)

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
		RequestID: authCtx.RequestId,
	}
	c.JSON(http.StatusOK, data)
}

// CreateWebConsoleSession 创建websocket session
func (s *service) CreateWebConsoleSession(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)

	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")
	consoleQuery := new(podmanager.ConsoleQuery)
	c.BindQuery(consoleQuery)

	// 封装一个独立函数, 统计耗时
	podCtx, err := func() (podCtx *types.PodContext, err error) {
		start := time.Now()
		defer func() {
			if consoleQuery.IsContainerDirectMode() {
				return
			}

			// 单独统计 pod metrics
			podReadyDuration := time.Since(start)
			metrics.SetRequestIgnoreDuration(c, podReadyDuration)

			metrics.CollectPodReady(
				podmanager.GetAdminClusterId(clusterId),
				podmanager.GetNamespace(),
				podmanager.GetPodName(clusterId, authCtx.Username),
				err,
				podReadyDuration,
			)
		}()

		podCtx, err = podmanager.QueryAuthPodCtx(c.Request.Context(), clusterId, authCtx.Username, consoleQuery)
		return
	}()
	if err != nil {
		APIError(c, i18n.GetMessage(err.Error()))
		return
	}

	podCtx.ProjectId = projectId
	podCtx.Username = authCtx.Username
	podCtx.Source = consoleQuery.Source

	sessionId, err := sessions.NewStore().WebSocketScope().Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", err)
		APIError(c, msg)
		return
	}

	data := types.APIResponse{
		Data: map[string]string{
			"session_id": sessionId,
			"ws_url":     makeWebSocketURL(sessionId, consoleQuery.Lang, false),
		},
		Code:      types.NoError,
		Message:   i18n.GetMessage("获取session成功"),
		RequestID: authCtx.RequestId,
	}
	c.JSON(http.StatusOK, data)
}

// CreatePortalSession
func (s *service) CreatePortalSession(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)
	if authCtx.BindSession == nil {
		msg := i18n.GetMessage("sessin_id不正确")
		APIError(c, msg)
		return
	}

	podCtx := authCtx.BindSession

	sessionId, err := sessions.NewStore().WebSocketScope().Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", err)
		APIError(c, msg)
		return
	}

	data := types.APIResponse{
		Data: map[string]string{
			"session_id": sessionId,
			"ws_url":     makeWebSocketURL(sessionId, "", false),
		},
		Code:      types.NoError,
		Message:   i18n.GetMessage("获取session成功"),
		RequestID: authCtx.RequestId,
	}
	c.JSON(http.StatusOK, data)
}

// CreateContainerPortalSession 创建 webconsole url api
func (s *service) CreateContainerPortalSession(c *gin.Context) {
	authCtx := route.MustGetAuthContext(c)

	consoleQuery := new(podmanager.OpenQuery)

	err := c.BindJSON(consoleQuery)
	if err != nil {
		msg := i18n.GetMessage(fmt.Sprintf("请求参数错误, %s", err))
		APIError(c, msg)
		return
	}

	// 自定义命令行
	var commands []string
	if consoleQuery.Command != "" {
		commands, err = shlex.Split(consoleQuery.Command)
		if err != nil {
			msg := i18n.GetMessage(fmt.Sprintf("请求参数错误, command not valid, %s", err))
			APIError(c, msg)
			return
		}
	}

	podCtx, err := podmanager.QueryOpenPodCtx(c.Request.Context(), authCtx.ClusterId, consoleQuery)
	if err != nil {
		msg := i18n.GetMessage(fmt.Sprintf("请求参数错误, %s", err))
		APIError(c, msg)
		return
	}

	podCtx.ProjectId = authCtx.ProjectId
	// bkapigw 校验, 使用 Operator 做用户标识
	podCtx.Username = consoleQuery.Operator

	if len(commands) > 0 {
		podCtx.Commands = commands
	}

	sessionId, err := sessions.NewStore().OpenAPIScope().Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", err)
		APIError(c, msg)
		return
	}

	data := map[string]string{
		"session_id":      sessionId,
		"web_console_url": makeWebConsoleURL(sessionId, podCtx),
	}

	// 这里直接置换新的session_id
	if consoleQuery.WSAcquire {
		wsSessionId, err := sessions.NewStore().WebSocketScope().Set(c.Request.Context(), podCtx)
		if err != nil {
			msg := i18n.GetMessage("获取session失败{}", err)
			APIError(c, msg)
			return
		}

		data["ws_url"] = makeWebSocketURL(wsSessionId, "", true)
	}

	respData := types.APIResponse{
		Data:      data,
		Code:      types.NoError,
		Message:   i18n.GetMessage("获取session成功"),
		RequestID: authCtx.RequestId,
	}

	c.JSON(http.StatusOK, respData)
}

// makeWebConsoleURL webconsole 页面访问地址
func makeWebConsoleURL(sessionId string, podCtx *types.PodContext) string {
	u := *config.G.Web.BaseURL
	u.Path = path.Join(u.Path, "/portal/container/") + "/"

	query := url.Values{}
	query.Set("session_id", sessionId)
	query.Set("container_name", podCtx.ContainerName)

	u.RawQuery = query.Encode()

	return u.String()
}

// makeWebSocketURL http 转换为 ws 协议链接
func makeWebSocketURL(sessionId, lang string, withScheme bool) string {
	u := *config.G.Web.BaseURL
	u.Path = path.Join(u.Path, "/ws/sessions/", sessionId) + "/"

	query := url.Values{}
	if lang != "" {
		query.Set("lang", lang)
	}

	u.RawQuery = query.Encode()

	// https 协议 转换为 wss
	if u.Scheme == "https" {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}

	// 去掉前缀, web 使用
	if !withScheme {
		u.Scheme = ""
		u.Host = ""
	}

	return u.String()
}

// CreateClusterPortalSession 集群级别的 webconsole openapi
func (s *service) CreateClusterPortalSession(c *gin.Context) {
	rest.AbortWithBadRequestError(c, errors.New("Not implemented"))
}

// APIError 简易的错误返回
func APIError(c *gin.Context, msg string) {
	authCtx := route.MustGetAuthContext(c)

	data := types.APIResponse{
		Code:      types.ApiErrorCode,
		Message:   msg,
		RequestID: authCtx.RequestId,
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, data)
}
