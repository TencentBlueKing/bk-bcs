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

package role

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	regd "github.com/Tencent/bk-bcs/bcs-common/common/RegisterDiscover"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
)

type RoleInterface interface {
	IsMaster() bool
}

func NewRoleController(localIP string, metricPort uint, zkAddr string, watchPath string) (RoleInterface, error) {
	if len(localIP) == 0 {
		return nil, fmt.Errorf("local ip must be configed")
	}
	r := &Role{
		localIP:     localIP,
		metricPort:  metricPort,
		CurrentRole: SlaveRole,
	}
	if err := r.WatchRole(zkAddr, watchPath); nil != err {
		return nil, fmt.Errorf("role controller watch zk failed. err: %v", err)
	}
	return r, nil
}

type RoleType string

const (
	MasterRole RoleType = "Master"
	SlaveRole  RoleType = "Slave"
)

type Role struct {
	locker      sync.RWMutex
	localIP     string
	metricPort  uint
	CurrentRole RoleType
}

func (r *Role) SetRole(role RoleType) {
	r.locker.Lock()
	defer r.locker.Unlock()
	if r.CurrentRole == role {
		return
	}
	r.CurrentRole = role
	blog.Warnf("change endpoint role to %s", r.CurrentRole)
}

func (r *Role) IsMaster() bool {
	r.locker.Lock()
	defer r.locker.Unlock()
	return r.CurrentRole == MasterRole
}

func (r *Role) GetRole() RoleType {
	r.locker.Lock()
	defer r.locker.Unlock()
	return r.CurrentRole
}

func (r *Role) WatchRole(zkAddr string, path string) error {
	disc := regd.NewRegDiscoverEx(zkAddr, time.Duration(5*time.Second))
	if err := disc.Start(); nil != err {
		return fmt.Errorf("start discover service failed. error:%v", err)
	}

	eventChan, err := disc.DiscoverService(path)
	if nil != err {
		return fmt.Errorf("start running discover service failed. error:%v", err)
	}

	go func() {
		for event := range eventChan {
			if event.Err != nil {
				blog.Error("received error %s from event chan.", event.Err)
				continue
			}

			if len(event.Server) == 0 {
				continue
			}
			info := types.MetricCollectorInfo{}

			if err := json.Unmarshal([]byte(event.Server[0]), &info); nil != err {
				blog.Errorf("unmarshal zk info failed, skip role handler. err: %v", err)
				continue
			}
			if info.IP == r.localIP && info.MetricPort == r.metricPort {
				r.SetRole(MasterRole)
			} else {
				r.SetRole(SlaveRole)
			}
		}
	}()

	go func() {
		ticker := time.Tick(time.Duration(60 * time.Second))
		for {
			select {
			case <-ticker:
				blog.Infof("-> running in %s role.", r.GetRole())
			}
		}
	}()
	return nil
}
