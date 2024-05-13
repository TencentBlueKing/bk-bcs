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

package app

import (
	"fmt"
	"math"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pbcloudagent "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetagent"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netagent/internal/deviceplugin"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netagent/internal/inspector"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netagent/internal/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/apimetric"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/grpclb"
	bcsclientset "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/clientset/versioned"
	cloudv1set "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/clientset/versioned/typed/cloud/v1"
	listercloudv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/listers/cloud/v1"

	"google.golang.org/grpc"
	"k8s.io/client-go/kubernetes"
	k8score "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Server server for cloud net agent
type Server struct {
	option *options.NetAgentOption

	inspector *inspector.NodeNetworkInspector

	cloudNetClient pbcloudnet.CloudNetserviceClient

	k8sClient k8score.CoreV1Interface

	k8sIPClient cloudv1set.CloudV1Interface

	k8sIPLister listercloudv1.CloudIPLister

	metricCollector *apimetric.Collector

	devicePluginOp *deviceplugin.DevicePluginOp

	fixedIPWorkloads []string
}

// New create server
func New(option *options.NetAgentOption) *Server {
	option.FixedIPWorkloads = strings.Replace(option.FixedIPWorkloads, ";", ",", -1)
	fixedIPWorkloads := strings.Split(option.FixedIPWorkloads, ",")
	return &Server{
		option:           option,
		fixedIPWorkloads: fixedIPWorkloads,
	}
}

func (s *Server) initCloudNetClient() error {
	cloudNetserviceEndpointsStr := strings.Replace(s.option.CloudNetserviceEndpoints, ";", ",", -1)
	cloudNetserviceEndpoints := strings.Split(cloudNetserviceEndpointsStr, ",")

	conn, err := grpc.Dial(
		"",
		grpc.WithInsecure(),
		grpc.WithBalancer(grpc.RoundRobin(grpclb.NewPseudoResolver(cloudNetserviceEndpoints))),
		grpc.WithBlock(),
	)
	if err != nil {
		blog.Errorf("init cloud netservice client failed, err %s", err.Error())
		return err
	}
	cloudNetClient := pbcloudnet.NewCloudNetserviceClient(conn)
	s.cloudNetClient = cloudNetClient
	return nil
}

func (s *Server) initInspector() error {
	s.inspector = inspector.New(s.option, s.cloudNetClient, s.devicePluginOp)
	if err := s.inspector.Init(); err != nil {
		return err
	}
	return nil
}

func (s *Server) initK8SClient() error {
	var restConfig *rest.Config
	var err error
	if len(s.option.Kubeconfig) == 0 {
		blog.Infof("access kube-apiserver using incluster mod")
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			blog.Errorf("get incluster config failed, err %s", err.Error())
			return fmt.Errorf("get incluster config failed, err %s", err.Error())
		}
	} else {
		blog.Infof("access kube-apiserver using kubeconfig %s", s.option.Kubeconfig)
		restConfig, err = clientcmd.BuildConfigFromFlags("", s.option.Kubeconfig)
		if err != nil {
			blog.Errorf("create internal client with kubeconfig %s failed, err %s", s.option.Kubeconfig, err.Error())
			return fmt.Errorf("create internal client with kubeconfig %s failed, err %s", s.option.Kubeconfig, err.Error())
		}
	}

	k8sClientSet, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("k8s NewForConfig failed err %s", err.Error())
		return fmt.Errorf("k8s NewForConfig failed err %s", err.Error())
	}

	bcsClientSet, err := bcsclientset.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("NewForConfig failed, err %s", err.Error())
		return fmt.Errorf("NewForConfig failed, err %s", err.Error())
	}

	s.k8sClient = k8sClientSet.CoreV1()
	s.k8sIPClient = bcsClientSet.CloudV1()
	return nil
}

func (s *Server) initMetrics() error {
	blog.Infof("init metrics handler")
	mux := http.NewServeMux()
	metricEndpoint := s.option.Address + ":" + strconv.Itoa(int(s.option.MetricPort))
	metricCollector := apimetric.NewCollector(metricEndpoint, "/metrics")
	metricCollector.Init("bcs_network", "agent")
	metricCollector.RegisterMux(mux)

	metricServer := &http.Server{
		Addr:    metricEndpoint,
		Handler: mux,
	}
	go func() {
		blog.Infof("start metrics and pprof server")
		err := metricServer.ListenAndServe()
		if err != nil {
			blog.Fatalf("metric server Listen failed, err %s", err.Error())
		}
	}()

	s.metricCollector = metricCollector
	return nil
}

func (s *Server) initDevicePluginServer() {
	if s.option.UseDevicePlugin {
		blog.Infof("init device plugin server")
		s.devicePluginOp = deviceplugin.NewDevicePluginOp(
			s.option.KubeletSockPath, s.option.DevicePluginSockPath, s.option.DevicePluginResourceName)
		go s.devicePluginOp.Start()
	}
}

// Init init server
func (s *Server) Init() {
	if err := s.initMetrics(); err != nil {
		blog.Fatalf("init metric Collector failed, err %s", err.Error())
	}
	if err := s.initK8SClient(); err != nil {
		blog.Fatalf("init k8s client failed, err %s", err.Error())
	}
	if err := s.initCloudNetClient(); err != nil {
		blog.Fatalf("init cloud netservice client, err %s", err.Error())
	}
	s.initDevicePluginServer()
	if err := s.initInspector(); err != nil {
		blog.Fatalf("init Inspector failed, err %s", err.Error())
	}

}

// Run run server
func (s *Server) Run() {

	lis, err := net.Listen("tcp",
		s.option.Address+":"+strconv.Itoa(int(s.option.Port)))
	if err != nil {
		blog.Fatalf("listen on endpoint failed, err %s", err.Error())
	}

	// run grpc server
	grpcServer := grpc.NewServer(grpc.MaxRecvMsgSize(math.MaxInt32))
	pbcloudagent.RegisterCloudNetagentServer(grpcServer, s)
	blog.Infof("registered cloud netagent grpc server")

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			blog.Fatalf("start grpc server with net listener failed, err %s", err.Error())
		}
	}()

	interupt := make(chan os.Signal, 10)
	signal.Notify(interupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM,
		syscall.SIGUSR1, syscall.SIGUSR2)
	for {
		select {
		case <-interupt:
			grpcServer.GracefulStop()
			blog.Infof("Get signal from system. Exit\n")
			return
		}
	}
}
