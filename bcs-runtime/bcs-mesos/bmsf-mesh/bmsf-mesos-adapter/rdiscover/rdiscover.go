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

package rdiscover

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	rd "github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
)

// RoleEvent event for role change
type RoleEvent string

const (
	// MasterToSlave event role change from master to slave
	MasterToSlave = "m2s"
	// SlaveToMaster event role change from slave to master
	SlaveToMaster = "s2m"
)

// AdapterDiscover service discovery and master election for mesos adapter
type AdapterDiscover struct {
	isMaster   bool
	rd         *rd.RegDiscover
	zkClient   *zkclient.ZkClient
	zkAddr     string
	clusterID  string
	ip         string
	metricPort uint
	eventQueue chan RoleEvent
	cancel     context.CancelFunc
}

// NewAdapterDiscover create Adapter Discover
func NewAdapterDiscover(zkAddr, ip, clusterID string, metricPort uint) (*AdapterDiscover, <-chan RoleEvent, error) {

	zkAddr = strings.Replace(zkAddr, ";", ",", -1)
	zkAddrs := strings.Split(zkAddr, ",")
	zkClient := zkclient.NewZkClient(zkAddrs)
	err := zkClient.Connect()
	if err != nil {
		blog.Errorf("AdapterDiscover connect zk failed, err %s", err.Error())
		return nil, nil, err
	}

	eventQueue := make(chan RoleEvent, 128)
	return &AdapterDiscover{
		isMaster:   false,
		rd:         rd.NewRegDiscoverEx(zkAddr, 20*time.Second),
		zkClient:   zkClient,
		zkAddr:     zkAddr,
		ip:         ip,
		clusterID:  clusterID,
		metricPort: metricPort,
		eventQueue: eventQueue,
	}, eventQueue, nil
}

// Start register zk path and monitor all registered adapters
func (ad *AdapterDiscover) Start() {
	var err error
	rootCxt, cancel := context.WithCancel(context.Background())
	ad.cancel = cancel
	err = ad.rd.Start()
	if err != nil {
		blog.Errorf("failed to start RegisterDiscover, err %s", err.Error())
		blog.Infof("restart AdapterDiscover after 3 second")
		time.Sleep(3 * time.Second)
		go ad.Start()
		return
	}
	err = ad.registerAdapter()
	if err != nil {
		blog.Errorf("failed to register mesos adapter, err %s", err.Error())
		blog.Infof("restart AdapterDiscover after 3 second")
		ad.rd.Stop()
		time.Sleep(3 * time.Second)
		go ad.Start()
		return
	}
	adaptersPath := ad.getZkDiscoveryPath()
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			children, err := ad.zkClient.GetChildren(adaptersPath)
			if err != nil {
				blog.Warnf("get children of %s failed, err %s", adaptersPath, err.Error())
				continue
			}
			children = ad.sortNode(children)
			adapters := []*types.ServerInfo{}
			for _, tmpChild := range children {
				tmpData, err := ad.zkClient.Get(adaptersPath + "/" + tmpChild)
				if err != nil {
					blog.Warnf("get data for child %s failed, err %s", tmpChild)
					continue
				}
				adapter := new(types.ServerInfo)
				err = json.Unmarshal([]byte(tmpData), adapter)
				if err != nil {
					blog.Warnf("failed to unmarshal(%s), err %s", tmpData, err.Error())
					continue
				}
				adapters = append(adapters, adapter)
			}
			if len(adapters) == 0 {
				blog.Warnf("found no registered adapters")
				if ad.isMaster {
					ad.isMaster = false
					ad.eventQueue <- MasterToSlave
				}
				blog.Infof("Role changed, I become Slave")
				continue
			}
			if adapters[0].IP == ad.ip && !ad.isMaster {
				ad.isMaster = true
				ad.eventQueue <- SlaveToMaster
				blog.Infof("Role changed, I become Master")
				continue
			} else if adapters[0].IP != ad.ip && ad.isMaster {
				ad.isMaster = false
				ad.eventQueue <- MasterToSlave
				blog.Infof("Role changed, I become Slave")
				continue
			}

		case <-rootCxt.Done():
			blog.Warnf("AdapterDiscover context done")
			return
		}
	}

}

func (ad *AdapterDiscover) registerAdapter() error {
	host, err := os.Hostname()
	if err != nil {
		blog.Error("mesos adapter get hostname err: %s", err.Error())
		host = "UNKOWN"
	}
	serverInfo := types.ServerInfo{}
	serverInfo.IP = ad.ip
	serverInfo.Port = ad.metricPort
	serverInfo.MetricPort = ad.metricPort
	serverInfo.Pid = os.Getpid()
	serverInfo.Version = version.GetVersion()
	serverInfo.Cluster = ad.clusterID
	serverInfo.HostName = host
	// TODO: support https
	serverInfo.Scheme = "http"
	serverInfoByte, err := json.Marshal(serverInfo)
	if err != nil {
		blog.Errorf("fail to marshal mesos-adapter info to bytes, err %s", err.Error())
		return err
	}
	serverRegisterPath := ad.getZkDiscoveryPath() + "/" + serverInfo.IP
	return ad.rd.RegisterAndWatchService(serverRegisterPath, serverInfoByte)
}

func (ad *AdapterDiscover) getZkDiscoveryPath() string {
	return types.BCS_SERV_BASEPATH + "/" + types.BCS_MODULE_MESOSADAPTER + "/" + ad.clusterID
}

func (ad *AdapterDiscover) sortNode(nodes []string) []string {
	var sortPart []int
	mapSortNode := make(map[int]string)
	for _, chNode := range nodes {
		if len(chNode) <= 10 {
			blog.V(3).Infof("node(%s) is less then 10, there is not the seq number", chNode)
			continue
		}

		p, err := strconv.Atoi(chNode[len(chNode)-10:])
		if err != nil {
			blog.V(3).Infof("fail to conv string to seq number for node(%s), err:%s", chNode, err.Error())
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
