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
	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi"
	cm "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	restclient "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/client"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/go-micro/plugins/v4/registry/etcd"
	microGrpcServer "github.com/go-micro/plugins/v4/server/grpc"
	_ "github.com/go-sql-driver/mysql" // nolint
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
	microsvc "go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"google.golang.org/grpc"
	grpccred "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/cmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	kafka2 "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/kafka"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/requester"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store"
	dmmongo "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store/mongo"
	dmtspider "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/store/tspider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/worker"
	datamanager "github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/proto/bcs-data-manager"
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
	mongoStore    store.Server
	tspiderStore  store.Server
	ctx           context.Context
	ctxCancelFunc context.CancelFunc
	// extra module server, [pprof, metrics, swagger]
	extraServer    *http.Server
	resourceGetter common.GetterInterface
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

	if err := s.initWorker(); err != nil {
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

// WaitForConsumerDone wait for consumer finish
func (s *Server) WaitForConsumerDone() {
	s.consumer.Done()
}

// initTLSConfig xxx
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

// initModel xxx
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

	// init tspiderConfig
	tspiderConfig, err := s.opt.ParseTspiderConfig()
	if err != nil {
		blog.Errorf("init tspider config failed, err %s", err.Error())
		return err
	}
	tspiderDBs := make(map[string]*sqlx.DB, 0)
	for _, conf := range tspiderConfig {
		dsn := conf.Connection
		db, err := sqlx.Connect("mysql", dsn)
		if err != nil {
			blog.Errorf("init tspider db(%s) failed, err %s", dsn, err.Error())
			return err
		}

		tspiderDBs[conf.StoreName] = db
	}
	blog.Infof("init tspider db successfully")

	// init bkbaseConfig
	bkbaseConfig, err := s.opt.ParseBkbaseConfig()
	if err != nil {
		blog.Errorf("init bkbase config failed, err %s", err.Error())
		return err
	}
	blog.Infof("init bkbaseConfig successfully")

	// set server's mongo and tspider model
	mongoModelSet := dmmongo.NewServer(mongoDB, bkbaseConfig)
	s.mongoStore = mongoModelSet
	spiderModelSet := dmtspider.NewServer(tspiderDBs, bkbaseConfig)
	s.tspiderStore = spiderModelSet
	blog.Infof("init store successfully")
	return nil
}

// initRegistry init micro service registry
func (s *Server) initRegistry() error {
	endpoints := strings.ReplaceAll(s.opt.Etcd.EtcdEndpoints, ";", ",")
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

// initHTTPGateway xxx
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
		grpcDialOpts = append(grpcDialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
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

// initHTTPService xxx
// init http service
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

// initMicro xxx
// init micro service
func (s *Server) initMicro() error {

	// with tls
	grpcSvr := microGrpcServer.NewServer(microGrpcServer.AuthTLS(s.tlsConfig))

	// New Service
	microService := microsvc.NewService(
		microsvc.Server(grpcSvr),
		microsvc.Cmd(util.NewDummyMicroCmd()),
		microsvc.Name(types.ServiceDomain),
		microsvc.Metadata(map[string]string{
			types.MicroMetaKeyHTTPPort: strconv.Itoa(int(s.opt.HTTPPort)),
		}),
		microsvc.Address(s.opt.Address+":"+strconv.Itoa(int(s.opt.Port))),
		microsvc.Registry(s.microRegistry),
		microsvc.Version(version.BcsVersion),
		microsvc.RegisterTTL(30*time.Second),
		microsvc.RegisterInterval(25*time.Second),
		microsvc.Context(s.ctx),
	)
	microService.Init()

	// create cluster manager server handler
	s.handler = handler.NewBcsDataManager(s.mongoStore, s.tspiderStore, s.resourceGetter)
	// Register handler
	err := datamanager.RegisterDataManagerHandler(microService.Server(), s.handler)
	if err != nil {
		blog.Errorf("RegisterDataManagerHandler error :%v", err)
		return err
	}
	s.microService = microService
	return nil
}

// initSignalHandler xxx
// init signal handler
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

// close service
func (s *Server) close() {
	closeCtx, closeCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer closeCancel()
	_ = s.httpServer.Shutdown(closeCtx)
	_ = s.extraServer.Shutdown(closeCtx)
	s.ctxCancelFunc()
}

// initWorker xxx
// init producer and worker
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

	selectClusters := strings.Split(s.opt.FilterRules.ClusterIDs, ",")
	blog.Infof("selected cluster: %v", selectClusters)
	blog.Infof("cluster env: %s", s.opt.FilterRules.Env)

	pmClient := s.initProjectManager()
	if pmClient == nil {
		blog.Errorf("init project manager cli error, client is nil")
		return fmt.Errorf("init project manager cli error, client is nil")
	}
	// init resourceGetter
	s.resourceGetter = common.NewGetter(s.opt.FilterRules.NeedFilter, selectClusters, s.opt.FilterRules.Env,
		pmClient, bcsMonitorCli)
	// init producer
	producerCron := cron.New()
	s.producer = worker.NewProducer(s.ctx, msgQueue, producerCron, cmCli, k8sStorageCli, mesosStorageCli,
		s.resourceGetter, s.opt.ProducerConfig.Concurrency, s.opt.NeedSendKafka)
	if err = s.producer.InitCronList(); err != nil {
		blog.Errorf("init producer cron list error: %v", err)
		return err
	}
	if s.opt.NeedSendKafka {
		if err = s.initKafkaConn(); err != nil {
			return err
		}
	}
	// init consumer
	handlerOpts := worker.HandlerOptions{ChanQueueNum: s.opt.HandleConfig.ChanQueueLen}
	handlerClients := worker.HandleClients{
		Store:            s.mongoStore,
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

// initQueue xxx
// init message queue, use rabbit mq
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
		ResourceToQueue: map[string]string{types.DataJobQueue: types.DataJobQueue},
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
			TopicName:    types.DataJobQueue,
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
			TopicName:         types.DataJobQueue,
			QueueName:         types.DataJobQueue,
			DisableAutoAck:    true,
			Durable:           true,
			AckOnSuccess:      true,
			RequeueOnError:    true,
			DeliverAllMessage: true,
			ManualAckMode:     true,
			EnableAckWait:     true,
			AckWaitDuration:   time.Duration(30) * time.Second,
			MaxInFlight:       0,
			QueueArguments:    map[string]interface{}{"x-message-ttl": 1800000},
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

// initClusterManager xxx
// init cluster manager cli
func (s *Server) initClusterManager() (cmanager.ClusterManagerClient, error) {
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

// initBcsMonitorCli xxx
// init bcs monitor/storeGW cli
func (s *Server) initBcsMonitorCli() bcsmonitor.ClientInterface {
	var realAppSecret []byte
	if s.opt.AppSecret != "" {
		realAppSecret, _ = encrypt.DesDecryptFromBase([]byte(s.opt.AppSecret))
	}
	realAuthToken, _ := encrypt.DesDecryptFromBase([]byte(s.opt.BcsAPIConf.AdminToken))
	bcsMonitorOpts := bcsmonitor.BcsMonitorClientOpt{
		Endpoint:  s.opt.BcsMonitorConf.BcsMonitorEndpoints,
		AppCode:   s.opt.AppCode,
		AppSecret: string(realAppSecret),
	}
	bcsMonitorRequester := requester.NewRequester()
	bcsMonitorCli := bcsmonitor.NewBcsMonitorClient(bcsMonitorOpts, bcsMonitorRequester)
	defaultHeader := http.Header{}
	defaultHeader.Add("Authorization", fmt.Sprintf("Bearer %s", realAuthToken))
	bcsMonitorCli.SetDefaultHeader(defaultHeader)
	blog.Infof("init monitor cli success")
	return bcsMonitorCli
}

// initStorageCli xxx
// init bcs storage cli
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
	k8sTransport.TLSClientConfig = s.clientTLSConfig
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

// initExtraModules xxx
// init pprof and metric
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

// initPProf xxx
// init pprof
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

// initMetric xxx
// init metric
func (s *Server) initMetric(mux *http.ServeMux) {
	blog.Infof("init metric handler")
	mux.Handle("/metrics", promhttp.Handler())
}

// initProjectManager init project manager client
func (s *Server) initProjectManager() bcsproject.BcsProjectManagerClient {
	realAuthToken, _ := encrypt.DesDecryptFromBase([]byte(s.opt.BcsAPIConf.AdminToken))
	opts := &bcsproject.Options{
		Module:          bcsproject.ModuleProjectManager,
		Address:         s.opt.BcsAPIConf.GrpcGWAddress,
		EtcdRegistry:    s.microRegistry,
		ClientTLSConfig: s.clientTLSConfig,
		AuthToken:       string(realAuthToken),
		UserName:        s.opt.BcsAPIConf.UserName,
	}
	return bcsproject.NewBcsProjectManagerClient(opts)
}

// initKafkaConn init kafka connection cli
func (s *Server) initKafkaConn() error {
	password := s.opt.KafkaConfig.Password
	if password != "" {
		realPwd, _ := encrypt.DesDecryptFromBase([]byte(password))
		password = string(realPwd)
	}
	mechanism, err := scram.Mechanism(scram.SHA512, s.opt.KafkaConfig.Username, password)
	if err != nil {
		blog.Errorf("init kafka mechanism error :%s", err.Error())
		return fmt.Errorf("init kafka mechanism error :%s", err.Error())
	}
	sharedTransport := &kafka.Transport{
		SASL: mechanism,
	}
	writer := &kafka.Writer{
		Addr: kafka.TCP(s.opt.KafkaConfig.Address),
		//Topic:                  "datamanager",
		MaxAttempts:            3,
		AllowAutoTopicCreation: true,
		Transport:              sharedTransport,
	}
	kafkaConn := kafka2.NewKafkaClient(writer, nil, s.opt.KafkaConfig.Topic)
	s.producer.ImportKafkaConn(kafkaConn)
	return nil
}
