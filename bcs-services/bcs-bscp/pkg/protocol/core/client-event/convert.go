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
