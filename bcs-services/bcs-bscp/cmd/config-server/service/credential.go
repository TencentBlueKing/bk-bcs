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

	"bscp.io/pkg/cc"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbcredential "bscp.io/pkg/protocol/core/credential"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/tools"
)

// CreateCredentials create a credential
func (s *Service) CreateCredentials(ctx context.Context, req *pbcs.CreateCredentialReq) (*pbcs.CreateCredentialResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.CreateCredentialResp)

	bizID := req.BizId
	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Credential, Action: meta.Create}, BizID: bizID}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	masterKey := cc.ConfigServer().Credential.MasterKey
	encryptionAlgorithm := cc.ConfigServer().Credential.EncryptionAlgorithm

	//create token
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

	resp = &pbcs.CreateCredentialResp{
		Id: rp.Id,
	}
	return resp, nil
}

// ListCredentials get Credentials
func (s *Service) ListCredentials(ctx context.Context, req *pbcs.ListCredentialsReq) (*pbcs.ListCredentialsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.ListCredentialsResp)

	bizID := req.BizId
	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Credential, Action: meta.Find}, BizID: bizID}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	r := &pbds.ListCredentialReq{
		BizId:     bizID,
		SearchKey: req.SearchKey,
		Start:     req.Start,
		Limit:     req.Limit,
	}

	rp, err := s.client.DS.ListCredentials(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("list credentials failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	for _, val := range rp.Details {
		credential, err := tools.DecryptCredential(val.Spec.EncCredential, cc.ConfigServer().Credential.MasterKey, val.Spec.EncAlgorithm)
		if err != nil {
			logs.Errorf("credentials decrypt failed, err: %v, rid: %s", err, grpcKit.Rid)
			return nil, err
		}
		val.Spec.EncCredential = credential
	}

	resp = &pbcs.ListCredentialsResp{
		Count:   rp.Count,
		Details: rp.Details,
	}
	return resp, nil
}

// DeleteCredential delete Credential
func (s *Service) DeleteCredential(ctx context.Context, req *pbcs.DeleteCredentialsReq) (*pbcs.DeleteCredentialsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.DeleteCredentialsResp)

	bizID := req.BizId
	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Credential, Action: meta.Delete}, BizID: bizID}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
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
func (s *Service) UpdateCredential(ctx context.Context, req *pbcs.UpdateCredentialsReq) (*pbcs.UpdateCredentialsResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateCredentialsResp)

	bizID := req.BizId
	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.Credential, Action: meta.Update}, BizID: bizID}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}
	r := &pbds.UpdateCredentialReq{
		Id: req.Id,
		Attachment: &pbcredential.CredentialAttachment{
			BizId: bizID,
		},
		Spec: &pbcredential.CredentialSpec{
			Enable: req.Enable,
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
