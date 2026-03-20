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

// Package mq defines message queue related interfaces, message structures, and implementations.
package mq

import (
	amqp "github.com/rabbitmq/amqp091-go"
)

// MQ is an interface that defines the methods for a message queue client.
type MQ interface {
	// Connect connects to the message queue.
	Connect() error

	// Close closes the connection to the message queue.
	Close() error

	// GetChannel gets a channel from the message queue.
	GetChannel() (*amqp.Channel, error)

	// EnsureExchange ensures the exchange exists in the message queue.
	EnsureExchange(*amqp.Channel) error

	// DeclareQueue declares a queue in the message queue.
	DeclareQueue(*amqp.Channel, string, amqp.Table) error

	// BindQueue binds a queue to an exchange in the message queue.
	BindQueue(*amqp.Channel, string, string, amqp.Table) error

	// Publish publishes a message to the message queue.
	Publish(routingKey string, body []byte) error

	// StartConsumer starts a consumer for the message queue.
	StartConsumer(*amqp.Channel, string, string, Handler, <-chan bool) error
}

// Handler is an interface for handling messages from the message queue.
type Handler interface {
	// HandleMsg processes messages from a delivery channel.
	HandleMsg(messages <-chan amqp.Delivery, done <-chan bool) error
}
