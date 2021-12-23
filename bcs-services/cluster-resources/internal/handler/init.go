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

package handler

import (
	"cluster-resources/internal/common"
	"cluster-resources/internal/utils"
	clusterRes "cluster-resources/proto/cluster-resources"
	"cluster-resources/swagger"
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	goBindataAssetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	microRgt "github.com/micro/go-micro/v2/registry"
	microEtcd "github.com/micro/go-micro/v2/registry/etcd"
	microSvc "github.com/micro/go-micro/v2/service"
	microGrpc "github.com/micro/go-micro/v2/service/grpc"
	"google.golang.org/grpc"
	grpcCred "google.golang.org/grpc/credentials"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

// ClusterResources 服务初始化执行集
func (cr *ClusterResources) Init() error {
	// 各个初始化方法依次执行
	for _, f := range []func() error{
		cr.initTLSConfig,
		cr.initRegistry,
		cr.initMicro,
		cr.initHTTPService,
	} {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

// ClusterResources 服务启动逻辑
func (cr *ClusterResources) Run() error {
	if err := cr.microSvc.Run(); err != nil {
		return err
	}
	return nil
}

// 初始化 MicroService
func (cr *ClusterResources) initMicro() error {
	svc := microGrpc.NewService(
		microSvc.Name(common.ServiceDomain),
		microGrpc.WithTLS(cr.tlsConfig),
		microSvc.Address(cr.opts.Server.Address+":"+strconv.Itoa(int(cr.opts.Server.Port))),
		microSvc.Registry(cr.microRtr),
		microSvc.RegisterTTL(30*time.Second),
		microSvc.RegisterInterval(25*time.Second),
		microSvc.Version("latest"),
	)
	svc.Init()

	if err := clusterRes.RegisterClusterResourcesHandler(svc.Server(), cr); err != nil {
		blog.Errorf("register cluster resources handler to micro failed: %s", err.Error())
		return err
	}

	cr.microSvc = svc
	blog.Infof("register cluster resources handler to micro successfully.")
	return nil
}

// 注册服务到 Etcd
func (cr *ClusterResources) initRegistry() error {
	etcdEndpoints := utils.SplitAddrString(cr.opts.Etcd.EtcdEndpoints)
	etcdSecure := false

	var etcdTLS *tls.Config
	var err error
	if len(cr.opts.Etcd.EtcdCa) != 0 && len(cr.opts.Etcd.EtcdCert) != 0 && len(cr.opts.Etcd.EtcdKey) != 0 {
		etcdSecure = true
		etcdTLS, err = ssl.ClientTslConfVerity(cr.opts.Etcd.EtcdCa, cr.opts.Etcd.EtcdCert, cr.opts.Etcd.EtcdKey, "")
		if err != nil {
			return err
		}
	}

	blog.Infof("registry: etcd endpoints: %v, secure: %t", etcdEndpoints, etcdSecure)

	cr.microRtr = microEtcd.NewRegistry(
		microRgt.Addrs(etcdEndpoints...),
		microRgt.Secure(etcdSecure),
		microRgt.TLSConfig(etcdTLS),
	)
	if err := cr.microRtr.Init(); err != nil {
		return err
	}
	return nil
}

// 自定义 HTTP Header Matcher
func CustomMatcher(key string) (string, bool) {
	switch key {
	case "X-Request-Id":
		return "X-Request-Id", true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// 初始化 Server 与 client TLS 配置
func (cr *ClusterResources) initTLSConfig() error {
	if len(cr.opts.Server.Cert) != 0 && len(cr.opts.Server.Key) != 0 && len(cr.opts.Server.Ca) != 0 {
		tlsConfig, err := ssl.ServerTslConfVerityClient(cr.opts.Server.Ca, cr.opts.Server.Cert,
			cr.opts.Server.Key, static.ServerCertPwd)
		if err != nil {
			blog.Errorf("load cluster resources server tls config failed: %s", err.Error())
			return err
		}
		cr.tlsConfig = tlsConfig
		blog.Infof("load cluster resources server tls config successfully")
	}

	if len(cr.opts.Client.Cert) != 0 && len(cr.opts.Client.Key) != 0 && len(cr.opts.Client.Ca) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(cr.opts.Client.Ca, cr.opts.Client.Cert,
			cr.opts.Client.Key, static.ClientCertPwd)
		if err != nil {
			blog.Errorf("load cluster resources client tls config failed: %s", err.Error())
			return err
		}
		cr.clientTLSConfig = tlsConfig
		blog.Infof("load cluster resources client tls config successfully")
	}
	return nil
}

// 初始化 HTTP 服务
func (cr *ClusterResources) initHTTPService() error {
	rmMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(CustomMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}),
	)

	grpcDialOpts := make([]grpc.DialOption, 0)
	if cr.tlsConfig != nil && cr.clientTLSConfig != nil {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(grpcCred.NewTLS(cr.clientTLSConfig)))
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
	}
	err := clusterRes.RegisterClusterResourcesGwFromEndpoint(
		context.TODO(),
		rmMux,
		cr.opts.Server.Address+":"+strconv.Itoa(int(cr.opts.Server.Port)),
		grpcDialOpts,
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
	if cr.opts.Swagger.Enabled && len(cr.opts.Swagger.Dir) != 0 {
		blog.Infof("swagger doc is enabled")
		// 挂载 swagger.json 文件目录
		originMux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path.Join(cr.opts.Swagger.Dir, strings.TrimPrefix(r.URL.Path, "/swagger/")))
		})
		// 配置 swagger-ui 服务
		fileServer := http.FileServer(&goBindataAssetfs.AssetFS{
			Asset:    swagger.Asset,
			AssetDir: swagger.AssetDir,
			Prefix:   "third_party/swagger-ui",
		})
		originMux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", fileServer))
	}

	httpAddr := cr.opts.Server.Address + ":" + strconv.Itoa(int(cr.opts.Server.HTTPPort))
	cr.httpServer = &http.Server{
		Addr:    httpAddr,
		Handler: originMux,
	}
	go func() {
		var err error
		blog.Infof("start http gateway server on address %s", httpAddr)
		if cr.tlsConfig != nil {
			cr.httpServer.TLSConfig = cr.tlsConfig
			err = cr.httpServer.ListenAndServeTLS("", "")
		} else {
			err = cr.httpServer.ListenAndServe()
		}
		if err != nil {
			blog.Errorf("start http gateway server failed, %s", err.Error())
			cr.stopCh <- struct{}{}
		}
	}()
	return nil
}
