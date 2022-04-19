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
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	microRgt "github.com/micro/go-micro/v2/registry"
	microEtcd "github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-micro/v2/server"
	serverGrpc "github.com/micro/go-micro/v2/server/grpc"
	microSvc "github.com/micro/go-micro/v2/service"
	microGrpc "github.com/micro/go-micro/v2/service/grpc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	grpcCred "google.golang.org/grpc/credentials"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common/config"
	conf "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/stringx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/version"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/wrapper"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
)

// Project describe a project instance
type ProjectService struct {
	opt *conf.ProjectConfig

	// mongo DB options
	mongoOptions *mongo.Options
	model        store.ProjectModel

	microSvc  microSvc.Service
	microRgt  microRgt.Registry
	discovery *discovery.ModuleDiscovery

	// http service
	httpServer *http.Server

	// metric service
	metricServer *http.Server

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
		p.initMongo,
		p.initRegistry,
		p.initDiscovery,
		p.initMicro,
		p.initHttpService,
		p.initMetric,
	} {
		if err := f(); err != nil {
			return err
		}
	}
	return nil
}

// Run helm manager server
func (p *ProjectService) Run() error {
	// run the service
	if err := p.microSvc.Run(); err != nil {
		logging.Error("run micro service failed, err: %s", err.Error())
		return err
	}
	return nil
}

// init server and client tls config
func (p *ProjectService) initTLSConfig() error {
	if len(p.opt.Server.Cert) != 0 && len(p.opt.Server.Key) != 0 && len(p.opt.Server.Ca) != 0 {
		tlsConfig, err := ssl.ServerTslConfVerityClient(p.opt.Server.Ca, p.opt.Server.Cert,
			p.opt.Server.Key, p.opt.Server.CertPwd)
		if err != nil {
			logging.Error("load project server tls config failed, err %s", err.Error())
			return err
		}
		p.tlsConfig = tlsConfig
		logging.Info("load project server tls config successfully")
	}

	if len(p.opt.Client.Cert) != 0 && len(p.opt.Client.Key) != 0 && len(p.opt.Client.Ca) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(p.opt.Client.Ca, p.opt.Client.Cert,
			p.opt.Client.Key, p.opt.Client.CertPwd)
		if err != nil {
			logging.Error("load project client tls config failed, err %s", err.Error())
			return err
		}
		p.clientTLSConfig = tlsConfig
		logging.Info("load project client tls config successfully")
	}
	return nil
}

// init mongo client
func (p *ProjectService) initMongo() error {
	if len(p.opt.Mongo.Address) == 0 {
		return fmt.Errorf("mongo address cannot be empty")
	}
	if len(p.opt.Mongo.Database) == 0 {
		return fmt.Errorf("mongo database cannot be empty")
	}
	// 判断 password 是否加密，如果加密需要解密获取到原始数据
	// 使用 bcs service 统一的 pwd
	password := p.opt.Mongo.Password
	if password != "" && p.opt.Mongo.Encrypted {
		realPwd, _ := encrypt.DesDecryptFromBase([]byte(password))
		password = string(realPwd)
	}

	mongoOptions := &mongo.Options{
		Hosts:                 strings.Split(p.opt.Mongo.Address, ","),
		ConnectTimeoutSeconds: int(p.opt.Mongo.ConnectTimeout),
		Database:              p.opt.Mongo.Database,
		Username:              p.opt.Mongo.Username,
		Password:              password,
		MaxPoolSize:           uint64(p.opt.Mongo.MaxPoolSize),
		MinPoolSize:           uint64(p.opt.Mongo.MinPoolSize),
	}
	p.mongoOptions = mongoOptions

	mongoDB, err := mongo.NewDB(mongoOptions)
	if err != nil {
		logging.Error("create mongo error, err: %s", err.Error())
		return err
	}
	if err = mongoDB.Ping(); err != nil {
		logging.Error("connect mongo error, err: %s", err.Error())
		return err
	}
	logging.Info("init mongo successfully")
	modelSet := store.New(mongoDB)
	p.model = modelSet
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
	p.discovery = discovery.NewModuleDiscovery(config.ServiceDomain, p.microRgt)
	logging.Info("init discovery for project service successfully")
	return nil
}

// init micro service
func (p *ProjectService) initMicro() error {
	// max size: 50M, add grpc address to access
	server := serverGrpc.NewServer(serverGrpc.MaxMsgSize(config.MaxMsgSize), server.Address(fmt.Sprintf(":%d", p.opt.Server.Port)))
	svc := microGrpc.NewService(
		microSvc.Name(config.ServiceDomain),
		microSvc.Metadata(map[string]string{
			config.MicroMetaKeyHTTPPort: strconv.Itoa(int(p.opt.Server.HTTPPort)),
		}),
		microGrpc.WithTLS(p.tlsConfig),
		microSvc.Address(p.opt.Server.Address+":"+strconv.Itoa(int(p.opt.Server.Port))),
		microSvc.Registry(p.microRgt),
		microSvc.Version(version.Version),
		microSvc.RegisterTTL(30*time.Second),      // add ttl to config
		microSvc.RegisterInterval(25*time.Second), // add interval to config
		microSvc.Context(p.ctx),
		microSvc.Server(server),
		microSvc.AfterStart(func() error {
			return p.discovery.Start()
		}),
		microSvc.BeforeStop(func() error {
			p.discovery.Stop()
			return nil
		}),
		microSvc.WrapHandler(
			wrapper.NewInjectRequestIDWrapper,
			wrapper.NewLogWrapper,
			wrapper.NewResponseWrapper,
			wrapper.NewValidatorWrapper,
		),
	)
	svc.Init()

	// project hander
	if err := proto.RegisterBCSProjectHandler(svc.Server(), handler.NewProject(p.model)); err != nil {
		logging.Error("register handler failed, err: %s", err.Error())
		return err
	}

	p.microSvc = svc
	logging.Info("success to register project service handler to micro")
	return nil
}

// init http gateway
func (p *ProjectService) initHTTPGateway(router *mux.Router) error {
	gwMux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(CustomMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			OrigName:     true,
			EmitDefaults: true,
		}),
	)
	grpcDialOpts := []grpc.DialOption{}
	if p.tlsConfig != nil && p.clientTLSConfig != nil {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(grpcCred.NewTLS(p.clientTLSConfig)))
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
	}

	grpcDialOpts = append(grpcDialOpts, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(config.MaxMsgSize), grpc.MaxCallSendMsgSize(config.MaxMsgSize)))
	err := proto.RegisterBCSProjectGwFromEndpoint(
		context.TODO(),
		gwMux,
		p.opt.Server.Address+":"+strconv.Itoa(int(p.opt.Server.Port)),
		grpcDialOpts,
	)
	if err != nil {
		logging.Error("register http gateway failed, err %s", err.Error())
		return fmt.Errorf("register http gateway failed, err %s", err.Error())
	}
	router.Handle("/{uri:.*}", gwMux)
	logging.Info("register grpc gateway handler to path /")
	return nil
}

// init swagger
func (p *ProjectService) initSwagger(mux *http.ServeMux) {
	if len(p.opt.Swagger.Dir) != 0 {
		logging.Info("swagger doc is enabled")
		mux.HandleFunc("/swagger/", func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, path.Join(p.opt.Swagger.Dir, strings.TrimPrefix(r.URL.Path, "/swagger/")))
		})
	}
}

// init http service
func (p *ProjectService) initHttpService() error {
	router := mux.NewRouter()
	// init micro http gateway
	if err := p.initHTTPGateway(router); err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", router)
	// init swagger
	p.initSwagger(mux)

	httpAddr := p.opt.Server.Address + ":" + strconv.Itoa(int(p.opt.Server.HTTPPort))
	p.httpServer = &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}
	go func() {
		var err error
		logging.Info("start http server on address %s", httpAddr)
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

func (p *ProjectService) initMetric() error {
	logging.Info("init metric handler")
	metricAddr := p.opt.Server.Address + ":" + strconv.Itoa(int(p.opt.Server.MetricPort))
	metricMux := http.NewServeMux()
	metricMux.Handle("/metrics", promhttp.Handler())
	p.metricServer = &http.Server{
		Addr:    metricAddr,
		Handler: metricMux,
	}

	go func() {
		var err error
		logging.Info("start metric server on address %s", metricAddr)
		if err = p.metricServer.ListenAndServe(); err != nil {
			logging.Error("start metric server failed, %s", err.Error())
			p.stopCh <- struct{}{}
		}
	}()
	return nil
}

// CustomMatcher for http header
func CustomMatcher(key string) (string, bool) {
	switch key {
	case "X-Request-Id":
		return "X-Request-Id", true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
