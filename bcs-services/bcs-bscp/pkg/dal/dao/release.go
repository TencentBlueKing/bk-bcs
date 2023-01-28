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

// Release supplies all the release related operations.
type Release interface {
	// CreateWithTx one release instance with tx.
	CreateWithTx(kit *kit.Kit, tx *sharding.Tx, release *table.Release) (uint32, error)
	// List releases with options.
	List(kit *kit.Kit, opts *types.ListReleasesOption) (*types.ListReleaseDetails, error)
}

var _ Release = new(releaseDao)

type releaseDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
}

// CreateWithTx one release instance with tx.
func (dao *releaseDao) CreateWithTx(kit *kit.Kit, tx *sharding.Tx, release *table.Release) (uint32, error) {
	if release == nil {
		return 0, errf.New(errf.InvalidParameter, "release is nil")
	}

	if err := release.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	if err := dao.validateAttachmentResExist(kit, release.Attachment); err != nil {
		return 0, err
	}

	// generate an release id and update to release.
	id, err := dao.idGen.One(kit, table.ReleaseTable)
	if err != nil {
		return 0, err
	}

	release.ID = id

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.ReleaseTable, table.ReleaseColumns.ColumnExpr(),
		table.ReleaseColumns.ColonNameExpr())

	if err = dao.orm.Txn(tx.Tx()).Insert(kit.Ctx, sql, release); err != nil {
		return 0, err
	}

	au := &AuditOption{Txn: tx.Tx(), ResShardingUid: tx.ShardingUid()}
	if err := dao.auditDao.Decorator(kit, release.Attachment.BizID,
		enumor.Release).AuditCreate(release, au); err != nil {
		return 0, fmt.Errorf("audit create release failed, err: %v", err)
	}

	return id, nil
}

// List releases with options.
func (dao *releaseDao) List(kit *kit.Kit, opts *types.ListReleasesOption) (
	*types.ListReleaseDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list releases options null")
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

	var sql string
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sql = fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ReleaseTable, whereExpr)
		var count uint32
		count, err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, sql)
		if err != nil {
			return nil, err
		}

		return &types.ListReleaseDetails{Count: count, Details: make([]*table.Release, 0)}, nil
	}

	// query release list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sql = fmt.Sprintf(`SELECT %s FROM %s %s %s`,
		table.ReleaseColumns.NamedExpr(), table.ReleaseTable, whereExpr, pageExpr)

	list := make([]*table.Release, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql)
	if err != nil {
		return nil, err
	}

	return &types.ListReleaseDetails{Count: 0, Details: list}, nil
}

// validateAttachmentResExist validate if attachment resource exists before creating release.
func (dao *releaseDao) validateAttachmentResExist(kit *kit.Kit, am *table.ReleaseAttachment) error {

	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable,
		fmt.Sprintf("WHERE id = %d AND biz_id = %d", am.AppID, am.BizID))
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("release attached app %d is not exist", am.AppID))
	}

	return nil
}
