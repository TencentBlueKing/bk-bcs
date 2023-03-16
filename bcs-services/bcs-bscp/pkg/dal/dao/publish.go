/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"

	"github.com/jmoiron/sqlx"
)

// Publish defines all the publish operation related operations.
type Publish interface {
	// Publish publish an app's release with its strategy.
	// once an app's strategy along with its release id is published,
	// all its released config items are effected immediately.
	Publish(kit *kit.Kit, opt *types.PublishOption) (id uint32, err error)

	// FinishPublish finish the strategy's publish process when a
	// strategy is in publishing state.
	FinishPublish(kit *kit.Kit, opt *types.FinishPublishOption) error

	// GetAppCPStrategies get an app's current published strategies for cache.
	GetAppCPStrategies(kt *kit.Kit, opts *types.GetAppCPSOption) ([]*types.PublishedStrategyCache, error)

	// GetAppCpsID get an app's current published strategy id.
	GetAppCpsID(kt *kit.Kit, opts *types.GetAppCpsIDOption) ([]uint32, error)

	// ListPSHistory list published strategy history with options.
	ListPSHistory(kit *kit.Kit, opts *types.ListPSHistoriesOption) (*types.ListPSHistoryDetails, error)
}

var _ Publish = new(pubDao)

type pubDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	event    Event
}

// Publish publish an app's release with its strategy.
// once an app's strategy along with its release id is published,
// all its released config items are effected immediately.
// return the published strategy history record id.
func (pd *pubDao) Publish(kit *kit.Kit, opt *types.PublishOption) (uint32, error) {

	if opt == nil {
		return 0, errf.New(errf.InvalidParameter, "publish strategy option is nil")
	}

	if err := opt.Validate(); err != nil {
		return 0, err
	}

	eDecorator := pd.event.Eventf(kit)
	var pshID uint32
	err := pd.sd.ShardingOne(opt.BizID).AutoTxn(kit, func(txn *sqlx.Tx, options *sharding.TxnOption) error {
		groups := make([]*table.Group, len(opt.Groups))
		if !opt.All {
			// list groups if gray release
			var sqlSentence []string
			sqlSentence = append(sqlSentence, "SELECT ", table.GroupColumns.NamedExpr(), " FROM ", table.GroupTable.Name(),
				" WHERE id IN (", tools.JoinUint32(opt.Groups, ","), ")")
			lgExpr := filter.SqlJoint(sqlSentence)
			if err := pd.orm.Do(pd.sd.MustSharding(opt.BizID)).Select(kit.Ctx, &groups, lgExpr); err != nil {
				logs.Errorf("get to be published groups(%s) failed, err: %v, rid: %s",
					tools.JoinUint32(opt.Groups, ","), err, kit.Rid)
				return errf.New(errf.DBOpFailed, err.Error())
			}
			// if any group can not match,return err
			if len(groups) != len(opt.Groups) {
				return errf.New(errf.DBOpFailed,
					fmt.Sprintf("groups num not matched with id(%s)", tools.JoinUint32(opt.Groups, ",")))
			}
		}
		// create strategy to publish it later
		now := time.Now()
		stgID, err := pd.idGen.One(kit, table.StrategyTable)
		stg := &table.Strategy{
			ID: stgID,
			Spec: &table.StrategySpec{
				// TODO: strategy name
				Name:      "TODO",
				ReleaseID: opt.ReleaseID,
				AsDefault: opt.All,
				Scope: &table.Scope{
					Groups: groups,
				},
				Mode: table.Normal,
				Memo: opt.Memo,
			},
			State: &table.StrategyState{
				PubState: table.Publishing,
			},
			Attachment: &table.StrategyAttachment{
				BizID: opt.BizID,
				AppID: opt.AppID,
			},
			Revision: &table.Revision{
				Creator:   kit.User,
				Reviser:   kit.User,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}
		var sqlSentence []string
		sqlSentence = append(sqlSentence, "INSERT INTO ", table.StrategyTable.Name(), " (", table.StrategyColumns.ColumnExpr(),
			")  VALUES(", table.StrategyColumns.ColonNameExpr(), ")")
		stgExpr := filter.SqlJoint(sqlSentence)

		if err = pd.orm.Txn(txn).Insert(kit.Ctx, stgExpr, stg); err != nil {
			return err
		}

		// audit this to create strategy details.
		auc := &AuditOption{Txn: txn, ResShardingUid: options.ShardingUid}
		if err = pd.auditDao.Decorator(kit, stg.Attachment.BizID,
			enumor.Strategy).AuditCreate(stg, auc); err != nil {
			return fmt.Errorf("audit create strategy failed, err: %v", err)
		}

		// audit this to publish strategy details.
		aup := &AuditOption{Txn: txn, ResShardingUid: options.ShardingUid}
		if err = pd.auditDao.Decorator(kit, stg.Attachment.BizID,
			enumor.Strategy).AuditPublish(stg, aup); err != nil {
			return fmt.Errorf("audit publish strategy failed, err: %v", err)
		}

		// upsert the published strategy to the CurrentPublishedStrategy table for record.
		if err := pd.upsertToCurrentPublishedStrategy(kit, txn, stg); err != nil {
			logs.Errorf("upsert to current published strategy table failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}

		// save history to the PublishedStrategyHistoryTable table for record.
		id, err := pd.recordPublishedStrategyHistory(kit, txn, stg, opt)
		if err != nil {
			logs.Errorf("record the published strategy history table failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}

		// add release publish num
		if err := pd.increaseReleasePublishNum(kit, txn, stg.Spec.ReleaseID); err != nil {
			logs.Errorf("increate release publish num failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}

		pshID = id

		// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
		one := types.Event{
			Spec: &table.EventSpec{
				Resource: table.Publish,
				// use the published strategy history id, which represent a real publish operation.
				ResourceID: opt.ReleaseID,
				OpType:     table.InsertOp,
			},
			Attachment: &table.EventAttachment{BizID: opt.BizID, AppID: opt.AppID},
			Revision:   &table.CreatedRevision{Creator: kit.User, CreatedAt: time.Now()},
		}
		if err := eDecorator.Fire(one); err != nil {
			logs.Errorf("fire publish strategy event failed, err: %v, rid: %s", err, kit.Rid)
			return errf.New(errf.DBOpFailed, "fire event failed, "+err.Error())
		}

		return nil
	})

	eDecorator.Finalizer(err)

	if err != nil {
		logs.Errorf("publish strategy failed, err: %v, rid: %s", err, kit.Rid)
		return 0, err
	}

	return pshID, nil
}

// upsertToCurrentPublishedStrategy upsert the published strategy to the CurrentPublishedStrategy table for record.
// the current published strategy table only record the all current published strategy, but not the published
// strategy history, so if the strategy is already in this table(which means this strategy has been published
// before), then update it, otherwise, insert it.
func (pd *pubDao) upsertToCurrentPublishedStrategy(kt *kit.Kit, txn *sqlx.Tx, s *table.Strategy) error {
	// generate the strategy id and update to strategy.
	id, err := pd.idGen.One(kt, table.CurrentPublishedStrategyTable)
	if err != nil {
		return err
	}

	cps := &table.CurrentPublishedStrategy{
		ID:         id,
		StrategyID: s.ID,
		Spec:       s.Spec,
		State: &table.StrategyState{
			PubState: "",
		},
		Attachment: s.Attachment,
		Revision: &table.CreatedRevision{
			Creator:   kt.User,
			CreatedAt: time.Now(),
		},
	}
	opts := orm.NewFieldOptions().AddIgnoredFields("strategy_id", "pub_state")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(cps, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.CurrentPublishedStrategyTable.Name(), " SET ", expr, " WHERE strategy_id = ", strconv.Itoa(int(cps.StrategyID)))
	sql := filter.SqlJoint(sqlSentence)
	result, err := txn.NamedExecContext(kt.Ctx, sql, toUpdate)
	if err != nil {
		return errf.New(errf.DBOpFailed, "update current published strategy failed, err: "+err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errf.New(errf.DBOpFailed, err.Error())
	}

	if rowsAffected > 1 {
		return fmt.Errorf("update current published strategy affected is %d", rowsAffected)
	}

	if rowsAffected == 1 {
		return nil
	}

	var sqlSentenceIn []string
	sqlSentenceIn = append(sqlSentenceIn, "INSERT INTO ", table.CurrentPublishedStrategyTable.Name(), " (", table.CurrentPublishedStrategyColumns.ColumnExpr(),
		") VALUES(", table.CurrentPublishedStrategyColumns.ColonNameExpr(), ")")
	sql = filter.SqlJoint(sqlSentenceIn)

	if _, err := txn.NamedExecContext(kt.Ctx, sql, cps); err != nil {
		// concurrency can cause deadlock problems and provide three retries
		if strings.Contains(err.Error(), orm.ErrDeadLock) {
			return sharding.ErrRetryTransaction
		}
		return errf.New(errf.DBOpFailed, "insert current published strategy failed, err: "+err.Error())
	}

	return nil
}

// recordPublishedStrategyHistory record the to be published strategy to its history table.
func (pd *pubDao) recordPublishedStrategyHistory(kit *kit.Kit, txn *sqlx.Tx, pubStrategy *table.Strategy,
	opt *types.PublishOption) (uint32, error) {

	id, err := pd.idGen.One(kit, table.PublishedStrategyHistoryTable)
	if err != nil {
		return 0, errf.New(errf.DBOpFailed, "generate published strategy history id failed, err: "+err.Error())
	}

	published := &table.PublishedStrategyHistory{
		ID:         id,
		StrategyID: pubStrategy.ID,
		Spec:       pubStrategy.Spec,
		State:      pubStrategy.State,
		Attachment: pubStrategy.Attachment,
		Revision:   opt.Revision,
	}
	published.State.PubState = ""

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.PublishedStrategyHistoryTable.Name(), " (", table.PubStrategyHistoryColumns.ColumnExpr(),
		")  VALUES(", table.PubStrategyHistoryColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)

	if _, err := txn.NamedExecContext(kit.Ctx, sql, published); err != nil {
		logs.Errorf("insert published strategy history failed, sql: %s, err: %v, rid: %s", sql, err, kit.Rid)
		return 0, errf.New(errf.DBOpFailed, "insert published strategy history failed, err: "+err.Error())
	}

	return id, nil
}

// increaseReleasePublishNum increase release publish num by 1
func (pd *pubDao) increaseReleasePublishNum(kit *kit.Kit, txn *sqlx.Tx, releaseID uint32) error {
	sql := fmt.Sprintf(`UPDATE %s SET publish_num = publish_num + 1 WHERE id = %d`, table.ReleaseTable, releaseID)
	if _, err := txn.ExecContext(kit.Ctx, sql); err != nil {
		logs.Errorf("increate release publish num failed, sql: %s, err: %v, rid: %s", sql, err, kit.Rid)
		return errf.New(errf.DBOpFailed, "insert published strategy history failed, err: "+err.Error())
	}
	return nil
}

// updateStrategyPublishState update a strategy's publish state.
// it uses the txn to update state if it is not nil, otherwise update the state directly.
func (pd *pubDao) updateStrategyPublishState(kit *kit.Kit, txn *sqlx.Tx, bizID uint32, strategyID uint32,
	state table.PublishState) error {

	if bizID <= 0 {
		return errf.New(errf.InvalidParameter, "biz_id is invalid")
	}

	if err := state.Validate(); err != nil {
		return errf.New(errf.InvalidParameter, "invalid strategy publish state: "+string(state))
	}
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.StrategyTable.Name(), " SET pub_state = '", string(state), "', reviser = '",
		kit.User, "', updated_at = '", time.Now().Format(constant.TimeStdFormat), "' ", "WHERE id = ", strconv.Itoa(int(strategyID)),
		" AND biz_id = ", strconv.Itoa(int(bizID)), " AND pub_state != '", string(state), "'")
	expr := filter.SqlJoint(sqlSentence)

	if txn == nil {
		// update state without an already existed transaction
		if _, err := pd.orm.Do(pd.sd.MustSharding(bizID)).Exec(kit.Ctx, expr); err != nil {
			return errf.New(errf.DBOpFailed, "update strategy state failed, err: "+err.Error())
		}

		return nil
	}

	// update state with an already existed transaction
	result, err := txn.ExecContext(kit.Ctx, expr)
	if err != nil {
		return errf.New(errf.DBOpFailed, "update strategy state failed, err: "+err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errf.New(errf.DBOpFailed, err.Error())
	}

	// if state is publishing and row affected is 0, it means that strategy
	// origin state is publishing, so we need to remind users to finish publishing.
	if state == table.Publishing && rowsAffected == 0 {
		return errf.New(errf.Aborted, "need to finish the last publish before the next publish")
	}

	return nil
}

// FinishPublish is to end publish strategy process.
// this operation only set the strategy's state to published.
func (pd *pubDao) FinishPublish(kit *kit.Kit, opt *types.FinishPublishOption) error {
	if opt == nil {
		return errf.New(errf.InvalidParameter, "finish publish option is nil")
	}

	if err := opt.Validate(); err != nil {
		return err
	}

	err := pd.sd.ShardingOne(opt.BizID).AutoTxn(kit, func(txn *sqlx.Tx, options *sharding.TxnOption) error {
		if err := pd.updateStrategyPublishState(kit, txn, opt.BizID, opt.StrategyID, table.Published); err != nil {
			return err
		}

		au := &AuditOption{Txn: txn, ResShardingUid: options.ShardingUid}
		if err := pd.auditDao.Decorator(kit, opt.BizID, enumor.Strategy).AuditFinishPublish(
			opt.StrategyID, opt.AppID, au); err != nil {
			return fmt.Errorf("audit finish publish strategy failed, err: %v", err)
		}
		return nil
	})
	if err != nil {
		logs.Errorf("finish publish strategy failed, err: %v, rid: %s", err, kit.Rid)
		return err
	}

	return nil
}

// ListPSHistory list published strategy history with options.
func (pd *pubDao) ListPSHistory(kit *kit.Kit, opts *types.ListPSHistoriesOption) (
	*types.ListPSHistoryDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list published strategy histories options is nil")
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

	var sql string
	var sqlSentence []string
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sqlSentence = append(sqlSentence, "SELECT COUNT(*) FROM ", table.PublishedStrategyHistoryTable.Name(), whereExpr)
		sql = filter.SqlJoint(sqlSentence)
		var count uint32
		count, err = pd.orm.Do(pd.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, sql, args...)
		if err != nil {
			return nil, err
		}

		return &types.ListPSHistoryDetails{Count: count, Details: make([]*table.PublishedStrategyHistory, 0)}, nil
	}

	// query published strategy history list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sqlSentence = append(sqlSentence, "SELECT ", table.PubStrategyHistoryColumns.NamedExpr(), " FROM ", table.PublishedStrategyHistoryTable.Name(),
		whereExpr, pageExpr)
	sql = filter.SqlJoint(sqlSentence)

	list := make([]*table.PublishedStrategyHistory, 0)
	err = pd.orm.Do(pd.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListPSHistoryDetails{Count: 0, Details: list}, nil
}

// GetAppCPStrategies get an app's current published strategies for cache.
func (pd *pubDao) GetAppCPStrategies(kt *kit.Kit, opts *types.GetAppCPSOption) ([]*types.PublishedStrategyCache,
	error) {

	po := &types.PageOption{MaxLimit: types.GetCPSMaxPageLimit}
	if err := opts.Validate(po); err != nil {
		return nil, err
	}

	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", types.PublishedStrategyCacheColumn, " FROM ", table.CurrentPublishedStrategyTable.Name(),
		" WHERE biz_id = ", strconv.Itoa(int(opts.BizID)), " AND app_id = ", strconv.Itoa(int(opts.AppID)), pageExpr)
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*types.PublishedStrategyCache, 0)
	if err = pd.orm.Do(pd.sd.ShardingOne(opts.BizID).DB()).Select(kt.Ctx, &list, sql); err != nil {
		return nil, err
	}

	return list, nil
}

// GetAppCpsID get an app's current published strategy ids.
func (pd *pubDao) GetAppCpsID(kt *kit.Kit, opts *types.GetAppCpsIDOption) ([]uint32, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	var sql string
	var sqlSentence []string
	if len(opts.Namespace) != 0 {
		// query namespace and default cps.
		sqlSentence = append(sqlSentence, "SELECT id FROM ", table.CurrentPublishedStrategyTable.Name(), " WHERE app_id = ", strconv.Itoa(int(opts.AppID)), " AND (namespace = '", opts.Namespace, "' OR as_default = true)")
		sql = filter.SqlJoint(sqlSentence)
	} else {
		// query app all cps.
		sqlSentence = append(sqlSentence, "SELECT id FROM ", table.CurrentPublishedStrategyTable.Name(), " WHERE biz_id = ", strconv.Itoa(int(opts.BizID)), " AND app_id = ", strconv.Itoa(int(opts.AppID)))
		sql = filter.SqlJoint(sqlSentence)
	}

	list := make([]uint32, 0)
	if err := pd.orm.Do(pd.sd.ShardingOne(opts.BizID).DB()).Select(kt.Ctx, &list, sql); err != nil {
		return nil, err
	}

	return list, nil
}
