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
	"errors"
	"fmt"

	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbatb "bscp.io/pkg/protocol/core/app-template-binding"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// CreateAppTemplateBinding create a app template binding
func (s *Service) CreateAppTemplateBinding(ctx context.Context, req *pbcs.CreateAppTemplateBindingReq) (*pbcs.
	CreateAppTemplateBindingResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateAppTemplateBindingResp)

	templateSetIDs, templateRevisionIDs, err := parseBindings(req.Bindings)
	if err != nil {
		logs.Errorf("create app template binding failed, parse bindings err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	repeatedTmplSetIDs := tools.SliceRepeatedElements(templateSetIDs)
	if len(repeatedTmplSetIDs) > 0 {
		return nil, fmt.Errorf("repeated template set ids: %v, id must be unique", repeatedTmplSetIDs)
	}
	repeatedTmplRevisiondIDs := tools.SliceRepeatedElements(templateRevisionIDs)
	if len(repeatedTmplRevisiondIDs) > 0 {
		return nil, fmt.Errorf("repeated template revision ids: %v, id must be unique", repeatedTmplRevisiondIDs)
	}
	if len(templateRevisionIDs) > 500 {
		return nil, fmt.Errorf("the length of template revision ids is %d, it must be within the range of [1,500]",
			len(templateRevisionIDs))
	}

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AppTemplateBinding, Action: meta.Create,
		ResourceID: req.AppId}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.CreateAppTemplateBindingReq{
		Attachment: &pbatb.AppTemplateBindingAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbatb.AppTemplateBindingSpec{
			Bindings: req.Bindings,
		},
	}
	rp, err := s.client.DS.CreateAppTemplateBinding(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create app template binding failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.CreateAppTemplateBindingResp{
		Id: rp.Id,
	}
	return resp, nil
}

// DeleteAppTemplateBinding delete a app template binding
func (s *Service) DeleteAppTemplateBinding(ctx context.Context, req *pbcs.DeleteAppTemplateBindingReq) (*pbcs.
	DeleteAppTemplateBindingResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteAppTemplateBindingResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AppTemplateBinding, Action: meta.Delete,
		ResourceID: req.BindingId}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.DeleteAppTemplateBindingReq{
		Id: req.BindingId,
		Attachment: &pbatb.AppTemplateBindingAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	}
	if _, err := s.client.DS.DeleteAppTemplateBinding(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("delete app template binding failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// UpdateAppTemplateBinding update a app template binding
func (s *Service) UpdateAppTemplateBinding(ctx context.Context, req *pbcs.UpdateAppTemplateBindingReq) (*pbcs.
	UpdateAppTemplateBindingResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateAppTemplateBindingResp)

	templateSetIDs, templateRevisionIDs, err := parseBindings(req.Bindings)
	if err != nil {
		logs.Errorf("create app template binding failed, parse bindings err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}
	repeatedTmplSetIDs := tools.SliceRepeatedElements(templateSetIDs)
	if len(repeatedTmplSetIDs) > 0 {
		return nil, fmt.Errorf("repeated template set ids: %v, id must be unique", repeatedTmplSetIDs)
	}
	repeatedTmplRevisiondIDs := tools.SliceRepeatedElements(templateRevisionIDs)
	if len(repeatedTmplRevisiondIDs) > 0 {
		return nil, fmt.Errorf("repeated template revision ids: %v, id must be unique", repeatedTmplRevisiondIDs)
	}
	if len(templateRevisionIDs) > 500 {
		return nil, fmt.Errorf("the length of template revision ids is %d, it must be within the range of [1,500]",
			len(templateRevisionIDs))
	}

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AppTemplateBinding, Action: meta.Update,
		ResourceID: req.BindingId}, BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.UpdateAppTemplateBindingReq{
		Id: req.BindingId,
		Attachment: &pbatb.AppTemplateBindingAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbatb.AppTemplateBindingSpec{
			Bindings: req.Bindings,
		},
	}
	if _, err := s.client.DS.UpdateAppTemplateBinding(grpcKit.RpcCtx(), r); err != nil {
		logs.Errorf("update app template binding failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// ListAppTemplateBindings list app template bindings
func (s *Service) ListAppTemplateBindings(ctx context.Context, req *pbcs.ListAppTemplateBindingsReq) (*pbcs.
	ListAppTemplateBindingsResp,
	error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListAppTemplateBindingsResp)

	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AppTemplateBinding, Action: meta.Find},
		BizID: req.BizId}
	if err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res); err != nil {
		return nil, err
	}

	r := &pbds.ListAppTemplateBindingsReq{
		BizId: req.BizId,
		AppId: req.AppId,
		Start: 0,
		Limit: uint32(types.DefaultMaxPageLimit),
	}

	rp, err := s.client.DS.ListAppTemplateBindings(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list app template bindings failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListAppTemplateBindingsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

func parseBindings(bindings []*pbatb.TemplateBinding) (templateSetIDs, templateRevisionIDs []uint32, err error) {
	if len(bindings) == 0 {
		return nil, nil, errors.New("bindings can't be empty")
	}
	for _, b := range bindings {
		if b.TemplateSetId <= 0 {
			return nil, nil, fmt.Errorf("invalid template set id of bindings member: %d", b.TemplateSetId)
		}
		if len(b.TemplateRevisionIds) == 0 {
			return nil, nil, errors.New("template revision ids of bindings member can't be empty")
		}
		templateSetIDs = append(templateSetIDs, b.TemplateSetId)
		templateRevisionIDs = append(templateRevisionIDs, b.TemplateRevisionIds...)
	}

	return templateSetIDs, templateRevisionIDs, nil
}
