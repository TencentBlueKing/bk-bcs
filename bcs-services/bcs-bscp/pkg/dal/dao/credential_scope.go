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

package dao

import (
	"errors"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
)

// CredentialScope supplies all the credential scope related operations.
type CredentialScope interface {
	// CreateWithTx create credential scope with transaction
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, credential *table.CredentialScope) (uint32, error)
	// Get get credential scopes
	Get(kit *kit.Kit, credentialId, bizID uint32) ([]*table.CredentialScope, int64, error)
	// DeleteWithTx delete credential scope with transaction
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, id uint32) error
	// UpdateWithTx update credential scope with transaction
	UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, credentialScope *table.CredentialScope) error
}

var _ CredentialScope = new(credentialScopeDao)

type credentialScopeDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// CreateWithTx create credential scope with transaction
func (dao *credentialScopeDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.CredentialScope) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate a Template id and update to Template.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	q := tx.CredentialScope.WithContext(kit.Ctx)
	if err := q.Create(g); err != nil {
		return 0, err
	}

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)
	if err := ad.Do(tx.Query); err != nil {
		return 0, err
	}

	return g.ID, nil
}

// Get get credential scope
func (dao *credentialScopeDao) Get(kit *kit.Kit, credentialId, bizID uint32) ([]*table.CredentialScope, int64, error) {
	m := dao.genQ.CredentialScope
	q := dao.genQ.CredentialScope.WithContext(kit.Ctx)

	result, err := q.Where(m.BizID.Eq(bizID), m.CredentialId.Eq(credentialId)).Find()
	if err != nil {
		return nil, 0, err
	}

	return result, int64(len(result)), nil
}

// DeleteWithTx delete credential scope with transaction
func (dao *credentialScopeDao) DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, id uint32) error {
	if bizID == 0 {
		return errors.New("biz id is zero")
	}

	if id == 0 {
		return errors.New("credential scope id is zero")
	}

	// 删除操作, 获取当前记录做审计
	m := tx.CredentialScope
	q := tx.CredentialScope.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(id), m.BizID.Eq(bizID)).Take()
	if err != nil {
		return err
	}

	if _, err := q.Where(m.BizID.Eq(bizID), m.ID.Eq(id)).Delete(); err != nil {
		return err
	}

	ad := dao.auditDao.DecoratorV2(kit, bizID).PrepareDelete(oldOne)
	if err := ad.Do(tx.Query); err != nil {
		return err
	}

	return nil
}

// UpdateWithTx update credential scope with transaction
func (dao *credentialScopeDao) UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.CredentialScope) error {
	if err := g.ValidateUpdate(); err != nil {
		return err
	}

	m := tx.CredentialScope

	// 更新操作, 获取当前记录做审计
	oldOne, err := tx.CredentialScope.WithContext(kit.Ctx).Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	if _, err := tx.CredentialScope.WithContext(kit.Ctx).Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
		Omit(m.BizID, m.ID).Updates(g); err != nil {
		return err
	}

	if err := ad.Do(tx.Query); err != nil {
		return err
	}

	return nil
}
