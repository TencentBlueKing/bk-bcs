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
	// BatchUpdateSelectFieldTx Update selected field
	BatchUpdateSelectFieldTx(kit *kit.Kit, tx *gen.QueryTx, messageType sfs.MessagingType,
		data []*table.Client) error
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
}

var _ Client = new(clientDao)

type clientDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// List Obtain client data according to conditions
func (dao *clientDao) List(kit *kit.Kit, bizID, appID uint32, heartbeatTime int64,
	search *pbclient.ClientQueryCondition, order *pbds.ListClientsReq_Order,
	opt *types.BasePage) ([]*table.Client, int64, error) {

	m := dao.genQ.Client
	q := dao.genQ.Client.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID))

	var err error
	var conds []rawgen.Condition
	if search != nil {
		conds, err = dao.handleSearch(kit, bizID, appID, search)
		if err != nil {
			return nil, 0, err
		}
	}

	if heartbeatTime > 0 {
		lastHeartbeatTime := time.Now().Add(time.Duration(-heartbeatTime) * time.Minute)
		conds = append(conds, m.LastHeartbeatTime.Gte(lastHeartbeatTime))
	}

	var exprs []field.Expr
	if order != nil {
		exprs = dao.handleOrder(order)
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
func (dao *clientDao) handleSearch(kit *kit.Kit, bizID, appID uint32, search *pbclient.ClientQueryCondition) (
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

	if len(search.GetReleaseChangeStatus()) > 0 {
		status := field.NewString(m.TableName(), "release_change_status")
		conds = append(conds, q.Where(status.In(search.GetReleaseChangeStatus()...)))
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
		OnStatus := field.NewString(m.TableName(), "online_status")
		conds = append(conds, q.Where(OnStatus.In(search.GetOnlineStatus()...)))
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

// BatchUpdateSelectFieldTx Update selected field
func (dao *clientDao) BatchUpdateSelectFieldTx(kit *kit.Kit, tx *gen.QueryTx,
	messageType sfs.MessagingType, data []*table.Client) error {
	if len(data) == 0 {
		return nil
	}
	m := dao.genQ.Client
	q := tx.Client.WithContext(kit.Ctx)

	// 根据类型更新字段
	// 拉取状态时没有上报资源信息所以忽略cpu和内存等信息
	switch messageType {
	case sfs.VersionChangeMessage:
		q = q.Omit(m.FirstConnectTime)
	case sfs.Heartbeat:
		q = q.Omit(m.ClientVersion, m.Ip, m.Labels, m.Annotations, m.FirstConnectTime, m.CurrentReleaseID,
			m.TargetReleaseID, m.FailedDetailReason, m.ReleaseChangeFailedReason)
	}

	return q.Save(data...)
}

// BatchUpdateWithTx batch update client instances with transaction.
func (dao *clientDao) BatchUpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, data []*table.Client) error {
	if len(data) == 0 {
		return nil
	}
	return tx.Client.WithContext(kit.Ctx).Save(data...)
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

	return tx.Client.WithContext(kit.Ctx).Save(data...)
}
