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

package cmd

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/tcp/listener"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	microEtcd "github.com/go-micro/plugins/v4/registry/etcd"
	microGrpc "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"google.golang.org/grpc"
	grpcCreds "google.golang.org/grpc/credentials"

	audit2 "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/audit"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/component/project"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	basicHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/basic"
	configHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/config"
	customResHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/customresource"
	hpaHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/hpa"
	multiHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/multicluster"
	nsHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/namespace"
	networkHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/network"
	nodeHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/node"
	rbacHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/rbac"
	resHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/resource"
	storageHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/storage"
	templateSetHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/templateset"
	viewHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/view"
	workloadHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/workload"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	httpUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/http"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/wrapper"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/swagger"
)

// clusterResourcesService
type clusterResourcesService struct {
	conf *config.ClusterResourcesConf

	ctx context.Context

	microSvc micro.Service
	microRtr registry.Registry

	httpServer   *http.Server
	metricServer *http.Server

	tlsConfig       *tls.Config
	clientTLSConfig *tls.Config

	model store.ClusterResourcesModel

	stopCh chan struct{}
}

// newClusterResourcesService 创建服务对象
func newClusterResourcesService(conf *config.ClusterResourcesConf) *clusterResourcesService {
	return &clusterResourcesService{conf: conf, ctx: context.TODO()}
}

// Init 服务初始化执行集
func (crSvc *clusterResourcesService) Init() error {
	// 各个初始化方法依次执行
	for _, f := range []func() error{
		crSvc.initTLSConfig,
		crSvc.initModel,
		crSvc.initRegistry,
		crSvc.initMicro,
		crSvc.initHandler,
		crSvc.initHTTPService,
		crSvc.initMetricService,
		crSvc.initComponentClient,
	} {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

// Run 服务启动逻辑
func (crSvc *clusterResourcesService) Run() error {
	if err := crSvc.microSvc.Run(); err != nil {
		return err
	}
	return nil
}

// initMicro 初始化 MicroService
func (crSvc *clusterResourcesService) initMicro() error {
	metadata := map[string]string{}
	dualStackListener := listener.NewDualStackListener()

	grpcPort := strconv.Itoa(crSvc.conf.Server.Port)
	grpcAddr := net.JoinHostPort(crSvc.conf.Server.Address, grpcPort)

	if err := dualStackListener.AddListenerWithAddr(grpcAddr); err != nil {
		return err
	}

	if crSvc.conf.Server.AddressIPv6 != "" {
		ipv6Addr := net.JoinHostPort(crSvc.conf.Server.AddressIPv6, grpcPort)
		metadata[types.IPV6] = ipv6Addr

		if err := dualStackListener.AddListenerWithAddr(ipv6Addr); err != nil {
			return err
		}
		log.Info(crSvc.ctx, "grpc serve dualStackListener with ipv6: %s", ipv6Addr)
	}

	// init micro server
	grpcServer := microGrpc.NewServer(
		server.Name(conf.ServiceDomain),
		microGrpc.AuthTLS(crSvc.tlsConfig),
		microGrpc.MaxMsgSize(conf.MaxGrpcMsgSize),
		microGrpc.Listener(dualStackListener),
		server.Address(grpcAddr),
		server.Registry(crSvc.microRtr),
		server.RegisterTTL(time.Duration(crSvc.conf.Server.RegisterTTL)*time.Second),
		server.RegisterInterval(time.Duration(crSvc.conf.Server.RegisterInterval)*time.Second),
		server.Version(version.Version),

		server.WrapHandler(
			//	链路追踪
			wrapper.NewTracingWrapper(),
		),
		server.WrapHandler(
			// context 信息注入
			wrapper.NewContextInjectWrapper(),
		),
		server.WrapHandler(
			// 格式化返回结果
			wrapper.NewResponseFormatWrapper(),
		),
		server.WrapHandler(
			// 记录 API 访问流水日志
			wrapper.NewLogWrapper(),
		),
		server.WrapHandler(
			// 自动执行参数校验
			wrapper.NewValidatorHandlerWrapper(),
		),
	)
	if err := grpcServer.Init(); err != nil {
		return err
	}

	crSvc.microSvc = micro.NewService(micro.AfterStop(func() error {
		audit2.GetAuditClient().Close()
		return nil
	}), micro.Server(grpcServer), micro.Metadata(metadata))
	log.Info(crSvc.ctx, "register cluster resources handler to micro successfully.")
	return nil
}

// initHandler 注册多个 Handler
func (crSvc *clusterResourcesService) initHandler() error { // nolint:cyclop
	if err := clusterRes.RegisterBasicHandler(crSvc.microSvc.Server(), basicHdlr.New()); err != nil {
		return err
	}
	if err := clusterRes.RegisterNodeHandler(crSvc.microSvc.Server(), nodeHdlr.New()); err != nil {
		return err
	}
	if err := clusterRes.RegisterNamespaceHandler(crSvc.microSvc.Server(), nsHdlr.New()); err != nil {
		return err
	}
	if err := clusterRes.RegisterWorkloadHandler(crSvc.microSvc.Server(), workloadHdlr.New()); err != nil {
		return err
	}
	if err := clusterRes.RegisterNetworkHandler(crSvc.microSvc.Server(), networkHdlr.New()); err != nil {
		return err
	}
	if err := clusterRes.RegisterConfigHandler(crSvc.microSvc.Server(), configHdlr.New()); err != nil {
		return err
	}
	if err := clusterRes.RegisterStorageHandler(crSvc.microSvc.Server(), storageHdlr.New()); err != nil {
		return err
	}
	if err := clusterRes.RegisterRBACHandler(crSvc.microSvc.Server(), rbacHdlr.New()); err != nil {
		return err
	}
	if err := clusterRes.RegisterHPAHandler(crSvc.microSvc.Server(), hpaHdlr.New()); err != nil {
		return err
	}
	if err := clusterRes.RegisterCustomResHandler(crSvc.microSvc.Server(), customResHdlr.New()); err != nil {
		return err
	}
	if err := clusterRes.RegisterResourceHandler(crSvc.microSvc.Server(), resHdlr.New()); err != nil {
		return err
	}
	if err := clusterRes.RegisterViewConfigHandler(crSvc.microSvc.Server(), viewHdlr.New(crSvc.model)); err != nil {
		return err
	}
	if err := clusterRes.RegisterTemplateSetHandler(
		crSvc.microSvc.Server(), templateSetHdlr.New(crSvc.model)); err != nil {
		return err
	}
	if err := clusterRes.RegisterMultiClusterHandler(crSvc.microSvc.Server(), multiHdlr.New(crSvc.model)); err != nil {
		return err
	}
	return nil
}

// initRegistry 注册服务到 Etcd
func (crSvc *clusterResourcesService) initRegistry() error {
	etcdEndpoints := stringx.Split(crSvc.conf.Etcd.EtcdEndpoints)
	etcdSecure := false

	var etcdTLS *tls.Config
	var err error
	if crSvc.conf.Etcd.EtcdCa != "" && crSvc.conf.Etcd.EtcdCert != "" && crSvc.conf.Etcd.EtcdKey != "" {
		etcdSecure = true
		etcdTLS, err = ssl.ClientTslConfVerity(
			crSvc.conf.Etcd.EtcdCa, crSvc.conf.Etcd.EtcdCert, crSvc.conf.Etcd.EtcdKey, "",
		)
		if err != nil {
			return err
		}
	}

	log.Info(crSvc.ctx, "registry: etcd endpoints: %v, secure: %t", etcdEndpoints, etcdSecure)

	crSvc.microRtr = microEtcd.NewRegistry(
		registry.Addrs(etcdEndpoints...),
		registry.Secure(etcdSecure),
		registry.TLSConfig(etcdTLS),
	)
	if err = crSvc.microRtr.Init(); err != nil {
		return err
	}
	return nil
}

// initTLSConfig 初始化 Server 与 client TLS 配置
func (crSvc *clusterResourcesService) initTLSConfig() error {
	if crSvc.conf.Server.Cert != "" && crSvc.conf.Server.Key != "" && crSvc.conf.Server.Ca != "" {
		tlsConfig, err := ssl.ServerTslConfVerityClient(
			crSvc.conf.Server.Ca, crSvc.conf.Server.Cert, crSvc.conf.Server.Key, crSvc.conf.Server.CertPwd,
		)
		if err != nil {
			log.Error(crSvc.ctx, "load cluster resources server tls config failed: %v", err)
			return err
		}
		crSvc.tlsConfig = tlsConfig
		log.Info(crSvc.ctx, "load cluster resources server tls config successfully")
	}

	if crSvc.conf.Client.Cert != "" && crSvc.conf.Client.Key != "" && crSvc.conf.Client.Ca != "" {
		tlsConfig, err := ssl.ClientTslConfVerity(
			crSvc.conf.Client.Ca, crSvc.conf.Client.Cert, crSvc.conf.Client.Key, crSvc.conf.Client.CertPwd,
		)
		if err != nil {
			log.Error(crSvc.ctx, "load cluster resources client tls config failed: %v", err)
			return err
		}
		crSvc.clientTLSConfig = tlsConfig
		log.Info(crSvc.ctx, "load cluster resources client tls config successfully")
	}
	return nil
}

// initHTTPService 初始化 HTTP 服务
func (crSvc *clusterResourcesService) initHTTPService() error {
	rmMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(httpUtil.CustomHeaderMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}),
	)

	grpcDialOpts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(conf.MaxGrpcMsgSize),
			grpc.MaxCallSendMsgSize(conf.MaxGrpcMsgSize),
		),
	}
	if crSvc.tlsConfig != nil && crSvc.clientTLSConfig != nil {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(grpcCreds.NewTLS(crSvc.clientTLSConfig)))
	} else {
		// nolint
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
	}

	// 循环注册各个 rpc service
	endpoint := net.JoinHostPort(crSvc.conf.Server.Address, strconv.Itoa(crSvc.conf.Server.Port))
	for _, epRegister := range []func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error{
		clusterRes.RegisterBasicGwFromEndpoint,
		clusterRes.RegisterNodeGwFromEndpoint, clusterRes.RegisterNamespaceGwFromEndpoint,
		clusterRes.RegisterWorkloadGwFromEndpoint, clusterRes.RegisterNetworkGwFromEndpoint,
		clusterRes.RegisterConfigGwFromEndpoint, clusterRes.RegisterStorageGwFromEndpoint,
		clusterRes.RegisterRBACGwFromEndpoint, clusterRes.RegisterHPAGwFromEndpoint,
		clusterRes.RegisterCustomResGwFromEndpoint, clusterRes.RegisterResourceGwFromEndpoint,
		clusterRes.RegisterViewConfigGwFromEndpoint, clusterRes.RegisterMultiClusterGwFromEndpoint,
		clusterRes.RegisterTemplateSetGwFromEndpoint,
	} {
		err := epRegister(crSvc.ctx, rmMux, endpoint, grpcDialOpts)
		if err != nil {
			log.Error(crSvc.ctx, "register http service failed: %v", err)
			return errorx.New(errcode.General, "register http service failed: %v", err)
		}
	}

	router := mux.NewRouter()
	router.Handle("/{uri:.*}", rmMux)
	log.Info(crSvc.ctx, "register grpc service handler to path /")
	originMux := http.NewServeMux()
	originMux.Handle("/", router)
	originMux.Handle("/clusterresources/api/v1/", NewAPIRouter(crSvc))

	// 检查是否需要启用 swagger 服务
	if crSvc.conf.Swagger.Enabled {
		log.Info(crSvc.ctx, "swagger doc is enabled")
		// 加载 swagger.json
		// 配置 swagger-ui 服务
		originMux.HandleFunc("/clusterresources/swagger/", handlerSwagger)
	}

	httpPort := strconv.Itoa(crSvc.conf.Server.HTTPPort)
	httpAddr := net.JoinHostPort(crSvc.conf.Server.Address, httpPort)

	crSvc.httpServer = &http.Server{
		Addr: httpAddr,
		Handler: wsproxy.WebsocketProxy(
			originMux,
			wsproxy.WithForwardedHeaders(httpUtil.WSHeaderForwarder),
		),
	}

	dualStackListener := listener.NewDualStackListener()
	if err := dualStackListener.AddListenerWithAddr(httpAddr); err != nil {
		return err
	}
	if crSvc.conf.Server.AddressIPv6 != "" {
		ipv6Addr := net.JoinHostPort(crSvc.conf.Server.AddressIPv6, httpPort)
		if err := dualStackListener.AddListenerWithAddr(ipv6Addr); err != nil {
			return err
		}
		log.Info(crSvc.ctx, "http serve dualStackListener with ipv6: %s", ipv6Addr)
	}

	go func() {
		crSvc.run(httpAddr, dualStackListener)
	}()
	return nil
}

func handlerSwagger(w http.ResponseWriter, r *http.Request) {
	// 提取 URL 的扩展名
	ext := filepath.Ext(r.URL.Path)
	// 检查扩展名是否为 ".json"
	if ext == ".json" {
		// 设置响应头
		w.Header().Set("Content-Type", "application/json")
		file, _ := swagger.Assets.ReadFile("data/cluster-resources.swagger.json")
		w.Write(file)
		return
	}
	httpSwagger.Handler(httpSwagger.URL("cluster-resources.swagger.json")).ServeHTTP(w, r)
}

func (crSvc *clusterResourcesService) run(httpAddr string, dualStackListener net.Listener) {
	var err error
	log.Info(crSvc.ctx, "start http gateway server on address %s", httpAddr)
	if crSvc.tlsConfig != nil {
		crSvc.httpServer.TLSConfig = crSvc.tlsConfig
		err = crSvc.httpServer.ServeTLS(dualStackListener, "", "")
	} else {
		err = crSvc.httpServer.Serve(dualStackListener)
	}
	if err != nil {
		log.Error(crSvc.ctx, "start http gateway server failed: %v", err)
		crSvc.stopCh <- struct{}{}
	}
}

// initMetricService 初始化 Metric 服务
func (crSvc *clusterResourcesService) initMetricService() error {
	log.Info(crSvc.ctx, "init cluster resource metric service")

	metricMux := http.NewServeMux()
	metricMux.Handle("/metrics", promhttp.Handler())

	metricPort := strconv.Itoa(crSvc.conf.Server.MetricPort)
	metricAddr := net.JoinHostPort(crSvc.conf.Server.Address, metricPort)
	crSvc.metricServer = &http.Server{
		Addr:    metricAddr,
		Handler: metricMux,
	}

	dualStackListener := listener.NewDualStackListener()
	if err := dualStackListener.AddListenerWithAddr(metricAddr); err != nil {
		return err
	}
	if crSvc.conf.Server.AddressIPv6 != "" {
		ipv6Addr := net.JoinHostPort(crSvc.conf.Server.AddressIPv6, metricPort)
		if err := dualStackListener.AddListenerWithAddr(ipv6Addr); err != nil {
			return err
		}
		log.Info(crSvc.ctx, "metric serve dualStackListener with ipv6: %s", ipv6Addr)
	}

	go func() {
		var err error
		log.Info(crSvc.ctx, "start metric server on address %s", metricAddr)
		if err = crSvc.metricServer.Serve(dualStackListener); err != nil {
			log.Error(crSvc.ctx, "start metric server failed: %v", err)
			crSvc.stopCh <- struct{}{}
		}
	}()
	return nil
}

// initComponentClient 初始化依赖组件 Client
func (crSvc *clusterResourcesService) initComponentClient() (err error) {
	// ClusterManager
	cluster.InitCMClient()
	// ProjectManager
	project.InitProjClient()
	return nil
}

// initModel decode the connection info from the config and init a new store.HelmManagerModel
func (crSvc *clusterResourcesService) initModel() error {
	if len(crSvc.conf.Mongo.Address) == 0 {
		log.Error(crSvc.ctx, "mongo address is empty")
		return nil
	}
	if len(crSvc.conf.Mongo.Database) == 0 {
		log.Error(crSvc.ctx, "mongo database is empty")
		return nil
	}
	password := crSvc.conf.Mongo.Password
	if password != "" && crSvc.conf.Mongo.Encrypted {
		realPwd, err := encrypt.DesDecryptFromBase([]byte(password))
		if err != nil {
			log.Error(crSvc.ctx, "decrypt password failed, err %s", err.Error())
			return nil
		}

		password = string(realPwd)
	}
	mongoOptions := &mongo.Options{
		Hosts:                 strings.Split(crSvc.conf.Mongo.Address, ","),
		ConnectTimeoutSeconds: int(crSvc.conf.Mongo.ConnectTimeout),
		AuthDatabase:          crSvc.conf.Mongo.AuthDatabase,
		Database:              crSvc.conf.Mongo.Database,
		Username:              crSvc.conf.Mongo.Username,
		Password:              password,
		MaxPoolSize:           uint64(crSvc.conf.Mongo.MaxPoolSize),
		MinPoolSize:           uint64(crSvc.conf.Mongo.MinPoolSize),
		Monitor:               otelmongo.NewMonitor(),
	}

	mongoDB, err := mongo.NewDB(mongoOptions)
	if err != nil {
		log.Error(crSvc.ctx, "init mongo db failed, err %s", err.Error())
		return nil
	}
	if err = mongoDB.Ping(); err != nil {
		log.Error(crSvc.ctx, "ping mongo db failed, err %s", err.Error())
		return nil
	}
	log.Info(crSvc.ctx, "init mongo db successfully")
	crSvc.model = store.New(mongoDB)
	log.Info(crSvc.ctx, "init store successfully")
	return nil
}
