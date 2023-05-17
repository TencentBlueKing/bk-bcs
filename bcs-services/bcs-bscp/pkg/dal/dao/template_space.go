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

// TemplateSpace supplies all the TemplateSpace related operations.
type TemplateSpace interface {
	// Create one TemplateSpace instance.
	Create(kit *kit.Kit, TemplateSpace *table.TemplateSpace) (uint32, error)
	// Update one TemplateSpace's info.
	Update(kit *kit.Kit, TemplateSpace *table.TemplateSpace) error
	// List TemplateSpaces with options.
	List(kit *kit.Kit, opts *types.ListTemplateSpacesOption) (*types.ListTemplateSpaceDetails, error)
	// Delete one strategy instance.
	Delete(kit *kit.Kit, strategy *table.TemplateSpace) error
	// GetByName get templateSpace by name.
	GetByName(kit *kit.Kit, bizID uint32, name string) (*table.TemplateSpace, error)
}

var _ TemplateSpace = new(templateSpaceDao)

type templateSpaceDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one TemplateSpace instance.
func (dao *templateSpaceDao) Create(kit *kit.Kit, g *table.TemplateSpace) (uint32, error) {

	if g == nil {
		return 0, errf.New(errf.InvalidParameter, "TemplateSpace is nil")
	}

	if err := g.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	// generate a TemplateSpace id and update to TemplateSpace.
	id, err := dao.idGen.One(kit, table.TemplateSpaceTable)
	if err != nil {
		return 0, err
	}

	g.ID = id
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.TemplateSpaceTable.Name(), " (", table.TemplateSpaceColumns.ColumnExpr(), ")  VALUES(", table.TemplateSpaceColumns.ColonNameExpr(), ")")

	sql := filter.SqlJoint(sqlSentence)

	err = dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			if err := dao.orm.Txn(txn).Insert(kit.Ctx, sql, g); err != nil {
				return err
			}

			// audit this to be created TemplateSpace details.
			au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
			if err = dao.auditDao.Decorator(kit, g.Attachment.BizID,
				enumor.TemplateSpace).AuditCreate(g, au); err != nil {
				return fmt.Errorf("audit create TemplateSpace failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		logs.Errorf("create TemplateSpace, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create TemplateSpace, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// Update one TemplateSpace instance.
func (dao *templateSpaceDao) Update(kit *kit.Kit, g *table.TemplateSpace) error {

	if g == nil {
		return errf.New(errf.InvalidParameter, "TemplateSpace is nil")
	}

	if err := g.ValidateUpdate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	opts := orm.NewFieldOptions().AddIgnoredFields(
		"id", "biz_id")
	expr, toUpdate, err := orm.RearrangeSQLDataWithOption(g, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql expr failed, err: %v", err)
	}

	ab := dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.TemplateSpace).PrepareUpdate(g)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.TemplateSpaceTable.Name(), " SET ", expr, " WHERE id = ", strconv.Itoa(int(g.ID)),
		" AND biz_id = ", strconv.Itoa(int(g.Attachment.BizID)))
	sql := filter.SqlJoint(sqlSentence)

	err = dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			var effected int64
			effected, err = dao.orm.Txn(txn).Update(kit.Ctx, sql, toUpdate)
			if err != nil {
				logs.Errorf("update TemplateSpace: %d failed, err: %v, rid: %v", g.ID, err, kit.Rid)
				return err
			}

			if effected == 0 {
				logs.Errorf("update one TemplateSpace: %d, but record not found, rid: %v", g.ID, kit.Rid)
				return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
			}

			if effected > 1 {
				logs.Errorf("update one TemplateSpace: %d, but got updated TemplateSpace count: %d, rid: %v", g.ID,
					effected, kit.Rid)
				return fmt.Errorf("matched TemplateSpace count %d is not as excepted", effected)
			}

			// do audit
			if err := ab.Do(&AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}); err != nil {
				return fmt.Errorf("do TemplateSpace update audit failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		return err
	}

	return nil
}

// List TemplateSpaces with options.
func (dao *templateSpaceDao) List(kit *kit.Kit, opts *types.ListTemplateSpacesOption) (
	*types.ListTemplateSpaceDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list TemplateSpace options null")
	}

	if err := opts.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id", "biz_id"},
		CrownedOption: &filter.CrownedOption{
			CrownedOp: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "biz_id",
					Op:    filter.Equal.Factory(),
					Value: opts.BizID,
				},
			},
		},
	}
	whereExpr, args, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}
	var sqlSentenceCount []string
	sqlSentenceCount = append(sqlSentenceCount, "SELECT COUNT(*) FROM ", table.TemplateSpaceTable.Name(), whereExpr)
	countSql := filter.SqlJoint(sqlSentenceCount)
	count, err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, countSql, args...)
	if err != nil {
		return nil, err
	}

	// query TemplateSpace list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.TemplateSpaceColumns.NamedExpr(), " FROM ", table.TemplateSpaceTable.Name(), whereExpr, pageExpr)
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*table.TemplateSpace, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListTemplateSpaceDetails{Count: count, Details: list}, nil
}

// Delete one TemplateSpace instance.
func (dao *templateSpaceDao) Delete(kit *kit.Kit, g *table.TemplateSpace) error {

	if g == nil {
		return errf.New(errf.InvalidParameter, "TemplateSpace is nil")
	}

	if err := g.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	ab := dao.auditDao.Decorator(kit, g.Attachment.BizID, enumor.TemplateSpace).PrepareDelete(g.ID)

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.TemplateSpaceTable.Name(), " WHERE id = ", strconv.Itoa(int(g.ID)),
		" AND biz_id = ", strconv.Itoa(int(g.Attachment.BizID)))
	expr := filter.SqlJoint(sqlSentence)

	err := dao.sd.ShardingOne(g.Attachment.BizID).AutoTxn(kit, func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
		// delete the TemplateSpace at first.
		err := dao.orm.Txn(txn).Delete(kit.Ctx, expr)
		if err != nil {
			return err
		}

		// audit this delete TemplateSpace details.
		auditOpt := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
		if err := ab.Do(auditOpt); err != nil {
			return fmt.Errorf("audit delete TemplateSpace failed, err: %v", err)
		}

		return nil
	})

	if err != nil {
		logs.Errorf("delete TemplateSpace: %d failed, err: %v, rid: %v", g.ID, err, kit.Rid)
		return fmt.Errorf("delete TemplateSpace, but run txn failed, err: %v", err)
	}

	return nil
}

// GetByName get by name
func (dao *templateSpaceDao) GetByName(kit *kit.Kit, bizID uint32, name string) (*table.TemplateSpace, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.TemplateSpaceColumns.NamedExpr(), " FROM ", table.TemplateSpaceTable.Name(),
		" WHERE name = '", name, "' AND biz_id = ", strconv.Itoa(int(bizID)))
	expr := filter.SqlJoint(sqlSentence)
	one := new(table.TemplateSpace)
	err := dao.orm.Do(dao.sd.Admin().DB()).Get(kit.Ctx, one, expr)
	if err != nil {
		return nil, fmt.Errorf("get app details failed, err: %v", err)
	}

	return one, nil
}
