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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registryv4"
	trestful "github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/restful"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/constants"
	mserver "github.com/Tencent/bk-bcs/bcs-services/bcs-storage/pkg/server"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/utils/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/utils/middle"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	"github.com/emicklei/go-restful"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StorageServer is a data struct of bcs storage server
type StorageServer struct {
	conf       *options.StorageOptions
	httpServer *httpserver.HttpServer
	// micro server
	microServer *mserver.MicroServer
	// etcd
	etcdRegistry registryv4.Registry

	// ??????
	etcdTLSConfig *tls.Config

	// ?????????
	ctx        context.Context
	cancelFunc context.CancelFunc
}

// NewStorageServer create storage server object
func NewStorageServer(op *options.StorageOptions) (*StorageServer, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())
	// ?????? StorageServer
	s := &StorageServer{
		// Configuration
		conf: op,
		// Http server
		httpServer: httpserver.NewHttpServer(op.Port, op.Address, ""),
		// micro server
		microServer: mserver.NewMicroServer(ctx, cancelFunc),
		// ?????????
		ctx:        ctx,
		cancelFunc: cancelFunc,
	}
	return s, nil
}

// Init ?????????bcs-storage
func (s *StorageServer) Init() (err error) {
	// ???????????????
	if err = s.initTlsConfig(); err != nil {
		return errors.Wrapf(err, "certificate initialization failed.")
	}

	// ?????????http
	if err = s.initHTTPServer(); err != nil {
		return errors.Wrapf(err, "http server initialization failed.")
	}

	// v1 http server ??????
	if err = s.registerV1HttpServerToRegistry(); err != nil {
		return errors.Wrapf(err, "http v1 initialization failed.")
	}

	// ?????????go-micro
	if err = s.microServer.Init(s.conf); err != nil {
		return errors.Wrapf(err, "v2 server initialization failed.")
	}

	return nil
}

// registerV1HttpServerToRegistry ??????v1 http server???Registry
func (s *StorageServer) registerV1HttpServerToRegistry() error {
	// RDiscover
	if s.conf.Etcd.Feature {
		// ??????metadata
		meta := make(map[string]string)
		// web server ?????????
		port := strconv.FormatUint(uint64(s.conf.Port), 10)
		// ipv4 ??????
		ipv4 := net.JoinHostPort(s.conf.Address, port)
		// ipv6 ???????????????????????????????????????
		if ipv6 := net.ParseIP(s.conf.IPv6Address); ipv6 != nil && !ipv6.IsLoopback() {
			// ???ipv6???????????????mata???
			meta[types.IPV6] = net.JoinHostPort(ipv6.String(), port)
		}

		// use go-micro v4 registry
		eoption := &registryv4.Options{
			Name:         constants.ServerName,
			Version:      version.BcsVersion,
			RegistryAddr: strings.Split(s.conf.Etcd.Address, ","),
			RegAddr:      ipv4,
			Config:       s.etcdTLSConfig,
			Meta:         meta,
		}
		blog.Infof("#############storage turn on etcd registry feature, options %+v ###############", eoption)
		s.etcdRegistry = registryv4.NewEtcdRegistry(eoption)
	}

	return nil
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

// initTlsConfig ????????? tls ??????
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
		return errors.Wrapf(err, "storage loading etcd registry tls config failed.")
	}
	// set etcd tls
	s.microServer.SetEtcdTLSConfig(s.etcdTLSConfig)

	/*
	   init micro server and client tls
	*/
	// loading client tls configuration
	cliConfig, err := ssl.ClientTslConfVerity(
		s.conf.ClientCert.CAFile,
		s.conf.ClientCert.CertFile,
		s.conf.ClientCert.KeyFile,
		s.conf.ClientCert.CertPwd,
	)
	if err != nil {
		return errors.Wrapf(err, "loading client side tls configuration failed.")
	}
	s.microServer.SetClientTLSConfig(cliConfig)

	// loading server tls configuration
	svrConfig, err := ssl.ServerTslConfVerityClient(
		s.conf.ServerCert.CAFile,
		s.conf.ServerCert.CertFile,
		s.conf.ServerCert.KeyFile,
		s.conf.ServerCert.CertPwd,
	)
	if err != nil {
		return errors.Wrapf(err, "loading server side tls config failed.")
	}
	s.microServer.SetServerTLSConfig(svrConfig)

	return nil
}

// initHTTPServer ?????????http server
func (s *StorageServer) initHTTPServer() error {

	// ApiResource
	a := apiserver.GetAPIResource()
	a.SetConfig(s.conf)
	a.InitActions()

	// register middleware
	filterFunctions := s.initFilterFunctions()

	// Api v1
	s.httpServer.RegisterWebServer(actions.PathV1, filterFunctions, a.ActionsV1)

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
	s.httpServer.RegisterWebServer("", nil, action)
}

// Start to run storage server
func (s *StorageServer) Start() error {
	chErr := make(chan error, 1)

	defer func() {
		// ????????????
		if s.conf.Etcd.Feature {
			s.etcdRegistry.Deregister()
		}
		// ????????????
		s.close()
	}()

	go func() {
		// run micro server
		chErr <- s.microServer.Run()
	}()

	go func() {
		// run http server
		blog.Info("run v1 http server")
		s.httpServer.SetAddressIPv6(s.conf.IPv6Address) // ???ipv6??????????????????httpServer???
		err := s.httpServer.ListenAndServe()
		chErr <- errors.Wrapf(err, "http listen and service failed.")
	}()

	runPrometheusMetrics(s.conf)

	// startDaemon
	actions.StartActionDaemon()

	// register and discover
	if s.conf.Etcd.Feature {
		if err := s.etcdRegistry.Register(); err != nil {
			chErr <- errors.Wrapf(err, "storage etcd registry failed.")
		}
	}

	select {
	case err := <-chErr:
		return errors.Wrapf(err, "exit!")
	}
}

// close ??????bcs-storage
func (s *StorageServer) close() {
	// ?????? micro server
	s.cancelFunc()
	// ?????? http server
	s.httpServer.Close()
}

// runPrometheusMetrics starting prometheus metrics handler
func runPrometheusMetrics(op *options.StorageOptions) {
	http.Handle("/metrics", promhttp.Handler())
	// ipv4 ipv6
	ips := []string{op.Address, op.IPv6Address}
	ipv6Server := ipv6server.NewIPv6Server(ips, strconv.Itoa(int(op.MetricPort)), "", nil)
	// ??????server???????????????v4???v6??????
	go ipv6Server.ListenAndServe()
}

func getRouteFunc(f http.HandlerFunc) restful.RouteFunction {
	return restful.RouteFunction(func(req *restful.Request, resp *restful.Response) {
		f(resp, req.Request)
	})
}
