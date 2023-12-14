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

	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbds "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
)

// ListCredentialScopes get credential scopes
func (s *Service) ListCredentialScopes(ctx context.Context,
	req *pbcs.ListCredentialScopesReq) (*pbcs.ListCredentialScopesResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.Credential, Action: meta.View}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.ListCredentialScopesReq{
		BizId:        req.BizId,
		CredentialId: req.CredentialId,
	}
	rp, err := s.client.DS.ListCredentialScopes(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list credential scope failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.ListCredentialScopesResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil

}

// UpdateCredentialScope  update credential scope
func (s *Service) UpdateCredentialScope(ctx context.Context,
	req *pbcs.UpdateCredentialScopeReq) (*pbcs.UpdateCredentialScopeResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	resp := new(pbcs.UpdateCredentialScopeResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.Credential, Action: meta.Manage}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.UpdateCredentialScopesReq{
		BizId:        req.BizId,
		CredentialId: req.CredentialId,
	}

	r.Created = append(r.Created, req.AddScope...)
	r.Updated = append(r.Updated, req.AlterScope...)
	r.Deleted = append(r.Deleted, req.DelId...)

	_, err = s.client.DS.UpdateCredentialScopes(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("update credential scope failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}
