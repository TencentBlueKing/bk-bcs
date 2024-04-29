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

	"gorm.io/datatypes"
	rawgen "gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gorm/clause"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbclient "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/client"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	sfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Client supplies all the client related operations.
type Client interface {
	// BatchCreateWithTx batch create client instances with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, data []*table.Client) error
	// BatchUpdateWithTx batch update client instances with transaction.
	BatchUpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, data []*table.Client) error
	// ListClientByTuple Query the client list according to multiple fields in
	ListClientByTuple(kit *kit.Kit, data [][]interface{}) ([]*table.Client, error)
	// ListByHeartbeatTimeOnlineState obtain data based on the last heartbeat time and online status
	ListByHeartbeatTimeOnlineState(kit *kit.Kit, heartbeatTime time.Time, onlineState string,
		limit int, id uint32) ([]*table.Client, error)
	// UpdateClientOnlineState Update the online status of the client
	UpdateClientOnlineState(kit *kit.Kit, heartbeatTime time.Time, onlineState string, ids []uint32) error
	// GetClientCountByCondition Get the total according to the condition
	GetClientCountByCondition(kit *kit.Kit, heartbeatTime time.Time, onlineState string) (int64, error)
	// List Obtain client data according to conditions
	List(kit *kit.Kit, bizID, appID uint32, heartbeatTime int64, search *pbclient.ClientQueryCondition,
		order *pbds.ListClientsReq_Order, opt *types.BasePage) ([]*table.Client, int64, error)
	// ListClientGroupByCurrentReleaseID 按当前版本 ID 列出客户端组
	ListClientGroupByCurrentReleaseID(kit *kit.Kit, bizID, appID uint32, heartbeatTime int64,
		search *pbclient.ClientQueryCondition) ([]types.ClientConfigVersionChart, error)
	// ListClientGroupByChangeStatus 按更改状态列出客户端组
	ListClientGroupByChangeStatus(kit *kit.Kit, bizID, appID uint32, heartbeatTime int64,
		search *pbclient.ClientQueryCondition) ([]types.ChangeStatusChart, error)
	// ListClientGroupByFailedReason 按失败原因列出客户端组
	ListClientGroupByFailedReason(kit *kit.Kit, bizID, appID uint32, heartbeatTime int64,
		search *pbclient.ClientQueryCondition) ([]types.FailedReasonChart, error)
	// GetResourceUsage 获取资源使用率
	GetResourceUsage(kit *kit.Kit, bizID, appID uint32, heartbeatTime int64,
		search *pbclient.ClientQueryCondition) (types.ResourceUsage, error)
	// ListClientByIDs 按多个 ID 列出客户端
	ListClientByIDs(kit *kit.Kit, bizID, appID uint32, ids []uint32) ([]*table.Client, error)
	// UpsertHeartbeat 更新插入心跳
	UpsertHeartbeat(kit *kit.Kit, tx *gen.QueryTx, data []*table.Client) error
	// UpsertVersionChange 更新插入版本更改
	UpsertVersionChange(kit *kit.Kit, tx *gen.QueryTx, data []*table.Client) error
}

var _ Client = new(clientDao)

type clientDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// GetResourceUsage 获取资源使用率
func (dao *clientDao) GetResourceUsage(kit *kit.Kit, bizID uint32, appID uint32, heartbeatTime int64,
	search *pbclient.ClientQueryCondition) (types.ResourceUsage, error) {

	m := dao.genQ.Client
	q := dao.genQ.Client.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID),
		m.AppID.Eq(appID))

	var err error
	var items types.ResourceUsage
	var conds []rawgen.Condition
	if search.String() != "" {
		conds, err = dao.handleSearch(kit, bizID, appID, search)
		if err != nil {
			return items, err
		}
	}
	if len(search.GetReleaseChangeStatus()) > 0 {
		conds = append(conds, q.Where(m.ReleaseChangeStatus.In(search.GetReleaseChangeStatus()...)))
	} else {
		conds = append(conds, m.ReleaseChangeStatus.Eq("Success"))
	}
	if heartbeatTime > 0 {
		lastHeartbeatTime := time.Now().Add(time.Duration(-heartbeatTime) * time.Minute)
		conds = append(conds, m.LastHeartbeatTime.Gte(lastHeartbeatTime))
	}
	err = q.Select(m.CpuMaxUsage.Max().As("cpu_max_usage"), m.MemoryMaxUsage.Max().As("memory_max_usage"),
		m.CpuMinUsage.Min().As("cpu_min_usage"), m.CpuAvgUsage.Avg().As("cpu_avg_usage"),
		m.MemoryMinUsage.Min().As("memory_min_usage"), m.MemoryAvgUsage.Avg().As("memory_avg_usage")).
		Where(conds...).
		Scan(&items)
	if err != nil {
		return items, err
	}
	return items, nil
}

// ListClientByIDs 按多个 ID 列出客户端
func (dao *clientDao) ListClientByIDs(kit *kit.Kit, bizID uint32, appID uint32, ids []uint32) ([]*table.Client, error) {
	m := dao.genQ.Client

	result, err := dao.genQ.Client.WithContext(kit.Ctx).
		Select(m.ID, m.BizID, m.AppID, m.ClientType).
		Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ID.In(ids...)).
		Find()

	return result, err
}

// ListClientGroupByFailedReason 按照失败原因列出客户端组
func (dao *clientDao) ListClientGroupByFailedReason(kit *kit.Kit, bizID uint32, appID uint32, heartbeatTime int64,
	search *pbclient.ClientQueryCondition) ([]types.FailedReasonChart, error) {

	m := dao.genQ.Client
	q := dao.genQ.Client.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID),
		m.AppID.Eq(appID), m.ReleaseChangeFailedReason.Neq(""))

	var err error
	var conds []rawgen.Condition
	if search.String() != "" {
		conds, err = dao.handleSearch(kit, bizID, appID, search)
		if err != nil {
			return nil, err
		}
	}
	if len(search.GetReleaseChangeStatus()) > 0 {
		conds = append(conds, q.Where(m.ReleaseChangeStatus.In(search.GetReleaseChangeStatus()...)))
	} else {
		conds = append(conds, m.ReleaseChangeStatus.Eq("Failed"))
	}
	if heartbeatTime > 0 {
		lastHeartbeatTime := time.Now().Add(time.Duration(-heartbeatTime) * time.Minute)
		conds = append(conds, m.LastHeartbeatTime.Gte(lastHeartbeatTime))
	}
	var items []types.FailedReasonChart
	err = q.Select(m.ReleaseChangeFailedReason, m.ID.Count().As("count")).Where(conds...).
		Group(m.ReleaseChangeFailedReason).
		Scan(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// ListClientGroupByChangeStatus 按更改状态列出客户端组
func (dao *clientDao) ListClientGroupByChangeStatus(kit *kit.Kit, bizID uint32, appID uint32, heartbeatTime int64,
	search *pbclient.ClientQueryCondition) ([]types.ChangeStatusChart, error) {

	m := dao.genQ.Client
	q := dao.genQ.Client.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID),
		m.AppID.Eq(appID))

	var err error
	var conds []rawgen.Condition
	if search.String() != "" {
		conds, err = dao.handleSearch(kit, bizID, appID, search)
		if err != nil {
			return nil, err
		}
	}
	if len(search.GetReleaseChangeStatus()) > 0 {
		conds = append(conds, q.Where(m.ReleaseChangeFailedReason.In(search.GetReleaseChangeStatus()...)))
	} else {
		conds = append(conds, m.ReleaseChangeStatus.In("Failed", "Success"))
	}
	if heartbeatTime > 0 {
		lastHeartbeatTime := time.Now().Add(time.Duration(-heartbeatTime) * time.Minute)
		conds = append(conds, m.LastHeartbeatTime.Gte(lastHeartbeatTime))
	}
	var items []types.ChangeStatusChart
	err = q.Select(m.ReleaseChangeStatus, m.ID.Count().As("count")).Where(conds...).
		Group(m.ReleaseChangeStatus).
		Scan(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// ListClientGroupByCurrentReleaseID 通过当前版本ID统计数量
func (dao *clientDao) ListClientGroupByCurrentReleaseID(kit *kit.Kit, bizID uint32, appID uint32, heartbeatTime int64,
	search *pbclient.ClientQueryCondition) ([]types.ClientConfigVersionChart, error) {
	m := dao.genQ.Client
	q := dao.genQ.Client.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.CurrentReleaseID.Neq(0))
	var err error
	var conds []rawgen.Condition
	if heartbeatTime > 0 {
		lastHeartbeatTime := time.Now().Add(time.Duration(-heartbeatTime) * time.Minute)
		conds = append(conds, m.LastHeartbeatTime.Gte(lastHeartbeatTime))
	}
	if search.String() != "" {
		conds, err = dao.handleSearch(kit, bizID, appID, search)
		if err != nil {
			return nil, err
		}
	}
	if len(search.GetReleaseChangeStatus()) > 0 {
		conds = append(conds, q.Where(m.ReleaseChangeStatus.In(search.GetReleaseChangeStatus()...)))
	}

	var items []types.ClientConfigVersionChart
	err = q.Select(m.CurrentReleaseID, m.ID.Count().As("count")).Where(conds...).
		Group(m.CurrentReleaseID).
		Scan(&items)
	if err != nil {
		return nil, err
	}
	return items, nil
}

// List Obtain client data according to conditions
func (dao *clientDao) List(kit *kit.Kit, bizID, appID uint32, heartbeatTime int64,
	search *pbclient.ClientQueryCondition, order *pbds.ListClientsReq_Order,
	opt *types.BasePage) ([]*table.Client, int64, error) {

	m := dao.genQ.Client
	q := dao.genQ.Client.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID))

	var err error
	var conds []rawgen.Condition
	if search.String() != "" {
		conds, err = dao.handleSearch(kit, bizID, appID, search)
		if err != nil {
			return nil, 0, err
		}
	}
	if len(search.GetReleaseChangeStatus()) > 0 {
		conds = append(conds, q.Where(m.ReleaseChangeStatus.In(search.GetReleaseChangeStatus()...)))
	}

	if heartbeatTime > 0 {
		lastHeartbeatTime := time.Now().Add(time.Duration(-heartbeatTime) * time.Minute)
		conds = append(conds, m.LastHeartbeatTime.Gte(lastHeartbeatTime))
	}

	var exprs []field.Expr
	if order.String() != "" {
		exprs = dao.handleOrder(order)
	} else {
		exprs = append(exprs, m.ID.Desc())
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
func (dao *clientDao) handleSearch(kit *kit.Kit, bizID, appID uint32, search *pbclient.ClientQueryCondition) ( // nolint
	[]rawgen.Condition, error) {

	var conds []rawgen.Condition
	m := dao.genQ.Client
	q := dao.genQ.Client.WithContext(kit.Ctx)
	rs := dao.genQ.Release
	ce := dao.genQ.ClientEvent
	if len(search.GetIp()) > 0 {
		conds = append(conds, q.Where(m.Ip.Like("%"+search.GetIp()+"%")))
	}
	if len(search.GetUid()) > 0 {
		conds = append(conds, q.Where(m.UID.Like("%"+search.GetUid()+"%")))
	}

	if len(search.GetCurrentReleaseName()) > 0 {
		var item []struct {
			ID uint32
		}
		err := rs.WithContext(kit.Ctx).Select(rs.ID).Where(rs.BizID.Eq(bizID), rs.AppID.Eq(appID),
			rs.Name.Like("%"+search.GetCurrentReleaseName()+"%")).Scan(&item)
		if err != nil {
			return conds, err
		}
		releaseID := []uint32{}
		for _, v := range item {
			releaseID = append(releaseID, v.ID)
		}
		conds = append(conds, q.Where(m.CurrentReleaseID.In(releaseID...)))
	}

	// 目标版本查询
	// 先获取releaseID, 根据 target_release_id = releaseID 获取 client_events 中的 client_id
	if len(search.GetTargetReleaseName()) > 0 {
		var item []struct {
			ID uint32
		}
		err := rs.WithContext(kit.Ctx).Select(rs.ID).Where(rs.BizID.Eq(bizID), rs.AppID.Eq(appID),
			rs.Name.Like("%"+search.GetTargetReleaseName()+"%")).Scan(&item)
		if err != nil {
			return conds, err
		}
		releaseID := []uint32{}
		for _, v := range item {
			releaseID = append(releaseID, v.ID)
		}
		var clientEvent []struct {
			ClientID uint32
		}
		err = ce.WithContext(kit.Ctx).Select(ce.ClientID).Where(ce.TargetReleaseID.In(releaseID...)).
			Group(ce.ClientID).Scan(&clientEvent)
		if err != nil {
			return conds, err
		}
		cid := []uint32{}
		for _, v := range clientEvent {
			cid = append(cid, v.ClientID)
		}
		conds = append(conds, q.Where(m.ID.In(cid...)))
	}

	// 处理拉取时间
	if search.GetPullTime() != "" {
		starTime, err := time.Parse("2006-01-02", search.GetPullTime())
		if err != nil {
			return nil, err
		}
		// 设置时分秒为 00:00:00
		starTime = time.Date(starTime.Year(), starTime.Month(), starTime.Day(), 0, 0, 0, 0, starTime.UTC().Location())
		endTime, err := time.Parse("2006-01-02", search.GetPullTime())
		// 设置时分秒为 23:59:59
		endTime = time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 0, endTime.UTC().Location())
		if err != nil {
			return nil, err
		}
		var clientEvent []struct {
			ClientID uint32
		}
		err = ce.WithContext(kit.Ctx).Select(ce.ClientID).Where(ce.StartTime.Gte(starTime), ce.EndTime.Lte(endTime)).
			Group(ce.ClientID).Scan(&clientEvent)
		if err != nil {
			return conds, err
		}
		cid := []uint32{}
		for _, v := range clientEvent {
			cid = append(cid, v.ClientID)
		}
		conds = append(conds, q.Where(m.ID.In(cid...)))
	}

	if len(search.GetReleaseChangeStatus()) > 0 {
		conds = append(conds, q.Where(m.ReleaseChangeStatus.In(search.GetReleaseChangeStatus()...)))
	}

	if search.GetLabel() != nil && len(search.GetLabel().GetFields()) != 0 {
		for k, v := range search.GetLabel().GetFields() {
			conds = append(conds, rawgen.Cond(datatypes.JSONQuery("labels").Equals(v.AsInterface(), k))...)
		}
	}

	if search.GetAnnotations() != nil && len(search.GetAnnotations().GetFields()) != 0 {
		for k, v := range search.GetLabel().GetFields() {
			conds = append(conds, rawgen.Cond(datatypes.JSONQuery("annotations").Equals(v.AsInterface(), k))...)
		}
	}

	if len(search.GetOnlineStatus()) > 0 {
		conds = append(conds, q.Where(m.OnlineStatus.In(search.GetOnlineStatus()...)))
	}

	if len(search.GetClientVersion()) > 0 {
		conds = append(conds, q.Where(m.ClientVersion.Like("%"+search.GetClientVersion()+"%")))
	}

	if len(search.GetClientType()) > 0 {
		conds = append(conds, q.Where(m.ClientType.Eq(search.GetClientType())))
	}

	if len(search.GetFailedReason()) > 0 {
		conds = append(conds, q.Where(m.ReleaseChangeFailedReason.Eq(search.GetFailedReason())))
	}

	return conds, nil
}

// 处理排序
func (dao *clientDao) handleOrder(order *pbds.ListClientsReq_Order) []field.Expr {
	var exprs []field.Expr
	m := dao.genQ.Client

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

// GetClientCountByCondition Get the total according to the condition
func (dao *clientDao) GetClientCountByCondition(kit *kit.Kit, heartbeatTime time.Time,
	onlineState string) (int64, error) {
	m := dao.genQ.Client
	count, err := dao.genQ.Client.WithContext(kit.Ctx).
		Where(m.LastHeartbeatTime.Lt(heartbeatTime)).
		Where(m.OnlineStatus.Eq(onlineState)).Count()
	return count, err
}

// UpdateClientOnlineState Update the online status of the client
func (dao *clientDao) UpdateClientOnlineState(kit *kit.Kit, heartbeatTime time.Time,
	onlineState string, ids []uint32) error {
	m := dao.genQ.Client
	_, err := dao.genQ.Client.WithContext(kit.Ctx).
		Where(m.ID.In(ids...)).
		Where(m.LastHeartbeatTime.Lt(heartbeatTime)).
		Where(m.OnlineStatus.Eq(onlineState)).
		Update(m.OnlineStatus, sfs.Offline.String())
	return err
}

// ListByHeartbeatTimeOnlineState obtain data based on the last heartbeat time and online status
func (dao *clientDao) ListByHeartbeatTimeOnlineState(kit *kit.Kit, heartbeatTime time.Time,
	onlineState string, limit int, id uint32) ([]*table.Client, error) {
	m := dao.genQ.Client
	q := dao.genQ.Client.WithContext(kit.Ctx)
	find, err := q.Select(m.ID, m.LastHeartbeatTime, m.OnlineStatus).
		Where(m.ID.Gt(id)).
		Where(m.LastHeartbeatTime.Lt(heartbeatTime)).
		Where(m.OnlineStatus.Eq(onlineState)).
		Limit(limit).
		Find()
	return find, err
}

// ListClientByTuple Query the client list according to multiple fields in
// data Example {{1, 1,"uid1"}, {2, 2,"uid2"}}
// SELECT * FROM `client` WHERE (`biz_id`, `app_id`,`uid`) IN ((1,1,"uid1"),(2,2,'uid2'));
func (dao *clientDao) ListClientByTuple(kit *kit.Kit, data [][]interface{}) ([]*table.Client, error) {
	m := dao.genQ.Client
	return dao.genQ.Client.WithContext(kit.Ctx).
		Select(m.ID, m.BizID, m.AppID, m.UID, m.FirstConnectTime, m.CurrentReleaseID).
		Where(m.WithContext(kit.Ctx).Columns(m.BizID, m.AppID, m.UID).
			In(field.Values(data))).
		Find()
}

// BatchCreateWithTx batch create client instances with transaction.
func (dao *clientDao) BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, data []*table.Client) error {
	// generate an config item id and update to config item.
	if len(data) == 0 {
		return nil
	}
	ids, err := dao.idGen.Batch(kit, table.ClientTable, len(data))
	if err != nil {
		return err
	}
	for i, item := range data {
		if err := item.ValidateCreate(); err != nil {
			return err
		}
		item.ID = ids[i]
	}

	return tx.Client.WithContext(kit.Ctx).CreateInBatches(data, 500)
}

// BatchUpdateWithTx batch update client instances with transaction.
func (dao *clientDao) BatchUpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, data []*table.Client) error {
	if len(data) == 0 {
		return nil
	}
	return tx.Client.WithContext(kit.Ctx).Save(data...)
}

// UpsertHeartbeat 更新插入心跳
func (dao *clientDao) UpsertHeartbeat(kit *kit.Kit, tx *gen.QueryTx, data []*table.Client) error {

	q := tx.Client.WithContext(kit.Ctx)
	return q.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "biz_id"}, {Name: "app_id"}, {Name: "uid"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"online_status", "last_heartbeat_time", "client_version", "ip", "annotations",
			"release_change_status",
			"cpu_usage", "cpu_max_usage", "cpu_min_usage", "cpu_avg_usage",
			"memory_usage", "memory_max_usage", "memory_min_usage", "memory_avg_usage",
		}),
	}).CreateInBatches(data, 500)
}

// UpsertVersionChange 更新插入版本更改
func (dao *clientDao) UpsertVersionChange(kit *kit.Kit, tx *gen.QueryTx, data []*table.Client) error {

	q := tx.Client.WithContext(kit.Ctx)
	return q.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "biz_id"}, {Name: "app_id"}, {Name: "uid"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"online_status", "last_heartbeat_time", "client_version", "client_type", "ip", "labels", "annotations",
			"current_release_id", "target_release_id", "specific_failed_reason",
			"release_change_status", "release_change_failed_reason", "failed_detail_reason",
			"cpu_usage", "cpu_max_usage", "cpu_min_usage", "cpu_avg_usage",
			"memory_usage", "memory_max_usage", "memory_min_usage", "memory_avg_usage",
		}),
	}).CreateInBatches(data, 500)
}
