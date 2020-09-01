/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and

limitations under the License.
*/
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	meshv1 "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/api/v1"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/controllers"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/meshmicro"

	"k8s.io/klog"
	rawgrpc "google.golang.org/grpc"
	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-micro/v2/service"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/service/grpc"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
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
	var enableLeaderElection bool
	conf := config.Config{}
	flag.StringVar(&conf.MetricsPort, "metrics-addr", ":9443", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", true,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&conf.DockerHub, "istio-docker-hub", "", "istio-operator docker hub")
	flag.StringVar(&conf.IstioOperatorCharts, "istiooperator-charts", "", "istio-operator charts")
	flag.StringVar(&conf.ServerAddress, "apigateway-addr", "", "apigateway address")
	flag.StringVar(&conf.UserToken, "user-token", "", "apigateway usertoken to control k8s cluster")
	flag.StringVar(&conf.Address, "address", "127.0.0.1", "server address")
	flag.IntVar(&conf.Port, "port", 8899, "server port")
	flag.StringVar(&conf.EtcdCaFile, "etcd-cafile", "", "SSL Certificate Authority file used to secure etcd communication")
	flag.StringVar(&conf.EtcdCertFile, "etcd-certfile", "", "SSL certification file used to secure etcd communication")
	flag.StringVar(&conf.EtcdKeyFile, "etcd-keyfile", "", "SSL key file used to secure etcd communication")
	flag.Parse()
	by,_ := json.Marshal(conf)
	klog.Infof("MeshManager config(%s)", string(by))
	conf.ServerAddress = "http://9.143.0.40:31000/tunnels/clusters/BCS-K8S-15091/"
	conf.UserToken = "mCdfmlzonNPiAeWhANX1nj91ouBeQckQ"
	conf.IstioOperatorCharts = "./istio-operator"
	conf.EtcdCertFile = "/data/bcs/cert/k8s/bcs-etcd.pem"
	conf.EtcdCaFile = "/data/bcs/cert/k8s/etcd-ca.pem"
	conf.EtcdKeyFile = "/data/bcs/cert/k8s/bcs-etcd-key.pem"

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: fmt.Sprintf("%s:%s", conf.Address, conf.MetricsPort),
		LeaderElection:     true,
		LeaderElectionID:   "meshmanager.bkbcs.tencent.com",
	})
	if err != nil {
		klog.Errorf("start manager failed: %s", err.Error())
		os.Exit(1)
	}
	if err = (&controllers.MeshClusterReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("MeshCluster"),
		Scheme: mgr.GetScheme(),
		MeshClusters: make(map[string]*controllers.MeshClusterManager),
		Conf: conf,
	}).SetupWithManager(mgr); err != nil {
		klog.Errorf("create MeshManager controller failed: %s", err.Error())
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	go func(){
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
	endpoint := fmt.Sprintf("%s:%d", conf.Address, conf.Port)
	grpcmux := grpcruntime.NewServeMux()
	opts := []rawgrpc.DialOption{rawgrpc.WithInsecure()}
	if err := meshmanager.RegisterMeshManagerHandlerFromEndpoint(ctx, grpcmux, endpoint, opts); err != nil {
		klog.Errorf("register grpc-gateway failed, %s", err.Error())
		os.Exit(1)
	}
	// http mux
	mux := http.NewServeMux()
	mux.Handle("/", grpcmux)
	go func(){
		klog.Infof("starting manager")
		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", conf.Address, conf.Port+1), mux); err != nil {
			klog.Errorf("running manager failed: %s", err.Error())
			os.Exit(1)
		}
	}()
	//tls
	tlsConf,err := ssl.ClientTslConfVerity(conf.EtcdCaFile, conf.EtcdCertFile, conf.EtcdKeyFile, "")
	if err!=nil {
		klog.Errorf("new client tsl conf failed: %s", err.Error())
		os.Exit(1)
	}
	// New Service
	regOption := func(e *registry.Options){
		e.Addrs = []string{"https://127.0.0.1:2379"}
		e.TLSConfig = tlsConf
	}
	grpcSvr := grpc.NewService(
		service.Name("bcs-mesh-manager.bkbcs.tencent.com"),
		service.Version("1.18.2-alpha"),
		service.Context(ctx),
		service.Address(endpoint),
		service.Registry(etcd.NewRegistry(regOption)),
	)
	grpc.NewService()
	// Initialise service
	grpcSvr.Init()
	// Register Handler, if we need more options control
	// try formation like: handler.BcsDataManager(CustomOption)
	meshHandler := handler.NewMeshHandler(conf, mgr.GetClient())
	err = meshmicro.RegisterMeshManagerHandler(grpcSvr.Server(), meshHandler)
	if err!=nil {
		klog.Errorf("RegisterMeshManagerHandler failed: %s", err.Error())
	}
	// Run service
	klog.Infof("Listen grpc/http server on endpoint(%s)", endpoint)
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