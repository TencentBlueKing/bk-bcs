/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/pprof"
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
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	restclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/cmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/worker"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"google.golang.org/grpc"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/handler"
	microsvc "github.com/micro/go-micro/v2/service"
	microgrpcsvc "github.com/micro/go-micro/v2/service/grpc"
	grpccred "google.golang.org/grpc/credentials"
)

// Server data manager server
type Server struct {
	microService    microsvc.Service
	microRegistry   registry.Registry
	tlsConfig       *tls.Config
	clientTLSConfig *tls.Config
	// http server
	httpServer    *http.Server
	opt           *DataManagerOptions
	handler       *handler.BcsDataManager
	producer      *worker.Producer
	consumer      *worker.Consumers
	store         store.Server
	ctx           context.Context
	ctxCancelFunc context.CancelFunc
	stopCtx       context.Context
	// extra module server, [pprof, metrics, swagger]
	extraServer *http.Server
}

// NewServer create server
func NewServer(ctx context.Context, cancel context.CancelFunc, opt *DataManagerOptions) *Server {
	return &Server{
		opt:           opt,
		ctx:           ctx,
		ctxCancelFunc: cancel,
	}
}

// Init init data manager server
func (s *Server) Init() error {
	// init server and client tls config
	if err := s.initTLSConfig(); err != nil {
		return err
	}
	// init model
	if err := s.initModel(); err != nil {
		return err
	}
	// init registry
	if err := s.initRegistry(); err != nil {
		return err
	}
	// init core micro service
	if err := s.initMicro(); err != nil {
		return err
	}
	// init http service
	if err := s.initHTTPService(); err != nil {
		return err
	}

	// init metric, pprof
	s.initExtraModules()
	// init system signal handler
	s.initSignalHandler()
	if err := s.initWorker(); err != nil {
		return err
	}

	return nil
}

// RunAsProducer server run as producer
func (s *Server) RunAsProducer() {
	s.producer.Run()
}

// RunAsConsumer server run as consumer
func (s *Server) RunAsConsumer() {
	s.consumer.Run()
}

// Run microservice run
func (s *Server) Run() {
	go func() {
		if err := s.microService.Run(); err != nil {
			blog.Fatalf("run data manager failed, err %s", err.Error())
		}
	}()
}

// StopProducer stop the producer
func (s *Server) StopProducer() {
	s.producer.Stop()
}

// StopConsumer stop the consumer
func (s *Server) StopConsumer() {
	s.consumer.Stop()
}

// init server and client tls config
func (s *Server) initTLSConfig() error {
	if len(s.opt.ServerCert) != 0 && len(s.opt.ServerKey) != 0 && len(s.opt.ServerCa) != 0 {
		tlsConfig, err := ssl.ServerTslConfVerityClient(s.opt.ServerCa, s.opt.ServerCert,
			s.opt.ServerKey, static.ServerCertPwd)
		if err != nil {
			blog.Errorf("load data manager server tls config failed, err %s", err.Error())
			return err
		}
		s.tlsConfig = tlsConfig
		blog.Infof("load data manager server tls config successfully")
	}

	if len(s.opt.ClientCert) != 0 && len(s.opt.ClientKey) != 0 && len(s.opt.ClientCa) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(s.opt.ClientCa, s.opt.ClientCert,
			s.opt.ClientKey, static.ClientCertPwd)
		if err != nil {
			blog.Errorf("load cluster manager client tls config failed, err %s", err.Error())
			return err
		}
		s.clientTLSConfig = tlsConfig
		blog.Infof("load data manager client tls config successfully")
	}
	return nil
}

// init mongo client
func (s *Server) initModel() error {
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
	}
	mongoDB, err := mongo.NewDB(mongoOptions)
	if err != nil {
		blog.Errorf("init mongo db failed, err %s", err.Error())
		return err
	}
	if err = mongoDB.Ping(); err != nil {
		blog.Errorf("ping mongo db failed, err %s", err.Error())
		return err
	}
	blog.Infof("init mongo db successfully")
	modelSet := store.NewServer(mongoDB)
	s.store = modelSet
	blog.Infof("init store successfully")
	return nil
}

func (s *Server) initRegistry() error {
	endpoints := strings.Replace(s.opt.Etcd.EtcdEndpoints, ";", ",", -1)
	etcdEndpoints := strings.Split(endpoints, ",")
	etcdSecure := false
	var etcdTLS *tls.Config
	var err error
	if len(s.opt.Etcd.EtcdCa) != 0 && len(s.opt.Etcd.EtcdCert) != 0 && len(s.opt.Etcd.EtcdKey) != 0 {
		etcdSecure = true
		etcdTLS, err = ssl.ClientTslConfVerity(s.opt.Etcd.EtcdCa, s.opt.Etcd.EtcdCert, s.opt.Etcd.EtcdKey, "")
		if err != nil {
			return err
		}
	}
	s.microRegistry = etcd.NewRegistry(
		registry.Addrs(etcdEndpoints...),
		registry.Secure(etcdSecure),
		registry.TLSConfig(etcdTLS),
	)
	if err := s.microRegistry.Init(); err != nil {
		return err
	}
	return nil
}

// init http grpc gateway
func (s *Server) initHTTPGateway(router *mux.Router) error {
	gwmux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			OrigName:     true,
			EmitDefaults: true,
		}),
	)
	grpcDialOpts := []grpc.DialOption{}
	if s.tlsConfig != nil && s.clientTLSConfig != nil {
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(grpccred.NewTLS(s.clientTLSConfig)))
	} else {
		grpcDialOpts = append(grpcDialOpts, grpc.WithInsecure())
	}
	err := datamanager.RegisterDataManagerGwFromEndpoint(
		context.TODO(),
		gwmux,
		s.opt.Address+":"+strconv.Itoa(int(s.opt.Port)),
		grpcDialOpts)
	if err != nil {
		blog.Errorf("register http gateway failed, err %s", err.Error())
		return fmt.Errorf("register http gateway failed, err %s", err.Error())
	}
	router.Handle("/{uri:.*}", gwmux)
	blog.Info("register grpc gateway handler to path /")
	return nil
}

func (s *Server) initHTTPService() error {
	router := mux.NewRouter()
	// init micro http gateway
	if err := s.initHTTPGateway(router); err != nil {
		return err
	}

	serverMux := http.NewServeMux()
	serverMux.Handle("/", router)

	httpAddr := s.opt.Address + ":" + strconv.Itoa(int(s.opt.HTTPPort))
	s.httpServer = &http.Server{
		Addr:    httpAddr,
		Handler: serverMux,
	}
	go func() {
		var err error
		blog.Infof("start http gateway server on address %s", httpAddr)
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

func (s *Server) initMicro() error {
	// New Service
	microService := microgrpcsvc.NewService(
		microsvc.Name(common.ServiceDomain),
		microsvc.Metadata(map[string]string{
			common.MicroMetaKeyHTTPPort: strconv.Itoa(int(s.opt.HTTPPort)),
		}),
		microgrpcsvc.WithTLS(s.tlsConfig),
		microsvc.Address(s.opt.Address+":"+strconv.Itoa(int(s.opt.Port))),
		microsvc.Registry(s.microRegistry),
		microsvc.Version(version.BcsVersion),
		microsvc.RegisterTTL(30*time.Second),
		microsvc.RegisterInterval(25*time.Second),
		microsvc.Context(s.ctx),
	)
	microService.Init()

	// create cluster manager server handler
	s.handler = handler.NewBcsDataManager(s.store)
	// Register handler
	err := datamanager.RegisterDataManagerHandler(microService.Server(), s.handler)
	if err != nil {
		blog.Errorf("RegisterDataManagerHandler error :%v", err)
		return err
	}
	s.microService = microService
	return nil
}

func (s *Server) initSignalHandler() {
	// listen system signal
	// to run in the container, should not trap SIGTERM
	interrupt := make(chan os.Signal, 10)
	signal.Notify(interrupt, syscall.SIGINT)
	go func() {
		select {
		case e := <-interrupt:
			blog.Infof("receive interrupt %s, do close", e.String())
			s.close()
		case <-s.ctx.Done():
			blog.Infof("stop channel, do close")
			s.close()
		}
	}()
}

func (s *Server) close() {
	closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer closeCancel()
	s.httpServer.Shutdown(closeCtx)
	s.extraServer.Shutdown(closeCtx)
	s.ctxCancelFunc()
}

func (s *Server) initWorker() error {
	bcsMonitorCli := s.initBcsMonitorCli()
	msgQueue, err := initQueue(s.opt.QueueConfig)
	if err != nil {
		blog.Errorf("init queue err:%v", err)
		return err
	}
	k8sStorageCli, mesosStorageCli, err := s.initStorageCli()
	if err != nil {
		blog.Errorf("init storage cli err:%v", err)
		return err
	}
	cmCli, err := s.initClusterManager()
	if err != nil {
		blog.Errorf("init cmCli err:%v", err)
		return err
	}
	// init producer
	producerCron := cron.New()
	selectClusters := strings.Split(s.opt.FilterRules.ClusterIDs, ",")
	blog.Infof("selected cluster: %v", selectClusters)
	resourceGetter := common.NewGetter(s.opt.FilterRules.NeedFilter, selectClusters)
	s.producer = worker.NewProducer(s.ctx, msgQueue, producerCron, cmCli, k8sStorageCli, mesosStorageCli, resourceGetter)
	if err = s.producer.InitCronList(); err != nil {
		blog.Errorf("init producer cron list error: %v", err)
		return err
	}
	handlerOpts := worker.HandlerOptions{ChanQueueNum: s.opt.HandleConfig.ChanQueueLen}
	handlerClients := worker.HandleClients{
		Store:            s.store,
		BcsMonitorClient: bcsMonitorCli,
		K8sStorageCli:    k8sStorageCli,
		MesosStorageCli:  mesosStorageCli,
		CmCli:            cmCli,
	}
	consumers := make([]worker.Consumer, 0)
	dataJobHandler := worker.NewDataJobHandler(handlerOpts, handlerClients, s.opt.HandleConfig.Concurrency)
	consumers = append(consumers, dataJobHandler)
	s.consumer = worker.NewConsumers(consumers, msgQueue)
	return nil
}

func initQueue(opts QueueConfig) (msgqueue.MessageQueue, error) {
	address := opts.QueueAddress
	schemas := strings.Split(address, "//")
	if len(schemas) != 2 {
		return nil, fmt.Errorf("passwd contain special char(//)")
	}
	accountServers := strings.Split(schemas[1], "@")
	if len(accountServers) != 2 {
		return nil, fmt.Errorf("queue account or passwd contain special char(@)")
	}
	accounts := strings.Split(accountServers[0], ":")
	if len(accounts) != 2 {
		return nil, fmt.Errorf("queue account or passwd contain special char(:)")
	}
	pwd := accounts[1]
	if pwd != "" {
		realPwd, _ := encrypt.DesDecryptFromBase([]byte(pwd))
		pwd = string(realPwd)
	}
	parseAddress := fmt.Sprintf("%s//%s:%s@%s", schemas[0], accounts[0], pwd, accountServers[1])
	commonOption := msgqueue.CommonOpts(&msgqueue.CommonOptions{
		QueueFlag:       opts.QueueFlag,
		QueueKind:       msgqueue.QueueKind("rabbitmq"),
		ResourceToQueue: map[string]string{common.DataJobQueue: common.DataJobQueue},
		Address:         parseAddress,
	})
	exchangeOption := msgqueue.Exchange(
		&msgqueue.ExchangeOptions{
			Name:           opts.ExchangeName,
			Durable:        true,
			PrefetchCount:  30,
			PrefetchGlobal: true,
		})
	natStreamingOption := msgqueue.NatsOpts(
		&msgqueue.NatsOptions{
			ClusterID:      opts.ClusterID,
			ConnectTimeout: time.Duration(opts.ConnectTimeout) * time.Second,
			ConnectRetry:   opts.ConnectRetry,
		})
	publishOption := msgqueue.PublishOpts(
		&msgqueue.PublishOptions{
			TopicName:    common.DataJobQueue,
			DeliveryMode: uint8(opts.PublishDelivery),
		})
	arguments := make(map[string]interface{})
	queueArgumentsRaw := opts.QueueArguments
	queueArguments := strings.Split(queueArgumentsRaw, ";")
	if len(queueArguments) > 0 {
		for _, data := range queueArguments {
			dList := strings.Split(data, ":")
			if len(dList) == 2 {
				arguments[dList[0]] = dList[1]
			}
		}
	}
	subscribeOption := msgqueue.SubscribeOpts(
		&msgqueue.SubscribeOptions{
			TopicName:         common.DataJobQueue,
			QueueName:         common.DataJobQueue,
			DisableAutoAck:    true,
			Durable:           true,
			AckOnSuccess:      true,
			RequeueOnError:    true,
			DeliverAllMessage: true,
			ManualAckMode:     true,
			EnableAckWait:     true,
			AckWaitDuration:   time.Duration(30) * time.Second,
			MaxInFlight:       0,
			QueueArguments: map[string]interface{}{
				"x-message-ttl": 1800000,
			},
		})
	msgQueue, err := msgqueue.NewMsgQueue(commonOption, exchangeOption, natStreamingOption, publishOption, subscribeOption)
	if err != nil {
		msgErr := fmt.Errorf("create queue failed, err %s", err.Error())
		blog.Errorf("create queue failed, err %s", err.Error())
		return nil, msgErr
	}
	blog.Infof("init queue successfully, sub queue[dataJob]")
	return msgQueue, nil
}

func (s *Server) initClusterManager() (*cmanager.ClusterManagerClient, error) {
	realAuthToken, _ := encrypt.DesDecryptFromBase([]byte(s.opt.BcsAPIConf.AdminToken))
	opts := &cmanager.Options{
		Module:          cmanager.ModuleClusterManager,
		Address:         s.opt.BcsAPIConf.GrpcGWAddress,
		EtcdRegistry:    s.microRegistry,
		ClientTLSConfig: s.clientTLSConfig,
		AuthToken:       string(realAuthToken),
	}

	cli := cmanager.NewClusterManagerClient(opts)
	if cli == nil {
		errMsg := fmt.Errorf("initClusterManager failed")
		return nil, errMsg
	}
	cmConn, err := cli.GetClusterManagerConn()
	if err != nil || cmConn == nil {
		return nil, fmt.Errorf("get cluster manager client error:%v", err)
	}
	cmCli := cli.NewGrpcClientWithHeader(s.ctx, cmConn)
	_, err = cmCli.Cli.ListCluster(cmCli.Ctx, &cm.ListClusterReq{})
	if err != nil {
		return nil, fmt.Errorf("dial cm failed:%v", err)
	}
	return cli, nil
}

func (s *Server) initBcsMonitorCli() bcsmonitor.ClientInterface {
	realPassword, _ := encrypt.DesDecryptFromBase([]byte(s.opt.BcsMonitorConf.Password))
	realAppSecret, _ := encrypt.DesDecryptFromBase([]byte(s.opt.AppSecret))
	bcsMonitorOpts := bcsmonitor.BcsMonitorClientOpt{
		Schema:    s.opt.BcsMonitorConf.Schema,
		Endpoint:  s.opt.BcsMonitorConf.BcsMonitorEndpoints,
		UserName:  s.opt.BcsMonitorConf.User,
		Password:  string(realPassword),
		AppCode:   s.opt.AppCode,
		AppSecret: string(realAppSecret),
	}
	bcsMonitorRequester := bcsmonitor.NewRequester()
	bcsMonitorCli := bcsmonitor.NewBcsMonitorClient(bcsMonitorOpts, bcsMonitorRequester)
	bcsMonitorCli.SetCompleteEndpoint()
	bcsMonitorCli.SetDefaultHeader(http.Header{})
	blog.Infof("init monitor cli success")
	return bcsMonitorCli
}

func (s *Server) initStorageCli() (bcsapi.Storage, bcsapi.Storage, error) {
	realAuthToken, _ := encrypt.DesDecryptFromBase([]byte(s.opt.BcsAPIConf.AdminToken))
	k8sStorageConfig := &bcsapi.Config{
		Hosts:     []string{s.opt.BcsAPIConf.BcsAPIGwURL},
		TLSConfig: s.clientTLSConfig,
		AuthToken: string(realAuthToken),
		Gateway:   true,
	}
	k8sStorageCli := &bcsapi.StorageCli{
		Config: k8sStorageConfig,
	}
	if k8sStorageConfig.TLSConfig != nil {
		k8sStorageCli.Client = restclient.NewRESTClientWithTLS(k8sStorageConfig.TLSConfig)
	} else {
		k8sStorageCli.Client = restclient.NewRESTClient()
	}
	k8sTransport := &http.Transport{}
	k8sStorageCli.Client.WithTransport(k8sTransport)
	_, err := k8sStorageCli.QueryK8SDeployment("test", "test")
	if err != nil {
		blog.Errorf("init k8s storage cli error: %v", err)
		return nil, nil, err
	}
	blog.Infof("init k8s storage cli success")

	mesosStorageConfig := &bcsapi.Config{
		Hosts:     []string{s.opt.BcsAPIConf.OldBcsAPIGwURL},
		AuthToken: string(realAuthToken),
		Gateway:   true,
	}
	mesosStorageCli := &bcsapi.StorageCli{
		Config: mesosStorageConfig,
	}
	mesosStorageCli.Client = restclient.NewRESTClient()
	mesosTransport := &http.Transport{}
	mesosStorageCli.Client.WithTransport(mesosTransport)
	_, err = mesosStorageCli.QueryMesosDeployment("test")
	if err != nil {
		blog.Errorf("init mesos storage cli error: %v", err)
		return nil, nil, err
	}
	blog.Infof("init mesos storage cli success")
	return k8sStorageCli, mesosStorageCli, nil
}

func (s *Server) initExtraModules() {
	extraMux := http.NewServeMux()
	s.initPProf(extraMux)
	s.initMetric(extraMux)
	extraServerEndpoint := s.opt.Address + ":" + strconv.Itoa(int(s.opt.MetricPort))
	s.extraServer = &http.Server{
		Addr:    extraServerEndpoint,
		Handler: extraMux,
	}

	go func() {
		var err error
		blog.Infof("start extra modules [pprof, metric] server %s", extraServerEndpoint)
		err = s.extraServer.ListenAndServe()
		if err != nil {
			blog.Errorf("extra modules server listen failed, err %s", err.Error())
			s.ctxCancelFunc()
		}
	}()
}

func (s *Server) initPProf(mux *http.ServeMux) {
	if !s.opt.Debug {
		blog.Infof("pprof is disabled")
		return
	}
	blog.Infof("pprof is enabled")
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
}

func (s *Server) initMetric(mux *http.ServeMux) {
	blog.Infof("init metric handler")
	mux.Handle("/metrics", promhttp.Handler())
}
