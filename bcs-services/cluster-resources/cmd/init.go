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

/*
 * init.go ClusterResources 模块初始化相关
 */

package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	goBindataAssetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	microRgt "github.com/micro/go-micro/v2/registry"
	microEtcd "github.com/micro/go-micro/v2/registry/etcd"
	microSvc "github.com/micro/go-micro/v2/service"
	microGrpc "github.com/micro/go-micro/v2/service/grpc"
	"google.golang.org/grpc"
	grpcCreds "google.golang.org/grpc/credentials"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/handler"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/utils"
	clusterRes "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/proto/cluster-resources"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/swagger"
)

type clusterResourcesService struct {
	conf *config.ClusterResourcesConf

	microSvc microSvc.Service
	microRtr microRgt.Registry

	httpServer *http.Server

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
		crSvc.initHTTPService,
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
	svc := microGrpc.NewService(
		microSvc.Name(common.ServiceDomain),
		microGrpc.WithTLS(crSvc.tlsConfig),
		microSvc.Address(crSvc.conf.Server.Address+":"+strconv.Itoa(crSvc.conf.Server.Port)),
		microSvc.Registry(crSvc.microRtr),
		microSvc.RegisterTTL(time.Duration(crSvc.conf.Server.RegisterTTL)*time.Second),
		microSvc.RegisterInterval(time.Duration(crSvc.conf.Server.RegisterInterval)*time.Second),
		microSvc.Version("latest"),
	)
	svc.Init()

	err := clusterRes.RegisterClusterResourcesHandler(svc.Server(), handler.NewClusterResourcesHandler())
	if err != nil {
		blog.Errorf("register cluster resources handler to micro failed: %s", err.Error())
		return err
	}

	crSvc.microSvc = svc
	blog.Infof("register cluster resources handler to micro successfully.")
	return nil
}

// 注册服务到 Etcd
func (crSvc *clusterResourcesService) initRegistry() error {
	etcdEndpoints := utils.SplitString(crSvc.conf.Etcd.EtcdEndpoints)
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

	blog.Infof("registry: etcd endpoints: %v, secure: %t", etcdEndpoints, etcdSecure)

	crSvc.microRtr = microEtcd.NewRegistry(
		microRgt.Addrs(etcdEndpoints...),
		microRgt.Secure(etcdSecure),
		microRgt.TLSConfig(etcdTLS),
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
			crSvc.conf.Server.Ca, crSvc.conf.Server.Cert, crSvc.conf.Server.Key, static.ServerCertPwd,
		)
		if err != nil {
			blog.Errorf("load cluster resources server tls config failed: %s", err.Error())
			return err
		}
		crSvc.tlsConfig = tlsConfig
		blog.Infof("load cluster resources server tls config successfully")
	}

	if len(crSvc.conf.Client.Cert) != 0 && len(crSvc.conf.Client.Key) != 0 && len(crSvc.conf.Client.Ca) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(
			crSvc.conf.Client.Ca, crSvc.conf.Client.Cert, crSvc.conf.Client.Key, static.ClientCertPwd,
		)
		if err != nil {
			blog.Errorf("load cluster resources client tls config failed: %s", err.Error())
			return err
		}
		crSvc.clientTLSConfig = tlsConfig
		blog.Infof("load cluster resources client tls config successfully")
	}
	return nil
}

// 初始化 HTTP 服务
func (crSvc *clusterResourcesService) initHTTPService() error {
	rmMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(utils.CustomHeaderMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}),
	)

	grpcDialconf := make([]grpc.DialOption, 0)
	if crSvc.tlsConfig != nil && crSvc.clientTLSConfig != nil {
		grpcDialconf = append(grpcDialconf, grpc.WithTransportCredentials(grpcCreds.NewTLS(crSvc.clientTLSConfig)))
	} else {
		grpcDialconf = append(grpcDialconf, grpc.WithInsecure())
	}
	err := clusterRes.RegisterClusterResourcesGwFromEndpoint(
		context.TODO(),
		rmMux,
		crSvc.conf.Server.Address+":"+strconv.Itoa(crSvc.conf.Server.Port),
		grpcDialconf,
	)
	if err != nil {
		blog.Errorf("register http service failed: %s", err)
		return fmt.Errorf("register http service failed: %s", err)
	}

	router := mux.NewRouter()
	router.Handle("/{uri:.*}", rmMux)
	blog.Info("register grpc service handler to path /")

	originMux := http.NewServeMux()
	originMux.Handle("/", router)

	// 检查是否需要启用 swagger 服务
	if crSvc.conf.Swagger.Enabled && len(crSvc.conf.Swagger.Dir) != 0 {
		blog.Infof("swagger doc is enabled")
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
		Handler: originMux,
	}
	go func() {
		var err error
		blog.Infof("start http gateway server on address %s", httpAddr)
		if crSvc.tlsConfig != nil {
			crSvc.httpServer.TLSConfig = crSvc.tlsConfig
			err = crSvc.httpServer.ListenAndServeTLS("", "")
		} else {
			err = crSvc.httpServer.ListenAndServe()
		}
		if err != nil {
			blog.Errorf("start http gateway server failed, %s", err.Error())
			crSvc.stopCh <- struct{}{}
		}
	}()
	return nil
}
