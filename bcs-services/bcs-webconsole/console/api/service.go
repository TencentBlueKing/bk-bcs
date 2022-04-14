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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/components/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/podmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/sessions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"

	"github.com/gin-gonic/gin"
	"github.com/google/shlex"
)

type service struct {
	opts *route.Options
}

func NewRouteRegistrar(opts *route.Options) route.Registrar {
	return service{opts: opts}
}

func (s service) RegisterRoute(router gin.IRoutes) {
	api := router.Use(route.APIAuthRequired())

	gp := metrics.New(s.opts.Router)
	api.Use(gp.Middleware())

	// 用户登入态鉴权, session鉴权
	api.GET("/api/projects/:projectId/clusters/:clusterId/session/", route.PermissionRequired(), s.CreateWebConsoleSession)
	api.GET("/api/projects/:projectId/clusters/", s.ListClusters)

	// 蓝鲸API网关鉴权 & App鉴权
	api.GET("/api/portal/sessions/:sessionId/", s.CreatePortalSession)
	api.POST("/api/portal/projects/:projectId/clusters/:clusterId/container/", route.CredentialRequired(), s.CreateContainerPortalSession)
	api.POST("/api/portal/projects/:projectId/clusters/:clusterId/cluster/", route.CredentialRequired(), s.CreateClusterPortalSession)

	// websocket协议, session鉴权
	api.GET("/ws/projects/:projectId/clusters/:clusterId/", s.BCSWebSocketHandler)
}

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

	podCtx, err := podmanager.QueryAuthPodCtx(c.Request.Context(), clusterId, authCtx.Username, consoleQuery)
	if err != nil {
		APIError(c, i18n.GetMessage(err.Error()))
		return
	}

	podCtx.ProjectId = projectId
	podCtx.Source = consoleQuery.Source

	store := sessions.NewRedisStore(projectId, clusterId)
	sessionId, err := store.Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", err)
		APIError(c, msg)
		return
	}

	query := url.Values{}
	query.Set("session_id", sessionId)
	if consoleQuery.Lang != "" {
		query.Set("lang", consoleQuery.Lang)
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
		RequestID: authCtx.RequestId,
	}
	c.JSON(http.StatusOK, data)
}

func (s *service) CreatePortalSession(c *gin.Context) {
	sessionId := c.Query("session_id")

	authCtx := route.MustGetAuthContext(c)
	if authCtx.BindSession == nil {
		msg := i18n.GetMessage("sessin_id不正确")
		APIError(c, msg)
		return
	}

	podCtx := authCtx.BindSession

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
		RequestID: authCtx.RequestId,
	}
	c.JSON(http.StatusOK, data)
}

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

	if len(commands) > 0 {
		podCtx.Commands = commands
	}

	store := sessions.NewRedisStore("open-session", "open-session")

	sessionId, err := store.Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", err)
		APIError(c, msg)
		return
	}

	webConsoleUrl := path.Join(s.opts.RoutePrefix, "/portal/container/") + "/"

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
		RequestID: authCtx.RequestId,
	}

	c.JSON(http.StatusOK, respData)
}

func (s *service) CreateClusterPortalSession(c *gin.Context) {
}

// APIError 简易的错误返回
func APIError(c *gin.Context, msg string) {
	authCtx := route.MustGetAuthContext(c)

	data := types.APIResponse{
		Code:      types.ApiErrorCode,
		Message:   msg,
		RequestID: authCtx.RequestId,
	}

	c.AbortWithStatusJSON(http.StatusOK, data)
}
