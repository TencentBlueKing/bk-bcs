/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
)

// ConfigItem supplies all the configItem related operations.
type ConfigItem interface {
	// CreateWithTx create one configItem instance.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, configItem *table.ConfigItem) (uint32, error)
	// BatchCreateWithTx batch create configItem instances with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, appID uint32, configItems []*table.ConfigItem) error
	// ValidateAppCINumber verify whether the current number of app config items has reached the maximum.
	ValidateAppCINumber(kt *kit.Kit, tx *gen.QueryTx, bizID, appID uint32) error
	// Update one configItem instance.
	Update(kit *kit.Kit, configItem *table.ConfigItem) error
	// BatchUpdateWithTx batch update configItem instances with transaction.
	BatchUpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, configItems []*table.ConfigItem) error
	// Get configItem by id
	Get(kit *kit.Kit, id, bizID uint32) (*table.ConfigItem, error)
	// SearchAll search all configItem with searchKey.
	SearchAll(kit *kit.Kit, searchKey string, appID, bizID uint32) ([]*table.ConfigItem, error)
	// ListAllByAppID list all configItem by appID
	ListAllByAppID(kit *kit.Kit, appID uint32, bizID uint32) ([]*table.ConfigItem, error)
	// Delete one configItem instance.
	Delete(kit *kit.Kit, configItem *table.ConfigItem) error
	// Delete one configItem instance with transaction.
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, configItem *table.ConfigItem) error
	// BatchDeleteWithTx batch configItem instances with transaction.
	BatchDeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, ids []uint32, bizID, appID uint32) error
	// GetCount bizID config count
	GetCount(kit *kit.Kit, bizID uint32, appId []uint32) ([]*table.ListConfigItemCounts, error)
}

var _ ConfigItem = new(configItemDao)

type configItemDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
	lock     LockDao
}

// CreateWithTx create one configItem instance with transaction.
func (dao *configItemDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, ci *table.ConfigItem) (uint32, error) {
	if ci == nil {
		return 0, errors.New("config item is nil")
	}

	if err := ci.ValidateCreate(); err != nil {
		return 0, err
	}

	if err := dao.validateAttachmentResExist(kit, ci.Attachment); err != nil {
		return 0, err
	}

	// generate an config item id and update to config item.
	id, err := dao.idGen.One(kit, table.ConfigItemTable)
	if err != nil {
		return 0, err
	}

	ci.ID = id
	ad := dao.auditDao.DecoratorV2(kit, ci.Attachment.BizID).PrepareCreate(ci)

	if err := tx.ConfigItem.WithContext(kit.Ctx).Create(ci); err != nil {
		return 0, err
	}

	if err := ad.Do(tx.Query); err != nil {
		return 0, fmt.Errorf("audit create config item failed, err: %v", err)
	}

	return id, nil
}

// BatchCreateWithTx batch create configItem instances with transaction.
// NOTE: 1. this method won't audit, because it's batch operation.
// 2. this method won't validate attachment resource exist, because it's batch operation.
func (dao *configItemDao) BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx,
	bizID, appID uint32, configItems []*table.ConfigItem) error {
	// generate an config item id and update to config item.
	if len(configItems) == 0 {
		return nil
	}
	ids, err := dao.idGen.Batch(kit, table.ConfigItemTable, len(configItems))
	if err != nil {
		return err
	}
	for i, configItem := range configItems {
		if err := configItem.ValidateCreate(); err != nil {
			return err
		}
		configItem.ID = ids[i]
	}
	if err := tx.ConfigItem.WithContext(kit.Ctx).Save(configItems...); err != nil {
		return err
	}
	return nil
}

// Update one configItem instance.
func (dao *configItemDao) Update(kit *kit.Kit, ci *table.ConfigItem) error {

	if ci == nil {
		return errf.New(errf.InvalidParameter, "config item is nil")
	}

	m := dao.genQ.ConfigItem
	q := dao.genQ.ConfigItem.WithContext(kit.Ctx)

	// if file mode not update, need to query this ci's file mode that used to validate unix and win file related info.
	if ci.Spec != nil && len(ci.Spec.FileMode) == 0 {
		fileMode, err := dao.queryFileMode(kit, ci.ID, ci.Attachment.BizID)
		if err != nil {
			return err
		}

		ci.Spec.FileMode = fileMode
	}

	if err := ci.ValidateUpdate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentAppExist(kit, ci.Attachment); err != nil {
		return err
	}

	oldOne, err := q.Where(m.ID.Eq(ci.ID), m.BizID.Eq(ci.Attachment.BizID)).Take()
	if err != nil {
		return err
	}

	ad := dao.auditDao.DecoratorV2(kit, ci.Attachment.BizID).PrepareUpdate(ci, oldOne)

	updateTx := func(tx *gen.Query) error {
		q = tx.ConfigItem.WithContext(kit.Ctx)
		if _, err = q.Omit(m.ID, m.BizID, m.AppID).
			Where(m.ID.Eq(ci.ID), m.BizID.Eq(ci.Attachment.BizID)).Updates(ci); err != nil {
			return err
		}

		if err = ad.Do(tx); err != nil {
			return fmt.Errorf("audit update config item failed, err: %v", err)
		}
		return nil
	}

	if err = dao.genQ.Transaction(updateTx); err != nil {
		logs.Errorf("update config item: %d failed, err: %v, rid: %v", ci.ID, err, kit.Rid)
		return err
	}

	return nil
}

// BatchUpdateWithTx batch update configItem instances with transaction.
// Note: 1. this method won't audit, because it's batch operation.
// 2. this method won't validate resource update, because it's batch operation.
// 3. this method won't validate attachment resource exist, because it's batch operation.
func (dao *configItemDao) BatchUpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, configItems []*table.ConfigItem) error {
	if len(configItems) == 0 {
		return nil
	}
	if err := tx.ConfigItem.WithContext(kit.Ctx).Save(configItems...); err != nil {
		return err
	}
	return nil
}

// Get configItem by ID.
// Note: !!!current db is sharded by biz_id,it can not adapt bcs project,need redesign
func (dao *configItemDao) Get(kit *kit.Kit, id, bizID uint32) (*table.ConfigItem, error) {

	if id == 0 {
		return nil, errf.New(errf.InvalidParameter, "config item id can not be 0")
	}

	m := dao.genQ.ConfigItem
	return dao.genQ.ConfigItem.WithContext(kit.Ctx).Where(m.ID.Eq(id), m.BizID.Eq(bizID)).Take()
}

// SearchAll search all configItem by searchKey
func (dao *configItemDao) SearchAll(kit *kit.Kit, searchKey string, appID, bizID uint32) ([]*table.ConfigItem, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID can not be 0")
	}
	m := dao.genQ.ConfigItem
	query := dao.genQ.ConfigItem.WithContext(kit.Ctx).Where(m.AppID.Eq(appID), m.BizID.Eq(bizID))
	if searchKey != "" {
		searchKey = "%" + searchKey + "%"
		query = query.Where(m.Name.Like(searchKey)).Or(m.Creator.Like(searchKey)).Or(m.Reviser.Like(searchKey))
	}
	return query.Find()
}

// ListAllByAppID list all configItem by appID
func (dao *configItemDao) ListAllByAppID(kit *kit.Kit, appID uint32, bizID uint32) ([]*table.ConfigItem, error) {
	if appID == 0 {
		return nil, errf.New(errf.InvalidParameter, "appID can not be 0")
	}
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID can not be 0")
	}
	m := dao.genQ.ConfigItem
	return dao.genQ.ConfigItem.WithContext(kit.Ctx).Where(m.AppID.Eq(appID), m.BizID.Eq(bizID)).Find()
}

// Delete one configItem instance.
func (dao *configItemDao) Delete(kit *kit.Kit, ci *table.ConfigItem) error {

	if ci == nil {
		return errf.New(errf.InvalidParameter, "config item is nil")
	}

	if err := ci.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	m := dao.genQ.ConfigItem
	q := dao.genQ.ConfigItem.WithContext(kit.Ctx)

	if err := dao.validateAttachmentAppExist(kit, ci.Attachment); err != nil {
		return err
	}

	oldOne, err := q.Where(m.ID.Eq(ci.ID), m.BizID.Eq(ci.Attachment.BizID)).Take()
	if err != nil {
		return err
	}

	ad := dao.auditDao.DecoratorV2(kit, ci.Attachment.BizID).PrepareDelete(oldOne)

	// delete config item with transaction.
	deleteTx := func(tx *gen.Query) error {
		q = tx.ConfigItem.WithContext(kit.Ctx)
		if _, err = q.Where(m.ID.Eq(ci.ID), m.BizID.Eq(ci.Attachment.BizID)).Delete(); err != nil {
			return err
		}

		if err = ad.Do(tx); err != nil {
			return err
		}
		// decrease the config item lock count after the deletion
		lock := lockKey.ConfigItem(ci.Attachment.BizID, ci.Attachment.AppID)
		if err = dao.lock.DecreaseCount(kit, tx, lock, 1); err != nil {
			return err
		}
		return nil
	}
	if err = dao.genQ.Transaction(deleteTx); err != nil {
		logs.Errorf("delete config item: %d failed, err: %v, rid: %v", ci.ID, err, kit.Rid)
		return err
	}

	return nil
}

// DeleteWithTx one configItem instance with transaction.
func (dao *configItemDao) DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, ci *table.ConfigItem) error {

	if ci == nil {
		return errf.New(errf.InvalidParameter, "config item is nil")
	}

	if err := ci.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentAppExist(kit, ci.Attachment); err != nil {
		return err
	}

	m := tx.ConfigItem
	q := tx.ConfigItem.WithContext(kit.Ctx)

	oldOne, err := q.Where(m.ID.Eq(ci.ID), m.BizID.Eq(ci.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, ci.Attachment.BizID).PrepareDelete(oldOne)

	result, err := q.Where(m.ID.Eq(ci.ID), m.BizID.Eq(ci.Attachment.BizID)).Delete()
	if err != nil {
		return err
	}
	if err = ad.Do(tx.Query); err != nil {
		return err
	}

	// decrease the config item lock count after the deletion
	lock := lockKey.ConfigItem(ci.Attachment.BizID, ci.Attachment.AppID)
	if err := dao.lock.DecreaseCount(kit, tx.Query, lock, uint32(result.RowsAffected)); err != nil {
		return err
	}

	return nil
}

// BatchDeleteWithTx batch configItem instances with transaction.
func (dao *configItemDao) BatchDeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, ids []uint32, bizID, appID uint32) error {
	m := dao.genQ.ConfigItem
	q := tx.ConfigItem.WithContext(kit.Ctx)
	result, err := q.Where(m.ID.In(ids...), m.BizID.Eq(bizID), m.AppID.Eq(appID)).Delete()
	if err != nil {
		return err
	}

	// decrease the config item lock count after the deletion
	lock := lockKey.ConfigItem(bizID, appID)
	if err := dao.lock.DecreaseCount(kit, tx.Query, lock, uint32(result.RowsAffected)); err != nil {
		return err
	}
	return nil
}

// validateAttachmentResExist validate if attachment resource exists before creating config item.
func (dao *configItemDao) validateAttachmentResExist(kit *kit.Kit, am *table.ConfigItemAttachment) error {
	return dao.validateAttachmentAppExist(kit, am)
}

// validateAttachmentAppExist validate if attachment resource exists before creating config item.
func (dao *configItemDao) validateAttachmentAppExist(kit *kit.Kit, am *table.ConfigItemAttachment) error {
	m := dao.genQ.App
	// validate if config item attached app exists.
	if _, err := m.WithContext(kit.Ctx).Where(m.ID.Eq(am.AppID), m.BizID.Eq(am.BizID)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("config item attached app %d not exist", am.AppID)
		}
		return fmt.Errorf("get config item attached app %d failed", am.AppID)
	}
	return nil
}

// ValidateAppCINumber verify whether the current number of app config items has reached the maximum.
func (dao *configItemDao) ValidateAppCINumber(kt *kit.Kit, tx *gen.QueryTx, bizID, appID uint32) error {
	m := tx.ConfigItem
	count, err := m.WithContext(kt.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).Count()
	if err != nil {
		return fmt.Errorf("count app %d's config items failed, err: %v", appID, err)
	}

	if err := table.ValidateAppCINumber(count); err != nil {
		return err
	}

	return nil
}

// queryFileMode query config item file mode field.
func (dao *configItemDao) queryFileMode(kt *kit.Kit, id, bizID uint32) (
	table.FileMode, error) {

	m := dao.genQ.ConfigItem

	ci, err := dao.genQ.ConfigItem.WithContext(kt.Ctx).Where(m.ID.Eq(id), m.BizID.Eq(bizID)).Take()
	if err != nil {
		return "", errf.New(errf.DBOpFailed, fmt.Sprintf("get config item %d file mode failed", id))
	}

	if err := ci.Spec.FileMode.Validate(); err != nil {
		return "", errf.New(errf.InvalidParameter, err.Error())
	}

	return ci.Spec.FileMode, nil
}

// GetCount get bizID config count
func (dao *configItemDao) GetCount(kit *kit.Kit, bizID uint32, appId []uint32) ([]*table.ListConfigItemCounts, error) {

	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "config item biz id can not be 0")
	}

	configItem := make([]*table.ListConfigItemCounts, 0)

	m := dao.genQ.ConfigItem
	q := dao.genQ.ConfigItem.WithContext(kit.Ctx)
	if err := q.Select(m.AppID, m.ID.Count().As("count"), m.UpdatedAt.Max().As("updated_at")).
		Where(m.BizID.Eq(bizID), m.AppID.In(appId...)).Group(m.AppID).Scan(&configItem); err != nil {
		return nil, err
	}

	return configItem, nil
}
