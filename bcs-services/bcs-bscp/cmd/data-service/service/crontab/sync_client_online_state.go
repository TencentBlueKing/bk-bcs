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

// Package crontab example Synchronize the online status of the client
package crontab

import (
	"context"
	"sync"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/dao"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
)

const (
	defaultSyncClientStateInterval = 60 * time.Second
)

// NewSyncClientOnlineState init client online state
func NewSyncClientOnlineState(set dao.Set, sd serviced.Service) ClientOnlineState {
	return ClientOnlineState{
		set:   set,
		state: sd,
	}
}

// ClientOnlineState xxx
type ClientOnlineState struct {
	set   dao.Set
	state serviced.Service
	mutex sync.Mutex
}

// Run the sync client online state task
func (c *ClientOnlineState) Run() {
	logs.Infof("example Start an online synchronization task for the client")
	notifier := shutdown.AddNotifier()
	go func() {
		ticker := time.NewTicker(defaultSyncClientStateInterval)
		defer ticker.Stop()
		for {
			kt := kit.New()
			ctx, cancel := context.WithCancel(kt.Ctx)
			kt.Ctx = ctx

			select {
			case <-notifier.Signal:
				logs.Infof("stop sync client online status success")
				cancel()
				notifier.Done()
				return
			case <-ticker.C:
				if !c.state.IsMaster() {
					logs.Infof("current service instance is slave, skip sync client online status")
					continue
				}
				logs.Infof("starts to synchronize the client online status")
				c.syncClientOnlineState(kt)
			}
		}
	}()
}

// sync the online status of the client
func (c *ClientOnlineState) syncClientOnlineState(kt *kit.Kit) {
	c.mutex.Lock()
	defer func() {
		c.mutex.Unlock()
	}()
	heartbeatTime := time.Now().Add(-60 * time.Second)
	onlineStatus := "online"
	var page uint32

	limit := 100
	count, err := c.set.Client().GetClientCountByCondition(kt, heartbeatTime, onlineStatus)
	if err != nil {
		return
	}
	if count == 0 {
		logs.Infof("there is no data to process,rid: %s, heartbeatTime: %s - onlineStatus: %s", kt.Rid,
			heartbeatTime, onlineStatus)
		return
	}
	listLen := int(count)
	for i := 0; i < listLen; i += limit {
		list, err := c.set.Client().ListByHeartbeatTimeOnlineState(kt, heartbeatTime, onlineStatus, limit, page)
		if err != nil {
			logs.Errorf("get client data failed, rid: %s, heartbeatTime: %s, page: %d, err: %s", kt.Rid,
				heartbeatTime, page, err.Error())
			return
		}
		if len(list) == 0 {
			logs.Infof("there is no data to process, rid: %s, heartbeatTime: %s, page: %s", kt.Rid,
				heartbeatTime, page)
			return
		}
		page = list[len(list)-1].ID
		err = c.updateClientOnlineState(kt, list, heartbeatTime, onlineStatus)
		if err != nil {
			logs.Errorf("update client online state failed, rid: %s, heartbeatTime: %s, page: %d, err: %s", kt.Rid,
				heartbeatTime, page, err.Error())
			return
		}
	}
}

// Update the online status of the client
func (c *ClientOnlineState) updateClientOnlineState(kt *kit.Kit, list []*table.Client,
	heartbeatTime time.Time, onlineStatus string) error {
	ids := []uint32{}
	for _, item := range list {
		ids = append(ids, item.ID)
	}
	err := c.set.Client().UpdateClientOnlineState(kt, heartbeatTime, onlineStatus, ids)
	if err != nil {
		return err
	}
	return nil
}
