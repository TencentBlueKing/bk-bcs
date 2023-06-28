/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

// Package service NOTES
package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"bscp.io/cmd/feed-server/bll"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/repository"
	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/runtime/handler"
	"bscp.io/pkg/runtime/shutdown"
	"bscp.io/pkg/serviced"
	"bscp.io/pkg/tools"
)

// Service do all the data service's work
type Service struct {
	bll *bll.BLL
	// authorizer auth related operations.
	authorizer auth.Authorizer
	serve      *http.Server
	state      serviced.State
	provider   repository.Provider

	// name feed server instance name.
	name string
	// dsSetting down stream related setting.
	dsSetting cc.Downstream
	mc        *metric
}

// NewService create a service instance.
func NewService(sd serviced.Discover, name string) (*Service, error) {

	state, ok := sd.(serviced.State)
	if !ok {
		return nil, errors.New("discover convert state failed")
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

	return &Service{
		bll:        bl,
		authorizer: authorizer,
		state:      state,
		name:       name,
		dsSetting:  cc.FeedServer().Downstream,
		provider:   provider,
		mc:         initMetric(name),
	}, nil
}

// ListenAndServeRest listen and serve the restful server
func (s *Service) ListenAndServeRest() error {
	network := cc.FeedServer().Network
	server := &http.Server{
		Addr:    net.JoinHostPort(network.BindIP, strconv.FormatUint(uint64(network.HttpPort), 10)),
		Handler: s.handler(),
	}

	if network.TLS.Enable() {
		tls := network.TLS
		tlsC, err := tools.ClientTLSConfVerify(tls.InsecureSkipVerify, tls.CAFile, tls.CertFile, tls.KeyFile,
			tls.Password)
		if err != nil {
			return fmt.Errorf("init restful tls config failed, err: %v", err)
		}

		server.TLSConfig = tlsC
	}

	logs.Infof("listen restful server on %s with secure(%v) now.", server.Addr, network.TLS.Enable())

	go func() {
		notifier := shutdown.AddNotifier()
		select {
		case <-notifier.Signal:
			defer notifier.Done()

			logs.Infof("start shutdown restful server gracefully...")

			ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Second)
			defer cancel()
			if err := server.Shutdown(ctx); err != nil {
				logs.Errorf("shutdown restful server failed, err: %v", err)
				return
			}

			logs.Infof("shutdown restful server success...")
		}
	}()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logs.Errorf("serve restful server failed, err: %v", err)
			shutdown.SignalShutdownGracefully()
		}
	}()

	s.serve = server

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
