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

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/runtime/selector"
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

	PublishWithTx(kit *kit.Kit, tx *sharding.Tx, opt *types.PublishOption) (id uint32, err error)
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
		groups := make([]*table.Group, 0, len(opt.Groups))
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
		}
		// create strategy to publish it later
		now := time.Now()
		stgID, err := pd.idGen.One(kit, table.StrategyTable)
		if err != nil {
			logs.Errorf("generate strategy id failed, err: %v, rid: %s", err, kit.Rid)
			return errf.New(errf.DBOpFailed, err.Error())
		}
		pshID = stgID

		stg := &table.Strategy{
			ID: stgID,
			Spec: &table.StrategySpec{
				Name:      now.Format(time.RFC3339),
				ReleaseID: opt.ReleaseID,
				AsDefault: opt.Default,
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

		// add release publish num
		if err := pd.increaseReleasePublishNum(kit, txn, stg.Spec.ReleaseID); err != nil {
			logs.Errorf("increate release publish num failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}

		if err := pd.upsertReleasedGroups(kit, txn, opt, stg); err != nil {
			logs.Errorf("upsert group current releases failed, err: %v, rid: %s", err, kit.Rid)
			return err
		}

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

// Publish publish with transaction
func (pd *pubDao) PublishWithTx(kit *kit.Kit, tx *sharding.Tx, opt *types.PublishOption) (uint32, error) {

	if opt == nil {
		return 0, errf.New(errf.InvalidParameter, "publish strategy option is nil")
	}

	if err := opt.Validate(); err != nil {
		return 0, err
	}

	groupIDs := opt.Groups
	if opt.All {
		// list groups if gray release
		var lgSql []string
		lgSql = append(lgSql, "SELECT group_id FROM ", table.ReleasedGroupTable.Name(),
			" WHERE group_id <> 0 AND app_id = ", strconv.Itoa(int(opt.AppID)),
			" AND biz_id = ", strconv.Itoa(int(opt.BizID)))
		lgExpr := filter.SqlJoint(lgSql)
		if err := pd.orm.Do(pd.sd.MustSharding(opt.BizID)).Select(kit.Ctx, &groupIDs, lgExpr); err != nil {
			logs.Errorf("get to be published groups(all) failed, err: %v, rid: %s", err, kit.Rid)
			return 0, errf.New(errf.DBOpFailed, err.Error())
		}
		opt.Default = true
	}

	eDecorator := pd.event.Eventf(kit)

	groups := make([]*table.Group, 0, len(groupIDs))
	// list groups if gray release
	if len(groupIDs) > 0 {
		var lgSentence []string
		lgSentence = append(lgSentence, "SELECT ", table.GroupColumns.NamedExpr(), " FROM ", table.GroupTable.Name(),
			" WHERE id IN (", tools.JoinUint32(groupIDs, ","), ")")
		lgExpr := filter.SqlJoint(lgSentence)
		if err := pd.orm.Do(pd.sd.MustSharding(opt.BizID)).Select(kit.Ctx, &groups, lgExpr); err != nil {
			logs.Errorf("get to be published groups(%s) failed, err: %v, rid: %s",
				tools.JoinUint32(groupIDs, ","), err, kit.Rid)
			return 0, errf.New(errf.DBOpFailed, err.Error())
		}
	}
	// create strategy to publish it later
	now := time.Now()
	stgID, err := pd.idGen.One(kit, table.StrategyTable)
	if err != nil {
		logs.Errorf("generate strategy id failed, err: %v, rid: %s", err, kit.Rid)
		return 0, errf.New(errf.DBOpFailed, err.Error())
	}
	pshID := stgID
	stg := &table.Strategy{
		ID: stgID,
		Spec: &table.StrategySpec{
			Name:      now.Format(time.RFC3339),
			ReleaseID: opt.ReleaseID,
			AsDefault: opt.Default,
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

	if err = pd.orm.Txn(tx.Tx()).Insert(kit.Ctx, stgExpr, stg); err != nil {
		return 0, err
	}

	// audit this to create strategy details.
	auc := &AuditOption{Txn: tx.Tx(), ResShardingUid: tx.ShardingUid()}
	if err = pd.auditDao.Decorator(kit, stg.Attachment.BizID,
		enumor.Strategy).AuditCreate(stg, auc); err != nil {
		return 0, fmt.Errorf("audit create strategy failed, err: %v", err)
	}

	// audit this to publish strategy details.
	aup := &AuditOption{Txn: tx.Tx(), ResShardingUid: tx.ShardingUid()}
	if err = pd.auditDao.Decorator(kit, stg.Attachment.BizID,
		enumor.Strategy).AuditPublish(stg, aup); err != nil {
		return 0, fmt.Errorf("audit publish strategy failed, err: %v", err)
	}

	// add release publish num
	if err := pd.increaseReleasePublishNum(kit, tx.Tx(), stg.Spec.ReleaseID); err != nil {
		logs.Errorf("increate release publish num failed, err: %v, rid: %s", err, kit.Rid)
		return 0, err
	}

	if err := pd.upsertReleasedGroups(kit, tx.Tx(), opt, stg); err != nil {
		logs.Errorf("upsert group current releases failed, err: %v, rid: %s", err, kit.Rid)
		return 0, err
	}

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
	if e := eDecorator.FireWithTx(tx, one); e != nil {
		logs.Errorf("fire publish strategy event failed, err: %v, rid: %s", e, kit.Rid)
		return 0, errf.New(errf.DBOpFailed, "fire event failed, "+e.Error())
	}

	return pshID, nil
}

// increaseReleasePublishNum increase release publish num by 1
func (pd *pubDao) increaseReleasePublishNum(kit *kit.Kit, txn *sqlx.Tx, releaseID uint32) error {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.ReleaseTable.Name(),
		" SET publish_num = publish_num + 1 WHERE id = ", strconv.Itoa(int(releaseID)))
	sql := filter.SqlJoint(sqlSentence)
	if _, err := txn.ExecContext(kit.Ctx, sql); err != nil {
		logs.Errorf("increate release publish num failed, sql: %s, err: %v, rid: %s", sql, err, kit.Rid)
		return errf.New(errf.DBOpFailed, "insert published strategy history failed, err: "+err.Error())
	}
	return nil
}

func (pd *pubDao) upsertReleasedGroups(kit *kit.Kit, txn *sqlx.Tx,
	opt *types.PublishOption, stg *table.Strategy) error {
	groups := stg.Spec.Scope.Groups
	now := time.Now()
	if opt.Default {
		groups = append(groups, &table.Group{
			ID: 0,
			Spec: &table.GroupSpec{
				Name:     "默认分组",
				Mode:     table.Default,
				Public:   true,
				Selector: new(selector.Selector),
				UID:      "",
			},
		})
	}
	for _, group := range groups {
		gcr := &table.ReleasedGroup{
			GroupID:    group.ID,
			AppID:      opt.AppID,
			ReleaseID:  opt.ReleaseID,
			StrategyID: stg.ID,
			Mode:       group.Spec.Mode,
			Selector:   group.Spec.Selector,
			UID:        group.Spec.UID,
			Edited:     false,
			BizID:      opt.BizID,
			Reviser:    kit.User,
			UpdatedAt:  now,
		}
		opts := orm.NewFieldOptions().AddIgnoredFields("id").AddBlankedFields("edited")
		expr, toUpdate, err := orm.RearrangeSQLDataWithOption(gcr, opts)
		var sqlSentence []string
		sqlSentence = append(sqlSentence, "UPDATE ", table.ReleasedGroupTable.Name(), " SET ", expr,
			" WHERE biz_id = ", strconv.Itoa(int(opt.BizID)), " AND group_id = ", strconv.Itoa(int(group.ID)),
			" AND app_id = ", strconv.Itoa(int(opt.AppID)))
		sql := filter.SqlJoint(sqlSentence)
		result, err := txn.NamedExecContext(kit.Ctx, sql, toUpdate)
		if err != nil {
			return errf.New(errf.DBOpFailed, "update group current releases failed, err: "+err.Error())
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return errf.New(errf.DBOpFailed, err.Error())
		}

		if rowsAffected > 1 {
			return fmt.Errorf("update group current releases affected is %d", rowsAffected)
		}

		if rowsAffected == 1 {
			continue
		}

		id, err := pd.idGen.One(kit, table.ReleasedGroupTable)
		if err != nil {
			return errf.New(errf.DBOpFailed, "generate group current releases id failed, err: "+err.Error())
		}
		gcr.ID = id

		var sqlSentenceIn []string
		sqlSentenceIn = append(sqlSentenceIn, "INSERT INTO ", table.ReleasedGroupTable.Name(), " (",
			table.ReleasedGroupColumns.ColumnExpr(), ") VALUES(", table.ReleasedGroupColumns.ColonNameExpr(), ")")
		sql = filter.SqlJoint(sqlSentenceIn)
		if _, err := txn.NamedExecContext(kit.Ctx, sql, gcr); err != nil {
			// concurrency can cause deadlock problems and provide three retries
			if strings.Contains(err.Error(), orm.ErrDeadLock) {
				return sharding.ErrRetryTransaction
			}
			return errf.New(errf.DBOpFailed, "insert group current releases failed, err: "+err.Error())
		}
	}
	return nil
}
