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
	"strconv"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"

	"github.com/jmoiron/sqlx"
)

// ConfigItem supplies all the configItem related operations.
type ConfigItem interface {
	// CreateWithTx create one configItem instance.
	CreateWithTx(kit *kit.Kit, tx *sharding.Tx, configItem *table.ConfigItem) (uint32, error)
	// Update one configItem instance.
	Update(kit *kit.Kit, configItem *table.ConfigItem) error
	// Get configItem by id
	Get(kit *kit.Kit, id, bizID uint32) (*table.ConfigItem, error)
	// List configItem with options.
	List(kit *kit.Kit, opts *types.ListConfigItemsOption) (*types.ListConfigItemDetails, error)
	// Delete one configItem instance.
	Delete(kit *kit.Kit, configItem *table.ConfigItem) error
	// GetCount bizID config count
	GetCount(kit *kit.Kit, bizID uint32, appId []uint32) ([]*table.ListConfigItemCounts, error)
	// TruncateWithTx truncate app config items with transaction.
	TruncateWithTx(kit *kit.Kit, tx *sharding.Tx, bizID, appID uint32) error
}

var _ ConfigItem = new(configItemDao)

type configItemDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	lock     LockDao
}

// CreateWithTx create one configItem instance with transaction.
func (dao *configItemDao) CreateWithTx(kit *kit.Kit, tx *sharding.Tx, ci *table.ConfigItem) (uint32, error) {
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
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.ConfigItemTable.Name(),
		" (", table.ConfigItemColumns.ColumnExpr(), ")  VALUES(", table.ConfigItemColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)

	if err = dao.validateAppCINumber(kit, ci.Attachment, &LockOption{Txn: tx.Tx()}); err != nil {
		return 0, err
	}

	if e := dao.orm.Txn(tx.Tx()).Insert(kit.Ctx, sql, ci); e != nil {
		return 0, err
	}

	// audit this to be create config item details.
	au := &AuditOption{Txn: tx.Tx(), ResShardingUid: tx.ShardingUid()}
	if err = dao.auditDao.Decorator(kit, ci.Attachment.BizID,
		enumor.ConfigItem).AuditCreate(ci, au); err != nil {
		return 0, fmt.Errorf("audit create config item failed, err: %v", err)
	}

	return id, nil
}

// Update one configItem instance.
func (dao *configItemDao) Update(kit *kit.Kit, ci *table.ConfigItem) error {

	if ci == nil {
		return errf.New(errf.InvalidParameter, "config item is nil")
	}

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

	opts := orm.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields("id", "biz_id", "app_id")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(ci, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	ab := dao.auditDao.Decorator(kit, ci.Attachment.BizID, enumor.ConfigItem).PrepareUpdate(ci)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.ConfigItemTable.Name(), " SET ", expr, " WHERE id = ", strconv.Itoa(int(ci.ID)), " AND biz_id = ", strconv.Itoa(int(ci.Attachment.BizID)))
	sql := filter.SqlJoint(sqlSentence)

	err = dao.sd.ShardingOne(ci.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			var effected int64
			effected, err = dao.orm.Txn(txn).Update(kit.Ctx, sql, toUpdate)
			if err != nil {
				logs.Errorf("update config item: %d failed, err: %v, rid: %v", ci.ID, err, kit.Rid)
				return err
			}

			if effected == 0 {
				logs.Errorf("update one config item: %d, but record not found, rid: %v", ci.ID, kit.Rid)
				return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
			}

			if effected > 1 {
				logs.Errorf("update one config item: %d, but got updated config item count: %d, rid: %v", ci.ID,
					effected, kit.Rid)
				return fmt.Errorf("matched config item count %d is not as excepted", effected)
			}

			// do audit
			if err := ab.Do(&AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}); err != nil {
				return fmt.Errorf("do config item update audit failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

// Get configItem by ID.
// TODO: !!!current db is sharded by biz_id,it can not adapt bcs project,need redesign
func (dao *configItemDao) Get(kit *kit.Kit, id, bizID uint32) (*table.ConfigItem, error) {

	if id == 0 {
		return nil, errf.New(errf.InvalidParameter, "config item id can not be 0")
	}
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.ConfigItemColumns.NamedExpr(), " FROM ", table.ConfigItemTable.Name(), " WHERE id = ", strconv.Itoa(int(id)))
	sql := filter.SqlJoint(sqlSentence)

	configItem := &table.ConfigItem{}
	if err := dao.orm.Do(dao.sd.ShardingOne(bizID).DB()).Get(kit.Ctx, configItem, sql); err != nil {
		return nil, err
	}
	return configItem, nil
}

// List configItems with options.
func (dao *configItemDao) List(kit *kit.Kit, opts *types.ListConfigItemsOption) (
	*types.ListConfigItemDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list config item options null")
	}

	po := &types.PageOption{
		EnableUnlimitedLimit: true,
		DisabledSort:         false,
	}

	if err := opts.Validate(po); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id", "biz_id", "app_id"},
		CrownedOption: &filter.CrownedOption{
			CrownedOp: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "biz_id",
					Op:    filter.Equal.Factory(),
					Value: opts.BizID,
				},
				&filter.AtomRule{
					Field: "app_id",
					Op:    filter.Equal.Factory(),
					Value: opts.AppID,
				},
			},
		},
	}
	whereExpr, args, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sqlSentenceCount []string
	sqlSentenceCount = append(sqlSentenceCount, "SELECT COUNT(*) FROM ", table.ConfigItemTable.Name(), whereExpr)
	countSql := filter.SqlJoint(sqlSentenceCount)
	count, err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, countSql, args...)
	if err != nil {
		return nil, err
	}

	// query config item list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.ConfigItemColumns.NamedExpr(), " FROM ", table.ConfigItemTable.Name(), whereExpr, pageExpr)
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*table.ConfigItem, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListConfigItemDetails{Count: count, Details: list}, nil
}

// Delete one configItem instance.
func (dao *configItemDao) Delete(kit *kit.Kit, ci *table.ConfigItem) error {

	if ci == nil {
		return errf.New(errf.InvalidParameter, "config item is nil")
	}

	if err := ci.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentAppExist(kit, ci.Attachment); err != nil {
		return err
	}

	ab := dao.auditDao.Decorator(kit, ci.Attachment.BizID, enumor.ConfigItem).PrepareDelete(ci.ID)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.ConfigItemTable.Name(), " WHERE id = ", strconv.Itoa(int(ci.ID)), " AND biz_id = ", strconv.Itoa(int(ci.Attachment.BizID)))
	expr := filter.SqlJoint(sqlSentence)

	err := dao.sd.ShardingOne(ci.Attachment.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		// delete the config item at first.
		err := dao.orm.Txn(txn).Delete(kit.Ctx, expr)
		if err != nil {
			return err
		}

		// audit this delete config item details.
		auditOpt := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err := ab.Do(auditOpt); err != nil {
			return fmt.Errorf("audit delete config item failed, err: %v", err)
		}

		// decrease the config item lock count after the deletion
		lock := lockKey.ConfigItem(ci.Attachment.BizID, ci.Attachment.AppID)
		if err := dao.lock.DecreaseCount(kit, lock, &LockOption{Txn: txn}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logs.Errorf("delete config item: %d failed, err: %v, rid: %v", ci.ID, err, kit.Rid)
		return fmt.Errorf("delete config item, but run txn failed, err: %v", err)
	}

	return nil
}

// validateAttachmentResExist validate if attachment resource exists before creating config item.
func (dao *configItemDao) validateAttachmentResExist(kit *kit.Kit, am *table.ConfigItemAttachment) error {
	return dao.validateAttachmentAppExist(kit, am)
}

// validateAttachmentAppExist validate if attachment resource exists before creating config item.
func (dao *configItemDao) validateAttachmentAppExist(kit *kit.Kit, am *table.ConfigItemAttachment) error {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "WHERE id = ", strconv.Itoa(int(am.AppID)), " AND biz_id = ", strconv.Itoa(int(am.BizID)))
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable, sql)
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("config item attached app %d is not exist", am.AppID))
	}

	return nil
}

// validateAppCINumber verify whether the current number of app config items has reached the maximum.
func (dao *configItemDao) validateAppCINumber(kt *kit.Kit, at *table.ConfigItemAttachment, lo *LockOption) error {
	// try lock config item to ensure the number is limited when creating concurrently
	lock := lockKey.ConfigItem(at.BizID, at.AppID)
	count, err := dao.lock.IncreaseCount(kt, lock, lo)
	if err != nil {
		return err
	}

	if err := table.ValidateAppCINumber(count); err != nil {
		return err
	}

	return nil
}

// queryFileMode query config item file mode field.
func (dao *configItemDao) queryFileMode(kt *kit.Kit, id, bizID uint32) (
	table.FileMode, error) {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.ConfigItemSpecColumns.NamedExpr(), " FROM ", table.ConfigItemTable.Name(), " WHERE id = ", strconv.Itoa(int(id)), " AND biz_id = ", strconv.Itoa(int(bizID)))
	expr := filter.SqlJoint(sqlSentence)

	one := new(table.ConfigItemSpec)
	if err := dao.orm.Do(dao.sd.MustSharding(bizID)).Get(kt.Ctx, one, expr); err != nil {
		return "", errf.New(errf.DBOpFailed, fmt.Sprintf("get file mode failed, err: %v", err))
	}

	if err := one.FileMode.Validate(); err != nil {
		return "", errf.New(errf.InvalidParameter, err.Error())
	}

	return one.FileMode, nil
}

// GetCount get bizID config count
func (dao *configItemDao) GetCount(kit *kit.Kit, bizID uint32, appId []uint32) ([]*table.ListConfigItemCounts, error) {

	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "config item biz id can not be 0")
	}

	expr := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			&filter.AtomRule{
				Field: "biz_id",
				Op:    filter.Equal.Factory(),
				Value: bizID,
			},
			&filter.AtomRule{
				Field: "app_id",
				Op:    filter.In.Factory(),
				Value: appId,
			},
		},
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"biz_id", "app_id"},
	}
	whereExpr, args, err := expr.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT app_id, COUNT(*) as count, max(updated_at) as update_at FROM ", table.ConfigItemTable.Name(), whereExpr, " GROUP BY app_id")
	sql := filter.SqlJoint(sqlSentence)

	configItem := make([]*table.ListConfigItemCounts, 0)
	if err := dao.orm.Do(dao.sd.ShardingOne(bizID).DB()).Select(kit.Ctx, &configItem, sql, args...); err != nil {
		return nil, err
	}
	return configItem, nil
}

// TruncateWithTx delete all config item by bizID and appID
func (dao *configItemDao) TruncateWithTx(kit *kit.Kit, tx *sharding.Tx, bizID, appID uint32) error {

	if bizID == 0 || appID == 0 {
		return errf.New(errf.InvalidParameter, "config item biz id or app id can not be 0")
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.ConfigItemTable.Name(),
		" WHERE app_id = ", strconv.Itoa(int(appID)), " AND biz_id = ", strconv.Itoa(int(bizID)))
	expr := filter.SqlJoint(sqlSentence)

	err := dao.orm.Txn(tx.Tx()).Delete(kit.Ctx, expr)
	if err != nil {
		return err
	}

	// decrease the config item lock count after the deletion
	lock := lockKey.ConfigItem(bizID, appID)
	if err := dao.lock.TruncateCount(kit, lock, &LockOption{Txn: tx.Tx()}); err != nil {
		return err
	}

	return nil
}
