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
 *
 */

package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/cmd/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/cmd/pkgs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/consumer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/proto/alertmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/server/service"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/types"

	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	microsvc "github.com/micro/go-micro/v2/service"
	grpcsvc "github.com/micro/go-micro/v2/service/grpc"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	apiPrefix = "/alertmanager"
)

var (
	// ErrServerNotInit server not init
	ErrServerNotInit = errors.New("server not init")
)

// AlertManager alert manager
type AlertManager struct {
	// options for alertmanager
	options *config.AlertManagerOptions

	// server http tls authentication
	tlsServerConfig *tls.Config

	// client http tls authentication
	tlsClientConfig *tls.Config

	// server handler
	serverHandler *service.AlertManager

	// main http server
	mainServer *http.Server

	// pprof、metrics、swagger server
	extraServer *http.Server

	// msgQueue consumer manager
	consumer *consumer.Consumers

	// micro service
	microService microsvc.Service
	// micro registry
	microRegistry registry.Registry

	ctx    context.Context
	cancel context.CancelFunc
	// http server err quit
	stop chan error
}

// NewAlertManager create alertmanager
func NewAlertManager(options *config.AlertManagerOptions) *AlertManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &AlertManager{
		options: options,
		ctx:     ctx,
		cancel:  cancel,
		stop:    make(chan error),
	}
}

// initSvrCliTLSConfig init server and client tlsConfig
func (am *AlertManager) initSvrCliTLSConfig() error {
	if am == nil {
		return ErrServerNotInit
	}

	if len(am.options.CAFile) != 0 && len(am.options.ServerCertFile) != 0 && len(am.options.ServerKeyFile) != 0 {
		tlsServerConfig, err := ssl.ServerTslConfVerityClient(am.options.CAFile, am.options.ServerCertFile,
			am.options.ServerKeyFile, static.ServerCertPwd)
		if err != nil {
			blog.Errorf("initServerTLSConfig failed: %v", err)
			return err
		}

		am.tlsServerConfig = tlsServerConfig
		blog.Infof("initServerTLSConfig successful")
	}

	if len(am.options.CertConfig.CAFile) != 0 && len(am.options.CertConfig.ClientCertFile) != 0 && len(am.options.ClientKeyFile) != 0 {
		tlsClientConfig, err := ssl.ClientTslConfVerity(am.options.CAFile, am.options.ClientCertFile,
			am.options.ClientKeyFile, static.ClientCertPwd)
		if err != nil {
			blog.Errorf("initClientTLSConfig failed: %v", err)
			return err
		}

		am.tlsClientConfig = tlsClientConfig
		blog.Infof("initClientTLSConfig successful")
	}

	return nil
}

// init consumer & run consumer
func (am *AlertManager) initConsumers() error {
	if am == nil {
		return ErrServerNotInit
	}

	consumers := pkgs.GetFactoryConsumers(am.options)
	msgQueue := pkgs.GetQueueClient(am.options)

	con := consumer.NewConsumers(consumers, msgQueue)
	if con == nil {
		panic("initConsumers failed")
	}
	am.consumer = con
	// run consumer sub handler
	am.consumer.Run()
	return nil
}

// init micro etcd registry
func (am *AlertManager) initRegistry() error {
	if am == nil {
		return ErrServerNotInit
	}
	if len(am.options.CMDOptions.Address) == 0 {
		errMsg := fmt.Errorf("etcdServers invalid")
		return errMsg
	}
	servers := strings.Split(am.options.CMDOptions.Address, ";")

	var (
		secureEtcd bool
		etcdTLS    *tls.Config
		err        error
	)

	if len(am.options.CMDOptions.CA) != 0 && len(am.options.CMDOptions.Cert) != 0 && len(am.options.CMDOptions.Key) != 0 {
		secureEtcd = true

		etcdTLS, err = ssl.ClientTslConfVerity(am.options.CMDOptions.CA, am.options.CMDOptions.Cert,
			am.options.CMDOptions.Key, "")
		if err != nil {
			return err
		}
	}

	am.microRegistry = etcd.NewRegistry(
		registry.Addrs(servers...),
		registry.Secure(secureEtcd),
		registry.TLSConfig(etcdTLS),
	)
	if err := am.microRegistry.Init(); err != nil {
		return err
	}

	return nil
}

// init HTTP Service
func (am *AlertManager) initPProf(router *mux.Router) {
	if am == nil {
		return
	}

	if !am.options.DebugMode {
		blog.Infof("pprof debugMode is off")
		return
	}

	blog.Infof("pprof debugMode is on")

	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

// init swagger server file
func (am *AlertManager) initServerSwaggerFile(mux *http.ServeMux) {
	if am == nil {
		return
	}

	if len(am.options.SwaggerConfigDir.Dir) != 0 {
		blog.Infof("swagger config dir is enabled")

		mux.HandleFunc(apiPrefix+"/swagger/", am.serveSwaggerFile)
	}
}

func (am *AlertManager) serveSwaggerFile(w http.ResponseWriter, r *http.Request) {
	swaggerFile := path.Join(am.options.SwaggerConfigDir.Dir, strings.TrimPrefix(r.URL.Path, apiPrefix+"/swagger/"))
	blog.Infof("Serving swagger-file: %s", swaggerFile)

	http.ServeFile(w, r, swaggerFile)
}

// init prometheus metrics handler
func (am *AlertManager) initMetrics(router *mux.Router) {
	blog.Infof("init metrics handler")
	router.Handle(apiPrefix+"/metrics", promhttp.Handler())
}

func customMatcher(key string) (string, bool) {
	switch key {
	case "X-Request-Id":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// init httpGateWay trans grpc to http
func (am *AlertManager) initHTTPGateWay(router *mux.Router) error {
	if am == nil {
		return ErrServerNotInit
	}
	// customizing_your_gateway:
	// https://github.com/grpc-ecosystem/grpc-gateway/blob/master/docs/docs/mapping/customizing_your_gateway.md
	gmux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(customMatcher),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			OrigName: true,
		}),
	)

	dialOpts := []grpc.DialOption{}
	if am.tlsServerConfig != nil && am.tlsClientConfig != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(am.tlsClientConfig)))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	err := alertmanager.RegisterAlertManagerGwFromEndpoint(
		am.ctx,
		gmux,
		am.options.ServiceConfig.Address+":"+strconv.Itoa(int(am.options.ServiceConfig.Port)),
		dialOpts)
	if err != nil {
		errMsg := fmt.Sprintf("http gateway RegisterAlertManagerGwFromEndpoint failed: %v", err)
		blog.Errorf(errMsg)
		return errors.New(errMsg)
	}

	// register handler to router
	router.Handle("/{uri:.*}", gmux)
	blog.Infof("register http gateway to router successful")

	return nil
}

// init main http server
func (am *AlertManager) initMainHTTPServer() error {
	if am == nil {
		return ErrServerNotInit
	}

	router := mux.NewRouter()

	err := am.initHTTPGateWay(router)
	if err != nil {
		return err
	}

	mainAddress := am.options.ServiceConfig.Address + ":" + fmt.Sprintf("%d", am.options.ServiceConfig.Port-1)
	am.mainServer = &http.Server{
		Addr:    mainAddress,
		Handler: router,
	}

	go func() {
		var err error
		blog.Infof("initMainHttpServer address: %s", mainAddress)
		if am.tlsServerConfig != nil {
			am.mainServer.TLSConfig = am.tlsServerConfig
			err = am.mainServer.ListenAndServeTLS("", "")
		} else {
			err = am.mainServer.ListenAndServe()
		}
		// server failed
		if err != nil {
			blog.Errorf("http server failed: %v", err)
			am.stop <- err
		}
	}()

	return nil
}

// init extra http server(metrics, serverSwagger, pprof)
func (am *AlertManager) initExtraHTTPServer() error {
	if am == nil {
		return ErrServerNotInit
	}

	router := mux.NewRouter()
	am.initMetrics(router)
	am.initPProf(router)

	mux := http.NewServeMux()
	mux.Handle("/", router)
	am.initServerSwaggerFile(mux)

	extraAddress := am.options.ServiceConfig.Address + ":" + strconv.Itoa(int(am.options.MetricPort))
	am.extraServer = &http.Server{
		Addr:    extraAddress,
		Handler: mux,
	}

	go func() {
		var err error
		blog.Infof("initExtraHttpServer address: %s", extraAddress)

		err = am.extraServer.ListenAndServe()
		if err != nil {
			blog.Errorf("initExtraHttpServer failed: %v", err)
			am.stop <- err
		}
	}()

	return nil
}

// init microService for grpc
func (am *AlertManager) initMicroService() error {
	if am == nil {
		return ErrServerNotInit
	}
	microService := grpcsvc.NewService(
		microsvc.Context(am.ctx),
		microsvc.Name(types.ServiceName),
		microsvc.Version(version.BcsVersion),
		microsvc.Address(am.options.ServiceConfig.Address+":"+strconv.Itoa(int(am.options.ServiceConfig.Port))),
		grpcsvc.WithTLS(am.tlsServerConfig),
		microsvc.Registry(am.microRegistry),
		microsvc.RegisterInterval(30*time.Second),
		microsvc.RegisterTTL(40*time.Second),
	)
	microService.Init()

	// create handler && register handler
	am.serverHandler = service.NewAlertManager(pkgs.GetAlertClient(am.options))
	alertmanager.RegisterAlertManagerHandler(microService.Server(), am.serverHandler)

	am.microService = microService

	return nil
}

// waitServerQuitSignal graceful shut down
func (am *AlertManager) waitServerQuitSignal() {
	if am == nil {
		return
	}

	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		select {
		case e := <-signalChan:
			blog.Infof("receive interrupt signal: %s", e.String())
			am.close()
		case <-am.stop:
			blog.Infof("http server quit")
			am.close()
		}
	}()
}

// close alertmanager
func (am *AlertManager) close() {
	if am == nil {
		return
	}

	timeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	am.consumer.Stop()
	am.extraServer.Shutdown(timeCtx)
	am.mainServer.Shutdown(timeCtx)

	am.cancel()
}

// initAlertManagerServer init alert manager server
func (am *AlertManager) initAlertManagerServer() error {
	if am == nil {
		return ErrServerNotInit
	}

	// init tlsConfig
	if err := am.initSvrCliTLSConfig(); err != nil {
		return err
	}
	// init register
	if err := am.initRegistry(); err != nil {
		return err
	}
	// init Consumer: msgQueue/consumer
	if err := am.initConsumers(); err != nil {
		return err
	}
	// init httpServer
	if err := am.initMainHTTPServer(); err != nil {
		return err
	}
	// init extraServer
	if err := am.initExtraHTTPServer(); err != nil {
		return err
	}
	// init microService
	if err := am.initMicroService(); err != nil {
		return err
	}
	// wait quitHandler
	am.waitServerQuitSignal()

	return nil
}

// Run init alertmanager & run microService
func (am *AlertManager) Run() error {
	if am == nil {
		return ErrServerNotInit
	}
	defer blog.CloseLogs()

	err := am.initAlertManagerServer()
	if err != nil {
		blog.Errorf("initAlertManagerServer failed: %v", err)
		return err
	}

	// run micro Grpc service: block here wait am.cancel()
	err = am.microService.Run()
	if err != nil {
		blog.Fatal("microService quit: %v", err)
		return err
	}

	return nil
}
