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

package client

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	regd "github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

// NewMasterTracker create new master tracker
func NewMasterTracker(zkAddr string) (*MasterTracker, error) {
	blog.Infof("starting master tracker.")
	disc := regd.NewRegDiscoverEx(zkAddr, time.Duration(5*time.Second))
	if err := disc.Start(); nil != err {
		return nil, fmt.Errorf("start get ccapi zk service failed. Error:%v", err)
	}
	watchPath := fmt.Sprintf("%s/%s/%s", types.BCS_SERV_BASEPATH, types.BCS_MODULE_HEALTH, "master")
	eventChan, eventErr := disc.DiscoverService(watchPath)
	if nil != eventErr {
		return nil, fmt.Errorf("start running discover service failed. Error:%v", eventErr)
	}
	m := &MasterTracker{
		eventChan: eventChan,
		svr: &server{
			healthServers: make(map[string]string),
		},
	}
	m.run()
	return m, nil
}

type MasterTracker struct {
	eventChan <-chan *regd.DiscoverEvent
	svr       *server
}

func (m *MasterTracker) run() {
	go func() {
		blog.Infof("start to sync bcs-health address from zk.")
		for svr := range m.eventChan {
			blog.Info("received one zk event which may contains bcs-health address.")
			if svr.Err != nil {
				blog.Errorf("get bcs-health addr failed. but will continue watch. err: %v", svr.Err)
				continue
			}
			if len(svr.Server) <= 0 {
				blog.Warnf("get 0 bcs-health master addr from zk.")
				continue
			}
			m.updateServers(svr.Server[0])
		}
	}()
}

// GetServers get first healthy server
func (m *MasterTracker) GetServers() string {
	m.svr.locker.Lock()
	defer m.svr.locker.Unlock()
	for svr := range m.svr.healthServers {
		return svr
	}
	return ""
}

func (m *MasterTracker) updateServers(svr string) {
	m.svr.locker.Lock()
	defer m.svr.locker.Unlock()

	info := types.ServerInfo{}
	if err := json.Unmarshal([]byte(svr), &info); nil != err {
		blog.Errorf("unmashal health server master info failed. reason: %v", err)
		return
	}
	if len(info.IP) == 0 || info.Port == 0 || len(info.Scheme) == 0 {
		blog.Errorf("get invalid health master info: %s", svr)
		return
	}
	addr := fmt.Sprintf("%s://%s:%d", info.Scheme, info.IP, info.Port)
	if _, exist := m.svr.healthServers[addr]; exist {
		return
	}

	m.svr.healthServers[addr] = ""
	blog.Infof("*** get new bcs-health master client, addr: %s ***", addr)

	for svr := range m.svr.healthServers {
		if svr != addr {
			delete(m.svr.healthServers, svr)
			blog.Infof("*** remove old bcs-health master server, addr: %s ***.", svr)
		}
	}
}

type server struct {
	locker        sync.Mutex
	healthServers map[string]string
}
