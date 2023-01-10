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

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbinstance "bscp.io/pkg/protocol/core/instance"
	pbrelease "bscp.io/pkg/protocol/core/release"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// PublishInstance publish an instance.
func (s *Service) PublishInstance(ctx context.Context, req *pbcs.PublishInstanceReq) (
	*pbcs.PublishInstanceResp, error) {

	kit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.PublishInstanceResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.CRInstance, Action: meta.Create,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kit, resp, authRes)
	if err != nil {
		return resp, nil
	}

	r := &pbds.CreateCRInstanceReq{
		Attachment: &pbrelease.ReleaseAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
		Spec: &pbinstance.ReleasedInstanceSpec{
			Uid:       req.Uid,
			ReleaseId: req.ReleaseId,
			Memo:      req.Memo,
		},
	}
	rp, err := s.client.DS.CreateCRInstance(kit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kit, resp)
		logs.Errorf("create current released instance failed, err: %v, rid: %s", err, kit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.PublishInstanceResp_RespData{
		Id: rp.Id,
	}
	return resp, nil
}

// DeletePublishedInstance delete a published instance
func (s *Service) DeletePublishedInstance(ctx context.Context, req *pbcs.DeletePublishedInstanceReq) (
	*pbcs.DeletePublishedInstanceResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeletePublishedInstanceResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.CRInstance, Action: meta.Delete,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, authRes)
	if err != nil {
		return resp, nil
	}

	r := &pbds.DeleteCRInstanceReq{
		Id: req.Id,
		Attachment: &pbrelease.ReleaseAttachment{
			BizId: req.BizId,
			AppId: req.AppId,
		},
	}
	_, err = s.client.DS.DeleteCRInstance(grpcKit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("delete current released instance failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	return resp, nil
}

// ListPublishedInstance list the published instance.
func (s *Service) ListPublishedInstance(ctx context.Context, req *pbcs.ListPublishedInstanceReq) (
	*pbcs.ListPublishedInstanceResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListPublishedInstanceResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.CRInstance, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, authRes)
	if err != nil {
		return resp, nil
	}

	if req.Page == nil {
		errf.Error(errf.New(errf.InvalidParameter, "page is null")).AssignResp(grpcKit, resp)
		return resp, nil
	}

	if err := req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		return resp, nil
	}

	r := &pbds.ListCRInstancesReq{
		BizId:  req.BizId,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListCRInstances(grpcKit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(grpcKit, resp)
		logs.Errorf("list current released instance failed, err: %v, rid: %s", err, grpcKit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.ListPublishedInstanceResp_RespData{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
