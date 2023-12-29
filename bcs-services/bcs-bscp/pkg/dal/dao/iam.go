/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dao

import (
	"fmt"
	"strconv"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/orm"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/sharding"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/client"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/sys"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// IAM only used to iam to pull resource callback.
type IAM interface {
	// ListInstances list instances with options.
	ListInstances(kt *kit.Kit, opts *types.ListInstancesOption) (*types.ListInstanceDetails, error)
	// FetchInstanceInfo fetch instance info with options.
	FetchInstanceInfo(kt *kit.Kit, opts *types.FetchInstanceInfoOption) (*types.FetchInstanceInfoDetails, error)
}

var _ IAM = new(iamDao)

type iamDao struct {
	orm  orm.Interface
	genQ *gen.Query
	sd   *sharding.Sharding
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

	var (
		count   int64
		details []*types.InstanceResource
	)
	switch opts.ResourceType {
	case meta.App.String():
		bizID, err := strconv.Atoi(opts.ParentID)
		if err != nil {
			return nil, err
		}
		m := r.genQ.App
		count, err = m.WithContext(kt.Ctx).
			Select(m.ID.As("id"), m.Name.As("name")).Where(m.BizID.Eq(uint32(bizID))).
			ScanByPage(&details, opts.Page.Offset(), opts.Page.LimitInt())
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("invalid resource type %s", opts.ResourceType)
	}

	return &types.ListInstanceDetails{Count: uint32(count), Details: details}, nil
}

// FetchInstanceInfo fetch instance info with options.
func (r *iamDao) FetchInstanceInfo(kt *kit.Kit, opts *types.FetchInstanceInfoOption) (
	*types.FetchInstanceInfoDetails, error) {

	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "fetch instance info options is null")
	}

	if err := opts.Validate(); err != nil {
		return nil, err
	}

	var details []*types.InstanceInfo
	switch opts.ResourceType {
	case meta.App.String():
		ids, err := tools.StringSliceToUint32Slice(opts.IDs)
		if err != nil {
			return nil, err
		}
		m := r.genQ.App
		apps, err := m.WithContext(kt.Ctx).Where(m.ID.In(ids...)).Find()
		if err != nil {
			return nil, err
		}
		for _, app := range apps {
			detail := &types.InstanceInfo{
				ID:          strconv.Itoa(int(app.ID)),
				DisplayName: app.Spec.Name,
				Approver:    []string{},
				Path:        []string{fmt.Sprintf("/%s,%d/", sys.Business, app.BizID)},
			}
			if app.Revision.Creator != "" {
				detail.Approver = append(detail.Approver, app.Revision.Creator)
			}
			if app.Revision.Reviser != "" {
				detail.Approver = append(detail.Approver, app.Revision.Reviser)
			}
			details = append(details, detail)
		}
	default:
		return nil, fmt.Errorf("invalid resource type %s", opts.ResourceType)
	}

	return &types.FetchInstanceInfoDetails{Details: details}, nil
}
