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
	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/controllers"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanagerv1"

	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-micro/v2/service"
	"github.com/micro/go-micro/v2/service/grpc"
	rawgrpc "google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
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
		conf.TLSConf = tlsConf
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
	managerStop := make(chan struct{})
	go func() {
		klog.Infof("starting manager")
		if err := mgr.Start(managerStop); err != nil {
			klog.Errorf("running manager failed: %s", err.Error())
			os.Exit(1)
		}
	}()
	//context for grpc gateway, go-micro and controllerManager
	ctx, cancel := context.WithCancel(context.Background())

	//http server
	grpcAddr := fmt.Sprintf("%s:%d", conf.Address, conf.Port)
	grpcmux := grpcruntime.NewServeMux()
	opts := []rawgrpc.DialOption{rawgrpc.WithInsecure()}
	if err := meshmanagerv1.RegisterMeshManagerHandlerFromEndpoint(ctx, grpcmux, grpcAddr, opts); err != nil {
		klog.Errorf("register grpc-gateway failed, %s", err.Error())
		os.Exit(1)
	}
	httpserver := &http.Server{Addr: fmt.Sprintf("%s:%d", conf.Address, conf.Port-1), Handler: grpcmux}
	// http backgroup listen
	go func() {
		var err error
		if conf.IsSsl {
			httpserver.TLSConfig = conf.TLSConf
			err = httpserver.ListenAndServeTLS("", "")
		} else {
			err = httpserver.ListenAndServe()
		}
		if err != nil {
			klog.Errorf("ListenAndServe %s failed: %s", httpserver.Addr, err.Error())
			//when httpserver shutdown, wait for resource clean
			time.Sleep(time.Second * 3)
			os.Exit(1)
		}
	}()

	go signalWatch(cancel, managerStop, httpserver)

	//grpc server setting
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
	grpcSvc := grpc.NewService(
		service.Context(ctx),
		service.Name("meshmanager.bkbcs.tencent.com"),
		service.Version(version.BcsVersion),
		service.Address(grpcAddr),
		service.Registry(etcd.NewRegistry(regOption)),
		grpc.WithTLS(conf.TLSConf),
		service.RegisterInterval(time.Second*30),
		service.RegisterTTL(time.Second*40),
	)
	// Initialise service
	grpcSvc.Init()
	// Register Handler, if we need more options control
	// try formation like: handler.BcsDataManager(CustomOption)
	meshHandler := handler.NewMeshHandler(conf, mgr.GetClient())
	err = meshmanager.RegisterMeshManagerHandler(grpcSvc.Server(), meshHandler)
	if err != nil {
		klog.Errorf("RegisterMeshManagerHandler failed: %s", err.Error())
	}
	// Run service
	klog.Infof("Listen grpc server on endpoint(%s)", grpcAddr)
	if err := grpcSvc.Run(); err != nil {
		klog.Errorf("run grpc server failed: %s", err.Error())
		os.Exit(1)
	}
}

func signalWatch(stop context.CancelFunc, manager chan struct{}, htpSvr *http.Server) {
	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	<-signalCh
	fmt.Printf("bcs-gateway-discovery catch exit signal, exit in 3 seconds...\n")
	close(manager)
	stop()
	htpSvr.Shutdown(context.Background())
	time.Sleep(time.Second * 3)
}
