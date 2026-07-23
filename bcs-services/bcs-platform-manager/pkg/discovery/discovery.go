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

// Package discovery service discovery
package discovery

import (
	"context"
	"crypto/tls"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	etcd "github.com/go-micro/plugins/v4/registry/etcd"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/server"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/projectmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/utils"
)

const serverNameSuffix = ".bkbcs.tencent.com"

// ServiceDiscovery service discovery
type ServiceDiscovery struct {
	ctx             context.Context
	srv             micro.Service
	microRgt        registry.Registry
	clientTLSConfig *tls.Config
	TLSConfig       *tls.Config
}

// NewServiceDiscovery :
func NewServiceDiscovery(ctx context.Context, name, version, bindaddr, bindPort, addrIPv6 string) (
	*ServiceDiscovery, error) {
	metadata := map[string]string{}
	if addrIPv6 != "" {
		metadata[types.IPV6] = utils.GetListenAddr(addrIPv6, bindPort)
	}

	svr := server.NewServer(
		server.Name(name+serverNameSuffix),
		server.Version(version),
		server.Context(ctx),
	)

	_ = svr.Init(server.Advertise(utils.GetListenAddr(bindaddr, bindPort)))
	service := micro.NewService(micro.Server(svr), micro.Metadata(metadata))

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
	// initTLSConfig client TLS 配置
	if err := s.initTLSConfig(); err != nil {
		return err
	}

	// etcd 服务发现注册
	etcdRegistry, err := s.initEtcdRegistry()
	if err != nil {
		return err
	}

	err = s.InitComponentConfig()
	if err != nil {
		return err
	}

	if etcdRegistry != nil {
		s.srv.Init(micro.Registry(etcdRegistry))
	}
	return nil
}

// initTLSConfig xxx
// init server and client tls config
func (s *ServiceDiscovery) initTLSConfig() error {
	if len(config.G.TLSConf.ServerCert) != 0 && len(config.G.TLSConf.ServerKey) != 0 &&
		len(config.G.TLSConf.ServerCa) != 0 {
		tlsConfig, err := ssl.ServerTslConfVerityClient(config.G.TLSConf.ServerCa, config.G.TLSConf.ServerCert,
			config.G.TLSConf.ServerKey, static.ServerCertPwd)
		if err != nil {
			blog.Errorf("load platform manager server tls config failed, err %s", err.Error())
			return err
		}
		s.TLSConfig = tlsConfig
		blog.Info("load platform manager server tls config successfully")
	}

	if len(config.G.TLSConf.ClientCert) != 0 && len(config.G.TLSConf.ClientKey) != 0 &&
		len(config.G.TLSConf.ClientCa) != 0 {
		tlsConfig, err := ssl.ClientTslConfVerity(config.G.TLSConf.ClientCa, config.G.TLSConf.ClientCert,
			config.G.TLSConf.ClientKey, static.ClientCertPwd)
		if err != nil {
			blog.Errorf("load platform manager client tls config failed, err %s", err.Error())
			return err
		}
		s.clientTLSConfig = tlsConfig
		blog.Info("load platform manager client tls config successfully")
	}
	return nil
}

// initEtcdRegistry etcd 服务注册
func (s *ServiceDiscovery) initEtcdRegistry() (registry.Registry, error) {
	endpoints := config.G.Viper.GetString("etcd.endpoints")

	// 添加环境变量
	if endpoints == "" {
		endpoints = config.BCS_ETCD_HOST
	}

	if endpoints == "" {
		return nil, nil
	}

	s.microRgt = etcd.NewRegistry(registry.Addrs(strings.Split(endpoints, ",")...))

	ca := config.G.Viper.GetString("etcd.ca")
	cert := config.G.Viper.GetString("etcd.cert")
	key := config.G.Viper.GetString("etcd.key")
	if ca != "" && cert != "" && key != "" {
		tlsConfig, err := ssl.ClientTslConfVerity(ca, cert, key, "")
		if err != nil {
			return nil, err
		}

		s.clientTLSConfig = tlsConfig
		_ = s.microRgt.Init(registry.TLSConfig(tlsConfig))
	}

	return s.microRgt, nil
}

// InitComponentConfig init component config
func (s *ServiceDiscovery) InitComponentConfig() error {
	err := projectmanager.SetClientConifg(s.clientTLSConfig, s.microRgt)
	if err != nil {
		return err
	}

	err = clustermanager.SetClientConifg(s.clientTLSConfig, s.microRgt)
	if err != nil {
		return err
	}

	// err = pushmanager.SetClientConifg(s.clientTLSConfig, s.microRgt)
	// if err != nil {
	// 	return err
	// }

	return nil
}
