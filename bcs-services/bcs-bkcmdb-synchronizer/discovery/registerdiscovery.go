/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package discovery

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

// Client client for service discovery
type Client struct {
	zkAddrs   string
	zkSubPath string
	svcInfo   *types.ServerInfo
	rd        *RegisterDiscover.RegDiscover
	rootCxt   context.Context
	cancel    context.CancelFunc

	services []*types.ServerInfo
	isMaster bool
	lock     sync.Mutex
}

// New create client for service discovery
func New(zkAddrs, zkSubPath string, svcInfo *types.ServerInfo) *Client {
	return &Client{
		zkAddrs:   zkAddrs,
		zkSubPath: zkSubPath,
		svcInfo:   svcInfo,
		rd:        RegisterDiscover.NewRegDiscover(zkAddrs),
	}
}

// Run run service discovery
func (c *Client) Run() error {
	c.rootCxt, c.cancel = context.WithCancel(context.Background())

	if err := c.rd.Start(); err != nil {
		blog.Errorf("fail to connect registry, err %s", err.Error())
		return err
	}

	if err := c.registerService(); err != nil {
		blog.Errorf("fail to register service, err %s", err.Error())
		return err
	}

	path := c.getServicePath()
	discoveryEvent, err := c.rd.DiscoverService(path)
	if err != nil {
		blog.Errorf("failed to discover service by path %s, err %s", path, err.Error())
		return err
	}

	go c.checkMasterStatus()
	for {
		select {
		case e := <-discoveryEvent:
			err := c.resolveServices(e.Server)
			if err != nil {
				blog.Warnf("resolve service %s failed, err %s", e.Server, err.Error())
			}
		case <-c.rootCxt.Done():
			blog.Warnf("service discovery done")
			return nil
		}
	}
}

// GetServers get server info
func (c *Client) GetServers() []*types.ServerInfo {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.services
}

// IsMaster return if it is master
func (c *Client) IsMaster() bool {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.isMaster
}

// Stop stop discovery
func (c *Client) Stop() error {
	c.cancel()
	if err := c.rd.Stop(); err != nil {
		return fmt.Errorf("stop discovery failed err %s", err.Error())
	}
	return nil
}

// resolve services
func (c *Client) resolveServices(svcStrs []string) error {
	blog.V(3).Infof("discovery bkcmdb-synchronizer [%+v]", svcStrs)

	svcs := []*types.ServerInfo{}
	for _, serverStr := range svcStrs {
		newSvc := new(types.ServerInfo)
		if err := json.Unmarshal([]byte(serverStr), newSvc); err != nil {
			blog.Warnf("failed to unmarshal %s, err %s", serverStr, err.Error())
			continue
		}
		svcs = append(svcs, newSvc)
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	c.services = svcs
	return nil
}

// check master status
func (c *Client) checkMasterStatus() {
	for {
		c.lock.Lock()
		if len(c.services) > 0 {
			if c.services[0].IP == c.svcInfo.IP {
				if !c.isMaster {
					blog.Infof("[Role changed] slave->master")
				}
				c.isMaster = true
			} else {
				c.isMaster = false
			}
		}
		c.lock.Unlock()
		time.Sleep(5 * time.Second)
	}
}

// registerService
func (c *Client) registerService() error {
	data, err := json.Marshal(c.svcInfo)
	if err != nil {
		blog.Errorf("failed to marshal service info, err %s", err.Error())
		return err
	}
	path := filepath.Join(c.getServicePath(), c.svcInfo.IP)

	return c.rd.RegisterAndWatchService(path, data)
}

// getServicePath
func (c *Client) getServicePath() string {
	return filepath.Join(types.BCS_SERV_BASEPATH, types.BCS_MODULE_BKCMDB_SYNCHRONIZER, c.zkSubPath)
}
