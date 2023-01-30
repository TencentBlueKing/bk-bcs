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
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"

	bscp "bscp.io"
)

// StaticCacheHandler 添加缓存配置
func StaticCacheHandler(c *gin.Context) {
	// cache one day
	c.Header("Cache-Control", "max-age=86400, public")
}

// StaticGZipHandler 支持 gzip 返回
func StaticGZipHandler(relativePath string, fs http.FileSystem) gin.HandlerFunc {
	fileServer := http.StripPrefix(relativePath, bscp.GZipFileServer(fs))

	return func(c *gin.Context) {
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}

// ReverseAPIHandler 代理请求， CORS 跨域问题
func ReverseAPIHandler(name, remoteURL string) gin.HandlerFunc {
	remote, err := url.Parse(remoteURL)
	if err != nil {
		panic(err)
	}

	if remote.Scheme != "http" && remote.Scheme != "https" {
		panic(fmt.Errorf("%s '%s' scheme not supported", name, remoteURL))
	}

	return func(c *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = c.Request.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			klog.InfoS("forward request", "name", name, "url", req.URL)
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// HealthyHandler 健康检查
func HealthyHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("OK"))
}

// ReadyHandler 健康检查
func ReadyHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte("OK"))
}
