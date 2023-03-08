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
	"bscp.io/pkg/logs"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"

	"github.com/jmoiron/sqlx"
)

// Commit supplies all the commit related operations.
type Commit interface {
	// Create one commit instance.
	Create(kit *kit.Kit, commit *table.Commit) (uint32, error)
	// List commits with options.
	List(kit *kit.Kit, opts *types.ListCommitsOption) (*types.ListCommitDetails, error)
}

var _ Commit = new(commitDao)

type commitDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one commit instance.
func (dao *commitDao) Create(kit *kit.Kit, commit *table.Commit) (uint32, error) {

	if commit == nil {
		return 0, errf.New(errf.InvalidParameter, "commit is nil")
	}

	if err := commit.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentResExist(kit, commit.Attachment); err != nil {
		return 0, err
	}

	// generate an commit id and update to commit.
	id, err := dao.idGen.One(kit, table.CommitsTable)
	if err != nil {
		return 0, err
	}

	commit.ID = id

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.CommitsTable, table.CommitsColumns.ColumnExpr(),
		table.CommitsColumns.ColonNameExpr())

	err = dao.sd.ShardingOne(commit.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			if err := dao.orm.Txn(txn).Insert(kit.Ctx, sql, commit); err != nil {
				return err
			}

			// audit this to be create commit details.
			au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
			if err = dao.auditDao.Decorator(kit, commit.Attachment.BizID,
				enumor.Content).AuditCreate(commit, au); err != nil {
				return fmt.Errorf("audit create commit failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		logs.Errorf("create commit, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create commit, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// List commits with options.
func (dao *commitDao) List(kit *kit.Kit, opts *types.ListCommitsOption) (
	*types.ListCommitDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list commits options null")
	}

	if err := opts.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id", "biz_id", "app_id", "config_item_id"},
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
	whereExpr, arg, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sql string
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sql = fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.CommitsTable, whereExpr)
		var count uint32
		count, err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, sql, arg)
		if err != nil {
			return nil, err
		}

		return &types.ListCommitDetails{Count: count, Details: make([]*table.Commit, 0)}, nil
	}

	// query commit list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sql = fmt.Sprintf(`SELECT %s FROM %s %s %s`,
		table.CommitsColumns.NamedExpr(), table.CommitsTable, whereExpr, pageExpr)

	list := make([]*table.Commit, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql)
	if err != nil {
		return nil, err
	}

	return &types.ListCommitDetails{Count: 0, Details: list}, nil
}

// validateAttachmentResExist validate if attachment resource exists before creating commit.
func (dao *commitDao) validateAttachmentResExist(kit *kit.Kit, am *table.CommitAttachment) error {

	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable,
		fmt.Sprintf("WHERE id = %d AND biz_id = %d", am.AppID, am.BizID))
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("commit attached app %d is not exist", am.AppID))
	}

	exist, err = isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.ConfigItemTable,
		fmt.Sprintf("WHERE id = %d AND biz_id = %d AND app_id = %d", am.ConfigItemID, am.BizID, am.AppID))
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("commit attached config item %d is not exist",
			am.ConfigItemID))
	}

	return nil
}
