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

package serviced

import (
	"context"
	"encoding/json"
	"sync"

	"go.etcd.io/etcd/api/v3/mvccpb"
	etcd3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/resolver"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// etcdBuilder creates a resolver that will be used to watch name resolution updates.
type etcdBuilder struct {
	cli *etcd3.Client
}

// newEtcdBuilder new etcdBuilder.
func newEtcdBuilder(cli *etcd3.Client) *etcdBuilder {
	return &etcdBuilder{cli: cli}
}

// Build creates and starts a etcd resolver that watches the name resolution of the target.
func (b *etcdBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (
	resolver.Resolver, error) {

	ctx, cancel := context.WithCancel(context.Background())
	r := &etcdResolver{
		cli:    b.cli,
		cc:     cc,
		target: target.Endpoint(),
		ctx:    ctx,
		cancel: cancel,
	}

	go r.watcher()
	return r, nil
}

// Scheme return grpc scheme.
func (b *etcdBuilder) Scheme() string {
	return "etcd"
}

// etcdResolver watches for the updates on the specified target.
// Updates include address updates and service config updates.
type etcdResolver struct {
	sync.RWMutex
	cli       *etcd3.Client
	cc        resolver.ClientConn
	target    string
	addresses map[string]resolver.Address
	ctx       context.Context
	cancel    context.CancelFunc
}

// ResolveNow will be called by gRPC to try to resolve the target name
// again. It's just a hint, resolver can ignore this if it's not necessary.
//
// It could be called multiple times concurrently.
func (r *etcdResolver) ResolveNow(resolver.ResolveNowOptions) {}

// Close closes the resolver.
func (r *etcdResolver) Close() {
	r.cancel()
}

func (r *etcdResolver) watcher() {
	// Use serialized request so resolution still works if the target etcd
	// server is partitioned away from the quorum.
	resp, err := r.cli.Get(r.ctx, r.target, etcd3.WithPrefix(), etcd3.WithSerializable())
	if err != nil {
		logs.Infof("get %s key failed, err: %v", r.target, err)
	}
	logs.V(3).Infof("watcher etcd resolver get target: %s, kvs: %#v", r.target, resp.Kvs)

	r.addresses = make(map[string]resolver.Address)
	if err == nil {
		for _, kv := range resp.Kvs {
			r.setAddress(string(kv.Key), string(kv.Value))
		}

		if e := r.cc.UpdateState(resolver.State{Addresses: r.getAddresses()}); e != nil {
			logs.Errorf("client conn update state failed, addr: %v, err: %v", r.addresses, e)
		}
	}
	logs.V(3).Infof("watcher addresses:%#v", r.addresses)

	opts := []etcd3.OpOption{
		etcd3.WithRev(resp.Header.Revision + 1),
		etcd3.WithPrefix(),
		etcd3.WithPrevKV(),
	}
	watch := r.cli.Watch(r.ctx, r.target, opts...)

	for response := range watch {
		for _, event := range response.Events {
			switch event.Type {
			case mvccpb.PUT:
				r.setAddress(string(event.Kv.Key), string(event.Kv.Value))
			case mvccpb.DELETE:
				r.delAddress(string(event.Kv.Key))
			default:
				logs.Infof("unknown event type, %d", event.Type)
				continue
			}
		}

		addresses := r.getAddresses()
		if err := r.cc.UpdateState(resolver.State{Addresses: addresses}); err != nil {
			logs.Errorf("client conn update state failed, addr: %v, err: %v", addresses, err)
		}
	}
}

// setAddress set etcdResolver addresses.
func (r *etcdResolver) setAddress(key, address string) {
	r.Lock()
	defer r.Unlock()

	addr := resolver.Address{}
	if err := json.Unmarshal([]byte(address), &addr); err != nil {
		logs.Errorf("etcd resolver unmarshal addr failed, addr: %s, err: %v", address, err)
		return
	}

	r.addresses[key] = addr
	logs.V(3).Infof("set address key:%s, address:%s, addresses:%#v", key, address, r.addresses)
}

// delAddress del etcdResolver addresses.
func (r *etcdResolver) delAddress(key string) {
	r.Lock()
	defer r.Unlock()
	delete(r.addresses, key)
	logs.V(3).Infof("del address key:%s, addresses:%#v", key, r.addresses)
}

// getAddresses get grpc addresses.
func (r *etcdResolver) getAddresses() []resolver.Address {
	addresses := make([]resolver.Address, 0, len(r.addresses))

	for _, address := range r.addresses {
		addresses = append(addresses, address)
	}

	return addresses
}
