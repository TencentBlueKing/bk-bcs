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
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	grpccli "github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	grpcsvr "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/gorilla/mux"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/tunnel"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
)

// NewProxy create new proxy instance
func NewProxy(opt *Options) *Proxy {
	cxt, cancel := context.WithCancel(context.Background())
	return &Proxy{
		cxt:    cxt,
		cancel: cancel,
		stops:  make([]utils.StopFunc, 0),
		option: opt,
	}
}

// Proxy implementation
type Proxy struct {
	cxt    context.Context
	cancel context.CancelFunc
	stops  []utils.StopFunc
	option *Options
	// http service serve for grpc-http gateway, manager proxy and websocket tunnel
	httpService *http.Server
	// micro service use for service discovery
	microService  micro.Service
	tunnelManager *tunnel.TunnelManager
}

// Init all service
func (p *Proxy) Init() error {
	initializer := []func() error{
		p.initMicroService, p.initTunnelManager, p.initHTTPService,
	}

	for _, initFunc := range initializer {
		if err := initFunc(); err != nil {
			return err
		}
	}
	return nil
}

// Run proxy server
func (p *Proxy) Run() error {
	runners := []func(){
		p.startMicroService, p.startTunnelManager,
		p.startSignalHandler, p.startHTTPService,
	}

	for _, run := range runners {
		time.Sleep(time.Second)
		go run()
	}

	<-p.cxt.Done()
	p.stop()

	blog.Infof("proxy is under graceful period, %d seconds...", gracefulexit)
	time.Sleep(time.Second * gracefulexit)
	return nil
}

// stop all services
func (p *Proxy) stop() {
	for _, stop := range p.stops {
		go stop()
	}
}

func (p *Proxy) initMicroService() error {
	svc := micro.NewService(
		micro.Client(grpccli.NewClient(grpccli.AuthTLS(p.option.ClientTLS))),
		micro.Server(grpcsvr.NewServer(grpcsvr.AuthTLS(p.option.ServerTLS))),
		micro.Name(common.ProxyName),
		micro.Metadata(map[string]string{
			common.MetaHTTPKey: fmt.Sprintf("%d", p.option.HTTPPort),
		}),
		micro.Address(fmt.Sprintf("%s:%d", p.option.Address, p.option.Port)),
		micro.Version(version.BcsVersion),
		micro.RegisterTTL(30*time.Second),
		micro.RegisterInterval(25*time.Second),
		micro.Context(p.cxt),
		micro.Registry(etcd.NewRegistry(
			registry.Addrs(strings.Split(p.option.Registry.Endpoints, ",")...),
			registry.TLSConfig(p.option.Registry.TLSConfig),
		)),
	)
	p.microService = svc
	blog.Infof("proxy init go-micro service successfully")
	return nil
}

func (p *Proxy) startMicroService() {
	err := p.microService.Run()
	if err != nil {
		blog.Fatalf("proxy start micro service failed, %s", err.Error())
		return
	}
	blog.Infof("proxy start micro service successfully")
}

func (p *Proxy) initTunnelManager() error {
	ctx, cancel := context.WithCancel(p.cxt) // nolint
	option := &tunnel.TunnelOptions{
		Context:           ctx,
		TunnelID:          fmt.Sprintf("%s:%d", p.option.Address, p.option.HTTPPort),
		TunnelToken:       p.option.PeerConnectToken,
		ClusterAddressKey: common.HeaderServerAddressKey,
		ConnectURL:        p.option.PeerConnectURL,
		PeerServiceName:   p.option.ServiceName,
		ClientTLS:         p.option.ClientTLS,
		// !micro service must init before tunnelMgr
		Registry: p.microService.Options().Registry,
		// proxy don't serve mutiple cluster,
		// just return gitopsmanager service
		Indexer: func(_ *http.Request) (string, error) {
			return common.ServiceName, nil
		},
	}
	p.tunnelManager = tunnel.NewTunnelManager(option)
	// register tunnel manager stop function
	p.stops = append(p.stops, utils.StopFunc(cancel))
	if err := p.tunnelManager.Init(); err != nil {
		return err
	}
	return nil
}

func (p *Proxy) startTunnelManager() {
	if err := p.tunnelManager.Start(); err != nil {
		blog.Fatalf("proxy start tunnel manager failed, %s", err.Error())
		return
	}
	blog.Infof("proxy start tunnel manager successfully")
}

func (p *Proxy) initHTTPService() error {
	router := mux.NewRouter()
	router.UseEncodedPath()
	// init websocket connect, pass it to tunnel server directly
	// URL: /gitopsproxy/websocket/connect
	router.Handle(common.ConnectURI, p.tunnelManager.GetTunnelServer())

	// init gitopsmanager proxy, there are two URLs for proxy
	// /gitopsmanager/v1/ for gitops services
	// /gitopsmanager/proxy/ for argocd proxy
	// but actually, proxy don't care subpath details
	router.PathPrefix("/gitopsmanager/").Handler(p.tunnelManager)

	// !!importance fix: golang strim %2f%2f to / in URL path
	bugWork := &proxy.BUG21955Workaround{Handler: router}

	// init http server
	p.httpService = &http.Server{
		Addr:      fmt.Sprintf("%s:%d", p.option.Address, p.option.HTTPPort),
		Handler:   bugWork,
		TLSConfig: p.option.ServerTLS,
	}
	return nil
}

func (p *Proxy) startHTTPService() {
	if p.httpService == nil {
		blog.Fatalf("proxy lost http server instance")
		return
	}
	p.stops = append(p.stops, p.stopHTTPService)
	err := p.httpService.ListenAndServeTLS("", "")
	if err != nil {
		if http.ErrServerClosed == err {
			blog.Warnf("proxy http service gracefully exit.")
			return
		}
		// start http gateway error, maybe port is conflict or something else
		blog.Fatal("proxy http service ListenAndServeTLS fatal, %s", err.Error())
	}
}

func (p *Proxy) startSignalHandler() {
	utils.StartSignalHandler(p.cancel, gracefulexit)
}

// stopHTTPService  gracefully stop
func (p *Proxy) stopHTTPService() {
	cxt, cancel := context.WithTimeout(p.cxt, time.Second*2)
	defer cancel()
	if err := p.httpService.Shutdown(cxt); err != nil {
		blog.Errorf("proxy gracefully shutdown http service failure: %s", err.Error())
		return
	}
	blog.Infof("proxy shutdown http service gracefully")
}
