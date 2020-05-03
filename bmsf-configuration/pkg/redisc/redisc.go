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

package redisc

import (
	"time"

	"gopkg.in/redis.v5"
)

var (
	// NoExist is a alias of redis.Nil returned when the redis key non-exist.
	NoExist = redis.Nil

	// pong is redis ping ack message.
	pong = "PONG"

	// defaultPingInterval is default interval of redis PING checking.
	defaultPingInterval = 100 * time.Millisecond
)

// RedisCli is redis client with balancer.
type RedisCli struct {
	// redis balancer.
	balancer *Balancer

	// target redis services addresses.
	addrs []string
}

// NewRedisCli creates a new RedisCli instance.
func NewRedisCli(addrs []string, pingInterval time.Duration, opts redis.Options) (*RedisCli, error) {
	options := []Options{}

	for i := 0; i < len(addrs); i++ {
		opts.Addr = addrs[i]

		opt := Options{
			Options:  opts,
			interval: pingInterval,
		}
		options = append(options, opt)
	}

	// create a balancer base on options.
	balancer, err := NewBalancer(options)
	if err != nil {
		return nil, err
	}

	// create new RedisCli success.
	cli := &RedisCli{
		balancer: balancer,
		addrs:    addrs,
	}

	return cli, nil
}

// Cli returns a redis client.
func (c *RedisCli) Cli() *redis.Client {
	return c.balancer.Next()
}

// Close closes the redis client handled in the balancer.
func (c *RedisCli) Close() {
	c.balancer.Close()
}
