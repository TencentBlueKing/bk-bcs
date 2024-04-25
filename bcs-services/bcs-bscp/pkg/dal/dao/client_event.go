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
	"gorm.io/gorm/clause"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// ClientEvent supplies all the client event related operations.
type ClientEvent interface {
	// BatchCreateWithTx batch create client event instances with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, data []*table.ClientEvent) error
	// ListClientByTuple Query the client list according to multiple fields in
	ListClientByTuple(kit *kit.Kit, data [][]interface{}) ([]*table.ClientEvent, error)
	// List list client event details
	List(kit *kit.Kit, bizID, appID, clientID uint32, startTime, endTime time.Time, searchValue string,
		order *pbds.ListClientEventsReq_Order, opt *types.BasePage) ([]*table.ClientEvent, int64, error)
	// GetMinMaxAvgTime 获取最小最大平均时间
	GetMinMaxAvgTime(kit *kit.Kit, bizID, appID uint32, clientID []uint32, releaseChangeStatus []string) (
		types.MinMaxAvgTimeChart, error)
	// GetPullTrend 获取拉取趋势
	GetPullTrend(kit *kit.Kit, bizID uint32, appID uint32, clientID []uint32, pullTime int64) ([]types.PullTrend, error)
	// UpsertHeartbeat 更新插入心跳
	UpsertHeartbeat(kit *kit.Kit, tx *gen.QueryTx, data []*table.ClientEvent) error
	// UpsertVersionChange 更新插入版本更改
	UpsertVersionChange(kit *kit.Kit, tx *gen.QueryTx, data []*table.ClientEvent) error
}

var _ ClientEvent = new(clientEventDao)

type clientEventDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// GetPullTrend 获取拉取趋势
func (dao *clientEventDao) GetPullTrend(kit *kit.Kit, bizID uint32, appID uint32, clientID []uint32, pullTime int64) (
	[]types.PullTrend, error) {

	m := dao.genQ.ClientEvent
	q := dao.genQ.ClientEvent.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID),
		m.AppID.Eq(appID))

	var conds []rawgen.Condition
	if pullTime > 0 {
		startTime := time.Now().AddDate(0, 0, -int(pullTime)).Truncate(24 * time.Hour)
		conds = append(conds, m.StartTime.Gte(startTime))
	}
	if len(clientID) > 0 {
		conds = append(conds, m.ClientID.In(clientID...))
	}

	var items []types.PullTrend

	err := q.Select(m.ClientID, m.StartTime.Date().As("pull_time")).
		Where(conds...).
		Group(m.ClientID, field.NewField("", "pull_time")).
		Scan(&items)
	return items, err
}

// GetMinMaxAvgTime 获取最小最大平均时间
func (dao *clientEventDao) GetMinMaxAvgTime(kit *kit.Kit, bizID uint32, appID uint32, clientID []uint32,
	releaseChangeStatus []string) (types.MinMaxAvgTimeChart, error) {

	m := dao.genQ.ClientEvent
	q := dao.genQ.ClientEvent.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID),
		m.AppID.Eq(appID)).Where(m.OriginalReleaseID.NeqCol(m.TargetReleaseID))

	var err error
	var items types.MinMaxAvgTimeChart
	var conds []rawgen.Condition
	if len(releaseChangeStatus) > 0 {
		conds = append(conds, q.Where(m.ReleaseChangeStatus.In(releaseChangeStatus...)))
	} else {
		conds = append(conds, m.ReleaseChangeStatus.Eq("Success"))
	}
	if len(clientID) > 0 {
		conds = append(conds, m.ClientID.In(clientID...))
	}
	q = q.Where(conds...)
	err = q.Select(m.TotalSeconds.Max().As("max"), m.TotalSeconds.Min().As("min"), m.TotalSeconds.Avg().As("avg")).
		Group(m.ReleaseChangeFailedReason).
		Scan(&items)

	return items, err
}

// List list client event details
func (dao *clientEventDao) List(kit *kit.Kit, bizID, appID, clientID uint32, startTime, endTime time.Time,
	searchValue string, order *pbds.ListClientEventsReq_Order, opt *types.BasePage) (
	[]*table.ClientEvent, int64, error) {

	m := dao.genQ.ClientEvent
	q := dao.genQ.ClientEvent.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID),
		m.ClientID.Eq(clientID)).Where(m.OriginalReleaseID.NeqCol(m.TargetReleaseID))
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
func (dao *clientEventDao) handleSearch(kit *kit.Kit, bizID, appID uint32, search string) ([]rawgen.Condition, error) { // nolint
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

	return tx.ClientEvent.WithContext(kit.Ctx).CreateInBatches(data, 500)
}

// UpsertHeartbeat 更新插入心跳
func (dao *clientEventDao) UpsertHeartbeat(kit *kit.Kit, tx *gen.QueryTx, data []*table.ClientEvent) error {

	q := tx.ClientEvent.WithContext(kit.Ctx)
	return q.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "biz_id"},
			{Name: "app_id"},
			{Name: "uid"},
			{Name: "cursor_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"download_file_size", "download_file_num", "release_change_status",
		}),
	}).CreateInBatches(data, 500)
}

// UpsertVersionChange 更新插入版本更改
func (dao *clientEventDao) UpsertVersionChange(kit *kit.Kit, tx *gen.QueryTx, data []*table.ClientEvent) error {

	q := tx.ClientEvent.WithContext(kit.Ctx)
	return q.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "biz_id"},
			{Name: "app_id"},
			{Name: "uid"},
			{Name: "cursor_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"client_mode", "original_release_id", "target_release_id", "start_time", "end_time",
			"release_change_status", "release_change_failed_reason", "failed_detail_reason",
			"download_file_size", "download_file_num", "total_seconds", "total_file_size",
			"total_file_num", "download_file_num", "specific_failed_reason",
		}),
	}).CreateInBatches(data, 500)
}
