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

package controller

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-dns/plugin/bcsscheduler/metrics"
	bcsSchedulerUtil "github.com/Tencent/bk-bcs/bcs-services/bcs-dns/plugin/bcsscheduler/util"
	clientGoCache "k8s.io/client-go/tools/cache"

	"github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"
)

//ZkClient interface to define zk operation
//interface is only use for dependency injection
type ZkClient interface {
	Close()
	Get(path string) ([]byte, *zk.Stat, error)
	GetW(path string) ([]byte, *zk.Stat, <-chan zk.Event, error)
	Children(path string) ([]string, *zk.Stat, error)
	ChildrenW(path string) ([]string, *zk.Stat, <-chan zk.Event, error)
	Exists(path string) (bool, error)
	ExistsW(path string) (bool, *zk.Stat, <-chan zk.Event, error)
}

//wrapperClient wrapper for zk client in bcs-common.zkclient
type wrapperClient struct {
	client *zkclient.ZkClient
}

func (wrapper *wrapperClient) Close() {
	wrapper.client.Close()
}

func (wrapper *wrapperClient) Get(path string) ([]byte, *zk.Stat, error) {
	return wrapper.client.GetEx(path)
}

func (wrapper *wrapperClient) GetW(path string) ([]byte, *zk.Stat, <-chan zk.Event, error) {
	return wrapper.client.GetW(path)
}

func (wrapper *wrapperClient) Children(path string) ([]string, *zk.Stat, error) {
	return wrapper.client.GetChildrenEx(path)
}

func (wrapper *wrapperClient) ChildrenW(path string) ([]string, *zk.Stat, <-chan zk.Event, error) {
	return wrapper.client.ChildrenW(path)
}

func (wrapper *wrapperClient) Exists(path string) (bool, error) {
	return wrapper.client.Exist(path)
}

func (wrapper *wrapperClient) ExistsW(path string) (bool, *zk.Stat, <-chan zk.Event, error) {
	return wrapper.client.ExistW(path)
}

type ZkController struct {
	conCxt       context.Context                          //context for exit signal
	conCancel    context.CancelFunc                       //stop all goroutine
	resyncperiod int                                      //resync all data period
	resynced     bool                                     //status resynced
	zkHost       []string                                 //zk host iplist
	watchpath    string                                   //zookeeper watch path
	client       ZkClient                                 //zookeeper client
	storage      cache.Store                              //cache storage
	nsStorage    map[string]context.CancelFunc            //storage for all namespace watcher
	nsLock       sync.Mutex                               //lock for nsStorage
	decoder      bcsSchedulerUtil.Decoder                 //decoder for storage
	funcs        *clientGoCache.ResourceEventHandlerFuncs //funcs for callback
}

//NewZkController create controller according Store, Decoder end EventFuncs
func NewZkController(zkHost []string, path string, period int, cache cache.Store, decoder bcsSchedulerUtil.Decoder, eventFunc *clientGoCache.ResourceEventHandlerFuncs) (*ZkController, error) {
	if len(zkHost) == 0 {
		return nil, fmt.Errorf("create Controller failed, empty zookeeper host list")
	}
	if len(path) == 0 {
		return nil, fmt.Errorf("create Controller failed, empty watch path")
	}
	cxt, cancel := context.WithCancel(context.Background())
	controller := &ZkController{
		conCxt:       cxt,
		conCancel:    cancel,
		resyncperiod: period,
		watchpath:    path,
		zkHost:       zkHost,
		storage:      cache,
		nsStorage:    make(map[string]context.CancelFunc),
		decoder:      decoder,
		funcs:        eventFunc,
	}
	//create connection to zookeeper
	client := zkclient.NewZkClient(zkHost)
	if conErr := client.ConnectEx(time.Second * 5); conErr != nil {
		return nil, fmt.Errorf("controller create zk conection failed, %s", conErr.Error())
	}
	controller.client = &wrapperClient{
		client: client,
	}
	return controller, nil
}

//RunController running controller, create goroutine watch zookeeper data
func (ctrl *ZkController) RunController(stopCh <-chan struct{}) error {
	//check path exist
	if err := ctrl.existWatch(ctrl.watchpath); err != nil {
		return err
	}
	//get all children node & setting watch
	if err := ctrl.dataTypeWatch(ctrl.watchpath); err != nil {
		return fmt.Errorf("Controller watch %s err, %s", ctrl.watchpath, err.Error())
	}

	//create resync event for all data
	tick := time.NewTicker(time.Second * time.Duration(ctrl.resyncperiod))
	defer tick.Stop()
	for {
		select {
		case <-ctrl.conCxt.Done():
			log.Printf("[WARN] controller ask exist")
			return nil
		case now := <-tick.C:
			//resync all data from watchpath
			if ctrl.Resynced() {
				log.Printf("[WARN] controller %s under resync now [%s], drop this tick.", ctrl.watchpath, now.String())
				continue
			}
			ctrl.resynced = true
			log.Printf("[INFO] resync %s tick, now: %s", ctrl.watchpath, now.String())
			ctrl.dataResync()
			log.Printf("[INFO] resync %s tick, end: %s", ctrl.watchpath, time.Now().String())
			ctrl.resynced = false
		}
	}
}

//StopController stop controller event, clean all data
func (ctrl *ZkController) StopController() {
	log.Printf("[INFO] controller %s stop", ctrl.watchpath)
	ctrl.conCancel()
}

//Resynced check controller is under resynced
func (ctrl *ZkController) Resynced() bool {
	return ctrl.resynced
}

/**
 * inner method for watching zookeeper data
 */

//existWatch check path is exist, block until path exist
func (ctrl *ZkController) existWatch(path string) error {
	if len(path) == 0 {
		return fmt.Errorf("controller err, get empty path")
	}
	ok, err := ctrl.client.Exists(path)
	if err != nil {
		return fmt.Errorf("controller check path %s err, %s", path, err.Error())
	}
	if !ok {
		//path do not exist, create block watch waits for path created
		log.Printf("[WARN] controller check path %s do not exist, setting watch waiting for create event.", path)
		_, _, events, err := ctrl.client.ExistsW(path)
		if err != nil {
			log.Printf("[ERROR] controller create %s exist watch failed, %s", path, err.Error())
			return err
		}
		select {
		case <-ctrl.conCxt.Done():
			log.Printf("[WARN] controller asked to exit when watching path %s", path)
			return fmt.Errorf("controller forced exit")
		case <-events:
			log.Printf("[INFO] controller get path %s existence event.", path)
			return nil
		}
	}
	return nil
}

//typeWatch watch data type node from zookeeper.
//data type comes from bcs-scheduler
func (ctrl *ZkController) dataTypeWatch(path string) error {
	log.Printf("[INFO] controller create watch for %s", path)
	namespaces, stat, events, err := ctrl.client.ChildrenW(path)
	if err != nil {
		log.Printf("[ERROR] Watch Node %s children failed, %s", path, err.Error())
		return err
	} else if stat == nil {
		log.Printf("[ERROR] Wath node %s state return nil, setting watch failed.", path)
		return fmt.Errorf("error watch stat for %s", path)
	}
	ctrl.namespaceCheck(namespaces)
	//wait for namespace node event
	go func() {
		select {
		case <-ctrl.conCxt.Done():
			log.Printf("controller ask to stop watch %s children event", path)
			return
		case event := <-events:
			if event.Type == zk.EventNodeChildrenChanged {
				allNS, err := ctrl.listChildrenNode(path)
				if err != nil {
					//list failed
					log.Printf("[ERROR] Controller list init path %s children failed in ChildrenChangeEvent, %s", path, err.Error())
				} else {
					ctrl.namespaceCheck(allNS)
				}
			} else {
				log.Printf("path trigger event we don't care, event: %s", event.Type.String())
			}
			//register watch again
			if err := ctrl.dataTypeWatch(path); err != nil {
				log.Printf("controller create init path %s watch failed, %s", path, err.Error())
			}
			return
		}
	}()
	//doing good if we get here
	return nil
}

//listNamespace list all namespace node
func (ctrl *ZkController) listChildrenNode(path string) ([]string, error) {
	children, stat, err := ctrl.client.Children(path)
	if err != nil {
		log.Printf("[ERROR] Get Node %s children failed, %s", path, err.Error())
		return nil, err
	} else if stat == nil {
		log.Printf("[ERROR] Get Node %s state return nil, setting watch failed.", path)
		return nil, fmt.Errorf("error watch stat for %s", path)
	}
	return children, nil
}

//namespaces check namespace is under watch
func (ctrl *ZkController) namespaceCheck(ns []string) {
	if len(ns) == 0 {
		return
	}
	for _, item := range ns {
		ctrl.nsLock.Lock()
		if _, ok := ctrl.nsStorage[item]; ok {
			log.Printf("[INFO] controller already got namespace %s in cache", item)
		} else {
			//create watch for namespace
			ns := filepath.Join(ctrl.watchpath, item)
			nsCxt, nsCancel := context.WithCancel(ctrl.conCxt)
			if err := ctrl.namespaceWatch(nsCxt, ns); err != nil {
				log.Printf("[ERROR] controller watch data namespace %s failed, %s. data will be fixed in next Resync event.", ns, err.Error())
			} else {
				ctrl.nsStorage[item] = nsCancel
				log.Printf("[WARN] controller add new namespace %s success", ns)
			}
		}
		ctrl.nsLock.Unlock()
	}
}

//namespaceClean clean name with name
func (ctrl *ZkController) namespaceClean(ns string) {
	if len(ns) == 0 {
		return
	}
	ctrl.nsLock.Lock()
	if item, ok := ctrl.nsStorage[ns]; ok {
		delete(ctrl.nsStorage, ns)
		item()
		log.Printf("controller got namespace %s, cancel all subContext and release", ns)
	} else {
		log.Printf("[ERROR] controller lost namespace %s in cache when clean", ns)
	}
	ctrl.nsLock.Unlock()
}

//namespaceWatch watch data node under namespace
func (ctrl *ZkController) namespaceWatch(cxt context.Context, path string) error {
	log.Printf("[INFO] controller prepare to watch namespace %s children", path)
	ns := filepath.Base(path)
	dataNodes, stat, events, err := ctrl.client.ChildrenW(path)
	if err != nil {
		log.Printf("[ERROR] Watch Node %s children failed, %s", path, err.Error())
		return err
	} else if stat == nil {
		log.Printf("[ERROR] Wath node %s state return nil, setting watch failed.", path)
		return fmt.Errorf("error watch stat for %s", path)
	}
	//goroutine watch
	ctrl.dataNodeCheck(cxt, ns, dataNodes)
	//wait for namespace node event
	go func() {
		//todo(developer): add panic protection for cleaning nsStorage dirty data
		select {
		case <-cxt.Done():
			log.Printf("[WARN] controller ask to stop watch %s children event", path)
			return
		case event := <-events:
			if event.Type == zk.EventNodeChildrenChanged {
				allData, err := ctrl.listChildrenNode(path)
				if err != nil {
					//list failed
					log.Printf("[ERROR] Controller list namespace path %s children failed in ChildrenChangeEvent, %s", path, err.Error())
				} else {
					ctrl.dataNodeCheck(cxt, ns, allData)
				}
			} else if event.Type == zk.EventNodeDeleted {
				//namespace node delete, clean namespace data & clean all context
				log.Printf("[WARN] controller get namespace %s delete event, stop watch.", path)
				ctrl.namespaceClean(ns)
				return
			} else {
				log.Printf("[WARN] path %s trigger event we don't care, event: %s", path, event.Type.String())
			}
			//register watch again
			if err := ctrl.namespaceWatch(cxt, path); err != nil {
				log.Printf("[FATAL] controller create init path %s watch failed, %s", path, err.Error())
			}
			return
		}
	}()
	//doing good if we get here
	return nil
}

func (ctrl *ZkController) dataNodeCheck(cxt context.Context, ns string, nodeName []string) {
	if len(nodeName) == 0 {
		log.Printf("[WARN] controller get no data under %s/%s", ctrl.watchpath, ns)
		return
	}
	for _, node := range nodeName {
		key := ns + "/" + node
		if _, exist, _ := ctrl.storage.GetByKey(key); !exist {
			//create watch for this new node
			dataCxt, _ := context.WithCancel(cxt)
			path := filepath.Join(ctrl.watchpath, ns, node)
			go ctrl.dataWatch(dataCxt, ns, path)
			log.Printf("[WARN] controller set path %s under watch", path)
		}
		//todo(developer): maybe we need to add checking mechanism
		//for checking goroutine of data node is healthy
	}
}

//dataContentWatch watch detail data content node
func (ctrl *ZkController) dataWatch(cxt context.Context, ns string, path string) {
	data, stat, events, err := ctrl.client.GetW(path)
	if err != nil {
		log.Printf("[ERROR] controller get node data %s failed: %s, watch stop", path, err.Error())
		return
	} else if stat == nil {
		log.Printf("[ERROR] controller get node data %s failed: state return nil", path)
		return
	}
	cur, derr := ctrl.decoder.Decode(data)
	if derr != nil {
		log.Printf("[ERROR] controller decode %s data failed: %s, wait next watch to fix", path, derr.Error())
	} else {
		old, exist, _ := ctrl.storage.Get(cur)
		if exist {
			ctrl.storage.Update(cur)
			ctrl.funcs.UpdateFunc(old, cur)
			log.Printf("[INFO] %s update content", path)
			metrics.ZkNotifyTotal.WithLabelValues(metrics.UpdateOperation).Inc()
		} else {
			ctrl.storage.Add(cur)
			ctrl.funcs.AddFunc(cur)
			log.Printf("[INFO] %s add new data object", path)
			metrics.DnsTotal.Inc()
			metrics.ZkNotifyTotal.WithLabelValues(metrics.AddOperation).Inc()
		}
	}
	//wait for watch event
	for {
		select {
		case <-cxt.Done():
			log.Printf("[WARN] controller stop path %s watching", path)
			return
		case event := <-events:
			if event.Type == zk.EventNodeDataChanged {
				go ctrl.dataWatch(cxt, ns, path)
				return
			} else if event.Type == zk.EventNodeDeleted {
				node := filepath.Base(path)
				ctrl.dataClean(ns, node)
				return
			} else {
				log.Printf("[WARN] data path %s trigger event we don't care, event: %s", path, event.Type.String())
				//todo(developer): watch will be lost if we go here
				//we need [[ go ctrl.dataWatch(cxt, ns, path)]]
				return
			}
		}
	}
}

//dataClean clean data object by key
func (ctrl *ZkController) dataClean(ns, node string) {
	if len(node) == 0 {
		return
	}
	key := filepath.Join(ns, node)
	key = strings.ToLower(key)
	old, exist, _ := ctrl.storage.GetByKey(key)
	if exist {
		ctrl.storage.Delete(old)
		ctrl.funcs.DeleteFunc(old)
		log.Printf("[WARN] controller delete %s under %s in cache", key, ctrl.watchpath)
		metrics.DnsTotal.Dec()
		metrics.ZkNotifyTotal.WithLabelValues(metrics.DeleteOperation).Inc()
	} else {
		log.Printf("[ERROR] controller lost %s in cache, somewhere go wrong", key)
	}
}

//dataResync resync all data from zookeeper
func (ctrl *ZkController) dataResync() {
	//list all namespace node
	namespaces, _, err := ctrl.client.Children(ctrl.watchpath)
	if err != nil {
		log.Printf("[ERROR] controller get path %s children failed, %s", ctrl.watchpath, err.Error())
		return
	}
	if len(namespaces) == 0 {
		log.Printf("[WARN] controller check path %s no children, wait next resync", ctrl.watchpath)
		return
	}
	//when we iterator all data in zookeeper, we checking:
	//1. update cache with zookeeper data by force
	//2. cache data is dirty or not(exist in cache but lost in zookeeper)

	//step 1
	existsIndex := make(map[string]bool)
	for _, ns := range namespaces {
		//get all data under namespace
		nspath := filepath.Join(ctrl.watchpath, ns)
		dataNode, err := ctrl.listChildrenNode(nspath)
		if err != nil {
			log.Printf("[ERROR] controller get path %s children data failed, %s, try next one", nspath, err.Error())
			continue
		}
		if len(dataNode) == 0 {
			log.Printf("[INFO] [INFO] controller get no children of %s", nspath)
			continue
		}
		//list all data
		for _, node := range dataNode {
			nodepath := filepath.Join(nspath, node)
			data, _, err := ctrl.client.Get(nodepath)
			if err != nil {
				log.Printf("[ERROR] controller get %s err, %s", nodepath, err.Error())
				continue
			}
			cur, derr := ctrl.decoder.Decode(data)
			if derr != nil {
				log.Printf("[ERROR] controller decode %s data failed: %s, wait next watch to fix", nodepath, derr.Error())
			} else {
				key := ns + "/" + node
				old, exist, _ := ctrl.storage.Get(cur)
				if exist {
					ctrl.storage.Update(cur)
					ctrl.funcs.UpdateFunc(old, cur)
					existsIndex[key] = true
				} else {
					//todo(developer): lost data in cache, we add it to cache,
					//and we still need add watch goroutine for
					//updating data from zookeeper
					log.Printf("[WARN] RESYNC found ###%s### lost in cache", nodepath)
					ctrl.storage.Add(cur)
					ctrl.funcs.AddFunc(cur)
					metrics.DnsTotal.Inc()
				}
			}
		}
	}

	//step 2
	cacheIndexs := ctrl.storage.ListKeys()
	for _, indexKey := range cacheIndexs {
		if _, ok := existsIndex[indexKey]; ok {
			continue
		}
		checkpath := filepath.Join(ctrl.watchpath, indexKey)
		exist, err := ctrl.client.Exists(checkpath)
		if err != nil {
			log.Printf("[ERROR] controller check cache index %s error in zookeeper, %s", indexKey, err.Error())
			continue
		}
		if exist {
			log.Printf("[WARN] %s all in cache& zk, but lost in tmp map, maybe added latest or trigger RESYNC warn", indexKey)
			continue
		}
		log.Printf("[WARN] %s lost in zookeeper, TODO: delete in cache.", indexKey)
		//todo(developer): logs statistic for repairing this warnning
	}
}
