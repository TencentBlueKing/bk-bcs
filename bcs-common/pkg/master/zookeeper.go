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

package master

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	zktype "github.com/samuel/go-zookeeper/zk"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcstypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
)

// NewZookeeperMaster create zk master
func NewZookeeperMaster(hosts []string, path string, self *bcstypes.ServerInfo) (Master, error) {
	if len(hosts) == 0 {
		return nil, fmt.Errorf("empty zookeeper host")
	}
	if len(path) == 0 {
		return nil, fmt.Errorf("empty zookeeper path")
	}
	if len(self.IP) == 0 || self.Port == 0 || self.Pid == 0 {
		return nil, fmt.Errorf("invalid self Server info")
	}
	cxt, cancel := context.WithCancel(context.Background())
	zk := &ZookeeperMaster{
		zkHosts:    hosts,
		parentPath: path,
		isMaster:   false,
		healthy:    false,
		exitCancel: cancel,
		exitCxt:    cxt,
		self:       self,
	}
	return zk, nil
}

// ZookeeperMaster implementation for master in zookeeper
type ZookeeperMaster struct {
	zkHosts    []string             // zk host info
	parentPath string               // parent path in zk
	selfPath   string               // self node
	isMaster   bool                 // master status
	healthy    bool                 // status
	exitCancel context.CancelFunc   // exit func
	exitCxt    context.Context      // exit context
	self       *bcstypes.ServerInfo // self server info
	client     *zkclient.ZkClient   // zk client from common
}

// Init init stage, like create connection
func (zk *ZookeeperMaster) Init() error {
	zk.isMaster = false
	// create connection to zookeeper
	zk.client = zkclient.NewZkClient(zk.zkHosts)
	if err := zk.client.ConnectEx(time.Second * 10); err != nil {
		return fmt.Errorf("Init failed when connect zk, %s", err.Error())
	}
	return nil
}

// Finit init stage, like create connection
func (zk *ZookeeperMaster) Finit() {
	// close connection to zookeeper
	if zk.client != nil {
		zk.client.Close()
	}
	zk.client = nil
}

// Register registry information to storage
func (zk *ZookeeperMaster) Register() error {
	if err := zk.createSelfNode(); err != nil {
		return err
	}
	// create event loop for master flag
	go zk.masterLoop()
	go zk.healthLoop()
	return nil

}

// Clean clean self node
func (zk *ZookeeperMaster) Clean() error {
	// delete self node
	zk.exitCancel()
	if len(zk.selfPath) > 0 || zk.client != nil {
		_ = zk.client.Del(zk.selfPath, 0)
	}
	zk.isMaster = false
	return nil
}

// IsMaster check if self is master or not
func (zk *ZookeeperMaster) IsMaster() bool {
	return zk.isMaster
}

// CheckSelfNode check self node exist, and data correct
func (zk *ZookeeperMaster) CheckSelfNode() (bool, error) {
	if zk.client == nil {
		return false, fmt.Errorf("zookeeper do not Init")
	}
	children, _, err := zk.client.GetChildrenEx(zk.parentPath)
	if err != nil {
		return false, err
	}
	if len(children) == 0 {
		return false, nil
	}
	found := false
	for _, node := range children {
		var nodepath string
		if strings.HasSuffix(zk.parentPath, "/") {
			nodepath = zk.parentPath + node
		} else {
			nodepath = zk.parentPath + "/" + node
		}
		data, _, err := zk.client.GetEx(nodepath)
		if err != nil {
			blog.Errorf("zookeeper master in check self node, get %s err, %s", nodepath, err.Error())
			continue
		}
		var info bcstypes.ServerInfo
		if err := json.Unmarshal(data, &info); err != nil {
			blog.Warnf("zookeeper master in check self node parse %s json failed, %s", nodepath, err.Error())
			continue
		}
		if info.IP == zk.self.IP && info.Port == zk.self.Port && info.Pid == zk.self.Pid {
			found = true
			break
		}
	}
	return found, nil
}

// GetAllNodes get all server nodes
func (zk *ZookeeperMaster) GetAllNodes() ([]*bcstypes.ServerInfo, error) {
	return nil, nil
}

// GetPath setting self info, now is ip address & port
func (zk *ZookeeperMaster) GetPath() string {
	return zk.selfPath
}

func (zk *ZookeeperMaster) createSelfNode() error {
	if zk.client == nil {
		return fmt.Errorf("zk master is not Init")
	}
	data, err := json.Marshal(zk.self)
	if err != nil {
		return fmt.Errorf("self data jsonlize failed, %s", err.Error())
	}
	var selfnode string
	if strings.HasSuffix(zk.parentPath, "/") {
		selfnode = zk.parentPath + zk.self.IP
	} else {
		selfnode = zk.parentPath + "/" + zk.self.IP
	}
	zk.selfPath, err = zk.client.CreateEphAndSeqEx(selfnode, data)
	if err != nil {
		return fmt.Errorf("register self node failed, %s", err.Error())
	}
	blog.Infof("ZookeeperMaster register node %s", zk.selfPath)
	return nil
}

func (zk *ZookeeperMaster) masterLoop() {
	zk.healthy = true
	// watch parent path for children changed
	children, _, evChannel, err := zk.client.ChildrenW(zk.parentPath)
	if err != nil {
		blog.Errorf("zookeeper master watch %s  all nodes failed, %s", zk.parentPath, err.Error())
		zk.healthy = false
		zk.isMaster = false
		return
	}
	if len(children) == 0 {
		blog.Errorf("zookeeper master get empty nodes under %s, even self node. status unhealthy", zk.parentPath)
		zk.healthy = false
		zk.isMaster = false
		return
	}
	// watch works
	nodes := zk.sortNodes(children)
	// check self is master
	var nodepath string
	if strings.HasSuffix(zk.parentPath, "/") {
		nodepath = zk.parentPath + nodes[0]
	} else {
		nodepath = zk.parentPath + "/" + nodes[0]
	}
	data, _, err := zk.client.GetEx(nodepath)
	if err != nil {
		blog.Errorf("zookeeper master get %s err, %s", nodepath, err.Error())
		zk.healthy = false
		zk.isMaster = false
		return
	}
	if len(data) > 0 {
		var info bcstypes.ServerInfo
		if err := json.Unmarshal(data, &info); err != nil {
			blog.Warnf("zookeeper master parse %s json failed, %s", nodepath, err.Error())
			zk.healthy = false
			zk.isMaster = false
			return
		}
		if info.IP == zk.self.IP && info.Port == zk.self.Port && info.Pid == zk.self.Pid {
			// self is Master
			zk.isMaster = true
			blog.Infof("#######Status chanaged, Self ndoe become Master#############")
		} else {
			zk.isMaster = false
		}
	} else {
		blog.Warnf("zookeeper master get empty data with first node %s", nodepath)
		zk.isMaster = false
	}
	select {
	case <-zk.exitCxt.Done():
		blog.Infof("zookeeper master asked to exit.")
		return
	case event := <-evChannel:
		if event.Type == zktype.EventNodeChildrenChanged {
			// check master status
			go zk.masterLoop()
			return
		}
		if event.Type == zktype.EventSession {
			blog.Warnf("zookeeper happened EventSession in children watch.")
			go zk.masterLoop()
			return
		}
		if event.Type == zktype.EventNotWatching {
			blog.Warnf("zookeeper happened EventNotWatching in children watch. try to watch again")
			go zk.masterLoop()
			return
		}
		// other events, include connection broken
		// depend on health check loop to recovery
	}
	zk.healthy = false
}

func (zk *ZookeeperMaster) healthLoop() {
	masterTick := time.NewTicker(time.Second * 2)
	selfTick := time.NewTicker(time.Second * 30)
	defer masterTick.Stop()
	defer selfTick.Stop()

	for {
		select {
		case <-zk.exitCxt.Done():
			blog.Infof("zookeeper master healthy Loop asked exit.")
			return
		case <-masterTick.C:
			if !zk.healthy {
				blog.Warnf("****************master check loop exit, arise this loop*****************")
				go zk.masterLoop()
			}
		case <-selfTick.C:
			blog.Infof("SelfNode Master Status ***%v***", zk.isMaster)
			ok, err := zk.CheckSelfNode()
			if err != nil {
				blog.Errorf("check self node error, %s, try next tick.", err.Error())
				continue
			}
			if !ok {
				blog.Errorf("###we lost self Node data. rebuild it##")
				if err := zk.createSelfNode(); err != nil {
					blog.Errorf("********rebuild seld node data failed, %s***********", err.Error())
				} else {
					blog.Warnf("we rebuild zookeeper self node success.")
				}
			}
		}
	}
}

func (zk *ZookeeperMaster) sortNodes(nodes []string) []string {
	if len(nodes) == 1 {
		return nodes
	}
	var sortPart []int
	mapSortNode := make(map[int]string)
	for _, chNode := range nodes {
		if len(chNode) <= 10 {
			fmt.Printf("node(%s) is less then 10, there is not the seq number\n", chNode)
			continue
		}

		p, err := strconv.Atoi(chNode[len(chNode)-10:])
		if err != nil {
			fmt.Printf("fail to conv string to seq number for node(%s), err:%s\n", chNode, err.Error())
			continue
		}

		sortPart = append(sortPart, p)
		mapSortNode[p] = chNode
	}

	sort.Ints(sortPart)

	var children []string
	for _, part := range sortPart {
		children = append(children, mapSortNode[part])
	}

	return children
}
