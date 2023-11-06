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

// Package server defines the vaultplugin server
package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	traceconst "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/constants"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/pkg/metric"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-vaultplugin-server/pkg/secret"
)

// Server defines the server of vault plugin
type Server struct {
	ctx context.Context
	op  *options.Options

	secretManager secret.SecretManagerWithVersion
	argoStore     store.Store

	httpService  *http.Server
	metricServer *http.Server
}

// NewServer create the instance of valut plugin server
func NewServer(ctx context.Context, op *options.Options) *Server {
	return &Server{
		ctx: ctx,
		op:  op,
	}
}

// Init clients what the server need
func (s *Server) Init() error {
	s.secretManager = secret.NewSecretManager(s.op)
	if err := s.secretManager.Init(); err != nil {
		return errors.Wrapf(err, "init secret manager failed")
	}
	s.argoStore = store.NewStore(&store.Options{
		Service: s.op.Argo.Service,
		User:    s.op.Argo.User,
		Pass:    s.op.Argo.Pass,
		Cache:   true,
	})
	if err := s.argoStore.Init(); err != nil {
		return errors.Wrapf(err, "init argo store failed")
	}
	if err := s.initHTTPService(); err != nil {
		return errors.Wrapf(err, "init http service failed")
	}
	s.initMetricServer()
	return nil
}

func (s *Server) initMetricServer() {
	metricMux := http.NewServeMux()
	s.initPProf(metricMux)
	s.initMetric(metricMux)
	extraServerEndpoint := net.JoinHostPort(s.op.Address, strconv.Itoa(int(s.op.MetricPort)))

	s.metricServer = &http.Server{
		Addr:    extraServerEndpoint,
		Handler: metricMux,
	}
	go func() {
		blog.Infof("start extra modules [pprof, metric] server %s", extraServerEndpoint)
		if err := s.metricServer.ListenAndServe(); err != nil {
			blog.Errorf("metric server listen failed, err %s", err.Error())
		}
	}()
}

// initPProf 初始化 pprof
func (s *Server) initPProf(mux *http.ServeMux) {
	if !s.op.Debug {
		blog.Infof("pprof is disabled")
		return
	}
	blog.Infof("pprof is enabled")
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

// initMetric 初始化metric路由
func (s *Server) initMetric(mux *http.ServeMux) {
	blog.Infof("init metric handler")
	mux.Handle("/metrics", promhttp.Handler())
}

func requestID(ctx context.Context) string {
	return ctx.Value(traceconst.RequestIDHeaderKey).(string)
}

func (s *Server) initHTTPService() error {
	router := mux.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			requestID := r.Header.Get(traceconst.RequestIDHeaderKey)
			if requestID == "" {
				requestID = uuid.New().String()
			}
			r = r.WithContext(context.WithValue(r.Context(), traceconst.RequestIDHeaderKey, requestID))
			next.ServeHTTP(w, r)
			endTime := time.Now()
			cost := endTime.Sub(startTime).Seconds()
			blog.Infof("RequestID[%s] [%s] %s %.2f\n", requestID, r.Method, r.URL.Path, cost)
			metric.RequestTotal.WithLabelValues().Inc()
			metric.RequestDuration.WithLabelValues().Observe(cost)
		})
	})
	v1api := router.PathPrefix("/api/v1").Subrouter()
	v1api.HandleFunc("/health", s.healthy).Methods(http.MethodGet)
	v1api.HandleFunc("/decryption/manifest", s.routerDecryptManifest).Methods(http.MethodPost)
	v1api.HandleFunc("/secrets/{project}/{path}", s.routerSaveOrUpdateSecret).
		Methods(http.MethodPost, http.MethodPut)
	v1api.HandleFunc("/secrets/{project}/{path}", s.routerDeleteSecret).Methods(http.MethodDelete)
	v1api.HandleFunc("/secrets/{project}/{path}", s.routerGetSecret).
		Methods(http.MethodGet).Queries("version", "{version}")

	// 这里因为path可能为空，无法在url中定义空字段，所以用parameter来做path
	v1api.HandleFunc("/secrets/{project}/list", s.routerListSecret).
		Methods(http.MethodGet).Queries("path", "{path}")
	v1api.HandleFunc("/secrets/{project}/{path}/metadata", s.routerGetMetadata).Methods(http.MethodGet)
	v1api.HandleFunc("/secrets/{project}/{path}/version", s.routerGetVersion).Methods(http.MethodGet)
	v1api.HandleFunc("/secrets/{project}/{path}/rollback", s.routerRollback).
		Methods(http.MethodPost).Queries("version", "{version}")
	v1api.HandleFunc("/secrets/init", s.routerInitProject).
		Methods(http.MethodPost).Queries("project", "{project}")
	v1api.HandleFunc("/secrets/annotation", s.routerGetSecretAnnotation).
		Methods(http.MethodGet).Queries("project", "{project}")

	s.httpService = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.op.Address, s.op.HTTPPort),
		Handler: router,
	}
	err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		tpl, err := route.GetPathTemplate()
		if err == nil {
			blog.Infof("Path: %s,", tpl)
		}
		handler := route.GetHandler()
		if handler != nil {
			blog.Infof("Handler: %s\n", handler)
		}
		return nil
	})
	if err != nil {
		return errors.Wrapf(err, "router walk failed")
	}
	return nil
}

func (s *Server) startHTTPService(errCh chan error) {
	defer func() {
		if r := recover(); r != nil {
			blog.Errorf("http_server panic, stacktrace:\n%s", debug.Stack())
			errCh <- fmt.Errorf("http_server panic")
		}
	}()

	blog.Infof("http_server is started.")
	if err := s.httpService.ListenAndServe(); err != nil {
		blog.Errorf("http_server stopped with err: %s", err.Error())
		errCh <- err
		return
	}
	blog.Infof("http_server stopped.")
}

// Run all the servers
func (s *Server) Run() error {
	errChan := make(chan error, 1)
	go s.startHTTPService(errChan)
	select {
	case err := <-errChan:
		return err
	case <-s.ctx.Done():
		return nil
	}
}
