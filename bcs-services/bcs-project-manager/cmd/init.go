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

// Package cmd xxx
package cmd

import (
	"context"
	"crypto/tls"
	"embed"
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/http/ipv6server"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/i18n"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/micro"
	microEtcd "github.com/go-micro/plugins/v4/registry/etcd"
	serverGrpc "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	microSvc "go-micro.dev/v4"
	microRgt "go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	grpcCred "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	i18n2 "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/i18n"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clientset"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/clustermanager"
	conf "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/etcd"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/manager"
	pmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/provider/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/runtimex"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/version"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/wrapper"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// ProjectService describe a project instance
type ProjectService struct {
	opt *conf.ProjectConfig

	// mongo DB options
	model store.ProjectModel
	// etcd options
	etcd *clientv3.Client

	microSvc         microSvc.Service
	microRgt         microRgt.Registry
	discovery        *discovery.ModuleDiscovery
	clusterDiscovery *discovery.ModuleDiscovery
	namespaceManager *manager.NamespaceManager

	// http service
	httpServer *ipv6server.IPv6Server

	// metric service
	metricServer *ipv6server.IPv6Server

	// tls config for server and client
	tlsConfig       *tls.Config
	clientTLSConfig *tls.Config

	ctx           context.Context
	ctxCancelFunc context.CancelFunc
	stopCh        chan struct{}
}

// newProjectSvc create a new project instance
func newProjectSvc(opt *conf.ProjectConfig) *ProjectService {
	ctx, cancel := context.WithCancel(context.Background())
	return &ProjectService{
		opt:           opt,
		ctx:           ctx,
		ctxCancelFunc: cancel,
		stopCh:        make(chan struct{}),
	}
}

// Init a project server
func (p *ProjectService) Init() error {
	for _, f := range []func() error{
		p.initTLSConfig,
		p.runTaskServer,
		p.initMongo,
		p.initCache,
		p.initEtcd,
		p.initRegistry,
		p.initDiscovery,
		p.initClientGroup,
		p.initJwtClient,
		p.initPermClient,
		p.initMicro,
		p.initHttpService,
		p.initNamespaceManager,
		p.initMetric,
		p.initI18n,
	} {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

// Run helm manager server
func (p *ProjectService) Run() error {
	// manage namespace scheduled task
	if p.opt.ITSM.Enable {
		go p.namespaceManager.Run()
	}
	// run the service
	if err := p.microSvc.Run(); err != nil {
		logging.Error("run micro service failed, err: %s", err.Error())
		return err
	}
	return nil
}

// initTLSConfig xxx
// init server and client tls config
func (p *ProjectService) initTLSConfig() error {
	if len(p.opt.Server.Cert) != 0 && len(p.opt.Server.Key) != 0 && len(p.opt.Server.Ca) != 0 {
		// 获取 cert paasword
		serverCertPwd := static.ServerCertPwd
		if p.opt.Server.CertPwd != "" {
			serverCertPwd = p.opt.Server.CertPwd
		}
		tlsConfig, err := ssl.ServerTslConfVerityClient(p.opt.Server.Ca, p.opt.Server.Cert,
			p.opt.Server.Key, serverCertPwd)
		if err != nil {
			logging.Error("load project server tls config failed, err %s", err.Error())
			return err
		}
		p.tlsConfig = tlsConfig
		logging.Info("load project server tls config successfully")
	}

	if len(p.opt.Client.Cert) != 0 && len(p.opt.Client.Key) != 0 && len(p.opt.Client.Ca) != 0 {
		// 获取 cert paasword
		clientCertPwd := static.ClientCertPwd
		if p.opt.Client.CertPwd != "" {
			clientCertPwd = p.opt.Client.CertPwd
		}
		tlsConfig, err := ssl.ClientTslConfVerity(p.opt.Client.Ca, p.opt.Client.Cert,
			p.opt.Client.Key, clientCertPwd)
		if err != nil {
			logging.Error("load project client tls config failed, err %s", err.Error())
			return err
		}
		p.clientTLSConfig = tlsConfig
		logging.Info("load project client tls config successfully")
	}
	return nil
}

// initMongo init mongo client
func (p *ProjectService) initMongo() error {
	store.InitMongo(&p.opt.Mongo)
	store.InitModel(store.GetMongo())
	p.model = store.GetModel()
	logging.Info("init mongo successfully")
	return nil
}

func (p *ProjectService) runTaskServer() error {
	_, err := pmanager.RunTaskManager()
	if err != nil {
		logging.Error("run task manager failed, err: %s", err.Error())
		return err
	}

	pmanager.RegisterQuotaMgrs()
	pmanager.RegisterValidateMgrs()

	return nil
}

// initCache init cache
func (p *ProjectService) initCache() error {
	cache.InitCache()
	return nil
}

// initEtcd init etcd client
func (p *ProjectService) initEtcd() error {

	err := etcd.Init(&p.opt.Etcd)
	if err != nil {
		logging.Info("init etcd client failed, err: %s", err.Error())
		return err
	}
	p.etcd, err = etcd.GetClient()
	if err != nil {
		logging.Info("init etcd client failed, err: %s", err.Error())
		return err
	}
	logging.Info("init etcd successfully")
	return nil
}

func (p *ProjectService) initRegistry() error {
	etcdEndpoints := stringx.SplitString(p.opt.Etcd.EtcdEndpoints)
	etcdSecure := false

	var etcdTLS *tls.Config
	var err error
	if len(p.opt.Etcd.EtcdCa) != 0 && len(p.opt.Etcd.EtcdCert) != 0 && len(p.opt.Etcd.EtcdKey) != 0 {
		etcdSecure = true
		etcdTLS, err = ssl.ClientTslConfVerity(p.opt.Etcd.EtcdCa, p.opt.Etcd.EtcdCert, p.opt.Etcd.EtcdKey, "")
		if err != nil {
			return err
		}
	}

	logging.Info("etcd endpoints for registry: %v, with secure %t", etcdEndpoints, etcdSecure)

	p.microRgt = microEtcd.NewRegistry(
		microRgt.Addrs(etcdEndpoints...),
		microRgt.Secure(etcdSecure),
		microRgt.TLSConfig(etcdTLS),
	)
	if err := p.microRgt.Init(); err != nil {
		logging.Error("register micro failed, err: %s", err.Error())
		return err
	}
	return nil
}

func (p *ProjectService) initDiscovery() error {
	p.discovery = discovery.NewModuleDiscovery(constant.ServiceDomain, p.microRgt)
	logging.Info("init discovery for project manager successfully")
	// enable discovery cluster manager module
	p.clusterDiscovery = discovery.NewModuleDiscovery(constant.ClusterManagerDomain, p.microRgt)
	clustermanager.SetClusterManagerClient(p.clientTLSConfig, p.clusterDiscovery)
	logging.Info("init discovery for cluster manager successfully")
	return nil
}

func (p *ProjectService) initClientGroup() error {
	logging.Info("init client group")
	clientset.SetClientGroup(p.opt.BcsGateway.Host, p.opt.BcsGateway.Token)
	return nil
}

func (p *ProjectService) initJwtClient() error {
	logging.Info("init jwt client")
	return auth.SetJwtClient()
}

func (p *ProjectService) initPermClient() error {
	logging.Info("init perm client")
	return auth.InitPermClient()
}

// initMicro init micro service
// NOCC:golint/fnsize(设计如此)
func (p *ProjectService) initMicro() error {

	// server listen ip
	ipv4 := p.opt.Server.Address
	ipv6 := p.opt.Server.Ipv6Address
	port := strconv.Itoa(p.opt.Server.Port)

	// service inject metadata to discovery center
	metadata := make(map[string]string)
	metadata[constant.MicroMetaKeyHTTPPort] = strconv.Itoa(p.opt.Server.HTTPPort)

	// 适配单栈环境（ipv6注册地址不能是本地回环地址）
	if v := net.ParseIP(ipv6); v != nil && !v.IsLoopback() {
		metadata[types.IPV6] = net.JoinHostPort(ipv6, port)
	}

	authWrapper := wrapper.NewAuthWrapper()

	// with tls
	server := serverGrpc.NewServer(
		serverGrpc.AuthTLS(p.tlsConfig),
		serverGrpc.MaxMsgSize(constant.MaxMsgSize),
	)

	svc := microSvc.NewService(
		microSvc.Server(server),
		microSvc.Cmd(util.NewDummyMicroCmd()),
		microSvc.Name(constant.ServiceDomain),
		microSvc.Metadata(metadata),
		microSvc.Address(net.JoinHostPort(ipv4, port)),
		microSvc.Registry(p.microRgt),
		microSvc.Version(version.Version),
		microSvc.RegisterTTL(30*time.Second),      // add ttl to config
		microSvc.RegisterInterval(25*time.Second), // add interval to config
		microSvc.Context(p.ctx),
		microSvc.AfterStart(func() error {
			if err := p.clusterDiscovery.Start(); err != nil {
				return err
			}
			return p.discovery.Start()
		}),
		microSvc.BeforeStop(func() error {
			p.clusterDiscovery.Stop()
			p.discovery.Stop()
			etcd.Close()
			return nil
		}),
		microSvc.AfterStop(func() error {
			// close audit client
			component.GetAuditClient().Close()
			return nil
		}),
		microSvc.WrapHandler(
			wrapper.NewAPILatencyWrapper,
			wrapper.NewInjectContextWrapper,
			wrapper.HandleLanguageWrapper,
			wrapper.NewResponseWrapper,
			wrapper.NewLogWrapper,
			wrapper.NewValidatorWrapper,
			wrapper.NewAuthHeaderAdapter,
			authWrapper.AuthenticationFunc,
			wrapper.NewAuthLogWrapper,
			authWrapper.AuthorizationFunc,
			wrapper.NewAuditWrapper,
			micro.NewTracingWrapper(),
		),
	)
	svc.Init()

	dualStackListener := listener.NewDualStackListener()
	if err := dualStackListener.AddListener(ipv4, port); err != nil {
		return err
	}
	if err := dualStackListener.AddListener(ipv6, port); err != nil {
		return err
	}
	// get grpc server
	grpcServer := svc.Server()
	// add dual stack listener to grpc server
	if err := grpcServer.Init(serverGrpc.Listener(dualStackListener)); err != nil {
		return err
	}

	// register grpc handlers
	if err := p.registerHandlers(grpcServer); err != nil {
		return err
	}

	p.microSvc = svc
	logging.Info("success to register project service handler to micro")
	return nil
}

func (p *ProjectService) registerHandlers(grpcServer server.Server) error {
	// 添加项目相关handler
	if err := proto.RegisterBCSProjectHandler(grpcServer, handler.NewProject(p.model)); err != nil {
		logging.Error("register project handler failed, err: %s", err.Error())
		return err
	}
	// 添加业务相关handler
	if err := proto.RegisterBusinessHandler(grpcServer, handler.NewBusiness(p.model)); err != nil {
		logging.Error("register business handler failed, err: %s", err.Error())
		return err
	}
	// 添加命名空间相关handler
	if err := proto.RegisterNamespaceHandler(grpcServer, handler.NewNamespace(p.model)); err != nil {
		logging.Error("register namespace handler failed, err: %s", err.Error())
		return err
	}
	// 添加变量相关的handler
	if err := proto.RegisterVariableHandler(grpcServer, handler.NewVariable(p.model)); err != nil {
		logging.Error("register variable handler failed, err: %s", err.Error())
		return err
	}
	// 添加健康检查相关handler
	if err := proto.RegisterHealthzHandler(grpcServer, handler.NewHealthz()); err != nil {
		logging.Error("register healthz handler failed, err: %s", err.Error())
		return err
	}
	// 添加项目额度相关handler
	if err := proto.RegisterBCSProjectQuotaHandler(grpcServer, handler.NewProjectQuota(p.model)); err != nil {
		logging.Error("register healthz handler failed, err: %s", err.Error())
		return err
	}

	return nil
}

// initHTTPGateway xxx
func (p *ProjectService) initHTTPGateway(router *mux.Router) error {
	gwMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(runtimex.CustomHeaderMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			OrigName:     true,
			EmitDefaults: true,
		}),
	)
	grpcDialOpts := []grpc.DialOption{}
	if p.tlsConfig != nil && p.clientTLSConfig != nil {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(grpcCred.NewTLS(p.clientTLSConfig)))
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	grpcDialOpts = append(grpcDialOpts, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(constant.MaxMsgSize), grpc.MaxCallSendMsgSize(constant.MaxMsgSize)))
	if err := p.registerGatewayFromEndPoint(gwMux, grpcDialOpts); err != nil {
		return err
	}
	router.Handle("/{uri:.*}", gwMux)
	logging.Info("register grpc gateway handler to path /")
	return nil
}

func (p *ProjectService) registerGatewayFromEndPoint(gwMux *runtime.ServeMux, grpcDialOpts []grpc.DialOption) error {
	// 注册项目功能 endpoint
	if err := proto.RegisterBCSProjectGwFromEndpoint(
		context.TODO(),
		gwMux,
		net.JoinHostPort(p.opt.Server.Address, strconv.Itoa(p.opt.Server.Port)),
		grpcDialOpts,
	); err != nil {
		logging.Error("register project endpoints to http gateway failed, err %s", err.Error())
		return err
	}
	// 注册业务功能 endpoint
	if err := proto.RegisterBusinessGwFromEndpoint(
		context.TODO(),
		gwMux,
		net.JoinHostPort(p.opt.Server.Address, strconv.Itoa(p.opt.Server.Port)),
		grpcDialOpts,
	); err != nil {
		logging.Error("register business endpoints to http gateway failed, err %s", err.Error())
		return err
	}
	// 注册命名空间相关 endpoint
	if err := proto.RegisterNamespaceGwFromEndpoint(
		context.TODO(),
		gwMux,
		net.JoinHostPort(p.opt.Server.Address, strconv.Itoa(p.opt.Server.Port)),
		grpcDialOpts,
	); err != nil {
		logging.Error("register namespace endpoints to gateway failed, err %s", err.Error())
		return err
	}
	// 注册变量相关 endpoint
	if err := proto.RegisterVariableGwFromEndpoint(
		context.TODO(),
		gwMux,
		net.JoinHostPort(p.opt.Server.Address, strconv.Itoa(p.opt.Server.Port)),
		grpcDialOpts,
	); err != nil {
		logging.Error("register variable endpoints to gateway failed, err %s", err.Error())
		return err
	}
	// 注册健康检查相关 endpoint
	if err := proto.RegisterHealthzGwFromEndpoint(
		context.TODO(),
		gwMux,
		net.JoinHostPort(p.opt.Server.Address, strconv.Itoa(p.opt.Server.Port)),
		grpcDialOpts,
	); err != nil {
		logging.Error("register healthz endpoints to gateway failed, err %s", err.Error())
		return err
	}
	// 注册额度管理相关 endpoint
	if err := proto.RegisterBCSProjectQuotaGwFromEndpoint(
		context.TODO(),
		gwMux,
		net.JoinHostPort(p.opt.Server.Address, strconv.Itoa(p.opt.Server.Port)),
		grpcDialOpts,
	); err != nil {
		logging.Error("register project quota endpoints to gateway failed, err %s", err.Error())
		return err
	}

	return nil
}

// initSwagger xxx
func (p *ProjectService) initSwagger(mux *http.ServeMux) {
	if p.opt.Swagger.Enable {
		logging.Info("swagger doc is enabled")
		mux.HandleFunc("/bcsproject/swagger/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path.Join(p.opt.Swagger.Dir, strings.TrimPrefix(r.URL.Path, "/bcsproject/swagger/")))
		})
	}
}

// initHttpService xxx
func (p *ProjectService) initHttpService() error {
	router := mux.NewRouter()

	// init micro http gateway
	if err := p.initHTTPGateway(router); err != nil {
		return err
	}

	mux := http.NewServeMux()
	// init swagger
	p.initSwagger(mux)
	mux.Handle("/", router)

	addresses := []string{p.opt.Server.Address}
	if len(p.opt.Server.Ipv6Address) > 0 {
		addresses = append(addresses, p.opt.Server.Ipv6Address)
	}
	p.httpServer = ipv6server.NewIPv6Server(addresses, strconv.Itoa(p.opt.Server.HTTPPort), "", mux)
	go func() {
		var err error
		logging.Info("start http server on address %+s", addresses)
		if p.tlsConfig != nil {
			p.httpServer.TLSConfig = p.tlsConfig
			err = p.httpServer.ListenAndServeTLS("", "")
		} else {
			err = p.httpServer.ListenAndServe()
		}
		if err != nil {
			logging.Error("start http server failed, err %s", err.Error())
			p.stopCh <- struct{}{}
		}
	}()
	return nil
}

func (p *ProjectService) initNamespaceManager() error {
	logging.Info("init namespace manager")
	p.namespaceManager = manager.NewNamespaceManager(p.ctx, p.model)
	return nil
}

func (p *ProjectService) initMetric() error {
	logging.Info("init metric handler")
	metricMux := http.NewServeMux()
	metricMux.Handle("/metrics", promhttp.Handler())
	addresses := []string{p.opt.Server.Address}
	if len(p.opt.Server.Ipv6Address) > 0 {
		addresses = append(addresses, p.opt.Server.Ipv6Address)
	}
	p.metricServer = ipv6server.NewIPv6Server(addresses, strconv.Itoa(p.opt.Server.MetricPort), "", metricMux)

	go func() {
		var err error
		logging.Info("start metric server on address %+v", addresses)
		if err = p.metricServer.ListenAndServe(); err != nil {
			logging.Error("start metric server failed, %s", err.Error())
			p.stopCh <- struct{}{}
		}
	}()
	return nil
}

// init i18n
func (p *ProjectService) initI18n() error {
	i18n.Instance()
	// 加载翻译文件路径
	i18n.SetPath([]embed.FS{i18n2.Assets})
	// 设置默认语言
	// 默认是 zh
	i18n.SetLanguage("zh")
	return nil
}
