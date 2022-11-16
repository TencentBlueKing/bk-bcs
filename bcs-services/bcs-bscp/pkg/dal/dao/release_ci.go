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
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// ReleasedCI supplies all the released config item related operations.
type ReleasedCI interface {
	// BulkCreateWithTx bulk create released config items with tx.
	BulkCreateWithTx(kit *kit.Kit, tx *sharding.Tx, items []*table.ReleasedConfigItem) error
	// List released config items with options.
	List(kit *kit.Kit, opts *types.ListReleasedCIsOption) (*types.ListReleasedCIsDetails, error)
}

var _ ReleasedCI = new(releasedCIDao)

type releasedCIDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
}

// BulkCreateWithTx bulk create released config items.
func (dao *releasedCIDao) BulkCreateWithTx(kit *kit.Kit, tx *sharding.Tx, items []*table.ReleasedConfigItem) error {
	if items == nil {
		return errf.New(errf.InvalidParameter, "released config items is nil")
	}

	// validate released config item field.
	for _, item := range items {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	// generate released config items id.
	ids, err := dao.idGen.Batch(kit, table.ReleasedConfigItemTable, len(items))
	if err != nil {
		return err
	}

	start := 0
	for _, item := range items {
		item.ID = ids[start]
		start++
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.ReleasedConfigItemTable,
		table.ReleasedConfigItemColumns.ColumnExpr(), table.ReleasedConfigItemColumns.ColonNameExpr())

	if err = dao.orm.Txn(tx.Tx()).BulkInsert(kit.Ctx, sql, items); err != nil {
		return err
	}

	au := &AuditOption{Txn: tx.Tx(), ResShardingUid: tx.ShardingUid()}
	if err := dao.auditDao.Decorator(kit, items[0].Attachment.BizID,
		enumor.Release).AuditCreate(items, au); err != nil {
		return fmt.Errorf("audit create released config items failed, err: %v", err)
	}

	return nil
}

// List released config items with options.
func (dao *releasedCIDao) List(kit *kit.Kit, opts *types.ListReleasedCIsOption) (
	*types.ListReleasedCIsDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list released config items options null")
	}

	po := &types.PageOption{
		// allows list released ci without page
		EnableUnlimitedLimit: true,
		DisabledSort:         false,
	}
	if err := opts.Validate(po); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id", "release_id", "biz_id", "app_id"},
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
	whereExpr, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sql string
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sql = fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ReleasedConfigItemTable, whereExpr)
		count, err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, sql)
		if err != nil {
			return nil, err
		}

		return &types.ListReleasedCIsDetails{Count: count, Details: make([]*table.ReleasedConfigItem, 0)}, nil
	}

	// query released config item list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sql = fmt.Sprintf(`SELECT %s FROM %s %s %s`,
		table.ReleasedConfigItemColumns.NamedExpr(), table.ReleasedConfigItemTable, whereExpr, pageExpr)

	list := make([]*table.ReleasedConfigItem, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql)
	if err != nil {
		return nil, err
	}

	return &types.ListReleasedCIsDetails{Count: 0, Details: list}, nil
}
