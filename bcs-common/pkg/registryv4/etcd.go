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

package registryv4

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/go-micro/plugins/v4/registry/etcd"
	"github.com/google/uuid"
	"go-micro.dev/v4/registry"
	"go-micro.dev/v4/util/backoff"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

// NewEtcdRegistry create etcd registry instance
func NewEtcdRegistry(option *Options) Registry {
	// setting default options
	if option.TTL == time.Duration(0) {
		option.TTL = time.Second * 40
	}
	if option.Interval == time.Duration(0) {
		option.Interval = time.Second * 30
	}
	// create etcd registry
	r := etcd.NewRegistry(
		registry.Addrs(option.RegistryAddr...),
		registry.TLSConfig(option.Config),
	)
	// creat local service
	option.id = uuid.New().String()
	if option.Meta == nil {
		option.Meta = make(map[string]string)
	}
	option.Meta[types.UUID] = option.id
	svc := &registry.Service{
		Name:    option.Name,
		Version: option.Version,
		Nodes: []*registry.Node{
			{
				Id:       fmt.Sprintf("%s-%s", option.Name, option.id),
				Address:  option.RegAddr,
				Metadata: option.Meta,
			},
		},
	}
	ctx, stop := context.WithCancel(context.Background())
	e := &etcdRegister{
		option:       option,
		ctx:          ctx,
		stop:         stop,
		etcdregistry: r,
		localService: svc,
		localCache:   make(map[string]*registry.Service),
		registered:   false,
	}
	// check event handler
	if len(option.Modules) != 0 {
		// setting module that watch
		e.localModules = make(map[string]bool)
		for _, name := range option.Modules {
			e.localModules[name] = true
			_, _ = e.innerGet(name)
		}
		// start to watch all event
		go e.innerWatch(e.ctx)
	}
	return e
}

// etcd registry implementation
type etcdRegister struct {
	option       *Options
	ctx          context.Context
	stop         context.CancelFunc
	etcdregistry registry.Registry
	localService *registry.Service
	localLock    sync.RWMutex
	localCache   map[string]*registry.Service
	localModules map[string]bool
	registered   bool
}

// Register service information to registry
// register do not block, if module want to
// clean registe information, call Deregister
func (e *etcdRegister) Register() error {
	if e.registered {
		return fmt.Errorf("already registered")
	}
	if err := e.innerRegister(); err != nil {
		return err
	}
	e.registered = true
	// start background goroutine for interval keep alive
	// because we setting ttl for register
	go func() {
		tick := time.NewTicker(e.option.Interval)
		defer tick.Stop()
		for {
			select {
			case <-tick.C:
				// ready to keepAlive registered node information
				if err := e.innerRegister(); err != nil {
					blog.Errorf("register %s information %++v failed, %s", e.localService.Name, e.localService, err.Error())
					blog.Warnf("try register next tick...")
				}
			case <-e.ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (e *etcdRegister) innerRegister() (err error) {
	for i := 0; i < 3; i++ {
		if err = e.etcdregistry.Register(e.localService, registry.RegisterTTL(e.option.TTL)); err != nil {
			// try again until max failed
			roption := e.etcdregistry.Options()
			blog.Errorf("etcd registry register err, %s, options: %+v\n", err.Error(), roption)
			time.Sleep(backoff.Do(i + 1))
			continue
		}
		// register success, clean error
		err = nil
		break
	}
	return err
}

// Deregister clean service information from registry
func (e *etcdRegister) Deregister() error {
	// stop background keepalive goroutine
	e.stop()
	// clean registered node information
	if err := e.etcdregistry.Deregister(e.localService); err != nil {
		blog.Warnf("Deregister %s information %++v failed, %s", e.localService.Name, e.localService, err.Error())
	}
	e.registered = false
	return nil
}

// Get get specified service by name
func (e *etcdRegister) Get(name string) (*registry.Service, error) {
	if len(name) == 0 {
		return nil, nil
	}
	e.localLock.RLock()
	defer e.localLock.RUnlock()
	svc, ok := e.localCache[name]
	if !ok {
		blog.Warnf("registry get no %s in local cache", name)
		return nil, nil
	}
	return svc, nil
}

// nolint
func (e *etcdRegister) innerGet(name string) (*registry.Service, error) {
	// first, get details from registry
	svcs, err := e.etcdregistry.GetService(name)
	if err == registry.ErrNotFound {
		blog.Warnf("registry found no module %s under registry, clean local cache.", name)
		e.localLock.Lock()
		delete(e.localCache, name)
		e.localLock.Unlock()
		return nil, nil
	}
	if err != nil {
		blog.Errorf("registry get specified module %s failed, %s", name, err.Error())
		return nil, err
	}
	if len(svcs) == 0 {
		blog.Warnf("registry no module %s information", name)
		return nil, nil
	}
	// merge all version instance to one service
	if len(svcs) > 1 {
		blog.Infof("registry merge module %s different version instance and sort in cache", name)
		for _, svc := range svcs[1:] {
			svcs[0].Nodes = append(svcs[0].Nodes, svc.Nodes...)
		}
		sort.Slice(svcs[0].Nodes, func(i, j int) bool {
			return svcs[0].Nodes[i].Address < svcs[0].Nodes[j].Address
		})
	}
	// write to local cache
	e.localLock.Lock()
	e.localCache[name] = svcs[0]
	e.localLock.Unlock()
	return svcs[0], nil
}

func (e *etcdRegister) innerWatch(ctx context.Context) {
	// check if discovery is stopped
	select {
	case <-ctx.Done():
		blog.Infof("registry is ready to exit...")
		return
	default:
		blog.Infof("registry begin watch all registry modules....")
	}

	watcher, err := e.etcdregistry.Watch(registry.WatchContext(ctx))
	if err != nil {
		blog.Errorf("registry create watcher for all registry modules failed, %s. retry after a tick", err.Error())
		// retry after
		<-time.After(time.Second * 3)
		go e.innerWatch(ctx)
		return
	}
	defer watcher.Stop()
	for {
		select {
		case <-ctx.Done():
			blog.Infof("registry close watch backgroup goroutine...")
			return
		default:
			r, err := watcher.Next()
			if err != nil {
				blog.Errorf("registry watch registry loop err, %s, try to watch again~", err.Error())
				go e.innerWatch(ctx)
				return
			}
			if r == nil {
				blog.Warnf("registry watch got empty service information in event stream, keep watching...")
				continue
			}
			blog.Infof("registry watch information: module %s, details [%s] %+v", r.Service.Name, r.Action, r.Service)
			e.handleEvent(r)
		}
	}
}

func (e *etcdRegister) handleEvent(r *registry.Result) {
	fullName := r.Service.Name
	e.localLock.RLock()
	if _, ok := e.localModules[fullName]; !ok {
		blog.Warnf("registry do not expect module %s event, skip", fullName)
		e.localLock.RUnlock()
		return
	}
	e.localLock.RUnlock()
	_, err := e.innerGet(r.Service.Name)
	if err != nil {
		blog.Errorf("registry get module %s information failed, %s", fullName, err.Error())
		return
	}
	// check event handler
	if e.option.EvtHandler != nil {
		e.option.EvtHandler(fullName)
	}
}
