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

package api

import (
	"crypto/tls"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"net"
	"net/http"
	"strconv"
	"time"

	restful "github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "netservice",
		Subsystem: "api",
		Name:      "request_total",
		Help:      "The total number of requests to netservice http api",
	}, []string{"handler", "status"})
	requestLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "netservice",
		Subsystem: "api",
		Name:      "request_latency_seconds",
		Help:      "BCS netservice api request latency statistic.",
		Buckets:   []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.0, 3.0},
	}, []string{"handler", "status"})
	//SUCCESS success string for response
	SUCCESS = "success"
)

func init() {
	//add golang basic metrics
	//prometheus.MustRegister(prometheus.NewGoCollector())
	prometheus.MustRegister(requestsTotal)
	prometheus.MustRegister(requestLatency)
}

func reportMetrics(handler, status string, started time.Time) {
	requestsTotal.WithLabelValues(handler, status).Inc()
	requestLatency.WithLabelValues(handler, status).Observe(time.Since(started).Seconds())
}

//RegisterMetrics register http metrics for netservice
func RegisterMetrics(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(addr, nil)
}

//NewHTTPService create http service for net-service
func NewHTTPService(addr string, port int) *HTTPService {
	service := &HTTPService{
		container: restful.NewContainer(),
		addr:      addr,
		port:      port,
	}
	return service
}

//HTTPService http container for register api for net-service
type HTTPService struct {
	listener  net.Listener       //listener for graceful stop
	container *restful.Container //root container for all url route
	addr      string             //address:port for http service
	port      int                //listen port
}

//Register register restful web service
func (svr *HTTPService) Register(wb *restful.WebService) {
	if wb != nil {
		svr.container.Add(wb)
	}
}

//ListenAndServe http listen
func (svr *HTTPService) ListenAndServe() error {
	listenAddr := svr.addr + ":" + strconv.Itoa(svr.port)
	blog.Info("HTTPService ready to Listen %s", listenAddr)
	var err error
	svr.listener, err = net.Listen("tcp", listenAddr)
	if err != nil {
		blog.Error("HTTPService Listen %s failed: %s", listenAddr, err.Error())
		return err
	}
	return http.Serve(svr.listener, svr.container)
}

//ListenAndServeTLS https listen
//param certFile: root ca cert
//param keyFile: http server private key file
//param pubFile: http server public key file
//param passwd: if private key encrypted, passwd required
func (svr *HTTPService) ListenAndServeTLS(certFile, keyFile, pubFile, passwd string) error {
	//loading root certificate
	config, err := ssl.ServerTslConfVerityClient(certFile, pubFile, keyFile, passwd)
	if err != nil {
		blog.Error("HTTPService load SSL config err, %v", err)
		return err
	}
	config.BuildNameToCertificate()

	//https listen
	listenAddr := svr.addr + ":" + strconv.Itoa(svr.port)
	blog.Info("HTTPService https Listen %s", listenAddr)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		blog.Error("HTTPService https listen %s failed: %s", listenAddr, err.Error())
		return err
	}
	svr.listener = tls.NewListener(ln, config)
	return http.Serve(svr.listener, svr.container)
}

//Stop stop http service
func (svr *HTTPService) Stop(sec int) {
	if err := svr.listener.Close(); err != nil {
		blog.Errorf("HTTPService close failed, %s", err.Error())
	}
}
