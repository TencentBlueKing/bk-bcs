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
	"github.com/google/uuid"
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
	api.GET("/api/projects/:projectId/clusters/", route.PermissionRequired(), s.ListClusters)

	// 蓝鲸API网关鉴权 & App鉴权
	api.POST("/api/gate/sessions/:sessionId/", s.CreateGateSession)
	api.POST("/api/gate/projects/:projectId/clusters/:clusterId/container/", route.CredentialRequired(), s.CreateContainerGateSession)
	api.POST("/api/gate/projects/:projectId/clusters/:clusterId/cluster/", route.CredentialRequired(), s.CreateClusterGateSession)

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

// CreateWebConsoleSession 创建websocket session
func (s *service) CreateWebConsoleSession(c *gin.Context) {
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")
	consoleQuery := new(podmanager.ConsoleQuery)
	c.BindQuery(consoleQuery)

	authCtx, err := route.GetAuthContext(c)
	if err != nil {
		APIError(c, i18n.GetMessage(err.Error()))
		return
	}

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
		RequestID: uuid.New().String(),
	}
	c.JSON(http.StatusOK, data)
}

func (s *service) CreateGateSession(c *gin.Context) {
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

func (s *service) CreateContainerGateSession(c *gin.Context) {
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")

	consoleQuery := new(podmanager.OpenQuery)

	err := c.BindJSON(consoleQuery)
	if err != nil {
		msg := i18n.GetMessage("请求参数错误")
		APIError(c, msg)
		return
	}

	podCtx, err := podmanager.QueryOpenPodCtx(c.Request.Context(), clusterId, consoleQuery)
	if err != nil {
		msg := i18n.GetMessage("请求参数错误")
		APIError(c, msg)
		return
	}
	podCtx.ProjectId = projectId

	store := sessions.NewRedisStore("open-session", "open-session")

	sessionId, err := store.Set(c.Request.Context(), podCtx)
	if err != nil {
		msg := i18n.GetMessage("获取session失败{}", err)
		APIError(c, msg)
		return
	}

	webConsoleUrl := path.Join(s.opts.RoutePrefix, "/gate/container/") + "/"

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

func (s *service) CreateClusterGateSession(c *gin.Context) {
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
