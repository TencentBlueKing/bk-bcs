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

package consumer

import (
	"testing"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/cmd/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/handler/eventhandler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/remote/alert"
)

func connectQueue() (msgqueue.MessageQueue, error) {
	var (
		resourceToQueue = map[string]string{
			"Pod":   "Pod",
			"Event": "Event",
		}
		queueOptions []msgqueue.QueueOption
	)

	commonOpts := msgqueue.CommonOpts(
		&msgqueue.CommonOptions{
			QueueFlag:       true,
			QueueKind:       "rabbitmq",
			ResourceToQueue: resourceToQueue,
			Address:         "amqp://root:123456@configure.test.hosts:5672", /*"amqp://rabbit:rabbit@127.0.0.1:5672"*/
		})

	exchangeOpts := msgqueue.Exchange(
		&msgqueue.ExchangeOptions{
			Name:           "bcs-storage", //"bcs-storage-test",
			Durable:        true,
			PrefetchCount:  0,
			PrefetchGlobal: false,
		})

	publishOpts := msgqueue.PublishOpts(
		&msgqueue.PublishOptions{
			TopicName:    "Pod",
			DeliveryMode: 2,
		})

	subscribeOpts := msgqueue.SubscribeOpts(&msgqueue.SubscribeOptions{
		TopicName:      "Pod",
		QueueName:      "Pod",
		DisableAutoAck: true,
		Durable:        true,
		AckOnSuccess:   true,
		RequeueOnError: true,
		QueueArguments: map[string]interface{}{
			"x-message-ttl": 60000,
		},
	})
	queueOptions = append(queueOptions, commonOpts, exchangeOpts, publishOpts, subscribeOpts)

	queue, err := msgqueue.NewMsgQueue(queueOptions...)
	if err != nil {
		return nil, err
	}

	return queue, err
}

func alertClient() alert.BcsAlarmInterface {
	defaultOptions := &config.AlertServerOptions{
		Server:      "http://xxx.xx.xx.xx.xx",
		AppCode:     "xxx",
		AppSecret:   "xxx",
		ServerDebug: true,
	}

	return alert.NewAlertServer(defaultOptions, alert.WithTestDebug(true))
}

func getEventHandler() *eventhandler.SyncEventHandler {
	eventHandler := eventhandler.NewSyncEventHandler(eventhandler.Options{
		AlertEventBatchNum: 100,
		ConcurrencyNum:     100,
		Client:             alertClient(),
	})

	return eventHandler
}

func TestConsumers_Run(t *testing.T) {

	consumers := []Consumer{getEventHandler()}
	queue, err := connectQueue()
	if err != nil {
		t.Fatal(err)
	}

	consumer := NewConsumers(consumers, queue)
	consumer.Run()

	select {}
}
