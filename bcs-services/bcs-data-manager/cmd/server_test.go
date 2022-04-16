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
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/common"
	"github.com/stretchr/testify/assert"
)

//
// func TestServer_Init(t *testing.T) {
//	type fields struct {
//		microService    service.Service
//		microRegistry   registry.Registry
//		tlsConfig       *tls.Config
//		clientTLSConfig *tls.Config
//		httpServer      *http.Server
//		opt             *DataManagerOptions
//		handler         *handler.BcsDataManager
//		producer        *worker.Producer
//		consumer        *worker.Consumers
//		store           store.Server
//		cron            *cron.Cron
//		ctx             context.Context
//		ctxCancelFunc   context.CancelFunc
//		stopCh          chan struct{}
//		leaderCh        chan sync.Leader
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				microService:    tt.fields.microService,
//				microRegistry:   tt.fields.microRegistry,
//				tlsConfig:       tt.fields.tlsConfig,
//				clientTLSConfig: tt.fields.clientTLSConfig,
//				httpServer:      tt.fields.httpServer,
//				opt:             tt.fields.opt,
//				handler:         tt.fields.handler,
//				producer:        tt.fields.producer,
//				consumer:        tt.fields.consumer,
//				store:           tt.fields.store,
//				cron:            tt.fields.cron,
//				ctx:             tt.fields.ctx,
//				ctxCancelFunc:   tt.fields.ctxCancelFunc,
//				stopCh:          tt.fields.stopCh,
//				leaderCh:        tt.fields.leaderCh,
//			}
//			if err := s.Init(); (err != nil) != tt.wantErr {
//				t.Errorf("Init() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
// }
//
// func TestServer_RunAsConsumer(t *testing.T) {
//	type fields struct {
//		microService    service.Service
//		microRegistry   registry.Registry
//		tlsConfig       *tls.Config
//		clientTLSConfig *tls.Config
//		httpServer      *http.Server
//		opt             *DataManagerOptions
//		handler         *handler.BcsDataManager
//		producer        *worker.Producer
//		consumer        *worker.Consumers
//		store           store.Server
//		cron            *cron.Cron
//		ctx             context.Context
//		ctxCancelFunc   context.CancelFunc
//		stopCh          chan struct{}
//		leaderCh        chan sync.Leader
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				microService:    tt.fields.microService,
//				microRegistry:   tt.fields.microRegistry,
//				tlsConfig:       tt.fields.tlsConfig,
//				clientTLSConfig: tt.fields.clientTLSConfig,
//				httpServer:      tt.fields.httpServer,
//				opt:             tt.fields.opt,
//				handler:         tt.fields.handler,
//				producer:        tt.fields.producer,
//				consumer:        tt.fields.consumer,
//				store:           tt.fields.store,
//				cron:            tt.fields.cron,
//				ctx:             tt.fields.ctx,
//				ctxCancelFunc:   tt.fields.ctxCancelFunc,
//				stopCh:          tt.fields.stopCh,
//				leaderCh:        tt.fields.leaderCh,
//			}
//		})
//	}
// }
//
// func TestServer_RunAsProducer(t *testing.T) {
//	type fields struct {
//		microService    service.Service
//		microRegistry   registry.Registry
//		tlsConfig       *tls.Config
//		clientTLSConfig *tls.Config
//		httpServer      *http.Server
//		opt             *DataManagerOptions
//		handler         *handler.BcsDataManager
//		producer        *worker.Producer
//		consumer        *worker.Consumers
//		store           store.Server
//		cron            *cron.Cron
//		ctx             context.Context
//		ctxCancelFunc   context.CancelFunc
//		stopCh          chan struct{}
//		leaderCh        chan sync.Leader
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				microService:    tt.fields.microService,
//				microRegistry:   tt.fields.microRegistry,
//				tlsConfig:       tt.fields.tlsConfig,
//				clientTLSConfig: tt.fields.clientTLSConfig,
//				httpServer:      tt.fields.httpServer,
//				opt:             tt.fields.opt,
//				handler:         tt.fields.handler,
//				producer:        tt.fields.producer,
//				consumer:        tt.fields.consumer,
//				store:           tt.fields.store,
//				cron:            tt.fields.cron,
//				ctx:             tt.fields.ctx,
//				ctxCancelFunc:   tt.fields.ctxCancelFunc,
//				stopCh:          tt.fields.stopCh,
//				leaderCh:        tt.fields.leaderCh,
//			}
//		})
//	}
// }
//
// func TestServer_Stop(t *testing.T) {
//	type fields struct {
//		microService    service.Service
//		microRegistry   registry.Registry
//		tlsConfig       *tls.Config
//		clientTLSConfig *tls.Config
//		httpServer      *http.Server
//		opt             *DataManagerOptions
//		handler         *handler.BcsDataManager
//		producer        *worker.Producer
//		consumer        *worker.Consumers
//		store           store.Server
//		cron            *cron.Cron
//		ctx             context.Context
//		ctxCancelFunc   context.CancelFunc
//		stopCh          chan struct{}
//		leaderCh        chan sync.Leader
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				microService:    tt.fields.microService,
//				microRegistry:   tt.fields.microRegistry,
//				tlsConfig:       tt.fields.tlsConfig,
//				clientTLSConfig: tt.fields.clientTLSConfig,
//				httpServer:      tt.fields.httpServer,
//				opt:             tt.fields.opt,
//				handler:         tt.fields.handler,
//				producer:        tt.fields.producer,
//				consumer:        tt.fields.consumer,
//				store:           tt.fields.store,
//				cron:            tt.fields.cron,
//				ctx:             tt.fields.ctx,
//				ctxCancelFunc:   tt.fields.ctxCancelFunc,
//				stopCh:          tt.fields.stopCh,
//				leaderCh:        tt.fields.leaderCh,
//			}
//		})
//	}
// }
//
// func TestServer_close(t *testing.T) {
//	type fields struct {
//		microService    service.Service
//		microRegistry   registry.Registry
//		tlsConfig       *tls.Config
//		clientTLSConfig *tls.Config
//		httpServer      *http.Server
//		opt             *DataManagerOptions
//		handler         *handler.BcsDataManager
//		producer        *worker.Producer
//		consumer        *worker.Consumers
//		store           store.Server
//		cron            *cron.Cron
//		ctx             context.Context
//		ctxCancelFunc   context.CancelFunc
//		stopCh          chan struct{}
//		leaderCh        chan sync.Leader
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				microService:    tt.fields.microService,
//				microRegistry:   tt.fields.microRegistry,
//				tlsConfig:       tt.fields.tlsConfig,
//				clientTLSConfig: tt.fields.clientTLSConfig,
//				httpServer:      tt.fields.httpServer,
//				opt:             tt.fields.opt,
//				handler:         tt.fields.handler,
//				producer:        tt.fields.producer,
//				consumer:        tt.fields.consumer,
//				store:           tt.fields.store,
//				cron:            tt.fields.cron,
//				ctx:             tt.fields.ctx,
//				ctxCancelFunc:   tt.fields.ctxCancelFunc,
//				stopCh:          tt.fields.stopCh,
//				leaderCh:        tt.fields.leaderCh,
//			}
//		})
//	}
// }
//
// func TestServer_initHTTPGateway(t *testing.T) {
//	type fields struct {
//		microService    service.Service
//		microRegistry   registry.Registry
//		tlsConfig       *tls.Config
//		clientTLSConfig *tls.Config
//		httpServer      *http.Server
//		opt             *DataManagerOptions
//		handler         *handler.BcsDataManager
//		producer        *worker.Producer
//		consumer        *worker.Consumers
//		store           store.Server
//		cron            *cron.Cron
//		ctx             context.Context
//		ctxCancelFunc   context.CancelFunc
//		stopCh          chan struct{}
//		leaderCh        chan sync.Leader
//	}
//	type args struct {
//		router *mux.Router
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				microService:    tt.fields.microService,
//				microRegistry:   tt.fields.microRegistry,
//				tlsConfig:       tt.fields.tlsConfig,
//				clientTLSConfig: tt.fields.clientTLSConfig,
//				httpServer:      tt.fields.httpServer,
//				opt:             tt.fields.opt,
//				handler:         tt.fields.handler,
//				producer:        tt.fields.producer,
//				consumer:        tt.fields.consumer,
//				store:           tt.fields.store,
//				cron:            tt.fields.cron,
//				ctx:             tt.fields.ctx,
//				ctxCancelFunc:   tt.fields.ctxCancelFunc,
//				stopCh:          tt.fields.stopCh,
//				leaderCh:        tt.fields.leaderCh,
//			}
//			if err := s.initHTTPGateway(tt.args.router); (err != nil) != tt.wantErr {
//				t.Errorf("initHTTPGateway() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
// }
//
// func TestServer_initHTTPService(t *testing.T) {
//	type fields struct {
//		microService    service.Service
//		microRegistry   registry.Registry
//		tlsConfig       *tls.Config
//		clientTLSConfig *tls.Config
//		httpServer      *http.Server
//		opt             *DataManagerOptions
//		handler         *handler.BcsDataManager
//		producer        *worker.Producer
//		consumer        *worker.Consumers
//		store           store.Server
//		cron            *cron.Cron
//		ctx             context.Context
//		ctxCancelFunc   context.CancelFunc
//		stopCh          chan struct{}
//		leaderCh        chan sync.Leader
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				microService:    tt.fields.microService,
//				microRegistry:   tt.fields.microRegistry,
//				tlsConfig:       tt.fields.tlsConfig,
//				clientTLSConfig: tt.fields.clientTLSConfig,
//				httpServer:      tt.fields.httpServer,
//				opt:             tt.fields.opt,
//				handler:         tt.fields.handler,
//				producer:        tt.fields.producer,
//				consumer:        tt.fields.consumer,
//				store:           tt.fields.store,
//				cron:            tt.fields.cron,
//				ctx:             tt.fields.ctx,
//				ctxCancelFunc:   tt.fields.ctxCancelFunc,
//				stopCh:          tt.fields.stopCh,
//				leaderCh:        tt.fields.leaderCh,
//			}
//			if err := s.initHTTPService(); (err != nil) != tt.wantErr {
//				t.Errorf("initHTTPService() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
// }
//
// func TestServer_initLeader(t *testing.T) {
//	type fields struct {
//		microService    service.Service
//		microRegistry   registry.Registry
//		tlsConfig       *tls.Config
//		clientTLSConfig *tls.Config
//		httpServer      *http.Server
//		opt             *DataManagerOptions
//		handler         *handler.BcsDataManager
//		producer        *worker.Producer
//		consumer        *worker.Consumers
//		store           store.Server
//		cron            *cron.Cron
//		ctx             context.Context
//		ctxCancelFunc   context.CancelFunc
//		stopCh          chan struct{}
//		leaderCh        chan sync.Leader
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				microService:    tt.fields.microService,
//				microRegistry:   tt.fields.microRegistry,
//				tlsConfig:       tt.fields.tlsConfig,
//				clientTLSConfig: tt.fields.clientTLSConfig,
//				httpServer:      tt.fields.httpServer,
//				opt:             tt.fields.opt,
//				handler:         tt.fields.handler,
//				producer:        tt.fields.producer,
//				consumer:        tt.fields.consumer,
//				store:           tt.fields.store,
//				cron:            tt.fields.cron,
//				ctx:             tt.fields.ctx,
//				ctxCancelFunc:   tt.fields.ctxCancelFunc,
//				stopCh:          tt.fields.stopCh,
//				leaderCh:        tt.fields.leaderCh,
//			}
//		})
//	}
// }
//
// func TestServer_initMicro(t *testing.T) {
//	type fields struct {
//		microService    service.Service
//		microRegistry   registry.Registry
//		tlsConfig       *tls.Config
//		clientTLSConfig *tls.Config
//		httpServer      *http.Server
//		opt             *DataManagerOptions
//		handler         *handler.BcsDataManager
//		producer        *worker.Producer
//		consumer        *worker.Consumers
//		store           store.Server
//		cron            *cron.Cron
//		ctx             context.Context
//		ctxCancelFunc   context.CancelFunc
//		stopCh          chan struct{}
//		leaderCh        chan sync.Leader
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				microService:    tt.fields.microService,
//				microRegistry:   tt.fields.microRegistry,
//				tlsConfig:       tt.fields.tlsConfig,
//				clientTLSConfig: tt.fields.clientTLSConfig,
//				httpServer:      tt.fields.httpServer,
//				opt:             tt.fields.opt,
//				handler:         tt.fields.handler,
//				producer:        tt.fields.producer,
//				consumer:        tt.fields.consumer,
//				store:           tt.fields.store,
//				cron:            tt.fields.cron,
//				ctx:             tt.fields.ctx,
//				ctxCancelFunc:   tt.fields.ctxCancelFunc,
//				stopCh:          tt.fields.stopCh,
//				leaderCh:        tt.fields.leaderCh,
//			}
//			if err := s.initMicro(); (err != nil) != tt.wantErr {
//				t.Errorf("initMicro() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
// }

func TestServer_initModel(t *testing.T) {
	mongoOptions := &mongo.Options{
		Hosts:                 []string{"127.0.0.1:27017"},
		ConnectTimeoutSeconds: 3,
		Database:              "datamanager_test",
		Username:              "data",
		Password:              "test1234",
	}
	mongoDB, err := mongo.NewDB(mongoOptions)
	assert.Equal(t, nil, err)
	err = mongoDB.Ping()
	assert.Equal(t, nil, err)
	fmt.Println("init mongo db successfully")
}

//
// func TestServer_initRegistry(t *testing.T) {
//	type fields struct {
//		microService    service.Service
//		microRegistry   registry.Registry
//		tlsConfig       *tls.Config
//		clientTLSConfig *tls.Config
//		httpServer      *http.Server
//		opt             *DataManagerOptions
//		handler         *handler.BcsDataManager
//		producer        *worker.Producer
//		consumer        *worker.Consumers
//		store           store.Server
//		cron            *cron.Cron
//		ctx             context.Context
//		ctxCancelFunc   context.CancelFunc
//		stopCh          chan struct{}
//		leaderCh        chan sync.Leader
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				microService:    tt.fields.microService,
//				microRegistry:   tt.fields.microRegistry,
//				tlsConfig:       tt.fields.tlsConfig,
//				clientTLSConfig: tt.fields.clientTLSConfig,
//				httpServer:      tt.fields.httpServer,
//				opt:             tt.fields.opt,
//				handler:         tt.fields.handler,
//				producer:        tt.fields.producer,
//				consumer:        tt.fields.consumer,
//				store:           tt.fields.store,
//				cron:            tt.fields.cron,
//				ctx:             tt.fields.ctx,
//				ctxCancelFunc:   tt.fields.ctxCancelFunc,
//				stopCh:          tt.fields.stopCh,
//				leaderCh:        tt.fields.leaderCh,
//			}
//			if err := s.initRegistry(); (err != nil) != tt.wantErr {
//				t.Errorf("initRegistry() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
// }
//
// func TestServer_initSignalHandler(t *testing.T) {
//	type fields struct {
//		microService    service.Service
//		microRegistry   registry.Registry
//		tlsConfig       *tls.Config
//		clientTLSConfig *tls.Config
//		httpServer      *http.Server
//		opt             *DataManagerOptions
//		handler         *handler.BcsDataManager
//		producer        *worker.Producer
//		consumer        *worker.Consumers
//		store           store.Server
//		cron            *cron.Cron
//		ctx             context.Context
//		ctxCancelFunc   context.CancelFunc
//		stopCh          chan struct{}
//		leaderCh        chan sync.Leader
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				microService:    tt.fields.microService,
//				microRegistry:   tt.fields.microRegistry,
//				tlsConfig:       tt.fields.tlsConfig,
//				clientTLSConfig: tt.fields.clientTLSConfig,
//				httpServer:      tt.fields.httpServer,
//				opt:             tt.fields.opt,
//				handler:         tt.fields.handler,
//				producer:        tt.fields.producer,
//				consumer:        tt.fields.consumer,
//				store:           tt.fields.store,
//				cron:            tt.fields.cron,
//				ctx:             tt.fields.ctx,
//				ctxCancelFunc:   tt.fields.ctxCancelFunc,
//				stopCh:          tt.fields.stopCh,
//				leaderCh:        tt.fields.leaderCh,
//			}
//		})
//	}
// }
//
// func TestServer_initTLSConfig(t *testing.T) {
//	type fields struct {
//		microService    service.Service
//		microRegistry   registry.Registry
//		tlsConfig       *tls.Config
//		clientTLSConfig *tls.Config
//		httpServer      *http.Server
//		opt             *DataManagerOptions
//		handler         *handler.BcsDataManager
//		producer        *worker.Producer
//		consumer        *worker.Consumers
//		store           store.Server
//		cron            *cron.Cron
//		ctx             context.Context
//		ctxCancelFunc   context.CancelFunc
//		stopCh          chan struct{}
//		leaderCh        chan sync.Leader
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				microService:    tt.fields.microService,
//				microRegistry:   tt.fields.microRegistry,
//				tlsConfig:       tt.fields.tlsConfig,
//				clientTLSConfig: tt.fields.clientTLSConfig,
//				httpServer:      tt.fields.httpServer,
//				opt:             tt.fields.opt,
//				handler:         tt.fields.handler,
//				producer:        tt.fields.producer,
//				consumer:        tt.fields.consumer,
//				store:           tt.fields.store,
//				cron:            tt.fields.cron,
//				ctx:             tt.fields.ctx,
//				ctxCancelFunc:   tt.fields.ctxCancelFunc,
//				stopCh:          tt.fields.stopCh,
//				leaderCh:        tt.fields.leaderCh,
//			}
//			if err := s.initTLSConfig(); (err != nil) != tt.wantErr {
//				t.Errorf("initTLSConfig() error = %v, wantErr %v", err, tt.wantErr)
//			}
//		})
//	}
// }
//
// func TestServer_initWorker(t *testing.T) {
//	type fields struct {
//		microService    service.Service
//		microRegistry   registry.Registry
//		tlsConfig       *tls.Config
//		clientTLSConfig *tls.Config
//		httpServer      *http.Server
//		opt             *DataManagerOptions
//		handler         *handler.BcsDataManager
//		producer        *worker.Producer
//		consumer        *worker.Consumers
//		store           store.Server
//		cron            *cron.Cron
//		ctx             context.Context
//		ctxCancelFunc   context.CancelFunc
//		stopCh          chan struct{}
//		leaderCh        chan sync.Leader
//	}
//	tests := []struct {
//		name   string
//		fields fields
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			s := &Server{
//				microService:    tt.fields.microService,
//				microRegistry:   tt.fields.microRegistry,
//				tlsConfig:       tt.fields.tlsConfig,
//				clientTLSConfig: tt.fields.clientTLSConfig,
//				httpServer:      tt.fields.httpServer,
//				opt:             tt.fields.opt,
//				handler:         tt.fields.handler,
//				producer:        tt.fields.producer,
//				consumer:        tt.fields.consumer,
//				store:           tt.fields.store,
//				cron:            tt.fields.cron,
//				ctx:             tt.fields.ctx,
//				ctxCancelFunc:   tt.fields.ctxCancelFunc,
//				stopCh:          tt.fields.stopCh,
//				leaderCh:        tt.fields.leaderCh,
//			}
//		})
//	}
// }

func Test_initQueue(t *testing.T) {
	commonOption := msgqueue.CommonOpts(&msgqueue.CommonOptions{
		QueueFlag:       true,
		QueueKind:       msgqueue.QueueKind("rabbitmq"),
		ResourceToQueue: map[string]string{common.DataJobQueue: common.DataJobQueue},
		Address:         "amqp://root:123456@127.0.0.1:5672",
	})
	exchangeOption := msgqueue.Exchange(
		&msgqueue.ExchangeOptions{
			Name:           "bcs-data-manager",
			Durable:        true,
			PrefetchCount:  30,
			PrefetchGlobal: true,
		})
	natStreamingOption := msgqueue.NatsOpts(
		&msgqueue.NatsOptions{
			ClusterID:      "",
			ConnectTimeout: time.Duration(300) * time.Second,
			ConnectRetry:   true,
		})
	publishOption := msgqueue.PublishOpts(
		&msgqueue.PublishOptions{
			TopicName:    common.DataJobQueue,
			DeliveryMode: uint8(2),
		})
	arguments := make(map[string]interface{})
	queueArgumentsRaw := "x-message-ttl:120000"
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
			QueueArguments:    arguments,
		})
	msgQueue, err := msgqueue.NewMsgQueue(commonOption, exchangeOption, natStreamingOption, publishOption, subscribeOption)
	t.Log(err)
	t.Log(msgQueue)
	assert.Nil(t, err)
	assert.NotNil(t, msgQueue)
	t.Log("init queue successfully, sub queue[dataJob]")

}

func Test_initClusterManager(t *testing.T) {
	ctx := context.Background()
	server := &Server{
		opt: &DataManagerOptions{
			BcsAPIConf: BcsAPIConfig{
				GrpcGWAddress: "",
			},
		},
		ctx: ctx,
	}
	cmCli, err := server.initClusterManager()
	assert.Equal(t, nil, err)
	assert.NotNil(t, cmCli)

}
