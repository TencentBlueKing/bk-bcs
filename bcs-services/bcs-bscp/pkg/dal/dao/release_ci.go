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
	"bytes"
	"fmt"
	"strconv"

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
	// Get released config item by id and released id
	Get(kit *kit.Kit, id, bizID, releasedID uint32) (*table.ReleasedConfigItem, error)
	// GetReleasedLately released config item by app id and biz id
	GetReleasedLately(kit *kit.Kit, appId, bizID uint32, searchKey string) ([]*table.ReleasedConfigItem, error)
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

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.ReleasedConfigItemTable.Name(), " (", table.ReleasedConfigItemColumns.ColumnExpr(),
		")  VALUES(", table.ReleasedConfigItemColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)

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

// Get released config item by ID and released id
func (dao *releasedCIDao) Get(kit *kit.Kit, id, bizID, releasedID uint32) (*table.ReleasedConfigItem, error) {

	if id == 0 {
		return nil, errf.New(errf.InvalidParameter, "config item id can not be 0")
	}

	var sqlSentenceCount []string
	sqlSentenceCount = append(sqlSentenceCount, "SELECT ", table.ReleasedConfigItemColumns.NamedExpr(), " FROM ", table.ReleasedConfigItemTable.Name(),
		" WHERE config_item_id = ", strconv.Itoa(int(id)), " AND release_id = ", strconv.Itoa(int(releasedID)))
	sql := filter.SqlJoint(sqlSentenceCount)

	releasedCI := &table.ReleasedConfigItem{}
	if err := dao.orm.Do(dao.sd.ShardingOne(bizID).DB()).Get(kit.Ctx, releasedCI, sql); err != nil {
		return nil, err
	}
	return releasedCI, nil
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
	whereExpr, args, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sqlSentenceCount []string
	sqlSentenceCount = append(sqlSentenceCount, "SELECT COUNT(*) FROM ", table.ReleasedConfigItemTable.Name(), whereExpr)
	countSql := filter.SqlJoint(sqlSentenceCount)
	count, err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, countSql, args...)
	if err != nil {
		return nil, err
	}

	// query released config item list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.ReleasedConfigItemColumns.NamedExpr(), " FROM ", table.ReleasedConfigItemTable.Name(), whereExpr, pageExpr)
	sql := filter.SqlJoint(sqlSentence)
	list := make([]*table.ReleasedConfigItem, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListReleasedCIsDetails{Count: count, Details: list}, nil
}

// GetReleasedLately
func (dao *releasedCIDao) GetReleasedLately(kit *kit.Kit, appId, bizID uint32, searchKey string) (
	[]*table.ReleasedConfigItem, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "biz_id can not be 0")
	}

	var sqlBuf bytes.Buffer
	sqlBuf.WriteString("SELECT ")
	sqlBuf.WriteString(table.ReleasedConfigItemColumns.NamedExpr())
	sqlBuf.WriteString(" FROM ")
	sqlBuf.WriteString(table.ReleasedConfigItemTable.Name())
	sqlBuf.WriteString(" WHERE biz_id = ? AND app_id = ?")
	sqlBuf.WriteString(" AND (name like ? OR creator like ? OR reviser like ?)")
	sqlBuf.WriteString(" AND release_id = (SELECT release_id from ")
	sqlBuf.WriteString(table.ReleasedConfigItemTable.Name())
	sqlBuf.WriteString(" WHERE app_id = ? ORDER BY release_id desc limit 1)")

	fileInfo := make([]*table.ReleasedConfigItem, 0)
	err := dao.orm.Do(dao.sd.ShardingOne(bizID).DB()).Select(kit.Ctx, &fileInfo, sqlBuf.String(),
		bizID, appId, "%"+searchKey+"%", "%"+searchKey+"%", "%"+searchKey+"%", appId)
	if err != nil {
		return nil, err
	}
	return fileInfo, nil
}
