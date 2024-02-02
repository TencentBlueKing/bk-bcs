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
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbcredential "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/credential"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// CreateCredential Create Credential
func (s *Service) CreateCredential(ctx context.Context, req *pbds.CreateCredentialReq) (*pbds.CreateResp, error) {
	kt := kit.FromGrpcContext(ctx)

	if _, err := s.dao.Credential().GetByName(kt, req.Attachment.BizId, req.Spec.Name); err == nil {
		return nil, fmt.Errorf("credential name %s already exists", req.Spec.Name)
	}

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

	// StrToUint32Slice the comma separated string goes to uint32 slice
	topIds, _ := tools.StrToUint32Slice(req.TopIds)
	details, count, err := s.dao.Credential().List(kt, req.BizId, req.SearchKey, opt, topIds)

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

	tx := s.dao.GenQuery().Begin()

	// 查看credential_scopes表中的数据
	if err := s.dao.CredentialScope().DeleteByCredentialIDWithTx(kt, tx, req.Attachment.BizId, req.Id); err != nil {
		logs.Errorf("delete credential scope by credential id failed, err: %v, rid: %s", err, kt.Rid)
		if e := tx.Rollback(); e != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", e, kt.Rid)
		}
		return nil, err
	}

	if err := s.dao.Credential().DeleteWithTx(kt, tx, req.Attachment.BizId, req.Id); err != nil {
		logs.Errorf("delete credential failed, err: %v, rid: %s", err, kt.Rid)
		if e := tx.Rollback(); e != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", e, kt.Rid)
		}
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logs.Errorf("transaction commit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return new(pbbase.EmptyResp), nil
}

// UpdateCredential update credential
func (s *Service) UpdateCredential(ctx context.Context, req *pbds.UpdateCredentialReq) (*pbbase.EmptyResp, error) {
	kt := kit.FromGrpcContext(ctx)

	old, err := s.dao.Credential().GetByName(kt, req.Attachment.BizId, req.Spec.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logs.Errorf("get credential failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	if !errors.Is(gorm.ErrRecordNotFound, err) && old.ID != req.Id {
		return nil, fmt.Errorf("credential name %s already exists", req.Spec.Name)
	}

	credential := &table.Credential{
		ID:         req.Id,
		Spec:       req.Spec.CredentialSpec(),
		Attachment: req.Attachment.CredentialAttachment(),
		Revision: &table.Revision{
			Reviser: kt.User,
		},
	}
	if e := s.dao.Credential().Update(kt, credential); e != nil {
		logs.Errorf("update credential failed, err: %v, rid: %s", e, kt.Rid)
		return nil, e
	}

	return new(pbbase.EmptyResp), nil
}
