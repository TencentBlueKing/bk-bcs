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
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/cloud-network-agent/controller"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/cloud-network-agent/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/pkg/eni"
	eniaws "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/pkg/eni/aws"
	eniqcloud "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/pkg/eni/qcloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/pkg/netservice"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/pkg/networkutil"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloudnetwork/pkg/nodenetwork"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server server for cloud network agent
type Server struct {
	instanceEth   string
	instanceIP    string
	hostname      string
	netutil       *networkutil.NetUtil
	cloudClient   eni.Interface
	nodeNetClient nodenetwork.Interface
	netsvcClient  netservice.Interface
	controller    *controller.NetworkController
	metricServer  *httpserver.HttpServer
	opt           *options.NetworkOption
	wg            sync.WaitGroup
}

// New create server
func New(opt *options.NetworkOption) *Server {
	return &Server{
		opt: opt,
	}
}

// init network util
func (s *Server) initNetworkUtil() error {
	// create net util
	netutil := new(networkutil.NetUtil)
	// get host ip
	ifacesStr := strings.Replace(s.opt.Ifaces, ";", ",", -1)
	ifaces := strings.Split(ifacesStr, ",")
	instanceIP, instanceEth, err := netutil.GetAvailableHostIP(ifaces)
	if err != nil {
		blog.Errorf("get node ip failed, err %s", err.Error())
		return fmt.Errorf("get node ip failed, err %s", err.Error())
	}
	// get hostname
	hostName, err := netutil.GetHostName()
	if err != nil {
		blog.Errorf("get hostname failed, err %s", err.Error())
		return fmt.Errorf("get node ip failed, err %s", err.Error())
	}

	s.netutil = netutil
	s.instanceEth = instanceEth
	s.instanceIP = instanceIP
	s.hostname = hostName
	return nil
}

// init crd client
func (s *Server) initNodeNetworkClient() error {
	nodeNetClient := nodenetwork.New(
		s.opt.Kubeconfig,
		s.opt.KubeResyncPeriod,
		s.opt.KubeCacheSyncTimeout)

	if err := nodeNetClient.Init(); err != nil {
		blog.Errorf("init node network client failed, err %s", err.Error())
		return fmt.Errorf("init node network client failed, err %s", err.Error())
	}

	s.nodeNetClient = nodeNetClient
	return nil
}

// init cloud client
func (s *Server) initCloudClient() error {
	var client eni.Interface
	switch s.opt.Cloud {
	case options.CloudAWS:
		blog.Infof("create aws cloud client")
		client = eniaws.New(s.instanceIP)
	case options.CloudTencent:
		blog.Infof("create qcloud cloud client")
		client = eniqcloud.New(s.instanceIP)
	default:
		return fmt.Errorf("invalid cloud %s", s.opt.Cloud)
	}

	s.cloudClient = client
	return nil
}

// initNetServiceClient
func (s *Server) initNetServiceClient() error {
	netsvcClient := netservice.New(
		s.opt.NetServiceZookeeper,
		s.opt.NetServiceKey,
		s.opt.NetServiceCert,
		s.opt.NetServiceCa,
	)
	err := netsvcClient.Init()
	if err != nil {
		return err
	}

	s.netsvcClient = netsvcClient
	return nil
}

// init metric collector
func (s *Server) initMetric() error {

	metricServer := httpserver.NewHttpServer(s.opt.MetricPort, s.instanceIP, "")
	metricServer.GetWebContainer().Handle("/metrics", promhttp.Handler())

	go func() {
		if err := metricServer.ListenAndServe(); err != nil {
			blog.Warnf("metric server serve failed, err %s", err.Error())
		}
	}()

	s.metricServer = metricServer
	return nil
}

// init network controller
func (s *Server) initNetworkController() error {
	controller := controller.New(
		s.instanceEth,
		s.hostname,
		s.opt,
		s.netsvcClient,
		s.nodeNetClient,
		s.cloudClient,
		s.netutil)

	s.nodeNetClient.Register(controller)

	if err := s.nodeNetClient.Run(); err != nil {
		blog.Errorf("run node network client failed, err %s", err.Error())
		return fmt.Errorf("run node network client failed, err %s", err.Error())
	}

	if err := controller.Init(); err != nil {
		blog.Errorf("init network controller failed, err %s", err.Error())
		return fmt.Errorf("init network controller failed, err %s", err.Error())
	}

	s.controller = controller
	return nil
}

// Init init server
func (s *Server) Init() error {

	if err := s.initNetworkUtil(); err != nil {
		blog.Fatalf("init network util failed, err %s", err.Error())
	}

	if err := s.initCloudClient(); err != nil {
		blog.Fatalf("init cloud client failed, err %s", err.Error())
	}

	if err := s.initNodeNetworkClient(); err != nil {
		blog.Fatalf("init node network config client, err %s", err.Error())
	}

	if err := s.initNetServiceClient(); err != nil {
		blog.Fatalf("init netservice failed, err %s", err.Error())
	}

	if err := s.initNetworkController(); err != nil {
		blog.Fatalf("init network controller, err %s", err.Error())
	}

	if err := s.initMetric(); err != nil {
		// just throw warning when metric init failed
		blog.Warnf("init metric failed, err %s", err.Error())
	}

	return nil
}

// Run run server
func (s *Server) Run() {

	ctx, cancel := context.WithCancel(context.Background())
	s.wg.Add(1)
	go s.controller.Run(ctx, &s.wg)

	interupt := make(chan os.Signal, 10)
	signal.Notify(interupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM,
		syscall.SIGUSR1, syscall.SIGUSR2)
	for {
		select {
		case <-interupt:
			blog.Infof("Get signal from system. Exit\n")
			cancel()
			s.wg.Wait()
			return
		}
	}
}
