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
	"sync"
	"time"

	"bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/types"

	"golang.org/x/net/context"
)

//serviceDiscovery  discover bcs mesos module, examples: bcs-scheduler、bcs-mesos-driver ...
type serviceDiscovery struct {
	sync.RWMutex

	rd     *RegisterDiscover.RegDiscover
	cliTls *tls.Config

	rootCxt context.Context
	cancel  context.CancelFunc

	servers      map[string][]interface{} //key=BCS_MODULE_K8SAPISERVER...; value=[]*types.BcsK8sApiserverInfo...
	events       chan *RegisterDiscover.DiscoverEvent
	eventHandler EventHandleFunc

	k8sapiClustersNodes   map[string]struct{}
	mesosapiClustersNodes map[string]struct{}
}

//NewserviceDiscovery create a object of serviceDiscovery
func NewServiceDiscovery(zkserv string) (ModuleDiscovery, error) {
	rd := &serviceDiscovery{
		rd: RegisterDiscover.NewRegDiscoverEx(zkserv, 10*time.Second),
		servers: map[string][]interface{}{
			types.BCS_MODULE_STORAGE:        make([]interface{}, 0),
			types.BCS_MODULE_NETSERVICE:     make([]interface{}, 0),
			types.BCS_MODULE_MESOSAPISERVER: make([]interface{}, 0),
			types.BCS_MODULE_K8SAPISERVER:   make([]interface{}, 0),
			types.BCS_MODULE_METRICSERVICE:  make([]interface{}, 0),
			types.BCS_MODULE_APISERVER:      make([]interface{}, 0),
		},
		events:                make(chan *RegisterDiscover.DiscoverEvent, 1024),
		k8sapiClustersNodes:   make(map[string]struct{}),
		mesosapiClustersNodes: make(map[string]struct{}),
	}

	err := rd.start()
	if err != nil {
		return nil, err
	}

	return rd, nil
}

// module: types.BCS_MODULE_SCHEDULER...
// list all servers
//if mesos-apiserver/k8s-apiserver module={module}/clusterid, for examples: mesosdriver/BCS-TESTBCSTEST01-10001
func (r *serviceDiscovery) GetModuleServers(moduleName string) ([]interface{}, error) {
	r.RLock()
	defer r.RUnlock()

	servs, ok := r.servers[moduleName]
	if !ok {
		return nil, fmt.Errorf("Module %s is invalid", moduleName)
	}

	if len(servs) == 0 {
		return nil, fmt.Errorf("Module %s have no servers endpoints", moduleName)
	}

	return servs, nil
}

// get random one server
func (r *serviceDiscovery) GetRandModuleServer(moduleName string) (interface{}, error) {
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
func (r *serviceDiscovery) RegisterEventFunc(handleFunc EventHandleFunc) {
	r.eventHandler = handleFunc
}

//Start the serviceDiscovery
func (r *serviceDiscovery) start() error {
	//create root context
	r.rootCxt, r.cancel = context.WithCancel(context.Background())

	//start regdiscover
	if err := r.rd.Start(); err != nil {
		blog.Error("fail to start register and discover serv. err:%s", err.Error())
		return err
	}
	zvs, err := r.rd.DiscoverNodes("/bcs/services/endpoints")
	if err != nil {
		blog.Errorf("discover nodes /bcs/services/endpoints error %s", err.Error())
		return err
	}
	by, _ := json.Marshal(zvs)
	blog.Infof("servers(%s)", string(by))
	//discover other bcs service
	for k := range r.servers {
		go r.discoverModules(k)
	}

	return nil
}

func (r *serviceDiscovery) discoverModules(k string) {
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
				r.discoverMesosApiserver(eve.Nodes)
			// netservice
			case types.BCS_MODULE_STORAGE:
				r.discoverStorageServ(eve.Server)
			// metric service
			case types.BCS_MODULE_NETSERVICE:
				r.discoverNetserviceServ(eve.Server)
			case types.BCS_MODULE_K8SAPISERVER:
				r.discoverK8sApiserver(eve.Nodes)
			case types.BCS_MODULE_METRICSERVICE:
				r.discoverMetricServer(eve.Server)
			case types.BCS_MODULE_APISERVER:
				r.discoverApiserver(eve.Server)
			}

		case <-r.rootCxt.Done():
			blog.Warn("zk register path %s and discover done", key)
			return
		}
	}
}

//Stop the serviceDiscovery
func (r *serviceDiscovery) stop() error {
	r.cancel()
	r.rd.Stop()
	return nil
}

func (r *serviceDiscovery) discoverMesosApiserver(nodes []string) error {

	for _, node := range nodes {
		_, ok := r.mesosapiClustersNodes[node]
		if ok {
			continue
		}

		r.mesosapiClustersNodes[node] = struct{}{}
		blog.V(3).Infof("start discover cluster %s mesosapi", node)
		key := fmt.Sprintf("%s/%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_MESOSAPISERVER, node)

		go r.discoverClusterMesosApiserver(key)
	}

	return nil
}

func (r *serviceDiscovery) discoverClusterMesosApiserver(key string) {
	event, err := r.rd.DiscoverService(key)
	if err != nil {
		blog.Error("fail to discover service %s err:%s", key, err.Error())
		time.Sleep(time.Second)
		go r.discoverClusterMesosApiserver(key)
		return
	}

	for {
		select {
		case eve := <-event:
			if eve.Err != nil {
				blog.Errorf("discover zk key %s error %s", key, err.Error())
				time.Sleep(time.Second)
				go r.discoverClusterMesosApiserver(key)
				return
			}

			mesosApis := make([]interface{}, 0)
			for _, serv := range eve.Server {
				blog.V(3).Infof("discover key %s mesos apiserver %s", key, serv)

				api := new(types.BcsMesosApiserverInfo)
				if err := json.Unmarshal([]byte(serv), api); err != nil {
					blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
					continue
				}

				mesosApis = append(mesosApis, api)
			}

			r.Lock()
			r.servers[fmt.Sprintf("%s/%s", types.BCS_MODULE_MESOSAPISERVER, path.Base(key))] = mesosApis
			r.Unlock()
			if r.eventHandler != nil {
				r.eventHandler(fmt.Sprintf("%s/%s", types.BCS_MODULE_MESOSAPISERVER, path.Base(key)))
			}
		case <-r.rootCxt.Done():
			blog.Warn("zk register path %s and discover done", key)
			return
		}
	}
}

func (r *serviceDiscovery) discoverK8sApiserver(nodes []string) error {

	for _, node := range nodes {
		_, ok := r.k8sapiClustersNodes[node]
		if ok {
			continue
		}

		r.k8sapiClustersNodes[node] = struct{}{}
		blog.V(3).Infof("start discover cluster %s k8sapi", node)
		key := fmt.Sprintf("%s/%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_K8SAPISERVER, node)

		go r.discoverClusterK8sApiserver(key)
	}

	return nil
}

func (r *serviceDiscovery) discoverClusterK8sApiserver(key string) {
	event, err := r.rd.DiscoverService(key)
	if err != nil {
		blog.Error("fail to discover service %s err:%s", key, err.Error())
		time.Sleep(time.Second)
		go r.discoverClusterK8sApiserver(key)
		return
	}

	for {
		select {
		case eve := <-event:
			if eve.Err != nil {
				blog.Errorf("discover zk key %s error %s", key, err.Error())
				time.Sleep(time.Second)
				go r.discoverClusterK8sApiserver(key)
				return
			}

			k8sApis := make([]interface{}, 0)
			for _, serv := range eve.Server {
				blog.V(3).Infof("discover cluster %s k8s apiserver %s", key, serv)

				api := new(types.BcsK8sApiserverInfo)
				if err := json.Unmarshal([]byte(serv), api); err != nil {
					blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
					continue
				}

				k8sApis = append(k8sApis, api)
			}

			r.Lock()
			r.servers[fmt.Sprintf("%s/%s", types.BCS_MODULE_K8SAPISERVER, path.Base(key))] = k8sApis
			r.Unlock()
			if r.eventHandler != nil {
				r.eventHandler(fmt.Sprintf("%s/%s", types.BCS_MODULE_K8SAPISERVER, path.Base(key)))
			}
		case <-r.rootCxt.Done():
			blog.Warn("zk register path %s and discover done", key)
			return
		}
	}
}

func (r *serviceDiscovery) discoverStorageServ(servInfos []string) error {
	blog.V(3).Infof("discover storage(%v)", servInfos)

	storages := make([]interface{}, 0)
	for _, serv := range servInfos {
		storage := new(types.BcsStorageInfo)
		if err := json.Unmarshal([]byte(serv), storage); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		storages = append(storages, storage)
	}

	r.Lock()
	r.servers[types.BCS_MODULE_STORAGE] = storages
	r.Unlock()
	if r.eventHandler != nil {
		r.eventHandler(types.BCS_MODULE_STORAGE)
	}

	return nil
}

func (r *serviceDiscovery) discoverNetserviceServ(servInfos []string) error {
	blog.V(3).Infof("discover netservice(%v)", servInfos)

	netservices := make([]interface{}, 0)
	for _, serv := range servInfos {
		netservice := new(types.NetServiceInfo)
		if err := json.Unmarshal([]byte(serv), netservice); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		netservices = append(netservices, netservice)
	}

	r.Lock()
	r.servers[types.BCS_MODULE_NETSERVICE] = netservices
	r.Unlock()
	if r.eventHandler != nil {
		r.eventHandler(types.BCS_MODULE_NETSERVICE)
	}

	return nil
}

func (r *serviceDiscovery) discoverMetricServer(servInfos []string) error {
	blog.V(3).Infof("discover metricservice(%v)", servInfos)

	clusters := make([]interface{}, 0)
	for _, serv := range servInfos {
		cluster := new(types.MetricServiceInfo)
		if err := json.Unmarshal([]byte(serv), cluster); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		clusters = append(clusters, cluster)
	}

	r.Lock()
	r.servers[types.BCS_MODULE_METRICSERVICE] = clusters
	r.Unlock()
	if r.eventHandler != nil {
		r.eventHandler(types.BCS_MODULE_METRICSERVICE)
	}

	return nil
}

func (r *serviceDiscovery) discoverApiserver(servInfos []string) error {
	blog.V(3).Infof("discover apiserver(%v)", servInfos)

	clusters := make([]interface{}, 0)
	for _, serv := range servInfos {
		cluster := new(types.APIServInfo)
		if err := json.Unmarshal([]byte(serv), cluster); err != nil {
			blog.Warn("fail to do json unmarshal(%s), err:%s", serv, err.Error())
			continue
		}

		clusters = append(clusters, cluster)
	}

	r.Lock()
	r.servers[types.BCS_MODULE_APISERVER] = clusters
	r.Unlock()
	if r.eventHandler != nil {
		r.eventHandler(types.BCS_MODULE_APISERVER)
	}

	return nil
}
