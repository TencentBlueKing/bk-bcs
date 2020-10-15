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
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/micro/go-micro/v2/registry"
)

const (
	defaultDomain = ".bkbcs.tencent.com"
)

func init() {
	rand.Seed(int64(time.Now().UnixNano()))
}

// NewDiscovery create micro discovery implementation
// @param modules, bkbcs modules define in common/types, meshmanager, logmanager, 30001.mesosdriver and etc.
// @param handler, event callback
// @param r, go-micro registry implementation
func NewDiscovery(modules []string, handler EventHandler, r registry.Registry) Discovery {
	md := &MicroDiscovery{
		modules:       make(map[string]*registry.Service),
		ctlFunc:       make(map[string]context.CancelFunc),
		microRegistry: r,
		eventHandler:  handler,
	}
	//initialize all module watch goroutines
	md.ctx, md.stopFunc = context.WithCancel(context.Background())
	for _, m := range modules {
		fullName := m + defaultDomain
		md.modules[fullName] = nil
		workerCxt, workerCancel := context.WithCancel(md.ctx)
		md.ctlFunc[fullName] = workerCancel
		//watch specified module info
		go md.worker(workerCxt, fullName)
	}
	return md
}

// MicroDiscovery bkbcs go micro discovery implementation
type MicroDiscovery struct {
	sync.RWMutex
	ctx      context.Context
	stopFunc context.CancelFunc
	// all modules information from micro registry
	// modules cache key is fullName, for example 30002.mesosdriver.bkbcs.tencent.com, storage.bkbcs.tencent.com
	modules map[string]*registry.Service
	ctlFunc map[string]context.CancelFunc
	//event callback
	eventHandler  EventHandler
	microRegistry registry.Registry
}

// GetModuleServer module: types.BCS_MODULE_SCHEDULER...
//if mesos-apiserver/k8s-apiserver module=clusterId.{module}, for examples: 10001.mesosdriver, storage
func (d *MicroDiscovery) GetModuleServer(module string) (*registry.Service, error) {
	fullName := fmt.Sprintf("%s%s", module, defaultDomain)
	d.RLock()
	defer d.RUnlock()
	svc, ok := d.modules[fullName]
	if !ok {
		blog.V(5).Infof("find no specified module %s(%s) in local watch cache", module, fullName)
		return nil, nil
	}
	return svc, nil
}

// GetRandomServerInstance get random one instance of server
//if mesos-apiserver/k8s-apiserver module=clusterId.{module}, for examples: 10001.mesosdriver, storage
func (d *MicroDiscovery) GetRandomServerInstance(module string) (*registry.Node, error) {
	fullName := fmt.Sprintf("%s%s", module, defaultDomain)
	d.RLock()
	defer d.RUnlock()
	svc, ok := d.modules[fullName]
	if !ok {
		blog.V(5).Infof("find no servie %s(%s) when get random instance", module, fullName)
		return nil, nil
	}
	if svc == nil {
		blog.V(3).Infof("discovery has no local cache information of %s(%s)", module, fullName)
		return nil, nil
	}
	nodeLength := len(svc.Nodes)
	if nodeLength == 0 {
		blog.V(3).Infof("discovery found no node information of %s(%s)", module, fullName)
		return nil, nil
	}
	selected := rand.Int() % nodeLength
	return svc.Nodes[selected], nil
}

//ListAllServer list all registed server information
func (d *MicroDiscovery) ListAllServer() ([]*registry.Service, error) {
	return d.microRegistry.ListServices()
}

// AddModuleWatch add new watch for specified module, Discovery will cache watched module info
func (d *MicroDiscovery) AddModuleWatch(module string) error {
	fullName := module + defaultDomain
	d.Lock()
	defer d.Unlock()
	if _, ok := d.modules[fullName]; ok {
		blog.V(3).Infof("module %s is already under watch, skip", fullName)
		return nil
	}
	d.modules[fullName] = nil
	//ready to start watch worker
	workerCxt, workerCancel := context.WithCancel(d.ctx)
	d.ctlFunc[fullName] = workerCancel
	go d.worker(workerCxt, fullName)
	return nil
}

// DeleteModuleWatch clean watch for specified module
func (d *MicroDiscovery) DeleteModuleWatch(module string) error {
	fullName := module + defaultDomain
	d.Lock()
	defer d.Unlock()
	if _, ok := d.modules[fullName]; !ok {
		blog.V(3).Infof("module %s is not under watch, skip", fullName)
		return nil
	}
	delete(d.modules, fullName)
	//ready to stop watch worker
	cancel := d.ctlFunc[fullName]
	delete(d.ctlFunc, fullName)
	cancel()
	return nil
}

// RegisterEventFunc register event handle function
func (d *MicroDiscovery) RegisterEventFunc(handleFunc EventHandler) {
	if d.eventHandler != nil {
		blog.V(5).Infof("micro discovery change event handler in runtime")
	}
	d.eventHandler = handleFunc
}

// Stop close discovery
func (d *MicroDiscovery) Stop() {
	d.stopFunc()
}

// worker watch event for specified module, push event to callbacks Handler
// @param ctx, context for exit control
// @param module, full name for module, likes usermanager.bkbcs.tencent.com
func (d *MicroDiscovery) worker(ctx context.Context, module string) {
	//first, get details into cache
	svcs, err := d.microRegistry.GetService(module)
	if err != nil && err != registry.ErrNotFound {
		blog.Errorf("discovery get specified module %s failed, %s. try next tick", module, err.Error())
		<-time.After(time.Second * 3)
		select {
		case <-ctx.Done():
			blog.Infof("discovery for module %s is already done, exit", module)
			return
		default:
			go d.worker(ctx, module)
		}
		return
	}
	if len(svcs) > 0 {
		blog.Infof("discovery merge module %s different version instance and sort in cache", module)
		for _, svc := range svcs[1:] {
			svcs[0].Nodes = append(svcs[0].Nodes, svc.Nodes...)
		}
		sort.Slice(svcs[0].Nodes, func(i, j int) bool {
			return svcs[0].Nodes[i].Address < svcs[0].Nodes[j].Address
		})
		d.Lock()
		d.modules[module] = svcs[0]
		d.Unlock()
	}
	blog.V(3).Infof("discovery begin watch module %s", module)
	watcher, err := d.microRegistry.Watch(registry.WatchService(module))
	if err != nil {
		blog.Errorf("discovery create watcher for %s failed, %s. retry after a tick", module, err.Error())
		//retry after
		<-time.After(time.Second * 3)
		go d.worker(ctx, module)
		return
	}
	defer watcher.Stop()
	for {
		select {
		case <-ctx.Done():
			blog.V(3).Infof("discovery close watch module %s", module)
			return
		default:
			r, err := watcher.Next()
			if err != nil {
				blog.Errorf("discovery watch %s in loop err, %s", module, err.Error())
				go d.worker(ctx, module)
				return
			}
			if r == nil {
				blog.Warnf("discovery watch %s got empty service information in event stream", module)
				continue
			}
			blog.V(5).Infof("discovery watch %s module, details [%s] %+v", module, r.Action, r.Service)
			d.handleEvent(module, r)
		}
	}
}

func (d *MicroDiscovery) handleEvent(module string, r *registry.Result) {
	switch r.Action {
	case "create":
	case "update":
		blog.V(3).Infof("module %s created/updated, push to local cache", module)
		//sort all nodes, bkbcs use first node as master in some situation
		sort.Slice(r.Service.Nodes, func(i, j int) bool {
			return r.Service.Nodes[i].Address < r.Service.Nodes[j].Address
		})
		//push to local cache
		d.Lock()
		d.modules[module] = r.Service
		d.Unlock()
	case "delete":
		//clean local cache
		blog.V(3).Infof("module %s delete, clean local cache", module)
		d.Lock()
		d.modules[module] = nil
		d.Unlock()
	default:
		blog.Errorf("discovery module %s got unknown action [%s]", module, r.Action)
		return
	}
	//check event handler
	if d.eventHandler != nil {
		conciseName := strings.ReplaceAll(module, defaultDomain, "")
		d.eventHandler(conciseName)
	}
}
