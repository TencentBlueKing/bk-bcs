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

package dao

import (
	"errors"
	"fmt"

	rawgen "gorm.io/gen"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/utils"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Credential supplies all the Credential related operations.
type Credential interface {
	// Get get credential
	Get(kit *kit.Kit, bizID, id uint32) (*table.Credential, error)
	// GetByCredentialString get credential by credential string
	GetByCredentialString(kit *kit.Kit, bizID uint32, credential string) (*table.Credential, error)
	// ListByCredentialString list credential by credential string array
	ListByCredentialString(kit *kit.Kit, bizID uint32, credentials []string) ([]*table.Credential, error)
	// BatchListByIDs batch list credential by ids
	BatchListByIDs(kit *kit.Kit, bizID uint32, ids []uint32) ([]*table.Credential, error)
	// Create one credential instance.
	Create(kit *kit.Kit, credential *table.Credential) (uint32, error)
	// List get credentials
	List(kit *kit.Kit, bizID uint32, searchKey string, opt *types.BasePage,
		topIds []uint32, encCredential string, enable *bool) ([]*table.Credential, int64, error)
	// DeleteWithTx delete credential with transaction
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, id uint32) error
	// Update update credential
	Update(kit *kit.Kit, credential *table.Credential) error
	// UpdateRevisionWithTx update credential revision with transaction
	UpdateRevisionWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, id uint32) error
	// GetByName get Credential by name.
	GetByName(kit *kit.Kit, bizID uint32, name string) (*table.Credential, error)
}

var _ Credential = new(credentialDao)

type credentialDao struct {
	genQ              *gen.Query
	idGen             IDGenInterface
	auditDao          AuditDao
	credentialSetting *cc.Credential
	event             Event
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

// GetByCredentialString get credential by encoded credential string.
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

// ListByCredentialString list credential by encoded credential string array.
func (dao *credentialDao) ListByCredentialString(kit *kit.Kit, bizID uint32, strArr []string) (
	[]*table.Credential, error) {
	if bizID == 0 {
		return nil, errors.New("bizID is empty")
	}
	if len(strArr) == 0 {
		return nil, errors.New("credential string is empty")
	}

	encryptedArr := make([]string, 0, len(strArr))

	for _, str := range strArr {
		// encode credential string
		encryptionAlgorithm := dao.credentialSetting.EncryptionAlgorithm
		masterKey := dao.credentialSetting.MasterKey
		encrypted, err := tools.EncryptCredential(str, masterKey, encryptionAlgorithm)
		if err != nil {
			return nil, errf.ErrCredentialInvalid
		}
		encryptedArr = append(encryptedArr, encrypted)
	}

	m := dao.genQ.Credential
	q := dao.genQ.Credential.WithContext(kit.Ctx)

	return q.Where(m.BizID.Eq(bizID), m.EncCredential.In(encryptedArr...)).Find()
}

// BatchListByIDs batch list credential by ids
func (dao *credentialDao) BatchListByIDs(kit *kit.Kit, bizID uint32, ids []uint32) ([]*table.Credential, error) {
	if bizID == 0 {
		return nil, errors.New("bizID is empty")
	}
	if len(ids) == 0 {
		return []*table.Credential{}, nil
	}

	m := dao.genQ.Credential

	return dao.genQ.Credential.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.ID.In(ids...)).Find()
}

// Create create credential
func (dao *credentialDao) Create(kit *kit.Kit, g *table.Credential) (uint32, error) {
	if err := g.ValidateCreate(kit); err != nil {
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
func (dao *credentialDao) List(kit *kit.Kit, bizID uint32, searchKey string, opt *types.BasePage,
	topIds []uint32, encCredential string, enable *bool) ([]*table.Credential, int64, error) {
	m := dao.genQ.Credential
	q := dao.genQ.Credential.WithContext(kit.Ctx)
	cs := dao.genQ.CredentialScope

	var conds []rawgen.Condition
	if searchKey != "" {
		searchVal := "(?i)" + searchKey

		var item []struct {
			CredentialID uint32
		}
		err := cs.WithContext(kit.Ctx).Select(cs.CredentialId).
			Where(cs.BizID.Eq(bizID), cs.CredentialScope.Regexp(searchVal)).Group(cs.CredentialId).Scan(&item)
		if err != nil {
			return nil, 0, err
		}
		if len(item) > 0 {
			credentialID := []uint32{}
			for _, v := range item {
				credentialID = append(credentialID, v.CredentialID)
			}
			conds = append(conds, q.Where(m.Memo.Regexp(searchVal)).Or(m.Reviser.Regexp(searchVal)).
				Or(m.Name.Regexp(searchVal)).Or(m.ID.In(credentialID...)).Or(m.EncCredential.Eq(encCredential)))
		} else {
			conds = append(conds, q.Where(m.Memo.Regexp(searchVal)).Or(m.Reviser.Regexp(searchVal)).
				Or(m.Name.Regexp(searchVal)).Or(m.EncCredential.Eq(encCredential)))
		}

	}

	if len(topIds) != 0 {
		q = q.Order(utils.NewCustomExpr(`CASE WHEN id IN (?) THEN 0 ELSE 1 END,name ASC`, []interface{}{topIds}))
	} else {
		q = q.Order(m.Name)
	}

	if enable != nil {
		q = q.Where(m.Enable.Is(*enable))
	}

	q = q.Where(m.BizID.Eq(bizID)).Where(conds...)
	if opt.All {
		result, err := q.Find()
		if err != nil {
			return nil, 0, err
		}
		return result, int64(len(result)), err
	}

	return q.FindByPage(opt.Offset(), opt.LimitInt())
}

// Delete delete credential
// !Note: delete credential should emit a delete event.
func (dao *credentialDao) DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, id uint32) error {
	// 参数校验
	if bizID == 0 {
		return errf.New(errf.InvalidParameter, "bizID is 0")
	}
	if id == 0 {
		return errf.New(errf.InvalidParameter, "id is 0")
	}

	// 删除操作, 获取当前记录做审计
	m := tx.Credential
	q := tx.Credential.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(id), m.BizID.Eq(bizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, bizID).PrepareDelete(oldOne)

	if _, e := q.Where(m.BizID.Eq(bizID), m.ID.Eq(id)).Delete(); e != nil {
		return e
	}

	// decode credential string
	masterKey := dao.credentialSetting.MasterKey
	encrypted, err := tools.DecryptCredential(oldOne.Spec.EncCredential, masterKey, oldOne.Spec.EncAlgorithm)
	if err != nil {
		return err
	}
	// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
	one := types.Event{
		Spec: &table.EventSpec{
			Resource:    table.CredentialEvent,
			ResourceID:  id,
			ResourceUid: encrypted,
			OpType:      table.DeleteOp,
		},
		Attachment: &table.EventAttachment{BizID: bizID},
		Revision:   &table.CreatedRevision{Creator: kit.User},
	}
	eDecorator := dao.event.Eventf(kit)
	if err = eDecorator.FireWithTx(tx, one); err != nil {
		logs.Errorf("fire delete credential: %d event failed, err: %v, rid: %s", id, err, kit.Rid)
		return errors.New("fire event failed, " + err.Error())
	}

	return ad.Do(tx.Query)
}

// Update update credential's name, description, enable
// !Note: update credential should emit a update event.
func (dao *credentialDao) Update(kit *kit.Kit, g *table.Credential) error {
	if err := g.ValidateUpdate(kit); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := dao.genQ.Credential
	q := dao.genQ.Credential.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}

	// decode credential string
	masterKey := dao.credentialSetting.MasterKey
	encrypted, err := tools.DecryptCredential(oldOne.Spec.EncCredential, masterKey, oldOne.Spec.EncAlgorithm)
	if err != nil {
		return err
	}
	// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
	one := types.Event{
		Spec: &table.EventSpec{
			Resource:    table.CredentialEvent,
			ResourceID:  g.ID,
			ResourceUid: encrypted,
			OpType:      table.UpdateOp,
		},
		Attachment: &table.EventAttachment{BizID: g.Attachment.BizID},
		Revision:   &table.CreatedRevision{Creator: kit.User},
	}
	eDecorator := dao.event.Eventf(kit)
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.Credential.WithContext(kit.Ctx)
		if _, e := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
			Select(m.Memo, m.Name, m.Enable, m.Reviser).Updates(g); e != nil {
			return e
		}

		if e := ad.Do(tx); e != nil {
			return e
		}

		if e := eDecorator.Fire(one); e != nil {
			logs.Errorf("fire update credential: %s event failed, err: %v, rid: %s", g.ID, e, kit.Rid)
			return errors.New("fire event failed, " + e.Error())
		}

		return nil
	}
	err = dao.genQ.Transaction(updateTx)

	eDecorator.Finalizer(err)

	return err
}

// UpdateRevisionWithTx update credential revision with transaction
// !Note: update credential should emit a update event.
func (dao *credentialDao) UpdateRevisionWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID uint32, id uint32) error {
	if bizID == 0 || id == 0 {
		return errors.New("credential bizID or id is zero")
	}

	m := tx.Credential
	oldOne, err := m.WithContext(kit.Ctx).Where(m.ID.Eq(id), m.BizID.Eq(bizID)).Take()
	if err != nil {
		return err
	}

	// decode credential string
	masterKey := dao.credentialSetting.MasterKey
	encrypted, err := tools.DecryptCredential(oldOne.Spec.EncCredential, masterKey, oldOne.Spec.EncAlgorithm)
	if err != nil {
		return err
	}

	q := tx.Credential.WithContext(kit.Ctx)
	if _, e := q.Where(m.BizID.Eq(bizID), m.ID.Eq(id)).
		Select(m.Reviser).Update(m.Reviser, kit.User); e != nil {
		return e
	}

	// fire the event with txn to ensure the if save the event failed then the business logic is failed anyway.
	one := types.Event{
		Spec: &table.EventSpec{
			Resource:    table.CredentialEvent,
			ResourceID:  id,
			ResourceUid: encrypted,
			OpType:      table.UpdateOp,
		},
		Attachment: &table.EventAttachment{BizID: bizID},
		Revision:   &table.CreatedRevision{Creator: kit.User},
	}
	eDecorator := dao.event.Eventf(kit)
	if err = eDecorator.FireWithTx(tx, one); err != nil {
		logs.Errorf("fire update credential: %d event failed, err: %v, rid: %s", id, err, kit.Rid)
		return errors.New("fire event failed, " + err.Error())
	}

	return nil
}

// GetByName get Credential by name.
func (dao *credentialDao) GetByName(kit *kit.Kit, bizID uint32, name string) (*table.Credential, error) {

	m := dao.genQ.Credential
	q := dao.genQ.Credential.WithContext(kit.Ctx)

	credential, err := q.Where(m.BizID.Eq(bizID), m.Name.Eq(name)).Take()
	if err != nil {
		return nil, err
	}

	return credential, nil
}
