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

// Package event handle publish
package event

import (
	"context"
	"fmt"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/cache-service/service/cache/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/bedis"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/dao"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

const (
	defaultPublishInterval = 1 * time.Second
)

// Publish xxx
type Publish struct {
	set   dao.Set
	state serviced.State
	bds   bedis.Client
	op    client.Interface
}

// NewPublish init publish
func NewPublish(set dao.Set, state serviced.State, bds bedis.Client, op client.Interface) Publish {
	return Publish{
		set:   set,
		state: state,
		bds:   bds,
		op:    op,
	}
}

// Run the publish task
func (cm *Publish) Run() {
	logs.Infof("start publish task")
	notifier := shutdown.AddNotifier()
	go func() {
		ticker := time.NewTicker(defaultPublishInterval)
		defer ticker.Stop()
		for {
			kt := kit.New()
			ctx, cancel := context.WithCancel(kt.Ctx)
			kt.Ctx = ctx

			select {
			case <-notifier.Signal:
				logs.Infof("stop handle client publish data success")
				cancel()
				notifier.Done()
				return
			case <-ticker.C:
				logs.Infof("start handle client publish data")

				if !cm.state.IsMaster() {
					logs.V(2).Infof("this is slave, do not need to handle, skip. rid: %s", kt.Rid)
					time.Sleep(sleepTime)
					continue
				}
				cm.updateStrategy(kt)
			}
		}
	}()
}

// 上线更新状态
// nolint funlen
func (cm *Publish) updateStrategy(kt *kit.Kit) {
	// 统一使用上海时区
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		logs.Errorf("load location failed, err: %s, rid: %s", err.Error(), kt.Rid)
		return
	}
	locateTime := time.Now().In(location)
	publishInfos, err := cm.op.GetPublishTime(kt, locateTime.Unix())
	if err != nil {
		logs.Errorf("get publish time failed, err: %s, rid: %s", err.Error(), kt.Rid)
		return
	}

	if len(publishInfos) == 0 {
		return
	}

	var strategyIds []uint32
	zrems := make(map[string][]string)
	for k, v := range publishInfos {
		strategyIds = append(strategyIds, k)
		zrems[v.Key] = append(zrems[v.Key], fmt.Sprintf("%d", k))
	}

	strategies, err := cm.set.Strategy().GetStrategyByIDs(kt, strategyIds)
	if err != nil {
		logs.Errorf("get strategy by ids failed, err: %s, rid: %s", err.Error(), kt.Rid)
		return
	}

	var manulStrategies []uint32
	var publishStrategies []uint32

	tx := cm.set.GenQuery().Begin()
	defer func() {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
	}()

	for _, v := range strategies {
		// 类型必须是定时上线
		if v.Spec.PublishType == table.Periodically {

			// 刚好到上线时间未审批的情况，更新为手动上线
			if v.Spec.PublishStatus == table.PendApproval {
				manulStrategies = append(manulStrategies, v.ID)
				continue
			}

			// 因环境问题未上线的，要么刚好到上线时间未审批的情况，要么刚好要上线没上线
			if v.Spec.PublishStatus == table.PendPublish {
				// 更新为手动上线
				if v.Revision.UpdatedAt.Unix() > publishInfos[v.ID].PublishTime {
					manulStrategies = append(manulStrategies, v.ID)
					continue
				}
				// 刚好要上线，以及因环境问题刚好要上线没上线
				publishStrategies = append(publishStrategies, v.ID)

				opt := types.PublishOption{
					BizID:     v.Attachment.BizID,
					AppID:     v.Attachment.AppID,
					ReleaseID: v.Spec.ReleaseID,
					All:       false,
				}

				if len(v.Spec.Scope.Groups) == 0 {
					opt.All = true
				}

				err = cm.set.Publish().UpsertPublishWithTx(kt, tx, &opt, v)
				if err != nil {
					logs.Errorf("update publish with tx failed, err: %s, rid: %s", err.Error(), kt.Rid)
					return
				}
			}
		}
	}

	err = cm.set.Strategy().UpdateByIDs(kt, tx, manulStrategies, map[string]interface{}{
		"publish_type": table.Manually,
	})
	if err != nil {
		logs.Errorf("update strategy by ids manually failed, err: %s, rid: %s", err.Error(), kt.Rid)
		return
	}

	err = cm.set.Strategy().UpdateByIDs(kt, tx, publishStrategies, map[string]interface{}{
		"publish_status": table.AlreadyPublish,
		"pub_state":      table.Publishing,
	})
	if err != nil {
		logs.Errorf("update strategy by ids already publish failed, err: %s, rid: %s", err.Error(), kt.Rid)
		return
	}

	// update audit details
	err = cm.set.AuditDao().UpdateByStrategyIDs(kt, tx, publishStrategies, map[string]interface{}{
		"status": table.AlreadyPublish,
	})
	if err != nil {
		logs.Errorf("update audit by strategy ids failed, err: %s, rid: %s", err.Error(), kt.Rid)
		return
	}

	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return
	}

	for k, v := range zrems {
		_, err := cm.bds.ZRem(kt.Ctx, k, v)
		if err != nil {
			logs.Errorf("zrem failed, err: %v, rid: %s", err, kt.Rid)
			return
		}
	}
}
