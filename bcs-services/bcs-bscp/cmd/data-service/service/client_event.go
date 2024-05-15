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
	"sort"
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	pbce "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/client-event"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	sfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

func (s *Service) doBatchCreateClientEvents(kt *kit.Kit, tx *gen.QueryTx, clientEvents []*pbce.ClientEvent,
	clientID map[string]uint32) error {

	if len(clientEvents) == 0 {
		return nil
	}

	var err error

	var toCreate []*table.ClientEvent
	var toUpdate map[string][]*table.ClientEvent

	toCreate, toUpdate, err = s.handleCreateClientEvents(kt, clientEvents, clientID)
	if err != nil {
		return err
	}
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

	// 更新时id不能为空
	// 更新 client_event 时需要clientID
	for _, data := range toUpdate {
		for _, item := range data {
			key := fmt.Sprintf("%d-%d-%s-%s", item.Attachment.BizID, item.Attachment.AppID, item.Attachment.UID,
				item.Attachment.CursorID)
			if item.ID == 0 {
				item.ID = createID[key]
			}
		}
	}

	// 先更新心跳，再更新变更
	errH := s.dao.ClientEvent().UpsertHeartbeat(kt, tx, toUpdate[sfs.Heartbeat.String()])
	errV := s.dao.ClientEvent().UpsertVersionChange(kt, tx, toUpdate[sfs.VersionChangeMessage.String()])
	if errH != nil && errV != nil {
		return fmt.Errorf("upsert heartbeat err: %v, upsert version change err: %v", errH, errV)
	}

	return nil
}

// handle client event data
func (s *Service) handleCreateClientEvents(kt *kit.Kit, clientEvents []*pbce.ClientEvent, clientID map[string]uint32) (
	toCreate []*table.ClientEvent, toUpdate map[string][]*table.ClientEvent, err error) {

	data := [][]interface{}{}
	for _, item := range clientEvents {
		data = append(data, []interface{}{item.Attachment.BizId, item.Attachment.AppId, item.Attachment.Uid,
			item.Attachment.CursorId})
	}
	list, err := s.dao.ClientEvent().ListClientByTuple(kt, data)
	if err != nil {
		return nil, nil, err
	}
	oldData := map[string]uint32{}
	for _, v := range list {
		key := fmt.Sprintf("%d-%d-%s-%s", v.Attachment.BizID, v.Attachment.AppID, v.Attachment.UID, v.Attachment.CursorID)
		oldData[key] = v.ID
	}

	// 以心跳时间排序时间asc
	sort.Slice(clientEvents, func(i, j int) bool {
		return clientEvents[i].HeartbeatTime.AsTime().Before(clientEvents[j].HeartbeatTime.AsTime())
	})

	// 如果该数据不在 client_event 中有以下两种情况:
	// 该数据的键不在 existingKeys 中,将其视为新增数据,并添加到 toCreate 中.
	// 该数据的键已经在 existingKeys 中,将其视为修改数据,并添加到 toUpdate 中.
	// 如果该数据在 client_event 中,将其视为修改数据,并添加到 toUpdate 中.
	existingKeys := make(map[string]bool)
	toCreate = []*table.ClientEvent{}
	toUpdate = make(map[string][]*table.ClientEvent)
	for _, item := range clientEvents {
		keyWithoutCursor := fmt.Sprintf("%d-%d-%s-%s", item.Attachment.BizId, item.Attachment.AppId, item.Attachment.Uid,
			item.Attachment.CursorId)
		v, ok := oldData[keyWithoutCursor]
		if item.Spec.EndTime.GetSeconds() <= 0 {
			item.Spec.EndTime = nil
		}
		fullKey := fmt.Sprintf("%d-%d-%s", item.Attachment.BizId, item.Attachment.AppId, item.Attachment.Uid)
		if clientID[fullKey] > 0 {
			item.Attachment.ClientId = clientID[fullKey]
			if !ok {
				if !existingKeys[keyWithoutCursor] {
					toCreate = append(toCreate, &table.ClientEvent{
						Attachment: item.Attachment.ClientEventAttachment(),
						Spec:       item.Spec.ClientEventSpec(),
					})
					existingKeys[keyWithoutCursor] = true
				} else {
					toUpdate[item.MessageType] = append(toUpdate[item.MessageType], &table.ClientEvent{
						Attachment: item.Attachment.ClientEventAttachment(),
						Spec:       item.Spec.ClientEventSpec(),
					})
				}
			} else {
				toUpdate[item.MessageType] = append(toUpdate[item.MessageType], &table.ClientEvent{
					ID:         v,
					Attachment: item.Attachment.ClientEventAttachment(),
					Spec:       item.Spec.ClientEventSpec(),
				})
			}
		}
	}

	return toCreate, toUpdate, nil
}

// ListClientEvents List client details query
func (s *Service) ListClientEvents(ctx context.Context, req *pbds.ListClientEventsReq) (
	*pbds.ListClientEventsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	var err error
	var starTime, endTime time.Time
	if len(req.GetStartTime()) > 0 {
		starTime, err = time.ParseInLocation("2006-01-02 15:04:05", req.GetStartTime(), time.UTC)
		if err != nil {
			return nil, err
		}
	}

	if len(req.GetEndTime()) > 0 {
		endTime, err = time.ParseInLocation("2006-01-02 15:04:05", req.GetEndTime(), time.UTC)
		if err != nil {
			return nil, err
		}
	}

	items, count, err := s.dao.ClientEvent().List(grpcKit, req.GetBizId(), req.GetAppId(), req.GetClientId(),
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

	resp := &pbds.ListClientEventsResp{
		Count:   uint32(count),
		Details: data,
	}
	return resp, nil
}
