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

	"github.com/jmoiron/sqlx"

	"bscp.io/pkg/criteria/enumor"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// Hook supplies all the hook related operations.
type Hook interface {
	// Create one hook instance.
	Create(kit *kit.Kit, hook *table.Hook) (uint32, error)
	// Update one hook's info.
	Update(kit *kit.Kit, hook *table.Hook) error
	// List hooks with options.
	List(kit *kit.Kit, opts *types.ListHooksOption) (*types.ListHookDetails, error)
	// Delete one strategy instance.
	Delete(kit *kit.Kit, strategy *table.Hook) error
}

var _ Hook = new(hookDao)

type hookDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one hook instance.
func (dao *hookDao) Create(kit *kit.Kit, g *table.Hook) (uint32, error) {

	if g == nil {
		return 0, errf.New(errf.InvalidParameter, "hook is nil")
	}

	if err := g.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentResExist(kit, g.Attachment); err != nil {
		return 0, err
	}

	// generate a hook id and update to hook.
	id, err := dao.idGen.One(kit, table.HookTable)
	if err != nil {
		return 0, err
	}

	g.ID = id
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", string(table.HookTable), " (", table.HookColumns.ColumnExpr(), ")  VALUES(", table.HookColumns.ColonNameExpr(), ")")

	sql := filter.SqlJoint(sqlSentence)

	err = dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			if err := dao.orm.Txn(txn).Insert(kit.Ctx, sql, g); err != nil {
				return err
			}

			// audit this to be created hook details.
			au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
			if err = dao.auditDao.Decorator(kit, g.Attachment.BizID,
				enumor.Hook).AuditCreate(g, au); err != nil {
				return fmt.Errorf("audit create hook failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		logs.Errorf("create hook, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create hook, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// Update one hook instance.
func (dao *hookDao) Update(kit *kit.Kit, g *table.Hook) error {

	if g == nil {
		return errf.New(errf.InvalidParameter, "hook is nil")
	}

	if err := g.ValidateUpdate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentAppExist(kit, g.Attachment); err != nil {
		return err
	}

	opts := orm.NewFieldOptions().AddIgnoredFields(
		"id", "biz_id", "app_id")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(g, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	ab := dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.Hook).PrepareUpdate(g)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", string(table.HookTable), " SET ", expr, " WHERE id = ", strconv.Itoa(int(g.ID)),
		" AND biz_id = ", strconv.Itoa(int(g.Attachment.BizID)))
	sql := filter.SqlJoint(sqlSentence)

	err = dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			var effected int64
			effected, err = dao.orm.Txn(txn).Update(kit.Ctx, sql, toUpdate)
			if err != nil {
				logs.Errorf("update hook: %d failed, err: %v, rid: %v", g.ID, err, kit.Rid)
				return err
			}

			if effected == 0 {
				logs.Errorf("update one hook: %d, but record not found, rid: %v", g.ID, kit.Rid)
				return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
			}

			if effected > 1 {
				logs.Errorf("update one hook: %d, but got updated hook count: %d, rid: %v", g.ID,
					effected, kit.Rid)
				return fmt.Errorf("matched hook count %d is not as excepted", effected)
			}

			// do audit
			if err := ab.Do(&AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}); err != nil {
				return fmt.Errorf("do hook update audit failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

// List hooks with options.
func (dao *hookDao) List(kit *kit.Kit, opts *types.ListHooksOption) (
	*types.ListHookDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list hook options null")
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
	var sqlSentenceCount []string
	sqlSentenceCount = append(sqlSentenceCount, "SELECT COUNT(*) FROM ", string(table.HookTable), whereExpr)
	countSql := filter.SqlJoint(sqlSentenceCount)
	count, err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, countSql, args...)
	if err != nil {
		return nil, err
	}

	// query hook list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.HookColumns.NamedExpr(), " FROM ", string(table.HookTable), whereExpr, pageExpr)
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*table.Hook, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListHookDetails{Count: count, Details: list}, nil
}

// Delete one hook instance.
func (dao *hookDao) Delete(kit *kit.Kit, g *table.Hook) error {

	if g == nil {
		return errf.New(errf.InvalidParameter, "hook is nil")
	}

	if err := g.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentAppExist(kit, g.Attachment); err != nil {
		return err
	}

	ab := dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.Hook).PrepareDelete(g.ID)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", string(table.HookTable), " WHERE id = ", strconv.Itoa(int(g.ID)),
		" AND biz_id = ", strconv.Itoa(int(g.Attachment.BizID)))
	expr := filter.SqlJoint(sqlSentence)

	err := dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		// delete the hook at first.
		err := dao.orm.Txn(txn).Delete(kit.Ctx, expr)
		if err != nil {
			return err
		}

		// audit this delete hook details.
		auditOpt := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err := ab.Do(auditOpt); err != nil {
			return fmt.Errorf("audit delete hook failed, err: %v", err)
		}

		return nil
	})

	if err != nil {
		logs.Errorf("delete hook: %d failed, err: %v, rid: %v", g.ID, err, kit.Rid)
		return fmt.Errorf("delete hook, but run txn failed, err: %v", err)
	}

	return nil
}

// validateAttachmentResExist validate if attachment resource exists before creating hook.
func (dao *hookDao) validateAttachmentResExist(kit *kit.Kit, am *table.HookAttachment) error {
	return dao.validateAttachmentAppExist(kit, am)
}

// validateAttachmentAppExist validate if attachment app exists before creating hook.
func (dao *hookDao) validateAttachmentAppExist(kit *kit.Kit, am *table.HookAttachment) error {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "WHERE id = ", strconv.Itoa(int(am.AppID)), " AND biz_id = ", strconv.Itoa(int(am.BizID)))
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable, sql)
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("hook attached app %d is not exist", am.AppID))
	}

	return nil
}
