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

package service

import (
	"context"
	"fmt"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/constant"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbds "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// ListTmplBoundCounts list template bound counts
func (s *Service) ListTmplBoundCounts(ctx context.Context, req *pbcs.ListTmplBoundCountsReq) (
	*pbcs.ListTmplBoundCountsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

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

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTmplBoundCountsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateIds:     req.TemplateIds,
	}

	rp, err := s.client.DS.ListTmplBoundCounts(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template bound counts failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTmplBoundCountsResp{
		Details: rp.Details,
	}
	return resp, nil
}

// ListTmplRevisionBoundCounts list template bound counts
func (s *Service) ListTmplRevisionBoundCounts(ctx context.Context, req *pbcs.ListTmplRevisionBoundCountsReq) (
	*pbcs.ListTmplRevisionBoundCountsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

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

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTmplRevisionBoundCountsReq{
		BizId:               req.BizId,
		TemplateSpaceId:     req.TemplateSpaceId,
		TemplateId:          req.TemplateId,
		TemplateRevisionIds: req.TemplateRevisionIds,
	}

	rp, err := s.client.DS.ListTmplRevisionBoundCounts(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template revision bound counts failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTmplRevisionBoundCountsResp{
		Details: rp.Details,
	}
	return resp, nil
}

// ListTmplSetBoundCounts list template bound counts
func (s *Service) ListTmplSetBoundCounts(ctx context.Context, req *pbcs.ListTmplSetBoundCountsReq) (
	*pbcs.ListTmplSetBoundCountsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

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

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTmplSetBoundCountsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateSetIds:  req.TemplateSetIds,
	}

	rp, err := s.client.DS.ListTmplSetBoundCounts(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template set bound counts failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTmplSetBoundCountsResp{
		Details: rp.Details,
	}
	return resp, nil
}

// ListTmplBoundUnnamedApps list template bound unnamed app details
func (s *Service) ListTmplBoundUnnamedApps(ctx context.Context, req *pbcs.ListTmplBoundUnnamedAppsReq) (
	*pbcs.ListTmplBoundUnnamedAppsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTmplBoundUnnamedAppsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateId:      req.TemplateId,
		SearchFields:    req.SearchFields,
		SearchValue:     req.SearchValue,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTmplBoundUnnamedApps(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template bound unnamed app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTmplBoundUnnamedAppsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTmplBoundNamedApps list template bound named app details
// Deprecated: not in use currently
// if use it, consider to add column app_name, release_name on table released_app_templates in case of app is deleted
func (s *Service) ListTmplBoundNamedApps(ctx context.Context, req *pbcs.ListTmplBoundNamedAppsReq) (
	*pbcs.ListTmplBoundNamedAppsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTmplBoundNamedAppsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateId:      req.TemplateId,
		SearchFields:    req.SearchFields,
		SearchValue:     req.SearchValue,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTmplBoundNamedApps(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template bound named app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTmplBoundNamedAppsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTmplBoundTmplSets list template bound template set details
func (s *Service) ListTmplBoundTmplSets(ctx context.Context, req *pbcs.ListTmplBoundTmplSetsReq) (
	*pbcs.ListTmplBoundTmplSetsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTmplBoundTmplSetsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateId:      req.TemplateId,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTmplBoundTmplSets(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template bound template set details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTmplBoundTmplSetsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListMultiTmplBoundTmplSets list multiple template bound template set details
func (s *Service) ListMultiTmplBoundTmplSets(ctx context.Context, req *pbcs.ListMultiTmplBoundTmplSetsReq) (
	*pbcs.ListMultiTmplBoundTmplSetsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// validate input param
	templateIDs, err := tools.GetUint32List(req.TemplateIds)
	if err != nil {
		return nil, fmt.Errorf("invalid template ids, %s", err)
	}
	idsLen := len(templateIDs)
	if idsLen == 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if e := s.authorizer.Authorize(grpcKit, res...); e != nil {
		return nil, e
	}

	r := &pbds.ListMultiTmplBoundTmplSetsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateIds:     templateIDs,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListMultiTmplBoundTmplSets(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list multiple template bound template set details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListMultiTmplBoundTmplSetsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTmplRevisionBoundUnnamedApps list template revision bound unnamed app details
func (s *Service) ListTmplRevisionBoundUnnamedApps(ctx context.Context, req *pbcs.ListTmplRevisionBoundUnnamedAppsReq) (
	*pbcs.ListTmplRevisionBoundUnnamedAppsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTmplRevisionBoundUnnamedAppsReq{
		BizId:              req.BizId,
		TemplateSpaceId:    req.TemplateSpaceId,
		TemplateId:         req.TemplateId,
		TemplateRevisionId: req.TemplateRevisionId,
		SearchFields:       req.SearchFields,
		SearchValue:        req.SearchValue,
		Start:              req.Start,
		Limit:              req.Limit,
		All:                req.All,
	}

	rp, err := s.client.DS.ListTmplRevisionBoundUnnamedApps(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template revision bound unnamed app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTmplRevisionBoundUnnamedAppsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTmplRevisionBoundNamedApps list template revision bound named app details
// Deprecated: not in use currently
// if use it, consider to add column app_name, release_name on table released_app_templates in case of app is deleted
func (s *Service) ListTmplRevisionBoundNamedApps(ctx context.Context,
	req *pbcs.ListTmplRevisionBoundNamedAppsReq) (*pbcs.ListTmplRevisionBoundNamedAppsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTmplRevisionBoundNamedAppsReq{
		BizId:              req.BizId,
		TemplateSpaceId:    req.TemplateSpaceId,
		TemplateId:         req.TemplateId,
		TemplateRevisionId: req.TemplateRevisionId,
		SearchFields:       req.SearchFields,
		SearchValue:        req.SearchValue,
		Start:              req.Start,
		Limit:              req.Limit,
		All:                req.All,
	}

	rp, err := s.client.DS.ListTmplRevisionBoundNamedApps(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template revision bound named app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTmplRevisionBoundNamedAppsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTmplSetBoundUnnamedApps list template set bound unnamed app details
func (s *Service) ListTmplSetBoundUnnamedApps(ctx context.Context, req *pbcs.ListTmplSetBoundUnnamedAppsReq) (
	*pbcs.ListTmplSetBoundUnnamedAppsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTmplSetBoundUnnamedAppsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateSetId:   req.TemplateSetId,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTmplSetBoundUnnamedApps(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template set bound unnamed app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTmplSetBoundUnnamedAppsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListMultiTmplSetBoundUnnamedApps list multiple template sets bound unnamed app details
func (s *Service) ListMultiTmplSetBoundUnnamedApps(ctx context.Context, req *pbcs.ListMultiTmplSetBoundUnnamedAppsReq) (
	*pbcs.ListMultiTmplSetBoundUnnamedAppsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	// validate input param
	templateSetIDs, err := tools.GetUint32List(req.TemplateSetIds)
	if err != nil {
		return nil, fmt.Errorf("invalid template set ids, %s", err)
	}
	idsLen := len(templateSetIDs)
	if idsLen == 0 || idsLen > constant.ArrayInputLenLimit {
		return nil, fmt.Errorf("the length of template set ids is %d, it must be within the range of [1,%d]",
			idsLen, constant.ArrayInputLenLimit)
	}

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if e := s.authorizer.Authorize(grpcKit, res...); e != nil {
		return nil, e
	}

	r := &pbds.ListMultiTmplSetBoundUnnamedAppsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateSetIds:  templateSetIDs,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListMultiTmplSetBoundUnnamedApps(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template set bound unnamed app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListMultiTmplSetBoundUnnamedAppsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListTmplSetBoundNamedApps list template set bound named app details
// Deprecated: not in use currently
// if use it, consider to add column app_name, release_name on table released_app_templates in case of app is deleted
func (s *Service) ListTmplSetBoundNamedApps(ctx context.Context, req *pbcs.ListTmplSetBoundNamedAppsReq) (
	*pbcs.ListTmplSetBoundNamedAppsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListTmplSetBoundNamedAppsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateSetId:   req.TemplateSetId,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListTmplSetBoundNamedApps(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list template set bound named app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListTmplSetBoundNamedAppsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// ListLatestTmplBoundUnnamedApps list the latest template bound unnamed app details
func (s *Service) ListLatestTmplBoundUnnamedApps(ctx context.Context, req *pbcs.ListLatestTmplBoundUnnamedAppsReq) (
	*pbcs.ListLatestTmplBoundUnnamedAppsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListLatestTmplBoundUnnamedAppsReq{
		BizId:           req.BizId,
		TemplateSpaceId: req.TemplateSpaceId,
		TemplateId:      req.TemplateId,
		Start:           req.Start,
		Limit:           req.Limit,
		All:             req.All,
	}

	rp, err := s.client.DS.ListLatestTmplBoundUnnamedApps(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list the latest template bound unnamed app details failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListLatestTmplBoundUnnamedAppsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
