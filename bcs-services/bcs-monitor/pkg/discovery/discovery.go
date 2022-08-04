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

package discovery

import (
	"context"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	etcd "github.com/asim/go-micro/plugins/registry/etcd/v4"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

const serverNameSuffix = ".bkbcs.tencent.com"

// serviceDiscovery
type serviceDiscovery struct {
	ctx context.Context
	srv micro.Service
}

// NewServiceDiscovery :
func NewServiceDiscovery(ctx context.Context, name, version, bindaddr, advertiseAddr string) (*serviceDiscovery, error) {
	svr := server.NewServer(
		server.Name(name+serverNameSuffix),
		server.Version(version),
		server.Context(ctx),
	)

	if advertiseAddr != "" {
		svr.Init(server.Advertise(advertiseAddr))
	} else {
		svr.Init(server.Advertise(bindaddr))
	}

	service := micro.NewService(micro.Server(svr))

	sd := &serviceDiscovery{srv: service, ctx: ctx}
	if err := sd.init(); err != nil {
		return nil, err
	}

	return sd, nil
}

// Run
func (s *serviceDiscovery) Run() error {
	return s.srv.Run()
}

func (s *serviceDiscovery) init() error {
	// etcd 服务发现注册
	etcdRegistry, err := s.initEtcdRegistry()
	if err != nil {
		return err
	}

	if etcdRegistry != nil {
		s.srv.Init(micro.Registry(etcdRegistry))
	}
	return nil
}

// initEtcdRegistry etcd 服务注册
func (s *serviceDiscovery) initEtcdRegistry() (registry.Registry, error) {
	endpoints := config.G.Viper.GetString("etcd.endpoints")
	if endpoints == "" {
		return nil, nil
	}

	etcdRegistry := etcd.NewRegistry(registry.Addrs(strings.Split(endpoints, ",")...))

	ca := config.G.Viper.GetString("etcd.ca")
	cert := config.G.Viper.GetString("etcd.cert")
	key := config.G.Viper.GetString("etcd.key")
	if ca != "" && cert != "" && key != "" {
		tlsConfig, err := ssl.ClientTslConfVerity(ca, cert, key, "")
		if err != nil {
			return nil, err
		}
		etcdRegistry.Init(registry.TLSConfig(tlsConfig))
	}

	return etcdRegistry, nil
}
