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

package manager

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"reflect"
	osruntime "runtime"
	"strings"
	ossync "sync"
	"time"

	grpccli "github.com/asim/go-micro/plugins/client/grpc/v4"
	"github.com/asim/go-micro/plugins/registry/etcd/v4"
	grpcsvr "github.com/asim/go-micro/plugins/server/grpc/v4"
	etcdsync "github.com/asim/go-micro/plugins/sync/etcd/v4"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/sync"
	"go-micro.dev/v4/util/file"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/jwt"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/handler"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/controller"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/proxy/argocd"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/tunnel"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/utils"
	pb "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/proto"
)

// NewServer create gitops-manager main server
func NewServer(opt *Options) *Server {
	cxt, cancel := context.WithCancel(context.Background())
	return &Server{
		cxt:    cxt,
		cancel: cancel,
		option: opt,
		stops:  make([]utils.StopFunc, 0),
	}
}

// Server for gitops
type Server struct {
	cxt    context.Context
	cancel context.CancelFunc
	stops  []utils.StopFunc
	option *Options
	// etcdSync for leader election,
	// only leader can create tunnel in tunnel mode
	etcdSync         sync.Sync
	waitLeaderResign chan struct{}
	microService     micro.Service
	httpService      *http.Server
	// controller for data sync
	clusterCtl controller.ClusterControl
	projectCtl controller.ProjectControl
	// gitops revese proxy, including auth plugin
	gitops proxy.GitOpsProxy
	// gitops data storage
	storage store.Store
}

// Init all subsystems
func (s *Server) Init() error {
	initializer := []func() error{
		s.initStorage, s.initController, s.initMicroService,
		s.initHTTPService, s.initLeaderElection,
	}
	for _, init := range initializer {
		if err := init(); err != nil {
			return err
		}
	}
	return nil
}

// Run all service, blocking
func (s *Server) Run() error {
	runners := []func(){
		s.startSignalHandler, s.startMicroService,
		s.startHTTPService, s.startLeaderElection,
	}
	for _, run := range runners {
		time.Sleep(time.Millisecond * 500)
		go run()
	}
	<-s.cxt.Done()
	s.stop()

	return nil
}

func (s *Server) stop() {
	wg := &ossync.WaitGroup{}
	wg.Add(len(s.stops))
	for _, stop := range s.stops {
		go func(f func()) {
			f()
			blog.Infof("manager stop func '%s' is finished",
				osruntime.FuncForPC(reflect.ValueOf(f).Pointer()).Name())
			wg.Done()
		}(stop)
	}
	wg.Wait()
}

func (s *Server) initStorage() error {
	opt := &store.Options{
		Service: s.option.GitOps.Service,
		User:    s.option.GitOps.User,
		Pass:    s.option.GitOps.Pass,
		Cache:   false,
	}
	s.storage = store.NewStore(opt)
	if err := s.storage.Init(); err != nil {
		blog.Errorf("manager init gitops storage failure, %s", err.Error())
		return fmt.Errorf("gitops storage failure")
	}
	s.stops = append(s.stops, s.storage.Stop)
	return nil
}

func (s *Server) initMicroService() error {
	svc := micro.NewService(
		micro.Client(grpccli.NewClient(grpccli.AuthTLS(s.option.ClientTLS))),
		micro.Server(grpcsvr.NewServer(grpcsvr.AuthTLS(s.option.ServerTLS))),
		micro.Name(common.ServiceName),
		micro.Metadata(map[string]string{
			common.MetaHTTPKey: fmt.Sprintf("%d", s.option.HTTPPort),
		}),
		micro.Address(fmt.Sprintf("%s:%d", s.option.Address, s.option.Port)),
		micro.Version(version.BcsVersion),
		micro.RegisterTTL(30*time.Second),
		micro.RegisterInterval(25*time.Second),
		micro.Context(s.cxt),
		micro.Registry(etcd.NewRegistry(
			registry.Addrs(strings.Split(s.option.Registry.Endpoints, ",")...),
			registry.TLSConfig(s.option.Registry.TLSConfig),
		)),
	)
	s.microService = svc
	opt := &handler.Options{
		Storage:        s.storage,
		AdminNamespace: s.option.GitOps.AdminNamespace,
		ClusterControl: s.clusterCtl,
		ProjectControl: s.projectCtl,
	}
	if err := pb.RegisterBcsGitopsManagerHandler(
		s.microService.Server(),
		handler.NewGitOpsHandler(opt)); err != nil {
		blog.Errorf("manager register GitOpsManager handler failed, %s", err.Error())
		return fmt.Errorf("GitOps Handler register failure")
	}
	blog.Infof("manager init micro service successfully")
	return nil
}

func (s *Server) startMicroService() {
	if err := s.microService.Run(); err != nil {
		blog.Fatalf("manager start micro service failed, %s", err.Error())
	}
}

func (s *Server) initHTTPService() error {
	router := mux.NewRouter()
	router.UseEncodedPath()
	// init grpc http proxy
	// proxy link: /gitopsmanager/v1/*
	if err := s.initGrpcGateway(router); err != nil {
		return err
	}

	// init gitops http proxy
	// proxy link: /gitopsmanager/proxy/*
	if err := s.initGitOpsProxy(router); err != nil {
		return err
	}

	// init apidoc proxy , path /gitopsmanager/swagger/
	if err := s.initAPIDocs(router); err != nil {
		return err
	}

	// !! important fix: golang strim %2f%2f to / in URL path
	bugWork := &proxy.BUG21955Workaround{Handler: router}

	// ready to create http Server for next starting
	s.httpService = &http.Server{
		Addr:      fmt.Sprintf("%s:%d", s.option.Address, s.option.HTTPPort),
		Handler:   bugWork,
		TLSConfig: s.option.ServerTLS,
	}
	blog.Infof("manager init http service succefully")
	return nil
}

func (s *Server) startHTTPService() {
	if s.httpService == nil {
		blog.Fatalf("lost http server instance")
		return
	}
	s.stops = append(s.stops, s.stopHTTPService)
	err := s.httpService.ListenAndServeTLS("", "")
	if err != nil {
		if http.ErrServerClosed == err {
			blog.Warnf("manager http service gracefully exit.")
			return
		}
		// start http gateway error, maybe port is conflict or something else
		blog.Fatal("manager http service ListenAndServeTLS fatal, %s", err.Error())
	}
}

// stopHTTPService  gracefully stop http server
func (s *Server) stopHTTPService() {
	cxt, cancel := context.WithTimeout(s.cxt, time.Second*2)
	defer cancel()
	if err := s.httpService.Shutdown(cxt); err != nil {
		blog.Errorf("manager gracefully shutdown http service failure: %s", err.Error())
		return
	}
	blog.Infof("manager shutdown http service gracefully")
}

// init grpc http proxy
// proxy link: /gitopsmanager/v1/*
func (s *Server) initGrpcGateway(router *mux.Router) error {
	gatewayCxt, gatewayCancel := context.WithCancel(s.cxt)
	gatewayMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	)
	s.stops = append(s.stops, utils.StopFunc(gatewayCancel))
	opts := []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(s.option.ServerTLS))}
	// register grpc handler to http gateway
	if err := pb.RegisterBcsGitopsManagerGwFromEndpoint(
		gatewayCxt, gatewayMux,
		fmt.Sprintf("%s:%d", s.option.Address, s.option.Port), opts,
	); err != nil {
		blog.Errorf("")
		return fmt.Errorf("manager register grpc http gateway failure")
	}
	// register grpc gateway path to root router
	router.PathPrefix("/gitopsmanager/v1").Handler(gatewayMux)
	blog.Infof("manager init grpc gateway successfully")
	return nil
}

// init gitops http proxy,
// proxy link: /gitopsmanager/proxy/* .
// gitops porxy must be implemented http.Handler, that we can
// change to other gitops solution easilly
func (s *Server) initGitOpsProxy(router *mux.Router) error {
	// first, init auth information
	iamClient, err := iam.NewIamClient(&iam.Options{
		SystemID:    s.option.Auth.SystemID,
		AppCode:     s.option.Auth.AppCode,
		AppSecret:   s.option.Auth.AppSecret,
		External:    s.option.Auth.External,
		GateWayHost: s.option.Auth.Gateway,
		IAMHost:     s.option.Auth.IAMHost,
		BkiIAMHost:  s.option.Auth.BKIAM,
	})
	if err != nil {
		blog.Errorf("manager init iam client failure, %s", err.Error())
		return err
	}
	// second, init Auth info decoder
	jwtClient, err := jwt.NewJWTClient(jwt.JWTOptions{
		VerifyKeyFile: s.option.Auth.VerifyKeyFile,
		SignKeyFile:   s.option.Auth.SignKeyFile,
	})
	if err != nil {
		blog.Errorf("manager init jwt client failure, %s", err.Error())
		return err
	}

	opt := &proxy.GitOpsOptions{
		Service:    s.option.GitOps.Service,
		PathPrefix: common.GitOpsProxyURL,
		JWTDecoder: jwtClient,
		IAMClient:  iamClient,
		Storage:    s.storage,
	}

	s.gitops = argocd.NewGitOpsProxy(opt)
	if err := s.gitops.Init(); err != nil {
		return err
	}
	// Handle "/gitopsmanager/proxy", s.gitops
	router.PathPrefix(common.GitOpsProxyURL).Handler(s.gitops)
	blog.Infof("manager init gitops proxy router successfully")
	return nil
}

func (s *Server) initAPIDocs(router *mux.Router) error {
	ok, err := file.Exists("./swagger/index.html")
	if err != nil {
		blog.Errorf("check api docs in local err: %s, skip api docs serving", err.Error())
		return nil
	}
	if !ok {
		blog.Errorf("lost api docs in local, skip api docs serving")
		return nil
	}
	// init api docs, no auth in this path
	router.HandleFunc("/gitopsmanager/swagger/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(
			w, r,
			path.Join("/data/bcs/bcs-gitops-manager/swagger", strings.TrimPrefix(r.URL.Path,
				"/gitopsmanager/swagger/")),
		)
	})
	return nil
}

func (s *Server) initController() error {
	ctx, cancel := context.WithCancel(s.cxt)
	s.stops = append(s.stops, utils.StopFunc(cancel))
	// cluster controller
	opt := &controller.Options{
		Context:     ctx,
		Mode:        s.option.Mode,
		ClientTLS:   s.option.ClientTLS,
		Registry:    s.option.Registry.Endpoints,
		RegistryTLS: s.option.Registry.TLSConfig,
		APIGateway:  s.option.APIGateway,
		APIToken:    s.option.APIGatewayToken,
		Interval:    s.option.ClusterSyncInterval,
		Storage:     s.storage,
	}
	s.clusterCtl = controller.NewClusterController(opt)
	if err := s.clusterCtl.Init(); err != nil {
		blog.Errorf("manager init cluster controller failure, %s", err.Error())
		return err // nolint
	}
	s.stops = append(s.stops, utils.StopFunc(s.clusterCtl.Stop))
	blog.Infof("manager init cluster controller successfully")

	s.projectCtl = controller.NewProjectController(opt)
	if err := s.projectCtl.Init(); err != nil {
		blog.Errorf("manager init project controller failure, %s", err.Error())
		return err
	}
	s.stops = append(s.stops, utils.StopFunc(s.projectCtl.Stop))
	return nil
}

// initLeaderElection init etcd leader election impelementation
func (s *Server) initLeaderElection() error {
	if s.option.Mode != common.ModeTunnel {
		blog.Infof("manager is not in tunnel mode, leader election is unnecessary")
		return nil
	}

	hosts := strings.Split(s.option.Registry.Endpoints, ",")
	// for Context control
	s.etcdSync = etcdsync.NewSync(
		sync.WithTLS(s.option.Registry.TLSConfig),
		sync.Nodes(hosts...),
	)
	blog.Infof("manager construct leader sync successfully")
	return s.etcdSync.Init()
}

// initLeaderElection init etcd leader election impelementation
func (s *Server) startLeaderElection() {
	if s.option.Mode != common.ModeTunnel {
		blog.Infof("manager is not in tunnel mode, start leader election is unnecessary")
		return
	}
	s.waitLeaderResign = make(chan struct{})
	s.stops = append(s.stops, func() {
		<-s.waitLeaderResign
	})
	blog.Infof("manager runs in tunnel mode, creating websocket tunnel")
	cxt, cancel := context.WithCancel(s.cxt)
	// if lost leader role, stop tunnel client by CancelFunc
	defer cancel()
	leaderID := common.ServiceName
	leader, err := s.etcdSync.Leader(leaderID)
	if err != nil {
		blog.Errorf("manager %s campaign leader role met error, %s", leaderID, err.Error())
		// maybe network error, server can start elect
		// again after backoff strategy
		time.Sleep(time.Second * gracefulPeriod)
		go s.startLeaderElection()
		return
	}
	blog.Infof("manager %s become leader, starting controller...", leaderID)
	lost := leader.Status()
	// server campaign leader successfully, construct tunnel for proxy.
	// tunnel must start successfully or there is bug need to fix
	if err := s.startTunnelClient(cxt); err != nil {
		blog.Fatalf("server start tunnel fatal, %s", err.Error())
		return
	}
	// start cluster controller for interval synchronization
	// only leader can start cluster controller
	go s.clusterCtl.SingleStart(cxt)

	// start tunnel successfully, check leader status continuously.
	// when server lost leader role, stop tunnel and try to elect again
	select {
	case <-lost:
		time.Sleep(time.Second * gracefulPeriod)
		go s.startLeaderElection()
		// nothing to stop, recycle resource by defer CancelFunc()
		return
	case <-cxt.Done():
		blog.Infof("manager leaderelection received context done, and will resign leader.")
		// server exit
		if err := leader.Resign(); err != nil {
			blog.Errorf("manager %s resign leader failure, %s, other servers wait until timeout",
				leaderID, err.Error())
		} else {
			blog.Infof("manager %s resign leader successfully, prepare exit", leaderID)
		}
		close(s.waitLeaderResign)
	}
}

// startTunnelClient with context, close tunnel by CancelFunc
func (s *Server) startTunnelClient(cxt context.Context) error {
	opt := &tunnel.ClientOptions{
		Context:       cxt,
		ProxyAddress:  fmt.Sprintf("wss://%s", s.option.APIGateway),
		ProxyToken:    s.option.APIGatewayToken,
		TLSConfig:     s.option.ClientTLS,
		LocalEndpoint: fmt.Sprintf("https://%s:%d", s.option.Address, s.option.HTTPPort),
		ClusterID:     common.ServiceName,
		ClusterToken:  s.option.APIConnectToken,
		ConnectURL:    s.option.APIConnectURL,
	}
	client := tunnel.NewClient(opt)
	if err := client.Init(); err != nil {
		return err
	}
	client.Start()
	blog.Infof("manager start tunnel client successufully")
	return nil
}

func (s *Server) startSignalHandler() {
	blog.Infof("manager start system signal successufully")
	utils.StartSignalHandler(s.cancel, 3)
}
