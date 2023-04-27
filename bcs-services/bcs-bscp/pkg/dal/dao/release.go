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

	"github.com/pkg/errors"

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
	// GetByName
	GetByName(kit *kit.Kit, bizID uint32, appID uint32, name string) (*table.Release, error)
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
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "INSERT INTO ", table.ReleaseTable.Name(), " (", table.ReleaseColumns.ColumnExpr(),
		")  VALUES(", table.ReleaseColumns.ColonNameExpr(), ")")
	sql := filter.SqlJoint(sqlSentence)

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

// GetByName 通过名称获取, 可以做唯一性校验
func (dao *releaseDao) GetByName(kit *kit.Kit, bizID uint32, appID uint32, name string) (*table.Release, error) {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.ReleaseColumns.NamedExpr(), " FROM ", table.ReleaseTable.Name(),
		" WHERE name = '", name, "' AND biz_id = ", strconv.Itoa(int(bizID)), " AND app_id = ", strconv.Itoa(int(appID)))
	expr := filter.SqlJoint(sqlSentence)

	one := new(table.Release)
	err := dao.orm.Do(dao.sd.Admin().DB()).Get(kit.Ctx, one, expr)
	if err != nil {
		return nil, errors.Wrapf(err, "get release name")
	}

	return one, nil
}

// List releases with options.
func (dao *releaseDao) List(kit *kit.Kit, opts *types.ListReleasesOption) (
	*types.ListReleaseDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list releases options null")
	}

	po := &types.PageOption{
		EnableUnlimitedLimit: true,
		DisabledSort:         false,
	}

	if err := opts.Validate(po); err != nil {
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
				&filter.AtomRule{
					Field: "deprecated",
					Op:    filter.Equal.Factory(),
					Value: opts.Deprecated,
				},
			},
		},
	}
	whereExpr, args, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sqlSentenceCount []string
	sqlSentenceCount = append(sqlSentenceCount, "SELECT COUNT(*) FROM ", table.ReleaseTable.Name(), whereExpr)
	countSql := filter.SqlJoint(sqlSentenceCount)
	count, err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Count(kit.Ctx, countSql, args...)
	if err != nil {
		return nil, err
	}

	// query release list for now.
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.ReleaseColumns.NamedExpr(),
		" FROM ", table.ReleaseTable.Name(), whereExpr, pageExpr)
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*table.Release, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListReleaseDetails{Count: count, Details: list}, nil
}

// validateAttachmentResExist validate if attachment resource exists before creating release.
func (dao *releaseDao) validateAttachmentResExist(kit *kit.Kit, am *table.ReleaseAttachment) error {
	var sqlSentence []string
	sqlSentence = append(sqlSentence, " WHERE id = ", strconv.Itoa(int(am.AppID)), " AND biz_id = ", strconv.Itoa(int(am.BizID)))
	sql := filter.SqlJoint(sqlSentence)
	exist, err := isResExist(kit, dao.orm, dao.sd.ShardingOne(am.BizID), table.AppTable, sql)
	if err != nil {
		return err
	}

	if !exist {
		return errf.New(errf.RelatedResNotExist, fmt.Sprintf("release attached app %d is not exist", am.AppID))
	}

	return nil
}
