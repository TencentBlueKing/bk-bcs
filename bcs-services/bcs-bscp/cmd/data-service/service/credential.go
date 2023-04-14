package service

import (
	"context"
	"time"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbcredential "bscp.io/pkg/protocol/core/credential"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateCredential
func (s *Service) CreateCredential(ctx context.Context, req *pbds.CreateCredentialReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	spec, err := req.Spec.CredentialSpec()
	if err != nil {
		logs.Errorf("get credential spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	now := time.Now()
	credential := &table.Credential{
		Spec:       spec,
		Attachment: req.Attachment.CredentialAttachment(),
		Revision: &table.CredentialRevision{
			Creator:   kt.User,
			Reviser:   kt.User,
			CreatedAt: now,
			UpdatedAt: now,
			ExpiredAt: now,
		},
	}
	id, err := s.dao.Credential().Create(kt, credential)
	if err != nil {
		logs.Errorf("create credential failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil

}

// ListCredentials
func (s *Service) ListCredentials(ctx context.Context, req *pbds.ListCredentialReq) (*pbds.ListCredentialResp, error) {
	kt := kit.FromGrpcContext(ctx)

	// parse pb struct filter to filter.Expression.
	filter, err := pbbase.UnmarshalFromPbStructToExpr(req.Filter)
	if err != nil {
		logs.Errorf("unmarshal pb struct to expression failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	query := &types.ListCredentialsOption{
		BizID:  req.BizId,
		Page:   req.Page.BasePage(),
		Filter: filter,
	}
	details, err := s.dao.Credential().List(kt, query)
	if err != nil {
		logs.Errorf("list credential failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	credentials, err := pbcredential.PbCredentials(details.Details)
	if err != nil {
		logs.Errorf("get pb credential failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListCredentialResp{
		Count:   details.Count,
		Details: credentials,
	}
	return resp, nil
}

// DeleteCredential
func (s *Service) DeleteCredential(ctx context.Context, req *pbds.DeleteCredentialReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	credential := &table.Credential{
		ID:         req.Id,
		Attachment: req.Attachment.CredentialAttachment(),
	}

	// 查看credential_scopes表中的数据
	credentialScopes, err := s.dao.CredentialScope().Get(kt, req.Id, req.Attachment.BizId)
	if err != nil {
		logs.Errorf("get credential scope failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if credentialScopes.Count != 0 {
		return nil, errf.New(errf.InvalidParameter, "delete Credential failed, credential scope have data")
	}

	if err := s.dao.Credential().Delete(kt, credential); err != nil {
		logs.Errorf("delete credential failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// UpdateCredential
func (s *Service) UpdateCredential(ctx context.Context, req *pbds.UpdateCredentialReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	spec, err := req.Spec.CredentialSpec()
	if err != nil {
		logs.Errorf("get credential spec from pb failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	now := time.Now()
	credential := &table.Credential{
		ID:         req.Id,
		Spec:       spec,
		Attachment: req.Attachment.CredentialAttachment(),
		Revision: &table.CredentialRevision{
			Reviser:   kt.User,
			UpdatedAt: now,
		},
	}
	if err := s.dao.Credential().Update(kt, credential); err != nil {
		logs.Errorf("update credential failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
