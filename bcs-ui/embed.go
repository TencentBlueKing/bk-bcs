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
 */

package bcsui

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
)

//go:embed frontend/dist
var frontendAssets embed.FS

var (
	allowCompressExtentions = map[string]bool{
		".js":  true,
		".css": true,
	}
)

// IndexConfig 前端配置
type IndexConfig struct {
	RunEnv    string
	StaticURL string
	APIURL    string
	ProxyAPI  bool
}

// EmbedWebServer
type EmbedWebServer interface {
	IndexHandler() http.Handler
}

type embedWeb struct {
	dist     fs.FS
	tpl      *template.Template
	root     http.FileSystem
	fsServer http.Handler
}

// NewEmbedWeb 初始化模版和fs
func NewEmbedWeb() *embedWeb {
	// dist 路径
	dist, err := fs.Sub(frontendAssets, "frontend/dist")
	if err != nil {
		panic(err)
	}

	// 模版路径
	tpl := template.Must(template.New("").ParseFS(frontendAssets, "frontend/dist/ce/*.html"))

	root := http.FS(dist)

	w := &embedWeb{
		dist:     dist,
		tpl:      tpl,
		root:     root,
		fsServer: http.FileServer(root),
	}
	return w
}

const (
	// SITE_URL 前端Vue配置, 修改影响用户路由
	SITE_URL   = "/bcs"
	STATIC_URL = "/web/static"
)

// IndexHandler Vue 模板渲染
func (e *embedWeb) IndexHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		data := map[string]string{
			"STATIC_URL":              STATIC_URL,
			"SITE_URL":                SITE_URL,
			"RUN_ENV":                 config.G.Base.RunEnv,
			"PREFERRED_DOMAINS":       config.G.Web.PreferredDomains,
			"DEVOPS_HOST":             config.G.FrontendConf.Host.DevOpsHost,
			"DEVOPS_BCS_API_URL":      config.G.FrontendConf.Host.DevOpsBCSAPIURL,
			"DEVOPS_ARTIFACTORY_HOST": config.G.FrontendConf.Host.DevOpsArtifactoryHost,
			"BK_IAM_APP_URL":          config.G.FrontendConf.Host.BKIAMAppURL,
			"PAAS_HOST":               config.G.FrontendConf.Host.PaaSHost,
			"BKMONITOR_HOST":          config.G.FrontendConf.Host.BKMonitorHOst,
			"BCS_API_HOST":            config.G.BCS.Host,
			"BK_CC_HOST":              config.G.FrontendConf.Host.BKCMDBHost,
		}

		if config.G.IsDevMode() {
			data["DEVOPS_BCS_API_URL"] = fmt.Sprintf("%s/backend", config.G.Web.Host)
			data["BCS_API_HOST"] = config.G.Web.Host
		}
		e.tpl.ExecuteTemplate(w, "index.html", data)
	}

	return http.HandlerFunc(fn)
}
