/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/types"
)

// TemplateRelease supplies all the template release related operations.
type TemplateRelease interface {
	// Create one template release instance.
	Create(kit *kit.Kit, templateRelease *table.TemplateRelease) (uint32, error)
	// CreateWithTx create one template release instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, template *table.TemplateRelease) (uint32, error)
	// List templates with options.
	List(kit *kit.Kit, bizID, templateID uint32, opt *types.BasePage) ([]*table.TemplateRelease, int64, error)
	// Delete one template release instance.
	Delete(kit *kit.Kit, templateRelease *table.TemplateRelease) error
	// GetByUniqueKey get template release by unique key.
	GetByUniqueKey(kit *kit.Kit, bizID, templateID uint32, releaseName string) (*table.TemplateRelease, error)
}

var _ TemplateRelease = new(templateReleaseDao)

type templateReleaseDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one template release instance.
func (dao *templateReleaseDao) Create(kit *kit.Kit, g *table.TemplateRelease) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
		return 0, err
	}

	// generate a TemplateRelease id and update to TemplateRelease.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.TemplateRelease.WithContext(kit.Ctx)
		if err := q.Create(g); err != nil {
			return err
		}

		if err := ad.Do(tx); err != nil {
			return err
		}

		return nil
	}
	if err := dao.genQ.Transaction(createTx); err != nil {
		return 0, err
	}

	return g.ID, nil
}

// CreateWithTx create one template release instance with transaction.
func (dao *templateReleaseDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.TemplateRelease) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate a TemplateRelease id and update to TemplateRelease.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	q := tx.TemplateRelease.WithContext(kit.Ctx)
	if err := q.Create(g); err != nil {
		return 0, err
	}

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)
	if err := ad.Do(tx.Query); err != nil {
		return 0, err
	}

	return g.ID, nil
}

// List template releases with options.
func (dao *templateReleaseDao) List(kit *kit.Kit, bizID, templateID uint32, opt *types.BasePage) (
	[]*table.TemplateRelease, int64, error) {
	m := dao.genQ.TemplateRelease
	q := dao.genQ.TemplateRelease.WithContext(kit.Ctx)

	result, count, err := q.Where(m.BizID.Eq(bizID), m.TemplateID.Eq(templateID)).
		FindByPage(opt.Offset(), opt.LimitInt())
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

// Delete one template release instance.
func (dao *templateReleaseDao) Delete(kit *kit.Kit, g *table.TemplateRelease) error {
	// 参数校验
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.TemplateRelease
	q := dao.genQ.TemplateRelease.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.TemplateRelease.WithContext(kit.Ctx)
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

// GetByUniqueKey get template release by unique key
func (dao *templateReleaseDao) GetByUniqueKey(kit *kit.Kit, bizID, templateID uint32, releaseName string) (
	*table.TemplateRelease, error) {
	m := dao.genQ.TemplateRelease
	q := dao.genQ.TemplateRelease.WithContext(kit.Ctx)

	templateRelease, err := q.Where(m.BizID.Eq(bizID), m.TemplateID.Eq(templateID),
		m.ReleaseName.Eq(releaseName)).Take()
	if err != nil {
		return nil, fmt.Errorf("get template release failed, err: %v", err)
	}

	return templateRelease, nil
}

// validateAttachmentExist validate if attachment resource exists before operating template release
func (dao *templateReleaseDao) validateAttachmentExist(kit *kit.Kit, am *table.TemplateReleaseAttachment) error {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)

	if _, err := q.Where(m.ID.Eq(am.TemplateID)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("template release attached template %d is not exist", am.TemplateID)
		}
		return fmt.Errorf("get template release attached template failed, err: %v", err)
	}

	return nil
}
