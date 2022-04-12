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

package web

import (
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/podmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"

	"github.com/gin-gonic/gin"
)

type service struct {
	opts *route.Options
}

func NewRouteRegistrar(opts *route.Options) route.Registrar {
	return service{opts: opts}
}

func (s service) RegisterRoute(router gin.IRoutes) {
	web := router.Use(route.WebAuthRequired())

	// 跳转 URL
	web.GET("/user/login/", s.UserLoginRedirect)
	web.GET("/user/perm_request/", route.APIAuthRequired(), s.UserPermRequestRedirect)

	// html 页面
	web.GET("/projects/:projectId/clusters/:clusterId/", s.IndexPageHandler)
	web.GET("/projects/:projectId/mgr/", s.MgrPageHandler)
	web.GET("/portal/container/", s.ContainerGatePageHandler)
	web.GET("/portal/cluster/", s.ClusterGatePageHandler)

	// 公共接口, 如 metrics, healthy, ready, pprof 等
	web.GET("/-/healthy", s.HealthyHandler)
	web.GET("/-/ready", s.HealthyHandler)
	web.GET("/metrics", metrics.HandlerFunc())
}

func (s *service) IndexPageHandler(c *gin.Context) {
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")
	consoleQuery := new(podmanager.ConsoleQuery)
	c.BindQuery(consoleQuery)

	// 登入Url
	loginUrl := path.Join(s.opts.RoutePrefix, "/user/login") + "/"

	// 权限申请Url
	promRequestQuery := url.Values{}
	promRequestQuery.Set("project_id", projectId)
	promRequestQuery.Set("cluster_id", clusterId)
	promRequestUrl := path.Join(s.opts.RoutePrefix, "/user/perm_request") + "/" + "?" + promRequestQuery.Encode()

	// webconsole Url
	sessionUrl := path.Join(s.opts.RoutePrefix, fmt.Sprintf("/api/projects/%s/clusters/%s/session", projectId, clusterId)) + "/"

	encodedQuery := consoleQuery.MakeEncodedQuery()
	if encodedQuery != "" {
		sessionUrl = fmt.Sprintf("%s?%s", sessionUrl, encodedQuery)
	}

	settings := map[string]string{
		"SITE_STATIC_URL":      s.opts.RoutePrefix,
		"COMMON_EXCEPTION_MSG": "",
	}

	data := gin.H{
		"title":            clusterId,
		"session_url":      sessionUrl,
		"login_url":        loginUrl,
		"perm_request_url": promRequestUrl,
		"settings":         settings,
	}

	c.HTML(http.StatusOK, "index.html", data)
}

func (s *service) MgrPageHandler(c *gin.Context) {
	projectId := c.Param("projectId")

	settings := map[string]string{"SITE_URL": s.opts.RoutePrefix}

	// 登入Url
	loginUrl := path.Join(s.opts.RoutePrefix, "/user/login") + "/"

	// 权限申请Url
	promRequestQuery := url.Values{}
	promRequestQuery.Set("project_id", projectId)
	promRequestUrl := path.Join(s.opts.RoutePrefix, "/user/perm_request") + "/" + "?" + promRequestQuery.Encode()

	data := gin.H{
		"settings":         settings,
		"project_id":       projectId,
		"login_url":        loginUrl,
		"perm_request_url": promRequestUrl,
	}

	c.HTML(http.StatusOK, "mgr.html", data)
}

// ContainerGatePageHandler 开放的页面WebConsole页面
func (s *service) ContainerGatePageHandler(c *gin.Context) {
	sessionId := c.Query("session_id")
	containerName := c.Query("container_name")

	if containerName == "" {
		containerName = "--"
	}

	sessionUrl := path.Join(s.opts.RoutePrefix, fmt.Sprintf("/api/portal/sessions/%s/", sessionId)) + "/"

	settings := map[string]string{
		"SITE_STATIC_URL":      s.opts.RoutePrefix,
		"COMMON_EXCEPTION_MSG": "",
	}

	data := gin.H{
		"title":       containerName,
		"session_url": sessionUrl,
		"settings":    settings,
	}

	c.HTML(http.StatusOK, "index.html", data)
}

// ClusterGatePageHandler 开放的页面WebConsole页面
func (s *service) ClusterGatePageHandler(c *gin.Context) {

}

func (s *service) HealthyHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("OK"))
}

func (s *service) ReadyHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("OK"))
}
