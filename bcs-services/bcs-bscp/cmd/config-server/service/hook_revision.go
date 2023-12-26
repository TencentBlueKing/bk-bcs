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
	"errors"

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbhr "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/hook-revision"
	pbds "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// CreateHookRevision create hook revision with option
func (s *Service) CreateHookRevision(ctx context.Context,
	req *pbcs.CreateHookRevisionReq) (*pbcs.CreateHookRevisionResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.CreateHookRevisionReq{
		Attachment: &pbhr.HookRevisionAttachment{
			BizId:  grpcKit.BizID,
			HookId: req.HookId,
		},
		Spec: &pbhr.HookRevisionSpec{
			Name:    tools.GenerateRevisionName(),
			Content: req.Content,
			Memo:    req.Memo,
		},
	}

	rp, err := s.client.DS.CreateHookRevision(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create HookRevision failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.CreateHookRevisionResp{
		Id: rp.Id,
	}
	return resp, nil
}

// ListHookRevisions list hook revisions with filter
func (s *Service) ListHookRevisions(ctx context.Context, req *pbcs.ListHookRevisionsReq) (
	*pbcs.ListHookRevisionsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListHookRevisionsReq{
		HookId:    req.HookId,
		SearchKey: req.SearchKey,
		BizId:     grpcKit.BizID,
		All:       req.All,
		State:     req.State,
	}

	if !req.All {
		if req.Limit == 0 {
			return nil, errors.New("limit has to be greater than 0")
		}
		r.Start = req.Start
		r.Limit = req.Limit
	}

	rp, err := s.client.DS.ListHookRevisions(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list HookRevisions failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	details := make([]*pbcs.ListHookRevisionsResp_ListHookRevisionsData, 0, len(rp.Details))

	for _, detail := range rp.Details {
		details = append(details, &pbcs.ListHookRevisionsResp_ListHookRevisionsData{
			HookRevision:  detail.HookRevision,
			BoundNum:      detail.BoundNum,
			ConfirmDelete: detail.ConfirmDelete,
		})
	}

	resp := &pbcs.ListHookRevisionsResp{
		Count:   rp.Count,
		Details: details,
	}

	return resp, nil
}

// DeleteHookRevision delete a HookRevision
func (s *Service) DeleteHookRevision(ctx context.Context,
	req *pbcs.DeleteHookRevisionReq) (*pbcs.DeleteHookRevisionResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteHookRevisionResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.DeleteHookRevisionReq{
		BizId:      req.BizId,
		HookId:     req.HookId,
		RevisionId: req.RevisionId,
		Force:      req.Force,
	}

	if _, err := s.client.DS.DeleteHookRevision(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete HookRevision failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// PublishHookRevision publish a revision
func (s *Service) PublishHookRevision(ctx context.Context, req *pbcs.
	PublishHookRevisionReq) (*pbcs.PublishHookRevisionResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.PublishHookRevisionResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.PublishHookRevisionReq{
		BizId:  req.BizId,
		HookId: req.HookId,
		Id:     req.RevisionId,
	}

	if _, err := s.client.DS.PublishHookRevision(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("publish HookRevision failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// GetHookRevision get a hookRevision
func (s *Service) GetHookRevision(ctx context.Context, req *pbcs.GetHookRevisionReq) (*pbhr.HookRevision, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.GetHookRevisionByIdReq{
		BizId:  req.BizId,
		HookId: req.HookId,
		Id:     req.RevisionId,
	}

	return s.client.DS.GetHookRevisionByID(grpcKit.RpcCtx(), r)
}

// UpdateHookRevision update a HookRevision
func (s *Service) UpdateHookRevision(ctx context.Context, req *pbcs.UpdateHookRevisionReq) (
	*pbcs.UpdateHookRevisionResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateHookRevisionResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.UpdateHookRevisionReq{
		Id: req.RevisionId,
		Attachment: &pbhr.HookRevisionAttachment{
			BizId:  req.BizId,
			HookId: req.HookId,
		},
		Spec: &pbhr.HookRevisionSpec{
			Name:    req.Name,
			Content: req.Content,
			Memo:    req.Memo,
		},
	}
	if _, err := s.client.DS.UpdateHookRevision(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update HookRevision failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil

}

// ListHookRevisionReferences 查询hook版本被引用列表
func (s *Service) ListHookRevisionReferences(ctx context.Context,
	req *pbcs.ListHookRevisionReferencesReq) (*pbcs.ListHookRevisionReferencesResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
	}
	if err := s.authorizer.Authorize(grpcKit, res...); err != nil {
		return nil, err
	}

	r := &pbds.ListHookRevisionReferencesReq{
		BizId:      req.BizId,
		HookId:     req.HookId,
		RevisionId: req.RevisionId,
		Limit:      req.Limit,
		Start:      req.Start,
		SearchKey:  req.SearchKey,
	}

	rp, err := s.client.DS.ListHookRevisionReferences(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list TemplateSpaces failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	details := []*pbcs.ListHookRevisionReferencesResp_Detail{}
	for _, detail := range rp.Details {
		details = append(details, &pbcs.ListHookRevisionReferencesResp_Detail{
			RevisionId:   detail.RevisionId,
			RevisionName: detail.RevisionName,
			AppId:        detail.AppId,
			AppName:      detail.AppName,
			ReleaseId:    detail.ReleaseId,
			ReleaseName:  detail.ReleaseName,
			Type:         detail.Type,
		})
	}

	resp := &pbcs.ListHookRevisionReferencesResp{
		Count:   rp.Count,
		Details: details,
	}
	return resp, nil
}
