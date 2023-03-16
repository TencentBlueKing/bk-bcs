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
	"time"

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

// Strategy supplies all the strategy related operations.
type Strategy interface {
	// Create one strategy instance.
	Create(kit *kit.Kit, strategy *table.Strategy) (uint32, error)
	// Update one strategy's info.
	Update(kit *kit.Kit, strategy *table.Strategy) error
	// List strategy with options.
	List(kit *kit.Kit, opts *types.ListStrategiesOption) (*types.ListStrategyDetails, error)
	// Delete one strategy instance.
	Delete(kit *kit.Kit, strategy *table.Strategy) error
}

var _ Strategy = new(strategyDao)

type strategyDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	event    Event
	lock     LockDao
}

// Create one strategy instance.
func (dao *strategyDao) Create(kt *kit.Kit, strategy *table.Strategy) (uint32, error) {

	if strategy == nil {
		return 0, errf.New(errf.InvalidParameter, "strategy is nil")
	}

	// strategy type inherits strategy set's.
	mode, err := getAppMode(kt, dao.orm, dao.sd, strategy.Attachment.BizID, strategy.Attachment.AppID)
	if err != nil {
		return 0, err
	}
	strategy.Spec.Mode = mode

	if err = strategy.ValidateCreate(); err != nil {
		return 0, err
	}

	if mode == table.Namespace && strategy.Spec.AsDefault {
		// this strategy is a default strategy, then set its namespace to the
		// system reserved default namespace manually.
		strategy.Spec.Namespace = table.DefaultNamespace
	}

	if err = dao.validateAttachmentResExist(kt, strategy.Attachment); err != nil {
		return 0, err
	}

	// validate strategy binding release exist.
	if err = dao.validateReleaseExist(kt, strategy); err != nil {
		return 0, err
	}

	// generate the strategy id and update to strategy.
	id, err := dao.idGen.One(kt, table.StrategyTable)
	if err != nil {
		return 0, err
	}
	strategy.ID = id

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.StrategyTable.Name(), " (", table.StrategyColumns.ColumnExpr(), ") ",
		"VALUES(", table.StrategyColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)
	err = dao.sd.ShardingOne(strategy.Attachment.BizID).AutoTxn(kt,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			lo := &LockOption{Txn: txn}

			// validate default strategy already exist.
			if strategy.Spec.AsDefault {
				if err = dao.validateDefaultStrategyNotExist(kt, strategy.Attachment, lo); err != nil {
					return err
				}
			}

			// validate namespace only under strategy set.
			if !strategy.Spec.AsDefault && mode == table.Namespace {
				if err = dao.validateNamespaceExist(kt, strategy.Attachment, strategy.Spec.Namespace, lo); err != nil {
					return err
				}
			}

			if err = dao.validateAppStrategyNumber(kt, strategy.Attachment, strategy.Spec.Mode, lo); err != nil {
				return err
			}

			if err = dao.orm.Txn(txn).Insert(kt.Ctx, sql, strategy); err != nil {
				return err
			}

			// audit this to create strategy details.
			au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
			if err = dao.auditDao.Decorator(kt, strategy.Attachment.BizID,
				enumor.Strategy).AuditCreate(strategy, au); err != nil {
				return fmt.Errorf("audit create strategy failed, err: %v", err)
			}

			return nil
		})
	if err != nil {
		logs.Errorf("create strategy, but do auto txn failed, err: %v, rid: %s", err, kt.Rid)
		return 0, fmt.Errorf("create strategy, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// Update one strategy instance.
func (dao *strategyDao) Update(kit *kit.Kit, strategy *table.Strategy) error {

	if strategy == nil {
		return errf.New(errf.InvalidParameter, "strategy is nil")
	}

	s, err := dao.getStrategy(kit, strategy.Attachment.BizID, strategy.Attachment.AppID, strategy.ID)
	if err != nil {
		return err
	}

	if err = strategy.ValidateUpdate(s.Spec.AsDefault, s.Spec.Mode == table.Namespace); err != nil {
		return err
	}

	if err = dao.validateAttachmentAppExist(kit, strategy.Attachment); err != nil {
		return err
	}

	// validate strategy binding release exist.
	if err = dao.validateReleaseExist(kit, strategy); err != nil {
		return err
	}

	opts := orm.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(
		"id", "biz_id", "app_id", "strategy_set_id", "namespace")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(strategy, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	ab := dao.auditDao.Decorator(kit, strategy.Attachment.BizID, enumor.Strategy).PrepareUpdate(strategy)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.StrategyTable.Name(), " SET ", expr, " WHERE id = ",
		strconv.Itoa(int(strategy.ID)), " AND biz_id = ", strconv.Itoa(int(strategy.Attachment.BizID)))
	sql := filter.SqlJoint(sqlSentence)
	err = dao.sd.ShardingOne(strategy.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			effected, err := dao.orm.Txn(txn).Update(kit.Ctx, sql, toUpdate)
			if err != nil {
				logs.Errorf("update strategy: %d failed, err: %v, rid: %v", strategy.ID, err, kit.Rid)
				return err
			}

			if effected == 0 {
				logs.Errorf("update one strategy: %d, but record not found, rid: %v", strategy.ID, kit.Rid)
				return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
			}

			if effected > 1 {
				logs.Errorf("update one strategy: %d, but got updated strategy count: %d, rid: %v", strategy.ID,
					effected, kit.Rid)
				return fmt.Errorf("matched strategy count %d is not as excepted", effected)
			}

			// do audit
			if err := ab.Do(&AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}); err != nil {
				return fmt.Errorf("do strategy update audit failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

// List strategies with options.
func (dao *strategyDao) List(kit *kit.Kit, opts *types.ListStrategiesOption) (
	*types.ListStrategyDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list strategies options null")
	}

	if err := opts.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id", "biz_id", "app_id", "strategy_set_id"},
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
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sqlSentence = append(sqlSentence, "SELECT COUNT(*) FROM ", table.StrategyTable.Name(), whereExpr)
		sql := filter.SqlJoint(sqlSentence)
		var count uint32
		count, err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, sql, args...)
		if err != nil {
			return nil, err
		}

		return &types.ListStrategyDetails{Count: count, Details: make([]*table.Strategy, 0)}, nil
	}

	// query strategy list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sqlSentence = append(sqlSentence, "SELECT ", table.StrategyColumns.NamedExpr(), " FROM ", table.StrategyTable.Name(),
		whereExpr, pageExpr)
	sql := filter.SqlJoint(sqlSentence)
	list := make([]*table.Strategy, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListStrategyDetails{Count: 0, Details: list}, nil
}

// Delete one strategy instance.
func (dao *strategyDao) Delete(kit *kit.Kit, strategy *table.Strategy) error {

	if strategy == nil {
		return errf.New(errf.InvalidParameter, "strategy is nil")
	}

	if err := strategy.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	// validate that publishing strategy cannot be deleted
	if err := dao.validateStrategyNotPublishing(kit, strategy.ID, strategy.Attachment.BizID); err != nil {
		return err
	}

	if strategy.Attachment.AppID <= 0 {
		return errf.New(errf.InvalidParameter, "app id should be set")
	}

	if err := dao.validateAttachmentAppExist(kit, strategy.Attachment); err != nil {
		return err
	}

	strategy, err := dao.getStrategy(kit, strategy.Attachment.BizID, strategy.Attachment.AppID, strategy.ID)
	if err != nil {
		return err
	}

	ab := dao.auditDao.Decorator(kit, strategy.Attachment.BizID, enumor.Strategy).PrepareDelete(strategy.ID)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.StrategyTable.Name(), " WHERE id = ", strconv.Itoa(int(strategy.ID)),
		" AND biz_id = ", strconv.Itoa(int(strategy.Attachment.BizID)))
	expr := filter.SqlJoint(sqlSentence)
	eDecorator := dao.event.Eventf(kit)
	err = dao.sd.ShardingOne(strategy.Attachment.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		// delete the strategy at first.
		err := dao.orm.Txn(txn).Delete(kit.Ctx, expr)
		if err != nil {
			return err
		}

		// audit this delete app details.
		auditOpt := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err := ab.Do(auditOpt); err != nil {
			return fmt.Errorf("audit delete strategy failed, err: %v", err)
		}

		// delete current published strategy.
		var currentSqlSentence []string
		currentSqlSentence = append(currentSqlSentence, "DELETE FROM ", table.CurrentPublishedStrategyTable.Name(),
			" WHERE strategy_id = ", strconv.Itoa(int(strategy.ID)), " AND app_id = ", strconv.Itoa(int(strategy.Attachment.AppID)))
		sql := filter.SqlJoint(currentSqlSentence)
		if err = dao.orm.Txn(txn).Delete(kit.Ctx, sql); err != nil {
			return fmt.Errorf("delete current published strategy failed, err: %v", err)
		}

		// decrease the strategy related locks count after the deletion
		if err := dao.tryDeleteLocks(kit, strategy, &LockOption{Txn: txn}); err != nil {
			return err
		}

		if strategy.State.PubState == table.Publishing || strategy.State.PubState == table.Published {
			// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
			one := types.Event{
				Spec: &table.EventSpec{
					Resource:   table.Publish,
					ResourceID: strategy.ID,
					OpType:     table.DeleteOp,
				},
				Attachment: &table.EventAttachment{BizID: strategy.Attachment.BizID, AppID: strategy.Attachment.AppID},
				Revision:   &table.CreatedRevision{Creator: kit.User, CreatedAt: time.Now()},
			}
			if err := eDecorator.Fire(one); err != nil {
				logs.Errorf("fire delete %d strategy publish event failed, err: %v, rid: %s", strategy.ID, err, kit.Rid)
				return errf.New(errf.DBOpFailed, "fire event failed, "+err.Error())
			}
		}
		return nil
	})
	if strategy.State.PubState == table.Publishing || strategy.State.PubState == table.Published {
		eDecorator.Finalizer(err)
	}

	if err != nil {
		logs.Errorf("delete strategy: %d failed, err: %v, rid: %v", strategy.ID, err, kit.Rid)
		return fmt.Errorf("delete strategy, but run txn failed, err: %v", err)
	}

	return nil
}

// validateAttachmentResExist validate if attachment resource exists before creating strategy.
func (dao *strategyDao) validateAttachmentResExist(kit *kit.Kit, am *table.StrategyAttachment) error {

	if err := dao.validateAttachmentAppExist(kit, am); err != nil {
		return err
	}

	if err := dao.validateAttachmentStrategySetExist(kit, am); err != nil {
		return err
	}

	return nil
}

// validateReleaseExist validate if strategy's release exists before creating or updating.
func (dao *strategyDao) validateReleaseExist(kt *kit.Kit, strategy *table.Strategy) error {
	if strategy == nil {
		return errf.New(errf.InvalidParameter, "strategy is required")
	}

	var sqlSentence []string

	// validate main strategy binding release exist.
	if strategy.Spec != nil && strategy.Spec.ReleaseID != 0 {
		sqlSentence = append(sqlSentence, " WHERE id = ", strconv.Itoa(int(strategy.Spec.ReleaseID)),
			" AND biz_id = ", strconv.Itoa(int(strategy.Attachment.BizID)))
		sql := filter.SqlJoint(sqlSentence)
		exist, err := isResExist(kt, dao.orm, dao.sd.ShardingOne(strategy.Attachment.BizID), table.ReleaseTable, sql)
		if err != nil {
			return err
		}

		if !exist {
			return errf.New(errf.RecordNotFound, fmt.Sprintf("strategy binding release %d is not exist",
				strategy.Spec.ReleaseID))
		}
	}

	return nil
}

// validateAttachmentAppExist validate if attachment app exists before creating strategy.
func (dao *strategyDao) validateAttachmentAppExist(kit *kit.Kit, am *table.StrategyAttachment) error {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, " WHERE id = ", strconv.Itoa(int(am.AppID)),
		" AND biz_id = ", strconv.Itoa(int(am.BizID)))
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable, sql)
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RecordNotFound, "requested app not exist")
	}

	return nil
}

// validateAttachmentStrategySetExist validate if attachment strategy set exists before creating strategy.
func (dao *strategyDao) validateAttachmentStrategySetExist(kit *kit.Kit, am *table.StrategyAttachment) error {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, " WHERE id = ", strconv.Itoa(int(am.StrategySetID)),
		" AND biz_id = ", strconv.Itoa(int(am.BizID)), " AND app_id = ", strconv.Itoa(int(am.AppID)))
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.StrategySetTable, sql)
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("strategy attached strategy set %d is not exist",
			am.StrategySetID))
	}

	return nil
}

func (dao *strategyDao) getStrategy(kit *kit.Kit, bizID, appID, strategyID uint32) (*table.Strategy, error) {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.StrategyColumns.NamedExpr(), " FROM ", table.StrategyTable.Name(),
		" WHERE id = ", strconv.Itoa(int(strategyID)), " AND biz_id = ", strconv.Itoa(int(bizID)), " AND app_id = ", strconv.Itoa(int(appID)))
	sql := filter.SqlJoint(sqlSentence)
	one := new(table.Strategy)
	err := dao.orm.Do(dao.sd.MustSharding(bizID)).Get(kit.Ctx, one, sql)
	if err != nil {
		return nil, errf.New(errf.DBOpFailed, fmt.Sprintf("get strategy details failed, err: %v", err))
	}

	return one, nil
}

// validateStrategyNotPublishing validate if strategy is not publishing, returns error if it is publishing
func (dao *strategyDao) validateStrategyNotPublishing(kit *kit.Kit, strategyID, bizID uint32) error {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, " WHERE id = ", strconv.Itoa(int(strategyID)), " AND biz_id = ", strconv.Itoa(int(bizID)),
		" AND pub_state = '", string(table.Publishing), "'")
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(bizID), table.StrategyTable, sql)
	if err != nil {
		return err
	}

	if exist {
		return errf.New(errf.InvalidParameter, "strategy should not be publishing")
	}

	return nil
}

// validateAppStrategyNumber verify whether the current number of app strategies have reached the maximum.
func (dao *strategyDao) validateAppStrategyNumber(kt *kit.Kit, at *table.StrategyAttachment,
	mode table.AppMode, lo *LockOption) error {

	// try lock strategy to ensure the number is limited when creating concurrently
	lock := lockKey.Strategy(at.BizID, at.AppID)
	count, err := dao.lock.IncreaseCount(kt, lock, lo)
	if err != nil {
		return err
	}

	if err := table.ValidateAppStrategyNumber(count, mode); err != nil {
		return err
	}

	return nil
}

// validateDefaultStrategyNotExist validate default strategy not exist in strategy set.
func (dao *strategyDao) validateDefaultStrategyNotExist(kt *kit.Kit, at *table.StrategyAttachment,
	lo *LockOption) error {

	// try lock default strategy to ensure the number is limited when creating concurrently
	lock := lockKey.DefaultStrategy(at.BizID, at.StrategySetID)
	isUnique, err := dao.lock.AddUnique(kt, lock, lo)
	if err != nil {
		return err
	}

	if !isUnique {
		return errf.New(errf.InvalidParameter, "a default strategy already exists in the current strategy set")
	}

	return nil
}

// validateNamespaceExist validate namespace only in strategy set.
func (dao *strategyDao) validateNamespaceExist(kt *kit.Kit, at *table.StrategyAttachment, ns string,
	lo *LockOption) error {

	// try lock namespace strategy to ensure the number is limited when creating concurrently
	lock := lockKey.NamespaceStrategy(at.BizID, at.StrategySetID, ns)
	isUnique, err := dao.lock.AddUnique(kt, lock, lo)
	if err != nil {
		return err
	}

	if !isUnique {
		return errf.New(errf.InvalidParameter, "namespace repeats under the current strategy set")
	}

	return nil
}

// tryDeleteLocks decrease the strategy lock count after the deletion
func (dao *strategyDao) tryDeleteLocks(kt *kit.Kit, strategy *table.Strategy, lo *LockOption) error {
	lock := lockKey.Strategy(strategy.Attachment.BizID, strategy.Attachment.AppID)
	if err := dao.lock.DecreaseCount(kt, lock, lo); err != nil {
		return err
	}

	if strategy.Spec.AsDefault {
		defLock := lockKey.DefaultStrategy(strategy.Attachment.BizID, strategy.Attachment.StrategySetID)
		if err := dao.lock.DeleteUnique(kt, defLock, lo); err != nil {
			return err
		}
	}

	if !strategy.Spec.AsDefault && strategy.Spec.Mode == table.Namespace {
		nsLock := lockKey.NamespaceStrategy(strategy.Attachment.BizID, strategy.Attachment.StrategySetID,
			strategy.Spec.Namespace)
		if err := dao.lock.DeleteUnique(kt, nsLock, lo); err != nil {
			return err
		}
	}

	return nil
}
