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

// StrategySet supplies all the strategy set related operations.
type StrategySet interface {
	// Create one strategy set instance.
	Create(kit *kit.Kit, strategySet *table.StrategySet) (uint32, error)
	// Update one strategy set's info.
	Update(kit *kit.Kit, strategySet *table.StrategySet) error
	// List strategy sets with options.
	List(kit *kit.Kit, opts *types.ListStrategySetsOption) (*types.ListStrategySetDetails, error)
	// Delete one strategy instance.
	Delete(kit *kit.Kit, strategy *table.StrategySet) error
}

var _ StrategySet = new(strategySetDao)

type strategySetDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	lock     LockDao
}

// Create one strategy set instance.
func (dao *strategySetDao) Create(kit *kit.Kit, ss *table.StrategySet) (uint32, error) {

	if ss == nil {
		return 0, errf.New(errf.InvalidParameter, "strategy set is nil")
	}

	mode, err := getAppMode(kit, dao.orm, dao.sd, ss.Attachment.BizID, ss.Attachment.AppID)
	if err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	// set the strategy set's mode.
	ss.Spec.Mode = mode

	if err = ss.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	if err = dao.validateAttachmentResExist(kit, ss.Attachment); err != nil {
		return 0, err
	}

	// generate a strategy set id and update to strategy set.
	id, err := dao.idGen.One(kit, table.StrategySetTable)
	if err != nil {
		return 0, err
	}

	ss.ID = id

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", string(table.StrategySetTable), " (", table.StrategySetColumns.ColumnExpr(),
		") ", " VALUES(", table.StrategySetColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)
	err = dao.sd.ShardingOne(ss.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			if err = dao.validateAppStrategySetNumber(kit, ss.Attachment, &LockOption{Txn: txn}); err != nil {
				return err
			}

			if err := dao.orm.Txn(txn).Insert(kit.Ctx, sql, ss); err != nil {
				return err
			}

			// audit this to be created strategy set details.
			au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
			if err = dao.auditDao.Decorator(kit, ss.Attachment.BizID,
				enumor.StrategySet).AuditCreate(ss, au); err != nil {
				return fmt.Errorf("audit create strategy set failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		logs.Errorf("create strategy set, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create strategy set, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// Update one strategy set instance.
func (dao *strategySetDao) Update(kit *kit.Kit, ss *table.StrategySet) error {

	if ss == nil {
		return errf.New(errf.InvalidParameter, "strategy set is nil")
	}

	if err := ss.ValidateUpdate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentAppExist(kit, ss.Attachment); err != nil {
		return err
	}

	opts := orm.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(
		"id", "biz_id", "app_id")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(ss, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	ab := dao.auditDao.Decorator(kit, ss.Attachment.BizID, enumor.StrategySet).PrepareUpdate(ss)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", string(table.StrategySetTable), " SET ", expr, " WHERE id = ", strconv.Itoa(int(ss.ID)), " AND biz_id = ", strconv.Itoa(int(ss.Attachment.BizID)))
	sql := filter.SqlJoint(sqlSentence)
	err = dao.sd.ShardingOne(ss.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			var effected int64
			effected, err = dao.orm.Txn(txn).Update(kit.Ctx, sql, toUpdate)
			if err != nil {
				logs.Errorf("update strategy set: %d failed, err: %v, rid: %v", ss.ID, err, kit.Rid)
				return err
			}

			if effected == 0 {
				logs.Errorf("update one strategy set: %d, but record not found, rid: %v", ss.ID, kit.Rid)
				return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
			}

			if effected > 1 {
				logs.Errorf("update one strategy set: %d, but got updated strategy set count: %d, rid: %v", ss.ID,
					effected, kit.Rid)
				return fmt.Errorf("matched strategy set count %d is not as excepted", effected)
			}

			// do audit
			if err := ab.Do(&AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}); err != nil {
				return fmt.Errorf("do strategy set update audit failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

// List strategy sets with options.
func (dao *strategySetDao) List(kit *kit.Kit, opts *types.ListStrategySetsOption) (
	*types.ListStrategySetDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list strategy set options null")
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
	whereExpr, args, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sqlSentence []string
	var sql string
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sqlSentence = append(sqlSentence, "SELECT COUNT(*) FROM ", string(table.StrategySetTable), whereExpr)
		sql = filter.SqlJoint(sqlSentence)
		var count uint32
		count, err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, sql, args...)
		if err != nil {
			return nil, err
		}

		return &types.ListStrategySetDetails{Count: count, Details: make([]*table.StrategySet, 0)}, nil
	}

	// query strategy set list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sqlSentence = append(sqlSentence, "SELECT ", table.StrategySetColumns.NamedExpr(), " FROM ", string(table.StrategySetTable), whereExpr, pageExpr)
	sql = filter.SqlJoint(sqlSentence)
	list := make([]*table.StrategySet, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListStrategySetDetails{Count: 0, Details: list}, nil
}

// Delete one strategy set instance.
func (dao *strategySetDao) Delete(kit *kit.Kit, ss *table.StrategySet) error {

	if ss == nil {
		return errf.New(errf.InvalidParameter, "strategy set is nil")
	}

	if err := ss.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentAppExist(kit, ss.Attachment); err != nil {
		return err
	}

	// validate strategy set under if strategy not exist.
	if err := dao.validateStrategyNotExist(kit, ss.ID, ss.Attachment.BizID); err != nil {
		return err
	}

	ab := dao.auditDao.Decorator(kit, ss.Attachment.BizID, enumor.StrategySet).PrepareDelete(ss.ID)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", string(table.StrategySetTable), " WHERE id = ", strconv.Itoa(int(ss.ID)), " AND biz_id = ", strconv.Itoa(int(ss.Attachment.BizID)))
	sql := filter.SqlJoint(sqlSentence)

	err := dao.sd.ShardingOne(ss.Attachment.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		// delete the strategy set at first.
		err := dao.orm.Txn(txn).Delete(kit.Ctx, sql)
		if err != nil {
			return err
		}

		// audit this delete strategy set details.
		auditOpt := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err := ab.Do(auditOpt); err != nil {
			return fmt.Errorf("audit delete strategy set failed, err: %v", err)
		}

		// decrease the strategy set lock count after the deletion
		lock := lockKey.StrategySet(ss.Attachment.BizID, ss.Attachment.AppID)
		if err := dao.lock.DecreaseCount(kit, lock, &LockOption{Txn: txn}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logs.Errorf("delete strategy set: %d failed, err: %v, rid: %v", ss.ID, err, kit.Rid)
		return fmt.Errorf("delete strategy set, but run txn failed, err: %v", err)
	}

	return nil
}

// validateAttachmentResExist validate if attachment resource exists before creating strategy set.
func (dao *strategySetDao) validateAttachmentResExist(kit *kit.Kit, am *table.StrategySetAttachment) error {
	return dao.validateAttachmentAppExist(kit, am)
}

// validateAttachmentAppExist validate if attachment app exists before creating strategy set.
func (dao *strategySetDao) validateAttachmentAppExist(kit *kit.Kit, am *table.StrategySetAttachment) error {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, " WHERE id = ", strconv.Itoa(int(am.AppID)), " AND biz_id = ", strconv.Itoa(int(am.BizID)))
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable, sql)
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("strategy set attached app %d is not exist", am.AppID))
	}

	return nil
}

// validateAppStrategySetNumber currently, only one strategy set is allowed to create in an application.
func (dao *strategySetDao) validateAppStrategySetNumber(kt *kit.Kit, at *table.StrategySetAttachment,
	lo *LockOption) error {

	// try lock strategy set to ensure the number is limited when creating concurrently
	lock := lockKey.StrategySet(at.BizID, at.AppID)
	count, err := dao.lock.IncreaseCount(kt, lock, lo)
	if err != nil {
		return err
	}

	if err := table.ValidateAppStrategySetNumber(count); err != nil {
		return err
	}

	return nil
}

// validateStrategyNotExist validate this strategy set under if strategy not exist.
func (dao *strategySetDao) validateStrategyNotExist(kt *kit.Kit, bizID, id uint32) error {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, " WHERE biz_id = ", strconv.Itoa(int(bizID)), " AND strategy_set_id = %d", strconv.Itoa(int(id)))
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kt, dao.orm, dao.sd.ShardingOne(bizID), table.StrategyTable, sql)
	if err != nil {
		return err
	}

	if exist {
		return errf.New(errf.InvalidParameter, "there are still strategy under the current strategy set")
	}

	return nil
}
