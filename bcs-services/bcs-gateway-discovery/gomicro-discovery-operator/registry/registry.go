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

package registry

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"

	"gomicro-discovery-operator/constant"

	gatewayv1beta1 "github.com/TencentBlueKing/blueking-apigateway-operator/api/v1beta1"
	frametypes "github.com/TencentBlueKing/blueking-apigateway-operator/pkg/discovery-operator-frame/types"

	etcd "github.com/go-micro/plugins/v4/registry/etcd"
	json "github.com/json-iterator/go"
	"github.com/rotisserie/eris"
	"go-micro.dev/v4/registry"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RegistryName ...
const RegistryName = "bk-gomicro-discovery"

// MicroRegistry ...
type MicroRegistry struct {
	sync.Mutex
	Client client.Client
	regMap map[string]registry.Registry
}

var logger = zap.L().With(zap.String("registry", RegistryName))

// Init ...
func (r *MicroRegistry) Init() {
	r.regMap = make(map[string]registry.Registry)
}

// Watch ...
func (r *MicroRegistry) Watch(
	ctx context.Context,
	svcName, namespace string,
	svcConfig map[string]interface{},
	callBack frametypes.CallBack,
) error {
	conf := &ServiceConfig{}
	by, _ := json.Marshal(svcConfig)
	err := json.Unmarshal(by, conf)
	logger := logger.With(zap.String("method", "watch")).With(zap.String("service", svcName))
	if err != nil {
		logger.Error("Marshal service config to ServiceConfig failed", zap.Any("error", err))
		return err
	}
	if ok, err := conf.Validate(r.Client, namespace); !ok {
		logger.Error("Validate service config failed", zap.Any("error", err))
		return err
	}
	reg := r.getOrCreateRegistry(conf)
	// watch
	watcher, err := reg.Watch(registry.WatchService(svcName))
	if err != nil {
		logger.Error("Watch service failed", zap.Any("error", err))
		return err
	}

	// list all nodes when first start watch
	eps, err := r.innerList(svcName, reg, conf)
	if err != nil {
		logger.Error("Failed GetService from gomicro registry", zap.Any("error", err))
	}
	err = callBack(eps)
	if err != nil {
		logger.Error("Watch callBack function failed", zap.Any("error", err))
	}

	messageChan := make(chan *registry.Result)
	go r.innerWatch(logger, watcher, messageChan)
	for {
		select {
		case <-ctx.Done():
			logger.Info("Watch context cancled, watch stopped", zap.Any("reason", ctx.Err))
			watcher.Stop()
			return nil
		case result, ok := <-messageChan:
			if !ok {
				logger.Error("watch channel in innerwatch method closed")
				return eris.Errorf("watch channel in innerwatch method closed")
			}
			eps, err := r.innerList(svcName, reg, conf)
			if err != nil {
				logger.Error("Failed GetService from gomicro registry", zap.Any("error", err))
				continue
			}
			err = callBack(eps)
			if err != nil {
				logger.Error("Watch callBack function failed", zap.Any("event", result.Action), zap.Any("error", err))
			}
		}
	}
}

func (r *MicroRegistry) innerWatch(logger *zap.Logger, watcher registry.Watcher, messageChan chan<- *registry.Result) {
	for {
		res, err := watcher.Next()
		if err != nil {
			logger.Error("Watch from etcd failed", zap.Any("error", err))
			close(messageChan)
			return
		}
		messageChan <- res
	}
}

// List ...
func (r *MicroRegistry) List(
	svcName, namespace string,
	svcConfig map[string]interface{},
) (*gatewayv1beta1.BkGatewayEndpointsSpec, error) {
	conf := &ServiceConfig{}
	by, _ := json.Marshal(svcConfig)
	err := json.Unmarshal(by, conf)
	logger := logger.With(zap.String("method", "list")).With(zap.String("service", svcName))
	if err != nil {
		logger.Error("Marshal service config to ServiceConfig failed", zap.Any("error", err))
		return nil, err
	}
	if ok, _ := conf.Validate(r.Client, namespace); !ok {
		logger.Error("Validate service config failed", zap.Any("error", err))
		return nil, err
	}
	reg := r.getOrCreateRegistry(conf)
	return r.innerList(svcName, reg, conf)
}

func (r *MicroRegistry) innerList(
	svcName string,
	reg registry.Registry,
	conf *ServiceConfig,
) (*gatewayv1beta1.BkGatewayEndpointsSpec, error) {
	svcs, err := reg.GetService(svcName)
	if err != nil && err != registry.ErrNotFound {
		logger.Error("List service failed", zap.Any("error", err))
		return nil, err
	}
	if len(svcs) == 0 {
		return nil, nil
	}
	nodes := make([]gatewayv1beta1.BkGatewayNode, 0)
	for _, node := range svcs[0].Nodes {
		hostType, gatewayNode, gatewayNodev6, err := r.extractNodeFromServiceNode(node, conf)
		if err != nil {
			continue
		}

		// case ipv4 only
		if conf.DisableIPv6 {
			if hostType == constant.HostTypeIPV6 {
				logger.Info("skip ipv6 address", zap.Any("service", svcName), zap.Any("address", node.Address))
			} else {
				nodes = append(nodes, *gatewayNode)
			}
			continue
		}

		// case ipv6 only
		if conf.IPv6Only {
			if hostType == constant.HostTypeIPV4 {
				logger.Info("skip ipv4 address", zap.Any("service", svcName), zap.Any("address", node.Address))
			} else {
				nodes = append(nodes, *gatewayNode)
			}
			if gatewayNodev6 != nil {
				nodes = append(nodes, *gatewayNodev6)
			}
			continue
		}

		// dual stack
		nodes = append(nodes, *gatewayNode)
		if gatewayNodev6 != nil {
			nodes = append(nodes, *gatewayNodev6)
		}
	}
	return &gatewayv1beta1.BkGatewayEndpointsSpec{Nodes: nodes}, nil
}

func (r *MicroRegistry) extractNodeFromServiceNode(
	node *registry.Node,
	conf *ServiceConfig,
) (constant.HostType, *gatewayv1beta1.BkGatewayNode, *gatewayv1beta1.BkGatewayNode, error) {
	host, portStr, err := net.SplitHostPort(node.Address)
	if err != nil {
		logger.Error("Split host port failed", zap.Any("error", err), zap.Any("address", node.Address))
		return constant.HostTypeNULL, nil, nil, err
	}
	port, _ := strconv.Atoi(portStr)
	ip := net.ParseIP(host)
	hostType := constant.HostTypeDomain
	if ip != nil {
		if ip.To4() != nil {
			hostType = constant.HostTypeIPV4
		} else {
			hostType = constant.HostTypeIPV6
			host = fmt.Sprintf("[%s]", host)
		}
	}
	var gatewayNode, gatewayNodeV6 *gatewayv1beta1.BkGatewayNode
	// node in node address
	gatewayNode = &gatewayv1beta1.BkGatewayNode{
		Host:     host,
		Port:     port,
		Weight:   100,
		Priority: intPtr(0),
	}
	if conf.portOverride != 0 {
		gatewayNode.Port = int(conf.portOverride)
	}

	// v6node in node metadata
	addressV6, ok := node.Metadata[constant.MetadataKeyIPv6]
	if ok {
		var portV6 int
		hostV6, portStrV6, err := net.SplitHostPort(addressV6)
		if err != nil {
			hostV6 = addressV6
			portV6 = port
		} else {
			portV6, _ = strconv.Atoi(portStrV6)
		}
		if net.ParseIP(hostV6).To4() != nil {
			return hostType, gatewayNode, nil, nil
		}
		if hostV6[0] != '[' {
			hostV6 = fmt.Sprintf("[%s]", hostV6)
		}
		gatewayNodeV6 = &gatewayv1beta1.BkGatewayNode{
			Host:     hostV6,
			Port:     portV6,
			Weight:   100,
			Priority: intPtr(0),
		}
		if conf.portOverride != 0 {
			gatewayNodeV6.Port = int(conf.portOverride)
		}
	}

	return hostType, gatewayNode, gatewayNodeV6, nil
}

func (r *MicroRegistry) getOrCreateRegistry(conf *ServiceConfig) registry.Registry {
	confKey := conf.String()
	var reg registry.Registry
	var ok bool
	if reg, ok = r.regMap[confKey]; !ok {
		r.Lock()
		defer r.Unlock()

		reg, ok = r.regMap[confKey]
		if ok {
			return reg
		}

		logger.Debug("Create new registry with config", zap.Any("registry_config", confKey))
		opts := []registry.Option{}
		opts = append(opts, registry.Addrs(conf.Addrs...))
		if conf.tlsConfig != nil {
			opts = append(opts, registry.TLSConfig(conf.tlsConfig))
		} else if !conf.noAuth {
			opts = append(opts, etcd.Auth(conf.Username, conf.Password))
		}
		reg = etcd.NewRegistry(opts...)
		r.regMap[confKey] = reg
		return reg
	}
	return reg
}

// DiscoveryMethods ...
func (r *MicroRegistry) DiscoveryMethods() frametypes.SupportMethods {
	return frametypes.WatchAndListSupported
}

// Name ...
func (r *MicroRegistry) Name() string {
	return RegistryName
}
