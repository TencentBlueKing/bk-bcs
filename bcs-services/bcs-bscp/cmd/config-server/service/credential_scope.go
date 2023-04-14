package service

import (
	"context"

	"bscp.io/pkg/iam/meta"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcs "bscp.io/pkg/protocol/config-server"
	pbcredential "bscp.io/pkg/protocol/core/credential"
	pbcrs "bscp.io/pkg/protocol/core/credential-scope"
	pbds "bscp.io/pkg/protocol/data-service"
)

// ListCredentialScopes
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

// UpdateCredentialScope
func (s *Service) UpdateCredentialScope(ctx context.Context, req *pbcs.UpdateCredentialScopeReq) (*pbcs.UpdateCredentialScopeResp, error) {
	grpcKit := kit.FromGrpcContext(ctx)
	resp := new(pbcs.UpdateCredentialScopeResp)

	bizID := req.BizId
	res := &meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.CredentialScope, Action: meta.Update}, BizID: bizID}
	err := s.authorizer.AuthorizeWithResp(grpcKit, resp, res)
	if err != nil {
		return nil, err
	}

	// delete Credential_scopes
	r := &pbds.DeleteCredentialScopesReq{
		Id: req.DelId,
		Attachment: &pbcrs.CredentialScopeAttachment{
			BizId:        bizID,
			CredentialId: req.CredentialId,
		},
	}

	_, err = s.client.DS.DeleteCredentialScopes(grpcKit.RpcCtx(), r)
	if err != nil {
		logs.Errorf("create credential scope failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// create Credential_scopes
	rs := &pbds.CreateCredentialScopeReq{
		Spec: req.AddScope,
		Attachment: &pbcrs.CredentialScopeAttachment{
			BizId:        bizID,
			CredentialId: req.CredentialId,
		},
	}

	_, err = s.client.DS.CreateCredentialScope(grpcKit.RpcCtx(), rs)
	if err != nil {
		logs.Errorf("create credential scope failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	// update credential
	rc := &pbds.UpdateCredentialReq{
		Id:   req.CredentialId,
		Spec: &pbcredential.CredentialSpec{},
		Attachment: &pbcredential.CredentialAttachment{
			BizId: bizID,
		},
	}
	_, err = s.client.DS.UpdateCredential(grpcKit.RpcCtx(), rc)
	if err != nil {
		logs.Errorf("update credential failed, err: %v, rid: %s", err, grpcKit.Rid)
		return nil, err
	}

	return resp, nil
}
