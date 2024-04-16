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

package sfs

import (
	"errors"

	"google.golang.org/protobuf/types/known/timestamppb"

	pbclient "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/client"
	pbce "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/client-event"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/jsoni"
)

// PbClientMetric Version Change Payload convert to client metric proto
func (v *VersionChangePayload) PbClientMetric() (*pbclient.Client, error) {
	if v == nil {
		return nil, errors.New("VersionChangePayload is nil, can not be convert to proto")
	}
	var currentReleaseId uint32
	// 在clien表中更新目标ID和当前ID
	// 如果成功了CurrentReleaseId和TargetReleaseId是一致的
	if v.Application.ReleaseChangeStatus == Success {
		currentReleaseId = v.Application.TargetReleaseID
	} else {
		currentReleaseId = v.Application.CurrentReleaseID
	}

	data := &pbclient.Client{
		Spec: &pbclient.ClientSpec{
			ClientVersion:     v.BasicData.ClientVersion,
			ClientType:        string(v.BasicData.ClientType),
			Ip:                v.BasicData.IP,
			Labels:            toString(v.Application.Labels),
			Annotations:       toString(v.BasicData.Annotations),
			FirstConnectTime:  timestamppb.New(v.BasicData.HeartbeatTime),
			LastHeartbeatTime: timestamppb.New(v.BasicData.HeartbeatTime),
			OnlineStatus:      v.BasicData.OnlineStatus.String(),
			Resource: &pbclient.ClientResource{
				CpuUsage:       v.ResourceUsage.CpuUsage,
				CpuMaxUsage:    v.ResourceUsage.CpuMaxUsage,
				MemoryUsage:    v.ResourceUsage.MemoryUsage,
				MemoryMaxUsage: v.ResourceUsage.MemoryMaxUsage,
				CpuMinUsage:    v.ResourceUsage.CpuMinUsage,
				CpuAvgUsage:    v.ResourceUsage.CpuAvgUsage,
				MemoryMinUsage: v.ResourceUsage.MemoryMinUsage,
				MemoryAvgUsage: v.ResourceUsage.MemoryAvgUsage,
			},
			CurrentReleaseId:          currentReleaseId,
			TargetReleaseId:           v.Application.TargetReleaseID,
			ReleaseChangeStatus:       v.Application.ReleaseChangeStatus.String(),
			ReleaseChangeFailedReason: v.Application.FailedReason.String(),
			FailedDetailReason:        v.Application.FailedDetailReason,
			SpecificFailedReason:      v.Application.SpecificFailedReason.String(),
		},
		Attachment: &pbclient.ClientAttachment{
			Uid:   v.Application.Uid,
			BizId: v.BasicData.BizID,
			AppId: v.Application.AppID,
		},
		MessageType: VersionChangeMessage.String(),
	}
	return data, nil
}

// PbClientEventMetric Version Change Payload convert to client event metric proto
func (v *VersionChangePayload) PbClientEventMetric() (*pbce.ClientEvent, error) {
	if v == nil {
		return nil, errors.New("VersionChangePayload is nil, can not be convert to proto")
	}
	data := &pbce.ClientEvent{
		Spec: &pbce.ClientEventSpec{
			OriginalReleaseId:         v.Application.CurrentReleaseID,
			TargetReleaseId:           v.Application.TargetReleaseID,
			StartTime:                 timestamppb.New(v.Application.StartTime),
			EndTime:                   timestamppb.New(v.Application.EndTime),
			TotalSeconds:              v.Application.TotalSeconds,
			TotalFileSize:             float64(v.Application.TotalFileSize),
			TotalFileNum:              uint32(v.Application.TotalFileNum),
			DownloadFileSize:          float64(v.Application.DownloadFileSize),
			DownloadFileNum:           uint32(v.Application.DownloadFileNum),
			ReleaseChangeStatus:       v.Application.ReleaseChangeStatus.String(),
			ReleaseChangeFailedReason: v.Application.FailedReason.String(),
			FailedDetailReason:        v.Application.FailedDetailReason,
			SpecificFailedReason:      v.Application.SpecificFailedReason.String(),
		},
		Attachment: &pbce.ClientEventAttachment{
			Uid:        v.Application.Uid,
			ClientMode: v.BasicData.ClientMode.String(),
			BizId:      v.BasicData.BizID,
			AppId:      v.Application.AppID,
			CursorId:   v.Application.CursorID,
		},
		MessageType:   VersionChangeMessage.String(),
		HeartbeatTime: timestamppb.New(v.BasicData.HeartbeatTime),
	}
	return data, nil
}

// PbClientMetric heart beat Payload convert to client metric proto
func (h *HeartbeatItem) PbClientMetric() (*pbclient.Client, error) {
	if h == nil {
		return nil, errors.New("HeartbeatItem is nil, can not be convert to proto")
	}
	// 过滤没有目标版本号的数据，无意义
	if h.Application.CursorID == "" {
		return nil, nil
	}

	data := &pbclient.Client{
		Spec: &pbclient.ClientSpec{
			ClientVersion:     h.BasicData.ClientVersion,
			ClientType:        string(h.BasicData.ClientType),
			Ip:                h.BasicData.IP,
			Labels:            toString(h.Application.Labels),
			Annotations:       toString(h.BasicData.Annotations),
			FirstConnectTime:  timestamppb.New(h.BasicData.HeartbeatTime),
			LastHeartbeatTime: timestamppb.New(h.BasicData.HeartbeatTime),
			OnlineStatus:      h.BasicData.OnlineStatus.String(),
			Resource: &pbclient.ClientResource{
				CpuUsage:       h.ResourceUsage.CpuUsage,
				CpuMaxUsage:    h.ResourceUsage.CpuMaxUsage,
				CpuMinUsage:    h.ResourceUsage.CpuMinUsage,
				CpuAvgUsage:    h.ResourceUsage.CpuAvgUsage,
				MemoryUsage:    h.ResourceUsage.MemoryUsage,
				MemoryMaxUsage: h.ResourceUsage.MemoryMaxUsage,
				MemoryMinUsage: h.ResourceUsage.MemoryMinUsage,
				MemoryAvgUsage: h.ResourceUsage.MemoryAvgUsage,
			},
			ReleaseChangeStatus: h.Application.ReleaseChangeStatus.String(),
			CurrentReleaseId:    h.Application.CurrentReleaseID,
			TargetReleaseId:     h.Application.TargetReleaseID,
		},
		Attachment: &pbclient.ClientAttachment{
			Uid:   h.Application.Uid,
			BizId: h.BasicData.BizID,
			AppId: h.Application.AppID,
		},
		MessageType: Heartbeat.String(),
	}

	return data, nil
}

// PbClientEventMetric heart beat Payload convert to client event metric proto
// 心跳事件 ClientEvent表 只需更新 ReleaseChangeStatus 、SuccessDownloads、SuccessFileSize
func (h *HeartbeatItem) PbClientEventMetric() (*pbce.ClientEvent, error) {
	if h == nil {
		return nil, errors.New("HeartbeatItem is nil, can not be convert to proto")
	}

	// 客户端启动时可能有客户端连接和心跳数据，
	// 但没有发生拉取和变更事件, 所以CursorID会存在空
	// 过滤CursorID为空的数据，该数据没有任何意义
	if h.Application.CursorID == "" {
		return nil, nil
	}
	data := &pbce.ClientEvent{
		Spec: &pbce.ClientEventSpec{
			ReleaseChangeStatus: h.Application.ReleaseChangeStatus.String(),
			DownloadFileSize:    float64(h.Application.DownloadFileSize),
			DownloadFileNum:     uint32(h.Application.DownloadFileNum),
		},
		Attachment: &pbce.ClientEventAttachment{
			Uid:        h.Application.Uid,
			ClientMode: h.BasicData.ClientMode.String(),
			BizId:      h.BasicData.BizID,
			AppId:      h.Application.AppID,
			CursorId:   h.Application.CursorID,
		},
		HeartbeatTime: timestamppb.New(h.BasicData.HeartbeatTime),
		MessageType:   Heartbeat.String(),
	}

	return data, nil
}

func toString(label interface{}) string {
	marshal, err := jsoni.Marshal(label)
	if err != nil {
		return "{}"
	}

	if string(marshal) == "null" {
		return "{}"
	}

	return string(marshal)
}
