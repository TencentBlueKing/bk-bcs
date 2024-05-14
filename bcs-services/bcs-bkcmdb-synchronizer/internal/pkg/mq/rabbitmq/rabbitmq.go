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

// Package rabbitmq define methods for rabbitmq
package rabbitmq

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/handler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/option"
)

// RabbitMQ rabbitmq instance
type RabbitMQ struct {
	config     *option.RabbitMQConfig
	connection *amqp.Connection
	lock       sync.Mutex
}

// NewRabbitMQ create a new rabbitmq instance
func NewRabbitMQ(config *option.RabbitMQConfig) *RabbitMQ {
	return &RabbitMQ{
		config: config,
	}
}

// Connect connect to rabbitmq
func (r *RabbitMQ) Connect() error {
	r.lock.Lock()
	defer r.lock.Unlock()
	// Check if connection is already available
	if r.connection == nil || r.connection.IsClosed() {
		// Try connecting
		con, err := amqp.DialConfig(fmt.Sprintf(
			"amqp://%s:%s@%s:%d/%s",
			r.config.Username,
			r.config.Password,
			r.config.Host,
			r.config.Port,
			r.config.Vhost,
		), amqp.Config{})
		if err != nil {
			blog.Errorf("connect to rabbitmq failed, err %s", err.Error())
			return err
		}
		r.connection = con
	}

	return nil
}

// GetChannel get a channel from rabbitmq
func (r *RabbitMQ) GetChannel() (*amqp.Channel, error) {
	if err := r.Connect(); err != nil {
		return nil, err
	}
	return r.connection.Channel()
}

// EnsureExchange ensure exchange exists
func (r *RabbitMQ) EnsureExchange(chn *amqp.Channel) error {
	exchangeName := fmt.Sprintf("%s.headers", r.config.SourceExchange)
	err := chn.ExchangeDeclare(
		exchangeName,
		"headers",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		blog.Errorf("ensure exchange %s.headers failed, err %s", r.config.SourceExchange, err.Error())
		return err
	}

	_ = chn.ExchangeBind(
		exchangeName,
		"*",
		r.config.SourceExchange,
		false,
		nil,
	)

	return nil
}

// DeclareQueue declare a queue
func (r *RabbitMQ) DeclareQueue(chn *amqp.Channel, queueName string, args amqp.Table) error {
	// Make sure we process 1 message at a time
	if err := chn.Qos(1, 0, false); err != nil {
		return err
	}
	_, err := chn.QueueDeclare(
		queueName,
		true,
		false,
		true,
		false,
		args,
		// amqp.Table{"x-expires": 60000},
	)
	if err != nil {
		fmt.Printf("Error creating queue with name: %s, err: %s", queueName, err.Error())
		return err
	}
	return nil
}

// BindQueue bind queue to exchange
func (r *RabbitMQ) BindQueue(chn *amqp.Channel, queueName, exchangeName string, args amqp.Table) error {
	err := chn.QueueBind(
		queueName,
		"#",
		exchangeName,
		false,
		args,
	)
	if err != nil {
		fmt.Printf("Error binding queue with name: %s, err: %s", queueName, err.Error())
		return err
	}
	return nil
}

// StartConsumer start a consumer
func (r *RabbitMQ) StartConsumer(chn *amqp.Channel, queueName string, handler handler.Handler, done <-chan bool) error {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	consumer := fmt.Sprintf("%s.%s", hostname, queueName)

	messages, err := chn.Consume(
		queueName,
		consumer,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		blog.Errorf("Unable to start consumer for webhook queue: %s, err: %s", queueName, err.Error())
		return err
	}

	handler.HandleMsg(chn, queueName, messages, done)

	return nil
}

// PublishMsg is a function that publishes a message to the RabbitMQ exchange.
func (r *RabbitMQ) PublishMsg(chn *amqp.Channel, msg amqp.Delivery) error {
	// Set the exchange name with the source exchange name from the configuration.
	exchangeName := fmt.Sprintf("%s.headers", r.config.SourceExchange)

	// Publish the message to the exchange with the specified routing key.
	err := chn.PublishWithContext(
		context.Background(),
		exchangeName,
		msg.RoutingKey,
		false,
		false,
		amqp.Publishing{
			Headers: msg.Headers,
			Body:    msg.Body,
		},
	)

	// If there is an error publishing the message, log the error.
	if err != nil {
		blog.Errorf("Error publishing message: %s", err.Error())
	}

	// Return the error if there is one, or nil if the message was published successfully.
	return err
}

// Close is a function that closes the RabbitMQ connection.
func (r *RabbitMQ) Close() error {
	// Close the RabbitMQ connection.
	return r.connection.Close()
}
