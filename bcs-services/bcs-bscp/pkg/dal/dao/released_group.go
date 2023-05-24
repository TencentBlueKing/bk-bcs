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
	"strconv"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// ReleasedGroup supplies all the group related operations.
type ReleasedGroup interface {
	// List list group current releases by option
	List(kit *kit.Kit, opts *types.ListReleasedGroupsOption) ([]*table.ReleasedGroup, error)
	// CountGroupsReleasedApps counts each group's published apps.
	CountGroupsReleasedApps(kit *kit.Kit, opts *types.CountGroupsReleasedAppsOption) (
		[]*types.GroupPublishedAppsCount, error)
	// UpdateEditedStatusWithTx update edited status with transaction
	UpdateEditedStatusWithTx(kit *kit.Kit, tx *sharding.Tx, edited bool, groupID, bizID uint32) error
}

var _ ReleasedGroup = new(releasedGroupDao)

type releasedGroupDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	lock     LockDao
}

// List list group current releases by option
func (dao *releasedGroupDao) List(kit *kit.Kit, opts *types.ListReleasedGroupsOption) (
	[]*table.ReleasedGroup, error) {
	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list group current releases option is nil")
	}

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"biz_id", "app_id"},
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
	sqlSentence = append(sqlSentence, "SELECT ", table.ReleasedGroupColumns.NamedExpr(), " FROM ",
		table.ReleasedGroupTable.Name(), whereExpr)
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*table.ReleasedGroup, 0)
	err = dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}
	return list, nil
}

// CountGroupsReleasedApps counts each group's published apps.
func (dao *releasedGroupDao) CountGroupsReleasedApps(kit *kit.Kit, opts *types.CountGroupsReleasedAppsOption) (
	[]*types.GroupPublishedAppsCount, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "count groups released apps option is nil")
	}

	if err := opts.Validate(nil); err != nil {
		return nil, err
	}

	var sqlBuf bytes.Buffer
	sqlBuf.WriteString("SELECT group_id, COUNT(DISTINCT app_id) AS counts, MAX(edited) AS edited FROM ")
	sqlBuf.WriteString(table.ReleasedGroupTable.Name())
	sqlBuf.WriteString(" WHERE biz_id = ? AND group_id IN (?) GROUP BY group_id")

	counts := make([]*types.GroupPublishedAppsCount, 0)
	err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx,
		&counts, sqlBuf.String(), opts.BizID, opts.Groups)
	if err != nil {
		return nil, err
	}
	return counts, nil
}

// UpdateEditedStatusWithTx update edited status with transaction
func (dao *releasedGroupDao) UpdateEditedStatusWithTx(kit *kit.Kit, tx *sharding.Tx, edited bool, groupID, bizID uint32) error {
	if bizID == 0 {
		return errf.New(errf.InvalidParameter, "bizID is 0")
	}
	if groupID == 0 {
		// group id is 0, means it is a default group,can not be edited
		return errf.New(errf.InvalidParameter, "groupID is 0")
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "UPDATE ", table.ReleasedGroupTable.Name(),
		" SET edited = ", strconv.FormatBool(edited), " WHERE biz_id = ", strconv.Itoa(int(bizID)))
	sql := filter.SqlJoint(sqlSentence)

	toUpdate := map[string]interface{}{
		"edited": edited,
	}

	_, err := dao.orm.Txn(tx.Tx()).Update(kit.Ctx, sql, toUpdate)
	if err != nil {
		return err
	}
	return nil
}
