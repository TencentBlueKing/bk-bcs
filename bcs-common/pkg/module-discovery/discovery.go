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

// Package modulediscovery xxx
package modulediscovery

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

// DiscoveryV2  discover bcs module, examples: bcs-api、bcs-scheduler、bcs-mesos-driver
// base on bkbcs zookeeper discovery mechanism
type DiscoveryV2 struct {
	sync.RWMutex

	// discover bcs endpoints client
	rd     *RegisterDiscover.RegDiscover
	cliTLS *tls.Config // nolint

	// context
	rootCxt context.Context
	cancel  context.CancelFunc

	// endpoint server infos
	// key = bcs-module or bcs-module/clusterid
	// example: bcs-api、bcs-storage、mesosdriver/BCS-MESOS-10000、kubernetedriver/BCS-K8S-15000
	// value = []byte array, []byte is marshal types.ServerInfo data
	servers map[string][]interface{}
	// register RegisterDiscover callback function
	// when endpoint node changed, call it
	eventHandler EventHandleFunc
	// if modules != nil
	// DiscoveryV2 only discover the specified modules
	// then lose sight of others
	modules []string
	// watched key
	watchedKey map[string]struct{}
}

// NewDiscoveryV2 create a object of DiscoveryV2
func NewDiscoveryV2(zkserv string, modules []string) (ModuleDiscovery, error) {
	blog.Infof("DiscoveryV2 start...")
	rd := &DiscoveryV2{
		rd:         RegisterDiscover.NewRegDiscoverEx(zkserv, 10*time.Second),
		servers:    make(map[string][]interface{}, 0),
		modules:    modules,
		watchedKey: make(map[string]struct{}),
	}

	err := rd.start()
	if err != nil {
		return nil, err
	}
	blog.Infof("DiscoveryV2 working...")
	return rd, nil
}

// GetModuleServers module: types.BCS_MODULE_SCHEDULER...
// list all servers
// if mesos-apiserver/k8s-apiserver module={module}/clusterid, for examples: mesosdriver/BCS-TESTBCSTEST01-10001
func (r *DiscoveryV2) GetModuleServers(moduleName string) ([]interface{}, error) {
	r.RLock()
	defer r.RUnlock()

	servs := make([]interface{}, 0)
	for k, v := range r.servers {
		if strings.Contains(k, moduleName) {
			servs = append(servs, v...)
		}
	}
	if len(servs) == 0 {
		return nil, fmt.Errorf("Module %s don't have endpoints", moduleName)
	}

	return servs, nil
}

// GetRandModuleServer get random one server
func (r *DiscoveryV2) GetRandModuleServer(moduleName string) (interface{}, error) {
	r.RLock()
	defer r.RUnlock()

	servs, ok := r.servers[moduleName]
	if !ok {
		return nil, fmt.Errorf("Module %s not found", moduleName)
	}

	if len(servs) == 0 {
		return nil, fmt.Errorf("Module %s don't have endpoints", moduleName)
	}

	// rand
	rand.Seed(int64(time.Now().Nanosecond()))
	serv := servs[rand.Intn(len(servs))] // nolint
	return serv, nil
}

// RegisterEventFunc register event handle function
func (r *DiscoveryV2) RegisterEventFunc(handleFunc EventHandleFunc) {
	r.eventHandler = handleFunc
}

// start the DiscoveryV2
func (r *DiscoveryV2) start() error {
	// create root context
	r.rootCxt, r.cancel = context.WithCancel(context.Background())

	// start regdiscover
	if err := r.rd.Start(); err != nil {
		blog.Error("fail to start register and discover serv. err:%s", err.Error())
		return err
	}

	// discover all bcs module serviceinfos
	return r.discoverEndpoints(types.BCS_SERV_BASEPATH)
}

// discoverEndpoints xxx
// recursive discover bcs module serverinfo
func (r *DiscoveryV2) discoverEndpoints(path string) error {
	blog.V(3).Infof("discover %s endpoints", path)
	if r.modules != nil && path != types.BCS_SERV_BASEPATH {
		exist := false
		for _, module := range r.modules {
			if strings.Contains(path, fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, module)) {
				exist = true
				break
			}
		}
		if !exist {
			blog.V(3).Infof("path %s not in modules(%v), and ingore", path, r.modules)
			return nil
		}
	}

	r.Lock()
	// get path children
	zvs, err := r.rd.DiscoverNodesV2(path)
	if err != nil {
		blog.V(3).Infof("discover %s nodes error %s", path, err.Error())
		return err
	}
	blog.V(3).Infof("module-discovery get path %s servers %d, nodes %d", path, len(zvs.Server), len(zvs.Nodes))
	// if leaf node, then parse bcs module serverinfo
	if len(zvs.Server) != 0 {
		key := strings.TrimPrefix(path, fmt.Sprintf("%s/", types.BCS_SERV_BASEPATH))
		val := make([]interface{}, 0)
		for _, v := range zvs.Server {
			val = append(val, v)
		}
		r.servers[key] = val
		blog.V(3).Infof("set server %s endpoints %v", key, val)
		if r.eventHandler != nil {
			r.eventHandler(key)
		}
	}
	blog.V(5).Infof("module-discovery get path %s nodes details: %+v", path, zvs.Nodes)
	// watch key
	_, ok := r.watchedKey[path]
	if !ok {
		go r.discoverModules(path, true)
	}
	r.Unlock()

	// discovery path's children node
	if len(zvs.Server) == 0 {
		for _, v := range zvs.Nodes {
			_ = r.discoverEndpoints(fmt.Sprintf("%s/%s", path, v))
		}
	}

	return nil
}

func (r *DiscoveryV2) discoverModules(key string, init bool) {
	blog.Infof("discover watch key %s start...", key)
	r.Lock()
	r.watchedKey[key] = struct{}{}
	r.Unlock()
	event, err := r.rd.DiscoverService(key)
	if err != nil {
		blog.Error("fail to register discover for api. err:%s", err.Error())
		r.cancel()
		os.Exit(1)
	}

	index := 0
	for {
		index++
		select {
		case eve := <-event:
			if eve.Err != nil {
				blog.V(3).Infof("discover zk key %s error %s", key, eve.Err.Error())
				time.Sleep(time.Second)
				go r.discoverModules(key, false)
				return
			}
			if index == 1 && init {
				blog.V(3).Infof("the init watch key %s event, then ignore", key)
			} else {
				_ = r.discoverEndpoints(eve.Key)
			}

		case <-r.rootCxt.Done():
			blog.V(3).Infof("zk register path %s and discover done", key)
			return
		}
	}
}

// Stop the DiscoveryV2
func (r *DiscoveryV2) Stop() {
	r.cancel()
	_ = r.rd.Stop()
}
