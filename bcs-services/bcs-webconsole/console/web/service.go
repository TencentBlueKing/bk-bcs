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
	router.Use(route.AuthRequired()).
		GET("/", s.SessionPageHandler).
		GET("/projects/:projectId/clusters/:clusterId/", s.IndexPageHandler).
		GET("/projects/:projectId/mgr/", s.MgrPageHandler).
		GET(path.Join(s.opts.RoutePrefix, "/")+"/", s.SessionPageHandler).
		GET(path.Join(s.opts.RoutePrefix, "/projects/:projectId/clusters/:clusterId/"), s.IndexPageHandler).
		GET(path.Join(s.opts.RoutePrefix, "/projects/:projectId/mgr/"), s.MgrPageHandler)
}

func (s *service) IndexPageHandler(c *gin.Context) {
	projectId := c.Param("projectId")
	clusterId := c.Param("clusterId")
	containerId := c.Query("container_id")

	query := url.Values{}

	if containerId != "" {
		query.Set("container_id", containerId)
	}

	sessionUrl := path.Join(s.opts.RoutePrefix, fmt.Sprintf("/api/projects/%s/clusters/%s/session", projectId, clusterId)) + "/"
	sessionUrl = fmt.Sprintf("%s?%s", sessionUrl, query.Encode())

	settings := map[string]string{
		"SITE_STATIC_URL":      s.opts.RoutePrefix,
		"COMMON_EXCEPTION_MSG": "",
	}

	data := gin.H{
		"title":       clusterId,
		"session_url": sessionUrl,
		"settings":    settings,
	}

	c.HTML(http.StatusOK, "index.html", data)
}

func (s *service) MgrPageHandler(c *gin.Context) {
	projectId := c.Param("projectId")

	settings := map[string]string{"SITE_URL": s.opts.RoutePrefix}

	data := gin.H{
		"settings":   settings,
		"project_id": projectId,
	}

	c.HTML(http.StatusOK, "mgr.html", data)
}

// SessionPageHandler 开放的页面WebConsole页面
func (s *service) SessionPageHandler(c *gin.Context) {
	sessionId := c.Query("session_id")
	containerName := c.Query("container_name")

	query := url.Values{}

	if containerName != "" {
		containerName = "--"
	}

	query.Set("session_id", sessionId)

	sessionUrl := path.Join(s.opts.RoutePrefix, "/api/open_session/") + "/"
	sessionUrl = fmt.Sprintf("%s?%s", sessionUrl, query.Encode())

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
