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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/podmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"path"
)

type service struct {
	opts *route.Options
}

// NewRouteRegistrar xxx
func NewRouteRegistrar(opts *route.Options) route.Registrar {
	return service{opts: opts}
}

// RegisterRoute xxx
func (s service) RegisterRoute(router gin.IRoutes) {
	web := router.Use(route.WebAuthRequired())

	// 跳转 URL
	web.GET("/user/login/", metrics.RequestCollect("UserLoginRedirect"), s.UserLoginRedirect)
	web.GET("/user/perm_request/", metrics.RequestCollect("UserPermRequestRedirect"), route.APIAuthRequired(),
		s.UserPermRequestRedirect)

	// html 页面
	web.GET("/projects/:projectId/clusters/:clusterId/", metrics.RequestCollect("IndexPage"), s.IndexPageHandler)
	web.GET("/projects/:projectId/mgr/", metrics.RequestCollect("MgrPage"), s.MgrPageHandler)
	web.GET("/portal/container/", metrics.RequestCollect("ContainerGatePage"), s.ContainerGatePageHandler)
	web.GET("/portal/cluster/", metrics.RequestCollect("ClusterGatePage"), s.ClusterGatePageHandler)

	// 公共接口, 如 metrics, healthy, ready, pprof 等
	web.GET("/-/healthy", s.HealthyHandler)
	web.GET("/-/ready", s.ReadyHandler)
	web.GET("/metrics", metrics.PromMetricHandler())
}

// IndexPageHandler index 页面
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
	promRequestQuery.Set("namespace", route.GetNamespace(c))
	promRequestUrl := path.Join(s.opts.RoutePrefix, "/user/perm_request") + "/" + "?" + promRequestQuery.Encode()

	// webconsole Url
	sessionUrl := path.Join(s.opts.RoutePrefix, fmt.Sprintf("/api/projects/%s/clusters/%s/session", projectId,
		clusterId)) + "/"

	encodedQuery := consoleQuery.MakeEncodedQuery()
	if encodedQuery != "" {
		sessionUrl = fmt.Sprintf("%s?%s", sessionUrl, encodedQuery)
	}

	settings := map[string]string{
		"SITE_STATIC_URL":      s.opts.RoutePrefix,
		"COMMON_EXCEPTION_MSG": "",
	}
	language, download := getLang(c)

	data := gin.H{
		"title":            clusterId,
		"session_url":      sessionUrl,
		"login_url":        loginUrl,
		"perm_request_url": promRequestUrl,
		"guide_doc_links":  config.G.WebConsole.GuideDocLinks,
		"project_id":       projectId,
		"cluster_id":       clusterId,
		"settings":         settings,
		"Language":         language,
		"download":         download,
	}

	c.HTML(http.StatusOK, "index.html", data)
}

// MgrPageHandler 多集群页面
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
	lang, download := getLang(c)
	sessionUrl = fmt.Sprintf("%s?lang=%s", sessionUrl, lang)

	settings := map[string]string{
		"SITE_STATIC_URL":      s.opts.RoutePrefix,
		"COMMON_EXCEPTION_MSG": "",
	}

	data := gin.H{
		"title":       containerName,
		"session_url": sessionUrl,
		"settings":    settings,
		"Language":    lang,
		"download":    download,
	}

	c.HTML(http.StatusOK, "index.html", data)
}

// ClusterGatePageHandler 开放的页面WebConsole页面
func (s *service) ClusterGatePageHandler(c *gin.Context) {

}

// HealthyHandler xxx
func (s *service) HealthyHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("OK"))
}

// ReadyHandler xxx
func (s *service) ReadyHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("OK"))
}

// getLang 返回lang参数和模版变量download
func getLang(c *gin.Context) (string, string) {
	//只有lang为en时返回en,其他都返回zh
	lang := c.Query("lang")
	if lang == "en" {
		return i18n.GetMessage("语言"), i18n.GetMessage("下载")
	} else {
		return i18n.GetMessage("语言"), i18n.GetMessage("下载")
	}
}
