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

package service

import (
	"context"
	"fmt"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/iam/client"
	"bscp.io/pkg/iam/sys"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// ListInstances list instances used to iam pull resource callback.
func (s *Service) ListInstances(ctx context.Context, req *pbds.ListInstancesReq) (*pbds.ListInstancesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pn struct to expression failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	var tableName table.Name
	switch client.TypeID(req.ResourceType) {
	case sys.Application:
		tableName = table.AppTable

	default:
		return nil, fmt.Errorf("resource type %s not support", req.ResourceType)
	}

	opts := &types.ListInstancesOption{
		BizID:     req.BizId,
		TableName: tableName,
		Filter:    filter,
		Page:      req.Page.BasePage(),
	}

	details, err := s.dao.IAM().ListInstances(kt, opts)
	if err != nil {
		logs.Errorf("list instances failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListInstancesResp{
		Count:   details.Count,
		Details: pbds.PbInstanceResources(details.Details),
	}

	return resp, nil
}
