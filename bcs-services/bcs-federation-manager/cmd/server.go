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

// Package cmd xxx
package cmd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	machinelog "github.com/RichardKnop/machinery/v2/log"
	grpccli "github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	grpcsvr "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/task"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	trace "github.com/Tencent/bk-bcs/bcs-common/pkg/otel/trace/micro"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/helm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/requester"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/thirdparty"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/clients/user"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store"
	fmmongo "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/store/mongo"
	fedtask "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/task"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/internal/utils"
	federationmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/proto/bcs-federation-manager"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
)

// Server federation manager server
type Server struct {
	microService  micro.Service
	microRegistry registry.Registry

	tlsConfig       *tls.Config
	clientTLSConfig *tls.Config

	httpServer  *http.Server
	opt         *FederationManagerOptions
	store       store.FederationMangerModel
	clusterCli  cluster.Client
	helmCli     helm.Client
	userCli     user.Client
	projectCli  project.Client
	iamCli      iam.PermClient
	thirdCli    thirdparty.Client
	taskmanager *task.TaskManager

	ctx           context.Context
	ctxCancelFunc context.CancelFunc
	stopChan      chan struct{}
}

// NewServer create federation manager instance
func NewServer(opt *FederationManagerOptions) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		opt:           opt,
		ctx:           ctx,
		ctxCancelFunc: cancel,
		stopChan:      make(chan struct{}),
	}
}

// Init init modules of server
func (s *Server) Init() error {
	// initializers by sequence
	initializer := []func() error{
		s.initTLSConfig,
		s.initRegistry,
		s.initClusterCli,
		s.initHelmCli,
		s.initUserCli,
		s.initProjectCli,
		s.initThirdpartyCli,
		s.initModel,
		s.initIAMClient,
		s.initJWTClient,
		s.initSkipClients,
		s.initSkipHandler,
		s.initTaskManager,
		s.initMicro,
		s.initHTTPService,
		s.initSyncNamespaceQuotaTicker,
	}

	// init
	for _, init := range initializer {
		if err := init(); err != nil {
			return err
		}
	}

	return nil
}

// Run run the server
func (s *Server) Run() error {
	// run task manager
	s.taskmanager.Run()

	eg, _ := errgroup.WithContext(s.ctx)

	eg.Go(func() error {
		return s.microService.Run()
	})

	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

// initTLSConfig init server and client tls config
func (s *Server) initTLSConfig() error {
	if len(s.opt.ServerCert) != 0 && len(s.opt.ServerKey) != 0 && len(s.opt.ServerCa) != 0 {
		tlsConfig, err := ssl.ServerTslConfVerityClient(s.opt.ServerCa, s.opt.ServerCert,
			s.opt.ServerKey, static.ServerCertPwd)
		if err != nil {
			blog.Errorf("load federation manager server tls config failed, err %s", err.Error())
			return err
		}
		s.tlsConfig = tlsConfig
		blog.Infof("load federation manager server tls config successfully")
	}

	if len(s.opt.ClientCert) != 0 && len(s.opt.ClientKey) != 0 && len(s.opt.ClientCa) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(s.opt.ClientCa, s.opt.ClientCert,
			s.opt.ClientKey, static.ClientCertPwd)
		if err != nil {
			blog.Errorf("load federation manager client tls config failed, err %s", err.Error())
			return err
		}
		s.clientTLSConfig = tlsConfig
		blog.Infof("load federation manager client tls config successfully")
	}

	// init tls config success
	blog.Infof("init tls config successfully")
	return nil
}

// initModel init mongo client
func (s *Server) initModel() error {
	mongoOpt, err := s.getMongoOption()
	if err != nil {
		return err
	}

	mongoDB, err := mongo.NewDB(mongoOpt)
	if err != nil {
		blog.Errorf("init mongo db failed, err %s", err.Error())
		return err
	}
	if err = mongoDB.Ping(); err != nil {
		blog.Errorf("ping mongo db failed, err %s", err.Error())
		return err
	}

	s.store = fmmongo.NewServer(mongoDB)
	store.SetStoreModel(s.store)

	// init mongo client success
	blog.Infof("init store successfully")
	return nil
}

// initRegistry init micro service registry
func (s *Server) initRegistry() error {
	// parse etcd endpoints
	address := strings.ReplaceAll(s.opt.Etcd.EtcdEndpoints, ";", ",")
	address = strings.ReplaceAll(address, " ", ",")
	etcdEndpoints := strings.Split(address, ",")
	etcdSecure := false

	var etcdTLS *tls.Config
	var err error
	if len(s.opt.Etcd.EtcdCa) != 0 && len(s.opt.Etcd.EtcdCert) != 0 && len(s.opt.Etcd.EtcdKey) != 0 {
		etcdSecure = true
		etcdTLS, err = ssl.ClientTslConfVerity(s.opt.Etcd.EtcdCa, s.opt.Etcd.EtcdCert, s.opt.Etcd.EtcdKey, "")
		if err != nil {
			return err
		}
		s.opt.Etcd.tlsConfig = etcdTLS
	}
	s.microRegistry = etcd.NewRegistry(
		registry.Addrs(etcdEndpoints...),
		registry.Secure(etcdSecure),
		registry.TLSConfig(etcdTLS),
	)
	if err := s.microRegistry.Init(); err != nil {
		return err
	}

	// init registry success
	blog.Infof("init registry successfully")
	return nil
}

// init micro service
func (s *Server) initMicro() error {
	authWrapper := middleware.NewGoMicroAuth(auth.GetJWTClient()).
		EnableSkipHandler(auth.SkipHandler).
		EnableSkipClient(auth.SkipClient).
		SetCheckUserPerm(auth.CheckUserPerm)

	//init micro service
	s.microService = micro.NewService(
		micro.Server(grpcsvr.NewServer(
			grpcsvr.AuthTLS(s.tlsConfig),
		)),
		micro.Client(grpccli.NewClient(
			grpccli.AuthTLS(s.clientTLSConfig),
		)),
		micro.Name(common.ServiceDomain),
		micro.Context(s.ctx),
		micro.Metadata(map[string]string{common.MicroMetaKeyHTTPPort: strconv.Itoa(int(s.opt.HTTPPort))}),
		micro.Address(net.JoinHostPort(s.opt.Address, strconv.Itoa(int(s.opt.Port)))),
		micro.Version(version.BcsVersion),
		micro.RegisterTTL(30*time.Second),
		micro.RegisterInterval(25*time.Second),
		micro.Registry(s.microRegistry),
		micro.Flags(&cli.StringFlag{
			Name:        "f",
			Usage:       "set config file path",
			DefaultText: "./bcs-federation-manager.json",
		}),
		micro.WrapHandler(
			utils.ResponseWrapper,
			authWrapper.AuthenticationFunc,
			authWrapper.AuthorizationFunc,
			trace.NewTracingWrapper(),
		),
	)
	s.microService.Init()

	// decode token
	realAuthToken, err := encrypt.DesDecryptFromBase([]byte(s.opt.Gateway.Token))
	if err != nil {
		return fmt.Errorf("init clusterCli failed, encrypt token error: %s", err.Error())
	}

	// register federationmanager handler
	if err := federationmanager.RegisterFederationManagerHandler(
		s.microService.Server(),
		handler.NewFederationManager(&handler.FederationManagerOptions{
			BcsGateway: &handler.GatewayConfig{
				Endpoint: s.opt.Gateway.Endpoint,
				Token:    string(realAuthToken),
			},
			Store:             s.store,
			ClusterManagerCli: s.clusterCli,
			HelmManagerCli:    s.helmCli,
			UserManagerCli:    s.userCli,
			ProjectManagerCli: s.projectCli,
			ThirdManagerCli:   s.thirdCli,
			TaskManager:       s.taskmanager,
		}),
	); err != nil {
		blog.Errorf("failed to register federation manager handler to micro, error: %s", err.Error())
		return err
	}

	blog.Infof("success to register federation manager handler to micro")
	return nil
}

// initHTTPService init http service
func (s *Server) initHTTPService() error {
	// init http gateway
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	)
	grpcDialOpts := []grpc.DialOption{}
	if s.tlsConfig != nil && s.clientTLSConfig != nil {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(credentials.NewTLS(s.clientTLSConfig)))
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
	}
	// RegisterFederationManagerGwFromEndpoint
	err := federationmanager.RegisterFederationManagerGwFromEndpoint(
		context.TODO(),
		gwmux,
		net.JoinHostPort(s.opt.Address, strconv.Itoa(int(s.opt.Port))),
		grpcDialOpts)
	if err != nil {
		blog.Errorf("register http gateway failed, err %s", err.Error())
		return fmt.Errorf("register http gateway failed, err %s", err.Error())
	}
	router := mux.NewRouter()
	router.Handle("/{uri:.*}", gwmux)
	blog.Info("register grpc gateway handler to path /")

	// init http server
	smux := http.NewServeMux()
	smux.Handle("/", router)

	httpAddress := net.JoinHostPort(s.opt.Address, strconv.Itoa(int(s.opt.HTTPPort)))

	s.httpServer = &http.Server{
		Addr:    httpAddress,
		Handler: smux,
	}

	// start http gateway server
	go func() {
		var err error
		blog.Infof("start http gateway server on address %s", httpAddress)
		if s.tlsConfig != nil {
			s.httpServer.TLSConfig = s.tlsConfig
			err = s.httpServer.ListenAndServeTLS("", "")
		} else {
			err = s.httpServer.ListenAndServe()
		}
		if err != nil {
			blog.Errorf("start http gateway server failed, err %s", err.Error())
			s.ctxCancelFunc()
		}
	}()

	return nil
}

// initClusterCli init k8s client by bcs gateway
func (s *Server) initClusterCli() error {
	realAuthToken, err := encrypt.DesDecryptFromBase([]byte(s.opt.Gateway.Token))
	if err != nil {
		return fmt.Errorf("init clusterCli failed, encrypt token error: %s", err.Error())
	}
	opts := &cluster.ClientOptions{
		ClientTLS: s.clientTLSConfig,
		BaseOptions: requester.BaseOptions{
			Endpoint: s.opt.Gateway.Endpoint,
			Token:    string(realAuthToken),
			Sender:   requester.NewRequester(),
		},
	}

	if err := cluster.SetClusterClient(opts); err != nil {
		return fmt.Errorf("failed to set cluster client: %v", err)
	}
	s.clusterCli = cluster.GetClusterClient()

	blog.Infof("init cluster client successfully")
	return nil
}

// initHelmCli init helm manager client
func (s *Server) initHelmCli() error {
	realAuthToken, err := encrypt.DesDecryptFromBase([]byte(s.opt.Gateway.Token))
	if err != nil {
		return fmt.Errorf("init helmCli failed, encrypt token error: %s", err.Error())
	}
	opts := &helm.ClientOptions{
		ClientTLS: s.clientTLSConfig,
		BaseOptions: requester.BaseOptions{
			Endpoint: s.opt.Gateway.Endpoint,
			Token:    string(realAuthToken),
			Sender:   requester.NewRequester(),
		},
		Charts: s.opt.Deploy,
	}

	if err := helm.SetHelmClient(opts); err != nil {
		return fmt.Errorf("failed to set helm client: %v", err)
	}
	s.helmCli = helm.GetHelmClient()

	blog.Infof("init helm manager client successfully")
	return nil
}

// initUserCli init user client
func (s *Server) initUserCli() error {
	realAuthToken, err := encrypt.DesDecryptFromBase([]byte(s.opt.Gateway.Token))
	if err != nil {
		return fmt.Errorf("init userCli failed, encrypt token error: %s", err.Error())
	}
	opts := &user.ClientOptions{
		ClientTLS: s.clientTLSConfig,
		BaseOptions: requester.BaseOptions{
			Endpoint: s.opt.Gateway.Endpoint,
			Token:    string(realAuthToken),
			Sender:   requester.NewRequester(),
		},
	}

	user.SetUserClient(opts)
	s.userCli = user.GetUserClient()

	blog.Infof("init user client successfully")
	return nil
}

// initProjectCli init project client
func (s *Server) initProjectCli() error {
	realAuthToken, err := encrypt.DesDecryptFromBase([]byte(s.opt.Gateway.Token))
	if err != nil {
		return fmt.Errorf("init projectCli failed, encrypt token error: %s", err.Error())
	}
	opts := &project.ClientOptions{
		ClientTLS: s.clientTLSConfig,
		BaseOptions: requester.BaseOptions{
			Endpoint: s.opt.Gateway.Endpoint,
			Token:    string(realAuthToken),
			Sender:   requester.NewRequester(),
		},
	}

	if err := project.SetProjectClient(opts); err != nil {
		return fmt.Errorf("failed to set project client: %v", err)
	}
	s.projectCli = project.GetProjectClient()

	blog.Infof("init project client successfully")
	return nil
}

// initIAMClient init iam client
func (s *Server) initIAMClient() error {
	var err error
	s.iamCli, err = iam.NewIamClient(&iam.Options{
		SystemID:    s.opt.IAM.SystemID,
		AppCode:     s.opt.IAM.AppCode,
		AppSecret:   s.opt.IAM.AppSecret,
		External:    s.opt.IAM.External,
		GateWayHost: s.opt.IAM.GatewayServer,
		IAMHost:     s.opt.IAM.IAMServer,
		BkiIAMHost:  s.opt.IAM.BkiIAMServer,
		Metric:      s.opt.IAM.Metric,
		Debug:       s.opt.IAM.Debug,
	})
	if err != nil {
		return fmt.Errorf("init iamCli failed, error: %s", err.Error())
	}
	// init perm client
	auth.InitPermClient(s.iamCli)

	blog.Infof("init iam client successfully")
	return nil
}

// initJWTClient init jwt client
func (s *Server) initJWTClient() error {
	err := auth.InitJWTClient(&auth.JWTOptions{
		PublicKeyFile:  s.opt.Auth.PublicKeyFile,
		PrivateKeyFile: s.opt.Auth.PrivateKeyFile,
	})
	if err != nil {
		return err
	}
	blog.Infof("init jwt client successfully")
	return nil
}

func (s *Server) initSkipClients() error {
	auth.ClientPermissions = make(map[string][]string, 0)
	if len(s.opt.Auth.ClientPermissions) == 0 {
		return nil
	}
	err := json.Unmarshal([]byte(s.opt.Auth.ClientPermissions), &auth.ClientPermissions)
	if err != nil {
		return fmt.Errorf("parse skip clients error: %s", err.Error())
	}
	blog.Infof("init skip clients successfully")
	return nil
}

// initSkipHandler
func (s *Server) initSkipHandler() error {
	if len(s.opt.Auth.NoAuthMethod) == 0 {
		return nil
	}
	methods := strings.Split(s.opt.Auth.NoAuthMethod, ",")
	auth.NoAuthMethod = append(auth.NoAuthMethod, methods...)

	auth.EnableAuth = s.opt.Auth.Enable

	blog.Infof("init skip handler successfully")
	return nil
}

// initTaskManager init task server
func (s *Server) initTaskManager() error {
	// set logger for go-machinery which is used by bcs task manager
	machinelog.Set(fedtask.NewLogger())

	mongoOpt, err := s.getMongoOption()
	if err != nil {
		return err
	}

	config := &task.ManagerConfig{
		ModuleName: common.ModuleName,
		WorkerNum:  100,
		Broker: &task.BrokerConfig{
			QueueAddress: s.opt.Broker.QueueAddress,
			Exchange:     s.opt.Broker.Exchange,
		},
		Backend:     mongoOpt,
		StepWorkers: fedtask.RegisterSteps(),
		CallBacks:   fedtask.RegisterCallbacks(),
	}

	tm := task.NewTaskManager()
	if err := tm.Init(config); err != nil {
		return err
	}

	fedtask.SetTaskManagerClient(tm)
	s.taskmanager = fedtask.GetTaskManagerClient()

	// init task manager successfully
	blog.Infof("init task manager successfully")
	return nil
}

func (s *Server) getMongoOption() (*mongo.Options, error) {
	if len(s.opt.Mongo.MongoEndpoints) == 0 {
		return nil, fmt.Errorf("mongo address cannot be empty")
	}
	if len(s.opt.Mongo.MongoDatabaseName) == 0 {
		return nil, fmt.Errorf("mongo database cannot be empty")
	}

	// get mongo password
	password := s.opt.Mongo.MongoPassword
	if password != "" {
		realPwd, err := encrypt.DesDecryptFromBase([]byte(password))
		if err != nil {
			blog.Errorf("decrypt password failed, err %s", err.Error())
			return nil, err
		}
		password = string(realPwd)
	}
	mongoOptions := &mongo.Options{
		Hosts:                 strings.Split(s.opt.Mongo.MongoEndpoints, ","),
		ConnectTimeoutSeconds: s.opt.Mongo.MongoConnectTimeout,
		Database:              s.opt.Mongo.MongoDatabaseName,
		Username:              s.opt.Mongo.MongoUsername,
		Password:              password,
		MaxPoolSize:           0,
		MinPoolSize:           0,
	}

	return mongoOptions, nil
}

// initThirdpartyCli init thirdparty service manager client
func (s *Server) initThirdpartyCli() error {
	realAuthToken, err := encrypt.DesDecryptFromBase([]byte(s.opt.Gateway.Token))
	if err != nil {
		return fmt.Errorf("init clusterCli failed, encrypt token error: %s", err.Error())
	}
	opts := &thirdparty.ClientOptions{
		ClientTLS:     s.clientTLSConfig,
		EtcdEndpoints: strings.Split(s.opt.Etcd.EtcdEndpoints, ","),
		EtcdTLS:       s.opt.Etcd.tlsConfig,
		BaseOptions: requester.BaseOptions{
			Endpoint: s.opt.Gateway.Endpoint,
			Token:    string(realAuthToken),
			Sender:   requester.NewRequester(),
		},
	}

	if err := thirdparty.InitThirdpartyClient(opts); err != nil {
		return fmt.Errorf("failed to initialize thirdparty client: %v", err)
	}
	s.thirdCli = thirdparty.GetThirdpartyClient()

	// init thirdparty client successfully
	blog.Infof("init thirdparty manager client successfully")
	return nil
}

// initSyncNamespaceQuotaTicker init start namespace quota ticker
func (s *Server) initSyncNamespaceQuotaTicker() error {

	cli := handler.NewFedNamespaceControllerManager()
	err := cli.StartLoop(s.ctx, s.store, s.taskmanager, s.clusterCli)
	if err != nil {
		blog.Errorf("FedNamespaceControllerManager start loop failed, err %s", err.Error())
		return err
	}

	// init start namespace quota ticker successfully
	blog.Infof("init sync namespace quota ticker successfully")
	return nil
}
