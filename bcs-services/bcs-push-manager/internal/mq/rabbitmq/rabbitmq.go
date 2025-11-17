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

// Package rabbitmq provides a RabbitMQ client implementation.
package rabbitmq

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/encrypt"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/mq"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/options"
)

// RabbitMQ is a struct that holds the RabbitMQ configuration and connection.
type RabbitMQ struct {
	config     *options.RabbitMQOption
	connection *amqp.Connection
	lock       sync.Mutex
}

// NewRabbitMQ creates a new RabbitMQ instance.
func NewRabbitMQ(config *options.RabbitMQOption) *RabbitMQ {
	return &RabbitMQ{
		config: config,
	}
}

// Connect connects to RabbitMQ.
func (r *RabbitMQ) Connect() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	// Check if connection is already available
	if r.connection == nil || r.connection.IsClosed() {
		realPwd, err := encrypt.DesDecryptFromBase([]byte(r.config.Password))
		if err != nil {
			blog.Errorf("failed to decrypt RabbitMQ password: %s", err.Error())
			return err
		}
		decryptedPassword := string(realPwd)

		// Try connecting
		con, err := amqp.DialConfig(fmt.Sprintf(
			"amqp://%s:%s@%s:%d/%s",
			r.config.Username,
			decryptedPassword,
			r.config.Host,
			r.config.Port,
			r.config.Vhost,
		), amqp.Config{})
		if err != nil {
			blog.Errorf("failed to connect to rabbitmq: %s", err.Error())
			return err
		}
		r.connection = con
	}

	return nil
}

// GetChannel gets a channel from RabbitMQ.
func (r *RabbitMQ) GetChannel() (*amqp.Channel, error) {
	if err := r.Connect(); err != nil {
		return nil, err
	}
	return r.connection.Channel()
}

// EnsureExchange ensures the exchange exists in RabbitMQ.
func (r *RabbitMQ) EnsureExchange(chn *amqp.Channel) error {
	exchangeName := fmt.Sprintf("%s.topic", r.config.SourceExchange)
	err := chn.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		blog.Errorf("failed to ensure exchange %s.topic: %s", r.config.SourceExchange, err.Error())
		return err
	}
	return nil
}

// DeclareQueue declares a queue in RabbitMQ.
func (r *RabbitMQ) DeclareQueue(chn *amqp.Channel, queueName string, args amqp.Table) error {
	if err := chn.Qos(10, 0, false); err != nil {
		return err
	}
	_, err := chn.QueueDeclare(
		queueName,
		true,
		true,
		false,
		false,
		args,
	)
	if err != nil {
		fmt.Printf("Error creating queue with name: %s, err: %s", queueName, err.Error())
		return err
	}
	return nil
}

// BindQueue binds a queue to an exchange in RabbitMQ.
func (r *RabbitMQ) BindQueue(chn *amqp.Channel, queueName, exchangeName string, args amqp.Table) error {
	err := chn.QueueBind(
		queueName,
		constant.MQRoutingKeyBindPattern,
		exchangeName,
		false,
		nil,
	)
	if err != nil {
		fmt.Printf("Error binding queue with name: %s, err: %s", queueName, err.Error())
		return err
	}
	return nil
}

// StartConsumer starts a consumer for a RabbitMQ queue.
func (r *RabbitMQ) StartConsumer(
	chn *amqp.Channel, consumer, queueName string, handler mq.Handler, done <-chan bool) error {
	messages, err := chn.Consume(
		queueName,
		consumer,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		blog.Errorf("unable to start consumer for webhook queue: %s, err: %s", queueName, err.Error())
		return err
	}

	err = handler.HandleMsg(messages, done)
	if err != nil {
		blog.Warnf("HandleMsg returned error: %s", err.Error())
		return err
	}

	return nil
}

// Publish publishes a message to the RabbitMQ exchange.
func (r *RabbitMQ) Publish(routingKey string, body []byte) error {
	chn, err := r.GetChannel()
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}
	defer chn.Close()

	if r.config.SourceExchange == "" {
		blog.Warnf("SourceExchange is empty, using default exchange name")
		r.config.SourceExchange = constant.DefaultExchangeName
	}
	exchangeName := fmt.Sprintf("%s.topic", r.config.SourceExchange)

	// Publish the message to the exchange with the specified routing key.
	err = chn.PublishWithContext(
		context.Background(),
		exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)

	if err != nil {
		blog.Errorf("failed to publish message to exchange: %s, routingKey: %s, error: %s", exchangeName, routingKey, err.Error())
		return err
	}

	return nil
}

// Close closes the RabbitMQ connection.
func (r *RabbitMQ) Close() error {
	return r.connection.Close()
}
