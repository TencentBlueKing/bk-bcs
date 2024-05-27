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

package discovery

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"go-micro.dev/v4/registry"
)

const (
	defaultDomain = ".bkbcs.tencent.com"
)

func init() {
}

// NewDiscovery create micro discovery implementation
// @param modules, bkbcs modules define in common/types, meshmanager, logmanager, mesosdriver and etc.
// @param handler, event callback
// @param r, go-micro registry implementation
func NewDiscovery(modules []string, handler EventHandler, r registry.Registry) Discovery {
	md := &MicroDiscovery{
		moduleFilter:  make(map[string]string),
		modules:       make(map[string]*registry.Service),
		microRegistry: r,
		eventHandler:  handler,
	}
	// initialize all module watch goroutines
	md.ctx, md.stopFunc = context.WithCancel(context.Background())
	for _, m := range modules {
		md.moduleFilter[m] = m
	}
	return md
}

// MicroDiscovery bkbcs go micro discovery implementation
type MicroDiscovery struct {
	sync.RWMutex
	ctx      context.Context
	stopFunc context.CancelFunc
	// module name that we expect watch, this is short name
	moduleFilter map[string]string
	// all modules information from micro registry
	// modules cache key is fullName, for example 30002.mesosdriver.bkbcs.tencent.com, storage.bkbcs.tencent.com
	modules map[string]*registry.Service
	// event callback
	eventHandler  EventHandler
	microRegistry registry.Registry
}

// Start init all necessary resource
func (d *MicroDiscovery) Start() error {
	// list all modules name
	svcs, err := d.ListAllServer()
	if err != nil {
		blog.Errorf("MicroDiscovery list all service failed, %s", err.Error())
		return err
	}
	if len(svcs) == 0 {
		blog.Warnf("etcd registry list all service in starting, found no data")
	}
	d.Lock()
	defer d.Unlock()
	for _, svc := range svcs {
		blog.Infof("start to init etcd registry info, %s", svc.Name)
		d.modules[svc.Name] = svc
	}
	go d.worker(d.ctx)
	return nil
}

// GetModuleServer module: modules.BCSModuleScheduler
// if mesos-apiserver/k8s-apiserver module=clusterId.{module}, for examples: 10001.mesosdriver, storage
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

// innerGetService xxx
// getService get registry service information and store in local cache
// @param: module, full name in registry, like meshmanager.bkbcs.tencent.com, 10003.mesosdriver.bkbcs.tencent.com
// @return: registry service, including all different version Node, empty if error
func (d *MicroDiscovery) innerGetService(module string) (*registry.Service, error) {
	if !strings.Contains(module, defaultDomain) {
		return nil, fmt.Errorf("lost domain `bkbcs.tencent.com`")
	}
	// first, get details from registry
	svcs, err := d.microRegistry.GetService(module)
	if err == registry.ErrNotFound {
		blog.Warnf("discovery found no module %s under registry, clean local cache.", module)
		d.Lock()
		delete(d.modules, module)
		d.Unlock()
		return nil, nil
	}
	if err != nil {
		blog.Errorf("discovery get specified module %s failed, %s", module, err.Error())
		return nil, err
	}
	if len(svcs) == 0 {
		blog.Warnf("etcd registry no module %s information", module)
		return nil, nil
	}
	// merge all version instance to one service
	if len(svcs) > 1 {
		blog.Infof("discovery merge module %s different version instance and sort in cache", module)
		for _, svc := range svcs[1:] {
			svcs[0].Nodes = append(svcs[0].Nodes, svc.Nodes...)
		}
		sort.Slice(svcs[0].Nodes, func(i, j int) bool {
			return svcs[0].Nodes[i].Address < svcs[0].Nodes[j].Address
		})
	}
	// write to local cache
	d.Lock()
	d.modules[module] = svcs[0]
	d.Unlock()
	return svcs[0], nil
}

// GetRandomServerInstance get random one instance of server
// if mesos-apiserver/k8s-apiserver module=clusterId.{module}, for examples: 10001.mesosdriver, storage
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
	selected, err := rand.Int(rand.Reader, big.NewInt(int64(nodeLength)))
	if err != nil {
		blog.Errorf("discovery rand err: %v", err)
		return nil, err
	}
	return svc.Nodes[selected.Int64()], nil
}

// ListAllServer list all registered server information
func (d *MicroDiscovery) ListAllServer() ([]*registry.Service, error) {
	svcList, err := d.microRegistry.ListServices()
	if err != nil {
		blog.Errorf("discovery list all micro registry failed, %s", err)
		return nil, err
	}
	if len(svcList) == 0 {
		blog.Warnf("discovery list no module registry information")
		return nil, nil
	}
	svcMap := make(map[string]*registry.Service)
	// merge all instance for same service and filter modules that we don't watch
	for _, svc := range svcList {
		shortName := strings.ReplaceAll(svc.Name, defaultDomain, "")
		IDName := strings.Split(shortName, ".")
		module := IDName[len(IDName)-1]
		// check module filter
		if _, ok := d.moduleFilter[module]; !ok {
			blog.V(5).Infof("module %s is not expected now, skip in list", svc.Name)
			continue
		}
		// same service with different version nodes
		if local, ok := svcMap[svc.Name]; ok {
			blog.V(3).Infof("module %s merge operation, version: %s", svc.Name, svc.Version)
			local.Nodes = append(local.Nodes, svc.Nodes...)
			sort.Slice(local.Nodes, func(i, j int) bool {
				return local.Nodes[i].Address < local.Nodes[j].Address
			})
			continue
		}
		blog.V(3).Infof("etcd registry list service %s", svc.Name)
		svcMap[svc.Name] = svc
	}
	var svcs []*registry.Service
	for k := range svcMap {
		svcs = append(svcs, svcMap[k])
	}
	return svcs, nil
}

// AddModuleWatch add new watch for specified module, Discovery will cache watched module info
// @param: module, bkbcs module name, such as meshmanager, logmanager, storage etc.
func (d *MicroDiscovery) AddModuleWatch(module string) error {
	// check if we already watch this module
	d.Lock()
	defer d.Unlock()
	if _, ok := d.moduleFilter[module]; ok {
		blog.V(3).Infof("module %s is already under watch, skip", module)
		return nil
	}
	d.moduleFilter[module] = module
	return nil
}

// DeleteModuleWatch clean watch for specified module
func (d *MicroDiscovery) DeleteModuleWatch(module string) error {
	d.Lock()
	defer d.Unlock()
	// clean expected module info for stopping event notification
	delete(d.moduleFilter, module)
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

// worker watch all events for all modules, only push event to callbacks Handler
// according specified module names
// @param ctx, context for exit control
func (d *MicroDiscovery) worker(ctx context.Context) {
	// check if discovery is stopped
	select {
	case <-ctx.Done():
		blog.V(3).Infof("discovery is ready to exit...")
		return
	default:
		blog.Infof("discovery begin watch all registry modules....")
	}

	watcher, err := d.microRegistry.Watch(registry.WatchContext(ctx))
	if err != nil {
		blog.Errorf("discovery create watcher for all registry modules failed, %s. retry after a tick", err.Error())
		// retry after
		<-time.After(time.Second * 3)
		go d.worker(ctx)
		return
	}
	defer watcher.Stop()
	for {
		select {
		case <-ctx.Done():
			blog.Infof("discovery close watch backgroup goroutine...")
			return
		default:
			r, err := watcher.Next()
			if err != nil {
				blog.Errorf("discovery watch registry loop err, %s, try to watch again~", err.Error())
				go d.worker(ctx)
				return
			}
			if r == nil {
				blog.Warnf("discovery watch got empty service information in event stream, keep watching...")
				continue
			}
			blog.Infof("discovery watch information: module %s, details [%s] %+v", r.Service.Name, r.Action, r.Service)
			d.handleEvent(r)
		}
	}
}

func (d *MicroDiscovery) handleEvent(r *registry.Result) {
	fullName := r.Service.Name
	allNodeSvc, err := d.innerGetService(r.Service.Name)
	if err != nil {
		blog.Errorf("discovery get module %s information failed, %s", fullName, err.Error())
		return
	}
	if allNodeSvc == nil {
		blog.Errorf("discovery found module %s all nodes exit...", r.Service.Name)
		return
	}
	shortName := strings.ReplaceAll(fullName, defaultDomain, "")
	bkbcsName := strings.Split(shortName, ".")
	d.RLock()
	defer d.RUnlock()
	if _, ok := d.moduleFilter[bkbcsName[len(bkbcsName)-1]]; !ok {
		blog.Warnf("discovery do not expect module %s[%s] event, skip", fullName, shortName)
		return
	}
	// check event handler
	if d.eventHandler != nil {
		d.eventHandler(shortName)
	}
}
