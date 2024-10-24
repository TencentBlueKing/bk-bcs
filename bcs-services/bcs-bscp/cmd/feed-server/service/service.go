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

// Package service NOTES
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"k8s.io/klog/v2"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/feed-server/bll/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/repository"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/auth"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/ratelimiter"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/rest"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/handler"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// Service do all the data service's work
type Service struct {
	bll *bll.BLL
	// authorizer auth related operations.
	authorizer auth.Authorizer
	serves     []*http.Server
	state      serviced.State
	provider   repository.Provider

	// name feed server instance name.
	name  string
	mc    *metric
	gwMux *runtime.ServeMux
	rl    *ratelimiter.RL
}

// NewService create a service instance.
func NewService(sd serviced.Discover, name string) (*Service, error) {

	state, ok := sd.(serviced.State)
	if !ok {
		return nil, errors.New("discover convert state failed")
	}

	gwMux, err := newFeedServerMux()
	if err != nil {
		return nil, fmt.Errorf("new gateway failed, err: %v", err)
	}

	authorizer, err := auth.NewAuthorizer(sd, cc.FeedServer().Network.TLS)
	if err != nil {
		return nil, fmt.Errorf("new authorizer failed, err: %v", err)
	}

	bl, err := bll.New(sd, authorizer, name)
	if err != nil {
		return nil, fmt.Errorf("initialize business logical layer failed, err: %v", err)
	}

	provider, err := repository.NewProvider(cc.FeedServer().Repository)
	if err != nil {
		return nil, fmt.Errorf("new repository provider failed, err: %v", err)
	}

	rl := ratelimiter.New(cc.FeedServer().RateLimiter)
	logs.Infof("init rate limiter, conf: %+v", cc.FeedServer().RateLimiter)

	return &Service{
		bll:        bl,
		authorizer: authorizer,
		state:      state,
		name:       name,
		provider:   provider,
		mc:         initMetric(name),
		gwMux:      gwMux,
		rl:         rl,
	}, nil
}

// ListenAndServeRest listen and serve the restful server
func (s *Service) ListenAndServeRest() error {
	network := cc.FeedServer().Network
	addr := tools.GetListenAddr(network.BindIP, int(network.HttpPort))
	dualStackListener := listener.NewDualStackListener()
	if e := dualStackListener.AddListenerWithAddr(addr); e != nil {
		return e
	}
	logs.Infof("http server listen address: %s", addr)

	for _, ip := range network.BindIPs {
		if ip == network.BindIP {
			continue
		}
		ipAddr := tools.GetListenAddr(ip, int(network.HttpPort))
		if e := dualStackListener.AddListenerWithAddr(ipAddr); e != nil {
			return e
		}
		logs.Infof("http server listen address: %s", ipAddr)
	}

	server := &http.Server{Handler: s.handler()}

	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init restful tls config failed, err: %v", err)
		}

		server.TLSConfig = tlsC
	}

	go func() {
		notifier := shutdown.AddNotifier()
		<-notifier.Signal
		defer notifier.Done()

		logs.Infof("start shutdown restful server gracefully...")

		ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logs.Errorf("shutdown restful server failed, err: %v", err)
			return
		}

		logs.Infof("shutdown restful server success...")
	}()

	go func() {
		if err := server.Serve(dualStackListener); err != nil && err != http.ErrServerClosed {
			logs.Errorf("serve restful server failed, err: %v", err)
			shutdown.SignalShutdownGracefully()
		}
	}()

	s.serves = append(s.serves, server)

	return nil
}

// ListenAndGwServerRest listen and grpc-gateway serve the restful server
func (s *Service) ListenAndGwServerRest() error {
	network := cc.FeedServer().Network

	addr := tools.GetListenAddr(network.BindIP, int(network.GwHttpPort))
	dualStackListener := listener.NewDualStackListener()
	if e := dualStackListener.AddListenerWithAddr(addr); e != nil {
		return e
	}
	logs.Infof("http server listen address: %s", addr)

	for _, ip := range network.BindIPs {
		if ip == network.BindIP {
			continue
		}
		ipAddr := tools.GetListenAddr(ip, int(network.GwHttpPort))
		if e := dualStackListener.AddListenerWithAddr(ipAddr); e != nil {
			return e
		}
		logs.Infof("http server listen address: %s", ipAddr)
	}

	server := &http.Server{Handler: s.handlerGw()}

	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init restful tls config failed, err: %v", err)
		}

		server.TLSConfig = tlsC
	}

	go func() {
		notifier := shutdown.AddNotifier()
		<-notifier.Signal
		defer notifier.Done()

		logs.Infof("start shutdown restful server gracefully...")

		ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logs.Errorf("shutdown restful server failed, err: %v", err)
			return
		}

		logs.Infof("shutdown restful server success...")
	}()

	go func() {
		if err := server.Serve(dualStackListener); err != nil && err != http.ErrServerClosed {
			logs.Errorf("serve restful server failed, err: %v", err)
			shutdown.SignalShutdownGracefully()
		}
	}()

	s.serves = append(s.serves, server)

	return nil
}

func (s *Service) handler() http.Handler {
	r := chi.NewRouter()
	r.Use(handler.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 公共方法
	r.Get("/-/healthy", s.HealthyHandler)
	r.Get("/-/ready", s.ReadyHandler)
	r.Get("/healthz", s.Healthz)

	r.Mount("/", handler.RegisterCommonToolHandler())
	return r
}

func (s *Service) handlerGw() http.Handler {
	r := chi.NewRouter()
	r.Use(handler.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/api/v1/feed", func(r chi.Router) {
		r.Get("/biz/{biz_id}/app/{app}/files/*", s.DownloadFile)
		r.Mount("/", s.gwMux)
	})
	return r
}

// DownloadFile download file from provider repo
// nolint:funlen
func (s *Service) DownloadFile(w http.ResponseWriter, r *http.Request) {
	kt := kit.FromGrpcContext(r.Context())

	// 获取token
	authorizationHeader := r.Header.Get("Authorization")
	if len(authorizationHeader) < 1 {
		render.Render(w, r, rest.Unauthorized(errors.New("missing authorization header")))
		return
	}

	authHeaderParts := strings.Split(authorizationHeader, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != "bearer" {
		render.Render(w, r, rest.Unauthorized(errors.New("invalid authorization header format")))
		return
	}

	bizIdStr := chi.URLParam(r, "biz_id")
	bizID, _ := strconv.Atoi(bizIdStr)
	if bizID == 0 {
		render.Render(w, r, rest.BadRequest(errors.New("biz id is required")))
		return
	}
	kt.BizID = uint32(bizID)

	appName := chi.URLParam(r, "app")
	if appName == "" {
		render.Render(w, r, rest.BadRequest(errors.New("app is required")))
		return
	}

	labels := r.URL.Query().Get("labels")

	remainingPath := chi.URLParam(r, "*")
	if remainingPath == "" {
		render.Render(w, r, rest.BadRequest(errors.New("file path is required")))
		return
	}

	filePath, fileName := tools.SplitPathAndName(remainingPath)
	if fileName == "" {
		render.Render(w, r, rest.BadRequest(errors.New("file name is required")))
		return
	}

	appID, err := s.bll.AppCache().GetAppID(kt, uint32(bizID), appName)
	if err != nil {
		render.Render(w, r, rest.BadRequest(fmt.Errorf("get app id failed, err: %v", err)))
		return
	}

	app, err := s.bll.AppCache().GetMeta(kt, kt.BizID, appID)
	if err != nil {
		render.Render(w, r, rest.BadRequest(fmt.Errorf("get app meta failed, err: %v", err)))
		return
	}

	// validate can file be downloaded by credential.
	match, err := s.bll.Auth().CanMatchCI(
		kt, uint32(bizID), app.Name, authHeaderParts[1], filePath, fileName)
	if err != nil {
		render.Render(w, r, rest.Unauthorized(fmt.Errorf("do authorization failed, err: %v", err)))
		return
	}

	if !match {
		render.Render(w, r, rest.PermissionDenied(errors.New("no permission to download file"), nil))
		return
	}

	labelsMap := make(map[string]string)
	if labels != "" {
		if err = json.Unmarshal([]byte(labels), &labelsMap); err != nil {
			render.Render(w, r, rest.BadRequest(errors.New("invalid labels format, not in the correct json format")))
			return
		}
	}

	meta := &types.AppInstanceMeta{
		BizID:  kt.BizID,
		App:    appName,
		AppID:  appID,
		Labels: labelsMap,
	}

	cancel := kt.CtxWithTimeoutMS(1500)
	defer cancel()

	metas, err := s.bll.Release().ListAppLatestReleaseMeta(kt, meta)
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
		return
	}

	data := findMatchingConfigItem(filePath, fileName, metas.ConfigItems)
	if data == nil {
		render.Render(w, r, rest.NotFound(fmt.Errorf("file does not exist")))
		return
	}

	body, contentLength, err := s.provider.Download(kt, data.CommitSpec.GetContent().GetSignature())
	if err != nil {
		render.Render(w, r, rest.BadRequest(err))
		return
	}
	defer body.Close()

	w.Header().Set("Content-Length", strconv.FormatInt(contentLength, 10))
	w.Header().Set("Content-Type", "application/octet-stream")
	_, err = io.Copy(w, body)
	if err != nil {
		klog.ErrorS(err, "download file", "sign", data.CommitSpec.GetContent().GetSignature())
		render.Render(w, r, rest.BadRequest(fmt.Errorf("download file failed, err: %v", err)))
		return
	}
}

// 查找匹配的 ConfigItem
func findMatchingConfigItem(filePath, fileName string,
	configItems []*types.ReleasedCIMeta) *types.ReleasedCIMeta {
	for _, v := range configItems {
		path1 := path.Join(filePath, fileName)
		path2 := path.Join(v.ConfigItemSpec.Path, v.ConfigItemSpec.Name)
		if path1 == path2 {
			return v
		}
	}
	return nil
}

// HealthyHandler livenessProbe 健康检查
func (s *Service) HealthyHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// ReadyHandler ReadinessProbe 健康检查
func (s *Service) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	s.Healthz(w, r)
}

// Healthz check whether the service is healthy.
func (s *Service) Healthz(w http.ResponseWriter, req *http.Request) {
	if shutdown.IsShuttingDown() {
		logs.Errorf("service healthz check failed, current service is shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "current service is shutting down"))
		return
	}

	if err := s.state.Healthz(); err != nil {
		logs.Errorf("etcd healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "etcd healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
}
