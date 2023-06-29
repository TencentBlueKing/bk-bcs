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

package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/vaultplugin-server/handler"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/vaultplugin-server/secret"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
)

// VaultPlugin implementation
type VaultPlugin struct {
	ctx           context.Context
	cancel        context.CancelFunc
	option        *Options
	httpService   *http.Server
	stops         []utils.StopFunc
	secret        secret.SecretManagerWithVersion
	gitopsStorage store.Store
}

// NewVaultPlugin create new VaultPlugin
func NewVaultPlugin(opt *Options) *VaultPlugin {
	ctx, cancel := context.WithCancel(context.Background())
	return &VaultPlugin{
		ctx:    ctx,
		cancel: cancel,
		stops:  make([]utils.StopFunc, 0),
		option: opt,
	}
}

// Init all service
func (p *VaultPlugin) Init() error {
	initializer := []func() error{
		p.initSecret, p.initStorage, p.initHTTPService,
	}

	for _, initFunc := range initializer {
		if err := initFunc(); err != nil {
			return err
		}
	}
	return nil
}

// Run VaultPlugin server
func (p *VaultPlugin) Run() error {
	runners := []func(){
		p.startHTTPService,
	}

	for _, runner := range runners {
		go runner()
	}

	<-p.ctx.Done()
	p.stop()

	blog.Infof("vaultplugin-server is under graceful period, %d seconds...", gracefulexit)
	time.Sleep(time.Second * gracefulexit)
	return nil
}

func (p *VaultPlugin) startHTTPService() {
	if p.httpService == nil {
		blog.Fatalf("vaultplugin-server lost http server instance")
		return
	}
	p.stops = append(p.stops, p.stopHTTPService)
	err := p.httpService.ListenAndServe()
	if err != nil {
		if http.ErrServerClosed == err {
			blog.Warnf("vaultplugin-server http service gracefully exit.")
			return
		}
		// start http gateway error, maybe port is conflict or something else
		blog.Fatal("vaultplugin-server http service ListenAndServe fatal, %s", err.Error())
	}
}

// stopHTTPService  gracefully stop
func (p *VaultPlugin) stopHTTPService() {
	ctx, cancel := context.WithTimeout(p.ctx, time.Second*2)
	defer cancel()
	if err := p.httpService.Shutdown(ctx); err != nil {
		blog.Errorf("vaultplugin-server gracefully shutdown http service failure: %s", err.Error())
		return
	}
	blog.Infof("vaultplugin-server shutdown http service gracefully")
}

// stop all services
func (p *VaultPlugin) stop() {
	for _, stop := range p.stops {
		go stop()
	}
}

func (p *VaultPlugin) initHTTPService() error {
	router := mux.NewRouter()

	v1api := router.PathPrefix("/api/v1").Subrouter()

	p.initVaultPluginHandler(v1api)
	// init http server
	p.httpService = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", p.option.Address, p.option.HTTPPort),
		Handler: router,
	}
	walkfunc := func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		tpl, err := route.GetPathTemplate()
		if err == nil {
			blog.Infof("Path: %s,", tpl)
		}

		handler := route.GetHandler()
		if handler != nil {
			blog.Infof("Handler: %s\n", handler)
		}

		return nil
	}
	router.Walk(walkfunc)
	return nil
}

func (s *VaultPlugin) initSecret() error {
	opt := &secret.Options{
		Type:      s.option.Secret.Type,
		Endpoints: s.option.Secret.Endpoints,
		Token:     s.option.Secret.Token,
	}
	s.secret = secret.NewSecretManager(opt)
	if err := s.secret.Init(); err != nil {
		blog.Errorf("manager init secret failure, %s", err.Error())
		return fmt.Errorf("secret failure")
	}
	s.stops = append(s.stops, s.secret.Stop)
	return nil
}

func (p *VaultPlugin) initVaultPluginHandler(apirouter *mux.Router) {
	h := handler.NewV1VaultPluginHandler(apirouter, handler.Options{
		Secret:      p.secret,
		GitopsStore: p.gitopsStorage,
	})
	h.Init()
}

func (p *VaultPlugin) initStorage() error {
	opt := &store.Options{
		Service: p.option.GitOps.Service,
		User:    p.option.GitOps.User,
		Pass:    p.option.GitOps.Pass,
		Cache:   false,
	}
	p.gitopsStorage = store.NewStore(opt)
	if err := p.gitopsStorage.Init(); err != nil {
		blog.Errorf("vaultplugin-server init gitops storage failure, %s", err.Error())
		return fmt.Errorf("gitops storage failure")
	}
	p.stops = append(p.stops, p.gitopsStorage.Stop)
	return nil
}
