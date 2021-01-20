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

package grpclb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3"
	etcdnaming "github.com/coreos/etcd/clientv3/naming"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/naming"
	"google.golang.org/grpc/status"
)

// Context is naming context.
type Context struct {
	// Target is target service name
	Target string

	// Port is target service port(Optional)
	Port uint16

	// EtcdConfig is config for etcd client.
	EtcdConfig clientv3.Config
}

// NewGRPCConn creates a gRPC client with naming balancer.
func NewGRPCConn(ctx *Context, opts ...grpc.DialOption) (*GRPCConn, error) {
	// etcd client
	etcdCli, err := clientv3.New(ctx.EtcdConfig)
	if err != nil {
		return nil, err
	}

	// grpc client
	grpcOpts := []grpc.DialOption{
		// You can build your own Balancer struct, add and use it here.
		grpc.WithBalancer(newRRBalance(ctx, etcdCli)),
	}
	grpcOpts = append(grpcOpts, opts...)

	grpcConn, err := grpc.Dial(fmt.Sprintf("%s/%s", DEFAULTSCHEMA, ctx.Target), grpcOpts...)
	if err != nil {
		return nil, err
	}

	return &GRPCConn{grpcConn: grpcConn, etcdCli: etcdCli}, nil
}

// GRPCConn is grpclb connection.
type GRPCConn struct {
	grpcConn *grpc.ClientConn
	etcdCli  *clientv3.Client
}

// Conn returns grpc connection.
func (c *GRPCConn) Conn() *grpc.ClientConn {
	return c.grpcConn
}

// Close closes grpc connection and etcd client.
func (c *GRPCConn) Close() {
	c.grpcConn.Close()
	c.etcdCli.Close()
}

// newRRBalance creates a round robin balancer base on etcd.
func newRRBalance(ctx *Context, etcdCli *clientv3.Client) grpc.Balancer {
	resolver := &Resolver{ctx: ctx, Client: etcdCli}
	return grpc.RoundRobin(resolver)
}

// Resolver is resolver base on etcd.
type Resolver struct {
	ctx    *Context
	Client *clientv3.Client
}

// Resolve resolve by etcd.
func (r *Resolver) Resolve(target string) (naming.Watcher, error) {
	ctx, cancel := context.WithCancel(context.Background())
	w := &gRPCWatcher{c: r.Client, target: target, ctx: ctx, lbCtx: r.ctx, cancel: cancel}
	return w, nil
}

// gRPCWatcher is etcd service updates watcher.
type gRPCWatcher struct {
	ctx    context.Context
	cancel context.CancelFunc
	lbCtx  *Context

	target string
	c      *clientv3.Client

	wch clientv3.WatchChan
	err error
}

// Next gets the next set of updates from the etcd resolver.
func (watcher *gRPCWatcher) Next() ([]*naming.Update, error) {
	if watcher.wch == nil {
		return watcher.first()
	}
	if watcher.err != nil {
		return nil, watcher.err
	}

	wr, ok := <-watcher.wch
	if !ok {
		watcher.err = status.Error(codes.Unavailable, etcdnaming.ErrWatcherClosed.Error())
		return nil, watcher.err
	}
	if watcher.err = wr.Err(); watcher.err != nil {
		return nil, watcher.err
	}

	// updates.
	updates := make([]*naming.Update, 0, len(wr.Events))

	for _, e := range wr.Events {
		var jupdate naming.Update
		var err error

		switch e.Type {
		case clientv3.EventTypePut:
			err = json.Unmarshal(e.Kv.Value, &jupdate)
			jupdate.Op = naming.Add
			jupdate.Addr = watcher.targetAddr(jupdate.Addr)

		case clientv3.EventTypeDelete:
			err = json.Unmarshal(e.PrevKv.Value, &jupdate)
			jupdate.Op = naming.Delete
			jupdate.Addr = watcher.targetAddr(jupdate.Addr)

		default:
			continue
		}

		if err == nil {
			updates = append(updates, &jupdate)
		}
	}

	return updates, nil
}

// targetAddr make a right address by optional port.
func (watcher *gRPCWatcher) targetAddr(raddr string) string {
	if watcher.lbCtx.Port == 0 {
		return raddr
	}

	ipPort := strings.Split(raddr, ":")
	if len(ipPort) != 2 {
		return raddr
	}
	return fmt.Sprintf("%s:%d", ipPort[0], watcher.lbCtx.Port)
}

func (watcher *gRPCWatcher) first() ([]*naming.Update, error) {
	resp, err := watcher.c.Get(watcher.ctx, watcher.target, clientv3.WithPrefix(), clientv3.WithSerializable())
	if watcher.err = err; err != nil {
		return nil, err
	}

	// updates.
	updates := make([]*naming.Update, 0, len(resp.Kvs))

	for _, kv := range resp.Kvs {
		var jupdate naming.Update
		if err := json.Unmarshal(kv.Value, &jupdate); err != nil {
			continue
		}

		jupdate.Addr = watcher.targetAddr(jupdate.Addr)
		updates = append(updates, &jupdate)
	}

	opts := []clientv3.OpOption{
		clientv3.WithRev(resp.Header.Revision + 1),
		clientv3.WithPrefix(),
		clientv3.WithPrevKV(),
	}
	watcher.wch = watcher.c.Watch(watcher.ctx, watcher.target, opts...)

	return updates, nil
}

// Close closes watcher and stop resolver.
func (watcher *gRPCWatcher) Close() {
	watcher.cancel()
}
