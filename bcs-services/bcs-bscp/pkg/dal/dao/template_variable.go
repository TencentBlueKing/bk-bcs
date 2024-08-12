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
	rawgen "gorm.io/gen"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/search"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// TemplateVariable supplies all the template variable related operations.
type TemplateVariable interface {
	// Create one template variable instance.
	Create(kit *kit.Kit, templateVariable *table.TemplateVariable) (uint32, error)
	// Update one template variable's info.
	Update(kit *kit.Kit, templateVariable *table.TemplateVariable) error
	// List template variables with options.
	List(kit *kit.Kit, bizID uint32, s search.Searcher, opt *types.BasePage) ([]*table.TemplateVariable, int64, error)
	// Delete one template variable instance.
	Delete(kit *kit.Kit, templateVariable *table.TemplateVariable) error
	// BatchCreateWithTx batch create variable instances with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, tmplVars []*table.TemplateVariable) error
	// BatchUpdateWithTx batch update variable instances with transaction.
	BatchUpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, tmplVars []*table.TemplateVariable) error
	// GetByUniqueKey get template variable by unique key.
	GetByUniqueKey(kit *kit.Kit, bizID uint32, name string) (*table.TemplateVariable, error)
	// FetchIDsExcluding 获取指定ID后排除的ID
	FetchIDsExcluding(kit *kit.Kit, bizID uint32, ids []uint32) ([]uint32, error)
}

var _ TemplateVariable = new(templateVariableDao)

type templateVariableDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// ListIdsExcluded 获取指定ID后排除的ID
func (dao *templateVariableDao) FetchIDsExcluding(kit *kit.Kit, bizID uint32, ids []uint32) ([]uint32, error) {

	m := dao.genQ.TemplateVariable
	q := dao.genQ.TemplateVariable.WithContext(kit.Ctx)

	var result []uint32
	if err := q.Select(m.ID).
		Where(m.BizID.Eq(bizID), m.ID.NotIn(ids...)).
		Pluck(m.ID, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Create one template variable instance.
func (dao *templateVariableDao) Create(kit *kit.Kit, g *table.TemplateVariable) (uint32, error) {
	if err := g.ValidateCreate(kit); err != nil {
		return 0, errf.ErrInvalidArgF(kit).WithCause(err)
	}

	tmplSpaceID, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, errf.ErrDBOpsFailedF(kit).WithCause(err)
	}
	g.ID = tmplSpaceID

	tmplSpaceAD := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		if err := tx.TemplateVariable.WithContext(kit.Ctx).Create(g); err != nil {
			return err
		}

		if err := tmplSpaceAD.Do(tx); err != nil {
			return err
		}
		return nil
	}
	if err := dao.genQ.Transaction(createTx); err != nil {
		return 0, errf.ErrDBOpsFailedF(kit).WithCause(err)
	}

	return g.ID, nil
}

// Update one template variable instance.
func (dao *templateVariableDao) Update(kit *kit.Kit, g *table.TemplateVariable) error {
	if err := g.ValidateUpdate(kit); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := dao.genQ.TemplateVariable
	q := dao.genQ.TemplateVariable.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.TemplateVariable.WithContext(kit.Ctx)
		if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
			Select(m.DefaultVal, m.Memo, m.Reviser).
			Updates(g); err != nil {
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

// List template variables with options.
func (dao *templateVariableDao) List(
	kit *kit.Kit, bizID uint32, s search.Searcher, opt *types.BasePage) ([]*table.TemplateVariable, int64, error) {
	m := dao.genQ.TemplateVariable
	q := dao.genQ.TemplateVariable.WithContext(kit.Ctx)

	var conds []rawgen.Condition
	// add search condition
	if s != nil {
		exprs := s.SearchExprs(dao.genQ)
		if len(exprs) > 0 {
			var do gen.ITemplateVariableDo
			for i := range exprs {
				if i == 0 {
					do = q.Where(exprs[i])
				}
				do = do.Or(exprs[i])
			}
			conds = append(conds, do)
		}
	}

	d := q.Where(m.BizID.Eq(bizID)).Where(conds...).Order(m.Name)
	if opt.All {
		result, err := d.Find()
		if err != nil {
			return nil, 0, err
		}
		return result, int64(len(result)), err
	}

	return d.FindByPage(opt.Offset(), opt.LimitInt())
}

// Delete one template variable instance.
func (dao *templateVariableDao) Delete(kit *kit.Kit, g *table.TemplateVariable) error {
	// 参数校验
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.TemplateVariable
	q := dao.genQ.TemplateVariable.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.TemplateVariable.WithContext(kit.Ctx)
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

// BatchCreateWithTx batch create variable instances with transaction.
// Note: batch operation won't audit.
func (dao *templateVariableDao) BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx,
	tmplVars []*table.TemplateVariable) error {
	if len(tmplVars) == 0 {
		return nil
	}
	ids, err := dao.idGen.Batch(kit, table.Name(tmplVars[0].TableName()), len(tmplVars))
	if err != nil {
		return err
	}
	for i, v := range tmplVars {
		if err := v.ValidateCreate(kit); err != nil {
			return err
		}
		v.ID = ids[i]
	}
	return tx.TemplateVariable.WithContext(kit.Ctx).Save(tmplVars...)
}

// BatchUpdateWithTx batch update variable instances with transaction.
// Note: batch operation won't audit.
func (dao *templateVariableDao) BatchUpdateWithTx(kit *kit.Kit, tx *gen.QueryTx,
	tmplVars []*table.TemplateVariable) error {
	if len(tmplVars) == 0 {
		return nil
	}
	for _, v := range tmplVars {
		if err := v.ValidateUpdate(kit); err != nil {
			return err
		}
		if err := v.Spec.Type.Validate(kit); err != nil {
			return err
		}
	}
	return tx.TemplateVariable.WithContext(kit.Ctx).Save(tmplVars...)
}

// GetByUniqueKey get template variable by unique key
func (dao *templateVariableDao) GetByUniqueKey(kit *kit.Kit, bizID uint32, name string) (*table.TemplateVariable,
	error) {
	m := dao.genQ.TemplateVariable
	q := dao.genQ.TemplateVariable.WithContext(kit.Ctx)
	return q.Where(m.BizID.Eq(bizID), m.Name.Eq(name)).Take()
}
