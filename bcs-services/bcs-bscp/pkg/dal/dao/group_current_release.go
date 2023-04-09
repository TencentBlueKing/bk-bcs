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

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// GroupCurrentRelease supplies all the group related operations.
type GroupCurrentRelease interface {
	// CountGroupsReleasedApps counts each group's published apps.
	CountGroupsReleasedApps(kit *kit.Kit, opts *types.CountGroupsReleasedAppsOption) (
		[]*types.GroupPublishedAppsCount, error)
	ListPublishedAppsByGrouID(kit *kit.Kit, groupID, bizID uint32) ([]*table.GroupCurrentRelease, error)
}

var _ GroupCurrentRelease = new(currentReleaseDao)

type currentReleaseDao struct {
	orm      orm.Interface
	sd       *sharding.Sharding
	idGen    IDGenInterface
	auditDao AuditDao
	lock     LockDao
}

// CountGroupsReleasedApps counts each group's published apps.
func (dao *currentReleaseDao) CountGroupsReleasedApps(kit *kit.Kit, opts *types.CountGroupsReleasedAppsOption) (
	[]*types.GroupPublishedAppsCount, error) {
	if err := opts.Validate(nil); err != nil {
		return nil, err
	}

	args := tools.JoinUint32(opts.Groups, ",")

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT group_id, COUNT(DISTINCT app_id) AS counts, MAX(edited) AS edited FROM ",
		table.GroupCurrentReleaseTable.Name(), fmt.Sprintf(" WHERE biz_id = %d AND group_id IN (%s) ", opts.BizID, args),
		" GROUP BY group_id ")
	sql := filter.SqlJoint(sqlSentence)

	counts := make([]*types.GroupPublishedAppsCount, 0)
	err := dao.orm.Do(dao.sd.ShardingOne(opts.BizID).DB()).Select(kit.Ctx, &counts, sql)
	if err != nil {
		return nil, err
	}
	return counts, nil
}

// ListPublishedAppsByGrouID list all published apps by group id
func (dao *currentReleaseDao) ListPublishedAppsByGrouID(kit *kit.Kit, groupID, bizID uint32) (
	[]*table.GroupCurrentRelease, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID is 0")
	}

	var sqlSentence []string
	sqlSentence = append(sqlSentence, "SELECT ", table.GroupCurrentReleaseColumns.NamedExpr(), " FROM ",
		table.GroupCurrentReleaseTable.Name(), " WHERE biz_id = ? AND group_id = ?")
	sql := filter.SqlJoint(sqlSentence)

	list := make([]*table.GroupCurrentRelease, 0)
	if err := dao.orm.Do(dao.sd.ShardingOne(bizID).DB()).Select(kit.Ctx, &list, sql, bizID, groupID); err != nil {
		return nil, err
	}
	return list, nil
}
