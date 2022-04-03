/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package discovery

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/micro/go-micro/v2/registry"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/logging"
)

const (
	rewatchSecond = 1
)

// EventHandler discovery event handler interface
type EventHandler func(svcs []*registry.Service)

// ModuleDiscovery discovery service endpoints
type ModuleDiscovery struct {
	sync.RWMutex
	module        string
	curServices   []*registry.Service
	microRegistry registry.Registry
	handler       EventHandler
	stop          chan bool
}

// NewModuleDiscovery create discovery
func NewModuleDiscovery(module string, r registry.Registry) *ModuleDiscovery {
	return &ModuleDiscovery{
		module:        module,
		microRegistry: r,
		stop:          make(chan bool),
	}
}

// Start start discovery
func (md *ModuleDiscovery) Start() error {
	watcher, err := md.microRegistry.Watch(registry.WatchService(md.module))
	if err != nil {
		return fmt.Errorf("failed to create registry watcher for module %s, err %s", md.module, err.Error())
	}
	go func() {
		var err error
		for {
			select {
			case <-md.stop:
				if watcher != nil {
					watcher.Stop()
				}
				return
			default:
				if watcher == nil {
					watcher, err = md.microRegistry.Watch(registry.WatchService(md.module))
					if err != nil {
						logging.Error("see empty watcher and failed to create registry watcher for module %s, err %s",
							md.module, err.Error())
						time.Sleep(rewatchSecond * time.Second)
						continue
					}
				}

				if err := md.watchRegistry(watcher); err != nil {
					logging.Error("failed when watching registry for module %s, err %s", md.module, err.Error())
					time.Sleep(rewatchSecond * time.Second)
				}

				if watcher != nil {
					watcher.Stop()
					watcher = nil
				}
			}
		}
	}()
	return nil
}

func (md *ModuleDiscovery) watchRegistry(w registry.Watcher) error {
	stop := make(chan bool)
	defer func() {
		close(stop)
	}()

	go func() {
		defer w.Stop()
		select {
		case <-stop:
			return
		case <-md.stop:
			return
		}
	}()

	// receive event, to update service
	svcs, err := md.microRegistry.GetService(md.module)
	if err != nil {
		logging.Error("failed to get service for module %s, err %s", md.module, err.Error())
		return err
	}
	md.Lock()
	md.curServices = svcs
	md.Unlock()
	if md.handler != nil {
		md.handler(svcs)
	}

	for {
		result, err := w.Next()
		if err != nil {
			if err != registry.ErrWatcherStopped {
				return err
			}
			break
		}
		if result != nil && result.Service != nil {
			logging.Info("service watch result, action %s, service %s", result.Action, result.Service.Name)
		}
		// receive event, to update service
		svcs, err := md.microRegistry.GetService(md.module)
		if err != nil {
			logging.Error("failed to get service for module %s, err %s", md.module, err.Error())
			continue
		}
		logging.Info("get services %v", svcs)

		md.Lock()
		md.curServices = svcs
		md.Unlock()
		if md.handler != nil {
			md.handler(svcs)
		}
	}
	return nil
}

// GetService get service from remote
func (md *ModuleDiscovery) GetService() []*registry.Service {
	md.RLock()
	defer md.RUnlock()
	return md.curServices
}

// GetRandomServiceNode get random instance by curServices
func (md *ModuleDiscovery) GetRandomServiceNode() (*registry.Node, error) {
	allServiceNodes := make([]*registry.Node, 0)

	if len(md.curServices) == 0 {
		logging.Info("discovery has no local service cache[%s]", md.module)
		return nil, errors.New("curServices is empty")
	}

	md.Lock()
	defer md.Unlock()
	for i := range md.curServices {
		allServiceNodes = append(allServiceNodes, md.curServices[i].Nodes...)
	}

	nodeLength := len(allServiceNodes)
	if nodeLength == 0 {
		logging.Info("discovery found no node information of %s", md.module)
		return nil, errors.New("allServiceNodes is empty")
	}
	selected := rand.Int() % nodeLength
	return allServiceNodes[selected], nil
}

// GetModuleName get module name
func (md *ModuleDiscovery) GetModuleName() string {
	return md.module
}

// RegisterEventHandler register event callback function
func (md *ModuleDiscovery) RegisterEventHandler(eHandler EventHandler) {
	md.handler = eHandler
}

// Stop stop discovery
func (md *ModuleDiscovery) Stop() {
	select {
	case <-md.stop:
		return
	default:
		close(md.stop)
	}
}
