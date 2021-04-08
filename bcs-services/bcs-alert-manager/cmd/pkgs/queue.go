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

package pkgs

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/msgqueue"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/cmd/config"
)

var (
	queueClineOnce sync.Once
	queueClient    msgqueue.MessageQueue
)

// GetQueueClient get messageQueue interface
func GetQueueClient(options *config.AlertManagerOptions) msgqueue.MessageQueue {
	queueClineOnce.Do(func() {
		var err error
		queueClient, err = initQueueClient(options.QueueConfig)
		if err != nil {
			blog.Errorf("parse queue config failed, err %s", err.Error())
			panic("init queueClient failed")
		}
	})

	return queueClient
}

func initQueueClient(queueConf config.QueueConfig) (msgqueue.MessageQueue, error) {
	var queueOptions []msgqueue.QueueOption

	resource := queueConf.Resource
	resourceToQueue := map[string]string{}
	arrayResource := strings.Split(resource, ",")
	for _, r := range arrayResource {
		resourceToQueue[r] = r
	}

	newAddress, err := transQueueAddressPwd(queueConf.Address)
	if err != nil {
		return nil, err
	}

	commonOption := msgqueue.CommonOpts(
		&msgqueue.CommonOptions{
			QueueFlag:       queueConf.QueueFlag,
			QueueKind:       msgqueue.QueueKind(queueConf.QueueKind),
			ResourceToQueue: resourceToQueue,
			Address:         newAddress,
		})

	exchangeOption := msgqueue.Exchange(
		&msgqueue.ExchangeOptions{
			Name:           queueConf.ExchangeName,
			Durable:        queueConf.ExchangeDurable,
			PrefetchCount:  queueConf.ExchangePrefetchCount,
			PrefetchGlobal: queueConf.ExchangePrefetchGlobal,
		})

	natStreamingOption := msgqueue.NatsOpts(
		&msgqueue.NatsOptions{
			ClusterID:      queueConf.ClusterID,
			ConnectTimeout: time.Duration(queueConf.ConnectTimeout) * time.Second,
			ConnectRetry:   queueConf.ConnectRetry,
		})

	publishOption := msgqueue.PublishOpts(
		&msgqueue.PublishOptions{
			DeliveryMode: uint8(queueConf.PublishDelivery),
		})

	arguments := parseQueueArguments(queueConf.QueueArguments)
	subscribeOption := msgqueue.SubscribeOpts(
		&msgqueue.SubscribeOptions{
			DisableAutoAck:    queueConf.SubDisableAutoAck,
			Durable:           queueConf.SubDurable,
			AckOnSuccess:      queueConf.SubAckOnSuccess,
			RequeueOnError:    queueConf.SubRequeueOnError,
			DeliverAllMessage: queueConf.SubDeliverAllMessage,
			ManualAckMode:     queueConf.SubManualAckMode,
			EnableAckWait:     queueConf.SubEnableAckWait,
			AckWaitDuration:   time.Duration(queueConf.SubAckWaitDuration) * time.Second,
			MaxInFlight:       queueConf.SubMaxInFlight,
			QueueArguments:    arguments,
		})

	queueOptions = append(queueOptions, commonOption, natStreamingOption, exchangeOption, publishOption, subscribeOption)

	queueClient, err := msgqueue.NewMsgQueue(queueOptions...)
	if err != nil {
		msgErr := fmt.Errorf("create queue failed, err %s", err.Error())
		blog.Errorf("create queue failed, err %s", err.Error())
		return nil, msgErr
	}
	queueKind, _ := queueClient.String()

	blog.Infof("init queueClient[%s] successfully", queueKind)
	return queueClient, nil
}

// https://github.com/streadway/amqp/blob/master/channel.go
// amqp channel.go: QueueDeclare limit value type: nil, bool, byte, int, int16, int32, int64, float32, float64, string, []byte, Decimal, time.Time
func parseQueueArguments(queueArguments map[string]interface{}) map[string]interface{} {
	arguments := map[string]interface{}{}

	for key, value := range queueArguments {
		if v, ok := value.(uint64); ok {
			arguments[key] = int64(v)
		} else {
			arguments[key] = value
		}
	}

	return arguments
}

func transQueueAddressPwd(address string) (string, error) {
	schemas := strings.Split(address, "//")
	if len(schemas) != 2 {
		return "", fmt.Errorf("passwd contain special char(//)")
	}

	accountServers := strings.Split(schemas[1], "@")
	if len(accountServers) != 2 {
		return "", fmt.Errorf("queue account or passwd contain special char(@)")
	}

	accounts := strings.Split(accountServers[0], ":")
	if len(accounts) != 2 {
		return "", fmt.Errorf("queue account or passwd contain special char(:)")
	}

	pwd := accounts[1]
	if pwd != "" {
		realPwd, _ := encrypt.DesDecryptFromBase([]byte(pwd))
		pwd = string(realPwd)
	}

	parseAddress := fmt.Sprintf("%s//%s:%s@%s", schemas[0], accounts[0], pwd, accountServers[1])
	return parseAddress, nil
}
