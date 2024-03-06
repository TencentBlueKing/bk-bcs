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
	"gorm.io/gen/field"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	sfs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/sf-share"
)

// ClientEvent supplies all the client event related operations.
type ClientEvent interface {
	// BatchCreateWithTx batch create client event instances with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, data []*table.ClientEvent) error
	// BatchUpdateSelectFieldTx Update selected field
	BatchUpdateSelectFieldTx(kit *kit.Kit, tx *gen.QueryTx, messageType sfs.MessagingType, data []*table.ClientEvent) error
	// ListClientByTuple Query the client list according to multiple fields in
	ListClientByTuple(kit *kit.Kit, data [][]interface{}) ([]*table.ClientEvent, error)
}

var _ ClientEvent = new(clientEventDao)

type clientEventDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// ListClientByTuple implements ClientEvent.
func (dao *clientEventDao) ListClientByTuple(kit *kit.Kit, data [][]interface{}) ([]*table.ClientEvent, error) {
	m := dao.genQ.ClientEvent
	return dao.genQ.ClientEvent.WithContext(kit.Ctx).
		Select(m.ID, m.BizID, m.AppID, m.UID, m.CursorID, m.ClientMode).
		Where(m.WithContext(kit.Ctx).Columns(m.BizID, m.AppID, m.UID, m.CursorID).
			In(field.Values(data))).
		Find()
}

// BatchUpdateSelectFieldTx implements ClientEvent.
// 版本变更时所有所有数据都可以更新
// 心跳主要更新下载数和下载文件大小、变更状态
func (dao *clientEventDao) BatchUpdateSelectFieldTx(kit *kit.Kit, tx *gen.QueryTx, messageType sfs.MessagingType,
	data []*table.ClientEvent) error {
	m := dao.genQ.ClientEvent
	q := tx.ClientEvent.WithContext(kit.Ctx)

	// 根据类型更新字段
	switch messageType {
	case sfs.VersionChangeMessage:
		q = q.Omit()
	case sfs.Heartbeat:
		q = q.Omit(m.ClientMode, m.TotalFileNum, m.TotalFileSize, m.TotalSeconds, m.OriginalReleaseID,
			m.TargetReleaseID, m.StartTime, m.EndTime, m.FailedDetailReason, m.ReleaseChangeFailedReason)
	}

	return q.Save(data...)
}

// BatchCreateWithTx batch create client event instances with transaction.
func (dao *clientEventDao) BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, data []*table.ClientEvent) error {
	// generate an config item id and update to config item.
	if len(data) == 0 {
		return nil
	}
	ids, err := dao.idGen.Batch(kit, table.ClientEventTable, len(data))
	if err != nil {
		return err
	}
	for i, item := range data {
		if err := item.ValidateCreate(); err != nil {
			return err
		}
		item.ID = ids[i]
	}
	if err := tx.ClientEvent.WithContext(kit.Ctx).Save(data...); err != nil {
		return err
	}
	return nil
}
