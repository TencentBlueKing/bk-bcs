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
	"errors"
	"time"

	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/logs"
	pbbase "bscp.io/pkg/protocol/core/base"
	pbcredential "bscp.io/pkg/protocol/core/credential"
	pbds "bscp.io/pkg/protocol/data-service"
	"bscp.io/pkg/types"
)

// CreateCredential Create Credential
func (s *Service) CreateCredential(ctx context.Context, req *pbds.CreateCredentialReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	credential := &table.Credential{
		Spec:       req.Spec.CredentialSpec(),
		Attachment: req.Attachment.CredentialAttachment(),
		Revision: &table.Revision{
			Creator: kt.User,
			Reviser: kt.User,
		},
	}
	credential.Spec.ExpiredAt = time.Now().UTC()
	id, err := s.dao.Credential().Create(kt, credential)
	if err != nil {
		logs.Errorf("create credential failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.CreateResp{Id: id}
	return resp, nil

}

// ListCredentials get credentials
func (s *Service) ListCredentials(ctx context.Context, req *pbds.ListCredentialReq) (*pbds.ListCredentialResp, error) {
	kt := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit)}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	details, count, err := s.dao.Credential().List(kt, req.BizId, req.SearchKey, opt)

	if err != nil {
		logs.Errorf("list credential failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListCredentialResp{
		Count:   uint32(count),
		Details: pbcredential.PbCredentials(details),
	}
	return resp, nil
}

// DeleteCredential delete credential
func (s *Service) DeleteCredential(ctx context.Context, req *pbds.DeleteCredentialReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	credential := &table.Credential{
		ID:         req.Id,
		Attachment: req.Attachment.CredentialAttachment(),
	}

	// 查看credential_scopes表中的数据
	_, count, err := s.dao.CredentialScope().Get(kt, req.Id, req.Attachment.BizId)
	if err != nil {
		logs.Errorf("get credential scope failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if count != 0 {
		return nil, errors.New("delete Credential failed, credential scope have data")
	}

	if err := s.dao.Credential().Delete(kt, credential); err != nil {
		logs.Errorf("delete credential failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// UpdateCredential update credential
func (s *Service) UpdateCredential(ctx context.Context, req *pbds.UpdateCredentialReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	credential := &table.Credential{
		ID:         req.Id,
		Spec:       req.Spec.CredentialSpec(),
		Attachment: req.Attachment.CredentialAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}
	if err := s.dao.Credential().Update(kt, credential); err != nil {
		logs.Errorf("update credential failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}
