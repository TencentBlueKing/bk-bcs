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
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/types"

	"golang.org/x/net/context"
)

//mesosDiscovery  discover bcs mesos module, examples: bcs-scheduler、bcs-mesos-driver ...
type mesosDiscovery struct {
	sync.RWMutex

	rd     *RegisterDiscover.RegDiscover
	cliTls *tls.Config

	rootCxt context.Context
	cancel  context.CancelFunc

	servers      map[string][]interface{} //key=BCS_MODULE_K8SAPISERVER...; value=[]*types.BcsK8sApiserverInfo...
	events       chan *RegisterDiscover.DiscoverEvent
	eventHandler EventHandleFunc

	loadbalanceGroupNodes map[string]struct{}
}

//NewmesosDiscovery create a object of mesosDiscovery
func NewMesosDiscovery(zkserv string) (ModuleDiscovery, error) {
	rd := &mesosDiscovery{
		rd: RegisterDiscover.NewRegDiscoverEx(zkserv, 10*time.Second),
		servers: map[string][]interface{}{
			types.BCS_MODULE_MESOSAPISERVER: make([]interface{}, 0),
			types.BCS_MODULE_SCHEDULER:      make([]interface{}, 0),
			types.BCS_MODULE_MESOSDATAWATCH: make([]interface{}, 0),
			types.BCS_MODULE_DNS:            make([]interface{}, 0),
			types.BCS_MODULE_LOADBALANCE:    make([]interface{}, 0),
		},
		events:                make(chan *RegisterDiscover.DiscoverEvent, 1024),
		loadbalanceGroupNodes: make(map[string]struct{}),
	}

	err := rd.start()
	if err != nil {
		return nil, err
	}

	return rd, nil
}

// module: types.BCS_MODULE_SCHEDULER...
// list all servers
func (r *mesosDiscovery) GetModuleServers(moduleName string) ([]interface{}, error) {
	r.RLock()
	defer r.RUnlock()

	var servs []interface{}
	for k, v := range r.servers {
		if strings.Contains(k, moduleName) {
			servs = append(servs, v...)
		}
	}

	if len(servs) == 0 {
		return nil, fmt.Errorf("Module %s have no servers endpoints", moduleName)
	}

	return servs, nil
}

//input: types.BCS_MODULE_SCHEDULER...
// get random one server
func (r *mesosDiscovery) GetRandModuleServer(moduleName string) (interface{}, error) {
	r.RLock()
	defer r.RUnlock()

	servs, ok := r.servers[moduleName]
	if !ok {
		return nil, fmt.Errorf("Module %s is invalid", moduleName)
	}

	if len(servs) == 0 {
		return nil, fmt.Errorf("Module %s have no servers endpoints", moduleName)
	}

	//rand
	rand.Seed(int64(time.Now().Nanosecond()))
	serv := servs[rand.Intn(len(servs))]
	return serv, nil
}

// register event handle function
func (r *mesosDiscovery) RegisterEventFunc(handleFunc EventHandleFunc) {
	r.eventHandler = handleFunc
}

//Start the mesosDiscovery
func (r *mesosDiscovery) start() error {
	//create root context
	r.rootCxt, r.cancel = context.WithCancel(context.Background())

	blog.V(3).Infof("mesosDiscovery start regdiscover...")
	//start regdiscover
	if err := r.rd.Start(); err != nil {
		blog.Error("fail to start register and discover serv. err:%s", err.Error())
		return err
	}
	blog.V(3).Infof("mesosDiscovery start regdiscover success")

	//discover other bcs service
	for k := range r.servers {
		go r.discoverModules(k)
	}

	return nil
}

func (r *mesosDiscovery) discoverModules(k string) {
	key := fmt.Sprintf("%s/%s", types.BCS_SERV_BASEPATH, k)
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
				blog.Errorf("discover zk key %s error %s", key, eve.Err.Error())
				time.Sleep(time.Second)
				go r.discoverModules(k)
				return
			}

			switch path.Base(eve.Key) {
			// mesos apiserver
			case types.BCS_MODULE_MESOSAPISERVER:
				r.discoverMesosdriver(eve.Server)
			// netservice
			case types.BCS_MODULE_SCHEDULER:
				r.discoverScheduler(eve.Server)
			// metric service
			case types.BCS_MODULE_MESOSDATAWATCH:
				r.discoverDatawatch(eve.Server)

			case types.BCS_MODULE_DNS:
				r.discoverDns(eve.Server)

			case types.BCS_MODULE_LOADBALANCE:
				r.discoverLoadbalance(eve.Nodes)
			}

		case <-r.rootCxt.Done():
			blog.Warn("zk register path %s and discover done", key)
			return
		}
	}
}

//Stop the mesosDiscovery
func (r *mesosDiscovery) stop() error {
	r.cancel()
	r.rd.Stop()
	return nil
}

func (r *mesosDiscovery) discoverMesosdriver(servInfos []string) error {
	blog.Info("discover mesos-driver(%v)", servInfos)

	drivers := make([]interface{}, 0)
	for _, serv := range servInfos {
		driver := new(types.BcsMesosApiserverInfo)
		if err := json.Unmarshal([]byte(serv), &driver); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		drivers = append(drivers, driver)
	}

	r.Lock()
	r.servers[types.BCS_MODULE_MESOSAPISERVER] = drivers
	r.Unlock()
	r.eventHandler(types.BCS_MODULE_MESOSAPISERVER)

	return nil
}

func (r *mesosDiscovery) discoverScheduler(servInfos []string) error {
	blog.Info("discover scheduler(%v)", servInfos)

	schedulers := make([]interface{}, 0)
	for _, serv := range servInfos {
		scheduler := new(types.SchedulerServInfo)
		if err := json.Unmarshal([]byte(serv), &scheduler); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		schedulers = append(schedulers, scheduler)
	}

	r.Lock()
	r.servers[types.BCS_MODULE_SCHEDULER] = schedulers
	r.Unlock()
	r.eventHandler(types.BCS_MODULE_SCHEDULER)

	return nil
}

func (r *mesosDiscovery) discoverDatawatch(servInfos []string) error {
	blog.Info("discover data-watch(%v)", servInfos)

	watches := make([]interface{}, 0)
	for _, serv := range servInfos {
		watch := new(types.MesosDataWatchServInfo)
		if err := json.Unmarshal([]byte(serv), &watch); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		watches = append(watches, watch)
	}

	r.Lock()
	r.servers[types.BCS_MODULE_MESOSDATAWATCH] = watches
	r.Unlock()
	r.eventHandler(types.BCS_MODULE_MESOSDATAWATCH)

	return nil
}

func (r *mesosDiscovery) discoverDns(servInfos []string) error {
	blog.Info("discover dns(%v)", servInfos)

	dnses := make([]interface{}, 0)
	for _, serv := range servInfos {
		dns := new(types.DNSInfo)
		if err := json.Unmarshal([]byte(serv), &dns); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		dnses = append(dnses, dns)
	}

	r.Lock()
	r.servers[types.BCS_MODULE_DNS] = dnses
	r.Unlock()
	r.eventHandler(types.BCS_MODULE_DNS)

	return nil
}

func (r *mesosDiscovery) discoverLoadbalance(nodes []string) error {

	for _, node := range nodes {
		_, ok := r.loadbalanceGroupNodes[node]
		if ok {
			continue
		}

		r.loadbalanceGroupNodes[node] = struct{}{}
		blog.V(3).Infof("start discover group %s loabalance", node)
		key := fmt.Sprintf("%s/%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_LOADBALANCE, node)

		go r.discoverGroupLoadbalance(key)
	}

	return nil
}

func (r *mesosDiscovery) discoverGroupLoadbalance(key string) {
	event, err := r.rd.DiscoverService(key)
	if err != nil {
		blog.Error("fail to discover service %s err:%s", key, err.Error())
		time.Sleep(time.Second)
		go r.discoverGroupLoadbalance(key)
		return
	}

	for {
		select {
		case eve := <-event:
			if eve.Err != nil {
				blog.Errorf("discover zk key %s error %s", key, err.Error())
				time.Sleep(time.Second)
				go r.discoverGroupLoadbalance(key)
				return
			}

			lbs := make([]interface{}, 0)
			for _, serv := range eve.Server {
				blog.V(3).Infof("discover key %s mesos apiserver %s", key, serv)

				lb := new(types.LoadBalanceInfo)
				if err := json.Unmarshal([]byte(serv), lb); err != nil {
					blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
					continue
				}

				lbs = append(lbs, lb)
			}

			r.Lock()
			r.servers[fmt.Sprintf("%s/%s", types.BCS_MODULE_LOADBALANCE, path.Base(key))] = lbs
			r.Unlock()
			if r.eventHandler != nil {
				r.eventHandler(types.BCS_MODULE_LOADBALANCE)
			}
		case <-r.rootCxt.Done():
			blog.Warn("zk register path %s and discover done", key)
			return
		}
	}
}
