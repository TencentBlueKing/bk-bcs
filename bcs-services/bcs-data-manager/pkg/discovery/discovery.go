/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package discovery

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/micro/go-micro/v2/registry"
)

// Discovery discovery interface
type Discovery interface {
	// Start() start discover in background
	Start() error
	// GetRandomServiceInstance get random service instance
	GetRandomServiceInstance() (*registry.Node, error)
	// RegisterEventHandler register callback handler
	RegisterEventHandler(callBackHandler EventHandler)
	// GetServiceList() get service instances list
	GetServiceList() []*registry.Service
	// GetModuleName() get service name
	GetModuleName() string
	// Stop() stop discover
	Stop()
}

var (
	// ErrServerNotInit server not init
	ErrServerNotInit = errors.New("server not init")
)

// EventHandler discovery event handler interface
type EventHandler func(svcs []*registry.Service)

// ServiceDiscovery discovery service endpoints
type ServiceDiscovery struct {
	sync.RWMutex
	service string
	// cache serviceInfo
	curServices   []*registry.Service
	microRegistry registry.Registry
	// call back handler
	handler EventHandler
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewServiceDiscovery create discovery
func NewServiceDiscovery(service string, r registry.Registry) *ServiceDiscovery {
	ctx, cancel := context.WithCancel(context.Background())
	return &ServiceDiscovery{
		service:       service,
		microRegistry: r,
		ctx:           ctx,
		cancel:        cancel,
	}
}

// Start start discovery
func (sd *ServiceDiscovery) Start() error {
	if sd == nil {
		return ErrServerNotInit
	}
	serviceList, err := sd.microRegistry.GetService(sd.service)
	if err != nil {
		blog.Errorf("failed to get service[%s], err: %v", sd.service, err.Error())
		return err
	}

	blog.Infof("sd.microRegistry.GetService: %+v", serviceList)

	sd.Lock()
	sd.curServices = serviceList
	sd.Unlock()
	if sd.handler != nil {
		sd.handler(serviceList)
	}

	// begin watch
	go sd.worker(sd.ctx)

	return nil
}

func (sd *ServiceDiscovery) worker(ctx context.Context) {
	select {
	case <-ctx.Done():
		blog.V(3).Infof("discovery is ready ti exit...")
		return
	default:
		blog.Infof("discovery begin watch service module")
	}

	watcher, err := sd.microRegistry.Watch(registry.WatchContext(ctx), registry.WatchService(sd.service))
	if err != nil {
		blog.Errorf("discovery create watcher failed: %v, retry after wait", err.Error())
		// retry after
		<-time.After(time.Second * 5)
		go sd.worker(ctx)
		return
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			blog.Infof("discovery close watcher goroutine... %v", ctx.Err())
			return
		default:
			r, err := watcher.Next()
			if err != nil {
				blog.Errorf("discovery watch registry loop err, %s, try to watch again~", err.Error())
				go sd.worker(ctx)
				return
			}

			fmt.Printf("%+v\n", r)

			if r == nil {
				blog.Warnf("discovery watch got empty service information in event stream, keep watching...")
				continue
			}

			if r != nil && r.Service != nil {
				blog.V(5).Infof("service watch result, action %s, service %+v", r.Action, r.Service)
			}

			// receive event and update service
			svcs, err := sd.microRegistry.GetService(sd.service)
			if err != nil {
				blog.Warnf("failed to get service for module %s, err %s", sd.service, err.Error())
				continue
			}
			blog.V(5).Infof("get services %v", svcs)

			sd.Lock()
			sd.curServices = svcs
			sd.Unlock()

			if sd.handler != nil {
				blog.Infof("event handler update discovery service module %s", sd.service)
				sd.handler(svcs)
			}
		}
	}
}

// GetRandomServiceInstance get random instance by curServices
func (sd *ServiceDiscovery) GetRandomServiceInstance() (*registry.Node, error) {
	if sd == nil {
		return nil, ErrServerNotInit
	}
	allServiceNodes := []*registry.Node{}

	if len(sd.curServices) == 0 {
		blog.Infof("discovery has no local service cache[%s]", sd.service)
		return nil, errors.New("curServices is empty")
	}

	sd.Lock()
	defer sd.Unlock()
	for i := range sd.curServices {
		allServiceNodes = append(allServiceNodes, sd.curServices[i].Nodes...)
	}

	nodeLength := len(allServiceNodes)
	if nodeLength == 0 {
		blog.V(3).Infof("discovery found no node information of %s", sd.service)
		return nil, errors.New("allServiceNodes is empty")
	}
	randNum, err := rand.Int(rand.Reader, big.NewInt(6))
	if err != nil {
		blog.Errorf("get rand num error: %v", err)
		return nil, errors.New("get rand num error")
	}
	selected := randNum.Int64() % int64(nodeLength)
	return allServiceNodes[selected], nil
}

// RegisterEventHandler register external callBackHandler
func (sd *ServiceDiscovery) RegisterEventHandler(callBackHandler EventHandler) {
	if sd == nil {
		return
	}
	sd.handler = callBackHandler
}

// GetServiceList get all instances for service
func (sd *ServiceDiscovery) GetServiceList() []*registry.Service {
	if sd == nil {
		return nil
	}
	sd.Lock()
	defer sd.Unlock()

	return sd.curServices
}

// GetModuleName get discovery service
func (sd *ServiceDiscovery) GetModuleName() string {
	if sd == nil {
		return ""
	}
	return sd.service
}

// Stop stop discovery
func (sd *ServiceDiscovery) Stop() {
	if sd != nil {
		sd.cancel()
	}
}
