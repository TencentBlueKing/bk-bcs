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
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/data-service/service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/components/itsm"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/enumor"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/dao"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/shutdown"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/serviced"
)

const (
	defaultSyncTicketStatusInterval = 30 * time.Second
)

// NewSyncTicketStatus init sync ticket status
func NewSyncTicketStatus(set dao.Set, sd serviced.Service, srv *service.Service) SyncTicketStatus {
	return SyncTicketStatus{
		set:   set,
		state: sd,
		srv:   srv,
	}
}

// SyncTicketStatus xxx
type SyncTicketStatus struct {
	set   dao.Set
	state serviced.Service
	mutex sync.Mutex
	srv   *service.Service
}

// Run the sync ticket status
func (c *SyncTicketStatus) Run() {
	logs.Infof("start synchronization task for the itsm tickets")
	notifier := shutdown.AddNotifier()
	go func() {
		ticker := time.NewTicker(defaultSyncTicketStatusInterval)
		defer ticker.Stop()
		for {
			kt := kit.New()
			ctx, cancel := context.WithCancel(kt.Ctx)
			kt.Ctx = ctx

			select {
			case <-notifier.Signal:
				logs.Infof("stop sync tickets status success")
				cancel()
				notifier.Done()
				return
			case <-ticker.C:
				if !c.state.IsMaster() {
					logs.Infof("current service instance is slave, skip sync tickets status")
					continue
				}
				logs.Infof("starts to synchronize the tickets status")
				c.syncTicketStatus(kt)
			}
		}
	}()
}

// sync the ticket status
// nolint: funlen
func (c *SyncTicketStatus) syncTicketStatus(kt *kit.Kit) {
	c.mutex.Lock()
	defer func() {
		c.mutex.Unlock()
	}()

	// 获取CREATED、待上线，待审批状态的strategy记录
	strategys, err := c.set.Strategy().ListStrategyByItsm(kt)
	if err != nil {
		logs.Errorf("list strategy by itsm failed: %s", err.Error())
	}
	snList := []string{}
	strategyMap := map[string]*table.Strategy{}
	for _, strategy := range strategys {
		snList = append(snList, strategy.Spec.ItsmTicketSn)
		strategyMap[strategy.Spec.ItsmTicketSn] = strategy
	}

	if len(snList) == 0 {
		return
	}
	tickets, err := itsm.ListTickets(kt.Ctx, snList)
	if err != nil {
		logs.Errorf("list approve itsm tickets %v failed, err: %s", snList, err.Error())
		return
	}

	for _, ticket := range tickets {
		// 正常状态的单据
		if ticket.CurrentStatus == constant.TicketRunningStatu {
			ticketStatus, err := itsm.GetTicketStatus(kt.Ctx, ticket.SN)
			if err != nil {
				logs.Errorf("get ticket status failed, err: %s", err.Error())
				return
			}
			req := &pbds.ApproveReq{
				BizId:         strategyMap[ticket.SN].Attachment.BizID,
				AppId:         strategyMap[ticket.SN].Attachment.AppID,
				ReleaseId:     strategyMap[ticket.SN].Spec.ReleaseID,
				PublishStatus: string(table.PendPublish),
				StrategyId:    strategyMap[ticket.SN].ID,
			}
			// 正常通过或者驳回的单据
			if len(ticketStatus.Data.CurrentSteps) == 0 {
				gan, errResult := itsm.GetApproveNodeResult(kt.Ctx, ticket.SN,
					strategyMap[ticket.SN].Spec.ItsmTicketStateID)
				if errResult != nil {
					logs.Errorf("GetApproveNodeResult failed, err: %s", errResult.Error())
					return
				}
				// itsm驳回了
				if !gan.Data.ApproveResult {
					req.PublishStatus = string(table.RejectedApproval)
					req.Reason = gan.Data.ApproveRemark
				}
				req.ApprovedBy = strings.Split(gan.Data.Processeduser, ",")

			} else {
				// 统计有多少人已经审批通过
				passApprover, errResult := itsm.GetTicketLogsByPass(kt.Ctx, ticket.SN)
				if errResult != nil {
					logs.Errorf("GetTicketLogsByPass failed, err: %s", errResult.Error())
					return
				}
				if len(passApprover) == 0 {
					continue
				}

				req.PublishStatus = string(table.PendApproval)
				req.ApprovedBy = passApprover
			}
			ctx := metadata.NewIncomingContext(kt.Ctx, metadata.MD{
				strings.ToLower(constant.OperateWayKey): []string{string(enumor.API)}, // 从定时任务调用的
			})
			_, err = c.srv.Approve(ctx, req)
			if err != nil {
				logs.Errorf("sync ticket status approve failed, err: %s", err.Error())
				continue
			}
		} else {
			// 其他状态的单据直接撤销
			req := &pbds.ApproveReq{
				BizId:         strategyMap[ticket.SN].Attachment.BizID,
				AppId:         strategyMap[ticket.SN].Attachment.AppID,
				ReleaseId:     strategyMap[ticket.SN].Spec.ReleaseID,
				PublishStatus: string(table.RevokedPublish),
				StrategyId:    strategyMap[ticket.SN].ID,
			}
			ctx := metadata.NewIncomingContext(kt.Ctx, metadata.MD{
				strings.ToLower(constant.UserKey):       []string{ticket.Creator},
				strings.ToLower(constant.OperateWayKey): []string{string(enumor.API)}, // 从定时任务调用的
			})
			_, err := c.srv.Approve(ctx, req)
			if err != nil {
				logs.Errorf("sync ticket status approve failed, err: %s", err.Error())
				continue
			}
		}
	}
}
