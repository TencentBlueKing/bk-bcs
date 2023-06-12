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
	"bscp.io/pkg/dal/gen"
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
	// CreateWithTx create one commit instance with transaction
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, commit *table.Commit) (uint32, error)
	// BatchCreateWithTx batch create commit instances with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, commits []*table.Commit) error
	// BatchListLatestCommits batch list config itmes' latest commit.
	BatchListLatestCommits(kit *kit.Kit, bizID, appID uint32, ids []uint32) ([]*table.Commit, error)
	// List commits with options.
	List(kit *kit.Kit, opts *types.ListCommitsOption) (*types.ListCommitDetails, error)
}

var _ Commit = new(commitDao)

type commitDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	genQ     *gen.Query
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
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.CommitsTable.Name(),
		" (", table.CommitsColumns.ColumnExpr(), ")  VALUES(", table.CommitsColumns.ColonNameExpr(), ")")

	sql := filter.SqlJoint(sqlSentence)

	err = dao.sd.ShardingOne(commit.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			if e := dao.orm.Txn(txn).Insert(kit.Ctx, sql, commit); e != nil {
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

// CreateWithTx create one commit instance with transaction
func (dao *commitDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, commit *table.Commit) (uint32, error) {

	if commit == nil {
		return 0, errf.New(errf.InvalidParameter, "commit is nil")
	}

	if err := commit.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	// generate an commit id and update to commit.
	id, err := dao.idGen.One(kit, table.CommitsTable)

	if err != nil {
		return 0, err
	}

	commit.ID = id
	if err := tx.Query.Commit.WithContext(kit.Ctx).Create(commit); err != nil {
		return 0, err
	}

	ad := dao.auditDao.DecoratorV2(kit, commit.Attachment.BizID).PrepareCreate(commit)
	if err := ad.Do(tx.Query); err != nil {
		return 0, fmt.Errorf("audit create commit failed, err: %v", err)
	}

	return id, nil
}

// BatchCreateWithTx batch create commit instances with transaction.
// NOTE: 1. this method won't audit, because it's batch operation.
// 2. this method won't validate attachment resource exist, because it's batch operation.
func (dao *commitDao) BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, commits []*table.Commit) error {
	if len(commits) == 0 {
		return nil
	}
	ids, err := dao.idGen.Batch(kit, table.CommitsTable, len(commits))
	if err != nil {
		return err
	}
	for i, commit := range commits {
		if err := commit.ValidateCreate(); err != nil {
			return err
		}
		commit.ID = ids[i]
	}
	if err := tx.Query.Commit.WithContext(kit.Ctx).Save(commits...); err != nil {
		return err
	}
	return nil
}

// BatchListLatestCommits batch list config itmes' latest commit.
func (dao *commitDao) BatchListLatestCommits(kit *kit.Kit, bizID, appID uint32, ids []uint32) ([]*table.Commit, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	m := dao.genQ.Commit
	subQuery := dao.genQ.Commit.WithContext(kit.Ctx).Select(m.ID.Max().As("commit_id")).Where(
		m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ConfigItemID.In(ids...)).Group(m.ConfigItemID)
	return dao.genQ.Commit.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID),
		dao.genQ.Commit.WithContext(kit.Ctx).Columns(m.ID).In(subQuery)).Find()
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
	whereExpr, args, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sql string
	var sqlSentence []string
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sqlSentence = append(sqlSentence, "SELECT COUNT(*) FROM ", table.CommitsTable.Name(), whereExpr)
		sql = filter.SqlJoint(sqlSentence)
		var count uint32
		count, err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, sql, args...)
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

	sqlSentence = append(sqlSentence, "SELECT ", table.CommitsColumns.NamedExpr(),
		" FROM ", table.CommitsTable.Name(), whereExpr, pageExpr)
	sql = filter.SqlJoint(sqlSentence)

	list := make([]*table.Commit, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListCommitDetails{Count: 0, Details: list}, nil
}

// validateAttachmentResExist validate if attachment resource exists before creating commit.
func (dao *commitDao) validateAttachmentResExist(kit *kit.Kit, am *table.CommitAttachment) error {

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "WHERE id = ", strconv.Itoa(int(am.AppID)),
		" AND biz_id = ", strconv.Itoa(int(am.BizID)))
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable, sql)
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("commit attached app %d is not exist", am.AppID))
	}

	var sqlSentenceRes []string
	sqlSentenceRes = append(sqlSentenceRes, "WHERE id = ", strconv.Itoa(int(am.ConfigItemID)),
		" AND biz_id = ", strconv.Itoa(int(am.BizID)), " AND app_id = ", strconv.Itoa(int(am.AppID)))
	sqlRes := filter.SqlJoint(sqlSentenceRes)
	exist, err = isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.ConfigItemTable, sqlRes)
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("commit attached config item %d is not exist",
			am.ConfigItemID))
	}

	return nil
}
