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

package app

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/cache"
	loadbalance "github.com/Tencent/bk-bcs/bcs-common/pkg/loadbalance/v2"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/option"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/util"

	"github.com/samuel/go-zookeeper/zk"
)

//DataReflector interface for unit test injection
type DataReflector interface {
	Lister() (http, https types.HTTPServiceInfoList, tcp, udp types.FourLayerServiceInfoList)
	Start() error
	Stop()
}

//ExportServiceKeyFunc key function format Service uniq key
func ExportServiceKeyFunc(obj interface{}) (string, error) {
	svr, ok := obj.(*loadbalance.ExportService)
	if !ok {
		return "", fmt.Errorf("ExportService type Assert failed")
	}
	return svr.Namespace + "." + svr.ServiceName, nil
}

//CheckBCSGroup check if local is in groups
func CheckBCSGroup(local string, svr *loadbalance.ExportService) bool {
	//done(developer): protect group info lost
	if svr.BCSGroup == nil || len(svr.BCSGroup) == 0 {
		errData, err := json.Marshal(svr)
		if err != nil {
			blog.Errorf("Reflector json encode err data failed, %s", err.Error())
			return false
		}
		blog.Errorf("Reflector got no group data. skip, serviceName: %s, original data: %s", svr.ServiceName, errData)
		return false
	}
	for _, item := range svr.BCSGroup {
		if item == local {
			return true
		}
	}
	return false
}

//GetName get real service name from zookeeper node name
func GetName(node string) string {
	index := strings.Index(node, ".")
	if index == -1 {
		return node
	}
	return node[index+1:]
}

//NewReflector create new ServiceReflector
func NewReflector(config *option.LBConfig, handler EventHandler) *ServiceReflector {
	hosts := strings.Split(config.Zookeeper, ",")
	return &ServiceReflector{
		dataCache:    cache.NewCache(ExportServiceKeyFunc),
		eventHanlder: handler,
		watchPath:    config.WatchPath,
		syncPeriod:   config.SyncPeriod,
		zkHosts:      hosts,
		zkConnFlag:   false,
		cfg:          config,
		exit:         make(chan struct{}),
	}
}

//ServiceReflector handling/holding service data coming from zookeeper
//reflector will do:
//1. sync all zookeeper data in period, default 30 seconds
//2. watch children of data path
//3. watch all children data of path
//4. cache all zookeeper data, notify EventHandler when data changed
//reflector will running in other goroutine,
type ServiceReflector struct {
	dataCache    cache.Store      //data cache for all service
	eventHanlder EventHandler     //callback handlers when data changed
	watchPath    string           //zk watch path
	syncPeriod   int              //period for sync all data
	zkHosts      []string         //zk host info
	zkConn       ZkClient         //zk client connection
	zkConnFlag   bool             //flag for zk reconnection
	cfg          *option.LBConfig //config item
	exit         chan struct{}    //exit flags
}

//Lister classify all service to 3 types:
//http service, https service and tcp service
func (reflector *ServiceReflector) Lister() (http, https types.HTTPServiceInfoList, tcp, udp types.FourLayerServiceInfoList) {
	//Get all data from cache
	exportList := reflector.dataCache.List()
	if len(exportList) == 0 {
		return
	}
	exportServiceList := make([]*loadbalance.ExportService, 0, len(exportList))
	for _, item := range exportList {
		tmp, ok := item.(*loadbalance.ExportService)
		if !ok {
			debugData, err := json.Marshal(item)
			if err != nil {
				blog.Errorf("Reflector parse cache data to json err: %s", err.Error())
				continue
			}
			blog.Errorf("Reflector got unsupport data type. detail: %s", string(debugData))
			continue
		}
		exportServiceList = append(exportServiceList, tmp)
	}

	http, https, tcp, udp = reflector.listData(exportServiceList)

	http.SortBackends()
	https.SortBackends()
	sort.Sort(http)
	sort.Sort(https)
	sort.Sort(tcp)
	sort.Sort(udp)
	return
}

func caculateBackendWeight(portInfo loadbalance.ExportPort, svr *loadbalance.ExportService) map[string]int {
	weightCopy := make(map[string]int)
	//weight only affects when backend coming from different POD
	backendCounts := make(map[string]int)
	for key := range svr.ServiceWeight {
		backendCounts[key] = 0
		weightCopy[key] = svr.ServiceWeight[key]
	}
	for _, backend := range portInfo.Backends {
		if backend.Label != nil && len(backend.Label) > 0 {
			if value, ok := backendCounts[backend.Label[0]]; ok {
				value++
				backendCounts[backend.Label[0]] = value
			}
		} else {
			blog.Warnf(
				"Backend %s/%d in Service %s/%s with protocol %s/%d lost label info",
				backend.TargetIP,
				backend.TargetPort,
				svr.Namespace,
				svr.ServiceName,
				portInfo.Protocol,
				portInfo.ServicePort,
			)
		}
	}
	//all backend weights
	for key := range svr.ServiceWeight {
		if backendCounts[key] == 0 {
			backendCounts[key] = 1
		}
		weightCopy[key] = int(math.Ceil(float64(svr.ServiceWeight[key]) / float64(backendCounts[key])))
	}
	return weightCopy
}

func convertPortBackends(portInfo loadbalance.ExportPort, svr *loadbalance.ExportService) types.BackendList {
	//weighCopy only affective when svr.ServiceWeight exists
	var weightCopy map[string]int
	//calculate backend weight for all nodes if needed
	if svr.ServiceWeight != nil && len(svr.ServiceWeight) > 1 {
		weightCopy = caculateBackendWeight(portInfo, svr)
	}
	var backends types.BackendList
	//backend
	for _, bk := range portInfo.Backends {
		if bk.TargetIP == "" || bk.TargetPort == 0 {
			blog.Errorf("Reflector got empty backend ip/port for %s/%s in service port %d", svr.Namespace, svr.ServiceName, portInfo.ServicePort)
			continue
		}
		var backend types.Backend
		backend.IP = bk.TargetIP
		backend.Port = bk.TargetPort
		backend.Host = svr.ServiceName + "_" + backend.String()
		if svr.ServiceWeight != nil && bk.Label != nil && len(bk.Label) > 0 {
			if value, ok := weightCopy[bk.Label[0]]; ok {
				backend.Weight = value
			} else {
				backend.Weight = 1
			}
		} else {
			backend.Weight = 1
		}
		backends = append(backends, backend)
	}
	if len(backends) != 0 {
		sort.Sort(backends)
	}
	return backends
}

func (reflector *ServiceReflector) listData(exportServiceList []*loadbalance.ExportService) (
	http, https types.HTTPServiceInfoList, tcp, udp types.FourLayerServiceInfoList) {
	for _, svr := range exportServiceList {
		if !CheckBCSGroup(reflector.cfg.Group, svr) {
			blog.Debug(fmt.Sprintf("Local Group %s, ExportService Group %s, skip service %s/%s", reflector.cfg.Group, svr.BCSGroup[0], svr.Namespace, svr.ServiceName))
			continue
		}
		//default balance
		if len(svr.Balance) == 0 {
			svr.Balance = "roundrobin"
		}
		for _, portInfo := range svr.ServicePort {
			//skip when no backends
			if portInfo.Backends == nil || len(portInfo.Backends) == 0 || portInfo.ServicePort == 0 {
				blog.Warnf("Get no backends in Service %s/%s in %s/%d", svr.Namespace, svr.ServiceName, portInfo.Protocol, portInfo.ServicePort)
				continue
			}

			backends := convertPortBackends(portInfo, svr)
			//protect empty backend list
			if len(backends) == 0 {
				blog.Warnf("Reflector got no backend for %s/%s in service port %d, discard data", svr.Namespace, svr.ServiceName, portInfo.ServicePort)
				continue
			}

			var srvInfo types.ServiceInfo
			validPath := util.TrimSpecialChar(portInfo.Path)
			if validPath == "" {
				srvInfo.Name = svr.ServiceName + "_" + strconv.Itoa(portInfo.ServicePort)
			} else {
				srvInfo.Name = svr.ServiceName + "_" + strconv.Itoa(portInfo.ServicePort) + "_" + validPath
			}
			srvInfo.ServicePort = portInfo.ServicePort
			srvInfo.Balance = svr.Balance
			srvInfo.MaxConn = 50000

			protocol := strings.ToLower(portInfo.Protocol)
			//done(developer): protect empty vhost for http(s)
			if (protocol == "http" || protocol == "https") && portInfo.BCSVHost == "" {
				blog.Errorf("http/s Service %s/%s/%s in port %d lost VHost info in export, discard.", svr.Namespace, svr.ServiceName, portInfo.Name, portInfo.ServicePort)
				continue
			}
			if protocol == "http" || protocol == "https" {
				httpSvrInfo := types.NewHTTPServiceInfo(srvInfo, portInfo.BCSVHost)
				if validPath == "" {
					httpSvrInfo.Name = svr.ServiceName + "_" + strconv.Itoa(portInfo.ServicePort)
				} else {
					httpSvrInfo.Name = svr.ServiceName + "_" + strconv.Itoa(portInfo.ServicePort) + "_" + validPath
				}
				var httpBackend types.HTTPBackend
				//预防HTTPS重定向HTTP或者HTTP重定向HTTPS，或者CLB把http改成https对外
				if portInfo.ServicePort != 80 && portInfo.ServicePort != 443 {
					// portInfo.Path 只是域名，不包含端口，当端口不为80或者443时，haproxy转发会有问题
					// nginx，clb转发不需要该端口信息，需要清理
					httpSvrInfo.BCSVHost = httpSvrInfo.BCSVHost + ":" + strconv.Itoa(portInfo.ServicePort)
				}
				httpBackend.Path = portInfo.Path
				if portInfo.Path == "" {
					httpBackend.Path = "/"
				}
				httpBackend.UpstreamName = httpSvrInfo.Name
				httpBackend.BackendList = backends
				httpSvrInfo.Backends = append(httpSvrInfo.Backends, httpBackend)
				if protocol == "http" {
					http.AddItem(httpSvrInfo)
				} else {
					https.AddItem(httpSvrInfo)
				}

			} else if protocol == "udp" {
				//udp
				udpSrvInfo := types.NewFourLayerServiceInfo(srvInfo, backends)
				udp = append(udp, udpSrvInfo)
			} else if protocol == "tcp" {
				//tcp
				srvInfo.SessionAffinity = true
				tcpSrvInfo := types.NewFourLayerServiceInfo(srvInfo, backends)
				tcp = append(tcp, tcpSrvInfo)
			} else {
				blog.Warnf("Get unknown protocol %s", portInfo.Protocol)
			}
		} //end of ServicePort
	}
	return
}

//Start reflector to syncing data from data source to local cache
//1. starting connection for zookeeper, watch cluster path, cluster children node data
//2. starting goroutine list all cluster service period
//3. starting channel receiving control events from upper app
func (reflector *ServiceReflector) Start() error {
	//step1: zkInit, all zookeeper event setting
	if err := reflector.zkInit(); err != nil {
		LoadbalanceZookeeperStateMetric.WithLabelValues(reflector.cfg.Name).Set(0)
		blog.Errorf("zookeeper init failed: %s", err.Error())
		return err
	}
	//step2
	go reflector.run()

	//TODO: step3
	return nil
}

//zkInit init zookeeper connection
func (reflector *ServiceReflector) zkInit() error {
	blog.Infof("Reflector init zookeeper connection with %s", reflector.cfg.Zookeeper)
	var conErr error
	//reflector.zkConn, _, conErr = zk.Connect(reflector.zkHosts, time.Second*5)
	reflector.zkConn, conErr = NewAdapterZkClient(reflector.zkHosts, time.Second*5)
	if conErr != nil {
		return conErr
	}
	blog.Infof("Connect to zookeeper %s success", reflector.cfg.Zookeeper)
	if err := reflector.watchClusterPath(); err != nil {
		return err
	}
	blog.Infof("Watch cluster path %s success", reflector.cfg.WatchPath)
	return nil
}

//watchServicePath Get all data first time, create zookeeper watch
func (reflector *ServiceReflector) watchClusterPath() error {
	if len(reflector.watchPath) == 0 {
		blog.Errorf("Cluster path is empty")
		return fmt.Errorf("Cluster path is empty")
	}
	//check path existence in zookeeper
	ok, _, err := reflector.zkConn.Exists(reflector.watchPath)
	if err != nil {
		blog.Errorf("Error happen when check cluster path existence: %s", err.Error())
		return err
	}
	if !ok {
		//node do not exist, create exist watch waiting node creation
		go func() {
			blog.Warnf("Cluster node %s do not exist, create existence watch", reflector.watchPath)
			_, _, existEvent, existErr := reflector.zkConn.ExistsW(reflector.watchPath)
			if existErr != nil {
				blog.Errorf("Create exist watch for %s error: %s", reflector.watchPath, existErr.Error())
				return
			}
			select {
			case <-reflector.exit:
				blog.Info("reflector exit in Cluster node ExistWatch")
				return
			case <-existEvent:
				//TODO: check if Event is error
				//cluster node created, ready to watch
				go reflector.childrenNodeWatch(reflector.watchPath)
				return
			}
		}()
	} else {
		//go watch children, wait for new children
		go reflector.childrenNodeWatch(reflector.watchPath)
	}
	return nil
}

//listChildrenNode list all children node designated
func (reflector *ServiceReflector) listChildrenNode(node string) []string {
	children, _, err := reflector.zkConn.Children(reflector.watchPath)
	if err != nil {
		LoadbalanceZookeeperStateMetric.WithLabelValues(reflector.cfg.Name).Set(0)
		blog.Errorf("reflector get cluster path children error: %s", err.Error())
		return nil
	}
	if len(children) == 0 {
		blog.Infof("Get no service children node from cluster path: %s", reflector.watchPath)
	}
	LoadbalanceZookeeperStateMetric.WithLabelValues(reflector.cfg.Name).Set(1)
	return children
}

//addNewNodeFromList check nodeList node, Get node data and then watch changed
func (reflector *ServiceReflector) addServiceNodeFromList(nodeList []string) {
	if len(nodeList) == 0 {
		blog.Warn("Node list is empty, no need to watch")
		return
	}
	for _, node := range nodeList {
		_, exist, err := reflector.dataCache.GetByKey(node)
		if err != nil {
			blog.Warnf("get data from cache by key %s failed, err %s", node, err.Error())
		}
		if exist {
			continue
		}
		//handle new node, create watch
		blog.Infof("New service %s found, ready to watch", node)
		go reflector.dataNodeWatch(node)
	}
}

//deleteServiceNode check nodeList node, Get node data and then watch changed
func (reflector *ServiceReflector) deleteServiceNode(node string) {
	data, exist, err := reflector.dataCache.GetByKey(node)
	if err != nil {
		blog.Warnf("get data from cache by key %s failed, err %s", node, err.Error())
	}
	if exist {
		delErr := reflector.dataCache.Delete(data)
		blog.Infof("Delete Service %s in local cache, delete ret: %+v", node, delErr)
		reflector.eventHanlder.OnDelete(data)
	}
}

//updateServiceNode check nodeList node, Get node data and then watch changed
func (reflector *ServiceReflector) updateServiceNode(node string) {
	serviceNode := filepath.Join(reflector.watchPath, node)
	data, _, err := reflector.zkConn.Get(serviceNode)
	if err != nil {
		blog.Errorf("Read %s data for Update failed: %s", node, err.Error())
		return
	}
	exportSvr := loadbalance.NewPtrExportService()
	if jsonErr := json.Unmarshal(data, exportSvr); jsonErr != nil {
		blog.Errorf("Decode %s json data failed: %s, original str: %s",
			node, jsonErr.Error(), string(data))
		return
	}
	//push to cache
	old, exsit, err := reflector.dataCache.Get(exportSvr)
	if err != nil {
		blog.Warnf("get data %v from data cache failed, err %s", exportSvr, err.Error())
	}
	if exsit {
		err = reflector.dataCache.Update(exportSvr)
		if err != nil {
			blog.Warnf("update data cache failed, new data: %v, err %s", exportSvr, err.Error())
		}
		reflector.eventHanlder.OnUpdate(old, exportSvr)
		blog.Debug(fmt.Sprintf("Update service %s success in reflector", node))
	} else {
		err = reflector.dataCache.Add(exportSvr)
		if err != nil {
			blog.Warnf("add data cache failed, new data: %v, err %s", exportSvr, err.Error())
		}
		reflector.eventHanlder.OnAdd(exportSvr)
		blog.Warnf("Update service %s warnning. service data Lost in cache", node)
	}
}

//listServiceNode check nodeList node, Get node data and then watch changed
func (reflector *ServiceReflector) listServiceNode(node string) {
	serviceNode := filepath.Join(reflector.watchPath, node)
	data, _, err := reflector.zkConn.Get(serviceNode)
	if err != nil {
		blog.Errorf("Read %s data for Update failed: %s", node, err.Error())
		return
	}
	exportSvr := loadbalance.NewPtrExportService()
	if jsonErr := json.Unmarshal(data, exportSvr); jsonErr != nil {
		blog.Errorf("Decode %s json data failed: %s, original str: %s",
			node, jsonErr.Error(), string(data))
		return
	}
	//push to cache
	old, exsit, err := reflector.dataCache.Get(exportSvr)
	if err != nil {
		blog.Warnf("get data cache by %v failed, err %s", exportSvr, err.Error())
	}
	if exsit {
		err = reflector.dataCache.Update(exportSvr)
		if err != nil {
			blog.Warnf("update data cache failed, new data %v, err %s", exportSvr, err.Error())
		}
		reflector.eventHanlder.OnUpdate(old, exportSvr)
		blog.Warnf("List service %s warnning. service data exist before", node)
	} else {
		err = reflector.dataCache.Add(exportSvr)
		if err != nil {
			blog.Warnf("add data cache failed, new data %v, err %s", exportSvr, err.Error())
		}
		reflector.eventHanlder.OnAdd(exportSvr)
		blog.Infof("List service %s success.", node)
	}
}

//dataNodeWatch watch detail service data changed
func (reflector *ServiceReflector) childrenNodeWatch(node string) {
	//ready to watch cluster service node,
	children, stat, eventChan, err := reflector.zkConn.ChildrenW(node)
	if err != nil {
		blog.Errorf("Watch Node %s children failed: %s", node, err.Error())
		time.Sleep(5 * time.Second)
		go reflector.childrenNodeWatch(node)
		return
	} else if stat == nil {
		blog.Errorf("Wath node %s state return nil", node)
		return
	}
	reflector.addServiceNodeFromList(children)
	//wait watch event
	for {
		select {
		case <-reflector.exit:
			blog.Info("Watch %s children Event exit.", node)
			return
		case event := <-eventChan:
			//TODO: check if event is error
			if event.Type == zk.EventNodeChildrenChanged {
				//Node num changed, maybe add or delete, only handle
				//add event. delete event will be handle by node watch
				childrenList := reflector.listChildrenNode(node)
				if len(childrenList) == 0 {
					blog.Info("Node children event trigger, all children clear")
				} else {
					//iterator all children node, finding new nodes, add watch
					reflector.addServiceNodeFromList(childrenList)
				}
			}
			//create next watch for event
			blog.Info("Children watch trigger done, create next watch")
			go reflector.childrenNodeWatch(node)
			return
		}
	}
}

//dataNodeWatch watch detail service data changed
func (reflector *ServiceReflector) dataNodeWatch(node string) {
	//ready to watch cluster service node
	serviceNode := filepath.Join(reflector.watchPath, node)
	data, stat, eventChan, err := reflector.zkConn.GetW(serviceNode)
	if err != nil {
		blog.Errorf("Watch Service Node %s failed: %s", serviceNode, err.Error())
		time.Sleep(5 * time.Second)
		go reflector.dataNodeWatch(node)
		return
	} else if stat == nil {
		blog.Errorf("Wath service node %s state return nil", serviceNode)
		return
	}
	//data watch success, reading data
	exSvr := loadbalance.NewPtrExportService()
	if jsonErr := json.Unmarshal(data, exSvr); jsonErr != nil {
		//event json format error, we still watch node data changed event
		//maybe json data will repaire next time
		blog.Errorf("Decode Node %s json failed: %s, original str: %s",
			serviceNode, jsonErr.Error(), string(data))
	} else {
		//push to cache
		old, exsit, err := reflector.dataCache.Get(exSvr)
		if err != nil {
			blog.Warnf("get data %v from data cache failed, err %s", exSvr, err.Error())
		}
		if exsit {
			err = reflector.dataCache.Update(exSvr)
			if err != nil {
				blog.Warnf("update data cache failed, new data %v, err %s", exSvr, err.Error())
			}
			reflector.eventHanlder.OnUpdate(old, exSvr)
			blog.Infof("dataNodeWatch %s update service info", node)
		} else {
			err = reflector.dataCache.Add(exSvr)
			if err != nil {
				blog.Warnf("add data cache failed, new data %v, err %s", exSvr, err.Error())
			}
			reflector.eventHanlder.OnAdd(exSvr)
			blog.Infof("dataNodeWatch %s Add service info", node)
		}
	}
	//wait data node event
	for {
		select {
		case <-reflector.exit:
			blog.Info("Watch service node %s exit.", serviceNode)
			return
		case event := <-eventChan:
			//TODO: check if event is error
			if event.Type == zk.EventNodeDeleted {
				//clean delete node event
				blog.Infof("Service Node %s trigger delete event. No watch registered", serviceNode)
				reflector.deleteServiceNode(node)
			} else if event.Type == zk.EventNodeDataChanged {
				//create next watch for event, data will update in before watch event
				blog.Infof("service node %s data changed trigger done, create next watch", serviceNode)
				go reflector.dataNodeWatch(node)
			} else {
				//unknown event, add another watch
				blog.Errorf("unknown event %d happened in service node %s, create next watch", event.Type, serviceNode)
				go reflector.dataNodeWatch(node)
			}
			return
		}
	}
}

//run run ticker for sync all data in zookeeper to local cache
func (reflector *ServiceReflector) run() {
	tick := time.NewTicker(time.Second * time.Duration(reflector.syncPeriod))
	defer tick.Stop()
	blog.Infof("Entry Ticker to sync total data")
	for {
		select {
		case <-reflector.exit:
			LoadbalanceZookeeperStateMetric.WithLabelValues(reflector.cfg.Name).Set(0)
			blog.Infof("Ticker receive exit event.")
			return
		case <-tick.C:
			blog.Infof("Ticker trigger, ready to sync all service data")
			nodeList := reflector.listChildrenNode(reflector.watchPath)
			if len(nodeList) == 0 {
				blog.Info("No service node in zookeeper. wait for next ticker")
			} else {
				//when get all service data from zookeeper, we need to do:
				//1. find out extra data in local cache, delete this dirty data
				//2. update dirty data in local cache

				//step 1
				oldkeys := reflector.dataCache.ListKeys()
				//extraKey := util.GetSubsection(nodeList, oldkeys)
				extraKey := util.GetSubsection(oldkeys, nodeList)
				for _, key := range extraKey {
					blog.Warnf("Fix extra dirty service data [%s]", key)
					reflector.deleteServiceNode(key)
				}
				//step 2
				for _, node := range nodeList {
					reflector.updateServiceNode(node)
				}
			}
			//end case now
		}
	}
}

//Stop reflector, send stop event to all goroutines
func (reflector *ServiceReflector) Stop() {
	close(reflector.exit)
}
