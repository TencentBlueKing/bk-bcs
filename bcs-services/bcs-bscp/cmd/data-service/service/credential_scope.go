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
	"time"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbcrs "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/credential-scope"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/runtime/credential"
)

// ListCredentialScopes  get credential scopes
func (s *Service) ListCredentialScopes(ctx context.Context,
	req *pbds.ListCredentialScopesReq) (*pbds.ListCredentialScopesResp, error) {

	kt := kit.FromGrpcContext(ctx)

	details, count, err := s.dao.CredentialScope().Get(kt, req.CredentialId, req.BizId)
	if err != nil {
		logs.Errorf("list credential scope failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	credentialScopes, err := pbcrs.PbCredentialScopes(details)
	if err != nil {
		logs.Errorf("get pb credential scope failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	resp := &pbds.ListCredentialScopesResp{
		Count:   uint32(count),
		Details: credentialScopes,
	}
	return resp, nil
}

// UpdateCredentialScopes update credential scopes
func (s *Service) UpdateCredentialScopes(ctx context.Context,
	req *pbds.UpdateCredentialScopesReq) (*pbds.UpdateCredentialScopesResp, error) {

	kt := kit.FromGrpcContext(ctx)

	tx := s.dao.GenQuery().Begin()

	for _, updated := range req.Updated {
		credScope, err := credential.New(updated.App, updated.Scope)
		if err != nil {
			return nil, err
		}

		credentialScope := &table.CredentialScope{
			ID: updated.Id,
			Spec: &table.CredentialScopeSpec{
				CredentialScope: credScope,
				ExpiredAt:       time.Now().UTC(),
			},
			Attachment: &table.CredentialScopeAttachment{
				BizID:        req.BizId,
				CredentialId: req.CredentialId,
			},
			Revision: &table.Revision{
				Reviser: kt.User,
			},
		}
		if err := s.dao.CredentialScope().UpdateWithTx(kt, tx, credentialScope); err != nil {
			logs.Errorf("update credential scope failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
	}
	for _, deleted := range req.Deleted {
		if err := s.dao.CredentialScope().DeleteWithTx(kt, tx, req.BizId, deleted); err != nil {
			logs.Errorf("delete credential scope failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
	}

	for _, created := range req.Created {
		credScope, err := credential.New(created.App, created.Scope)
		if err != nil {
			return nil, err
		}

		credentialScope := &table.CredentialScope{
			Spec: &table.CredentialScopeSpec{
				CredentialScope: credScope,
				ExpiredAt:       time.Now().UTC(),
			},
			Attachment: &table.CredentialScopeAttachment{
				BizID:        req.BizId,
				CredentialId: req.CredentialId,
			},
			Revision: &table.Revision{
				Creator: kt.User,
				Reviser: kt.User,
			},
		}
		if _, err := s.dao.CredentialScope().CreateWithTx(kt, tx, credentialScope); err != nil {
			logs.Errorf("create credential scope failed, err: %v, rid: %s", err, kt.Rid)
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
			}
			return nil, err
		}
	}

	if err := s.dao.Credential().UpdateRevisionWithTx(kt, tx, req.BizId, req.CredentialId); err != nil {
		logs.Errorf("update credential revision failed, err: %v, rid: %s", err, kt.Rid)
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kt.Rid)
		}
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	resp := &pbds.UpdateCredentialScopesResp{}
	return resp, nil
}
