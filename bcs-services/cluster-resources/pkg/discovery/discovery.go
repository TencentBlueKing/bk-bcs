/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package discovery

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"go-micro.dev/v4/registry"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/common/errcode"
	log "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/logging"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/errorx"
)

// EventHandler ...
type EventHandler func(svcs []*registry.Service)

// ServiceDiscovery ...
type ServiceDiscovery struct {
	sync.RWMutex
	serviceName string
	curServices []*registry.Service
	microRtr    registry.Registry
	handler     EventHandler
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewServiceDiscovery ...
func NewServiceDiscovery(serviceName string, rtr registry.Registry) *ServiceDiscovery {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServiceDiscovery{
		serviceName: serviceName,
		microRtr:    rtr,
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动服务发现，获取指定模块可用 Services
func (d *ServiceDiscovery) Start() error {
	svcs, err := d.microRtr.GetService(d.serviceName)
	if err != nil {
		return errorx.New(errcode.ComponentErr, "no available service for %s found", d.serviceName)
	}
	log.Info(d.ctx, "in start(), get module %s services: %v", d.serviceName, svcs)

	d.Lock()
	d.curServices = svcs
	d.Unlock()
	if d.handler != nil {
		d.handler(svcs)
	}

	go d.watch(d.ctx)

	return nil
}

// 持续监听 Etcd 事件，及时更新指定模块的 Services
// nolint:cyclop
func (d *ServiceDiscovery) watch(ctx context.Context) {
	select {
	case <-ctx.Done():
		log.Info(ctx, "discovery %s is ready exit ...", d.serviceName)
	default:
		log.Info(ctx, "discovery %s begin watch service module ...", d.serviceName)
	}

	watcher, err := d.microRtr.Watch(registry.WatchContext(ctx), registry.WatchService(d.serviceName))
	if err != nil {
		log.Error(ctx, "discovery %s create watcher failed: %v, retry after wait ...", d.serviceName, err)
		<-time.After(time.Second * 5)
		go d.watch(ctx)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info(ctx, "discovery %s close watcher goroutine: %v", d.serviceName, ctx.Err())
			return
		default:
			event, watchErr := watcher.Next()
			if watchErr != nil {
				log.Error(ctx, "discovery %s watch registry loop err: %v, try to watch again", d.serviceName, watchErr)
				go d.watch(ctx)
				return
			}

			log.Info(ctx, "get registry watch event: %v", event)
			if event == nil {
				log.Warn(ctx, "get empty registry event, keep watching ...")
				continue
			}

			if event != nil && event.Service != nil {
				log.Info(ctx, "get registry event, action: %s, service %v", event.Action, event.Service)
			}

			svcs, getServiceErr := d.microRtr.GetService(d.serviceName)
			if getServiceErr != nil {
				log.Warn(ctx, "failed to get service for module %s, err: %v", d.serviceName, err)
				continue
			}
			log.Info(ctx, "in watch(), get module %s services: %v", d.serviceName, svcs)

			d.Lock()
			d.curServices = svcs
			d.Unlock()

			if d.handler != nil {
				log.Info(ctx, "update event handler in discovery %s", d.serviceName)
				d.handler(svcs)
			}
		}
	}
}

// GetRandServiceInst 随机获取可用服务实例
func (d *ServiceDiscovery) GetRandServiceInst(ctx context.Context) (*registry.Node, error) {
	allNodes := []*registry.Node{}

	if len(d.curServices) == 0 {
		log.Error(ctx, "discovery %s has no local service cache!", d.serviceName)
		return nil, errorx.New(errcode.ComponentErr, "依赖服务 %s 不可用", d.serviceName)
	}

	d.Lock()
	defer d.Unlock()

	for _, svc := range d.curServices {
		allNodes = append(allNodes, svc.Nodes...)
	}
	nodeLen := len(allNodes)
	if nodeLen == 0 {
		log.Error(ctx, "found no available node for service: %s", d.serviceName)
		return nil, errorx.New(errcode.ComponentErr, "依赖服务 %s 不可用", d.serviceName)
	}
	return allNodes[rand.Int()%nodeLen], nil
}

// RegisterEventHandler 注册事件回调函数
func (d *ServiceDiscovery) RegisterEventHandler(callback EventHandler) {
	d.handler = callback
}

// GetServiceList 获取当前缓存的服务实例
func (d *ServiceDiscovery) GetServiceList() []*registry.Service {
	d.Lock()
	defer d.Unlock()

	return d.curServices
}

// GetServiceName 获取服务发现指定模块
func (d *ServiceDiscovery) GetServiceName() string {
	return d.serviceName
}

// Stop 停止服务发现
func (d *ServiceDiscovery) Stop() {
	d.cancel()
}
