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
 *
 */

// Package bcscustom xxx
package bcscustom

import (
	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/etcd"
	"github.com/coredns/coredns/plugin/etcd/msg"
	"github.com/coredns/coredns/request"

	"github.com/coredns/coredns/plugin/pkg/upstream"
	etcdcv3 "github.com/coreos/etcd/clientv3"
	"github.com/miekg/dns"
	"golang.org/x/net/context"
)

// TLS tls info for plugin
type TLS struct {
	CaFile   string
	KeyFile  string
	CertFile string
}

// BcsCustom represent a bcscustom server class
// Etcd is a plugin talks to an etcd cluster.
// BcsCustom is a plugin talks to an etcd cluster.
type BcsCustom struct {
	Next        plugin.Handler
	FallThrough bool
	Zones       []string
	RootPrefix  string
	Upstream    upstream.Upstream // Proxy for looking up names during the resolution process
	EtcdCli     *etcdcv3.Client
	EtcdPlugin  *etcd.Etcd
	Ctx         context.Context
	Listen      string
	SvrTLS      TLS

	endpoints []string // Stored here as well, to aid in testing.
}

// Services implements the coredns.plugin.ServiceBackend interface.
func (bc *BcsCustom) Services(state request.Request, exact bool, opt plugin.Options) (services []msg.Service,
	err error) {
	return bc.EtcdPlugin.Services(state, exact, opt)
}

// Reverse implements the coredns.plugin.ServiceBackend interface.
// Reverse communicates with the backend to retrieve service definition based on a IP address
// instead of a name. I.e. a reverse DNS lookup.
func (bc *BcsCustom) Reverse(state request.Request, exact bool, opt plugin.Options) ([]msg.Service, error) {
	return bc.EtcdPlugin.Reverse(state, exact, opt)
}

// Lookup implements the coredns.plugin.ServiceBackend interface.
// Lookup is used to find records else where.
func (bc *BcsCustom) Lookup(state request.Request, name string, typ uint16) (*dns.Msg, error) {
	return bc.EtcdPlugin.Lookup(state, name, typ)
}

// Records implements the coredns.plugin.ServiceBackend interface.
// Records _all_ services that matches a certain name.
// Note: it does not implement a specific service.
func (bc *BcsCustom) Records(state request.Request, exact bool) ([]msg.Service, error) {
	return bc.EtcdPlugin.Records(state, exact)
}

// IsNameError implements the coredns.plugin.ServiceBackend interface.
// IsNameError return true if err indicated a record not found condition
func (bc *BcsCustom) IsNameError(err error) bool {
	return bc.EtcdPlugin.IsNameError(err)
}

// Serial implements the coredns.plugin.ServiceBackend interface.
// Serial returns a SOA serial number to construct a SOA record.
func (bc *BcsCustom) Serial(state request.Request) uint32 {
	return bc.EtcdPlugin.Serial(state)
}

// MinTTL implements the coredns.plugin.ServiceBackend interface.
// MinTTL returns the minimum TTL to be used in the SOA record.
func (bc *BcsCustom) MinTTL(state request.Request) uint32 {
	return bc.EtcdPlugin.MinTTL(state)
}

// Transfer implements the coredns.plugin.ServiceBackend interface.
// Transfer handles a zone transfer it writes to the client just
// like any other handler.
func (bc *BcsCustom) Transfer(ctx context.Context, state request.Request) (int, error) {
	return bc.EtcdPlugin.Transfer(ctx, state)
}
