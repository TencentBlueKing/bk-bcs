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
	"crypto/tls"
	"errors"
	"strings"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/client/grpc"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	bcsRegistry "github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/util"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-alert-manager/pkg/proto/alertmanager"
)

var (
	// ErrServerInit server is null
	ErrServerInit = errors.New("server is null")
	// ErrServiceNotFound show service not found
	ErrServiceNotFound = errors.New("service name not found")
	// ErrEtcdServer show etcdServers invalid
	ErrEtcdServer = errors.New("etcdServers invalid")
)

// ModuleName for module name
type ModuleName string

const (
	// AlertManager alertmanager module
	AlertManager ModuleName = "alertmanager"
)

var serviceNameToModule = map[ModuleName]string{
	AlertManager: "alertmanager.bkbcs.tencent.com",
}

// Discovery interface for discovery service by go-micro registry
type Discovery interface {
	GetMicroServiceByName(name ModuleName) (interface{}, error)
}

// DiscoveryService for discovery client by micro registry service
type DiscoveryService struct {
	MicroRegistry registry.Registry
	MicroClient   client.Client
}

func initEtcdRegistry(etcdOpt bcsRegistry.CMDOptions) (registry.Registry, error) {
	if len(etcdOpt.Address) == 0 {
		return nil, ErrEtcdServer
	}

	servers := strings.Split(etcdOpt.Address, ";")

	var (
		secureEtcd bool
		etcdTLS    *tls.Config
		err        error
	)

	if len(etcdOpt.CA) != 0 && len(etcdOpt.Cert) != 0 && len(etcdOpt.Key) != 0 {
		secureEtcd = true

		etcdTLS, err = ssl.ClientTslConfVerity(etcdOpt.CA, etcdOpt.Cert, etcdOpt.Key, "")
		if err != nil {
			return nil, err
		}
	}

	etcdRegistry := etcd.NewRegistry(
		registry.Addrs(servers...),
		registry.Secure(secureEtcd),
		registry.TLSConfig(etcdTLS),
	)
	if err = etcdRegistry.Init(); err != nil {
		return nil, err
	}

	return etcdRegistry, nil
}

// NewDiscoveryService for init DiscoveryService
func NewDiscoveryService(schedConf util.SchedConfig) (Discovery, error) {
	etcdRegistry, err := initEtcdRegistry(schedConf.EtcdConf)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{}
	if schedConf.Scheduler.ClientCertFile != "" && schedConf.Scheduler.ClientKeyFile != "" {
		tlsConfig, err = ssl.ClientTslConfVerity(schedConf.Scheduler.ClientCAFile, schedConf.Scheduler.ClientCertFile,
			schedConf.Scheduler.ClientKeyFile, static.ClientCertPwd)
		if err != nil {
			blog.Errorf("NewDiscoveryService init ClientTslConfVerity failed: %v", err)
			return nil, err
		}
	}

	grpcClient := grpc.NewClient(
		client.Registry(etcdRegistry),
		grpc.AuthTLS(tlsConfig),
	)
	grpcClient.Init()

	return &DiscoveryService{
		MicroRegistry: etcdRegistry,
		MicroClient:   grpcClient,
	}, nil
}

// GetMicroServiceByName for get service object by moduleName
func (ds *DiscoveryService) GetMicroServiceByName(name ModuleName) (interface{}, error) {
	if ds == nil {
		return nil, ErrServerInit
	}

	module, ok := serviceNameToModule[name]
	if !ok {
		return nil, ErrServiceNotFound
	}

	switch module {
	case "alertmanager.bkbcs.tencent.com":
		return alertmanager.NewAlertManagerService(module, ds.MicroClient), nil
	default:
		return nil, ErrServiceNotFound
	}
}
