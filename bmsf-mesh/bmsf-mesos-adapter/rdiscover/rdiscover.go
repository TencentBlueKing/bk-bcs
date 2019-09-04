/*
Copyright (C) 2019 The BlueKing Authors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package rdiscover

import (
	rd "bk-bcs/bcs-common/common/RegisterDiscover"
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/common/version"
	"context"
	"encoding/json"
	"os"
	"time"
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
	zkAddr     string
	clusterID  string
	ip         string
	metricPort uint
	eventQueue chan RoleEvent
	cancel     context.CancelFunc
}

// NewAdapterDiscover create Adapter Discover
func NewAdapterDiscover(zkAddr, ip, clusterID string, metricPort uint) (*AdapterDiscover, <-chan RoleEvent) {
	eventQueue := make(chan RoleEvent, 128)
	return &AdapterDiscover{
		isMaster:   false,
		rd:         rd.NewRegDiscoverEx(zkAddr, 20*time.Second),
		zkAddr:     zkAddr,
		ip:         ip,
		clusterID:  clusterID,
		metricPort: metricPort,
		eventQueue: eventQueue,
	}, eventQueue
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
	discoveryEvent, err := ad.rd.DiscoverService(adaptersPath)
	if err != nil {
		blog.Errorf("failed to register discover for mesos-adapter, err %s", err.Error())
		blog.Infof("restart AdapterDiscover after 3 second")
		ad.rd.Stop()
		time.Sleep(3 * time.Second)
		go ad.Start()
		return
	}
	for {
		select {
		case curEvent := <-discoveryEvent:
			servers := curEvent.Server
			blog.V(3).Infof("discover mesos adapters(%v)", servers)
			adapters := []*types.ServerInfo{}
			for _, server := range servers {
				adapter := new(types.ServerInfo)
				err = json.Unmarshal([]byte(server), adapter)
				if err != nil {
					blog.Warnf("failed to unmarshal(%s), err %s", server, err.Error())
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
