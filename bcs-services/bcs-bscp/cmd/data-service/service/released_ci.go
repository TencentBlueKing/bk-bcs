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
	pbrci "bscp.io/pkg/protocol/core/released-ci"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/search"
	"bscp.io/pkg/types"
)

// GetReleasedConfigItem get released config item
func (s *Service) GetReleasedConfigItem(ctx context.Context, req *pbds.GetReleasedCIReq) (
	*pbrci.ReleasedConfigItem, error) {

	kt := kit.FromGrpcContext(ctx)

	releasedCI, err := s.dao.ReleasedCI().Get(kt, req.ConfigItemId, req.BizId, req.ReleaseId)
	if err != nil {
		logs.Errorf("get released config item failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return pbrci.PbReleasedConfigItem(releasedCI), nil
}

// ListReleasedConfigItems list app bound template revisions.
func (s *Service) ListReleasedConfigItems(ctx context.Context,
	req *pbds.ListReleasedConfigItemsReq) (
	*pbds.ListReleasedConfigItemsResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// validate the page params
	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	searcher, err := search.NewSearcher(req.SearchFields, req.SearchValue, search.ReleasedConfigItem)
	if err != nil {
		return nil, err
	}

	details, count, err := s.dao.ReleasedCI().List(kt, req.BizId, req.AppId, req.ReleaseId, searcher, opt)
	if err != nil {
		logs.Errorf("list released app bound templates revisions failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListReleasedConfigItemsResp{
		Count:   uint32(count),
		Details: pbrci.PbReleasedConfigItems(details),
	}
	return resp, nil
}
