/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	bcsclientset "github.com/Tencent/bk-bcs/bcs-mesos/mesosv2/generated/clientset/versioned"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-networkpolicy/controller"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-networkpolicy/datainformer"
	infrk8s "github.com/Tencent/bk-bcs/bcs-network/bcs-networkpolicy/datainformer/kubernetes"
	infrmesos "github.com/Tencent/bk-bcs/bcs-network/bcs-networkpolicy/datainformer/mesos"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-networkpolicy/options"

	restful "github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Server server for network policy
type Server struct {
	opt                 *options.NetworkPolicyOption
	httpServer          *httpserver.HttpServer
	infr                datainformer.Interface
	netPolicyController *controller.NetworkPolicyController
}

// New create server
func New(opt *options.NetworkPolicyOption) *Server {
	return &Server{
		opt: opt,
	}
}

// Init create datainformer and netpolicy controller
func (s *Server) Init() error {
	var err error
	var clientconfig *rest.Config
	if len(s.opt.Kubeconfig) != 0 {
		clientconfig, err = clientcmd.BuildConfigFromFlags(s.opt.KubeMaster, s.opt.Kubeconfig)
		if err != nil {
			return fmt.Errorf("build configuration from %s, %s failed, err %s",
				s.opt.KubeMaster, s.opt.Kubeconfig, err.Error())
		}
	} else {
		clientconfig, err = rest.InClusterConfig()
		if err != nil {
			return fmt.Errorf("init inCluster config failed, err %s", err.Error())
		}
	}

	clientset, err := kubernetes.NewForConfig(clientconfig)
	if err != nil {
		return fmt.Errorf("create client set failed, err %s", err.Error())
	}

	bcsClientset, err := bcsclientset.NewForConfig(clientconfig)
	if err != nil {
		return fmt.Errorf("create bcs client set failed, err %s", err.Error())
	}

	var infr datainformer.Interface
	switch s.opt.ServiceRegistry {
	case options.ServiceRegistryKubernetes:
		kInfr := infrk8s.New(s.opt)
		kInfr.Init(clientset)
		infr = kInfr
	case options.ServiceRegistryMesos:
		mInfr := infrmesos.New(s.opt)
		mInfr.Init(clientset, bcsClientset)
		infr = mInfr
	}

	npc, err := controller.NewNetworkPolicyController(clientset, infr, s.opt)
	if err != nil {
		blog.Errorf("create network policy controller failed, err %s", err.Error())
		return fmt.Errorf("create network policy controller failed, err %s", err.Error())
	}

	infr.AddPodEventHandler(npc.PodEventHandler)
	infr.AddNamespaceEventHandler(npc.NamespaceEventHandler)
	infr.AddNetworkpolicyEventHandler(npc.NetworkPolicyEventHandler)

	s.infr = infr
	s.netPolicyController = npc

	// init http server
	httpServer := httpserver.NewHttpServer(s.opt.Port, s.opt.Address, "")
	if len(s.opt.CAFile) != 0 || len(s.opt.ServerCertFile) != 0 || len(s.opt.ServerKeyFile) != 0 {
		httpServer.SetSsl(s.opt.CAFile, s.opt.ServerCertFile, s.opt.ServerKeyFile, static.ServerCertPwd)
	}
	httpServer.GetWebContainer().Handle("/metrics", promhttp.Handler())
	if s.opt.Debug {
		debugActions := []*httpserver.Action{
			httpserver.NewAction("GET", "/debug/pprof/", nil, getRouteFunc(pprof.Index)),
			httpserver.NewAction("GET", "/debug/pprof/{uri:*}", nil, getRouteFunc(pprof.Index)),
			httpserver.NewAction("GET", "/debug/pprof/cmdline", nil, getRouteFunc(pprof.Cmdline)),
			httpserver.NewAction("GET", "/debug/pprof/profile", nil, getRouteFunc(pprof.Profile)),
			httpserver.NewAction("GET", "/debug/pprof/symbol", nil, getRouteFunc(pprof.Symbol)),
			httpserver.NewAction("GET", "/debug/pprof/trace", nil, getRouteFunc(pprof.Trace)),
		}
		httpServer.RegisterWebServer("", nil, debugActions)
	}

	s.httpServer = httpServer

	return nil
}

// Run run the server
func (s *Server) Run(stopCh chan struct{}) error {

	// start policy controller
	// WaitGroup is needed because network policy controller needs WaitGroup
	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		go s.netPolicyController.Run(stopCh, &wg)
		wg.Wait()
	}()

	// start data informer
	if err := s.infr.Run(); err != nil {
		return fmt.Errorf("start data informer failed, err %s", err.Error())
	}

	// start http server
	if err := s.httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("listen failed, err %s", err.Error())
	}

	return nil
}

// Stop stop the server
func (s *Server) Stop() {
	blog.Infof("stop server")
	s.infr.Stop()
}

func getRouteFunc(f http.HandlerFunc) restful.RouteFunction {
	return restful.RouteFunction(func(req *restful.Request, resp *restful.Response) {
		f(resp, req.Request)
	})
}
