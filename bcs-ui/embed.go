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
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"html/template"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
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

var funcMap = template.FuncMap{
	"STATIC_URL":              STATIC_URL,
	"REGION":                  REGION,
	"LOGIN_FULL":              LOGIN_FULL,
	"DEVOPS_HOST":             DEVOPS_HOST,
	"DEVOPS_BCS_API_URL":      DEVOPS_BCS_API_URL,
	"RUN_ENV":                 RUN_ENV,
	"DEVOPS_ARTIFACTORY_HOST": DEVOPS_ARTIFACTORY_HOST,
	"SITE_URL":                SITE_URL,
	"BK_IAM_APP_URL":          BK_IAM_APP_URL,
	"PAAS_HOST":               PAAS_HOST,
	"BKMONITOR_HOST":          BKMONITOR_HOST,
	"BCS_API_HOST":            BCS_API_HOST,
	"PREFERRED_DOMAINS":       PREFERRED_DOMAINS,
	"BK_CC_HOST":              BK_CC_HOST,
	"BCS_DEBUG_API_HOST":      BCS_DEBUG_API_HOST,
}

func STATIC_URL() string {
	return "/web/static"
}

func REGION() string {
	return "ce"
}
func LOGIN_FULL() string {
	return "/LOGIN_FULL"
}
func DEVOPS_HOST() string {
	return "htto://localhost:8080"
}
func DEVOPS_BCS_API_URL() string {
	return "/DEVOPS_BCS_API_URL"
}
func RUN_ENV() string {
	return "/RUN_ENV"
}

func DEVOPS_ARTIFACTORY_HOST() string {
	return "/DEVOPS_ARTIFACTORY_HOST"
}

func SITE_URL() string {
	return "/bcs"
}
func BK_IAM_APP_URL() string {
	return "/BK_IAM_APP_URL"
}
func PAAS_HOST() string {
	return "/PAAS_HOST"
}
func BKMONITOR_HOST() string {
	return "/BKMONITOR_HOST"
}
func BCS_API_HOST() string {
	return "/BCS_API_HOST"
}
func PREFERRED_DOMAINS() string {
	return "/PREFERRED_DOMAINS"
}
func BK_CC_HOST() string {
	return "/BK_CC_HOST"
}
func BCS_DEBUG_API_HOST() string {
	return "/BCS_DEBUG_API_HOST"
}

// EmbedWebServer
type EmbedWebServer interface {
	//RenderIndexHandler(conf *IndexConfig) http.Handler
	//FaviconHandler(w http.ResponseWriter, r *http.Request)
	//StaticFileHandler(prefix string) http.Handler
	IndexHandler() http.Handler
}

type gzipFileInfo struct {
	contentType  string
	contentSize  string
	lastModified string
	filePath     string
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
	tpl := template.Must(template.New("").Funcs(funcMap).ParseFS(frontendAssets, "frontend/dist/ce/*.html"))

	root := http.FS(dist)

	w := &embedWeb{
		dist:     dist,
		tpl:      tpl,
		root:     root,
		fsServer: http.FileServer(root),
	}
	return w
}

// FaviconHandler favicon Handler
func (e *embedWeb) FaviconHandler(w http.ResponseWriter, r *http.Request) {
	// 填写实际的 icon 路径
	r.URL.Path = "/favicon.ico"

	// 添加缓存
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "max-age=86400, public")

	e.fsServer.ServeHTTP(w, r)
}

// RenderIndexHandler vue html 模板渲染
func (e *embedWeb) RenderIndexHandler(conf *IndexConfig) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		tplData := map[string]string{
			"BK_STATIC_URL":   conf.StaticURL,
			"RUN_ENV":         conf.RunEnv,
			"BK_BCS_BSCP_API": conf.APIURL,
		}

		// 本地开发模式 / 代理请求
		if conf.ProxyAPI {
			tplData["BK_BCS_BSCP_API"] = "/bscp"
		}

		e.tpl.ExecuteTemplate(w, "index.html", tplData)
	}

	return http.HandlerFunc(fn)
}

func (e *embedWeb) shouldCompress(r *http.Request) (bool, *gzipFileInfo) {
	// 必须包含 gzip 编码
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		return false, nil
	}

	// 其他不支持的场景
	if strings.Contains(r.Header.Get("Connection"), "Upgrade") ||
		strings.Contains(r.Header.Get("Accept"), "text/event-stream") {
		return false, nil
	}

	upath := r.URL.Path
	fileExt := filepath.Ext(upath)
	if ok, exist := allowCompressExtentions[fileExt]; !exist || !ok {
		return false, nil
	}

	ctype := mime.TypeByExtension(fileExt)
	if ctype == "" {
		return false, nil
	}

	filePath := upath + ".gz"
	gzipFile, err := e.root.Open(filePath)
	if err != nil {
		return false, nil
	}

	fileInfo, err := gzipFile.Stat()
	if err != nil {
		return false, nil
	}

	info := &gzipFileInfo{
		filePath:     filePath,
		contentType:  ctype,
		contentSize:  strconv.FormatInt(fileInfo.Size(), 10),
		lastModified: fileInfo.ModTime().Format(http.TimeFormat),
	}

	return true, info
}

// StaticFileHandler 静态文件处理函数
func (e *embedWeb) StaticFileHandler(prefix string) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if ok, fileInfo := e.shouldCompress(r); ok {
			r.URL.Path = fileInfo.filePath

			w.Header().Add("Vary", "Accept-Encoding")
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Length", fileInfo.contentSize)
			w.Header().Set("Content-Type", fileInfo.contentType)
			// 添加缓存
			w.Header().Set("Cache-Control", "max-age=86400, public")
			// issue https://github.com/golang/go/issues/44854
			// w.Header().Set("Last-Modified", fileInfo.lastModified)
			w.Header().Del("Transfer-Encoding")
		}

		e.fsServer.ServeHTTP(w, r)
	}

	return http.StripPrefix(prefix, http.HandlerFunc(fn))
}

const (
// SITE_URL 前端Vue配置, 修改影响用户路由
//SITE_URL = "/bcs"
//STATIC_URL = "/web/static"
)

// IndexHandler Vue 模板渲染
func (e *embedWeb) IndexHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		data := map[string]string{
			"STATIC_URL":              STATIC_URL(),
			"SITE_URL":                SITE_URL(),
			"REGION":                  "ce",
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
