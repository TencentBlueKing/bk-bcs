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

package bcs

import (
	"errors"
	"net/http"
	"net/http/pprof"
	"strconv"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	watchoptions "github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/options"
)

var (
	httpServerOnce sync.Once
	httpServer     *HTTPServerWrap
)

// GetHTTPServer init HttpServerWrap object
func GetHTTPServer(config *watchoptions.WatchConfig, opts ...Option) *HTTPServerWrap {
	httpServerOnce.Do(func() {
		httpServer = newHTTPServerWrap(config, opts...)
	})

	return httpServer
}

const (
	// DefaultFalseHTTPDebug for pprof false
	DefaultFalseHTTPDebug = false
	// DefaultTrueHTTPDebug for pprof true
	DefaultTrueHTTPDebug = true
)

var (
	// ErrHTTPServerNotInit http server not init
	ErrHTTPServerNotInit = errors.New("http server is nil")
)

// CertConfig is configuration of Cert
type CertConfig struct {
	CAFile   string
	CertFile string
	KeyFile  string
	CertPwd  string
	IsSSL    bool
}

type options struct {
	debug   bool
	certCfg CertConfig
}

// Option for http server parameter
type Option func(opt *options)

// WithCertConfig set certConfig for http server
func WithCertConfig(certConfig CertConfig) Option {
	return func(opt *options) {
		opt.certCfg = certConfig
	}
}

// WithDebug determine to on/off pprof
func WithDebug(debug bool) Option {
	return func(opt *options) {
		opt.debug = debug
	}
}

// HTTPServerWrap http server object
type HTTPServerWrap struct {
	server *httpserver.HttpServer
	opt    *options
}

func newHTTPServerWrap(config *watchoptions.WatchConfig, opts ...Option) *HTTPServerWrap {
	defaultCertConfig := CertConfig{
		CertPwd: static.ServerCertPwd,
		IsSSL:   false,
	}

	httpOptions := &options{
		debug:   DefaultFalseHTTPDebug,
		certCfg: defaultCertConfig,
	}

	for _, opt := range opts {
		opt(httpOptions)
	}

	if len(httpOptions.certCfg.CertFile) > 0 && len(httpOptions.certCfg.KeyFile) > 0 {
		httpOptions.certCfg.IsSSL = true
	}

	setDefaultHTTPServer(config)

	httpServer := httpserver.NewHttpServer(config.Port, config.Address, "")
	if httpOptions.certCfg.IsSSL {
		httpServer.SetSsl(httpOptions.certCfg.CAFile, httpOptions.certCfg.CertFile,
			httpOptions.certCfg.KeyFile, httpOptions.certCfg.CertPwd)
		blog.Infof("http server set ssl successful")
	}

	serverWrap := &HTTPServerWrap{
		server: httpServer,
		opt:    httpOptions,
	}

	// register debug pprof & health api
	_ = serverWrap.registerInitAction()

	return serverWrap
}

func (s *HTTPServerWrap) registerInitAction() error {
	if s == nil {
		return ErrHTTPServerNotInit
	}
	actions := []*httpserver.Action{}

	// debug action
	if s.opt.debug {
		debugAction := []*httpserver.Action{
			httpserver.NewAction("GET", "/debug/pprof/", nil, getRouteFunc(pprof.Index)),
			httpserver.NewAction("GET", "/debug/pprof/{uri:*}", nil, getRouteFunc(pprof.Index)),
			httpserver.NewAction("GET", "/debug/pprof/cmdline", nil, getRouteFunc(pprof.Cmdline)),
			httpserver.NewAction("GET", "/debug/pprof/profile", nil, getRouteFunc(pprof.Profile)),
			httpserver.NewAction("GET", "/debug/pprof/symbol", nil, getRouteFunc(pprof.Symbol)),
			httpserver.NewAction("GET", "/debug/pprof/trace", nil, getRouteFunc(pprof.Trace)),
		}
		actions = append(actions, debugAction...)
	}

	// health API
	healthAction := []*httpserver.Action{
		httpserver.NewAction("GET", "/healthz", nil, func(req *restful.Request, resp *restful.Response) {
			resp.WriteHeader(http.StatusOK)
			_, _ = resp.Write([]byte("ok"))
		}),
	}

	actions = append(actions, healthAction...)

	if len(actions) > 0 {
		_ = s.server.RegisterWebServer("", nil, actions)
	}

	return nil
}

// RegisterWebServer will register rootPath and router actions
func (s *HTTPServerWrap) RegisterWebServer(rootPath string, actions []*httpserver.Action) error {
	if s == nil {
		return ErrHTTPServerNotInit
	}

	_ = s.server.RegisterWebServer(rootPath, nil, actions)
	return nil
}

// ListenAndServe listen server
func (s *HTTPServerWrap) ListenAndServe() error {
	if s == nil {
		return ErrHTTPServerNotInit
	}

	err := s.server.ListenAndServe()
	if err != nil {
		return err
	}

	return nil
}

func setDefaultHTTPServer(config *watchoptions.WatchConfig) {
	if config.Address == "" {
		config.Address = "127.0.0.1"
	}

	if config.Port <= 0 {
		config.Port = 8080
	}
}

var apiActions = make([]*httpserver.Action, 0, 100)

// GetHTTPServerAction get HTTP server register Actions
func GetHTTPServerAction() []*httpserver.Action {
	return apiActions
}

// RegisterAction register action, router url bind handler
func RegisterAction(action httpserver.Action) {
	apiActions = append(apiActions, httpserver.NewAction(action.Verb, action.Path, action.Params, action.Handler))
}

// getRouteFunc http.HandlerFunc trans to restful.RouteFunction
func getRouteFunc(f http.HandlerFunc) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		f(resp, req.Request)
	}
}

func getRouteHandlerFunc(f http.Handler) restful.RouteFunction { //nolint
	return func(req *restful.Request, resp *restful.Response) {
		f.ServeHTTP(resp, req.Request)
	}
}

// RunPrometheusMetricsServer starting prometheus metrics handler
func RunPrometheusMetricsServer(config *watchoptions.WatchConfig) {
	http.Handle("/metrics", promhttp.Handler())
	addr := config.Address + ":" + strconv.Itoa(int(config.MetricPort))
	go http.ListenAndServe(addr, nil) // nolint
}
