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

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// GroupAppBind supplies all the group related operations.
type GroupAppBind interface {
	// BatchCreateWithTx batch create group app with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *sharding.Tx, items []*table.GroupAppBind) error
	// BatchDeleteByGroupIDWithTx batch delete group app by group id with transaction.
	BatchDeleteByGroupIDWithTx(kit *kit.Kit, tx *sharding.Tx, groupID, bizID uint32) error
	// BatchListByGroupIDs batch list group app by group ids.
	List(kit *kit.Kit, opts *types.ListGroupAppBindsOption) ([]*table.GroupAppBind, error)
	// Get get GroupAppBind by group id and app id.
	Get(kit *kit.Kit, groupID, appID, bizID uint32) (*table.GroupAppBind, error)
}

var _ GroupAppBind = new(groupAppDao)

type groupAppDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	lock     LockDao
}

// BatchCreateWithTx batch create group app with transaction.
func (dao *groupAppDao) BatchCreateWithTx(kit *kit.Kit, tx *sharding.Tx, items []*table.GroupAppBind) error {
	// validate released config item field.
	for _, item := range items {
		if err := item.ValidateCreate(); err != nil {
			return err
		}
	}

	// generate released config items id.
	ids, err := dao.idGen.Batch(kit, table.GroupAppBindTable, len(items))
	if err != nil {
		return err
	}

	start := 0
	for _, item := range items {
		item.ID = ids[start]
		start++
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.GroupAppBindTable.Name(), " (", table.GroupAppBindColumns.ColumnExpr(),
		")  VALUES(", table.GroupAppBindColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)

	return dao.orm.Txn(tx.Tx()).BulkInsert(kit.Ctx, sql, items)
}

// BatchDeleteByGroupIDWithTx batch delete group app by group id with transaction.
func (dao *groupAppDao) BatchDeleteByGroupIDWithTx(kit *kit.Kit, tx *sharding.Tx, groupID, bizID uint32) error {

	if groupID == 0 {
		return errf.New(errf.InvalidParameter, "group id is 0")
	}

	if bizID == 0 {
		return errf.New(errf.InvalidParameter, "biz id is 0")
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "DELETE FROM ", table.GroupAppBindTable.Name(), " WHERE group_id = ? AND biz_id = ?")

	sql := filter.SqlJoint(sqlSentence)

	if iErr := dao.orm.Txn(tx.Tx()).Delete(kit.Ctx, sql, groupID, bizID); iErr != nil {
		return iErr
	}

	return nil
}

func (dao *groupAppDao) List(kit *kit.Kit, opts *types.ListGroupAppBindsOption) ([]*table.GroupAppBind, error) {
	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "opts is nil")
	}
	if err := opts.Validate(); err != nil {
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
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.GroupAppBindColumns.NamedExpr(), " FROM ",
		table.GroupAppBindTable.Name(), whereExpr)
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*table.GroupAppBind, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// Get get GroupAppBind by group id and app id.
func (dao *groupAppDao) Get(kit *kit.Kit, groupID, appID, bizID uint32) (*table.GroupAppBind, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is 0")
	}
	if groupID == 0 {
		return nil, errf.New(errf.InvalidParameter, "group id is 0")
	}
	if appID == 0 {
		return nil, errf.New(errf.InvalidParameter, "app id is 0")
	}

	var sqlBuf bytes.Buffer
	sqlBuf.WriteString("SELECT ")
	sqlBuf.WriteString(table.GroupAppBindColumns.NamedExpr())
	sqlBuf.WriteString(" FROM ")
	sqlBuf.WriteString(table.GroupAppBindTable.Name())
	sqlBuf.WriteString(" WHERE group_id = ? AND app_id = ? AND biz_id = ?")

	item := &table.GroupAppBind{}
	if err := dao.orm.Do(dao.sd.ShardingOne(bizID).DB()).Get(kit.Ctx, item, sqlBuf.String(),
		groupID, appID, bizID); err != nil {
		return nil, err
	}
	return item, nil
}
