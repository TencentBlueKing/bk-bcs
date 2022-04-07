/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"context"
	"crypto/tls"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/tunnel"

	gClient "github.com/asim/go-micro/plugins/client/grpc/v4"
	microEtcd "github.com/asim/go-micro/plugins/registry/etcd/v4"
	gServer "github.com/asim/go-micro/plugins/server/grpc/v4"
	"github.com/gorilla/mux"
	ggRuntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	microSvc "go-micro.dev/v4"
	microRgt "go-micro.dev/v4/registry"
)

// ArgocdProxy describe the main proxy server
type ArgocdProxy struct {
	opt *options.ProxyOptions

	microSvc  microSvc.Service
	microRgt  microRgt.Registry
	discovery *discovery.ModuleDiscovery

	// http service
	httpServer *http.Server

	// metric service
	metricServer *http.Server

	// tls config for helm manager service and client side
	tlsConfig       *tls.Config
	clientTLSConfig *tls.Config

	peerManager *tunnel.PeerManager

	ctx           context.Context
	ctxCancelFunc context.CancelFunc
	stopCh        chan struct{}
}

// NewArgocdProxy create a new ArgocdProxy
func NewArgocdProxy(opt *options.ProxyOptions) *ArgocdProxy {
	ctx, cancel := context.WithCancel(context.Background())
	return &ArgocdProxy{
		opt:           opt,
		ctx:           ctx,
		ctxCancelFunc: cancel,
		stopCh:        make(chan struct{}),
	}
}

// Init bcs argocd proxy
func (ap *ArgocdProxy) Init() error {
	for _, f := range []func() error{
		ap.initTLSConfig,
		ap.initRegistry,
		ap.initDiscovery,
		ap.initMicro,
		ap.initHTTPService,
	} {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

// Run bcs argocd proxy
func (ap *ArgocdProxy) Run() error {
	// run the service
	if err := ap.microSvc.Run(); err != nil {
		blog.Fatal(err)
	}
	blog.CloseLogs()
	return nil
}

// init server and client tls config
func (ap *ArgocdProxy) initTLSConfig() error {
	if len(ap.opt.ServerCert) != 0 && len(ap.opt.ServerKey) != 0 && len(ap.opt.ServerCa) != 0 {
		tlsConfig, err := ssl.ServerTslConfVerityClient(ap.opt.ServerCa, ap.opt.ServerCert,
			ap.opt.ServerKey, static.ServerCertPwd)
		if err != nil {
			blog.Errorf("load bcs argocd proxy tls config failed, err %s", err.Error())
			return err
		}
		ap.tlsConfig = tlsConfig
		blog.Infof("load bcs argocd proxy tls config successfully")
	}
	if len(ap.opt.ClientCert) != 0 && len(ap.opt.ClientKey) != 0 && len(ap.opt.ClientCa) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(ap.opt.ClientCa, ap.opt.ClientCert,
			ap.opt.ClientKey, static.ClientCertPwd)
		if err != nil {
			blog.Errorf("load bcs argocd proxy client tls config failed, err %s", err.Error())
			return err
		}
		ap.clientTLSConfig = tlsConfig
		blog.Infof("load bcs argocd proxy client tls config successfully")
	}
	return nil
}
func (ap *ArgocdProxy) initRegistry() error {
	etcdEndpoints := common.SplitAddrString(ap.opt.Etcd.EtcdEndpoints)
	etcdSecure := false
	var etcdTLS *tls.Config
	var err error
	if len(ap.opt.Etcd.EtcdCa) != 0 && len(ap.opt.Etcd.EtcdCert) != 0 && len(ap.opt.Etcd.EtcdKey) != 0 {
		etcdSecure = true
		etcdTLS, err = ssl.ClientTslConfVerity(ap.opt.Etcd.EtcdCa, ap.opt.Etcd.EtcdCert, ap.opt.Etcd.EtcdKey, "")
		if err != nil {
			return err
		}
	}
	blog.Infof("get etcd endpoints for registry: %v, with secure %t", etcdEndpoints, etcdSecure)
	ap.microRgt = microEtcd.NewRegistry(
		microRgt.Addrs(etcdEndpoints...),
		microRgt.Secure(etcdSecure),
		microRgt.TLSConfig(etcdTLS),
	)
	if err := ap.microRgt.Init(); err != nil {
		return err
	}
	return nil
}

func (ap *ArgocdProxy) initDiscovery() error {
	ap.discovery = discovery.NewModuleDiscovery(common.ServiceDomain, ap.microRgt)
	blog.Infof("init discovery for bcs argocd proxy successfully")
	return nil
}

func (ap *ArgocdProxy) initMicro() error {
	svc := microSvc.NewService(
		microSvc.Client(gClient.NewClient(gClient.AuthTLS(ap.tlsConfig))),
		microSvc.Server(gServer.NewServer(gServer.AuthTLS(ap.tlsConfig))),
		microSvc.Name(common.ServiceDomain),
		microSvc.Metadata(map[string]string{
			common.MicroMetaKeyHTTPPort: strconv.Itoa(int(ap.opt.HTTPPort)),
		}),
		microSvc.Address(ap.opt.Address+":"+strconv.Itoa(int(ap.opt.Port))),
		microSvc.Registry(ap.microRgt),
		microSvc.Version(version.BcsVersion),
		microSvc.RegisterTTL(30*time.Second),
		microSvc.RegisterInterval(25*time.Second),
		microSvc.Context(ap.ctx),
		microSvc.BeforeStart(func() error {
			return nil
		}),
		microSvc.AfterStart(func() error {
			return ap.discovery.Start()
		}),
		microSvc.BeforeStop(func() error {
			ap.discovery.Stop()
			return nil
		}),
	)

	ap.microSvc = svc
	blog.Infof("success to register bcs argocd proxy handlers to micro")
	return nil
}

func (ap *ArgocdProxy) initHTTPService() error {
	router := mux.NewRouter()

	if err := ap.initTunnelServer(router); err != nil {
		return err
	}

	if err := ap.initHTTPGateway(router); err != nil {
		return err
	}
	originMux := http.NewServeMux()
	originMux.Handle("/", router)
	if len(ap.opt.Swagger.Dir) != 0 {
		blog.Infof("swagger doc is enabled")
		originMux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path.Join(ap.opt.Swagger.Dir, strings.TrimPrefix(r.URL.Path, "/swagger/")))
		})
	}

	httpAddr := ap.opt.Address + ":" + strconv.Itoa(int(ap.opt.HTTPPort))
	ap.httpServer = &http.Server{
		Addr:    httpAddr,
		Handler: originMux,
	}

	if ap.tlsConfig != nil {
		ap.httpServer.TLSConfig = ap.tlsConfig

		// brings up another insecure port
		go ap.bringsUp(&http.Server{
			Addr:    ap.opt.InsecureAddress + ":" + strconv.Itoa(int(ap.opt.HTTPInsecurePort)),
			Handler: originMux,
		})
	}

	go ap.bringsUp(ap.httpServer)
	return nil
}

func (ap *ArgocdProxy) bringsUp(svr *http.Server) {
	blog.Infof("start http gateway server on address %s", svr.Addr)

	var err error
	if svr.TLSConfig != nil {
		err = svr.ListenAndServeTLS("", "")
	} else {
		err = svr.ListenAndServe()
	}

	if err != nil {
		blog.Errorf("start http gateway server failed, %s", err.Error())
		ap.stopCh <- struct{}{}
	}
}

func (ap *ArgocdProxy) initHTTPGateway(router *mux.Router) error {
	rmMux := ggRuntime.NewServeMux(
		ggRuntime.WithIncomingHeaderMatcher(CustomMatcher),
		ggRuntime.WithMarshalerOption(ggRuntime.MIMEWildcard, &ggRuntime.JSONPb{OrigName: true, EmitDefaults: true}),
		ggRuntime.WithDisablePathLengthFallback(),
	)

	router.Handle("/{uri:.*}", rmMux)
	return nil
}

const (
	subPathVarName = "sub_path"
	tunnelSvrUrl   = "/websocket/connect"
	proxyPassUrl   = "/{" + subPathVarName + ":argocdmanager/.*}"
)

func (ap *ArgocdProxy) initTunnelServer(router *mux.Router) error {
	callback := tunnel.NewWsTunnelServerCallback()
	ap.peerManager = tunnel.NewPeerManager(ap.opt, ap.clientTLSConfig, callback.GetTunnelServer(), ap.discovery)
	if err := ap.peerManager.Start(); err != nil {
		return err
	}
	dispatcher := NewWsTunnelDispatcher(subPathVarName, ap.opt, callback)

	// register the connecting handler
	router.Handle(tunnelSvrUrl, callback.GetTunnelServer())
	blog.Infof("register tunnel server handler to path %s", tunnelSvrUrl)

	// register other urls
	router.Handle(proxyPassUrl, dispatcher)
	blog.Infof("register dispatcher proxy pass server to path %s", proxyPassUrl)
	return nil
}

// CustomMatcher for http header
func CustomMatcher(key string) (string, bool) {
	switch key {
	case "X-Request-Id":
		return "X-Request-Id", true
	default:
		return ggRuntime.DefaultHeaderMatcher(key)
	}
}
