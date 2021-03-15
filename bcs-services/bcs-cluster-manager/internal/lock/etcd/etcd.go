/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package etcd

import (
	"context"
	"errors"
	"path"
	gosync "sync"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/lock"
	client "github.com/coreos/etcd/clientv3"
	cc "github.com/coreos/etcd/clientv3/concurrency"
)

// Client for election
type Client struct {
	options lock.Options
	path    string
	client  *client.Client

	mtx   gosync.Mutex
	locks map[string]*etcdLock
}

type etcdLock struct {
	s *cc.Session
	m *cc.Mutex
}

// New create etcd locker
func New(opts ...lock.Option) (*Client, error) {
	var options lock.Options
	for _, o := range opts {
		o(&options)
	}

	var endpoints []string

	for _, addr := range options.Endpoints {
		if len(addr) > 0 {
			endpoints = append(endpoints, addr)
		}
	}

	if len(endpoints) == 0 {
		endpoints = []string{"http://127.0.0.1:2379"}
	}

	var conf client.Config
	if options.TLSConfig != nil {
		conf = client.Config{
			Endpoints: endpoints,
			TLS:       options.TLSConfig,
		}
	} else {
		conf = client.Config{
			Endpoints: endpoints,
		}
	}

	c, err := client.New(conf)
	if err != nil {
		return nil, err
	}

	return &Client{
		path:    "/lock.bkbcs.tencent.com",
		client:  c,
		options: options,
		locks:   make(map[string]*etcdLock),
	}, nil
}

// Init init etcd lock
func (c *Client) Init(opts ...lock.Option) error {
	for _, o := range opts {
		o(&c.options)
	}
	return nil
}

// Lock lock for certain id
func (c *Client) Lock(id string, opts ...lock.LockOption) error {
	var options lock.LockOptions
	for _, o := range opts {
		o(&options)
	}

	// make path
	var lpath string
	if len(c.options.Prefix) != 0 {
		lpath = path.Join(c.path, c.options.Prefix)
	}
	lpath = path.Join(lpath, id)

	var sopts []cc.SessionOption
	if options.TTL > 0 {
		sopts = append(sopts, cc.WithTTL(int(options.TTL.Seconds())))
	}

	s, err := cc.NewSession(c.client, sopts...)
	if err != nil {
		return err
	}

	m := cc.NewMutex(s, lpath)

	if err := m.Lock(context.TODO()); err != nil {
		return err
	}

	c.mtx.Lock()
	c.locks[id] = &etcdLock{
		s: s,
		m: m,
	}
	c.mtx.Unlock()
	return nil
}

// Unlock unlock for certain id
func (c *Client) Unlock(id string) error {
	c.mtx.Lock()
	defer c.mtx.Unlock()
	v, ok := c.locks[id]
	if !ok {
		return errors.New("lock not found")
	}
	err := v.m.Unlock(context.Background())
	delete(c.locks, id)
	return err
}
