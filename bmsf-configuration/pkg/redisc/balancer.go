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
	"errors"
	"sync/atomic"

	"gopkg.in/redis.v5"
)

// Balancer is redis balancer.
type Balancer struct {
	// pool is redis backend pool.
	pool pool

	// index is used for pick backend in random mode.
	index int32
}

// NewBalancer creates a new balancer.
func NewBalancer(opts []Options) (*Balancer, error) {
	if len(opts) == 0 {
		return nil, errors.New("invalid options")
	}

	// create balancer base on given options.
	balancer := &Balancer{pool: make(pool, len(opts))}

	for i := 0; i < len(opts); i++ {
		balancer.pool[i] = newBackend(&opts[i])
	}

	return balancer, nil
}

// Next returns the next client in redis backend.
func (b *Balancer) Next() *redis.Client {
	return b.pick().cli
}

// Close closes all connecitons here.
func (b *Balancer) Close() error {
	var cErr error

	for _, backend := range b.pool {
		if err := backend.close(); err != nil {
			cErr = err
		}
	}

	return cErr
}

// pick returns a redis backend instance.
func (b *Balancer) pick() *backend {
	backend := b.pool.up().at(int(atomic.AddInt32(&b.index, 1)))

	if backend == nil {
		backend = b.pool.rand()
	}

	return backend
}
