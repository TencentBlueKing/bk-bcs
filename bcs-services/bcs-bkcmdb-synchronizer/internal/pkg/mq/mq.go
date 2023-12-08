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

// Package mq define mq interface
package mq

import (
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/handler"
)

// MQ is an interface that defines the methods for a message queue.
type MQ interface {
	// Connect is a method that connects to the message queue.
	Connect() error

	// Close is a method that closes the connection to the message queue.
	Close() error

	// GetChannel is a method that gets a channel from the message queue.
	GetChannel() (*amqp.Channel, error)

	// EnsureExchange is a method that ensures the exchange exists in the message queue.
	EnsureExchange(*amqp.Channel) error

	// DeclareQueue is a method that declares a queue in the message queue.
	DeclareQueue(*amqp.Channel, string, amqp.Table) error

	// BindQueue is a method that binds a queue to an exchange in the message queue.
	BindQueue(*amqp.Channel, string, string, amqp.Table) error

	// StartConsumer is a method that starts a consumer for the message queue.
	StartConsumer(*amqp.Channel, string, handler.Handler) error
}
