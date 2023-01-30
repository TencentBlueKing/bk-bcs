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

// Config 前端配置
type Config struct {
	Docs map[string]string `json:"docs"`
}

// WebStatic 静态资源
func WebStatic() fs.FS {
	static, err := fs.Sub(frontendAssets, "ui/dist")
	if err != nil {
		panic(err)
	}
	return static
}

// WebFaviconPath 站点 icon 路径
func WebFaviconPath() string {
	entrys, err := frontendAssets.ReadDir("ui/dist/static")
	if err != nil {
		panic(err)
	}
	for _, v := range entrys {
		if v.IsDir() {
			continue
		}
		if strings.Contains(v.Name(), "favicon") {
			return "/static/" + v.Name()
		}
	}
	panic("favicon not found")
}

// WebTemplate html 摸版
func WebTemplate() *template.Template {
	tpl := template.Must(template.New("").ParseFS(frontendAssets, "ui/dist/*.html"))
	return tpl
}

type gzipFileHandler struct {
	root     http.FileSystem
	fsServer http.Handler
}

type gzipFileInfo struct {
	contentType  string
	contentSize  string
	lastModified string
	filePath     string
}

func (h *gzipFileHandler) shouldCompress(r *http.Request) (bool, *gzipFileInfo) {
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
	gzipFile, err := h.root.Open(filePath)
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

func (h *gzipFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ok, fileInfo := h.shouldCompress(r); ok {
		r.URL.Path = fileInfo.filePath
		w.Header().Add("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Content-Length", fileInfo.contentSize)
		w.Header().Set("Content-Type", fileInfo.contentType)
		// issue https://github.com/golang/go/issues/44854
		// w.Header().Set("Last-Modified", fileInfo.lastModified)
		w.Header().Del("Transfer-Encoding")
	}

	h.fsServer.ServeHTTP(w, r)
}

// GZipFileServer GZIP 文件服务
func GZipFileServer(root http.FileSystem) http.Handler {
	return &gzipFileHandler{
		root:     root,
		fsServer: http.FileServer(root),
	}
}
