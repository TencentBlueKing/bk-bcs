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
	"gorm.io/gen/field"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/utils"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/search"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Template supplies all the template related operations.
type Template interface {
	// Create one template instance.
	Create(kit *kit.Kit, template *table.Template) (uint32, error)
	// CreateWithTx create one template instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, template *table.Template) (uint32, error)
	// Update one template's info.
	Update(kit *kit.Kit, template *table.Template) error
	// UpdateWithTx Update one template instance with transaction.
	UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, template *table.Template) error
	// List templates with options.
	List(kit *kit.Kit, bizID, templateSpaceID uint32, s search.Searcher,
		opt *types.BasePage, topIds []uint32, searchValue string) ([]*table.Template, int64, error)
	// Delete one template instance.
	Delete(kit *kit.Kit, template *table.Template) error
	// DeleteWithTx delete one template instance with transaction.
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, template *table.Template) error
	// GetByUniqueKey get template by unique key.
	GetByUniqueKey(kit *kit.Kit, bizID, templateSpaceID uint32, name, path string) (*table.Template, error)
	// GetByID get template by id.
	GetByID(kit *kit.Kit, bizID, templateID uint32) (*table.Template, error)
	// ListByIDs list templates by template ids.
	ListByIDs(kit *kit.Kit, ids []uint32) ([]*table.Template, error)
	// ListByIDsWithTx list templates by template ids with transaction.
	ListByIDsWithTx(kit *kit.Kit, tx *gen.QueryTx, ids []uint32) ([]*table.Template, error)
	// ListAllIDs list all template ids.
	ListAllIDs(kit *kit.Kit, bizID, templateSpaceID uint32) ([]uint32, error)
	// BatchCreateWithTx batch create template instances with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, templates []*table.Template) error
	// BatchUpdateWithTx batch update template instances with transaction.
	BatchUpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, templates []*table.Template) error
	// ListTemplateByTuple 按照多个字段in查询template 列表
	ListTemplateByTuple(kit *kit.Kit, data [][]interface{}) ([]*table.Template, error)
	// ListByExclusionIDs list templates by template exclusion ids.
	ListByExclusionIDs(kit *kit.Kit, ids []uint32) ([]*table.Template, error)
}

var _ Template = new(templateDao)

type templateDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// UpdateWithTx Update one template instance with transaction.
func (dao *templateDao) UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.Template) error {
	if err := g.ValidateUpdate(kit); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := tx.Template
	q := tx.Template.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)
	if err := ad.Do(tx.Query); err != nil {
		return err
	}

	if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).UpdateColumns(g); err != nil {
		return err
	}

	return nil
}

// ListByExclusionIDs list templates by template exclusion ids.
func (dao *templateDao) ListByExclusionIDs(kit *kit.Kit, ids []uint32) ([]*table.Template, error) {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)
	result, err := q.Where(m.ID.NotIn(ids...)).Find()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListTemplateByTuple 按照多个字段in查询template 列表
func (dao *templateDao) ListTemplateByTuple(kit *kit.Kit, data [][]interface{}) ([]*table.Template, error) {
	m := dao.genQ.Template
	return dao.genQ.Template.WithContext(kit.Ctx).
		Select(m.ID, m.BizID, m.TemplateSpaceID, m.Name, m.Path).
		Where(m.WithContext(kit.Ctx).Columns(m.BizID, m.TemplateSpaceID, m.Name, m.Path).
			In(field.Values(data))).
		Find()
}

// BatchUpdateWithTx batch update template instances with transaction.
func (dao *templateDao) BatchUpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, templates []*table.Template) error {
	if len(templates) == 0 {
		return nil
	}
	if err := tx.Template.WithContext(kit.Ctx).Save(templates...); err != nil {
		return err
	}
	return nil
}

// BatchCreateWithTx batch create template instances with transaction.
func (dao *templateDao) BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, templates []*table.Template) error {
	if len(templates) == 0 {
		return nil
	}
	ids, err := dao.idGen.Batch(kit, table.TemplateTable, len(templates))
	if err != nil {
		return err
	}

	for i, item := range templates {
		if err = item.ValidateCreate(kit); err != nil {
			return err
		}
		item.ID = ids[i]
	}
	err = tx.Template.WithContext(kit.Ctx).Create(templates...)
	if err != nil {
		return err
	}
	return nil
}

// Create one template instance.
func (dao *templateDao) Create(kit *kit.Kit, g *table.Template) (uint32, error) {
	if err := g.ValidateCreate(kit); err != nil {
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
	if err := g.ValidateCreate(kit); err != nil {
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
	if err := g.ValidateUpdate(kit); err != nil {
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
func (dao *templateDao) List(kit *kit.Kit, bizID, templateSpaceID uint32, s search.Searcher,
	opt *types.BasePage, topIds []uint32, searchValue string) ([]*table.Template, int64, error) {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)

	var conds []rawgen.Condition
	// add search condition
	if s != nil {
		exprs := s.SearchExprs(dao.genQ)
		if len(exprs) > 0 {
			var do gen.ITemplateDo
			for i := range exprs {
				if i == 0 {
					do = q.Where(exprs[i])
				}
				do = do.Or(exprs[i])
			}

			do = do.Or(utils.RawCond(`CASE WHEN RIGHT(path, 1) = '/' THEN CONCAT(path,name)
			 ELSE CONCAT_WS('/', path, name) END LIKE ?`, "%"+searchValue+"%"))
			conds = append(conds, do)
		}
	}

	d := q.Where(m.BizID.Eq(bizID), m.TemplateSpaceID.Eq(templateSpaceID)).Where(conds...)
	if len(topIds) != 0 {
		d = d.Order(utils.NewCustomExpr("CASE WHEN id IN (?) THEN 0 ELSE 1 END,"+
			"CASE WHEN RIGHT(path, 1) = '/' THEN CONCAT(path,`name`) ELSE CONCAT_WS('/', path, `name`) END",
			[]interface{}{topIds}))
	} else {
		d = d.Order(utils.NewCustomExpr("CASE WHEN RIGHT(path, 1) = '/' THEN CONCAT(path,`name`) ELSE "+
			"CONCAT_WS('/', path, `name`) END", nil))
	}

	if opt.All {
		result, err := d.Find()
		if err != nil {
			return nil, 0, err
		}
		return result, int64(len(result)), err
	}

	return d.FindByPage(opt.Offset(), opt.LimitInt())
}

// Delete one template instance.
func (dao *templateDao) Delete(kit *kit.Kit, g *table.Template) error {
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

// DeleteWithTx delete one template instance with transaction.
func (dao *templateDao) DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.Template) error {
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := tx.Template
	q := tx.Template.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)
	if err := ad.Do(tx.Query); err != nil {
		return err
	}

	if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID)).Delete(g); err != nil {
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

// ListByIDs list templates by template ids.
func (dao *templateDao) ListByIDs(kit *kit.Kit, ids []uint32) ([]*table.Template, error) {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)
	result, err := q.Where(m.ID.In(ids...)).
		Order(utils.NewCustomExpr("CASE WHEN RIGHT(path, 1) = '/' THEN CONCAT(path,`name`) "+
			"ELSE CONCAT_WS('/', path, `name`) END", nil)).Find()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListByIDsWithTx list templates by template ids with transaction.
func (dao *templateDao) ListByIDsWithTx(kit *kit.Kit, tx *gen.QueryTx, ids []uint32) ([]*table.Template, error) {
	m := tx.Template
	q := tx.Template.WithContext(kit.Ctx)
	result, err := q.Where(m.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// ListAllIDs list all template ids.
func (dao *templateDao) ListAllIDs(kit *kit.Kit, bizID, templateSpaceID uint32) ([]uint32, error) {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)

	var templateIDs []uint32
	if err := q.Select(m.ID).
		Where(m.BizID.Eq(bizID), m.TemplateSpaceID.Eq(templateSpaceID)).
		Pluck(m.ID, &templateIDs); err != nil {
		return nil, err
	}

	return templateIDs, nil
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
