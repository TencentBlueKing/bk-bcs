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
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbce "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/client-event"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	sfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

func (s *Service) doBatchCreateClientEvents(kt *kit.Kit, tx *gen.QueryTx, clientEvents []*pbce.ClientEvent, // nolint
	clientID map[string]uint32) error {

	// 处理clientID
	clientEventsData := []*pbce.ClientEvent{}
	for _, item := range clientEvents {
		key := fmt.Sprintf("%d-%d-%s", item.Attachment.BizId, item.Attachment.AppId, item.Attachment.Uid)
		id, ok := clientID[key]
		if ok {
			item.Attachment.ClientId = id
			clientEventsData = append(clientEventsData, &pbce.ClientEvent{
				Attachment:  item.Attachment,
				Spec:        item.Spec,
				MessageType: item.MessageType,
			})
		}
	}

	if len(clientEventsData) == 0 {
		return nil
	}

	// 根据 bizID + appID + UID + CursorID 查找存在的数据
	data := [][]interface{}{}
	for _, item := range clientEventsData {
		data = append(data, []interface{}{item.Attachment.BizId, item.Attachment.AppId, item.Attachment.Uid,
			item.Attachment.CursorId})
	}
	list, err := s.dao.ClientEvent().ListClientByTuple(kt, data)
	if err != nil {
		return err
	}
	clientEventID := map[string]uint32{}
	for _, v := range list {
		key := fmt.Sprintf("%d-%d-%s-%s", v.Attachment.BizID, v.Attachment.AppID, v.Attachment.UID, v.Attachment.CursorID)
		clientEventID[key] = v.ID
	}

	toCreate, toUpdate := s.handleBatchCreateClientEvents(clientEventsData, clientEventID)

	err = s.dao.ClientEvent().BatchCreateWithTx(kt, tx, toCreate)
	if err != nil {
		return err
	}

	// 获取创建后的ID
	createID := make(map[string]uint32)
	for _, item := range toCreate {
		key := fmt.Sprintf("%d-%d-%s-%s", item.Attachment.BizID, item.Attachment.AppID, item.Attachment.UID,
			item.Attachment.CursorID)
		createID[key] = item.ID
	}

	// 判断类型更新对应字段
	heartbeatData := []*table.ClientEvent{}
	versionChangeData := []*table.ClientEvent{}

	for _, v := range toUpdate {
		key := fmt.Sprintf("%d-%d-%s-%s", v.Attachment.BizId, v.Attachment.AppId, v.Attachment.Uid,
			v.Attachment.CursorId)
		uid := v.Id
		if uid == 0 {
			cid, ok := createID[key]
			if !ok {
				uid = 0
			} else {
				uid = cid
			}
		}
		if uid != 0 {
			switch v.MessageType {
			case "Heartbeat":
				heartbeatData = append(heartbeatData, &table.ClientEvent{
					ID:         uid,
					Attachment: v.Attachment.ClientEventAttachment(),
					Spec:       v.Spec.ClientEventSpec(),
				})
			case "VersionChange":
				versionChangeData = append(versionChangeData, &table.ClientEvent{
					ID:         uid,
					Attachment: v.Attachment.ClientEventAttachment(),
					Spec:       v.Spec.ClientEventSpec(),
				})
			}
		}
	}

	if len(heartbeatData) != 0 {
		err = s.dao.ClientEvent().BatchUpdateSelectFieldTx(kt, tx, sfs.Heartbeat, heartbeatData)
		if err != nil {
			return err
		}
	}

	if len(versionChangeData) != 0 {
		err = s.dao.ClientEvent().BatchUpdateSelectFieldTx(kt, tx, sfs.VersionChangeMessage, versionChangeData)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) handleBatchCreateClientEvents(clientEventsData []*pbce.ClientEvent,
	clientEventID map[string]uint32) (toCreate []*table.ClientEvent, clientEUpdateData []*pbce.ClientEvent) {
	clientEUpdateData = []*pbce.ClientEvent{}
	clientECreateData := []*pbce.ClientEvent{}
	// 验证哪些数据需要新增和修改
	for _, item := range clientEventsData {
		key := fmt.Sprintf("%d-%d-%s-%s", item.Attachment.BizId, item.Attachment.AppId, item.Attachment.Uid,
			item.Attachment.CursorId)
		id, ok := clientEventID[key]

		if item.Spec.EndTime.GetSeconds() < 0 {
			item.GetSpec().EndTime = nil
		}

		if !ok {
			clientECreateData = append(clientECreateData, &pbce.ClientEvent{
				Attachment:  item.GetAttachment(),
				Spec:        item.GetSpec(),
				MessageType: item.MessageType,
			})
		} else {
			clientEUpdateData = append(clientEUpdateData, &pbce.ClientEvent{
				Id:          id,
				Attachment:  item.GetAttachment(),
				Spec:        item.GetSpec(),
				MessageType: item.MessageType,
			})
		}
	}

	// Client Event数据会存在同一维度下多个类型的消息
	// 如果同一维度下的数据都存在那么都是更新
	// 如果同一维度下的数据不存在，只允许一条创建其他都是更新操作
	createData := make(map[string]*pbce.ClientEvent)
	otherCreateData := make([]*pbce.ClientEvent, 0)
	for _, item := range clientECreateData {
		key := fmt.Sprintf("%d-%d-%s", item.Attachment.BizId, item.Attachment.AppId, item.Attachment.CursorId)
		_, ok := createData[key]
		if ok {
			otherCreateData = append(otherCreateData, item)
		} else {
			createData[key] = item
		}
	}

	// 该数据是最终需要创建的数据
	toCreate = []*table.ClientEvent{}
	for _, v := range createData {
		toCreate = append(toCreate, &table.ClientEvent{
			Spec:       v.Spec.ClientEventSpec(),
			Attachment: v.Attachment.ClientEventAttachment(),
		})
	}
	clientEUpdateData = append(clientEUpdateData, otherCreateData...)
	return toCreate, clientEUpdateData
}

// ListClientMetricEvents List client metric details query
func (s *Service) ListClientMetricEvents(ctx context.Context, req *pbds.ListClientMetricEventsReq) (
	*pbds.ListClientMetricEventsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	var err error
	var starTime, endTime time.Time
	if len(req.GetStartTime()) > 0 {
		starTime, err = time.Parse("2006-01-02", req.GetStartTime())
		if err != nil {
			return nil, err
		}
	}
	if len(req.GetEndTime()) > 0 {
		endTime, err = time.Parse("2006-01-02", req.GetEndTime())
		if err != nil {
			return nil, err
		}
	}

	items, count, err := s.dao.ClientEvent().List(grpcKit, req.GetBizId(), req.GetAppId(), req.GetClientMetricsId(),
		starTime,
		endTime,
		req.GetSearchValue(),
		req.GetOrder(),
		&types.BasePage{
			Start: req.GetStart(),
			Limit: uint(req.GetLimit()),
			All:   req.GetAll(),
		})
	if err != nil {
		return nil, err
	}

	// 处理当前和目标版本名称
	releaseMap := make(map[uint32]struct{})
	for _, v := range items {
		if v.Spec.OriginalReleaseID > 0 {
			releaseMap[v.Spec.OriginalReleaseID] = struct{}{}
		}
	}
	for _, v := range items {
		if v.Spec.TargetReleaseID > 0 {
			releaseMap[v.Spec.TargetReleaseID] = struct{}{}
		}
	}

	releaseIDs := make([]uint32, 0, len(releaseMap))
	for id := range releaseMap {
		releaseIDs = append(releaseIDs, id)
	}
	releases, err := s.dao.Release().ListAllByIDs(grpcKit, releaseIDs, req.BizId)
	if err != nil {
		return nil, err
	}

	releaseNames := map[uint32]string{}
	for _, v := range releases {
		releaseNames[v.ID] = v.Spec.Name
	}

	data := pbce.PbClientEvents(items)
	for _, v := range data {
		v.OriginalReleaseName = releaseNames[v.Spec.OriginalReleaseId]
		v.TargetReleaseName = releaseNames[v.Spec.TargetReleaseId]
	}

	resp := &pbds.ListClientMetricEventsResp{
		Count:   uint32(count),
		Details: data,
	}
	return resp, nil
}
