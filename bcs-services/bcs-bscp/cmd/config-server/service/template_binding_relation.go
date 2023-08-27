/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"
	"fmt"

	"bscp.io/pkg/criteria/constant"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/tools"
)

// ListTemplateBoundCounts list template bound counts
func (s *Service) ListTemplateBoundCounts(ctx context.Context, req *pbcs.ListTemplateBoundCountsReq) (
	*pbcs.ListTemplateBoundCountsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateBoundCountsResp)

	// validate input param
	ids := tools.SliceRepeatedElements(req.TemplateIds)
	if len(ids) > 0 {
		return nil, fmt.Errorf("repeated template ids: %v, id must be unique", ids)
	}
	idsLen := len(req.TemplateIds)
	if idsLen == 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateBoundCountsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateIds:     req.TemplateIds,
	}

	rp, err := s.client.DS.ListTemplateBoundCounts(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template bound counts failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateBoundCountsResp{
		Details: rp.Details,
	}
	return resp, nil
}

// ListTemplateRevisionBoundCounts list template bound counts
func (s *Service) ListTemplateRevisionBoundCounts(ctx context.Context, req *pbcs.ListTemplateRevisionBoundCountsReq) (
	*pbcs.ListTemplateRevisionBoundCountsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateRevisionBoundCountsResp)

	// validate input param
	ids := tools.SliceRepeatedElements(req.TemplateRevisionIds)
	if len(ids) > 0 {
		return nil, fmt.Errorf("repeated template revision ids: %v, id must be unique", ids)
	}
	idsLen := len(req.TemplateRevisionIds)
	if idsLen <= 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template revision ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateRevisionBoundCountsReq{
		BizId:               req.BizId,
		TemplateSpaceId:     req.TemplateSpaceId,
		TemplateId:          req.TemplateId,
		TemplateRevisionIds: req.TemplateRevisionIds,
	}

	rp, err := s.client.DS.ListTemplateRevisionBoundCounts(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template revision bound counts failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateRevisionBoundCountsResp{
		Details: rp.Details,
	}
	return resp, nil
}

// ListTemplateSetBoundCounts list template bound counts
func (s *Service) ListTemplateSetBoundCounts(ctx context.Context, req *pbcs.ListTemplateSetBoundCountsReq) (
	*pbcs.ListTemplateSetBoundCountsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateSetBoundCountsResp)

	// validate input param
	ids := tools.SliceRepeatedElements(req.TemplateSetIds)
	if len(ids) > 0 {
		return nil, fmt.Errorf("repeated template set ids: %v, id must be unique", ids)
	}
	idsLen := len(req.TemplateSetIds)
	if idsLen <= 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template set ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateSetBoundCountsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateSetIds:  req.TemplateSetIds,
	}

	rp, err := s.client.DS.ListTemplateSetBoundCounts(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template set bound counts failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateSetBoundCountsResp{
		Details: rp.Details,
	}
	return resp, nil
}

// ListTemplateBoundUnnamedAppDetails list template bound unnamed app details
func (s *Service) ListTemplateBoundUnnamedAppDetails(ctx context.Context,
	req *pbcs.ListTemplateBoundUnnamedAppDetailsReq) (
	*pbcs.ListTemplateBoundUnnamedAppDetailsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateBoundUnnamedAppDetailsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateBoundUnnamedAppDetailsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateId:      req.TemplateId,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTemplateBoundUnnamedAppDetails(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template bound unnamed app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateBoundUnnamedAppDetailsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTemplateBoundNamedAppDetails list template bound named app details
func (s *Service) ListTemplateBoundNamedAppDetails(ctx context.Context,
	req *pbcs.ListTemplateBoundNamedAppDetailsReq) (
	*pbcs.ListTemplateBoundNamedAppDetailsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateBoundNamedAppDetailsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateBoundNamedAppDetailsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateId:      req.TemplateId,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTemplateBoundNamedAppDetails(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template bound named app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateBoundNamedAppDetailsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTemplateBoundTemplateSetDetails list template bound template set details
func (s *Service) ListTemplateBoundTemplateSetDetails(ctx context.Context,
	req *pbcs.ListTemplateBoundTemplateSetDetailsReq) (
	*pbcs.ListTemplateBoundTemplateSetDetailsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateBoundTemplateSetDetailsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateBoundTemplateSetDetailsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateId:      req.TemplateId,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTemplateBoundTemplateSetDetails(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template bound template set details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateBoundTemplateSetDetailsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListMultiTemplateBoundTemplateSetDetails list multiple template bound template set details
func (s *Service) ListMultiTemplateBoundTemplateSetDetails(ctx context.Context,
	req *pbcs.ListMultiTemplateBoundTemplateSetDetailsReq) (
	*pbcs.ListMultiTemplateBoundTemplateSetDetailsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListMultiTemplateBoundTemplateSetDetailsResp)

	templateIDs, err := tools.GetUint32List(req.TemplateIds)
	if err != nil {
		return nil, fmt.Errorf("invalid template ids, %s", err)
	}
	idsLen := len(templateIDs)
	if idsLen == 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListMultiTemplateBoundTemplateSetDetailsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateIds:     templateIDs,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListMultiTemplateBoundTemplateSetDetails(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list multiple template bound template set details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListMultiTemplateBoundTemplateSetDetailsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTemplateRevisionBoundUnnamedAppDetails list template revision bound unnamed app details
func (s *Service) ListTemplateRevisionBoundUnnamedAppDetails(ctx context.Context,
	req *pbcs.ListTemplateRevisionBoundUnnamedAppDetailsReq) (
	*pbcs.ListTemplateRevisionBoundUnnamedAppDetailsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateRevisionBoundUnnamedAppDetailsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateRevisionBoundUnnamedAppDetailsReq{
		BizId:              req.BizId,
		TemplateSpaceId:    req.TemplateSpaceId,
		TemplateId:         req.TemplateId,
		TemplateRevisionId: req.TemplateRevisionId,
		Start:              req.Start,
		Limit:              req.Limit,
		All:                req.All,
	}

	rp, err := s.client.DS.ListTemplateRevisionBoundUnnamedAppDetails(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template revision bound unnamed app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateRevisionBoundUnnamedAppDetailsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTemplateRevisionBoundNamedAppDetails list template revision bound named app details
func (s *Service) ListTemplateRevisionBoundNamedAppDetails(ctx context.Context,
	req *pbcs.ListTemplateRevisionBoundNamedAppDetailsReq) (
	*pbcs.ListTemplateRevisionBoundNamedAppDetailsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateRevisionBoundNamedAppDetailsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateRevisionBoundNamedAppDetailsReq{
		BizId:              req.BizId,
		TemplateSpaceId:    req.TemplateSpaceId,
		TemplateId:         req.TemplateId,
		TemplateRevisionId: req.TemplateRevisionId,
		Start:              req.Start,
		Limit:              req.Limit,
		All:                req.All,
	}

	rp, err := s.client.DS.ListTemplateRevisionBoundNamedAppDetails(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template revision bound named app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateRevisionBoundNamedAppDetailsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTemplateSetBoundUnnamedAppDetails list template set bound unnamed app details
func (s *Service) ListTemplateSetBoundUnnamedAppDetails(ctx context.Context,
	req *pbcs.ListTemplateSetBoundUnnamedAppDetailsReq) (
	*pbcs.ListTemplateSetBoundUnnamedAppDetailsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateSetBoundUnnamedAppDetailsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateSetBoundUnnamedAppDetailsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateSetId:   req.TemplateSetId,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTemplateSetBoundUnnamedAppDetails(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template set bound unnamed app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateSetBoundUnnamedAppDetailsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListMultiTemplateSetBoundUnnamedAppDetails list multiple template sets bound unnamed app details
func (s *Service) ListMultiTemplateSetBoundUnnamedAppDetails(ctx context.Context,
	req *pbcs.ListMultiTemplateSetBoundUnnamedAppDetailsReq) (
	*pbcs.ListMultiTemplateSetBoundUnnamedAppDetailsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListMultiTemplateSetBoundUnnamedAppDetailsResp)

	templateSetIDs, err := tools.GetUint32List(req.TemplateSetIds)
	if err != nil {
		return nil, fmt.Errorf("invalid template set ids, %s", err)
	}
	idsLen := len(templateSetIDs)
	if idsLen == 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template set ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListMultiTemplateSetBoundUnnamedAppDetailsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateSetIds:  templateSetIDs,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListMultiTemplateSetBoundUnnamedAppDetails(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template set bound unnamed app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListMultiTemplateSetBoundUnnamedAppDetailsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTemplateSetBoundNamedAppDetails list template set bound named app details
func (s *Service) ListTemplateSetBoundNamedAppDetails(ctx context.Context,
	req *pbcs.ListTemplateSetBoundNamedAppDetailsReq) (
	*pbcs.ListTemplateSetBoundNamedAppDetailsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListTemplateSetBoundNamedAppDetailsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListTemplateSetBoundNamedAppDetailsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateSetId:   req.TemplateSetId,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTemplateSetBoundNamedAppDetails(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template set bound named app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListTemplateSetBoundNamedAppDetailsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListLatestTemplateBoundUnnamedAppDetails list the latest template bound unnamed app details
func (s *Service) ListLatestTemplateBoundUnnamedAppDetails(ctx context.Context,
	req *pbcs.ListLatestTemplateBoundUnnamedAppDetailsReq) (
	*pbcs.ListLatestTemplateBoundUnnamedAppDetailsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListLatestTemplateBoundUnnamedAppDetailsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Template, Action: meta.Find}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListLatestTemplateBoundUnnamedAppDetailsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateId:      req.TemplateId,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListLatestTemplateBoundUnnamedAppDetails(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list the latest template bound unnamed app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListLatestTemplateBoundUnnamedAppDetailsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
