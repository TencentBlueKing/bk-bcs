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

	"bscp.io/cmd/feed-server/bll"
	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/auth"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/metrics"
	"bscp.io/pkg/rest"
	"bscp.io/pkg/runtime/ctl"
	"bscp.io/pkg/runtime/shutdown"
	"bscp.io/pkg/serviced"
	sfs "bscp.io/pkg/sf-share"
	"bscp.io/pkg/tools"

	"github.com/emicklei/go-restful/v3"
)

// Service do all the data service's work
type Service struct {
	bll *bll.BLL
	// authorizer auth related operations.
	authorizer auth.Authorizer
	serve      *http.Server
	state      serviced.State
	// name feed server instance name.
	name string
	// dsSetting down stream related setting.
	dsSetting cc.Downstream
	// repositoryTLS used send to sidecar to download file. if repository not in use tls, repositoryTLS is nil.
	repositoryTLS *sfs.TLSBytes
	mc            *metric
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

	tlsBytes, err := sfs.LoadTLSBytes(cc.FeedServer().Repository.TLS)
	if err != nil {
		return nil, fmt.Errorf("conv tls to tls bytes failed, err: %v", err)
	}

	return &Service{
		bll:           bl,
		authorizer:    authorizer,
		state:         state,
		name:          name,
		dsSetting:     cc.FeedServer().Downstream,
		repositoryTLS: tlsBytes,
		mc:            initMetric(name),
	}, nil
}

// ListenAndServeRest listen and serve the restful server
func (s *Service) ListenAndServeRest() error {

	root := http.NewServeMux()
	root.HandleFunc("/", s.apiSet().ServeHTTP)
	root.HandleFunc("/healthz", s.Healthz)
	root.HandleFunc("/debug/", http.DefaultServeMux.ServeHTTP)
	root.HandleFunc("/metrics", metrics.Handler().ServeHTTP)
	root.HandleFunc("/ctl", ctl.Handler().ServeHTTP)

	network := cc.FeedServer().Network
	server := &http.Server{
		Addr:    net.JoinHostPort(network.BindIP, strconv.FormatUint(uint64(network.HttpPort), 10)),
		Handler: root,
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

func (s *Service) apiSet() *restful.Container {

	handler := rest.NewHandler()
	handler.Add("pbfs.ListFileAppLatestReleaseMeta", "POST", "/api/v1/feed/list/app/release/type/file/latest",
		s.ListFileAppLatestReleaseMetaRest)

	handler.Add("pbfs.AuthRepo", "POST", "/api/v1/feed/auth/repository/file_pull", s.AuthRepoRest)

	c := restful.NewContainer()
	handler.Load(c)

	return c
}

// Healthz check whether the service is healthy.
func (s *Service) Healthz(w http.ResponseWriter, req *http.Request) {
	if shutdown.IsShuttingDown() {
		logs.Errorf("service healthz check failed, current service is shutting down")
		w.WriteHeader(http.StatusServiceUnavailable)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "current service is shutting down"))
		return
	}

	if err := s.state.Healthz(cc.FeedServer().Service.Etcd); err != nil {
		logs.Errorf("etcd healthz check failed, err: %v", err)
		rest.WriteResp(w, rest.NewBaseResp(errf.UnHealth, "etcd healthz error, "+err.Error()))
		return
	}

	rest.WriteResp(w, rest.NewBaseResp(errf.OK, "healthy"))
	return
}
