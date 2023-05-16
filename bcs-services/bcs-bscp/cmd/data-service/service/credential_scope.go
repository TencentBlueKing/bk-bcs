package service

import (
	"context"
	"time"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbcrs "bscp.io/pkg/protocol/core/credential-scope"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/runtime/credential"
)

// ListCredentialScopes  get credential scopes
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

// UpdateCredentialScopes update credential scopes
func (s *Service) UpdateCredentialScopes(ctx context.Context, req *pbds.UpdateCredentialScopesReq) (*pbds.UpdateCredentialScopesResp, error) {
	kt := kit.FromGrpcContext(ctx)

	tx, err := s.dao.BeginTx(kt, req.BizId)
	if err != nil {
		logs.Errorf("begin transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	now := time.Now()
	for _, updated := range req.Updated {
		credentialScope := &table.CredentialScope{
			ID: updated.Id,
			Spec: &table.CredentialScopeSpec{
				CredentialScope: credential.CredentialScope(updated.Scope),
				ExpiredAt:       time.Now(),
			},
			Attachment: &table.CredentialScopeAttachment{
				BizID:        req.BizId,
				CredentialId: updated.Id,
			},
			Revision: &table.Revision{
				Reviser:   kt.User,
				UpdatedAt: now,
			},
		}
		if err = s.dao.CredentialScope().UpdateWithTx(kt, tx, credentialScope); err != nil {
			logs.Errorf("update credential scope failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback(kt)
			return nil, err
		}
	}
	for _, deleted := range req.Deleted {
		if err = s.dao.CredentialScope().DeleteWithTx(kt, tx, req.BizId, deleted); err != nil {
			logs.Errorf("delete credential scope failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback(kt)
			return nil, err
		}
	}

	for _, created := range req.Created {
		credentialScope := &table.CredentialScope{
			Spec: &table.CredentialScopeSpec{
				CredentialScope: credential.CredentialScope(created),
				ExpiredAt:       time.Now(),
			},
			Attachment: &table.CredentialScopeAttachment{
				BizID:        req.BizId,
				CredentialId: req.CredentialId,
			},
			Revision: &table.Revision{
				Creator:   kt.User,
				Reviser:   kt.User,
				CreatedAt: now,
				UpdatedAt: now,
			},
		}
		if _, err = s.dao.CredentialScope().CreateWithTx(kt, tx, credentialScope); err != nil {
			logs.Errorf("create credential scope failed, err: %v, rid: %s", err, kt.Rid)
			tx.Rollback(kt)
			return nil, err
		}
	}

	if err := s.dao.Credential().UpdateRevisionWithTx(kt, tx, req.BizId, req.CredentialId); err != nil {
		logs.Errorf("update credential revision failed, err: %v, rid: %s", err, kt.Rid)
		tx.Rollback(kt)
		return nil, err
	}
	tx.Commit(kt)
	resp := &pbds.UpdateCredentialScopesResp{}
	return resp, nil
}
