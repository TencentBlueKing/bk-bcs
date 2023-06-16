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

// Template supplies all the template related operations.
type Template interface {
	// Create one template instance.
	Create(kit *kit.Kit, template *table.Template) (uint32, error)
	// CreateWithTx create one template instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, template *table.Template) (uint32, error)
	// Update one template's info.
	Update(kit *kit.Kit, template *table.Template) error
	// List templates with options.
	List(kit *kit.Kit, bizID, templateSpaceID uint32, opt *types.BasePage) ([]*table.Template, int64, error)
	// Delete one template instance.
	Delete(kit *kit.Kit, template *table.Template) error
	// GetByUniqueKey get template by unique key.
	GetByUniqueKey(kit *kit.Kit, bizID, templateSpaceID uint32, name, path string) (*table.Template, error)
	// GetByID get template by id.
	GetByID(kit *kit.Kit, bizID, templateID uint32) (*table.Template, error)
}

var _ Template = new(templateDao)

type templateDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one template instance.
func (dao *templateDao) Create(kit *kit.Kit, g *table.Template) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
		return 0, err
	}

	// generate a Template id and update to Template.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.Template.WithContext(kit.Ctx)
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

// CreateWithTx create one template instance with transaction.
func (dao *templateDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.Template) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate a Template id and update to Template.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	q := tx.Template.WithContext(kit.Ctx)
	if err := q.Create(g); err != nil {
		return 0, err
	}

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)
	if err := ad.Do(tx.Query); err != nil {
		return 0, err
	}

	return g.ID, nil
}

// Update one template instance.
func (dao *templateDao) Update(kit *kit.Kit, g *table.Template) error {
	if err := g.ValidateUpdate(); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.Template.WithContext(kit.Ctx)
		if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
			Select(m.Memo, m.Reviser).Updates(g); err != nil {
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

// List templates with options.
func (dao *templateDao) List(kit *kit.Kit, bizID, templateSpaceID uint32, opt *types.BasePage) (
	[]*table.Template, int64, error) {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)

	result, count, err := q.Where(m.BizID.Eq(bizID), m.TemplateSpaceID.Eq(templateSpaceID)).
		FindByPage(opt.Offset(), opt.LimitInt())
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

// Delete one template instance.
func (dao *templateDao) Delete(kit *kit.Kit, g *table.Template) error {
	// 参数校验
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.Template.WithContext(kit.Ctx)
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

// GetByUniqueKey get template by unique key
func (dao *templateDao) GetByUniqueKey(kit *kit.Kit, bizID, templateSpaceID uint32, name, path string) (
	*table.Template, error) {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)

	template, err := q.Where(m.BizID.Eq(bizID), m.TemplateSpaceID.Eq(templateSpaceID), m.Name.Eq(name),
		m.Path.Eq(path)).Take()
	if err != nil {
		return nil, fmt.Errorf("get template failed, err: %v", err)
	}

	return template, nil
}

// GetByID get template by id
func (dao *templateDao) GetByID(kit *kit.Kit, bizID, templateID uint32) (*table.Template, error) {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)

	template, err := q.Where(m.BizID.Eq(bizID), m.ID.Eq(templateID)).Take()
	if err != nil {
		return nil, fmt.Errorf("get template failed, err: %v", err)
	}

	return template, nil
}

// validateAttachmentExist validate if attachment resource exists before operating template
func (dao *templateDao) validateAttachmentExist(kit *kit.Kit, am *table.TemplateAttachment) error {
	m := dao.genQ.TemplateSpace
	q := dao.genQ.TemplateSpace.WithContext(kit.Ctx)

	if _, err := q.Where(m.ID.Eq(am.TemplateSpaceID)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("template attached template space %d is not exist", am.TemplateSpaceID)
		}
		return fmt.Errorf("get template attached template space failed, err: %v", err)
	}

	return nil
}
