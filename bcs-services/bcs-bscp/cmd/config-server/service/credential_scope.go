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

	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbds "bscp.io/pkg/protocol/data-service"
)

// ListCredentialScopes get credential scopes
func (s *Service) ListCredentialScopes(ctx context.Context, req *pbcs.ListCredentialScopesReq) (*pbcs.ListCredentialScopesResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListCredentialScopesResp)

	bizID := req.BizId
	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.CredentialScope, Action: meta.Find}, BizID: bizID}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	r := &pbds.ListCredentialScopesReq{
		BizId:        bizID,
		CredentialId: req.CredentialId,
	}
	rp, err := s.client.DS.ListCredentialScopes(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list credential scope failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp = &pbcs.ListCredentialScopesResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil

}

// UpdateCredentialScope  update credential scope
func (s *Service) UpdateCredentialScope(ctx context.Context, req *pbcs.UpdateCredentialScopeReq) (*pbcs.UpdateCredentialScopeResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateCredentialScopeResp)

	bizID := req.BizId
	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.CredentialScope, Action: meta.Update}, BizID: bizID}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	r := &pbds.UpdateCredentialScopesReq{
		BizId:        req.BizId,
		CredentialId: req.CredentialId,
	}

	for _, add := range req.AddScope {
		r.Created = append(r.Created, add)
	}

	for _, updated := range req.AlterScope {
		r.Updated = append(r.Updated, updated)
	}

	for _, del := range req.DelId {
		r.Deleted = append(r.Deleted, del)
	}

	_, err = s.client.DS.UpdateCredentialScopes(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("update credential scope failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}
