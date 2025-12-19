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

// Package cmd provides the entry point for the bcs-push-manager service.
package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/discovery"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	grpccli "github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	grpcsvr "github.com/go-micro/plugins/v4/server/grpc"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	cli "github.com/urfave/cli/v2"
	micro "go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/action"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/mq/rabbitmq"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/options"
	mongostore "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/mongo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/thirdparty"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/proto"
)

// Server encapsulates the service's dependencies and lifecycle management.
type Server struct {
	microService        micro.Service
	microRegistry       registry.Registry
	thirdpartyDiscovery *discovery.ModuleDiscovery
	opt                 *options.ServiceOptions
	httpServer          *http.Server
	mqClient            *rabbitmq.RabbitMQ
	mongoServer         *mongostore.Server
	tlsConfig           *tls.Config
	clientTLSConfig     *tls.Config

	ctx           context.Context
	ctxCancelFunc context.CancelFunc
	stopChan      chan struct{}
}

// NewServer creates a new Server instance, initializing context and stopChan.
func NewServer(opt *options.ServiceOptions) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		opt:           opt,
		ctx:           ctx,
		ctxCancelFunc: cancel,
		stopChan:      make(chan struct{}),
	}
}

// Init initializes all dependencies, calling initializers in sequence.
func (s *Server) Init() error {
	initializer := []func() error{
		s.initTLSConfig,
		s.initRegistry,
		s.initStore,
		s.initMQ,
		s.initMicro,
		s.initThirdpartyDiscovery,
		s.initHTTPGateway,
	}
	for _, init := range initializer {
		if err := init(); err != nil {
			return err
		}
	}
	return nil
}

// Run microservice// Run starts all services and manages their lifecycle.
func (s *Server) Run() error {
	eg, ctx := errgroup.WithContext(s.ctx)

	eg.Go(func() error {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		select {
		case sig := <-sigChan:
			blog.Infof("received signal %s, shutting down...", sig)
			s.ctxCancelFunc()
		case <-ctx.Done():
			blog.Infof("context canceled, exiting signal goroutine.")
		}
		return nil
	})

	eg.Go(func() error {
		blog.Infof("starting gRPC service on %s", s.microService.Server().Options().Address)
		return s.microService.Run()
	})

	eg.Go(func() error {
		blog.Infof("starting HTTP gateway on %s", s.httpServer.Addr)
		go func() {
			<-ctx.Done()
			blog.Infof("shutting down HTTP gateway...")
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
				blog.Errorf("HTTP server graceful shutdown failed: %v", err)
			}
		}()
		if s.httpServer.TLSConfig != nil {
			if err := s.httpServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				return err
			}
		} else {
			if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return err
			}
		}
		return nil
	})

	eg.Go(func() error {
		blog.Infof("starting RabbitMQ consumer...")
		err := s.startRabbitMQConsumer(ctx)
		blog.Infof("RabbitMQ consumer stopped.")
		return err
	})

	if err := eg.Wait(); err != nil {
		blog.Errorf("Service exiting with error: %v", err)
		s.ctxCancelFunc()
		return err
	}

	blog.Infof("all services are running. Waiting for exit signal (Ctrl+C or kill). Note: Ctrl+D is not a signal and will not stop the service.")
	if err := eg.Wait(); err != nil && err != context.Canceled {
		blog.Errorf("service group encountered an error: %v", err)
		return err
	}

	blog.Infof("all services stopped gracefully.")

	return nil
}

// initTLSConfig initializes the service's TLS configuration.
func (s *Server) initTLSConfig() error {
	if len(s.opt.ServerCert) != 0 && len(s.opt.ServerKey) != 0 && len(s.opt.ServerCa) != 0 {
		tlsConfig, err := ssl.ServerTslConfVerityClient(s.opt.ServerCa, s.opt.ServerCert,
			s.opt.ServerKey, static.ServerCertPwd)
		if err != nil {
			blog.Errorf("load server tls config failed, err %s", err.Error())
			return err
		}
		s.tlsConfig = tlsConfig
		blog.Infof("load server tls config successfully")
	}

	if len(s.opt.ClientCert) != 0 && len(s.opt.ClientKey) != 0 && len(s.opt.ClientCa) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(s.opt.ClientCa, s.opt.ClientCert,
			s.opt.ClientKey, static.ClientCertPwd)
		if err != nil {
			blog.Errorf("load client tls config failed, err %s", err.Error())
			return err
		}
		s.clientTLSConfig = tlsConfig
		blog.Infof("load client tls config successfully")
	}
	return nil
}

// initRegistry initializes the service registry.
func (s *Server) initRegistry() error {
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
		s.opt.Etcd.TlsConfig = etcdTLS
	}

	s.microRegistry = etcd.NewRegistry(
		registry.Addrs(etcdEndpoints...),
		registry.Secure(etcdSecure),
		registry.TLSConfig(etcdTLS),
	)

	if err := s.microRegistry.Init(); err != nil {
		return err
	}
	blog.Infof("init registry successfully")
	return nil
}

// initStore initializes the MongoDB store.
func (s *Server) initStore() error {
	if len(s.opt.Mongo.MongoEndpoints) == 0 {
		return fmt.Errorf("mongo address cannot be empty")
	}
	if len(s.opt.Mongo.MongoDatabaseName) == 0 {
		return fmt.Errorf("mongo database cannot be empty")
	}
	password := s.opt.Mongo.MongoPassword
	if password != "" {
		realPwd, _ := encrypt.DesDecryptFromBase([]byte(password))
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
	instance, err := mongo.NewDB(mongoOptions)
	if err != nil {
		return fmt.Errorf("storage create mongo instance failed, %s", err.Error())
	}
	if pingErr := instance.Ping(); pingErr != nil {
		return fmt.Errorf("storage connection test failed, %s", pingErr.Error())
	}
	s.mongoServer = mongostore.NewServer(instance)
	return nil
}

// initMQ initializes the RabbitMQ client.
func (s *Server) initMQ() error {
	mqClient := rabbitmq.NewRabbitMQ(s.opt.RabbitMQ)
	if err := mqClient.Connect(); err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	s.mqClient = mqClient
	return nil
}

// initThirdpartyDiscovery initializes the thirdparty service discovery.
func (s *Server) initThirdpartyDiscovery() error {
	if !discovery.UseServiceDiscovery() {
		s.thirdpartyDiscovery = discovery.NewModuleDiscovery(constant.ModuleThirdpartyServiceManager, s.microRegistry)
		blog.Infof("init discovery for thirdparty service successfully")

		if err := s.thirdpartyDiscovery.Start(); err != nil {
			return fmt.Errorf("failed to start thirdparty discovery: %v", err)
		}

		maxRetries := 10
		for i := 0; i < maxRetries; i++ {
			services := s.thirdpartyDiscovery.GetService()
			if len(services) > 0 && len(services[0].Nodes) > 0 {
				blog.Infof("thirdparty service endpoints discovered successfully")
				break
			}
			if i == maxRetries-1 {
				return fmt.Errorf("thirdparty service endpoints not available after %d retries", maxRetries)
			}
			blog.Infof("waiting for thirdparty service endpoints, retry %d/%d", i+1, maxRetries)
			time.Sleep(500 * time.Millisecond)
		}
	}

	thirdpartyOpts := &thirdparty.ClientOptions{
		ClientTLS: s.clientTLSConfig,
		Discovery: s.thirdpartyDiscovery,
	}
	if err := thirdparty.InitThirdpartyClient(thirdpartyOpts); err != nil {
		return fmt.Errorf("failed to initialize thirdparty client: %v", err)
	}

	return nil
}

// initMicro initializes the gRPC service.
func (s *Server) initMicro() error {
	pushEventAction := action.NewPushEventAction(s.mongoServer.GetPushEventModel())
	pushWhitelistAction := action.NewPushWhitelistAction(s.mongoServer.GetPushWhitelistModel())
	pushTemplateAction := action.NewPushTemplateAction(s.mongoServer.GetPushTemplateModel())

	svcHandler := handler.NewPushManagerService(
		pushEventAction,
		pushWhitelistAction,
		pushTemplateAction,
		s.mqClient,
	)

	s.microService = micro.NewService(
		micro.Server(grpcsvr.NewServer(
			grpcsvr.AuthTLS(s.tlsConfig),
		)),
		micro.Client(grpccli.NewClient(
			grpccli.AuthTLS(s.clientTLSConfig),
		)),
		micro.Name(constant.ModulePushManager),
		micro.Context(s.ctx),
		micro.Metadata(map[string]string{constant.MicroMetaKeyHTTPPort: strconv.Itoa(int(s.opt.HTTPPort))}),
		micro.Address(net.JoinHostPort(s.opt.ServerConfig.Address, strconv.Itoa(int(s.opt.Port)))),
		micro.Version(version.BcsVersion),
		micro.RegisterTTL(30*time.Second),
		micro.RegisterInterval(25*time.Second),
		micro.Registry(s.microRegistry),
		micro.BeforeStop(func() error {
			if s.thirdpartyDiscovery != nil {
				s.thirdpartyDiscovery.Stop()
			}
			return nil
		}),
		micro.AfterStop(func() error {
			if err := thirdparty.CloseThirdpartyClient(); err != nil {
				blog.Errorf("failed to close thirdparty client: %w", err)
			}
			return nil
		}),
		micro.Flags(&cli.StringFlag{
			Name:        "f",
			Usage:       "set config file path",
			DefaultText: "./bcs-push-manager.json",
		}),
	)
	s.microService.Init()

	if err := pb.RegisterPushManagerHandler(
		s.microService.Server(),
		svcHandler,
	); err != nil {
		blog.Errorf("failed to register handler to micro, error: %s", err.Error())
		return err
	}
	blog.Infof("success to register handler to micro")
	return nil
}

// initHTTPGateway initializes the HTTP Gateway.
func (s *Server) initHTTPGateway() error {
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{}),
	)
	grpcDialOpts := []grpc.DialOption{}
	if s.tlsConfig != nil && s.clientTLSConfig != nil {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(credentials.NewTLS(s.clientTLSConfig)))
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure()) // nolint
	}
	if err := pb.RegisterPushManagerGwFromEndpoint(
		context.TODO(),
		gwmux,
		net.JoinHostPort(s.opt.ServerConfig.Address, strconv.Itoa(int(s.opt.ServerConfig.Port))),
		grpcDialOpts,
	); err != nil {
		blog.Errorf("register http gateway failed, err %s", err.Error())
		return fmt.Errorf("register http gateway failed, err %s", err.Error())
	}

	router := mux.NewRouter()
	router.Handle("/{uri:.*}", gwmux)
	blog.Info("register grpc gateway handler to path /")

	// init http server
	smux := http.NewServeMux()
	smux.Handle("/", router)

	httpAddress := net.JoinHostPort(s.opt.ServerConfig.Address, strconv.Itoa(int(s.opt.ServerConfig.HTTPPort)))

	s.httpServer = &http.Server{
		Addr:    httpAddress,
		Handler: smux,
	}
	if s.tlsConfig != nil {
		s.httpServer.TLSConfig = s.tlsConfig
	}
	return nil
}

// startConsumer initializes and starts the RabbitMQ consumer, supporting ctx-based shutdown.
func (s *Server) startRabbitMQConsumer(ctx context.Context) error {
	blog.Infof("starting RabbitMQ consumer...")
	defer blog.Infof("RabbitMQ consumer stopped.")

	// Create notification action with dependencies
	notificationAction := &action.NotificationAction{
		ThirdpartyClient: thirdparty.GetThirdpartyClient(),
		WhitelistStore:   s.mongoServer.GetPushWhitelistModel(),
		EventStore:       s.mongoServer.GetPushEventModel(),
		MaxRetry:         3,
		RetryInterval:    5 * time.Second,
	}

	// Open RabbitMQ channel
	channel, err := s.mqClient.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %v", err)
	}
	defer channel.Close()

	// Ensure exchange exists
	if err := s.mqClient.EnsureExchange(channel); err != nil {
		return fmt.Errorf("failed to ensure exchange: %v", err)
	}

	// Generate consumer identifier
	hostname, err := os.Hostname()
	if err != nil {
		blog.Errorf("failed to get hostname: %v, using default hostname", err)
		hostname = fmt.Sprintf("unknown-host-%x", time.Now().UnixNano())
	}
	queueName := constant.NotificationActionQueueName
	consumerName := fmt.Sprintf("consumer-%s-%d", hostname, os.Getpid())
	exchangeName := fmt.Sprintf("%s.topic", s.opt.RabbitMQ.SourceExchange)

	// Declare and bind queue
	if err := s.mqClient.DeclareQueue(channel, queueName, nil); err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	if err := s.mqClient.BindQueue(channel, queueName, exchangeName, nil); err != nil {
		return fmt.Errorf("failed to bind queue: %v", err)
	}

	// Start consumer with retry logic
	errChan := make(chan error, 1)
	done := make(chan bool, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				blog.Infof("context canceled, stopping consumer...")
				return
			default:
				if err := s.mqClient.StartConsumer(channel, consumerName, queueName, notificationAction, done); err != nil {
					blog.Errorf("consumer error: %v, retrying in 5 seconds...", err)
					errChan <- err
					time.Sleep(5 * time.Second)
				} else {
					blog.Infof("consumer started successfully")
					return
				}
			}
		}
	}()

	// Wait for context cancellation or consumer error
	select {
	case <-ctx.Done():
		blog.Info("context canceled, stopping consumer...")
		return nil
	case err := <-errChan:
		return fmt.Errorf("RabbitMQ consumer failed: %v", err)
	}
}
