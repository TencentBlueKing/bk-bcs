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
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/orm"
	"bscp.io/pkg/dal/sharding"
	"bscp.io/pkg/iam/client"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/runtime/filter"
	"bscp.io/pkg/types"
)

// IAM only used to iam to pull resource callback.
type IAM interface {
	// ListInstances list instances with options.
	ListInstances(kt *kit.Kit, opts *types.ListInstancesOption) (*types.ListInstanceDetails, error)
}

var _ IAM = new(iamDao)

type iamDao struct {
	orm orm.Interface
	sd  *sharding.Sharding
}

// ListInstances list instances with options.
func (r *iamDao) ListInstances(kt *kit.Kit, opts *types.ListInstancesOption) (
	*types.ListInstanceDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list instances options is null")
	}

	// enable unlimited query, because this is iam pull resource callback.
	po := &types.PageOption{MaxLimit: client.BkIAMMaxPageSize}
	if err := opts.Validate(po); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"biz_id", "app_id", "name"},
		CrownedOption: &filter.CrownedOption{
			CrownedOp: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "biz_id",
					Op:    filter.Equal.Factory(),
					Value: opts.BizID,
				}},
		},
	}
	whereExpr, args, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sql string
	var sqlSentence []string
	if opts.Page.Count {
		// count instance data by whereExpr
		sqlSentence = append(sqlSentence, "SELECT COUNT(*) FROM ", string(opts.TableName), whereExpr)
		sql = filter.SqlJoint(sqlSentence)
		var count uint32
		count, err = r.orm.Do(r.sd.ShardingOne(opts.BizID).DB()).Count(kt.Ctx, sql, args...)
		if err != nil {
			return nil, err
		}

		return &types.ListInstanceDetails{Count: count, Details: make([]*types.InstanceResource, 0)}, nil
	}

	// select instance data by whereExpr
	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sqlSentence = append(sqlSentence, "SELECT id, name FROM ", string(opts.TableName), whereExpr, pageExpr)
	sql = filter.SqlJoint(sqlSentence)
	list := make([]*types.InstanceResource, 0)
	err = r.orm.Do(r.sd.ShardingOne(opts.BizID).DB()).Select(kt.Ctx, &list, sql, args...)
	if err != nil {
		return nil, err
	}

	return &types.ListInstanceDetails{Count: 0, Details: list}, nil
}
