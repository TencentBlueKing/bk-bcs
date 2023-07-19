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

	"gorm.io/datatypes"
	rawgen "gorm.io/gen"
	"gorm.io/gorm"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/tools"
	"bscp.io/pkg/types"
)

// TemplateSet supplies all the template set related operations.
type TemplateSet interface {
	// Create one template set instance.
	Create(kit *kit.Kit, templateSpace *table.TemplateSet) (uint32, error)
	// Update one template set's info.
	Update(kit *kit.Kit, templateSpace *table.TemplateSet) error
	// List template sets with options.
	List(kit *kit.Kit, bizID, templateSpaceID uint32, searchKey string, opt *types.BasePage) ([]*table.TemplateSet, int64, error)
	// Delete one template set instance.
	Delete(kit *kit.Kit, templateSet *table.TemplateSet) error
	// DeleteWithTx delete one template set instance with transaction.
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, templateSet *table.TemplateSet) error
	// GetByUniqueKey get template set by unique key.
	GetByUniqueKey(kit *kit.Kit, bizID, templateSpaceID uint32, name string) (*table.TemplateSet, error)
	// ListByIDs list template sets by template set ids.
	ListByIDs(kit *kit.Kit, ids []uint32) ([]*table.TemplateSet, error)
	// AddTemplateToTemplateSets add a template to template sets.
	AddTemplateToTemplateSets(kit *kit.Kit, tmplID uint32, tmplSetIDs []uint32) error
	// DeleteTmplFromTmplSetsWithTx delete a template from template sets with transaction.
	DeleteTmplFromTmplSetsWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, tmplID uint32) error
}

var _ TemplateSet = new(templateSetDao)

type templateSetDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one template set instance.
func (dao *templateSetDao) Create(kit *kit.Kit, g *table.TemplateSet) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
		return 0, err
	}

	if err := dao.validateTemplatesExist(kit, g.Spec.TemplateIDs); err != nil {
		return 0, err
	}

	// generate a template set id and update to template set.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.TemplateSet.WithContext(kit.Ctx)
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

// Update one template set instance.
func (dao *templateSetDao) Update(kit *kit.Kit, g *table.TemplateSet) error {
	if err := g.ValidateUpdate(); err != nil {
		return err
	}

	if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
		return err
	}

	m := dao.genQ.TemplateSet

	// 更新操作, 获取当前记录做审计
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.TemplateSet.WithContext(kit.Ctx)
		if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).Select(m.Memo, m.Reviser).Updates(g); err != nil {
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

// List template sets with options.
func (dao *templateSetDao) List(kit *kit.Kit, bizID, templateSpaceID uint32, searchKey string, opt *types.BasePage) ([]*table.TemplateSet, int64, error) {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)

	var conds []rawgen.Condition
	if searchKey != "" {
		conds = append(conds, q.Where(m.Name.Regexp("(?i)"+searchKey)).Or(m.Memo.Regexp("(?i)"+searchKey)))
	}

	result, count, err := q.Where(m.BizID.Eq(bizID), m.TemplateSpaceID.Eq(templateSpaceID)).
		Where(conds...).
		FindByPage(opt.Offset(), opt.LimitInt())
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

// Delete one template set instance.
func (dao *templateSetDao) Delete(kit *kit.Kit, g *table.TemplateSet) error {
	// 参数校验
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.TemplateSet.WithContext(kit.Ctx)
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

// DeleteWithTx delete one template set instance with transaction.
func (dao *templateSetDao) DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.TemplateSet) error {
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	m := tx.TemplateSet
	q := tx.TemplateSet.WithContext(kit.Ctx)
	if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID)).Delete(g); err != nil {
		return err
	}

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)
	if err := ad.Do(tx.Query); err != nil {
		return err
	}

	return nil
}

// GetByUniqueKey get template set by unique key
func (dao *templateSetDao) GetByUniqueKey(kit *kit.Kit, bizID, templateSpaceID uint32, name string) (
	*table.TemplateSet, error) {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)

	tplSet, err := q.Where(m.BizID.Eq(bizID), m.TemplateSpaceID.Eq(templateSpaceID), m.Name.Eq(name)).Take()
	if err != nil {
		return nil, fmt.Errorf("get templateSpace failed, err: %v", err)
	}

	return tplSet, nil
}

// ListByIDs list template sets by template set ids.
func (dao *templateSetDao) ListByIDs(kit *kit.Kit, ids []uint32) ([]*table.TemplateSet, error) {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)
	result, err := q.Where(m.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// AddTemplateToTemplateSets add a template to template sets.
func (dao *templateSetDao) AddTemplateToTemplateSets(kit *kit.Kit, tmplID uint32, tmplSetIDs []uint32) error {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)
	if _, err := q.Where(m.ID.In(tmplSetIDs...)).
		Not(rawgen.Cond(datatypes.JSONArrayQuery("template_ids").Contains(tmplID))...).
		Update(m.TemplateIDs, gorm.Expr("JSON_ARRAY_APPEND(template_ids, '$', ?)", tmplID)); err != nil {
		return err
	}
	return nil
}

// DeleteTmplFromTmplSetsWithTx delete a template from template sets with transaction.
func (dao *templateSetDao) DeleteTmplFromTmplSetsWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, tmplID uint32) error {
	m := tx.TemplateSet
	q := tx.TemplateSet.WithContext(kit.Ctx)
	// subQuery get the array of template ids after delete the target template id, set it to '[]' if no records found
	subQuery := "COALESCE ((SELECT JSON_ARRAYAGG(oid) new_oids FROM " +
		"JSON_TABLE (template_ids, '$[*]' COLUMNS (oid BIGINT (1) UNSIGNED PATH '$')) AS t1 WHERE oid<> ?), '[]')"
	if _, err := q.Where(m.BizID.Eq(bizID)).
		Where(rawgen.Cond(datatypes.JSONArrayQuery("template_ids").Contains(tmplID))...).
		Update(m.TemplateIDs, gorm.Expr(subQuery, tmplID)); err != nil {
		return err
	}
	return nil
}

// validateAttachmentExist validate if attachment resource exists before operating template
func (dao *templateSetDao) validateAttachmentExist(kit *kit.Kit, am *table.TemplateSetAttachment) error {
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

// validateTemplatesExist validate if all templates resource exists before operating template set
func (dao *templateSetDao) validateTemplatesExist(kit *kit.Kit, templateIDs []uint32) error {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)
	var existIDs []uint32
	if err := q.Where(m.ID.In(templateIDs...)).Pluck(m.ID, &existIDs); err != nil {
		return fmt.Errorf("validate templates exist failed, err: %v", err)
	}

	diffIDs := tools.SliceDiff(templateIDs, existIDs)
	if len(diffIDs) > 0 {
		return fmt.Errorf("template id in %v is not exist", diffIDs)
	}

	return nil
}
