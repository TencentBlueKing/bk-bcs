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
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

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

const (
	confFilePath     = "frontend/dist/static/config.json"
	defaultStaticURL = "/web"
	// SITE_URL 前端Vue配置, 修改影响用户路由
	defaultSiteURL = "/bcs"
	// siteURLHeaderKey 前端前缀URL
	siteURLHeaderKey = "X-BCS-SiteURL"
)

// EmbedWebServer
type EmbedWebServer interface {
	IndexHandler() http.Handler
	FaviconHandler(w http.ResponseWriter, r *http.Request)
	StaticFileHandler(prefix string) http.Handler
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
	tpl := template.Must(template.New("").ParseFS(frontendAssets, "frontend/dist/*.html"))

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
	r.URL.Path = "/static/images/favicon.ico"

	// 添加缓存
	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "max-age=86400, public")

	e.fsServer.ServeHTTP(w, r)
}

// readConfigFile 读取前端配置文件
func readConfigFile() (map[string]string, error) {
	data, err := frontendAssets.ReadFile(confFilePath)
	if err != nil {
		return nil, err
	}

	c := new(map[string]string)
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return *c, nil
}

// mergeConfig 合并默认和自定义配置
func mergeConfig() ([]byte, error) {
	c, err := readConfigFile()
	if err != nil {
		return nil, err
	}
	for k, v := range config.G.FrontendConf.Docs {
		c[k] = v
	}

	bcsConfigBytes, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}
	return bcsConfigBytes, nil
}

// IndexHandler Vue 模板渲染
func (e *embedWeb) IndexHandler() http.Handler {
	bcsConfigBytes, err := mergeConfig()
	if err != nil {
		panic(fmt.Errorf("init bcs config err, %s", err))
	}
	bcsConfig := string(bcsConfigBytes)

	fn := func(w http.ResponseWriter, r *http.Request) {
		// 头部指定 SiteURL, 使用头部的， 多域名访问场景
		siteURL := r.Header.Get(siteURLHeaderKey)
		if siteURL == "" {
			siteURL = path.Join(config.G.Web.RoutePrefix, defaultSiteURL)
		}

		// 首页根路径下重定向跳转到 siteURL 前缀
		if r.URL.Path == "/" {
			http.Redirect(w, r, siteURL, http.StatusMovedPermanently)
		}

		data := map[string]string{
			"STATIC_URL":              path.Join(config.G.Web.RoutePrefix, defaultStaticURL),
			"SITE_URL":                siteURL,
			"RUN_ENV":                 config.G.Base.RunEnv,
			"REGION":                  config.G.Base.Region,
			"PREFERRED_DOMAINS":       config.G.Web.PreferredDomains,
			"DEVOPS_HOST":             config.G.FrontendConf.Host.DevOpsHost,
			"DEVOPS_BCS_API_URL":      config.G.FrontendConf.Host.DevOpsBCSAPIURL,
			"DEVOPS_ARTIFACTORY_HOST": config.G.FrontendConf.Host.DevOpsArtifactoryHost,
			"BK_IAM_APP_URL":          config.G.FrontendConf.Host.BKIAMAppURL,
			"PAAS_HOST":               config.G.FrontendConf.Host.BKPaaSHost,
			"BKMONITOR_HOST":          config.G.FrontendConf.Host.BKMonitorHost,
			"BK_CC_HOST":              config.G.FrontendConf.Host.BKCCHost,
			"BCS_API_HOST":            config.G.BCS.Host,
			"BCS_DEBUG_API_HOST":      config.G.BCSDebugAPIHost(),
			"BCS_CONFIG":              bcsConfig,
		}

		if config.G.IsLocalDevMode() {
			data["DEVOPS_BCS_API_URL"] = fmt.Sprintf("%s/backend", config.G.Web.Host)
			data["BCS_API_HOST"] = config.G.Web.Host
		}

		e.tpl.ExecuteTemplate(w, "index.html", data)
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

// StaticFileHandler 静态文件资源
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
