package service

import (
	"context"
	"time"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcrs "bscp.io/pkg/protocol/core/credential-scope"
	pbds "bscp.io/pkg/protocol/data-service"
)

// CreateCredentialScope
func (s *Service) CreateCredentialScope(ctx context.Context, req *pbds.CreateCredentialScopeReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	now := time.Now()
	for _, value := range req.Spec {
		credentialScope := &table.CredentialScope{
			Spec: &table.CredentialScopeSpec{
				CredentialScope: value,
			},
			Attachment: req.Attachment.CredentialAttachment(),
			Revision: &table.CredentialRevision{
				Creator:   kt.User,
				Reviser:   kt.User,
				CreatedAt: now,
				UpdatedAt: now,
				ExpiredAt: now,
			},
		}
		_, err := s.dao.CredentialScope().Create(kt, credentialScope)
		if err != nil {
			logs.Errorf("create credential scope failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}
	resp := &pbds.CreateResp{}
	return resp, nil
}

// ListCredentialScopes
func (s *Service) ListCredentialScopes(ctx context.Context, req *pbds.ListCredentialScopesReq) (*pbds.ListCredentialScopesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	details, err := s.dao.CredentialScope().Get(kt, req.CredentialId, req.BizId)
	if err != nil {
		logs.Errorf("list credential scope failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	credentialScopes, err := pbcrs.PbCredentialScopes(details.Details)
	if err != nil {
		logs.Errorf("get pb credential scope failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListCredentialScopesResp{
		Count:   details.Count,
		Details: credentialScopes,
	}
	return resp, nil
}

// DeleteCredentialScopes
func (s *Service) DeleteCredentialScopes(ctx context.Context, req *pbds.DeleteCredentialScopesReq) (*pbds.DeleteCredentialScopesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	for _, value := range req.Id {
		credentialScope := &table.CredentialScope{
			ID:         value,
			Attachment: req.Attachment.CredentialAttachment(),
		}
		err := s.dao.CredentialScope().Delete(kt, credentialScope)
		if err != nil {
			logs.Errorf("delete credential scope failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}
	resp := &pbds.DeleteCredentialScopesResp{}
	return resp, nil
}
