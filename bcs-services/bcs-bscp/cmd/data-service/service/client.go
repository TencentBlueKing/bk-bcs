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

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbclient "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/client"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	sfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// BatchUpsertClientMetrics 批量操作client metrics
func (s *Service) BatchUpsertClientMetrics(ctx context.Context, req *pbds.BatchUpsertClientMetricsReq) ( // nolint
	*pbds.BatchUpsertClientMetricsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	var err error
	toCreate := make([]*table.Client, 0)
	clientUpdateData := make([]*pbclient.Client, 0)
	if len(req.GetClientItems()) != 0 {
		toCreate, clientUpdateData, err = s.handleBatchCreateClients(kt, req.GetClientItems())
		if err != nil {
			return nil, err
		}
	}

	tx := s.dao.GenQuery().Begin()
	err = s.dao.Client().BatchCreateWithTx(kt, tx, toCreate)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	createID := make(map[string]uint32)
	for _, item := range toCreate {
		key := fmt.Sprintf("%d-%d-%s", item.Attachment.BizID, item.Attachment.AppID, item.Attachment.UID)
		createID[key] = item.ID
	}

	updateID := make(map[string]uint32)
	if len(clientUpdateData) > 0 {
		s.updatePrimaryKey(clientUpdateData, createID)
		// 根据类型更新对应字段
		heartbeatData := []*table.Client{}
		versionChangeData := []*table.Client{}
		for _, item := range clientUpdateData {
			key := fmt.Sprintf("%d-%d-%s", item.Attachment.BizId, item.Attachment.AppId, item.Attachment.Uid)
			updateID[key] = item.Id
			switch item.MessageType {
			case "Heartbeat":
				heartbeatData = append(heartbeatData, &table.Client{
					ID:         item.Id,
					Attachment: item.Attachment.ClientAttachment(),
					Spec:       item.Spec.ClientSpec(),
				})
			case "VersionChange":
				versionChangeData = append(versionChangeData, &table.Client{
					ID:         item.Id,
					Attachment: item.Attachment.ClientAttachment(),
					Spec:       item.Spec.ClientSpec(),
				})
			}
		}
		if len(heartbeatData) != 0 {
			err = s.dao.Client().BatchUpdateSelectFieldTx(kt, tx, sfs.Heartbeat, heartbeatData)
			if err != nil {
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
				}
				return nil, err
			}
		}
		if len(versionChangeData) != 0 {
			err = s.dao.Client().BatchUpdateSelectFieldTx(kt, tx, sfs.VersionChangeMessage, versionChangeData)
			if err != nil {
				if rErr := tx.Rollback(); rErr != nil {
					logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
				}
				return nil, err
			}
		}
	}

	if len(req.GetClientEventItems()) != 0 {
		mergedMap := make(map[string]uint32)
		for key, value := range createID {
			mergedMap[key] = value
		}
		for key, value := range updateID {
			if _, exists := mergedMap[key]; !exists {
				mergedMap[key] = value
			}
		}

		err = s.doBatchCreateClientEvents(kt, tx, req.GetClientEventItems(), mergedMap)
		if err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}
	return &pbds.BatchUpsertClientMetricsResp{}, nil
}

func (s *Service) handleBatchCreateClients(kt *kit.Kit, clients []*pbclient.Client) (toCreate []*table.Client,
	clientUpdateData []*pbclient.Client, err error) {
	data := [][]interface{}{}
	oldData := make(map[string]*table.Client)

	clientCreateData := []*pbclient.Client{}
	clientUpdateData = []*pbclient.Client{}

	for _, item := range clients {
		data = append(data, []interface{}{item.Attachment.BizId, item.Attachment.AppId, item.Attachment.Uid})
	}
	tuple, err := s.dao.Client().ListClientByTuple(kt, data)
	if err != nil {
		return nil, nil, err
	}
	if len(tuple) != 0 {
		for _, item := range tuple {
			key := fmt.Sprintf("%d-%d-%s", item.Attachment.BizID, item.Attachment.AppID, item.Attachment.UID)
			oldData[key] = item
		}
	}

	// 以心跳时间排序时间asc
	sort.Slice(clients, func(i, j int) bool {
		return clients[i].Spec.LastHeartbeatTime.AsTime().Before(clients[j].Spec.LastHeartbeatTime.AsTime())
	})
	for _, item := range clients {
		key := fmt.Sprintf("%d-%d-%s", item.Attachment.BizId, item.Attachment.AppId, item.Attachment.Uid)
		v, ok := oldData[key]
		// 判断数据是否存在
		if !ok {
			clientCreateData = append(clientCreateData, &pbclient.Client{
				Attachment:  item.GetAttachment(),
				Spec:        item.GetSpec(),
				MessageType: item.MessageType,
			})
		} else {
			// 处理第一次连接时间
			item.Spec.FirstConnectTime = timestamppb.New(v.Spec.FirstConnectTime)
			if item.Spec.ReleaseChangeStatus != sfs.Success.String() {
				item.Spec.CurrentReleaseId = v.Spec.CurrentReleaseID
			}
			clientUpdateData = append(clientUpdateData, &pbclient.Client{
				Id:          v.ID, // 只需要填充ID
				Attachment:  item.GetAttachment(),
				Spec:        item.Spec,
				MessageType: item.MessageType,
			})
		}
	}

	// 通过时间排序
	sort.Slice(clientCreateData, func(i, j int) bool {
		return clientCreateData[i].Spec.LastHeartbeatTime.AsTime().Before(clientCreateData[j].Spec.LastHeartbeatTime.AsTime())
	})

	// Client数据会存在同一维度下多个类型的消息
	// 如果同一维度下的数据都存在那么都是更新
	// 如果同一维度下的数据不存在，只允许一条创建其他都是更新操作
	createData := make(map[string]*pbclient.Client)
	updateData := make([]*pbclient.Client, 0)
	for _, item := range clientCreateData {
		key := fmt.Sprintf("%d-%d-%s", item.Attachment.BizId, item.Attachment.AppId, item.Attachment.Uid)
		if _, ok := createData[key]; !ok {
			createData[key] = item
		} else {
			updateData = append(updateData, item)
		}
	}

	// 该数据是最终需要创建的数据
	toCreate = []*table.Client{}
	for _, item := range createData {
		toCreate = append(toCreate, &table.Client{
			Attachment: item.GetAttachment().ClientAttachment(),
			Spec:       item.GetSpec().ClientSpec(),
		})
	}
	clientUpdateData = append(clientUpdateData, updateData...)
	return toCreate, clientUpdateData, nil
}

func (s *Service) updatePrimaryKey(clientData []*pbclient.Client, createID map[string]uint32) {
	for _, item := range clientData {
		key := fmt.Sprintf("%d-%d-%s", item.Attachment.BizId, item.Attachment.AppId, item.Attachment.Uid)
		v, ok := createID[key]
		if ok {
			item.Id = v
		}
	}
}

// ListClients list clients
func (s *Service) ListClients(ctx context.Context, req *pbds.ListClientsReq) (
	*pbds.ListClientsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	items, count, err := s.dao.Client().List(grpcKit, req.BizId, req.AppId,
		req.GetLastHeartbeatTime(),
		req.GetSearch(),
		req.GetOrder(),
		&types.BasePage{
			Start: req.Start,
			Limit: uint(req.Limit),
			All:   req.All,
		})
	if err != nil {
		return nil, err
	}

	// 获取发布版本信息
	releaseIDs := []uint32{}
	for _, v := range items {
		releaseIDs = append(releaseIDs, v.Spec.CurrentReleaseID)
	}

	releases, err := s.dao.Release().ListAllByIDs(grpcKit, releaseIDs, req.BizId)
	if err != nil {
		return nil, err
	}

	releaseNames := map[uint32]string{}
	for _, v := range releases {
		releaseNames[v.ID] = v.Spec.Name
	}
	data := pbclient.PbClients(items)
	for _, v := range data {
		v.Spec.CurrentReleaseName = releaseNames[v.Spec.CurrentReleaseId]
		v.Spec.Resource.CpuUsage = math.Round(v.Spec.Resource.CpuUsage*10) / 10
		v.Spec.Resource.CpuMaxUsage = math.Round(v.Spec.Resource.CpuMaxUsage*10) / 10
		v.Spec.Resource.MemoryUsage /= (1024 * 1024)
		v.Spec.Resource.MemoryMaxUsage /= (1024 * 1024)
	}

	resp := &pbds.ListClientsResp{
		Details: data,
		Count:   uint32(count),
	}

	return resp, nil

}

// clientConfigVersionChart 客户端配置版本图表
func (s *Service) clientConfigVersionChart(kit *kit.Kit, bizID, appID uint32, heartbeatTime int64,
	search *pbclient.ClientQueryCondition) ([]interface{}, error) {

	// 获取客户端当前的配置数据
	items, err := s.dao.Client().ListClientGroupByCurrentReleaseID(kit, bizID, appID, heartbeatTime, search)
	if err != nil {
		return nil, err
	}

	// 获取版本名称以及计算占比
	releaseIDs := []uint32{}
	for _, v := range items {
		if v.CurrentReleaseID > 0 {
			releaseIDs = append(releaseIDs, v.CurrentReleaseID)
		}
	}
	releases, err := s.dao.Release().ListAllByIDs(kit, releaseIDs, bizID)
	if err != nil {
		return nil, err
	}
	releaseNames := map[uint32]string{}
	for _, v := range releases {
		releaseNames[v.ID] = v.Spec.Name
	}

	totalSum := 0
	for _, v := range items {
		totalSum += v.Count
	}

	// 将结果转换为数组结构体 charts
	var charts []interface{}
	for _, v := range items {
		chart := make(map[string]interface{})
		ratio := float64(v.Count) / float64(totalSum)
		chart["current_release_id"] = v.CurrentReleaseID
		chart["current_release_name"] = releaseNames[v.CurrentReleaseID]
		chart["count"] = v.Count
		chart["percent"] = ratio
		charts = append(charts, chart)
	}
	return charts, nil
}

// clientVersionDistribution 客户端组件版本分布
func (s *Service) clientVersionDistribution(kit *kit.Kit, bizID, appID uint32, chartType string, heartbeatTime int64,
	search *pbclient.ClientQueryCondition) (map[string]interface{}, error) {
	items, count, err := s.dao.Client().List(kit, bizID, appID,
		heartbeatTime,
		search,
		&pbds.ListClientsReq_Order{},
		&types.BasePage{
			All: true,
		})
	if err != nil {
		return nil, err
	}

	if count <= 0 {
		return nil, err
	}

	var formatter types.ClientComponentVersionFormat
	if types.ChartType(chartType) == types.Sunburst {
		formatter = types.SunburstFormatter{}
	} else {
		formatter = types.BarFormatter{}
	}
	charts := formatter.Format(items)
	resp := make(map[string]interface{})
	resp["config_version_distribution"] = charts

	// 获取资源使用率
	resourceUsage, err := s.getResourceUsage(kit, bizID, appID, heartbeatTime, search)
	if err != nil {
		return nil, err
	}
	resp["resource_usage"] = resourceUsage

	return resp, nil
}

// clientPullInfo 客户端拉取信息
func (s *Service) clientPullInfo(kit *kit.Kit, bizID, appID uint32, heartbeatTime int64,
	search *pbclient.ClientQueryCondition) (map[string]interface{}, error) {

	changeStatus, err := s.dao.Client().ListClientGroupByChangeStatus(kit, bizID, appID, heartbeatTime, search)
	if err != nil {
		return nil, err
	}

	changeStatusCount := 0
	for _, v := range changeStatus {
		changeStatusCount += v.Count
	}

	// 将结果转换为数组结构体 charts
	var changeStatusCharts []interface{}
	for _, v := range changeStatus {
		chart := make(map[string]interface{})
		ratio := float64(v.Count) / float64(changeStatusCount)
		chart["release_change_status"] = v.ReleaseChangeStatus
		chart["count"] = v.Count
		chart["percent"] = ratio
		changeStatusCharts = append(changeStatusCharts, chart)
	}

	// 获取具体失败的比例
	failedReasons, err := s.dao.Client().ListClientGroupByFailedReason(kit, bizID, appID, heartbeatTime, search)
	if err != nil {
		return nil, err
	}

	failedReasonCount := 0
	for _, v := range failedReasons {
		failedReasonCount += v.Count
	}

	// 将结果转换为数组结构体 charts
	var failedReasonCharts []interface{}
	for _, v := range failedReasons {
		chart := make(map[string]interface{})
		ratio := float64(v.Count) / float64(failedReasonCount)
		chart["release_change_failed_reason"] = v.ReleaseChangeFailedReason
		chart["count"] = v.Count
		chart["percent"] = ratio
		failedReasonCharts = append(failedReasonCharts, chart)
	}

	// 获取最小最大平均时间
	// 通过查询条件获取clientID
	items, count, err := s.dao.Client().List(kit, bizID, appID, heartbeatTime, search,
		&pbds.ListClientsReq_Order{}, &types.BasePage{All: true})
	if err != nil {
		return nil, err
	}
	var ClientID []uint32
	if count > 0 {
		for _, v := range items {
			ClientID = append(ClientID, v.ID)
		}
	}

	time, err := s.dao.ClientEvent().GetMinMaxAvgTime(kit, bizID, appID, ClientID, search.GetReleaseChangeStatus())
	if err != nil {
		return nil, err
	}
	timeChart := make(map[string]interface{})
	timeChart["min"] = time.Min
	timeChart["max"] = time.Max
	timeChart["avg"] = time.Avg

	resp := make(map[string]interface{})
	resp["change_status"] = changeStatusCharts
	resp["failed_reason"] = failedReasonCharts
	resp["time_consuming"] = timeChart

	return resp, nil
}

// getResourceUsage 获取资源使用率cpu、mem
func (s *Service) getResourceUsage(kit *kit.Kit, bizID, appID uint32, heartbeatTime int64,
	search *pbclient.ClientQueryCondition) (map[string]interface{}, error) {

	item, err := s.dao.Client().GetResourceUsage(kit, bizID, appID, heartbeatTime, search)
	if err != nil {
		return nil, err
	}
	usage := make(map[string]interface{})
	usage["cpu_max_usage"] = math.Round(item.CpuMaxUsage*10) / 10
	usage["cpu_min_usage"] = math.Round(item.CpuMinUsage*10) / 10
	usage["cpu_avg_usage"] = math.Round(item.CpuAvgUsage*10) / 10
	usage["memory_max_usage"] = item.MemoryMaxUsage / (1024 * 1024)
	usage["memory_min_usage"] = item.MemoryMinUsage / (1024 * 1024)
	usage["memory_avg_usage"] = item.MemoryAvgUsage / (1024 * 1024)

	return usage, nil
}

// ClientStatisticsAnalyze 客户端配置版本统计、拉取成功率统计、失败原因统计、客户端组件信息统计
func (s *Service) ClientStatisticsAnalyze(ctx context.Context, req *pbds.ClientStatisticsAnalyzeReq) (
	*pbds.ClientStatisticsAnalyzeResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	configVersion, err := s.clientConfigVersionChart(grpcKit, req.GetBizId(), req.GetAppId(),
		req.GetLastHeartbeatTime(), req.GetSearch())
	if err != nil {
		return nil, err
	}

	pullInfo, err := s.clientPullInfo(grpcKit, req.GetBizId(), req.GetAppId(), req.GetLastHeartbeatTime(), req.GetSearch())
	if err != nil {
		return nil, err
	}

	distribution, err := s.clientVersionDistribution(grpcKit, req.GetBizId(), req.GetAppId(), req.GetChartType(),
		req.GetLastHeartbeatTime(), req.GetSearch())
	if err != nil {
		return nil, err
	}

	data := make(map[string]interface{})
	data["client_config_version"] = configVersion
	data["client_pull_info"] = pullInfo
	data["client_component_info_statistics"] = distribution

	details, err := structpb.NewStruct(data)
	if err != nil {
		return nil, err
	}

	return &pbds.ClientStatisticsAnalyzeResp{
		Details: details,
	}, nil
}

// ClientPullTrendAnalyze 客户端拉取数量趋势统计
func (s *Service) ClientPullTrendAnalyze(ctx context.Context, req *pbds.ClientPullTrendAnalyzeReq) (
	*pbds.ClientPullTrendAnalyzeResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	var ClientIDs []uint32
	// 根据搜索条件搜索主表获取clientID
	if req.GetSearch() != nil {
		items, count, err := s.dao.Client().List(grpcKit, req.GetBizId(), req.GetAppId(), req.GetLastHeartbeatTime(),
			req.GetSearch(), &pbds.ListClientsReq_Order{}, &types.BasePage{All: true})
		if err != nil {
			return nil, err
		}
		if count > 0 {
			for _, v := range items {
				ClientIDs = append(ClientIDs, v.ID)
			}
		}
	}

	data, err := s.dao.ClientEvent().GetPullTrend(grpcKit, req.GetBizId(), req.GetAppId(), ClientIDs, req.GetPullTime())
	if err != nil {
		return nil, err
	}

	var ids []uint32
	// 获取每个客户端类型
	for _, v := range data {
		ids = append(ids, v.ClientID)
	}

	clients, err := s.dao.Client().ListClientByIDs(grpcKit, req.GetBizId(), req.GetAppId(), ids)
	if err != nil {
		return nil, err
	}

	ClientTypes := make(map[uint32]string)
	for _, v := range clients {
		ClientTypes[v.ID] = string(v.Spec.ClientType)
	}

	// 统计时间维度下的客户端类型
	typeCountByDate := make(map[string]map[string]int)
	for _, v := range data {
		clientType := ClientTypes[v.ClientID]
		dateKey := v.PullTime.Format("2006/01/02")
		if typeCountByDate[dateKey] == nil {
			typeCountByDate[dateKey] = make(map[string]int)
		}
		typeCountByDate[dateKey][clientType]++
	}
	// 将结果转换为数组结构体 charts
	var charts []interface{}
	for date, counts := range typeCountByDate {
		chart := make(map[string]interface{})
		chart["date"] = date
		chart["total"] = sum(counts)
		for t, count := range counts {
			chart[t] = count
		}
		charts = append(charts, chart)
	}

	resp := make(map[string]interface{})
	resp["pull_trend"] = charts

	details, err := structpb.NewStruct(resp)
	if err != nil {
		return nil, err
	}

	return &pbds.ClientPullTrendAnalyzeResp{
		Details: details,
	}, nil
}

// ClientTagsAndExtraInfoAnalyze 客户端标签、客户端附加信息分布
func (s *Service) ClientTagsAndExtraInfoAnalyze(ctx context.Context, req *pbds.ClientTagsAndExtraInfoAnalyzeReq) (
	*pbds.ClientTagsAndExtraInfoAnalyzeResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	items, _, err := s.dao.Client().List(grpcKit, req.GetBizId(), req.GetAppId(), req.GetLastHeartbeatTime(),
		req.GetSearch(), &pbds.ListClientsReq_Order{}, &types.BasePage{All: true})
	if err != nil {
		return nil, err
	}

	counts := make(map[string]map[string]int)
	total := len(items)

	for _, item := range items {
		lable := map[string]string{}
		_ = json.Unmarshal([]byte(item.Spec.Labels), &lable)
		for key, value := range lable {
			if counts[key] == nil {
				counts[key] = make(map[string]int)
			}
			counts[key][value]++
		}
	}

	resp := make(map[string]interface{})
	for key, value := range counts {
		var items []interface{}
		for k, v := range value {
			items = append(items, map[string]interface{}{
				"key":     key,
				"value":   k,
				"count":   v,
				"percent": float64(v) / float64(total) * 100,
			})
		}
		resp[key] = items
	}

	details, err := structpb.NewStruct(resp)
	if err != nil {
		return nil, err
	}

	return &pbds.ClientTagsAndExtraInfoAnalyzeResp{Details: details}, nil
}

func sum(m map[string]int) int {
	total := 0
	for _, v := range m {
		total += v
	}
	return total
}
