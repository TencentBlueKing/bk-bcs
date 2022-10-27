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
	"net"
	"net/http"
	"net/http/pprof"
	"strconv"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/ipv6server"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
	trestful "github.com/Tencent/bk-bcs/bcs-common/pkg/tracing/restful"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/utils/metrics"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/actions/utils/middle"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-storage/storage/apiserver"
	"github.com/opentracing/opentracing-go"

	restful "github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StorageServer is a data struct of bcs storage server
type StorageServer struct {
	conf         *options.StorageOptions
	httpServer   *httpserver.HttpServer
	etcdRegistry registry.Registry
}

// NewStorageServer create storage server object
func NewStorageServer(op *options.StorageOptions) (*StorageServer, error) {
	s := &StorageServer{}

	// Configuration
	s.conf = op

	// Http server
	s.httpServer = httpserver.NewHttpServer(s.conf.Port, s.conf.Address, "")
	if s.conf.ServerCert.IsSSL {
		s.httpServer.SetSsl(
			s.conf.ServerCert.CAFile,
			s.conf.ServerCert.CertFile,
			s.conf.ServerCert.KeyFile,
			s.conf.ServerCert.CertPwd)
	}

	// RDiscover
	if s.conf.Etcd.Feature {
		tlsCfg, err := s.conf.Etcd.GetTLSConfig()
		if err != nil {
			blog.Errorf("storage loading etcd registry tls config failed, %s", err.Error())
			return nil, err
		}
		// 创建Meta
		meta := make(map[string]string)
		// 端口号
		port := strconv.FormatUint(uint64(s.conf.Port), 10)
		// ipv4地址
		ipv4 := net.JoinHostPort(s.conf.Address, port)
		// ipv6注册地址不能是本地回环地址
		if ipv6 := net.ParseIP(s.conf.IPv6Address); ipv6 != nil && !ipv6.IsLoopback() {
			// 把ipv6地址写入到mata中
			meta[types.IPV6] = net.JoinHostPort(ipv6.String(), port)
		}
		// init go-micro registry
		eoption := &registry.Options{
			Name:         "storage.bkbcs.tencent.com",
			Version:      version.BcsVersion,
			RegistryAddr: strings.Split(s.conf.Etcd.Address, ","),
			RegAddr:      ipv4,
			Config:       tlsCfg,
			Meta:         meta,
		}
		blog.Infof("#############storage turn on etcd registry feature, options %+v ###############", eoption)
		s.etcdRegistry = registry.NewEtcdRegistry(eoption)
	}

	// ApiResource
	a := apiserver.GetAPIResource()
	a.SetConfig(op)
	a.InitActions()

	return s, nil
}

func (s *StorageServer) initFilterFunctions() []restful.FilterFunction {
	filterFunctions := []restful.FilterFunction{}

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

func (s *StorageServer) initHTTPServer() error {
	a := apiserver.GetAPIResource()

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

	s.initHTTPServer()

	go func() {
		s.httpServer.SetAddressIPv6(s.conf.IPv6Address) // 把ipv6地址，加入到httpServer中
		err := s.httpServer.ListenAndServe()
		blog.Errorf("http listen and service failed! err:%s", err.Error())
		chErr <- err
	}()

	runPrometheusMetrics(s.conf)

	// startDaemon
	actions.StartActionDaemon()

	// register and discover
	if s.conf.Etcd.Feature {
		if err := s.etcdRegistry.Register(); err != nil {
			blog.Errorf("storage etcd registry failed, %s", err.Error())
			chErr <- err
		}
	}

	select {
	case err := <-chErr:
		blog.Errorf("exit! err:%s", err.Error())
		if s.conf.Etcd.Feature {
			s.etcdRegistry.Deregister()
		}
		return err
	}
}

// runPrometheusMetrics starting prometheus metrics handler
func runPrometheusMetrics(op *options.StorageOptions) {
	http.Handle("/metrics", promhttp.Handler())
	// ipv4 ipv6
	ips := []string{op.Address, op.IPv6Address}
	ipv6Server := ipv6server.NewIPv6Server(ips, strconv.Itoa(int(op.MetricPort)), "", nil)
	// 启动server，同时监听v4、v6地址
	go ipv6Server.ListenAndServe()
}

func getRouteFunc(f http.HandlerFunc) restful.RouteFunction {
	return restful.RouteFunction(func(req *restful.Request, resp *restful.Response) {
		f(resp, req.Request)
	})
}
