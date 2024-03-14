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
	"fmt"
	"math"
	"sort"

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

// ListClientMetrics implements pbds.DataServer.
func (s *Service) ListClientMetrics(ctx context.Context, req *pbds.ListClientMetricsReq) (
	*pbds.ListClientMetricsResp, error) {
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

	resp := &pbds.ListClientMetricsResp{
		Details: data,
		Count:   uint32(count),
	}

	return resp, nil

}
