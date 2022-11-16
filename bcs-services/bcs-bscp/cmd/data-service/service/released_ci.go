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

	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbrci "bscp.io/pkg/protocol/core/released-ci"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// ListReleasedConfigItems list released config items.
func (s *Service) ListReleasedConfigItems(ctx context.Context, req *pbds.ListReleasedCIsReq) (
	*pbds.ListReleasedCIsResp, error) {

	kt := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	query := &types.ListReleasedCIsOption{
		BizID:  req.BizId,
		Filter: filter,
		Page:   req.Page.BasePage(),
	}

	details, err := s.dao.ReleasedCI().List(kt, query)
	if err != nil {
		logs.Errorf("list released config item failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListReleasedCIsResp{
		Count:   details.Count,
		Details: pbrci.PbReleasedConfigItems(details.Details),
	}
	return resp, nil
}
