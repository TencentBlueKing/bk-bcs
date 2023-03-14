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

// GroupCategory supplies all the groupCategory related operations.
type GroupCategory interface {
	// Create one groupCategory instance.
	Create(kit *kit.Kit, groupCategory *table.GroupCategory) (uint32, error)
	// Update one groupCategory's info.
	Update(kit *kit.Kit, groupCategory *table.GroupCategory) error
	// List groupCategorys with options.
	List(kit *kit.Kit, opts *types.ListGroupCategoriesOption) (*types.ListGroupCategoriesDetails, error)
	// Delete one groupCategory instance.
	Delete(kit *kit.Kit, gc *table.GroupCategory) error
}

var _ GroupCategory = new(groupCategoryDao)

type groupCategoryDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	lock     LockDao
}

// Create one groupCategory instance.
func (dao *groupCategoryDao) Create(kit *kit.Kit, gc *table.GroupCategory) (uint32, error) {

	if gc == nil {
		return 0, errf.New(errf.InvalidParameter, "groupCategory is nil")
	}

	if err := gc.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentResExist(kit, gc.Attachment); err != nil {
		return 0, err
	}

	// generate a groupCategory id and update to groupCategory.
	id, err := dao.idGen.One(kit, table.GroupCategoryTable)
	if err != nil {
		return 0, err
	}

	gc.ID = id

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", string(table.GroupCategoryTable), " (", table.GroupCategoryColumns.ColumnExpr(), ")  VALUES(", table.GroupCategoryColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)

	err = dao.sd.ShardingOne(gc.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			if err := dao.orm.Txn(txn).Insert(kit.Ctx, sql, gc); err != nil {
				return err
			}

			// audit this to be created groupCategory details.
			au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
			if err = dao.auditDao.Decorator(kit, gc.Attachment.BizID,
				enumor.GroupCategory).AuditCreate(gc, au); err != nil {
				return fmt.Errorf("audit create groupCategory failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		logs.Errorf("create groupCategory, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create groupCategory, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// Update one groupCategory instance.
func (dao *groupCategoryDao) Update(kit *kit.Kit, gc *table.GroupCategory) error {

	if gc == nil {
		return errf.New(errf.InvalidParameter, "groupCategory is nil")
	}

	if err := gc.ValidateUpdate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentAppExist(kit, gc.Attachment); err != nil {
		return err
	}

	opts := orm.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(
		"id", "biz_id", "app_id")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(gc, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	ab := dao.auditDao.Decorator(kit, gc.Attachment.BizID, enumor.GroupCategory).PrepareUpdate(gc)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", string(table.GroupCategoryTable), " SET ", expr,
		" WHERE id = ", strconv.Itoa(int(gc.ID)), " AND biz_id = ", strconv.Itoa(int(gc.Attachment.BizID)))
	sql := filter.SqlJoint(sqlSentence)

	err = dao.sd.ShardingOne(gc.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			var effected int64
			effected, err = dao.orm.Txn(txn).Update(kit.Ctx, sql, toUpdate)
			if err != nil {
				logs.Errorf("update groupCategory: %d failed, err: %v, rid: %v", gc.ID, err, kit.Rid)
				return err
			}

			if effected == 0 {
				logs.Errorf("update one groupCategory: %d, but record not found, rid: %v", gc.ID, kit.Rid)
				return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
			}

			if effected > 1 {
				logs.Errorf("update one groupCategory: %d, but got updated groupCategory count: %d, rid: %v", gc.ID,
					effected, kit.Rid)
				return fmt.Errorf("matched groupCategory count %d is not as excepted", effected)
			}

			// do audit
			if err := ab.Do(&AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}); err != nil {
				return fmt.Errorf("do groupCategory update audit failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

// List groupCategorys with options.
func (dao *groupCategoryDao) List(kit *kit.Kit, opts *types.ListGroupCategoriesOption) (
	*types.ListGroupCategoriesDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list groupCategory options null")
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

	var sql string
	var sqlSentence []string
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sqlSentence = append(sqlSentence, "SELECT COUNT(*) FROM ", string(table.GroupCategoryTable), whereExpr)
		sql = filter.SqlJoint(sqlSentence)
		var count uint32
		count, err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, sql, args...)
		if err != nil {
			return nil, err
		}

		return &types.ListGroupCategoriesDetails{Count: count, Details: make([]*table.GroupCategory, 0)}, nil
	}

	// query groupCategory list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sqlSentence = append(sqlSentence, "SELECT ", table.GroupCategoryColumns.NamedExpr(), " FROM ", string(table.GroupCategoryTable), whereExpr, pageExpr)
	sql = filter.SqlJoint(sqlSentence)

	list := make([]*table.GroupCategory, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListGroupCategoriesDetails{Count: 0, Details: list}, nil
}

// Delete one groupCategory instance.
func (dao *groupCategoryDao) Delete(kit *kit.Kit, gc *table.GroupCategory) error {

	if gc == nil {
		return errf.New(errf.InvalidParameter, "groupCategory is nil")
	}

	if err := gc.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentAppExist(kit, gc.Attachment); err != nil {
		return err
	}

	// validate groupCategory under if group not exist.
	if err := dao.validateGroupNotExist(kit, gc.ID, gc.Attachment.BizID); err != nil {
		return err
	}

	ab := dao.auditDao.Decorator(kit, gc.Attachment.BizID, enumor.GroupCategory).PrepareDelete(gc.ID)
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", string(table.GroupCategoryTable), " WHERE id = ", strconv.Itoa(int(gc.ID)),
		" AND biz_id = ", strconv.Itoa(int(gc.Attachment.BizID)))
	expr := filter.SqlJoint(sqlSentence)

	err := dao.sd.ShardingOne(gc.Attachment.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		// delete the groupCategory at first.
		err := dao.orm.Txn(txn).Delete(kit.Ctx, expr)
		if err != nil {
			return err
		}

		// audit this delete groupCategory details.
		auditOpt := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err := ab.Do(auditOpt); err != nil {
			return fmt.Errorf("audit delete groupCategory failed, err: %v", err)
		}

		// decrease the groupCategory lock count after the deletion
		lock := lockKey.GroupCategory(gc.Attachment.BizID, gc.Attachment.AppID)
		if err := dao.lock.DecreaseCount(kit, lock, &LockOption{Txn: txn}); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logs.Errorf("delete groupCategory: %d failed, err: %v, rid: %v", gc.ID, err, kit.Rid)
		return fmt.Errorf("delete groupCategory, but run txn failed, err: %v", err)
	}

	return nil
}

// validateAttachmentResExist validate if attachment resource exists before creating groupCategory.
func (dao *groupCategoryDao) validateAttachmentResExist(kit *kit.Kit, am *table.GroupCategoryAttachment) error {
	return dao.validateAttachmentAppExist(kit, am)
}

// validateAttachmentAppExist validate if attachment app exists before creating groupCategory.
func (dao *groupCategoryDao) validateAttachmentAppExist(kit *kit.Kit, am *table.GroupCategoryAttachment) error {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, " WHERE id = ", strconv.Itoa(int(am.AppID)), " AND biz_id = ", strconv.Itoa(int(am.BizID)))
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable, sql)
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("groupCategory attached app %d is not exist", am.AppID))
	}

	return nil
}

// validateGroupNotExist validate this groupCategory under if group not exist.
func (dao *groupCategoryDao) validateGroupNotExist(kt *kit.Kit, bizID, id uint32) error {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, " WHERE biz_id = ", strconv.Itoa(int(bizID)), " AND group_category_id = ", strconv.Itoa(int(id)))
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kt, dao.orm, dao.sd.ShardingOne(bizID), table.GroupTable, sql)
	if err != nil {
		return err
	}

	if exist {
		return errf.New(errf.InvalidParameter, "there are still group under the current groupCategory")
	}

	return nil
}
