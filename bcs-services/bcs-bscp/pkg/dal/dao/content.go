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

// Content supplies all the content related operations.
type Content interface {
	// Create one content instance.
	Create(kit *kit.Kit, content *table.Content) (uint32, error)
	// List contents with options.
	List(kit *kit.Kit, opts *types.ListContentsOption) (*types.ListContentDetails, error)
}

var _ Content = new(contentDao)

type contentDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one content instance
func (dao *contentDao) Create(kit *kit.Kit, content *table.Content) (uint32, error) {

	if content == nil {
		return 0, errf.New(errf.InvalidParameter, "content is nil")
	}

	if err := content.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentResExist(kit, content.Attachment); err != nil {
		return 0, err
	}

	// generate an content id and update to content.
	id, err := dao.idGen.One(kit, table.ContentTable)
	if err != nil {
		return 0, err
	}

	content.ID = id

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.ContentTable, table.ContentColumns.ColumnExpr(),
		table.ContentColumns.ColonNameExpr())

	err = dao.sd.ShardingOne(content.Attachment.BizID).AutoTxn(kit,
		func(txn *sqlx.Tx, opt *sharding.TxnOption) error {
			if err := dao.orm.Txn(txn).Insert(kit.Ctx, sql, content); err != nil {
				return err
			}

			// audit this to be create content details.
			au := &AuditOption{Txn: txn, ResShardingUid: opt.ShardingUid}
			if err = dao.auditDao.Decorator(kit, content.Attachment.BizID,
				enumor.Content).AuditCreate(content, au); err != nil {
				return fmt.Errorf("audit create content failed, err: %v", err)
			}

			return nil
		})

	if err != nil {
		logs.Errorf("create content, but do auto txn failed, err: %v, rid: %s", err, kit.Rid)
		return 0, fmt.Errorf("create content, but auto run txn failed, err: %v", err)
	}

	return id, nil
}

// List contents with options.
func (dao *contentDao) List(kit *kit.Kit, opts *types.ListContentsOption) (
	*types.ListContentDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list content options null")
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
	whereExpr, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sql string
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sql = fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ContentTable, whereExpr)
		var count uint32
		count, err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, sql)
		if err != nil {
			return nil, err
		}

		return &types.ListContentDetails{Count: count, Details: make([]*table.Content, 0)}, nil
	}

	// query content list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sql = fmt.Sprintf(`SELECT %s FROM %s %s %s`,
		table.ContentColumns.NamedExpr(), table.ContentTable, whereExpr, pageExpr)

	list := make([]*table.Content, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql)
	if err != nil {
		return nil, err
	}

	return &types.ListContentDetails{Count: 0, Details: list}, nil
}

// validateAttachmentResExist validate if attachment resource exists before creating content.
func (dao *contentDao) validateAttachmentResExist(kit *kit.Kit, am *table.ContentAttachment) error {

	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable,
		fmt.Sprintf("WHERE id = %d AND biz_id = %d", am.AppID, am.BizID))
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("content attached app %d is not exist", am.AppID))
	}

	exist, err = isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.ConfigItemTable,
		fmt.Sprintf("WHERE id = %d AND biz_id = %d AND app_id = %d", am.ConfigItemID, am.BizID, am.AppID))
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("content attached config item %d is not exist",
			am.ConfigItemID))
	}

	return nil
}
