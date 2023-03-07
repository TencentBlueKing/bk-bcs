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

// Group supplies all the group related operations.
type Group interface {
	// Create one group instance.
	Create(kit *kit.Kit, group *table.Group) (uint32, error)
	// Update one group's info.
	Update(kit *kit.Kit, group *table.Group) error
	// List groups with options.
	List(kit *kit.Kit, opts *types.ListGroupsOption) (*types.ListGroupDetails, error)
	// Delete one strategy instance.
	Delete(kit *kit.Kit, strategy *table.Group) error
}

var _ Group = new(groupDao)

type groupDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	lock     LockDao
}

// Create one group instance.
func (dao *groupDao) Create(kit *kit.Kit, g *table.Group) (uint32, error) {

	if g == nil {
		return 0, errf.New(errf.InvalidParameter, "group is nil")
	}

	if err := g.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentResExist(kit, g.Attachment); err != nil {
		return 0, err
	}

	// generate a group id and update to group.
	id, err := dao.idGen.One(kit, table.GroupTable)
	if err != nil {
		return 0, err
	}

	g.ID = id

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.GroupTable,
		table.GroupColumns.ColumnExpr(), table.GroupColumns.ColonNameExpr())

	err = dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			if err := dao.orm.Txn(txn).Insert(kit.Ctx, sql, g); err != nil {
				return err
			}

			// audit this to be created group details.
			au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
			if err = dao.auditDao.Decorator(kit, g.Attachment.BizID,
				enumor.Group).AuditCreate(g, au); err != nil {
				return fmt.Errorf("audit create group failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		logs.Errorf("create group, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create group, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// Update one group instance.
func (dao *groupDao) Update(kit *kit.Kit, g *table.Group) error {

	if g == nil {
		return errf.New(errf.InvalidParameter, "group is nil")
	}

	if err := g.ValidateUpdate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentAppExist(kit, g.Attachment); err != nil {
		return err
	}

	opts := orm.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(
		"id", "biz_id", "app_id")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(g, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	ab := dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.Group).PrepareUpdate(g)

	sql := fmt.Sprintf(`UPDATE %s SET %s WHERE id = %d AND biz_id = %d`,
		table.GroupTable, expr, g.ID, g.Attachment.BizID)

	err = dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			var effected int64
			effected, err = dao.orm.Txn(txn).Update(kit.Ctx, sql, toUpdate)
			if err != nil {
				logs.Errorf("update group: %d failed, err: %v, rid: %v", g.ID, err, kit.Rid)
				return err
			}

			if effected == 0 {
				logs.Errorf("update one group: %d, but record not found, rid: %v", g.ID, kit.Rid)
				return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
			}

			if effected > 1 {
				logs.Errorf("update one group: %d, but got updated group count: %d, rid: %v", g.ID,
					effected, kit.Rid)
				return fmt.Errorf("matched group count %d is not as excepted", effected)
			}

			// do audit
			if err := ab.Do(&AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}); err != nil {
				return fmt.Errorf("do group update audit failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

// List groups with options.
func (dao *groupDao) List(kit *kit.Kit, opts *types.ListGroupsOption) (
	*types.ListGroupDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list group options null")
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

	countSql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.GroupTable, whereExpr)
	count, err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, countSql)
	if err != nil {
		return nil, err
	}

	// query group list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`,
		table.GroupColumns.NamedExpr(), table.GroupTable, whereExpr, pageExpr)

	list := make([]*table.Group, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql)
	if err != nil {
		return nil, err
	}

	return &types.ListGroupDetails{Count: count, Details: list}, nil
}

// Delete one group instance.
func (dao *groupDao) Delete(kit *kit.Kit, g *table.Group) error {

	if g == nil {
		return errf.New(errf.InvalidParameter, "group is nil")
	}

	if err := g.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentAppExist(kit, g.Attachment); err != nil {
		return err
	}

	// validate group under if strategy not exist.
	if err := dao.validateStrategyNotExist(kit, g.ID, g.Attachment.BizID); err != nil {
		return err
	}

	ab := dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.Group).PrepareDelete(g.ID)

	expr := fmt.Sprintf(`DELETE FROM %s WHERE id = %d AND biz_id = %d`, table.GroupTable,
		g.ID, g.Attachment.BizID)

	err := dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		// delete the group at first.
		err := dao.orm.Txn(txn).Delete(kit.Ctx, expr)
		if err != nil {
			return err
		}

		// audit this delete group details.
		auditOpt := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err := ab.Do(auditOpt); err != nil {
			return fmt.Errorf("audit delete group failed, err: %v", err)
		}

		// decrease the group lock count after the deletion
		lock := lockKey.Group(g.Attachment.BizID, g.Attachment.AppID)
		if err := dao.lock.DecreaseCount(kit, lock, &LockOption{Txn: txn}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logs.Errorf("delete group: %d failed, err: %v, rid: %v", g.ID, err, kit.Rid)
		return fmt.Errorf("delete group, but run txn failed, err: %v", err)
	}

	return nil
}

// validateAttachmentResExist validate if attachment resource exists before creating group.
func (dao *groupDao) validateAttachmentResExist(kit *kit.Kit, am *table.GroupAttachment) error {
	return dao.validateAttachmentAppExist(kit, am)
}

// validateAttachmentAppExist validate if attachment app exists before creating group.
func (dao *groupDao) validateAttachmentAppExist(kit *kit.Kit, am *table.GroupAttachment) error {

	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable,
		fmt.Sprintf("WHERE id = %d AND biz_id = %d", am.AppID, am.BizID))
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("group attached app %d is not exist", am.AppID))
	}

	return nil
}

// validateStrategyNotExist validate this group under if strategy not exist.
func (dao *groupDao) validateStrategyNotExist(kt *kit.Kit, bizID, id uint32) error {
	exist, err := isResExist(kt, dao.orm, dao.sd.ShardingOne(bizID), table.StrategyTable,
		fmt.Sprintf("WHERE biz_id = %d AND strategy_set_id = %d", bizID, id))
	if err != nil {
		return err
	}

	if exist {
		return errf.New(errf.InvalidParameter, "there are still strategy under the current group")
	}

	return nil
}
