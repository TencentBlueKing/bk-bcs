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

// Package msgqueuev4 xxx
package msgqueuev4

import (
	"time"

	"github.com/go-micro/plugins/v4/broker/rabbitmq"
	"github.com/go-micro/plugins/v4/broker/stan"
	"github.com/pkg/errors"
	"go-micro.dev/v4/broker"
)

// rabbitmq broker init: brokerOptions/init/connect
func rabbitmqBroker(q *QueueOptions) (broker.Broker, error) {
	var brokerOpts []broker.Option
	brokerOpts = append(brokerOpts, broker.Addrs(q.CommonOptions.Address), rabbitmq.ExchangeName(q.Exchange.Name),
		rabbitmq.PrefetchCount(q.Exchange.PrefetchCount))
	// exchange durable feature
	if q.Exchange.Durable {
		brokerOpts = append(brokerOpts, rabbitmq.DurableExchange())
	}
	// prefetchGlobal feature
	if q.Exchange.PrefetchGlobal {
		brokerOpts = append(brokerOpts, rabbitmq.PrefetchGlobal())
	}

	brokerRabbit := rabbitmq.NewBroker(brokerOpts...)

	// init rabbitmq broker
	err := brokerRabbit.Init()
	if err != nil {
		return nil, errors.Wrapf(err, "brokerRabbitmq init failed")
	}

	// create connect
	if err = brokerRabbit.Connect(); err != nil {
		return nil, errors.Wrapf(err, "can't connect to rabbit broker")
	}

	return brokerRabbit, nil
}

// natstreaming broker init: natsreaming options/init/connect
func natstreamingBroker(q *QueueOptions) (broker.Broker, error) {
	var brokerOpts []broker.Option
	brokerOpts = append(brokerOpts, stan.ClusterID(q.NatsOptions.ClusterID), broker.Addrs(q.CommonOptions.Address))
	// exchange durable feature
	if q.NatsOptions.ConnectRetry {
		brokerOpts = append(brokerOpts, stan.ConnectTimeout(time.Minute*5), stan.ConnectRetry(true))
	}

	brokerNatstreaming := stan.NewBroker(brokerOpts...)

	// init natstreaming broker
	err := brokerNatstreaming.Init()
	if err != nil {
		return nil, errors.Wrapf(err, "brokerNatstreaming init failed")
	}

	// create connect
	if err = brokerNatstreaming.Connect(); err != nil {
		return nil, errors.Wrapf(err, "can't connect to natstreaming broker")
	}

	return brokerNatstreaming, nil
}

// NewQueueBroker connect queue instance by queue kind
func NewQueueBroker(options *QueueOptions) (broker.Broker, error) {

	if !options.CommonOptions.QueueFlag {
		return nil, errors.New("queue flag is off")
	}

	var (
		err error
		b   broker.Broker
	)

	switch options.CommonOptions.QueueKind {
	case RABBITMQ:
		// validate rabbitmq configOptions
		err = validateRabbitMqOptions(options)
		if err != nil {
			return nil, err
		}
		// init rabbitmq broker
		b, err = rabbitmqBroker(options)
		if err != nil {
			return nil, err
		}
	case NATSTREAMING:
		// validate nats configOptions
		err = validateNatsMqOptions(options)
		if err != nil {
			return nil, err
		}
		// init nats broker
		b, err = natstreamingBroker(options)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unSupported queue kind")
	}

	return b, nil
}
