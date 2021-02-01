/*
Tencent is pleased to support the open source community by making Blueking Container Service available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package mq

import (
	"crypto/tls"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// Subscriber is base mq subscribe tools.
type Subscriber struct {
	nc    *nats.Conn
	addrs []string
	tls   *tls.Config
	sub   *nats.Subscription
}

// NewSubscriber creates a new Subscriber.
func NewSubscriber(addrs []string, tls *tls.Config) *Subscriber {
	return &Subscriber{addrs: addrs, tls: tls}
}

// Init initializes a new Subscriber.
func (s *Subscriber) Init(timeout, reconWait time.Duration, maxRecons int) error {
	opts := []nats.Option{}

	opts = append(opts, nats.Timeout(timeout))
	opts = append(opts, nats.ReconnectWait(reconWait))
	opts = append(opts, nats.MaxReconnects(maxRecons))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		// do nothing
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		// do nothing
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		// do nothing
	}))

	// TLS option.
	if s.tls != nil {
		opts = append(opts, nats.Secure(s.tls))
	}

	// connect nats
	nc, err := nats.Connect(strings.Join(s.addrs, ","), opts...)
	if err != nil {
		return err
	}
	s.nc = nc

	return nil
}

// Subscribe subscribes message data of target topic.
func (s *Subscriber) Subscribe(topic string, cb func([]byte)) error {
	sub, err := s.nc.Subscribe(topic, func(m *nats.Msg) {
		cb(m.Data)
	})
	if err != nil {
		return err
	}
	s.sub = sub

	return nil
}

// UnSubscribe stops subscribing message data of target topic.
func (s *Subscriber) UnSubscribe() {
	s.sub.Unsubscribe()
	s.sub.Drain()
}

// Close closes the Subscriber.
func (s *Subscriber) Close() {
	s.nc.Drain()
}
