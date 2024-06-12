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

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/iam/meta"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/config-server"
	pbcredential "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/credential"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
)

// CreateCredentials create a credential
func (s *Service) CreateCredentials(ctx context.Context,
	req *pbcs.CreateCredentialReq) (*pbcs.CreateCredentialResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	bizID := req.BizId
	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.Credential, Action: meta.Manage}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	masterKey := cc.ConfigServer().Credential.MasterKey
	encryptionAlgorithm := cc.ConfigServer().Credential.EncryptionAlgorithm

	// create token
	credential, err := tools.CreateCredential(masterKey, encryptionAlgorithm)
	if err != nil {
		return nil, err
	}

	// create Credential
	r := &pbds.CreateCredentialReq{
		Attachment: &pbcredential.CredentialAttachment{
			BizId: bizID,
		},
		Spec: &pbcredential.CredentialSpec{
			Memo:           req.Memo,
			Name:           req.Name,
			CredentialType: table.BearToken.String(),
			Enable:         true,
			EncAlgorithm:   encryptionAlgorithm,
			EncCredential:  credential,
		},
	}

	rp, err := s.client.DS.CreateCredential(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create credential failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	resp := &pbcs.CreateCredentialResp{
		Id: rp.Id,
	}
	return resp, nil
}

// ListCredentials get Credentials
func (s *Service) ListCredentials(ctx context.Context,
	req *pbcs.ListCredentialsReq) (*pbcs.ListCredentialsResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)

	bizID := req.BizId
	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.Credential, Action: meta.View}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	// 为了搜索密钥 把搜索条件加密
	encCredential, _ := tools.EncryptCredential(req.SearchKey, cc.ConfigServer().Credential.MasterKey, tools.AES)
	r := &pbds.ListCredentialReq{
		BizId:         bizID,
		SearchKey:     req.SearchKey,
		Start:         req.Start,
		Limit:         req.Limit,
		TopIds:        req.TopIds,
		All:           req.All,
		EncCredential: encCredential,
		Enable:        req.Enable,
	}

	rp, err := s.client.DS.ListCredentials(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list credentials failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	for _, val := range rp.Details {
		credential, e := tools.DecryptCredential(val.Spec.EncCredential, cc.ConfigServer().Credential.MasterKey,
			val.Spec.EncAlgorithm)
		if e != nil {
			logs.Errorf("credentials decrypt failed, err: %v, rid: %s", e, grpcKit.Rid)
			return nil, e
		}
		val.Spec.EncCredential = credential
	}

	resp := &pbcs.ListCredentialsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// DeleteCredential delete Credential
func (s *Service) DeleteCredential(ctx context.Context,
	req *pbcs.DeleteCredentialsReq) (*pbcs.DeleteCredentialsResp, error) {

	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteCredentialsResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.Credential, Action: meta.Manage}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	r := &pbds.DeleteCredentialReq{
		Id: req.Id,
		Attachment: &pbcredential.CredentialAttachment{
			BizId: req.BizId,
		},
	}
	_, err = s.client.DS.DeleteCredential(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("delete credential failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// UpdateCredential update credential
func (s *Service) UpdateCredential(ctx context.Context,
	req *pbcs.UpdateCredentialsReq) (*pbcs.UpdateCredentialsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	resp := new(pbcs.UpdateCredentialsResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.Credential, Action: meta.Manage}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}
	r := &pbds.UpdateCredentialReq{
		Id: req.Id,
		Attachment: &pbcredential.CredentialAttachment{
			BizId: req.BizId,
		},
		Spec: &pbcredential.CredentialSpec{
			Enable: req.Enable,
			Name:   req.Name,
			Memo:   req.Memo,
		},
	}
	_, err = s.client.DS.UpdateCredential(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("update credential failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}

// CheckCredentialName Check if the credential name exists
func (s *Service) CheckCredentialName(ctx context.Context, req *pbcs.CheckCredentialNameReq) (
	*pbcs.CheckCredentialNameResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.Credential, Action: meta.View}, BizID: req.BizId},
	}

	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	credential, err := s.client.DS.CheckCredentialName(grpcKit.Ctx, &pbds.CheckCredentialNameReq{
		BizId:          req.BizId,
		CredentialName: req.CredentialName,
	})
	if err != nil {
		return nil, err
	}

	return &pbcs.CheckCredentialNameResp{Exist: credential.Exist}, nil
}

// CredentialScopePreview 关联规则预览配置项
func (s *Service) CredentialScopePreview(ctx context.Context, req *pbcs.CredentialScopePreviewReq) (
	*pbcs.CredentialScopePreviewResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)

	resp := new(pbcs.CredentialScopePreviewResp)

	res := []*meta.ResourceAttribute{
		{Basic: meta.Basic{Type: meta.Biz, Action: meta.FindBusinessResource}, BizID: req.BizId},
		{Basic: meta.Basic{Type: meta.Credential, Action: meta.View}, BizID: req.BizId},
	}
	err := s.authorizer.Authorize(grpcKit, res...)
	if err != nil {
		return nil, err
	}

	data, err := s.client.DS.CredentialScopePreview(grpcKit.RpcCtx(), &pbds.CredentialScopePreviewReq{
		BizId:       req.BizId,
		AppName:     req.AppName,
		Scope:       req.Scope,
		Limit:       req.Limit,
		Start:       req.Start,
		SearchValue: req.SearchValue,
	})
	if err != nil {
		return resp, err
	}

	items := make([]*pbcs.CredentialScopePreviewResp_Detail, 0)
	for _, v := range data.Details {
		items = append(items, &pbcs.CredentialScopePreviewResp_Detail{
			Name: v.GetName(),
			Path: v.GetPath(),
		})
	}
	resp.Details = items
	resp.Count = data.Count
	return resp, nil
}
