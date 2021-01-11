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

// Publisher is base mq publish tools.
type Publisher struct {
	nc    *nats.Conn
	addrs []string
	tls   *tls.Config
}

// NewPublisher creates a new Publisher.
func NewPublisher(addrs []string, tls *tls.Config) *Publisher {
	return &Publisher{addrs: addrs, tls: tls}
}

// Init initializes a new Publisher.
func (p *Publisher) Init(timeout, reconWait time.Duration, maxRecons int) error {
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
	if p.tls != nil {
		opts = append(opts, nats.Secure(p.tls))
	}

	// connect nats
	nc, err := nats.Connect(strings.Join(p.addrs, ","), opts...)
	if err != nil {
		return err
	}
	p.nc = nc

	return nil
}

// Publish publishs message data with target topic.
func (p *Publisher) Publish(topic string, bytes []byte) error {
	if err := p.nc.Publish(topic, bytes); err != nil {
		return err
	}
	return p.nc.Flush()
}

// Close closes the Publisher.
func (p *Publisher) Close() error {
	return p.nc.Drain()
}
