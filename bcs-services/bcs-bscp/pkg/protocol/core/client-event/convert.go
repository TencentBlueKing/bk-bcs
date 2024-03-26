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

// Package pbce xxx
package pbce

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
)

// ClientEventSpec convert pb ClientEventSpec to table ClientEventSpec
func (c *ClientEventSpec) ClientEventSpec() *table.ClientEventSpec {
	if c == nil {
		return nil
	}

	return &table.ClientEventSpec{
		OriginalReleaseID:         c.OriginalReleaseId,
		TargetReleaseID:           c.TargetReleaseId,
		StartTime:                 c.StartTime.AsTime(),
		EndTime:                   c.EndTime.AsTime(),
		TotalSeconds:              c.TotalSeconds,
		TotalFileSize:             c.TotalFileSize,
		DownloadFileSize:          c.DownloadFileSize,
		TotalFileNum:              c.TotalFileNum,
		DownloadFileNum:           c.DownloadFileNum,
		ReleaseChangeStatus:       table.Status(c.ReleaseChangeStatus),
		ReleaseChangeFailedReason: table.FailedReason(c.ReleaseChangeFailedReason),
		FailedDetailReason:        c.FailedDetailReason,
	}
}

// PbClientEventSpec convert table ClientEventSpec to pb ClientEventSpec
func PbClientEventSpec(spec *table.ClientEventSpec) *ClientEventSpec { //nolint:revive
	if spec == nil {
		return nil
	}

	return &ClientEventSpec{
		OriginalReleaseId:         spec.OriginalReleaseID,
		TargetReleaseId:           spec.TargetReleaseID,
		StartTime:                 timestamppb.New(spec.StartTime),
		EndTime:                   timestamppb.New(spec.EndTime),
		TotalSeconds:              spec.TotalSeconds,
		TotalFileSize:             spec.TotalFileSize,
		DownloadFileSize:          spec.DownloadFileSize,
		TotalFileNum:              spec.TotalFileNum,
		DownloadFileNum:           spec.DownloadFileNum,
		ReleaseChangeStatus:       string(spec.ReleaseChangeStatus),
		ReleaseChangeFailedReason: string(spec.ReleaseChangeFailedReason),
		FailedDetailReason:        spec.FailedDetailReason,
	}
}

// ClientEventAttachment convert pb ClientEventAttachment to table ClientEventAttachment
func (c *ClientEventAttachment) ClientEventAttachment() *table.ClientEventAttachment {
	if c == nil {
		return nil
	}

	return &table.ClientEventAttachment{
		ClientID:   c.ClientId,
		CursorID:   c.CursorId,
		UID:        c.Uid,
		BizID:      c.BizId,
		AppID:      c.AppId,
		ClientMode: table.ClientMode(c.ClientMode),
	}
}

// PbClientEventAttachment convert table PbClientEventAttachment to pb PbClientEventAttachment
func PbClientEventAttachment(attachment *table.ClientEventAttachment) *ClientEventAttachment { // nolint
	if attachment == nil {
		return nil
	}
	return &ClientEventAttachment{
		ClientId:   attachment.ClientID,
		CursorId:   attachment.CursorID,
		Uid:        attachment.UID,
		BizId:      attachment.BizID,
		AppId:      attachment.AppID,
		ClientMode: string(attachment.ClientMode),
	}
}

// PbClientEvent convert table ClientEvent to pb ClientEvent
func PbClientEvent(c *table.ClientEvent) *ClientEvent {
	if c == nil {
		return nil
	}

	return &ClientEvent{
		Id:         c.ID,
		Spec:       PbClientEventSpec(c.Spec),
		Attachment: PbClientEventAttachment(c.Attachment),
	}
}

// PbClientEvents convert table ClientEvent to pb ClientEvent
func PbClientEvents(c []*table.ClientEvent) []*ClientEvent {
	if c == nil {
		return make([]*ClientEvent, 0)
	}
	result := make([]*ClientEvent, 0)
	for _, v := range c {
		result = append(result, PbClientEvent(v))
	}
	return result
}
