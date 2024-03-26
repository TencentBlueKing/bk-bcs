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
	"strings"
	"time"

	rawgen "gorm.io/gen"
	"gorm.io/gen/field"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	sfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// ClientEvent supplies all the client event related operations.
type ClientEvent interface {
	// BatchCreateWithTx batch create client event instances with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, data []*table.ClientEvent) error
	// BatchUpdateSelectFieldTx Update selected field
	BatchUpdateSelectFieldTx(kit *kit.Kit, tx *gen.QueryTx, messageType sfs.MessagingType, data []*table.ClientEvent) error
	// ListClientByTuple Query the client list according to multiple fields in
	ListClientByTuple(kit *kit.Kit, data [][]interface{}) ([]*table.ClientEvent, error)
	// List list client event details
	List(kit *kit.Kit, bizID, appID, clientID uint32, startTime, endTime time.Time, searchValue string,
		order *pbds.ListClientEventsReq_Order, opt *types.BasePage) ([]*table.ClientEvent, int64, error)
}

var _ ClientEvent = new(clientEventDao)

type clientEventDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// List list client event details
func (dao *clientEventDao) List(kit *kit.Kit, bizID, appID, clientID uint32, startTime, endTime time.Time,
	searchValue string, order *pbds.ListClientEventsReq_Order, opt *types.BasePage) (
	[]*table.ClientEvent, int64, error) {

	m := dao.genQ.ClientEvent
	q := dao.genQ.ClientEvent.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ClientID.Eq(clientID))
	var err error
	var conds []rawgen.Condition
	if len(searchValue) > 0 {
		conds, err = dao.handleSearch(kit, bizID, appID, searchValue)
		if err != nil {
			return nil, 0, err
		}
	}

	var exprs []field.Expr
	if order != nil {
		exprs = dao.handleOrder(order)
	}

	zeroTime := time.Time{}
	if startTime != zeroTime {
		conds = append(conds, m.StartTime.Gte(startTime))
	}
	if endTime != zeroTime {
		conds = append(conds, m.EndTime.Lte(endTime))
	}
	d := q.Where(conds...).Order(exprs...)
	if opt.All {
		result, err := d.Find()
		if err != nil {
			return nil, 0, err
		}
		return result, int64(len(result)), err
	}
	return d.FindByPage(opt.Offset(), opt.LimitInt())
}

// 处理搜索
func (dao *clientEventDao) handleSearch(kit *kit.Kit, bizID, appID uint32, search string) ([]rawgen.Condition, error) {
	var conds []rawgen.Condition

	m := dao.genQ.ClientEvent
	q := dao.genQ.ClientEvent.WithContext(kit.Ctx)
	rs := dao.genQ.Release

	status := field.NewString(m.TableName(), "release_change_status")

	var item []struct {
		ID uint32
	}

	err := rs.WithContext(kit.Ctx).Select(rs.ID).Where(rs.BizID.Eq(bizID), rs.AppID.Eq(appID),
		rs.Name.Like("%"+search+"%")).Scan(&item)
	if err != nil {
		return conds, err
	}
	if len(item) > 0 {
		releaseID := []uint32{}
		for _, v := range item {
			releaseID = append(releaseID, v.ID)
		}
		conds = append(conds, q.Or(m.OriginalReleaseID.In(releaseID...)).
			Or(m.TargetReleaseID.In(releaseID...)).Or(status.Eq(search)))
	} else {
		conds = append(conds, q.Or(status.Eq(search)))
	}

	return conds, nil
}

// 处理排序
func (dao *clientEventDao) handleOrder(order *pbds.ListClientEventsReq_Order) []field.Expr {
	var exprs []field.Expr
	m := dao.genQ.ClientEvent

	if len(order.GetDesc()) > 0 {
		desc := strings.Split(order.GetDesc(), ",")
		for _, v := range desc {
			orderCol, ok := m.GetFieldByName(v)
			if ok {
				exprs = append(exprs, orderCol.Desc())
			}
		}
	}
	if len(order.GetAsc()) > 0 {
		asc := strings.Split(order.GetAsc(), ",")
		for _, v := range asc {
			orderCol, ok := m.GetFieldByName(v)
			if ok {
				exprs = append(exprs, orderCol)
			}
		}
	}

	return exprs
}

// ListClientByTuple Query the client list according to multiple fields in
func (dao *clientEventDao) ListClientByTuple(kit *kit.Kit, data [][]interface{}) ([]*table.ClientEvent, error) {
	m := dao.genQ.ClientEvent
	return dao.genQ.ClientEvent.WithContext(kit.Ctx).
		Select(m.ID, m.BizID, m.AppID, m.UID, m.CursorID, m.ClientMode).
		Where(m.WithContext(kit.Ctx).Columns(m.BizID, m.AppID, m.UID, m.CursorID).
			In(field.Values(data))).
		Find()
}

// BatchUpdateSelectFieldTx Update selected field
// 版本变更时所有所有数据都可以更新
// 心跳主要更新下载数和下载文件大小、变更状态
func (dao *clientEventDao) BatchUpdateSelectFieldTx(kit *kit.Kit, tx *gen.QueryTx, messageType sfs.MessagingType,
	data []*table.ClientEvent) error {
	m := dao.genQ.ClientEvent
	q := tx.ClientEvent.WithContext(kit.Ctx)

	// 根据类型更新字段
	switch messageType {
	case sfs.VersionChangeMessage:
		q = q.Omit()
	case sfs.Heartbeat:
		q = q.Omit(m.ClientMode, m.TotalFileNum, m.TotalFileSize, m.TotalSeconds, m.OriginalReleaseID,
			m.TargetReleaseID, m.StartTime, m.EndTime, m.FailedDetailReason, m.ReleaseChangeFailedReason)
	}

	return q.Save(data...)
}

// BatchCreateWithTx batch create client event instances with transaction.
func (dao *clientEventDao) BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, data []*table.ClientEvent) error {
	// generate an config item id and update to config item.
	if len(data) == 0 {
		return nil
	}
	ids, err := dao.idGen.Batch(kit, table.ClientEventTable, len(data))
	if err != nil {
		return err
	}
	for i, item := range data {
		if err := item.ValidateCreate(); err != nil {
			return err
		}
		item.ID = ids[i]
	}
	if err := tx.ClientEvent.WithContext(kit.Ctx).Save(data...); err != nil {
		return err
	}
	return nil
}
