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

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
)

// Content supplies all the content related operations.
type Content interface {
	// Create one content instance.
	Create(kit *kit.Kit, content *table.Content) (uint32, error)
	// CreateWithTx create one content instance with transaction
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, content *table.Content) (uint32, error)
	// BatchCreateWithTx batch create content instances with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, contents []*table.Content) error
	// Get get content by id
	Get(kit *kit.Kit, id, bizID uint32) (*table.Content, error)
	// BatchDeleteWithTx batch delete content data instance with transaction.
	BatchDeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, contentIDs []uint32) error
}

var _ Content = new(contentDao)

type contentDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// BatchDeleteWithTx batch delete content data instance with transaction.
func (*contentDao) BatchDeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, contentIDs []uint32) error {

	m := tx.Content
	q := tx.Content.WithContext(kit.Ctx)
	_, err := q.Where(m.ID.In(contentIDs...)).Delete()
	if err != nil {
		return err
	}
	return nil
}

// Create one content instance
func (dao *contentDao) Create(kit *kit.Kit, content *table.Content) (uint32, error) {

	if content == nil {
		return 0, errf.New(errf.InvalidParameter, "content is nil")
	}

	if err := content.ValidateCreate(kit); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentResExist(kit, content.Attachment); err != nil {
		return 0, err
	}

	// generate an content id and update to content.
	id, err := dao.idGen.One(kit, table.ContentTable)
	if err != nil {
		return 0, err
	}

	content.ID = id

	ad := dao.auditDao.DecoratorV2(kit, content.Attachment.BizID).PrepareCreate(content)

	createTx := func(tx *gen.Query) error {
		q := tx.Content.WithContext(kit.Ctx)
		if err = q.Create(content); err != nil {
			return err
		}

		if err = ad.Do(tx); err != nil {
			return err
		}

		return nil
	}
	if err = dao.genQ.Transaction(createTx); err != nil {
		return 0, err
	}

	return id, nil
}

// CreateWithTx create one content instance with transaction
func (dao *contentDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, content *table.Content) (uint32, error) {

	if content == nil {
		return 0, errf.New(errf.InvalidParameter, "content is nil")
	}

	if err := content.ValidateCreate(kit); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	// generate an content id and update to content.
	id, err := dao.idGen.One(kit, table.ContentTable)
	if err != nil {
		return 0, err
	}

	content.ID = id

	if err := tx.Content.WithContext(kit.Ctx).Create(content); err != nil {
		return 0, err
	}

	ad := dao.auditDao.DecoratorV2(kit, content.Attachment.BizID).PrepareCreate(content)
	if err := ad.Do(tx.Query); err != nil {
		return 0, fmt.Errorf("audit create content failed, err: %v", err)
	}
	return id, nil
}

// BatchCreateWithTx batch create content instances with transaction.
// NOTE: 1. this method won't audit, because it's batch operation.
// 2. this method won't validate attachment resource exist, because it's batch operation.
func (dao *contentDao) BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, contents []*table.Content) error {
	if len(contents) == 0 {
		return nil
	}
	ids, err := dao.idGen.Batch(kit, table.ContentTable, len(contents))
	if err != nil {
		return err
	}
	for i, content := range contents {
		if err := content.ValidateCreate(kit); err != nil {
			return err
		}
		content.ID = ids[i]
	}
	if err := tx.Content.WithContext(kit.Ctx).Save(contents...); err != nil {
		return err
	}
	return nil
}

// Get content by id.
// Note: !!!current db is sharded by biz_id,it can not adapt bcs project,need redesign
func (dao *contentDao) Get(kit *kit.Kit, id, bizID uint32) (*table.Content, error) {

	if id == 0 {
		return nil, errf.New(errf.InvalidParameter, "content id can not be 0")
	}

	m := dao.genQ.Content

	return m.WithContext(kit.Ctx).Where(m.ID.Eq(id), m.BizID.Eq(bizID)).Take()
}

// validateAttachmentResExist validate if attachment resource exists before creating content.
func (dao *contentDao) validateAttachmentResExist(kit *kit.Kit, am *table.ContentAttachment) error {

	appQ := dao.genQ.App
	// validate if content attached app exists.
	if _, err := appQ.WithContext(kit.Ctx).Where(appQ.ID.Eq(am.AppID), appQ.BizID.Eq(am.BizID)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("content attached config item %d not exist", am.ConfigItemID)
		}
		return fmt.Errorf("get content attached app %d failed", am.AppID)
	}

	ciQ := dao.genQ.ConfigItem
	// validate if content attached config item exists.
	if _, err := ciQ.WithContext(kit.Ctx).Where(
		ciQ.BizID.Eq(am.BizID), ciQ.AppID.Eq(am.AppID), ciQ.ID.Eq(am.ConfigItemID)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("content attached config item %d not exist", am.ConfigItemID)
		}
		return fmt.Errorf("get content attached config item %d failed", am.ConfigItemID)
	}

	return nil
}
