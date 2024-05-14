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

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// AppTemplateBinding supplies all the app template binding related operations.
type AppTemplateBinding interface {
	// Create one app template binding instance.
	Create(kit *kit.Kit, atb *table.AppTemplateBinding) (uint32, error)
	// Update one app template binding's info.
	Update(kit *kit.Kit, atb *table.AppTemplateBinding) error
	// UpdateWithTx Update one app template binding's info with transaction.
	UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, atb *table.AppTemplateBinding) error
	// List app template bindings with options.
	List(kit *kit.Kit, bizID, appID uint32, opt *types.BasePage) ([]*table.AppTemplateBinding, int64, error)
	// Delete one app template binding instance.
	Delete(kit *kit.Kit, atb *table.AppTemplateBinding) error
	// DeleteByAppIDWithTx delete one app template binding instance by app id with transaction.
	DeleteByAppIDWithTx(kit *kit.Kit, tx *gen.QueryTx, appID uint32) error
}

var _ AppTemplateBinding = new(appTemplateBindingDao)

type appTemplateBindingDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one app template binding instance.
func (dao *appTemplateBindingDao) Create(kit *kit.Kit, g *table.AppTemplateBinding) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}
	if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
		return 0, err
	}

	// generate a app template binding id and update to app template binding.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.AppTemplateBinding.WithContext(kit.Ctx)
		if err = q.Create(g); err != nil {
			return err
		}

		if err = ad.Do(tx); err != nil {
			return err
		}

		return nil
	}
	if err = dao.genQ.Transaction(createTx); err != nil {
		return 0, err
	}

	return g.ID, nil
}

// Update one app template binding instance.
func (dao *appTemplateBindingDao) Update(kit *kit.Kit, g *table.AppTemplateBinding) error {
	if err := g.ValidateUpdate(); err != nil {
		return err
	}
	if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.AppTemplateBinding.WithContext(kit.Ctx)
		if _, err = q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
			Select(m.Bindings, m.TemplateSpaceIDs, m.TemplateSetIDs, m.TemplateIDs, m.TemplateRevisionIDs,
				m.LatestTemplateIDs, m.Creator, m.Reviser, m.UpdatedAt).
			Updates(g); err != nil {
			return err
		}

		if err = ad.Do(tx); err != nil {
			return err
		}
		return nil
	}
	if err = dao.genQ.Transaction(updateTx); err != nil {
		return err
	}

	return nil
}

// UpdateWithTx Update one app template binding's info with transaction.
func (dao *appTemplateBindingDao) UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx,
	g *table.AppTemplateBinding) error {
	if err := g.ValidateUpdate(); err != nil {
		return err
	}
	if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := tx.AppTemplateBinding
	q := tx.AppTemplateBinding.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)
	if err = ad.Do(tx.Query); err != nil {
		return err
	}

	q = tx.AppTemplateBinding.WithContext(kit.Ctx)
	if _, err = q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
		Select(m.Bindings, m.TemplateSpaceIDs, m.TemplateSetIDs, m.TemplateIDs, m.TemplateRevisionIDs,
			m.LatestTemplateIDs, m.Creator, m.Reviser, m.UpdatedAt).
		Updates(g); err != nil {
		return err
	}

	return nil
}

// List app template bindings with options.
func (dao *appTemplateBindingDao) List(kit *kit.Kit, bizID, appID uint32,
	opt *types.BasePage) ([]*table.AppTemplateBinding, int64, error) {
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)

	d := q.Where(m.BizID.Eq(bizID), m.AppID.Eq(appID))
	if opt.All {
		result, err := d.Find()
		if err != nil {
			return nil, 0, err
		}
		return result, int64(len(result)), err
	}

	return d.FindByPage(opt.Offset(), opt.LimitInt())
}

// Delete one app template binding instance.
func (dao *appTemplateBindingDao) Delete(kit *kit.Kit, g *table.AppTemplateBinding) error {
	// 参数校验
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.AppTemplateBinding
	q := dao.genQ.AppTemplateBinding.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.AppTemplateBinding.WithContext(kit.Ctx)
		if _, err = q.Where(m.BizID.Eq(g.Attachment.BizID)).Delete(g); err != nil {
			return err
		}

		if err = ad.Do(tx); err != nil {
			return err
		}
		return nil
	}
	if err = dao.genQ.Transaction(deleteTx); err != nil {
		return err
	}

	return nil
}

// DeleteByAppIDWithTx delete one app template binding instance by app id with transaction.
func (dao *appTemplateBindingDao) DeleteByAppIDWithTx(kit *kit.Kit, tx *gen.QueryTx, appID uint32) error {
	m := tx.AppTemplateBinding
	q := tx.AppTemplateBinding.WithContext(kit.Ctx)
	_, err := q.Where(m.AppID.Eq(appID)).Delete()
	return err
}

// validateAttachmentExist validate if attachment resource exists before operating template
func (dao *appTemplateBindingDao) validateAttachmentExist(kit *kit.Kit, am *table.AppTemplateBindingAttachment) error {
	m := dao.genQ.App
	q := dao.genQ.App.WithContext(kit.Ctx)

	if _, err := q.Where(m.ID.Eq(am.AppID)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("template attached app %d is not exist", am.AppID)
		}
		return fmt.Errorf("get template attached app failed, err: %v", err)
	}

	return nil
}
