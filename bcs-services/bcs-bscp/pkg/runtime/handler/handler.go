/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package handler NOTES
package handler

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof"
	"net/url"
	"strings"

	"k8s.io/klog/v2"

	"bscp.io/pkg/metrics"
	"bscp.io/pkg/runtime/ctl"
)

// Handler defines http handlers to add to http service
type Handler struct {
	Pattern string
	Handler http.Handler
}

// HTTPMiddleware is the http middleware that adds common handler to root http handler
func HTTPMiddleware(root http.Handler, handlers ...Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", root.ServeHTTP)

	// add metrics handler
	mux.HandleFunc("/metrics", metrics.Handler().ServeHTTP)

	// add pprof handler
	mux.HandleFunc("/debug/", http.DefaultServeMux.ServeHTTP)

	// add tools handler
	mux.HandleFunc("/ctl", ctl.Handler().ServeHTTP)

	for _, handler := range handlers {
		mux.HandleFunc(handler.Pattern, handler.Handler.ServeHTTP)
	}

	return mux
}

// RegisterCommonHandler 统一的 metrics  函数
func RegisterCommonHandler() http.Handler {
	mux := http.NewServeMux()

	// add metrics handler
	mux.HandleFunc("/metrics", metrics.Handler().ServeHTTP)

	// add pprof handler
	mux.HandleFunc("/debug/", http.DefaultServeMux.ServeHTTP)

	return mux
}

// RegisterCommonToolHandler 统一的 metrics / ctl 函数
func RegisterCommonToolHandler() http.Handler {
	mux := http.NewServeMux()

	// add metrics handler
	mux.HandleFunc("/metrics", metrics.Handler().ServeHTTP)

	// add pprof handler
	mux.HandleFunc("/debug/", http.DefaultServeMux.ServeHTTP)

	// add tools handler
	mux.HandleFunc("/ctl", ctl.Handler().ServeHTTP)

	return mux
}

// ReverseProxyHandler 代理请求
func ReverseProxyHandler(name, remoteURL string) http.Handler {
	remote, err := url.Parse(remoteURL)
	if err != nil {
		panic(fmt.Errorf("%s '%s' not valid: %s", name, remoteURL, err))
	}

	if remote.Scheme != "http" && remote.Scheme != "https" {
		panic(fmt.Errorf("%s '%s' scheme not supported", name, remoteURL))
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(remote)
		proxy.Director = func(req *http.Request) {
			req.Header = r.Header
			req.Host = remote.Host
			req.URL.Scheme = remote.Scheme
			req.URL.Host = remote.Host
			klog.InfoS("forward request", "name", name, "url", req.URL)
		}

		proxy.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

// CORS 跨域
func CORS(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// cors 处理
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		allowHeaders := []string{"Origin", "Content-Length", "Content-Type", "X-Requested-With"}
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowHeaders, ","))

		allowMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowMethods, ","))

		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
