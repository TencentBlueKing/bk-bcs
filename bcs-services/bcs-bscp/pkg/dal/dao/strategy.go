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

package dao

import (
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// Strategy supplies all the Strategy related operations.
type Strategy interface {
	// Get last strategy.
	GetLast(kit *kit.Kit, bizID, appID, releasedID, strategyID uint32) (*table.Strategy, error)
	// GetStrategyByIDs Get strategy by ids.
	GetStrategyByIDs(kit *kit.Kit, strategyIDs []uint32) ([]*table.Strategy, error)
	// ListStrategyByItsm list strategy by itsm.
	ListStrategyByItsm(kit *kit.Kit) ([]*table.Strategy, error)
	// UpdateByID update strategy kv by id.
	UpdateByID(kit *kit.Kit, tx *gen.QueryTx, strategyID uint32, m map[string]interface{}) error
	// UpdateByIDs update strategy kv by ids
	UpdateByIDs(
		kit *kit.Kit, tx *gen.QueryTx, strategyID []uint32, m map[string]interface{}) error
}

var _ Strategy = new(strategyDao)

type strategyDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// GetLast Get strategy kv.
func (dao *strategyDao) GetLast(kit *kit.Kit, bizID, appID, releasedID, strategyID uint32) (*table.Strategy, error) {
	m := dao.genQ.Strategy
	temp := m.WithContext(kit.Ctx).Where(
		m.BizID.Eq(bizID), m.AppID.Eq(appID))
	if releasedID != 0 {
		temp.Where(m.ReleaseID.Eq(releasedID))
	}
	if strategyID != 0 {
		temp.Where(m.ID.Eq(strategyID))
	}
	return temp.Last()
}

// GetStrategyByIDs Get strategy by ids.
func (dao *strategyDao) GetStrategyByIDs(kit *kit.Kit, strategyIDs []uint32) ([]*table.Strategy, error) {
	m := dao.genQ.Strategy
	return m.WithContext(kit.Ctx).Where(m.ID.In(strategyIDs...)).Find()
}

// GetStrategyByIDs Get strategy by ids.
func (dao *strategyDao) ListStrategyByItsm(kit *kit.Kit) ([]*table.Strategy, error) {
	m := dao.genQ.Strategy
	return m.WithContext(kit.Ctx).Where(m.ItsmTicketStatus.Eq(constant.ItsmTicketStatusCreated),
		m.ItsmTicketStateID.Neq(0), m.ItsmTicketSn.Neq(""),
		m.PublishStatus.In(string(table.PendApproval), string(table.PendPublish))).Find()
}

// UpdateByID update strategy kv by id
func (dao *strategyDao) UpdateByID(kit *kit.Kit, tx *gen.QueryTx, strategyID uint32, m map[string]interface{}) error {
	s := tx.Strategy
	_, err := s.WithContext(kit.Ctx).Where(s.ID.Eq(strategyID)).Updates(m)
	return err
}

// UpdateByIDs update strategy kv by ids
func (dao *strategyDao) UpdateByIDs(
	kit *kit.Kit, tx *gen.QueryTx, strategyIDs []uint32, m map[string]interface{}) error {
	if len(strategyIDs) == 0 {
		return nil
	}
	s := tx.Strategy
	_, err := s.WithContext(kit.Ctx).Where(s.ID.In(strategyIDs...)).Updates(m)
	return err
}
