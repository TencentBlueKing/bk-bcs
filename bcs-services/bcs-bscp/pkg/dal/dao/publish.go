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
	"bscp.io/pkg/types"

	"github.com/jmoiron/sqlx"
)

// Publish defines all the publish operation related operations.
type Publish interface {
	// PublishStrategy publish an app's release with its strategy.
	// once an app's strategy along with its release id is published,
	// all its released config items are effected immediately.
	PublishStrategy(kit *kit.Kit, opt *types.PublishStrategyOption) (id uint32, err error)

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

// PublishStrategy publish an app's release with its strategy.
// once an app's strategy along with its release id is published,
// all its released config items are effected immediately.
// return the published strategy history record id.
func (pd *pubDao) PublishStrategy(kit *kit.Kit, opt *types.PublishStrategyOption) (uint32, error) {

	if opt == nil {
		return 0, errf.New(errf.InvalidParameter, "publish strategy option is nil")
	}

	if err := opt.Validate(); err != nil {
		return 0, err
	}

	// get the strategy details to publish it later.
	pubStrategy := new(table.Strategy)
	stgExpr := fmt.Sprintf(`SELECT %s FROM %s WHERE id = %d AND biz_id = %d AND app_id = %d`,
		table.StrategyColumns.NamedExpr(), table.StrategyTable, opt.StrategyID, opt.BizID, opt.AppID)

	if err := pd.orm.Do(pd.sd.MustSharding(opt.BizID)).Get(kit.Ctx, pubStrategy, stgExpr); err != nil {
		// if it can not find this to be published strategy, return err with error code.
		if err == orm.ErrRecordNotFound {
			return 0, errf.New(errf.RecordNotFound, fmt.Sprintf("strategy with id(%d) not found", opt.StrategyID))
		}

		logs.Errorf("get to be published strategy(%d) failed, err: %v, rid: %s", opt.StrategyID, err, kit.Rid)
		return 0, errf.New(errf.DBOpFailed, err.Error())
	}

	eDecorator := pd.event.Eventf(kit)
	var pshID uint32
	err := pd.sd.ShardingOne(opt.BizID).AutoTxn(kit, func(txn *sqlx.Tx, options *sharding.TxnOption) error {
		// should first update the strategy state to publishing to ensure that the current state is not publishing.
		if err := pd.updateStrategyPublishState(kit, txn, opt.BizID, opt.StrategyID, table.Publishing); err != nil {
			logs.Errorf("update the strategy(%d) state to publishing failed, err: %v, rid: %s", opt.StrategyID,
				err, kit.Rid)
			return err
		}

		// upsert the published strategy to the CurrentPublishedStrategy table for record.
		if err := pd.upsertToCurrentPublishedStrategy(kit, txn, pubStrategy); err != nil {
			logs.Errorf("upsert to current published strategy table failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}

		// save history to the PublishedStrategyHistoryTable table for record.
		id, err := pd.recordPublishedStrategyHistory(kit, txn, pubStrategy, opt)
		if err != nil {
			logs.Errorf("record the published strategy history table failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}
		pshID = id

		au := &AuditOption{Txn: txn, ResShardingUid: options.ShardingUid}
		if err := pd.auditDao.Decorator(kit, opt.BizID, enumor.Strategy).AuditPublish(pubStrategy, au); err != nil {
			return fmt.Errorf("audit publish strategy failed, err: %v", err)
		}

		// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
		one := types.Event{
			Spec: &table.EventSpec{
				Resource: table.PublishStrategy,
				// use the published strategy history id, which represent a real publish operation.
				ResourceID: opt.StrategyID,
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

	sql := fmt.Sprintf(`UPDATE %s SET %s WHERE strategy_id = %d`, table.CurrentPublishedStrategyTable,
		expr, cps.StrategyID)
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

	sql = fmt.Sprintf("INSERT INTO %s (%s) VALUES(%s)",
		table.CurrentPublishedStrategyTable, table.CurrentPublishedStrategyColumns.ColumnExpr(),
		table.CurrentPublishedStrategyColumns.ColonNameExpr())

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
	opt *types.PublishStrategyOption) (uint32, error) {

	id, err := pd.idGen.One(kit, table.PublishedStrategyHistoryTable)
	if err != nil {
		return 0, errf.New(errf.DBOpFailed, "generate published strategy history id failed, err: "+err.Error())
	}

	published := &table.PublishedStrategyHistory{
		ID:         id,
		StrategyID: opt.StrategyID,
		Spec:       pubStrategy.Spec,
		State:      pubStrategy.State,
		Attachment: pubStrategy.Attachment,
		Revision:   opt.Revision,
	}
	published.State.PubState = ""

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.PublishedStrategyHistoryTable,
		table.PubStrategyHistoryColumns.ColumnExpr(), table.PubStrategyHistoryColumns.ColonNameExpr())

	if _, err := txn.NamedExecContext(kit.Ctx, sql, published); err != nil {
		logs.Errorf("insert published strategy history failed, sql: %s, err: %v, rid: %s", sql, err, kit.Rid)
		return 0, errf.New(errf.DBOpFailed, "insert published strategy history failed, err: "+err.Error())
	}

	return id, nil
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

	expr := fmt.Sprintf("UPDATE %s SET pub_state = '%s', reviser = '%s', updated_at = '%s' "+
		"WHERE id = %d AND biz_id = %d AND pub_state != '%s'", table.StrategyTable, state, kit.User,
		time.Now().Format(constant.TimeStdFormat), strategyID, bizID, state)

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
	whereExpr, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sql string
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sql = fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.PublishedStrategyHistoryTable, whereExpr)
		var count uint32
		count, err = pd.orm.Do(pd.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, sql)
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

	sql = fmt.Sprintf(`SELECT %s FROM %s %s %s`,
		table.PubStrategyHistoryColumns.NamedExpr(), table.PublishedStrategyHistoryTable, whereExpr, pageExpr)

	list := make([]*table.PublishedStrategyHistory, 0)
	err = pd.orm.Do(pd.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql)
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

	sql := fmt.Sprintf("SELECT %s FROM %s WHERE biz_id = %d AND app_id = %d %s", types.PublishedStrategyCacheColumn,
		table.CurrentPublishedStrategyTable, opts.BizID, opts.AppID, pageExpr)

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
	if len(opts.Namespace) != 0 {
		// query namespace and default cps.
		sql = fmt.Sprintf("SELECT id FROM %s WHERE app_id = %d AND (namespace = '%s' OR as_default = true)",
			table.CurrentPublishedStrategyTable, opts.AppID, opts.Namespace)
	} else {
		// query app all cps.
		sql = fmt.Sprintf("SELECT id FROM %s WHERE biz_id = %d AND app_id = %d", table.CurrentPublishedStrategyTable,
			opts.BizID, opts.AppID)
	}

	list := make([]uint32, 0)
	if err := pd.orm.Do(pd.sd.ShardingOne(opts.BizID).DB()).Select(kt.Ctx, &list, sql); err != nil {
		return nil, err
	}

	return list, nil
}
