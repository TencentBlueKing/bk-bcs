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
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	gintrace "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/gin"
	"github.com/gin-gonic/gin"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/podmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/tracing"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/route"
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
	web := router.Use(route.WebAuthRequired(), gintrace.Middleware(tracing.ServiceName))

	// 跳转 URL
	web.GET("/user/perm_request/", metrics.RequestCollect("UserPermRequestRedirect"), route.APIAuthRequired(),
		s.UserPermRequestRedirect)

	// html 页面
	web.GET("/projects/:projectId/clusters/:clusterId/", metrics.RequestCollect("IndexPage"), s.IndexPageHandler)
	web.GET("/projects/:projectId/mgr/", metrics.RequestCollect("MgrPage"), s.MgrPageHandler)
	web.GET("/portal/container/", metrics.RequestCollect("ContainerGatePage"), s.ContainerGatePageHandler)
	web.GET("/portal/cluster/", metrics.RequestCollect("ClusterGatePage"), s.ClusterGatePageHandler)
	web.GET("/replay/files", metrics.RequestCollect("ReplayFilesPage"), route.APIAuthRequired(), route.ManagersRequired(), s.ReplayFoldersPageHandler)
	web.GET("/replay/files/:folder", metrics.RequestCollect("ReplayFilesPage"), route.APIAuthRequired(), route.ManagersRequired(), s.ReplayFilesPageHandler)
	web.GET("/replay/:folderName/:fileName", metrics.RequestCollect("ReplayDetailPage"), route.APIAuthRequired(), route.ManagersRequired(), s.ReplayDetailPageHandler)

	// 公共接口, 如 metrics, healthy, ready, pprof 等
	web.GET("/-/healthy", s.HealthyHandler)
	web.GET("/-/ready", s.ReadyHandler)
	web.GET("/metrics", metrics.PromMetricHandler())
}

// ReplayFilesPageHandler 回放文件
func (s *service) ReplayFilesPageHandler(c *gin.Context) {
	folderName := c.Param("folder")
	baseDir := config.G.Audit.DataDir
	dir := filepath.Join(baseDir, folderName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		klog.Errorf("read dir err", err)
		return
	}
	fileNames := make([]string, 0)
	for _, entry := range entries {
		if entry.Type().IsRegular() {
			fileNames = append(fileNames, entry.Name())
		}
	}
	data := gin.H{
		"folder_name":     folderName,
		"file_names":      fileNames,
		"SITE_STATIC_URL": s.opts.RoutePrefix,
	}
	c.HTML(http.StatusOK, "replay.html", data)
}

// ReplayFoldersPageHandler 回放文件目录
func (s *service) ReplayFoldersPageHandler(c *gin.Context) {
	dirname := config.G.Audit.DataDir
	entries, err := os.ReadDir(dirname)
	if err != nil {
		klog.Errorf("read dir err", err)
		return
	}
	folderNames := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			folderNames = append(folderNames, entry.Name())
		}
	}
	data := gin.H{
		"folder_names":    folderNames,
		"SITE_STATIC_URL": s.opts.RoutePrefix,
	}
	c.HTML(http.StatusOK, "replay.html", data)
}

// ReplayDetailPageHandler 回放终端记录文件
func (s service) ReplayDetailPageHandler(c *gin.Context) {
	folder := c.Param("folderName")
	file := c.Param("fileName")
	routePrefix := config.G.Web.RoutePrefix
	data := fmt.Sprintf("%s/casts/%s/%s", routePrefix, folder, file)
	res := gin.H{
		"data":            data,
		"SITE_STATIC_URL": s.opts.RoutePrefix,
	}
	c.HTML(http.StatusOK, "asciinema.html", res)
}

// IndexPageHandler index 页面
func (s *service) IndexPageHandler(c *gin.Context) {
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")
	consoleQuery := new(podmanager.ConsoleQuery)
	c.BindQuery(consoleQuery)

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
	language, download := i18n.GetMessage(c, "语言"), i18n.GetMessage(c, "下载")

	data := gin.H{
		"title":            clusterId,
		"session_url":      sessionUrl,
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
	lang := c.Query("lang")

	settings := map[string]string{"SITE_URL": s.opts.RoutePrefix}

	// 权限申请Url
	promRequestQuery := url.Values{}
	promRequestQuery.Set("project_id", projectId)
	promRequestUrl := path.Join(s.opts.RoutePrefix, "/user/perm_request") + "/" + "?" + promRequestQuery.Encode()

	data := gin.H{
		"settings":         settings,
		"project_id":       projectId,
		"perm_request_url": promRequestUrl,
		"Language":         lang,
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
	lang, download := i18n.GetMessage(c, "语言"), i18n.GetMessage(c, "下载")
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
