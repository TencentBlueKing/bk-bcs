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

	rawgen "gorm.io/gen"
	"gorm.io/gorm"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/search"
	"bscp.io/pkg/types"
)

// TemplateRevision supplies all the template revision related operations.
type TemplateRevision interface {
	// Create one template revision instance.
	Create(kit *kit.Kit, templateRevision *table.TemplateRevision) (uint32, error)
	// CreateWithTx create one template revision instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, template *table.TemplateRevision) (uint32, error)
	// List templates with options.
	List(kit *kit.Kit, bizID, templateID uint32, s search.Searcher, opt *types.BasePage) ([]*table.TemplateRevision, int64, error)
	// Delete one template revision instance.
	Delete(kit *kit.Kit, templateRevision *table.TemplateRevision) error
	// GetByUniqueKey get template revision by unique key.
	GetByUniqueKey(kit *kit.Kit, bizID, templateID uint32, revisionName string) (*table.TemplateRevision, error)
	// ListByIDs list template revisions by template revision ids.
	ListByIDs(kit *kit.Kit, ids []uint32) ([]*table.TemplateRevision, error)
	// DeleteForTmplWithTx delete template revision for one template with transaction.
	DeleteForTmplWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, templateID uint32) error
}

var _ TemplateRevision = new(templateRevisionDao)

type templateRevisionDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one template revision instance.
func (dao *templateRevisionDao) Create(kit *kit.Kit, g *table.TemplateRevision) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
		return 0, err
	}

	// generate a TemplateRevision id and update to TemplateRevision.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.TemplateRevision.WithContext(kit.Ctx)
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

// CreateWithTx create one template revision instance with transaction.
func (dao *templateRevisionDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.TemplateRevision) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate a TemplateRevision id and update to TemplateRevision.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	q := tx.TemplateRevision.WithContext(kit.Ctx)
	if err := q.Create(g); err != nil {
		return 0, err
	}

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)
	if err := ad.Do(tx.Query); err != nil {
		return 0, err
	}

	return g.ID, nil
}

// List template revisions with options.
func (dao *templateRevisionDao) List(kit *kit.Kit, bizID, templateID uint32, s search.Searcher, opt *types.BasePage) (
	[]*table.TemplateRevision, int64, error) {
	m := dao.genQ.TemplateRevision
	q := dao.genQ.TemplateRevision.WithContext(kit.Ctx)

	var conds []rawgen.Condition
	// add search condition
	exprs := s.SearchExprs(dao.genQ)
	if len(exprs) > 0 {
		var do gen.ITemplateRevisionDo
		for i := range exprs {
			if i == 0 {
				do = q.Where(exprs[i])
			}
			do = do.Or(exprs[i])
		}
		conds = append(conds, do)
	}

	d := q.Where(m.BizID.Eq(bizID), m.TemplateID.Eq(templateID)).Where(conds...)
	if opt.All {
		result, err := d.Find()
		if err != nil {
			return nil, 0, err
		}
		return result, int64(len(result)), err
	}

	return d.FindByPage(opt.Offset(), opt.LimitInt())
}

// Delete one template revision instance.
func (dao *templateRevisionDao) Delete(kit *kit.Kit, g *table.TemplateRevision) error {
	// 参数校验
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.TemplateRevision
	q := dao.genQ.TemplateRevision.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.TemplateRevision.WithContext(kit.Ctx)
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

// GetByUniqueKey get template revision by unique key
func (dao *templateRevisionDao) GetByUniqueKey(kit *kit.Kit, bizID, templateID uint32, revisionName string) (
	*table.TemplateRevision, error) {
	m := dao.genQ.TemplateRevision
	q := dao.genQ.TemplateRevision.WithContext(kit.Ctx)

	templateRevision, err := q.Where(m.BizID.Eq(bizID), m.TemplateID.Eq(templateID),
		m.RevisionName.Eq(revisionName)).Take()
	if err != nil {
		return nil, fmt.Errorf("get template revision failed, err: %v", err)
	}

	return templateRevision, nil
}

// ListByIDs list template revisions by template revision ids.
func (dao *templateRevisionDao) ListByIDs(kit *kit.Kit, ids []uint32) ([]*table.TemplateRevision, error) {
	m := dao.genQ.TemplateRevision
	q := dao.genQ.TemplateRevision.WithContext(kit.Ctx)
	result, err := q.Where(m.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// DeleteForTmplWithTx delete template revision for one template with transaction.
func (dao *templateRevisionDao) DeleteForTmplWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, templateID uint32) error {
	m := tx.TemplateRevision
	q := tx.TemplateRevision.WithContext(kit.Ctx)
	if _, err := q.Where(m.BizID.Eq(bizID), m.TemplateID.Eq(templateID)).Delete(); err != nil {
		return err
	}
	return nil
}

// validateAttachmentExist validate if attachment resource exists before operating template revision
func (dao *templateRevisionDao) validateAttachmentExist(kit *kit.Kit, am *table.TemplateRevisionAttachment) error {
	m := dao.genQ.Template
	q := dao.genQ.Template.WithContext(kit.Ctx)

	if _, err := q.Where(m.ID.Eq(am.TemplateID)).Take(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("template revision attached template %d is not exist", am.TemplateID)
		}
		return fmt.Errorf("get template revision attached template failed, err: %v", err)
	}

	return nil
}
