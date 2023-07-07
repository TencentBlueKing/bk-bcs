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
	"fmt"

	"bscp.io/pkg/cc"
	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// Credential supplies all the Credential related operations.
type Credential interface {
	// Get get credential
	Get(kit *kit.Kit, bizID, id uint32) (*table.Credential, error)
	// GetByCredentialString get credential by credential string
	GetByCredentialString(kit *kit.Kit, bizID uint32, credential string) (*table.Credential, error)
	// Create one credential instance.
	Create(kit *kit.Kit, credential *table.Credential) (uint32, error)
	// List get credentials
	List(kit *kit.Kit, bizID uint32, searchKey string, opt *types.BasePage) ([]*table.Credential, int64, error)
	// Delete delete credential
	Delete(kit *kit.Kit, strategy *table.Credential) error
	// Update update credential
	Update(kit *kit.Kit, credential *table.Credential) error
	// UpdateRevisionWithTx update credential revision with transaction
	UpdateRevisionWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, id uint32) error
}

var _ Credential = new(credentialDao)

type credentialDao struct {
	genQ              *gen.Query
	idGen             IDGenInterface
	auditDao          AuditDao
	credentialSetting *cc.Credential
}

// Get ..
func (dao *credentialDao) Get(kit *kit.Kit, bizID, id uint32) (*table.Credential, error) {
	if bizID == 0 {
		return nil, errors.New("bizID is empty")
	}
	if id == 0 {
		return nil, errors.New("credential id is empty")
	}

	m := dao.genQ.Credential
	q := dao.genQ.Credential.WithContext(kit.Ctx)

	credential, err := q.Where(m.BizID.Eq(bizID), m.ID.Eq(id)).Take()
	if err != nil {
		return nil, fmt.Errorf("get credential failed, err: %v", err)
	}

	return credential, nil
}

// Get Credential by encoded credential string.
func (dao *credentialDao) GetByCredentialString(kit *kit.Kit, bizID uint32, str string) (*table.Credential, error) {
	if bizID == 0 {
		return nil, errors.New("bizID is empty")
	}
	if str == "" {
		return nil, errors.New("credential string is empty")
	}

	// encode credential string
	encryptionAlgorithm := dao.credentialSetting.EncryptionAlgorithm
	masterKey := dao.credentialSetting.MasterKey
	encrypted, err := tools.EncryptCredential(str, masterKey, encryptionAlgorithm)
	if err != nil {
		return nil, errf.ErrCredentialInvalid
	}

	m := dao.genQ.Credential
	q := dao.genQ.Credential.WithContext(kit.Ctx)

	credential, err := q.Where(m.BizID.Eq(bizID), m.EncCredential.Eq(encrypted)).Take()
	if err != nil {
		return nil, fmt.Errorf("get credential failed, err: %v", err)
	}

	return credential, nil
}

// Create create credential
func (dao *credentialDao) Create(kit *kit.Kit, g *table.Credential) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate a credential id and update to credential.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.Credential.WithContext(kit.Ctx)
		if err := q.Create(g); err != nil {
			return err
		}

		if err := ad.Do(tx); err != nil {
			return err
		}

		return nil
	}
	if err := dao.genQ.Transaction(createTx); err != nil {
		return 0, nil
	}

	return g.ID, nil
}

// List get credentials
func (dao *credentialDao) List(kit *kit.Kit, bizID uint32, searchKey string, opt *types.BasePage) (
	[]*table.Credential, int64, error) {
	m := dao.genQ.Credential
	q := dao.genQ.Credential.WithContext(kit.Ctx)

	result, count, err := q.Where(m.BizID.Eq(bizID)).
		Where(q.Where(m.Memo.Regexp("(?i)"+searchKey)).Or(m.Reviser.Regexp("(?i)"+searchKey))).
		Order(m.ID.Desc()).
		FindByPage(opt.Offset(), opt.LimitInt())
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

// Delete delete credential
func (dao *credentialDao) Delete(kit *kit.Kit, g *table.Credential) error {
	// 参数校验
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.Credential
	q := dao.genQ.Credential.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.Credential.WithContext(kit.Ctx)
		if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID)).Delete(g); err != nil {
			return err
		}

		if err := ad.Do(tx); err != nil {
			return err
		}
		return nil
	}
	if err := dao.genQ.Transaction(deleteTx); err != nil {
		return err
	}

	return nil
}

// Update update credential
// Note: only update name, description, enable
func (dao *credentialDao) Update(kit *kit.Kit, g *table.Credential) error {
	if err := g.ValidateUpdate(); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := dao.genQ.Credential
	q := dao.genQ.Credential.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.Credential.WithContext(kit.Ctx)
		if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
			Select(m.Memo, m.Enable).Updates(g); err != nil {
			return err
		}

		if err := ad.Do(tx); err != nil {
			return err
		}
		return nil
	}
	if err := dao.genQ.Transaction(updateTx); err != nil {
		return err
	}

	return nil
}

// UpdateRevisionWithTx update credential revision with transaction
func (dao *credentialDao) UpdateRevisionWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID uint32, id uint32) error {
	if bizID == 0 || id == 0 {
		return errors.New("credential bizID or id is zero")
	}

	m := tx.Credential
	q := tx.Credential.WithContext(kit.Ctx)
	if _, err := q.Where(m.BizID.Eq(bizID), m.ID.Eq(id)).
		Select(m.Reviser).Update(m.Reviser, kit.User); err != nil {
		return err
	}

	return nil
}
