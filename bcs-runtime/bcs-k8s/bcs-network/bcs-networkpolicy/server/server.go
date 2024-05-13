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

package server

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	bcsclientset "github.com/Tencent/bk-bcs/bcs-mesos/mesosv2/generated/clientset/versioned"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/controller"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/controller/networkpolicy"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/controller/podpolicy"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/datainformer"
	infrk8s "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/datainformer/kubernetes"
	infrmesos "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/datainformer/mesos"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/options"
	"github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Server server for network policy
type Server struct {
	opt              *options.NetworkPolicyOption
	httpServer       *httpserver.HttpServer
	infr             datainformer.Interface
	policyController controller.Controller
}

// New create server
func New(opt *options.NetworkPolicyOption) *Server {
	return &Server{
		opt: opt,
	}
}

// Init create dataInformer and networkPolicy controller
func (s *Server) Init() error {
	clientSet, bcsClientSet, err := s.buildClientSet()
	if err != nil {
		return err
	}

	var dataInformer datainformer.Interface
	switch s.opt.ServiceRegistry {
	case options.ServiceRegistryKubernetes:
		k8sInformer := infrk8s.New(s.opt)
		k8sInformer.Init(clientSet)
		dataInformer = k8sInformer
	case options.ServiceRegistryMesos:
		mesosInformer := infrmesos.New(s.opt)
		mesosInformer.Init(clientSet, bcsClientSet)
		dataInformer = mesosInformer
	default:
		return fmt.Errorf("unknown serviceRegistry '%s'", s.opt.ServiceRegistry)
	}
	blog.Infof("Using service registry: %s", s.opt.ServiceRegistry)

	var npc controller.Controller
	switch s.opt.WorkMode {
	case options.WorkModeGlobal:
		npc, err = networkpolicy.NewNetworkPolicyController(clientSet, dataInformer, s.opt)
	case options.WorkModePod:
		npc, err = podpolicy.NewPodPolicyController(clientSet, dataInformer, s.opt)
	default:
		return fmt.Errorf("unknown workMode '%s'", s.opt.WorkMode)
	}
	if err != nil {
		err = fmt.Errorf("create network policy controller failed, err: %s", err.Error())
		return err
	}
	blog.Infof("Using workMode: %s", s.opt.ServiceRegistry)

	dataInformer.AddPodEventHandler(npc.GetPodEventHandler())
	dataInformer.AddNamespaceEventHandler(npc.GetNamespaceEventHandler())
	dataInformer.AddNetworkpolicyEventHandler(npc.GetNetworkPolicyEventHandler())

	s.infr = dataInformer
	s.policyController = npc

	// init http server
	s.httpServer = s.buildHttpServer()

	return nil
}

// buildClientSet return kubernetes and bcs clientSet
func (s *Server) buildClientSet() (client kubernetes.Interface, bcsClient bcsclientset.Interface, err error) {
	var clientConfig *rest.Config
	if len(s.opt.Kubeconfig) != 0 {
		clientConfig, err = clientcmd.BuildConfigFromFlags(s.opt.KubeMaster, s.opt.Kubeconfig)
		if err != nil {
			return client, bcsClient, fmt.Errorf("build configuration from %s, %s failed, err %s",
				s.opt.KubeMaster, s.opt.Kubeconfig, err.Error())
		}
	} else {
		clientConfig, err = rest.InClusterConfig()
		if err != nil {
			return client, bcsClient, fmt.Errorf("init inCluster config failed, err %s", err.Error())
		}
	}

	client, err = kubernetes.NewForConfig(clientConfig)
	if err != nil {
		return client, bcsClient, fmt.Errorf("create client set failed, err %s", err.Error())
	}

	bcsClient, err = bcsclientset.NewForConfig(clientConfig)
	if err != nil {
		return client, bcsClient, fmt.Errorf("create bcs client set failed, err %s", err.Error())
	}
	return client, bcsClient, nil
}

// buildHttpServer return httpServer object
func (s *Server) buildHttpServer() *httpserver.HttpServer {
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
		_ = httpServer.RegisterWebServer("", nil, debugActions)
	}
	return httpServer
}

// Run run the server
func (s *Server) Run(stopCh chan struct{}) error {
	// Start data informer
	if err := s.infr.Run(); err != nil {
		return fmt.Errorf("start data informer failed, err %s", err.Error())
	}
	blog.Infof("DataInformer is started.")

	// Update dataInformer sync status of networkPolicy controller
	s.policyController.SetDataInformerSynced()

	// Start http server
	httpErr := make(chan error)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				blog.Errorf("HTTP Server occurred panic, stacktrace from panic:\n%s", string(debug.Stack()))
				httpErr <- fmt.Errorf("HTTPServer Panic")
			}
		}()

		if err := s.httpServer.ListenAndServe(); err != nil {
			httpErr <- err
		}
	}()

	// Start networkPolicy controller
	npcErr := make(chan error)
	go func() {
		defer func() {
			blog.Infof("NetworkPolicy Controller is stopped.")
			if r := recover(); r != nil {
				blog.Errorf("NetworkPolicy occurred panic, stacktrace:\n%s", string(debug.Stack()))
				npcErr <- fmt.Errorf("NetworkPolicy Controller Panic")
			}
		}()

		// WaitGroup is needed because network_policy controller needs WaitGroup
		var wg sync.WaitGroup
		wg.Add(1)
		err := s.policyController.Run(stopCh, &wg)
		wg.Wait()

		npcErr <- err
	}()

	// if httpServer or npcContainer is closed, finish the server
	select {
	case e := <-httpErr:
		{
			if e != nil {
				blog.Errorf("HttpServer stopped with error.")
				return e
			}
			blog.Infof("HttpServer is stopped.")
			return nil
		}
	case e := <-npcErr:
		{
			if e != nil {
				blog.Errorf("NetworkPolicyController stopped with error.")
				return e
			}
			blog.Infof("NetworkPolicyController is stopped.")
			return nil
		}
	}
}

// Stop stop the server
func (s *Server) Stop() {
	blog.Infof("Stop data informer.")
	s.infr.Stop()
}

func getRouteFunc(f http.HandlerFunc) restful.RouteFunction {
	return func(req *restful.Request, resp *restful.Response) {
		f(resp, req.Request)
	}
}
