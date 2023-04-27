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
		BizId: req.BizId,
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
