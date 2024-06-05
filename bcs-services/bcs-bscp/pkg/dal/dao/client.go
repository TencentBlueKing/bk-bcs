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
	"errors"
	"fmt"
	"regexp"
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

	if heartbeatTime > 0 {
		lastHeartbeatTime := time.Now().UTC().Add(time.Duration(-heartbeatTime) * time.Minute)
		conds = append(conds, q.Where(m.LastHeartbeatTime.Gte(lastHeartbeatTime)))
	}

	// 过滤最小0值
	conds = append(conds, q.Where(m.CpuMinUsage.Neq(0), m.MemoryMinUsage.Neq(0)))
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

	// 默认搜索失败
	if len(search.GetReleaseChangeStatus()) == 0 {
		conds = append(conds, q.Where(m.ReleaseChangeStatus.Eq(string(table.Failed))))
	}

	if heartbeatTime > 0 {
		lastHeartbeatTime := time.Now().UTC().Add(time.Duration(-heartbeatTime) * time.Minute)
		conds = append(conds, q.Where(m.LastHeartbeatTime.Gte(lastHeartbeatTime)))
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

	// 默认搜索失败和成功
	if len(search.GetReleaseChangeStatus()) == 0 {
		conds = append(conds, q.Where(m.ReleaseChangeStatus.In(string(table.Failed), string(table.Success))))
	}

	if heartbeatTime > 0 {
		lastHeartbeatTime := time.Now().UTC().Add(time.Duration(-heartbeatTime) * time.Minute)
		conds = append(conds, q.Where(m.LastHeartbeatTime.Gte(lastHeartbeatTime)))
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
		lastHeartbeatTime := time.Now().UTC().Add(time.Duration(-heartbeatTime) * time.Minute)
		conds = append(conds, q.Where(m.LastHeartbeatTime.Gte(lastHeartbeatTime)))
	}

	if search.String() != "" {
		conds, err = dao.handleSearch(kit, bizID, appID, search)
		if err != nil {
			return nil, err
		}
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

	if heartbeatTime > 0 {
		lastHeartbeatTime := time.Now().UTC().Add(time.Duration(-heartbeatTime) * time.Minute)
		conds = append(conds, q.Where(m.LastHeartbeatTime.Gte(lastHeartbeatTime)))
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

	// 根据IP搜索
	if len(search.GetIp()) > 0 {
		ips := replaceWithPipe(search.GetIp())
		for i, v := range ips {
			if i == 0 {
				q = q.Where(m.Ip.Like("%" + v + "%"))
			} else {
				q = q.Or(m.Ip.Like("%" + v + "%"))
			}
		}
		conds = append(conds, q)
	}

	// 根据客户端uid搜索
	if len(search.GetUid()) > 0 {
		uids := replaceWithPipe(search.GetUid())
		for i, v := range uids {
			if i == 0 {
				q = q.Where(m.UID.Like("%" + v + "%"))
			} else {
				q = q.Or(m.UID.Like("%" + v + "%"))
			}
		}
		conds = append(conds, q)
	}

	// 根据当前版本名称搜索
	if len(search.GetCurrentReleaseName()) > 0 {
		rns := replaceWithPipe(search.GetCurrentReleaseName())
		rq := rs.WithContext(kit.Ctx).Select(rs.ID).Where(rs.BizID.Eq(bizID), rs.AppID.Eq(appID))
		for i, v := range rns {
			if i == 0 {
				rq = rq.Where(rs.Name.Like("%" + v + "%"))
			} else {
				rq = rq.Or(rs.Name.Like("%" + v + "%"))
			}
		}

		var item []struct {
			ID uint32
		}
		err := rq.Where(rq).Scan(&item)
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
	var clientEventConds []rawgen.Condition
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

		clientEventConds = append(clientEventConds, q.Where(ce.TargetReleaseID.In(releaseID...)))
	}

	// 根据客户端时间表中的开始和结束时间搜索
	if search.GetStartPullTime() != "" {
		starTime, err := parseTime(search.GetStartPullTime())
		if err != nil {
			return nil, err
		}
		clientEventConds = append(clientEventConds, q.Where(ce.StartTime.Gte(starTime)))
	}
	if search.GetEndPullTime() != "" {
		endTime, err := parseTime(search.GetEndPullTime())
		if err != nil {
			return nil, err
		}
		clientEventConds = append(clientEventConds, q.Where(ce.EndTime.Lte(endTime)))
	}

	if len(clientEventConds) > 0 {
		var clientEvent []struct {
			ClientID uint32
		}
		err := ce.WithContext(kit.Ctx).Select(ce.ClientID).Where(clientEventConds...).
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

	// 根据变更状态搜索
	if len(search.GetReleaseChangeStatus()) > 0 {
		conds = append(conds, q.Where(m.ReleaseChangeStatus.In(search.GetReleaseChangeStatus()...)))
	}

	// 根据标签搜索
	// 支持多个 key:valul 以及 key
	// key 会存在各种格式：符号、数字、中文等，只要使用双引号包围就可以
	if search.GetLabel() != nil && len(search.GetLabel().GetFields()) != 0 {
		for k, v := range search.GetLabel().GetFields() {
			if k != "" {
				ks := replaceWithPipe(k)
				vs := replaceWithPipe(v.GetStringValue())

				// "label":{"app":"test|test2"}
				// JSON_EXTRACT(`labels`,'$."app"') = 'test' OR JSON_EXTRACT(`labels`,'$."app"') = 'test2'
				if len(ks) == 1 && v.GetStringValue() != "" {
					for i, v := range vs {
						if i == 0 {
							q = q.Where(rawgen.Cond(datatypes.JSONQuery("labels").Equals(v, fmt.Sprintf(`"%s"`, k)))...)
						} else {
							q = q.Or(rawgen.Cond(datatypes.JSONQuery("labels").Equals(v, fmt.Sprintf(`"%s"`, k)))...)
						}
					}
				}

				// label":{"app|env":""}
				// JSON_EXTRACT(`labels`,'$."app"') IS NOT NULL OR JSON_EXTRACT(`labels`,'$."env"') IS NOT NULL
				if len(ks) >= 1 && v.GetStringValue() == "" {
					for i, v := range ks {
						if i == 0 {
							q = q.Where(rawgen.Cond(datatypes.JSONQuery("labels").HasKey(fmt.Sprintf(`"%s"`, v)))...)
						} else {
							q = q.Or(rawgen.Cond(datatypes.JSONQuery("labels").HasKey(fmt.Sprintf(`"%s"`, v)))...)
						}
					}
				}

				// "label":{"app|env":"test|test2"}
				// JSON_EXTRACT(`labels`,'$."app"') = 'test' OR JSON_EXTRACT(`labels`,'$."env"') = 'test2'
				// "label":{"app|env":"test"}
				// JSON_EXTRACT(`labels`,'$."app"') = 'test' OR JSON_EXTRACT(`labels`,'$."env"') IS NOT NULL
				// "label":{"app|env":"test|test1|test2"}  直接忽略test2
				if len(ks) >= 1 && v.GetStringValue() != "" {
					for i, k := range ks {
						if i == 0 {
							q = q.Where(rawgen.Cond(datatypes.JSONQuery("labels").Equals(vs[i], fmt.Sprintf(`"%s"`, k)))...)
						} else {
							if i < len(vs) && vs[i] != "" {
								q = q.Or(rawgen.Cond(datatypes.JSONQuery("labels").Equals(vs[i], fmt.Sprintf(`"%s"`, k)))...)
							} else {
								q = q.Or(rawgen.Cond(datatypes.JSONQuery("labels").HasKey(fmt.Sprintf(`"%s"`, k)))...)
							}
						}
					}
				}
				conds = append(conds, q)
			}
		}
	}

	// 根据附加标签搜索
	if search.GetAnnotations() != nil && len(search.GetAnnotations().GetFields()) != 0 {
		for k, v := range search.GetLabel().GetFields() {
			if k != "" {
				conds = append(conds, q.Where(rawgen.Cond(datatypes.JSONQuery("annotations").
					Equals(v.AsInterface(), fmt.Sprintf(`"%s"`, k)))...))
			}
		}
	}

	// 根据在线状态搜索
	if len(search.GetOnlineStatus()) > 0 {
		conds = append(conds, q.Where(m.OnlineStatus.In(search.GetOnlineStatus()...)))
	}

	// 根据客户端版本搜索
	if len(search.GetClientVersion()) > 0 {
		cvs := replaceWithPipe(search.GetClientVersion())
		for i, v := range cvs {
			if i == 0 {
				q = q.Where(m.ClientVersion.Like("%" + v + "%"))
			} else {
				q = q.Or(m.ClientVersion.Like("%" + v + "%"))
			}
		}
		conds = append(conds, q)
	}

	// 根据客户端类型搜索
	if len(search.GetClientType()) > 0 {
		conds = append(conds, q.Where(m.ClientType.Eq(search.GetClientType())))
	}

	// 根据失败原因搜索
	if len(search.GetFailedReason()) > 0 {
		conds = append(conds, q.Where(m.ReleaseChangeFailedReason.Eq(search.GetFailedReason())))
	}

	return conds, nil
}

// 定义一个正则表达式，匹配换行、半角逗号、空字符（空格、\t等）、分号
func replaceWithPipe(input string) []string {
	re := regexp.MustCompile(`[\s,;]+`)
	// 使用正则表达式替换匹配到的部分为竖线
	result := re.ReplaceAllString(input, "|")
	// 去掉结果字符串开头和结尾的竖线
	result = strings.Trim(result, "|")
	if result == "" {
		return []string{}
	}
	// 使用 strings.Split 函数按竖线分隔字符串
	parts := strings.Split(result, "|")
	return parts
}

func parseTime(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, errors.New("time cannot be empty")
	}
	t, err := time.ParseInLocation(time.RFC3339, timeStr, time.UTC)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
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
			"release_change_status", "labels",
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
