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

// Package discovery xxx
package discovery

import (
	"context"
	"crypto/tls"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	etcd "github.com/go-micro/plugins/v4/registry/etcd"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"
	"go-micro.dev/v4/util/cmd"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component/bcs/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/component/bcs/project"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
)

const serverNameSuffix = ".bkbcs.tencent.com"

// ServiceDiscovery service discovery
type ServiceDiscovery struct {
	ctx             context.Context
	srv             micro.Service
	clientTLSConfig *tls.Config
}

// NewServiceDiscovery :
func NewServiceDiscovery(ctx context.Context, name, version, bindaddr, advertiseAddr,
	addrIPv6 string) (*ServiceDiscovery, error) {
	metadata := map[string]string{}
	if addrIPv6 != "" {
		metadata[types.IPV6] = addrIPv6
	}

	svr := server.NewServer(
		server.Name(name+serverNameSuffix),
		server.Version(version),
		server.Context(ctx),
	)

	if advertiseAddr != "" {
		_ = svr.Init(server.Advertise(advertiseAddr))
	} else {
		_ = svr.Init(server.Advertise(bindaddr))
	}

	service := micro.NewService(
		micro.Server(svr),
		micro.Metadata(metadata),
		micro.Cmd(NewDummyCmd()),
		micro.RegisterTTL(time.Second*30),
		micro.RegisterInterval(time.Second*15),
	)

	sd := &ServiceDiscovery{srv: service, ctx: ctx}
	if err := sd.init(); err != nil {
		return nil, err
	}

	return sd, nil
}

// Run xxx
func (s *ServiceDiscovery) Run() error {
	return s.srv.Run()
}

func (s *ServiceDiscovery) init() error {
	// etcd 服务发现注册
	etcdRegistry, err := s.initEtcdRegistry()
	if err != nil {
		return err
	}

	if etcdRegistry != nil {
		s.srv.Init(micro.Registry(etcdRegistry))
	}
	err = s.initTLSConfig()
	if err != nil {
		return err
	}
	err = s.initComponent(etcdRegistry)
	if err != nil {
		return err
	}
	return nil
}

// initTLSConfig 初始化client TLS 配置
func (s *ServiceDiscovery) initTLSConfig() error {

	// client tls config
	if config.G.Client.Cert != "" && config.G.Client.Key != "" && config.G.Client.Ca != "" {
		tlsConfig, err := ssl.ClientTslConfVerity(
			config.G.Client.Ca, config.G.Client.Cert, config.G.Client.Key, config.G.Client.CertPwd,
		)
		if err != nil {
			blog.Error("load bcs-ui client tls config failed: %v", err)
			return err
		}
		s.clientTLSConfig = tlsConfig
		blog.Info("load bcs-ui client tls config successfully")
	}
	return nil
}

func (s *ServiceDiscovery) initComponent(microRgt registry.Registry) error {
	err := project.NewClient(s.clientTLSConfig, microRgt)
	if err != nil {
		blog.Error("init project client error, %s", err.Error())
		return err
	}
	err = clustermanager.NewClient(s.clientTLSConfig, microRgt)
	if err != nil {
		blog.Error("init clustermanager client error, %s", err.Error())
		return err
	}
	blog.Info("init all client successfully")
	return nil
}

// initEtcdRegistry etcd 服务注册
func (s *ServiceDiscovery) initEtcdRegistry() (registry.Registry, error) {
	endpoints := config.G.Etcd.Endpoints
	if endpoints == "" {
		return nil, nil
	}

	etcdRegistry := etcd.NewRegistry(registry.Addrs(strings.Split(endpoints, ",")...))

	ca := config.G.Etcd.Ca
	cert := config.G.Etcd.Cert
	key := config.G.Etcd.Key
	if ca != "" && cert != "" && key != "" {
		tlsConfig, err := ssl.ClientTslConfVerity(ca, cert, key, "")
		if err != nil {
			return nil, err
		}
		_ = etcdRegistry.Init(registry.TLSConfig(tlsConfig))
	}

	return etcdRegistry, nil
}

// DummyCmd : 去掉 go-micro 命令行使用
type DummyCmd struct{}

// NewDummyCmd :
func NewDummyCmd() *DummyCmd {
	return &DummyCmd{}
}

// App :
func (c *DummyCmd) App() *cli.App {
	return &cli.App{}
}

// Init :
func (c *DummyCmd) Init(opts ...cmd.Option) error {
	return nil
}

// Options :
func (c *DummyCmd) Options() cmd.Options {
	return cmd.Options{}
}
