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

package module_discovery

import (
	"crypto/tls"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/types"

	"golang.org/x/net/context"
)

//DiscoveryV2  discover bcs module, examples: bcs-api、bcs-scheduler、bcs-mesos-driver ...
type DiscoveryV2 struct {
	sync.RWMutex

	//discover bcs endpoints client
	rd     *RegisterDiscover.RegDiscover
	cliTls *tls.Config

	//context
	rootCxt context.Context
	cancel  context.CancelFunc

	//endpoint server infos
	//key = bcs-module or bcs-module/clusterid
	//example: bcs-api、bcs-storage、mesosdriver/BCS-MESOS-10000、kubernetedriver/BCS-K8S-15000
	//value = []byte array, []byte is marshal types.ServerInfo data
	servers map[string][]interface{}
	//register RegisterDiscover callback function
	//when endpoint node changed, call it
	eventHandler EventHandleFunc
	//if modules != nil
	//DiscoveryV2 only discover the specified modules
	//then lose sight of others
	modules []string
}

//NewDiscoveryV2 create a object of DiscoveryV2
func NewDiscoveryV2(zkserv string, modules []string) (ModuleDiscovery, error) {
	blog.Infof("DiscoveryV2 start...")
	rd := &DiscoveryV2{
		rd:      RegisterDiscover.NewRegDiscoverEx(zkserv, 10*time.Second),
		servers: make(map[string][]interface{}, 0),
		modules: modules,
	}

	err := rd.start()
	if err != nil {
		return nil, err
	}
	blog.Infof("DiscoveryV2 working...")
	return rd, nil
}

// module: types.BCS_MODULE_SCHEDULER...
// list all servers
//if mesos-apiserver/k8s-apiserver module={module}/clusterid, for examples: mesosdriver/BCS-TESTBCSTEST01-10001
func (r *DiscoveryV2) GetModuleServers(moduleName string) ([]interface{}, error) {
	r.RLock()
	defer r.RUnlock()

	servs, ok := r.servers[moduleName]
	if !ok {
		return nil, fmt.Errorf("Module %s not found", moduleName)
	}

	if len(servs) == 0 {
		return nil, fmt.Errorf("Module %s don't have endpoints", moduleName)
	}

	return servs, nil
}

// get random one server
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

	//rand
	rand.Seed(int64(time.Now().Nanosecond()))
	serv := servs[rand.Intn(len(servs))]
	return serv, nil
}

// register event handle function
func (r *DiscoveryV2) RegisterEventFunc(handleFunc EventHandleFunc) {
	r.eventHandler = handleFunc
}

//Start the DiscoveryV2
func (r *DiscoveryV2) start() error {
	//create root context
	r.rootCxt, r.cancel = context.WithCancel(context.Background())

	//start regdiscover
	if err := r.rd.Start(); err != nil {
		blog.Error("fail to start register and discover serv. err:%s", err.Error())
		return err
	}

	//discover all bcs module serviceinfos
	err := r.discoverEndpoints(types.BCS_SERV_BASEPATH)
	if err != nil {
		return err
	}

	//watch bcs module serviceinfo event
	go r.discoverModules(types.BCS_SERV_BASEPATH)
	return nil
}

//recursive discover bcs module serverinfo
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

	//get path children
	zvs, err := r.rd.DiscoverNodesV2(path)
	if err != nil {
		blog.V(3).Infof("discover nodes %s error %s", path, err.Error())
		return err
	}

	//if leaf node, then parse bcs module serverinfo
	if len(zvs.Server) != 0 {
		key := strings.TrimLeft(path, fmt.Sprintf("%s/", types.BCS_SERV_BASEPATH))
		val := make([]interface{}, 0)
		for _, v := range zvs.Server {
			val = append(val, v)
		}
		r.Lock()
		r.servers[key] = val
		r.Unlock()
		return nil
	}

	for _, v := range zvs.Nodes {
		err = r.discoverEndpoints(fmt.Sprintf("%s/%s", path, v))
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *DiscoveryV2) discoverModules(key string) {
	blog.V(3).Infof("start discover service key %s", key)
	event, err := r.rd.DiscoverService(key)
	if err != nil {
		blog.Error("fail to register discover for api. err:%s", err.Error())
		r.cancel()
		os.Exit(1)
	}

	for {
		select {
		case eve := <-event:
			if eve.Err != nil {
				blog.V(3).Infof("discover zk key %s error %s", key, eve.Err.Error())
				time.Sleep(time.Second)
				go r.discoverModules(key)
				return
			}
			r.discoverEndpoints(eve.Key)

		case <-r.rootCxt.Done():
			blog.V(3).Infof("zk register path %s and discover done", key)
			return
		}
	}
}

//Stop the DiscoveryV2
func (r *DiscoveryV2) stop() error {
	r.cancel()
	r.rd.Stop()
	return nil
}
