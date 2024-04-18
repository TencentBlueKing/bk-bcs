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
	"strings"
	"time"

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
func (s *Service) BatchUpsertClientMetrics(ctx context.Context, req *pbds.BatchUpsertClientMetricsReq) (
	*pbds.BatchUpsertClientMetricsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	var err error
	var toCreate []*table.Client
	var toUpdate map[string][]*table.Client

	toCreate, toUpdate, err = s.handleBatchCreateClients(kt, req.GetClientItems())
	if err != nil {
		return nil, err
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

	// 更新时id不能为空
	// 更新 client_event 时需要clientID
	for _, data := range toUpdate {
		for _, v := range data {
			key := fmt.Sprintf("%d-%d-%s", v.Attachment.BizID, v.Attachment.AppID, v.Attachment.UID)
			if v.ID == 0 {
				v.ID = createID[key]
			}
			if _, exists := createID[key]; !exists {
				createID[key] = v.ID
			}
		}
	}

	// 先更新心跳，再更新变更
	errH := s.dao.Client().UpsertHeartbeat(kt, tx, toUpdate[sfs.Heartbeat.String()])
	errV := s.dao.Client().UpsertVersionChange(kt, tx, toUpdate[sfs.VersionChangeMessage.String()])
	if errH != nil && errV != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, fmt.Errorf("upsert heartbeat err: %v, upsert version change err: %v", errH, errV)
	}

	err = s.doBatchCreateClientEvents(kt, tx, req.GetClientEventItems(), createID)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}

	if e := tx.Commit(); e != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}
	return &pbds.BatchUpsertClientMetricsResp{}, nil
}

// handle client data
func (s *Service) handleBatchCreateClients(kt *kit.Kit, clients []*pbclient.Client) (toCreate []*table.Client,
	toUpdate map[string][]*table.Client, err error) {

	if len(clients) == 0 {
		return nil, nil, nil
	}

	data := [][]interface{}{}
	for _, item := range clients {
		data = append(data, []interface{}{item.Attachment.BizId, item.Attachment.AppId, item.Attachment.Uid})
	}

	oldData := make(map[string]*table.Client)
	tuple, err := s.dao.Client().ListClientByTuple(kt, data)
	if err != nil {
		return nil, nil, err
	}
	for _, item := range tuple {
		key := fmt.Sprintf("%d-%d-%s", item.Attachment.BizID, item.Attachment.AppID, item.Attachment.UID)
		oldData[key] = item
	}

	// 以心跳时间排序时间asc
	sort.Slice(clients, func(i, j int) bool {
		return clients[i].Spec.LastHeartbeatTime.AsTime().Before(clients[j].Spec.LastHeartbeatTime.AsTime())
	})

	// 如果该数据不在 client 中有以下两种情况:
	// 该数据的键不在 existingKeys 中,将其视为新增数据,并添加到 toCreate 中.
	// 该数据的键已经在 existingKeys 中,将其视为修改数据,并添加到 toUpdate 中.
	// 如果该数据在 client 中,将其视为修改数据,并添加到 toUpdate 中,还需处理第一次连接时间和ID.
	existingKeys := make(map[string]bool)
	toCreate = []*table.Client{}
	toUpdate = make(map[string][]*table.Client)
	for _, item := range clients {
		key := fmt.Sprintf("%d-%d-%s", item.Attachment.BizId, item.Attachment.AppId, item.Attachment.Uid)
		v, ok := oldData[key]
		if !ok {
			if !existingKeys[key] {
				toCreate = append(toCreate, &table.Client{
					Attachment: item.GetAttachment().ClientAttachment(),
					Spec:       item.GetSpec().ClientSpec(),
				})
				existingKeys[key] = true
			} else {
				toUpdate[item.MessageType] = append(toUpdate[item.MessageType], &table.Client{
					Attachment: item.GetAttachment().ClientAttachment(),
					Spec:       item.GetSpec().ClientSpec(),
				})
			}
		} else {
			item.Spec.FirstConnectTime = timestamppb.New(v.Spec.FirstConnectTime)
			if item.Spec.ReleaseChangeStatus != sfs.Success.String() {
				item.Spec.CurrentReleaseId = v.Spec.CurrentReleaseID
			}
			toUpdate[item.MessageType] = append(toUpdate[item.MessageType], &table.Client{
				ID:         v.ID,
				Attachment: item.GetAttachment().ClientAttachment(),
				Spec:       item.Spec.ClientSpec(),
			})
		}
	}

	return toCreate, toUpdate, nil
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
		v.Spec.Resource.CpuUsage = math.Round(v.Spec.Resource.CpuUsage*1000) / 1000
		v.Spec.Resource.CpuMaxUsage = math.Round(v.Spec.Resource.CpuMaxUsage*1000) / 1000
		v.Spec.Resource.MemoryUsage /= (1024 * 1024)
		v.Spec.Resource.MemoryMaxUsage /= (1024 * 1024)
	}

	resp := &pbds.ListClientsResp{
		Details: data,
		Count:   uint32(count),
	}

	return resp, nil

}

// ClientConfigVersionStatistics 客户端配置版本统计
func (s *Service) ClientConfigVersionStatistics(ctx context.Context, req *pbclient.ClientCommonReq) (
	*structpb.Struct, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	configVersion, err := s.clientConfigVersionChart(grpcKit, req.GetBizId(),
		req.GetAppId(), req.GetLastHeartbeatTime(), req.GetSearch())
	if err != nil {
		return nil, err
	}

	resp := make(map[string]interface{})
	resp["client_config_version"] = configVersion

	return structpb.NewStruct(resp)
}

// ClientPullTrendStatistics 客户端拉取趋势统计
func (s *Service) ClientPullTrendStatistics(ctx context.Context, req *pbclient.ClientCommonReq) (
	*structpb.Struct, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	var ClientIDs []uint32
	if req.GetSearch().String() != "" {
		items, _, err := s.dao.Client().List(grpcKit, req.GetBizId(), req.GetAppId(), req.GetLastHeartbeatTime(),
			req.GetSearch(), &pbds.ListClientsReq_Order{}, &types.BasePage{All: true})
		if err != nil {
			return nil, err
		}
		for _, v := range items {
			ClientIDs = append(ClientIDs, v.ID)
		}
	}

	resp := make(map[string]interface{})
	if req.GetSearch().String() == "" || len(ClientIDs) > 0 {
		data, err := s.dao.ClientEvent().GetPullTrend(grpcKit, req.GetBizId(), req.GetAppId(), ClientIDs, req.GetPullTime())
		if err != nil {
			return nil, err
		}
		var ids []uint32
		for _, v := range data {
			ids = append(ids, v.ClientID)
		}

		// 反查客户端类型
		clients, err := s.dao.Client().ListClientByIDs(grpcKit, req.GetBizId(), req.GetAppId(), ids)
		if err != nil {
			return nil, err
		}
		ClientTypes := make(map[uint32]string)
		for _, v := range clients {
			ClientTypes[v.ID] = string(v.Spec.ClientType)
		}

		// 根据 time + type 统计数量
		typeCountByTimeAndType := make(map[string]int)
		for _, v := range data {
			ct, ok := ClientTypes[v.ClientID]
			if ok {
				key := fmt.Sprintf("%s_%s", v.PullTime.Format("2006/01/02"), ct)
				typeCountByTimeAndType[key]++
			}
		}

		var typeAndTime []interface{}
		for key, count := range typeCountByTimeAndType {
			parts := strings.Split(key, "_")
			item := map[string]interface{}{
				"time": parts[0], "value": count, "type": parts[1],
			}
			typeAndTime = append(typeAndTime, item)
		}

		// 根据 time 统计数量
		typeCountByTime := make(map[string]int)
		for _, v := range data {
			key := v.PullTime.Format("2006/01/02")
			typeCountByTime[key]++
		}

		var byTime []interface{}
		for key, count := range typeCountByTime {
			item := map[string]interface{}{
				"time": key, "count": count,
			}
			byTime = append(byTime, item)
		}

		sort.Slice(typeAndTime, func(i, j int) bool {
			return compareTime(typeAndTime[i], typeAndTime[j])
		})

		sort.Slice(byTime, func(i, j int) bool {
			return compareTime(byTime[i], byTime[j])
		})

		resp["time_and_type"] = typeAndTime
		resp["time"] = byTime
	}

	return structpb.NewStruct(resp)
}

// ClientPullStatistics 客户端拉取信息统计
func (s *Service) ClientPullStatistics(ctx context.Context, req *pbclient.ClientCommonReq) (
	*structpb.Struct, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp, err := s.clientPullInfo(grpcKit, req.GetBizId(), req.GetAppId(),
		req.GetLastHeartbeatTime(), req.GetSearch())
	if err != nil {
		return nil, err
	}

	return structpb.NewStruct(resp)
}

// ClientLabelStatistics 客户端标签统计
func (s *Service) ClientLabelStatistics(ctx context.Context, req *pbclient.ClientCommonReq) (
	*structpb.Struct, error) {

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

	return structpb.NewStruct(resp)
}

// ClientAnnotationStatistics 客户端附加信息统计
func (s *Service) ClientAnnotationStatistics(_ context.Context, _ *pbclient.ClientCommonReq) (
	*structpb.Struct, error) {

	resp := make(map[string]interface{})
	return structpb.NewStruct(resp)
}

// ClientVersionStatistics 客户端版本统计
func (s *Service) ClientVersionStatistics(ctx context.Context, req *pbclient.ClientCommonReq) (
	*structpb.Struct, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	distribution, err := s.clientVersionDistribution(grpcKit, req.GetBizId(), req.GetAppId(),
		req.GetLastHeartbeatTime(), req.GetSearch())
	if err != nil {
		return nil, err
	}

	return structpb.NewStruct(distribution)
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
func (s *Service) clientVersionDistribution(kit *kit.Kit, bizID, appID uint32, heartbeatTime int64,
	search *pbclient.ClientQueryCondition) (map[string]interface{}, error) {

	items, _, err := s.dao.Client().List(kit, bizID, appID,
		heartbeatTime,
		search,
		&pbds.ListClientsReq_Order{},
		&types.BasePage{
			All: true,
		})
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int)
	for _, item := range items {
		counts[string(item.Spec.ClientType)+"_"+item.Spec.ClientVersion]++
	}

	totalCount := 0
	for _, count := range counts {
		totalCount += count
	}

	var data []map[string]interface{}
	for _, item := range items {
		key := string(item.Spec.ClientType) + "_" + item.Spec.ClientVersion
		count := counts[key]
		percent := float64(count) / float64(totalCount) * 100
		data = append(data, map[string]interface{}{
			"client_type":    string(item.Spec.ClientType),
			"client_version": item.Spec.ClientVersion,
			"percent":        percent,
			"value":          count,
		})
	}

	filteredOutputData := make(map[string][]map[string]interface{})
	for _, item := range data {
		key := item["client_type"].(string) + "_" + item["client_version"].(string)
		filteredOutputData[key] = append(filteredOutputData[key], item)
	}

	var charts []interface{}
	for _, data := range filteredOutputData {
		charts = append(charts, data[0])
	}

	resp := make(map[string]interface{})
	resp["version_distribution"] = charts

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
	var ClientID []uint32
	if search.String() != "" {
		items, _, err := s.dao.Client().List(kit, bizID, appID, heartbeatTime, search,
			&pbds.ListClientsReq_Order{}, &types.BasePage{All: true})
		if err != nil {
			return nil, err
		}
		for _, v := range items {
			ClientID = append(ClientID, v.ID)
		}
	}
	timeChart := make(map[string]interface{})
	if search.String() == "" || len(ClientID) > 0 {
		time, err := s.dao.ClientEvent().GetMinMaxAvgTime(kit, bizID, appID, ClientID, search.GetReleaseChangeStatus())
		if err != nil {
			return nil, err
		}
		timeChart["min"] = time.Min
		timeChart["max"] = time.Max
		timeChart["avg"] = time.Avg
	}

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

	usage := map[string]interface{}{}
	usage["cpu_max_usage"] = math.Round(item.CpuMaxUsage*10) / 10
	usage["cpu_min_usage"] = math.Round(item.CpuMinUsage*10) / 10
	usage["cpu_avg_usage"] = math.Round(item.CpuAvgUsage*10) / 10
	usage["memory_max_usage"] = item.MemoryMaxUsage / (1024 * 1024)
	usage["memory_min_usage"] = item.MemoryMinUsage / (1024 * 1024)
	usage["memory_avg_usage"] = item.MemoryAvgUsage / (1024 * 1024)

	return usage, nil
}

func compareTime(a, b interface{}) bool {
	timeI, err := time.Parse("2006/01/02", getTimeFromElement(a))
	if err != nil {
		return true
	}
	timeJ, err := time.Parse("2006/01/02", getTimeFromElement(b))
	if err != nil {
		return true
	}
	return timeI.Before(timeJ)
}

func getTimeFromElement(elem interface{}) string {
	switch v := elem.(type) {
	case map[string]interface{}:
		timeVal, _ := v["time"].(string)
		return timeVal
	default:
		return ""
	}
}

// ListClientLabelAndAnnotation 列出客户端标签和注释
func (s *Service) ListClientLabelAndAnnotation(ctx context.Context, req *pbds.ListClientLabelAndAnnotationReq) (
	*structpb.Struct, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	items, _, err := s.dao.Client().List(grpcKit, req.GetBizId(), req.GetAppId(), 0,
		nil, &pbds.ListClientsReq_Order{}, &types.BasePage{All: true})
	if err != nil {
		return nil, err
	}

	lableKeys := make(map[string]bool)
	annotationKeys := make(map[string]bool)
	for _, v := range items {
		var label map[string]string
		if err := json.Unmarshal([]byte(v.Spec.Labels), &label); err != nil {
			logs.Errorf("json parsing failed, err: %v", err)
			continue
		}
		for key := range label {
			lableKeys[key] = true
		}

		var annotation map[string]interface{}
		if err := json.Unmarshal([]byte(v.Spec.Annotations), &annotation); err != nil {
			logs.Errorf("json parsing failed, err: %v", err)
			continue
		}
		for key := range annotation {
			annotationKeys[key] = true
		}
	}

	var lables []interface{}
	for key := range lableKeys {
		lables = append(lables, key)
	}
	var annotations []interface{}
	for key := range annotationKeys {
		annotations = append(annotations, key)
	}

	resp := make(map[string]interface{})
	resp["labels"] = lables
	resp["annotations"] = annotations
	return structpb.NewStruct(resp)
}
