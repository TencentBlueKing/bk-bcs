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
	"fmt"

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
	// Create one configItem instance.
	Create(kit *kit.Kit, configItem *table.ConfigItem) (uint32, error)
	// Update one configItem instance.
	Update(kit *kit.Kit, configItem *table.ConfigItem) error
	// List configItem with options.
	List(kit *kit.Kit, opts *types.ListConfigItemsOption) (*types.ListConfigItemDetails, error)
	// Delete one configItem instance.
	Delete(kit *kit.Kit, configItem *table.ConfigItem) error
}

var _ ConfigItem = new(configItemDao)

type configItemDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	lock     LockDao
}

// Create one configItem instance.
func (dao *configItemDao) Create(kit *kit.Kit, ci *table.ConfigItem) (uint32, error) {

	if ci == nil {
		return 0, errf.New(errf.InvalidParameter, "config item is nil")
	}

	if err := ci.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
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

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.ConfigItemTable,
		table.ConfigItemColumns.ColumnExpr(), table.ConfigItemColumns.ColonNameExpr())

	err = dao.sd.ShardingOne(ci.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			if err := dao.validateAppCINumber(kit, ci.Attachment, &LockOption{Txn: txn}); err != nil {
				return err
			}

			if err := dao.orm.Txn(txn).Insert(kit.Ctx, sql, ci); err != nil {
				return err
			}

			// audit this to be create config item details.
			au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
			if err := dao.auditDao.Decorator(kit, ci.Attachment.BizID,
				enumor.ConfigItem).AuditCreate(ci, au); err != nil {
				return fmt.Errorf("audit create config item failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		logs.Errorf("create config item, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create config item, but auto run txn failed, err: %v", err)
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

	sql := fmt.Sprintf(`UPDATE %s SET %s WHERE id = %d AND biz_id = %d`,
		table.ConfigItemTable, expr, ci.ID, ci.Attachment.BizID)

	err = dao.sd.ShardingOne(ci.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			effected, err := dao.orm.Txn(txn).Update(kit.Ctx, sql, toUpdate)
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

// List configItems with options.
func (dao *configItemDao) List(kit *kit.Kit, opts *types.ListConfigItemsOption) (
	*types.ListConfigItemDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list config item options null")
	}

	if err := opts.Validate(types.DefaultPageOption); err != nil {
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
	whereExpr, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sql string
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sql = fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ConfigItemTable, whereExpr)
		count, err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, sql)
		if err != nil {
			return nil, err
		}

		return &types.ListConfigItemDetails{Count: count, Details: make([]*table.ConfigItem, 0)}, nil
	}

	// query config item list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sql = fmt.Sprintf(`SELECT %s FROM %s %s %s`,
		table.ConfigItemColumns.NamedExpr(), table.ConfigItemTable, whereExpr, pageExpr)

	list := make([]*table.ConfigItem, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql)
	if err != nil {
		return nil, err
	}

	return &types.ListConfigItemDetails{Count: 0, Details: list}, nil
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

	expr := fmt.Sprintf(`DELETE FROM %s WHERE id = %d AND biz_id = %d`, table.ConfigItemTable,
		ci.ID, ci.Attachment.BizID)

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

	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable,
		fmt.Sprintf("WHERE id = %d AND biz_id = %d", am.AppID, am.BizID))
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

	expr := fmt.Sprintf(`SELECT %s FROM %s WHERE id = %d AND biz_id = %d`, table.ConfigItemSpecColumns.NamedExpr(),
		table.ConfigItemTable, id, bizID)

	one := new(table.ConfigItemSpec)
	if err := dao.orm.Do(dao.sd.MustSharding(bizID)).Get(kt.Ctx, one, expr); err != nil {
		return "", errf.New(errf.DBOpFailed, fmt.Sprintf("get file mode failed, err: %v", err))
	}

	if err := one.FileMode.Validate(); err != nil {
		return "", errf.New(errf.InvalidParameter, err.Error())
	}

	return one.FileMode, nil
}
