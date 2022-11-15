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
	"bscp.io/pkg/protocol/config-server"
	"bscp.io/pkg/protocol/core/commit"
	"bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateCommit create commit with options
func (s *Service) CreateCommit(ctx context.Context, req *pbcs.CreateCommitReq) (*pbcs.CreateCommitResp, error) {
	kit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateCommitResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Commit, Action: meta.Create,
		ResourceID: req.AppId}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kit, resp, authRes)
	if err != nil {
		return resp, nil
	}

	r := &pbds.CreateCommitReq{
		Attachment: &pbcommit.CommitAttachment{
			BizId:        req.BizId,
			AppId:        req.AppId,
			ConfigItemId: req.ConfigItemId,
		},
		ContentId: req.ContentId,
		Memo:      req.Memo,
	}
	rp, err := s.client.DS.CreateCommit(kit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kit, resp)
		logs.Errorf("create commit failed, err: %v, rid: %s", err, kit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.CreateCommitResp_RespData{
		Id: rp.Id,
	}
	return resp, nil
}

// ListCommits list commit with filter
func (s *Service) ListCommits(ctx context.Context, req *pbcs.ListCommitsReq) (*pbcs.ListCommitsResp, error) {
	kit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListCommitsResp)

	authRes := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Commit, Action: meta.Find}, BizID: req.BizId}
	err := s.authorizer.AuthorizeWithResp(kit, resp, authRes)
	if err != nil {
		return resp, nil
	}

	if req.Page == nil {
		errf.Error(errf.New(errf.InvalidParameter, "page is null")).AssignResp(kit, resp)
		return resp, nil
	}

	if err := req.Page.BasePage().Validate(types.DefaultPageOption); err != nil {
		errf.Error(err).AssignResp(kit, resp)
		return resp, nil
	}

	r := &pbds.ListCommitsReq{
		BizId:  req.BizId,
		AppId:  req.AppId,
		Filter: req.Filter,
		Page:   req.Page,
	}
	rp, err := s.client.DS.ListCommits(kit.RpcCtx(), r)
	if err != nil {
		errf.Error(err).AssignResp(kit, resp)
		logs.Errorf("list commits failed, err: %v, rid: %s", err, kit.Rid)
		return resp, nil
	}

	resp.Code = errf.OK
	resp.Data = &pbcs.ListCommitsResp_RespData{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}
