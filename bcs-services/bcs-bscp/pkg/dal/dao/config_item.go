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

	"github.com/samber/lo"
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
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
	// GetByUniqueKey get config item by unique key
	GetByUniqueKey(kit *kit.Kit, bizID, appID uint32, name, path string) (*table.ConfigItem, error)
	// GetUniqueKeys get unique keys of all config items in one app
	GetUniqueKeys(kit *kit.Kit, bizID, appID uint32) ([]types.CIUniqueKey, error)
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
	// ListConfigItemByTuple 按照多个字段in查询config item 列表
	ListConfigItemByTuple(kit *kit.Kit, data [][]interface{}) ([]*table.ConfigItem, error)
	// RecoverConfigItem 恢复单个配置项(恢复的配置项使用原来的ID)
	RecoverConfigItem(kit *kit.Kit, tx *gen.QueryTx, configItem *table.ConfigItem) error
	// UpdateWithTx one configItem instance with transaction.
	UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, configItem *table.ConfigItem) error
	// GetConfigItemCount 获取配置项数量
	GetConfigItemCount(kit *kit.Kit, bizID uint32, appID uint32) (int64, error)
	// ListConfigItemCount 展示配置项数量
	ListConfigItemCount(kit *kit.Kit, bizID uint32, appID []uint32) ([]types.ListConfigItemCount, error)
	// GetConfigItemCount 获取配置项数量带有事务
	GetConfigItemCountWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID uint32, appID uint32) (int64, error)
}

var _ ConfigItem = new(configItemDao)

type configItemDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
	lock     LockDao
}

// GetConfigItemCountWithTx 获取配置项数量带有事务
func (dao *configItemDao) GetConfigItemCountWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID uint32,
	appID uint32) (int64, error) {

	m := dao.genQ.ConfigItem

	return tx.ConfigItem.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).
		Count()
}

// ListConfigItemCount 展示配置项数量
func (dao *configItemDao) ListConfigItemCount(kit *kit.Kit, bizID uint32, appID []uint32) (
	[]types.ListConfigItemCount, error) {
	m := dao.genQ.ConfigItem

	var result []types.ListConfigItemCount
	err := dao.genQ.ConfigItem.WithContext(kit.Ctx).Select(m.AppID, m.ID.Count().As("count")).
		Where(m.BizID.Eq(bizID), m.AppID.In(appID...)).Group(m.AppID).Scan(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetConfigItemCount 获取配置项数量
func (dao *configItemDao) GetConfigItemCount(kit *kit.Kit, bizID uint32, appID uint32) (int64, error) {

	m := dao.genQ.ConfigItem

	return dao.genQ.ConfigItem.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).
		Count()
}

// UpdateWithTx one configItem instance with transaction.
func (dao *configItemDao) UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, ci *table.ConfigItem) error {
	if ci == nil {
		return errf.New(errf.InvalidParameter, "config item is nil")
	}

	m := tx.ConfigItem
	q := tx.ConfigItem.WithContext(kit.Ctx)

	// if file mode not update, need to query this ci's file mode that used to validate unix and win file related info.
	if ci.Spec != nil && len(ci.Spec.FileMode) == 0 {
		fileMode, err := dao.queryFileMode(kit, ci.ID, ci.Attachment.BizID)
		if err != nil {
			return err
		}

		ci.Spec.FileMode = fileMode
	}

	if err := ci.ValidateUpdate(kit); err != nil {
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
		if _, err = q.Select(m.Name, m.Path, m.FileType, m.FileMode, m.Memo, m.User, m.UserGroup,
			m.Privilege, m.Reviser, m.UpdatedAt).
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

// RecoverConfigItem 恢复单个配置项(恢复的配置项使用原来的ID)
func (dao *configItemDao) RecoverConfigItem(kit *kit.Kit, tx *gen.QueryTx, ci *table.ConfigItem) error {
	if ci == nil {
		return errors.New("config item is nil")
	}

	if err := ci.ValidateRecover(kit); err != nil {
		return err
	}

	if err := dao.validateAttachmentResExist(kit, ci.Attachment); err != nil {
		return err
	}

	ad := dao.auditDao.DecoratorV2(kit, ci.Attachment.BizID).PrepareCreate(ci)

	if err := tx.ConfigItem.WithContext(kit.Ctx).Create(ci); err != nil {
		return err
	}

	if err := ad.Do(tx.Query); err != nil {
		return fmt.Errorf("audit recover config item failed, err: %v", err)
	}

	return nil
}

// ListConfigItemByTuple 按照多个字段in查询config item 列表
func (dao *configItemDao) ListConfigItemByTuple(kit *kit.Kit, data [][]interface{}) ([]*table.ConfigItem, error) {
	m := dao.genQ.ConfigItem
	return dao.genQ.ConfigItem.WithContext(kit.Ctx).
		Select(m.ID, m.BizID, m.AppID, m.Name, m.Path, m.FileType, m.User,
			m.UserGroup, m.Privilege, m.Memo, m.FileMode).
		Where(m.WithContext(kit.Ctx).Columns(m.BizID, m.AppID, m.Name, m.Path).
			In(field.Values(data))).
		Find()
}

// CreateWithTx create one configItem instance with transaction.
func (dao *configItemDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, ci *table.ConfigItem) (uint32, error) {
	if ci == nil {
		return 0, errors.New("config item is nil")
	}

	if err := ci.ValidateCreate(kit); err != nil {
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
		if err := configItem.ValidateCreate(kit); err != nil {
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

	if err := ci.ValidateUpdate(kit); err != nil {
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

// GetByUniqueKey get config item by unique key
func (dao *configItemDao) GetByUniqueKey(kit *kit.Kit, bizID, appID uint32, name, path string) (
	*table.ConfigItem, error) {
	m := dao.genQ.ConfigItem
	q := dao.genQ.ConfigItem.WithContext(kit.Ctx)

	configItem, err := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.Name.Eq(name),
		m.Path.Eq(path)).Take()
	if err != nil {
		return nil, err
	}

	return configItem, nil
}

// GetUniqueKeys get unique keys of all config items in one app
func (dao *configItemDao) GetUniqueKeys(kit *kit.Kit, bizID, appID uint32) ([]types.CIUniqueKey, error) {
	m := dao.genQ.ConfigItem
	q := dao.genQ.ConfigItem.WithContext(kit.Ctx)

	var rs []types.CIUniqueKey
	if err := q.Select(m.Name, m.Path).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).Scan(&rs); err != nil {
		return nil, err
	}
	return rs, nil
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

	_, err = q.Where(m.ID.Eq(ci.ID), m.BizID.Eq(ci.Attachment.BizID)).Delete()
	if err != nil {
		return err
	}
	if err = ad.Do(tx.Query); err != nil {
		return err
	}

	return nil
}

// BatchDeleteWithTx batch configItem instances with transaction.
func (dao *configItemDao) BatchDeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, ids []uint32, bizID, appID uint32) error {
	m := dao.genQ.ConfigItem
	q := tx.ConfigItem.WithContext(kit.Ctx)
	_, err := q.Where(m.ID.In(ids...), m.BizID.Eq(bizID), m.AppID.Eq(appID)).Delete()
	return err
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
// the number is the total count of template and non-template config items
func (dao *configItemDao) ValidateAppCINumber(kt *kit.Kit, tx *gen.QueryTx, bizID, appID uint32) error {
	// get non-template config count
	m := tx.ConfigItem
	count, err := m.WithContext(kt.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).Count()
	if err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kt, "count app %d's config items failed, err: %v", appID, err))
	}

	// get template config count
	tm := tx.AppTemplateBinding
	tcount := 0
	var atb *table.AppTemplateBinding
	atb, err = tm.WithContext(kt.Ctx).Where(tm.BizID.Eq(bizID), tm.AppID.Eq(appID)).Take()
	if err != nil {
		// if not found, means the count should be 0
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return errf.Errorf(errf.InvalidRequest, i18n.T(kt, "get app %d's template binding failed, err: %v", appID, err))
		}
	} else {
		tcount = len(atb.Spec.TemplateRevisionIDs)
	}

	total := int(count) + tcount
	appConfigCnt := getAppConfigCnt(bizID)
	if total > appConfigCnt {
		return errf.New(errf.InvalidParameter,
			i18n.T(kt, "the total number of app %d's config items(including template and non-template)exceeded the limit %d",
				appID, appConfigCnt))
	}

	return nil
}

func getAppConfigCnt(bizID uint32) int {
	if resLimit, ok := cc.DataService().FeatureFlags.ResourceLimit.Spec[fmt.Sprintf("%d", bizID)]; ok {
		if resLimit.AppConfigCnt > 0 {
			return int(resLimit.AppConfigCnt)
		}
	}
	return int(cc.DataService().FeatureFlags.ResourceLimit.Default.AppConfigCnt)
}

// queryFileMode query config item file mode field.
func (dao *configItemDao) queryFileMode(kt *kit.Kit, id, bizID uint32) (
	table.FileMode, error) {

	m := dao.genQ.ConfigItem

	ci, err := dao.genQ.ConfigItem.WithContext(kt.Ctx).Where(m.ID.Eq(id), m.BizID.Eq(bizID)).Take()
	if err != nil {
		return "", errf.New(errf.DBOpFailed, fmt.Sprintf("get config item %d file mode failed", id))
	}

	if err := ci.Spec.FileMode.Validate(kt); err != nil {
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

	ma := dao.genQ.AppTemplateBinding
	qa := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	result, err := qa.Select(ma.AppID, ma.TemplateIDs, ma.UpdatedAt).Where(
		ma.BizID.Eq(bizID), ma.AppID.In(appId...)).Find()
	if err != nil {
		return nil, err
	}

	// 泛型处理
	configItemMap := lo.KeyBy(configItem, func(c *table.ListConfigItemCounts) uint32 {
		return c.AppId
	})

	// 累加配置模板计数
	for _, r := range result {
		appID := r.AppID()
		if _, ok := configItemMap[appID]; ok {
			configItemMap[appID].Count += uint32(len(r.Spec.TemplateIDs))
		} else {
			configItemMap[appID] = &table.ListConfigItemCounts{
				AppId:     appID,
				Count:     uint32(len(r.Spec.TemplateIDs)),
				UpdatedAt: r.Revision.UpdatedAt,
			}
		}
	}

	v := lo.Values(configItemMap)
	return v, nil
}
