/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"context"
	"crypto/tls"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	microEtcd "github.com/asim/go-micro/plugins/registry/etcd/v4"
	microGrpc "github.com/asim/go-micro/plugins/server/grpc/v4"
	goBindataAssetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"
	"google.golang.org/grpc"
	grpcCreds "google.golang.org/grpc/credentials"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	basicHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/basic"
	configHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/config"
	customResHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/customresource"
	hpaHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/hpa"
	nsHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/namespace"
	networkHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/network"
	rbacHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/rbac"
	resHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/resource"
	storageHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/storage"
	workloadHdlr "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler/workload"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
	httpUtil "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/http"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/version"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/wrapper"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/swagger"
)

type clusterResourcesService struct {
	conf *config.ClusterResourcesConf

	microSvc micro.Service
	microRtr registry.Registry

	httpServer   *http.Server
	metricServer *http.Server

	tlsConfig       *tls.Config
	clientTLSConfig *tls.Config

	stopCh chan struct{}
}

// newClusterResourcesService 创建服务对象
func newClusterResourcesService(conf *config.ClusterResourcesConf) *clusterResourcesService {
	return &clusterResourcesService{conf: conf}
}

// Init 服务初始化执行集
func (crSvc *clusterResourcesService) Init() error {
	// 各个初始化方法依次执行
	for _, f := range []func() error{
		crSvc.initTLSConfig,
		crSvc.initRegistry,
		crSvc.initMicro,
		crSvc.initHandler,
		crSvc.initHTTPService,
		crSvc.initMetricService,
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

// 初始化 MicroService
func (crSvc *clusterResourcesService) initMicro() error {
	grpcServer := microGrpc.NewServer(
		server.Name(conf.ServiceDomain),
		microGrpc.AuthTLS(crSvc.tlsConfig),
		server.Address(crSvc.conf.Server.Address+":"+strconv.Itoa(crSvc.conf.Server.Port)),
		server.Registry(crSvc.microRtr),
		server.RegisterTTL(time.Duration(crSvc.conf.Server.RegisterTTL)*time.Second),
		server.RegisterInterval(time.Duration(crSvc.conf.Server.RegisterInterval)*time.Second),
		server.Version(version.Version),
		server.WrapHandler(
			// context 信息注入
			wrapper.NewContextInjectWrapper(),
		),
		server.WrapHandler(
			// 格式化返回结果
			wrapper.NewResponseFormatWrapper(),
		),
		server.WrapHandler(
			// 自动执行参数校验
			wrapper.NewValidatorHandlerWrapper(),
		),
	)
	if err := grpcServer.Init(); err != nil {
		return err
	}

	crSvc.microSvc = micro.NewService(micro.Server(grpcServer))
	log.Info("register cluster resources handler to micro successfully.")
	return nil
}

// 注册多个 Handler
func (crSvc *clusterResourcesService) initHandler() error { // nolint:cyclop
	if err := clusterRes.RegisterBasicHandler(crSvc.microSvc.Server(), basicHdlr.New()); err != nil {
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
	return nil
}

// 注册服务到 Etcd
func (crSvc *clusterResourcesService) initRegistry() error {
	etcdEndpoints := stringx.Split(crSvc.conf.Etcd.EtcdEndpoints)
	etcdSecure := false

	var etcdTLS *tls.Config
	var err error
	if len(crSvc.conf.Etcd.EtcdCa) != 0 && len(crSvc.conf.Etcd.EtcdCert) != 0 && len(crSvc.conf.Etcd.EtcdKey) != 0 {
		etcdSecure = true
		etcdTLS, err = ssl.ClientTslConfVerity(
			crSvc.conf.Etcd.EtcdCa, crSvc.conf.Etcd.EtcdCert, crSvc.conf.Etcd.EtcdKey, "",
		)
		if err != nil {
			return err
		}
	}

	log.Info("registry: etcd endpoints: %v, secure: %t", etcdEndpoints, etcdSecure)

	crSvc.microRtr = microEtcd.NewRegistry(
		registry.Addrs(etcdEndpoints...),
		registry.Secure(etcdSecure),
		registry.TLSConfig(etcdTLS),
	)
	if err := crSvc.microRtr.Init(); err != nil {
		return err
	}
	return nil
}

// 初始化 Server 与 client TLS 配置
func (crSvc *clusterResourcesService) initTLSConfig() error {
	if len(crSvc.conf.Server.Cert) != 0 && len(crSvc.conf.Server.Key) != 0 && len(crSvc.conf.Server.Ca) != 0 {
		tlsConfig, err := ssl.ServerTslConfVerityClient(
			crSvc.conf.Server.Ca, crSvc.conf.Server.Cert, crSvc.conf.Server.Key, crSvc.conf.Server.CertPwd,
		)
		if err != nil {
			log.Error("load cluster resources server tls config failed: %v", err)
			return err
		}
		crSvc.tlsConfig = tlsConfig
		log.Info("load cluster resources server tls config successfully")
	}

	if len(crSvc.conf.Client.Cert) != 0 && len(crSvc.conf.Client.Key) != 0 && len(crSvc.conf.Client.Ca) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(
			crSvc.conf.Client.Ca, crSvc.conf.Client.Cert, crSvc.conf.Client.Key, crSvc.conf.Client.CertPwd,
		)
		if err != nil {
			log.Error("load cluster resources client tls config failed: %v", err)
			return err
		}
		crSvc.clientTLSConfig = tlsConfig
		log.Info("load cluster resources client tls config successfully")
	}
	return nil
}

// 初始化 HTTP 服务
func (crSvc *clusterResourcesService) initHTTPService() error {
	rmMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(httpUtil.CustomHeaderMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}),
	)

	grpcDialconf := make([]grpc.DialOption, 0)
	if crSvc.tlsConfig != nil && crSvc.clientTLSConfig != nil {
		grpcDialconf = append(grpcDialconf, grpc.WithTransportCredentials(grpcCreds.NewTLS(crSvc.clientTLSConfig)))
	} else {
		grpcDialconf = append(grpcDialconf, grpc.WithInsecure())
	}

	// 循环注册各个 rpc service
	ctx, endpoint := context.TODO(), crSvc.conf.Server.Address+":"+strconv.Itoa(crSvc.conf.Server.Port)
	for _, epRegister := range []func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error{
		clusterRes.RegisterBasicGwFromEndpoint,
		clusterRes.RegisterNamespaceGwFromEndpoint,
		clusterRes.RegisterWorkloadGwFromEndpoint,
		clusterRes.RegisterNetworkGwFromEndpoint,
		clusterRes.RegisterConfigGwFromEndpoint,
		clusterRes.RegisterStorageGwFromEndpoint,
		clusterRes.RegisterRBACGwFromEndpoint,
		clusterRes.RegisterHPAGwFromEndpoint,
		clusterRes.RegisterCustomResGwFromEndpoint,
		clusterRes.RegisterResourceGwFromEndpoint,
	} {
		err := epRegister(ctx, rmMux, endpoint, grpcDialconf)
		if err != nil {
			log.Error("register http service failed: %v", err)
			return errorx.New(errcode.General, "register http service failed: %v", err)
		}
	}

	router := mux.NewRouter()
	router.Handle("/{uri:.*}", rmMux)
	log.Info("register grpc service handler to path /")

	originMux := http.NewServeMux()
	originMux.Handle("/", router)

	// 检查是否需要启用 swagger 服务
	if crSvc.conf.Swagger.Enabled && len(crSvc.conf.Swagger.Dir) != 0 {
		log.Info("swagger doc is enabled")
		// 挂载 swagger.json 文件目录
		originMux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path.Join(crSvc.conf.Swagger.Dir, strings.TrimPrefix(r.URL.Path, "/swagger/")))
		})
		// 配置 swagger-ui 服务
		fileServer := http.FileServer(&goBindataAssetfs.AssetFS{
			Asset:    swagger.Asset,
			AssetDir: swagger.AssetDir,
			Prefix:   "third_party/swagger-ui",
		})
		originMux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", fileServer))
	}

	httpAddr := crSvc.conf.Server.Address + ":" + strconv.Itoa(crSvc.conf.Server.HTTPPort)
	crSvc.httpServer = &http.Server{
		Addr:    httpAddr,
		Handler: wsproxy.WebsocketProxy(originMux),
	}
	go func() {
		var err error
		log.Info("start http gateway server on address %s", httpAddr)
		if crSvc.tlsConfig != nil {
			crSvc.httpServer.TLSConfig = crSvc.tlsConfig
			err = crSvc.httpServer.ListenAndServeTLS("", "")
		} else {
			err = crSvc.httpServer.ListenAndServe()
		}
		if err != nil {
			log.Error("start http gateway server failed: %v", err)
			crSvc.stopCh <- struct{}{}
		}
	}()
	return nil
}

// 初始化 Metric 服务
func (crSvc *clusterResourcesService) initMetricService() error {
	log.Info("init cluster resource metric service")

	metricMux := http.NewServeMux()
	metricMux.Handle("/metrics", promhttp.Handler())

	metricAddr := crSvc.conf.Server.Address + ":" + strconv.Itoa(crSvc.conf.Server.MetricPort)
	crSvc.metricServer = &http.Server{
		Addr:    metricAddr,
		Handler: metricMux,
	}

	go func() {
		var err error
		log.Info("start metric server on address %s", metricAddr)
		if err = crSvc.metricServer.ListenAndServe(); err != nil {
			log.Error("start metric server failed: %v", err)
			crSvc.stopCh <- struct{}{}
		}
	}()
	return nil
}
