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

// Package handler NOTES
package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	_ "net/http/pprof" //nolint
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"k8s.io/klog/v2"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/metrics"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/ctl"
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
func ReverseProxyHandler(name, prefix, remoteURL string) http.Handler {
	remote, err := url.Parse(remoteURL)
	if err != nil {
		panic(fmt.Errorf("%s '%s' not valid: %s", name, remoteURL, err))
	}

	if remote.Scheme != "http" && remote.Scheme != "https" {
		panic(fmt.Errorf("%s '%s' scheme not supported", name, remoteURL))
	}

	return &httputil.ReverseProxy{
		Rewrite: func(req *httputil.ProxyRequest) {
			req.SetURL(remote)
			req.SetXForwarded()
			if prefix != "" {
				req.Out.URL.Path = strings.TrimPrefix(req.Out.URL.Path, prefix)
				req.Out.URL.RawPath = strings.TrimPrefix(req.Out.URL.RawPath, prefix)
			}
			klog.InfoS("http proxy request",
				"name", name, "origionPath", req.In.URL.String(), "targetURL", req.Out.URL.String())
		},
	}
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

		allowHeaders := []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"X-Requested-With",
			"X-Bkapi-File-Content-Id",
			"X-Bkapi-File-Content-Overwrite",
			"X-Bscp-App-Id",
			"X-Bscp-Template-Space-Id",
			"X-Bscp-Unzip",
			"X-Bscp-Upload-Id",
			"X-Bscp-Part-Num",
			"X-Bscp-Operate-Way",
		}
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

// RequestIdGenerator request_id
func RequestIdGenerator() string {
	uid := uuid.New().String()
	requestId := strings.ReplaceAll(uid, "-", "")
	return requestId
}

// RequestID middleware
func RequestID(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(components.RequestIDHeaderKey)
		if requestID == "" {
			requestID = RequestIdGenerator()
		}

		ctx = components.WithRequestIDValue(ctx, requestID)
		ctx = context.WithValue(ctx, middleware.RequestIDKey, requestID)

		w.Header().Set(components.RequestIDHeaderKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(fn)
}

// RequestBodyLogger 记录 requetBody 的中间件
func RequestBodyLogger(ignorePattern ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			for _, p := range ignorePattern {
				if strings.Contains(r.RequestURI, p) {
					next.ServeHTTP(w, r)
					return
				}
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			buf := bytes.NewBuffer(nil)
			ww.Tee(buf)

			body, err := io.ReadAll(r.Body)
			if err != nil {
				render.Render(w, r, rest.BadRequest(err))
				return
			}

			defer func() {
				klog.Infof("REQ: url: %s, method: %s, body: %s, remote_addr: %s\nRESP: status: %d, body: %s",
					r.RequestURI,
					r.Method,
					body,
					r.RemoteAddr,
					ww.Status(),
					buf.String(),
				)
			}()

			r.Body = io.NopCloser(bytes.NewBuffer(body))
			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
