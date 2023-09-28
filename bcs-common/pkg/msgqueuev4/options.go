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

package msgqueuev4

import (
	"regexp"
	"time"

	"github.com/pkg/errors"
)

// QueueKind define queueType
type QueueKind string

const (
	// RABBITMQ queueType
	RABBITMQ QueueKind = "rabbitmq"
	// NATSTREAMING queueType
	NATSTREAMING QueueKind = "nats-streaming"
)

// DefaultQueue Options
var (
	// default queue flag
	DefaultQueueFlag = true
	// default queue kind
	DefaultQueueKind QueueKind = "rabbitmq"
	// default rabbitmq connection
	DefaultRabbitURL = "amqp://rabbit:rabbit@127.0.0.1:5672"
	// default monitor resource
	DefaultResourceToQueue = map[string]string{
		"Pod":         "Pod",
		"Event":       "Event",
		"StatefulSet": "StatefulSet",
		"DaemonSet":   "DaemonSet",
	}
	// default common options
	DefaulCommonOptions = &CommonOptions{
		QueueFlag:       DefaultQueueFlag,
		QueueKind:       DefaultQueueKind,
		ResourceToQueue: DefaultResourceToQueue,
		Address:         DefaultRabbitURL,
	}
	// default exchange options
	DefaultExchangeOptions = &ExchangeOptions{
		Name:           "micro",
		Durable:        true,
		PrefetchCount:  0,
		PrefetchGlobal: false,
	}
	// default publish options
	DefaultPublishOptions = &PublishOptions{
		DeliveryMode: 2,
	}
	// default subscribe options
	DefaultSubscribeOptions = &SubscribeOptions{
		TopicName:      "Pod",
		QueueName:      "Pod",
		Durable:        false,
		DisableAutoAck: true,
		AckOnSuccess:   true,
		RequeueOnError: false,
	}
)

// GetDefaultOptions returns default configuration options for the client.
func GetDefaultOptions() *QueueOptions {
	return &QueueOptions{
		CommonOptions:    DefaulCommonOptions,
		Exchange:         DefaultExchangeOptions,
		PublishOptions:   DefaultPublishOptions,
		SubscribeOptions: DefaultSubscribeOptions,
	}
}

// CommonOptions commonOptions for queue
type CommonOptions struct {
	// queue flag&kind
	QueueFlag bool      `json:"queueFlag"`
	QueueKind QueueKind `json:"queueKind"`
	// subscribe resourceType due to restrain queueInfo
	ResourceToQueue map[string]string `json:"resourceToQueue"`
	// queue broker Url
	Address string `json:"address"`
}

// QueueOptions init options: rabbitmq or natstreaming
type QueueOptions struct {
	CommonOptions    *CommonOptions    `json:"commonOptions"`
	Exchange         *ExchangeOptions  `json:"exchange"`
	NatsOptions      *NatsOptions      `json:"natsOptions"`
	PublishOptions   *PublishOptions   `json:"publishOptions"`
	SubscribeOptions *SubscribeOptions `json:"subscribeOptions"`
}

// NatsOptions initOptions for nats
type NatsOptions struct {
	ClusterID      string        `json:"clusterId"`
	ClientID       string        `json:"clientId"`
	ConnectTimeout time.Duration `json:"connectTimeout"`
	ConnectRetry   bool          `json:"connectRetry"`
}

// ExchangeOptions for rabbitmq exchange
type ExchangeOptions struct {
	Name    string `json:"name"`
	Durable bool   `json:"durable"`
	// subscribe option
	PrefetchCount  int  `json:"prefetchCount"`
	PrefetchGlobal bool `json:"prefetchGlobal"`
}

// PublishOptions publish init parameters
type PublishOptions struct {
	// exchange bind queue by topic
	TopicName string `json:"topicName"`
	// message durable
	DeliveryMode uint8 `json:"deliveryMode"`
}

// SubscribeOptions subscribe init parameters
type SubscribeOptions struct {
	// subscribe topic && queue
	TopicName string `json:"topicName"`
	QueueName string `json:"queueName"`

	// common options
	DisableAutoAck bool                   `json:"disableAutoAck"`
	Durable        bool                   `json:"durable"`
	QueueArguments map[string]interface{} `json:"argument"`

	// subscribe ackOptions by handler result
	AckOnSuccess   bool `json:"ackOnSuccess"`
	RequeueOnError bool `json:"requeueOnError"`

	// natstreaming subscribe options
	DeliverAllMessage bool          `json:"deliverAllMessage"`
	ManualAckMode     bool          `json:"manualAckMode"`
	EnableAckWait     bool          `json:"enableAckWait"`
	AckWaitDuration   time.Duration `json:"ackWaitDuration"`
	MaxInFlight       int           `json:"maxInFlight"`
}

// QueueOption function for init queueOptions
type QueueOption func(options *QueueOptions)

// CommonOpts set queue flag/kind
func CommonOpts(commonOpts *CommonOptions) QueueOption {
	return func(q *QueueOptions) {
		q.CommonOptions = commonOpts
	}
}

// Exchange set exchange name && durable
func Exchange(exchange *ExchangeOptions) QueueOption {
	return func(q *QueueOptions) {
		q.Exchange = exchange
	}
}

// PublishOpts set producer topic && deliveryMode
func PublishOpts(publishOpts *PublishOptions) QueueOption {
	return func(q *QueueOptions) {
		q.PublishOptions = publishOpts
	}
}

// SubscribeOpts ser consumer to sub topic/name/durable/ack handler
func SubscribeOpts(subscribeOpts *SubscribeOptions) QueueOption {
	return func(q *QueueOptions) {
		q.SubscribeOptions = subscribeOpts
	}
}

// NatsOpts connect options
func NatsOpts(natsOpts *NatsOptions) QueueOption {
	return func(q *QueueOptions) {
		q.NatsOptions = natsOpts
	}
}

// MetaData meta dataInfo
type MetaData struct {
	ClusterID    string `json:"clusterId"`
	NameSpace    string `json:"namespace"`
	ResourceType string `json:"resourceType"`
	Event        string `json:"event"`
	Name         string `json:"name"`
}

// PodInfo for pod
type PodInfo struct {
	MetaData `json:",inline"`
	Data     interface{} `json:"data"`
}

// EventInfo for event
type EventInfo struct {
	MetaData `json:",inline"`
	Data     interface{} `json:"data"`
}

// DeploymentInfo for deployment
type DeploymentInfo struct {
	MetaData `json:",inline"`
	Data     interface{} `json:"data"`
}

// StatefulSetInfo for statefulSet
type StatefulSetInfo struct {
	MetaData `json:",inline"`
	Data     interface{} `json:"data"`
}

// check options validation
func validateNatsMqOptions(n *QueueOptions) error {
	if !(len(n.CommonOptions.Address) > 0 && regexp.MustCompile("^nat(s)?://.*").MatchString(n.CommonOptions.Address)) {
		return errors.Errorf("natstreaming options address '%s' error", n.CommonOptions.Address)
	}

	if n.NatsOptions.ClusterID == "" || n.NatsOptions.ClientID == "" {
		return errors.New("nststreaming options clusterId|clientId is null")
	}

	if n.PublishOptions.TopicName == "" {
		return errors.New("natstreaming PublishOptions topic name is null")
	}

	if n.SubscribeOptions.TopicName == "" || n.SubscribeOptions.QueueName == "" {
		return errors.New("natstreaming SubscribeOptions topic|queue name is null")
	}

	return nil
}

func validateRabbitMqOptions(r *QueueOptions) error {
	if !(len(r.CommonOptions.Address) > 0 && regexp.MustCompile("^amqp(s)?://.*").MatchString(r.CommonOptions.Address)) {
		return errors.Errorf("rabbitmq address '%s' error", r.CommonOptions.Address)
	}

	if r.Exchange.Name == "" {
		return errors.New("exchange name is null, please input correct parameter")
	}

	return nil
}
