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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"
	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/controllers"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanagerv1"

	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"k8s.io/client-go/tools/clientcmd"
	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-micro/v2/service/grpc"
	rawgrpc "google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = meshv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	//conf := config.Config{}
	/*flag.StringVar(&conf.MetricsPort, "metric-port", "9443", "The address the metric endpoint binds to.")
	flag.StringVar(&conf.DockerHub, "istio-docker-hub", "", "istio-operator docker hub")
	flag.StringVar(&conf.IstioOperatorCharts, "istiooperator-charts", "", "istio-operator charts")
	flag.StringVar(&conf.ServerAddress, "apigateway-addr", "", "apigateway address")
	flag.StringVar(&conf.UserToken, "user-token", "", "apigateway usertoken to control k8s cluster")
	flag.StringVar(&conf.Address, "address", "127.0.0.1", "server address")
	flag.IntVar(&conf.Port, "port", 8899, "grpc server port")
	flag.StringVar(&conf.EtcdCaFile, "etcd-cafile", "", "SSL Certificate Authority file used to secure etcd communication")
	flag.StringVar(&conf.EtcdCertFile, "etcd-certfile", "", "SSL certification file used to secure etcd communication")
	flag.StringVar(&conf.EtcdKeyFile, "etcd-keyfile", "", "SSL key file used to secure etcd communication")
	flag.StringVar(&conf.EtcdServers, "etcd-servers", "", "List of etcd servers to connect with (scheme://ip:port), comma separated")
	flag.StringVar(&conf.ServerCaFile, "ca-file", "", "If set, any request presenting a certificate signed by one of the authorities in the ca-file is authenticated with an identity corresponding to the CommonName of the client certificate.")
	flag.StringVar(&conf.ServerCertFile, "tls-cert-file", "", "File containing the default x509 Certificate for HTTPS.")
	flag.StringVar(&conf.ServerKeyFile, "tls-private-key-file", "", "File containing the default x509 private key matching")*/
	//flag.Parse()
	conf := config.ParseConfig()
	by, _ := json.Marshal(conf)
	klog.Infof("MeshManager config(%s)", string(by))
	if conf.ServerCaFile != "" && conf.ServerCertFile != "" && conf.ServerKeyFile != "" {
		conf.IsSsl = true
		tlsConf, err := ssl.ServerTslConf(conf.ServerCaFile, conf.ServerCertFile, conf.ServerKeyFile, static.ServerCertPwd)
		if err != nil {
			klog.Errorf("ServerTslConf failed: %s", err.Error())
			os.Exit(1)
		}
		conf.TlsConf = tlsConf
	}
	kubecfg, err := clientcmd.BuildConfigFromFlags("", conf.Kubeconfig)
	if err != nil {
		klog.Errorf("build kubeconfig %s error %s", conf.Kubeconfig, err.Error())
		os.Exit(1)
	}
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	mgr, err := ctrl.NewManager(kubecfg, ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: fmt.Sprintf("%s:%s", conf.Address, conf.MetricsPort),
		/*LeaderElection:     true,
		LeaderElectionID:   "meshmanager.bkbcs.tencent.com",*/
	})
	if err != nil {
		klog.Errorf("start manager failed: %s", err.Error())
		os.Exit(1)
	}
	if err = (&controllers.MeshClusterReconciler{
		Client:       mgr.GetClient(),
		Log:          ctrl.Log.WithName("controllers").WithName("MeshCluster"),
		Scheme:       mgr.GetScheme(),
		MeshClusters: make(map[string]*controllers.MeshClusterManager),
		Conf:         conf,
	}).SetupWithManager(mgr); err != nil {
		klog.Errorf("create MeshManager controller failed: %s", err.Error())
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	go func() {
		klog.Infof("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			klog.Errorf("running manager failed: %s", err.Error())
			os.Exit(1)
		}
	}()
	//context for grpc gateway & go-micro
	ctx, cancel := context.WithCancel(context.Background())
	go signalWatch(cancel)
	//http server
	grpcAddr := fmt.Sprintf("%s:%d", conf.Address, conf.Port)
	grpcmux := grpcruntime.NewServeMux()
	opts := []rawgrpc.DialOption{rawgrpc.WithInsecure()}
	if err := meshmanagerv1.RegisterMeshManagerHandlerFromEndpoint(ctx, grpcmux, grpcAddr, opts); err != nil {
		klog.Errorf("register grpc-gateway failed, %s", err.Error())
		os.Exit(1)
	}
	// http mux
	mux := http.NewServeMux()
	mux.Handle("/", grpcmux)
	go func() {
		httpserver := &http.Server{Addr: fmt.Sprintf("%s:%d", conf.Address, conf.Port-1), Handler: mux}
		var err error
		if conf.IsSsl {
			httpserver.TLSConfig = conf.TlsConf
			err = httpserver.ListenAndServeTLS("", "")
		} else {
			err = httpserver.ListenAndServe()
		}
		if err != nil {
			klog.Errorf("ListenAndServe %s failed: %s", httpserver.Addr, err.Error())
			os.Exit(1)
		}
	}()
	//tls
	tlsConf, err := ssl.ClientTslConfVerity(conf.EtcdCaFile, conf.EtcdCertFile, conf.EtcdKeyFile, "")
	if err != nil {
		klog.Errorf("new client tsl conf failed: %s", err.Error())
		os.Exit(1)
	}
	// New Service
	regOption := func(e *registry.Options) {
		e.Addrs = strings.Split(conf.EtcdServers, ",")
		e.TLSConfig = tlsConf
	}
	sevOption := func(o *server.Options) {
		o.TLSConfig = conf.TlsConf
		o.Name = "meshmanager.bkbcs.tencent.com"
		o.Version = version.GetVersion()
		o.Context = ctx
		o.Address = grpcAddr
		o.Registry = etcd.NewRegistry(regOption)
	}
	grpcSvr := grpc.NewService()
	grpcSvr.Server().Init(sevOption)
	// Initialise service
	grpcSvr.Init()
	// Register Handler, if we need more options control
	// try formation like: handler.BcsDataManager(CustomOption)
	meshHandler := handler.NewMeshHandler(conf, mgr.GetClient())
	err = meshmanager.RegisterMeshManagerHandler(grpcSvr.Server(), meshHandler)
	if err != nil {
		klog.Errorf("RegisterMeshManagerHandler failed: %s", err.Error())
	}
	// Run service
	klog.Infof("Listen grpc server on endpoint(%s)", grpcAddr)
	if err := grpcSvr.Run(); err != nil {
		klog.Errorf("run grpc server failed: %s", err.Error())
		os.Exit(1)
	}
}

func signalWatch(stop context.CancelFunc) {
	close := make(chan os.Signal, 10)
	signal.Notify(close, syscall.SIGINT, syscall.SIGTERM)
	<-close
	fmt.Printf("bcs-gateway-dicovery catch exit signal, exit in 3 seconds...\n")
	stop()
	time.Sleep(time.Second * 3)
}
