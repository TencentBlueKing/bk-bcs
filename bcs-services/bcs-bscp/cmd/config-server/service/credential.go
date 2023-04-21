package service

import (
	"context"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbcredential "bscp.io/pkg/protocol/core/credential"
	pbcrs "bscp.io/pkg/protocol/core/credential-scope"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/filter"
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

	// create Credential_scopes
	rs := &pbds.CreateCredentialScopeReq{
		Spec: req.Scope,
		Attachment: &pbcrs.CredentialScopeAttachment{
			BizId:        bizID,
			CredentialId: rp.Id,
		},
	}
	_, err = s.client.DS.CreateCredentialScope(grpcKit.RpcCtx(), rs)
	if err != nil {
		logs.Errorf("create credential scope failed, err: %v, rid: %s", err, grpcKit.Rid)
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

	page := &pbbase.BasePage{
		Start: req.Start,
		Limit: req.Limit,
	}

	ft := &filter.Expression{
		Op:    filter.Or,
		Rules: []filter.RuleFactory{},
	}

	if req.SearchKey != "" {
		ft.Rules = append(ft.Rules, &filter.AtomRule{
			Field: "enable",
			Op:    filter.ContainsInsensitive.Factory(),
			Value: req.SearchKey,
		}, &filter.AtomRule{
			Field: "memo",
			Op:    filter.ContainsInsensitive.Factory(),
			Value: req.SearchKey,
		}, &filter.AtomRule{
			Field: "creator",
			Op:    filter.ContainsInsensitive.Factory(),
			Value: req.SearchKey,
		})
	}

	ftpb, err := ft.MarshalPB()
	if err != nil {
		return nil, err
	}
	r := &pbds.ListCredentialReq{
		BizId:  bizID,
		Page:   page,
		Filter: ftpb,
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
