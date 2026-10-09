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

package bcsstorage

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/ipv6server"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
	trestful "github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/restful"
	restful "github.com/emicklei/go-restful/v3"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	mserver "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/server"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/utils/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/utils/middle"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
)

// StorageServer is a data struct of bcs storage server
type StorageServer struct {
	conf       *options.StorageOptions
	httpServer *httpserver.HttpServer
	// micro server
	microServer *mserver.MicroServer
	// etcd
	etcdRegistry registry.Registry

	// 证书
	etcdTLSConfig *tls.Config

	// 上下文
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewStorageServer create storage server object
func NewStorageServer(op *options.StorageOptions) (*StorageServer, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	// 创建 StorageServer
	s := &StorageServer{
		// Configuration
		conf: op,
		// Http server
		httpServer: httpserver.NewHttpServer(op.Port, op.Address, ""),
		// micro server
		microServer: mserver.NewMicroServer(ctx, cancelFunc),
		// 上下文
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}
	return s, nil
}

// Init 初始化bcs-storage
func (s *StorageServer) Init() (err error) {
	// 证书初始化
	if err = s.initTlsConfig(); err != nil {
		return errors.Wrapf(err, "certificate initialization failed")
	}

	// 初始化http
	if err = s.initHTTPServer(); err != nil {
		return errors.Wrapf(err, "http server initialization failed")
	}

	// v1 http server 注册
	s.registerV1HttpServerToRegistry()

	// 初始化go-micro
	if err = s.microServer.Init(s.conf); err != nil {
		return errors.Wrapf(err, "v2 server initialization failed")
	}

	return nil
}

// registerV1HttpServerToRegistry 注册v1 http server到Registry
func (s *StorageServer) registerV1HttpServerToRegistry() {
	// RDiscover
	if s.conf.Etcd.Feature {
		// 创建metadata
		meta := make(map[string]string)
		// web server 端口号
		port := strconv.FormatUint(uint64(s.conf.Port), 10)
		// ipv4 地址
		ipv4 := net.JoinHostPort(s.conf.Address, port)
		// ipv6 注册地址不能是本地回环地址
		if ipv6 := net.ParseIP(s.conf.IPv6Address); ipv6 != nil && !ipv6.IsLoopback() {
			// 把ipv6地址写入到mata中
			meta[types.IPV6] = net.JoinHostPort(ipv6.String(), port)
		}

		// use go-micro v4 registry
		eoption := &registry.Options{
			Name:         constants.ServerName,
			Version:      version.BcsVersion,
			RegistryAddr: strings.Split(s.conf.Etcd.Address, ","),
			RegAddr:      ipv4,
			Config:       s.etcdTLSConfig,
			Meta:         meta,
		}
		blog.Infof("#############storage turn on etcd registry feature, options %+v ###############", eoption)
		s.etcdRegistry = registry.NewEtcdRegistry(eoption)
	}

}

func (s *StorageServer) initFilterFunctions() []restful.FilterFunction {
	filterFunctions := make([]restful.FilterFunction, 0)

	// register middleware
	mdlw := middle.New(middle.Options{
		Recorder: metrics.NewRecorder(metrics.Config{
			Prefix: middle.MetricsPrefix,
		}),
		GroupedStatus: true,
	})

	filterFunctions = append(filterFunctions, trestful.NewOTFilter(opentracing.GlobalTracer()))
	filterFunctions = append(filterFunctions, middle.MetricsMiddleHandler(mdlw))

	return filterFunctions
}

// initTlsConfig 初始化 tls 配置
func (s *StorageServer) initTlsConfig() (err error) {
	// set http tls
	if s.conf.ServerCert.IsSSL {
		s.httpServer.SetSsl(
			s.conf.ServerCert.CAFile,
			s.conf.ServerCert.CertFile,
			s.conf.ServerCert.KeyFile,
			s.conf.ServerCert.CertPwd,
		)
	}

	// get etcd tls
	if s.etcdTLSConfig, err = s.conf.Etcd.GetTLSConfig(); err != nil {
		return errors.Wrapf(err, "storage loading etcd registry tls config failed")
	}
	// set etcd tls
	s.microServer.SetEtcdTLSConfig(s.etcdTLSConfig)

	/*
	   init micro server and client tls
	*/
	// loading client tls configuration
	if len(s.conf.ClientCert.CertFile) != 0 && len(s.conf.ClientCert.KeyFile) != 0 &&
		len(s.conf.ClientCert.CAFile) != 0 {
		cliConfig, err := ssl.ClientTslConfVerity(
			s.conf.ClientCert.CAFile,
			s.conf.ClientCert.CertFile,
			s.conf.ClientCert.KeyFile,
			s.conf.ClientCert.CertPwd,
		)
		if err != nil {
			return errors.Wrapf(err, "loading client side tls configuration failed")
		}
		s.microServer.SetClientTLSConfig(cliConfig)
	}

	if len(s.conf.ServerCert.CertFile) != 0 && len(s.conf.ServerCert.KeyFile) != 0 &&
		len(s.conf.ServerCert.CertFile) != 0 {
		// loading server tls configuration
		svrConfig, err := ssl.ServerTslConfVerityClient(
			s.conf.ServerCert.CAFile,
			s.conf.ServerCert.CertFile,
			s.conf.ServerCert.KeyFile,
			s.conf.ServerCert.CertPwd,
		)
		if err != nil {
			return errors.Wrapf(err, "loading server side tls config failed")
		}
		s.microServer.SetServerTLSConfig(svrConfig)
	}

	return nil
}

// initHTTPServer 初始化http server
func (s *StorageServer) initHTTPServer() error {

	// ApiResource
	a := apiserver.GetAPIResource()
	if err := a.SetConfig(s.conf); err != nil {
		return err
	}
	a.InitActions()

	// register middleware
	filterFunctions := s.initFilterFunctions()

	// Api v1
	if err := s.httpServer.RegisterWebServer(actions.PathV1, filterFunctions, a.ActionsV1); err != nil {
		return err
	}

	if a.Conf.DebugMode {
		s.initDebug()
	}
	return nil
}

func (s *StorageServer) initDebug() {
	action := []*httpserver.Action{
		httpserver.NewAction("GET", "/debug/pprof/", nil, getRouteFunc(pprof.Index)),
		httpserver.NewAction("GET", "/debug/pprof/{uri:*}", nil, getRouteFunc(pprof.Index)),
		httpserver.NewAction("GET", "/debug/pprof/cmdline", nil, getRouteFunc(pprof.Cmdline)),
		httpserver.NewAction("GET", "/debug/pprof/profile", nil, getRouteFunc(pprof.Profile)),
		httpserver.NewAction("GET", "/debug/pprof/symbol", nil, getRouteFunc(pprof.Symbol)),
		httpserver.NewAction("GET", "/debug/pprof/trace", nil, getRouteFunc(pprof.Trace)),
	}
	_ = s.httpServer.RegisterWebServer("", nil, action)
}

// Start to run storage server
func (s *StorageServer) Start() error {
	chErr := make(chan error, 1)

	defer func() {
		// 注销服务
		if s.conf.Etcd.Feature {
			_ = s.etcdRegistry.Deregister()
		}
		// 关闭资源
		s.close()
	}()

	go func() {
		// run micro server
		chErr <- s.microServer.Run()
	}()

	go func() {
		// run http server
		blog.Info("run v1 http server")
		s.httpServer.SetAddressIPv6(s.conf.IPv6Address) // 把ipv6地址，加入到httpServer中
		err := s.httpServer.ListenAndServe()
		chErr <- errors.Wrapf(err, "http listen and service failed")
	}()

	runPrometheusMetrics(s.conf)

	// startDaemon
	actions.StartActionDaemon()

	// register and discover
	if s.conf.Etcd.Feature {
		if err := s.etcdRegistry.Register(); err != nil {
			chErr <- errors.Wrapf(err, "storage etcd registry failed")
		}
	}

	err := <-chErr
	return errors.Wrapf(err, "exit")
}

// close 关闭bcs-storage
func (s *StorageServer) close() {
	// 关闭 micro server
	s.cancelFunc()
	// 关闭 http server
	_ = s.httpServer.Close()
}

// runPrometheusMetrics starting prometheus metrics handler
func runPrometheusMetrics(op *options.StorageOptions) {
	http.Handle("/metrics", promhttp.Handler())
	// ipv4 ipv6
	ips := []string{op.Address}
	if op.IPv6Address != "" && op.IPv6Address != op.Address {
		ips = append(ips, op.IPv6Address)
	}
	ipv6Server := ipv6server.NewIPv6Server(ips, strconv.Itoa(int(op.MetricPort)), "", nil)
	// 启动server，同时监听v4、v6地址
	// nolint
	go ipv6Server.ListenAndServe()
}

func getRouteFunc(f http.HandlerFunc) restful.RouteFunction {
	return restful.RouteFunction(func(req *restful.Request, resp *restful.Response) {
		f(resp, req.Request)
	})
}
