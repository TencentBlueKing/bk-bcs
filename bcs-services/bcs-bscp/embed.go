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

package bscp

import (
	"embed"
	"html/template"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

//go:embed ui/dist
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
	IAMHost   string
	APIURL    string
	SiteURL   string // vue 路由前缀
	ProxyAPI  bool
}

// EmbedWebServer 前端 web server
type EmbedWebServer interface {
	RenderIndexHandler(conf *IndexConfig) http.Handler
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
	dist, err := fs.Sub(frontendAssets, "ui/dist")
	if err != nil {
		panic(err)
	}

	// 模版路径
	tpl := template.Must(template.New("").ParseFS(frontendAssets, "ui/dist/*.html"))

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
			"BK_IAM_HOST":     conf.IAMHost,
			"SITE_URL":        conf.SiteURL,
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
