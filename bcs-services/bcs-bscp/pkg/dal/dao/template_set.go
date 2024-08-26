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

	"gorm.io/datatypes"
	rawgen "gorm.io/gen"
	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	dtypes "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/search"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// TemplateSet supplies all the template set related operations.
type TemplateSet interface {
	// Create one template set instance.
	Create(kit *kit.Kit, templateSet *table.TemplateSet) (uint32, error)
	// Update one template set's info.
	Update(kit *kit.Kit, templateSet *table.TemplateSet) error
	// UpdateWithTx update one template set's info with transaction.
	UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, templateSet *table.TemplateSet) error
	// List template sets with options.
	List(kit *kit.Kit, bizID, templateSpaceID uint32, s search.Searcher, opt *types.BasePage) (
		[]*table.TemplateSet, int64, error)
	// Delete one template set instance.
	Delete(kit *kit.Kit, templateSet *table.TemplateSet) error
	// DeleteWithTx delete one template set instance with transaction.
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, templateSet *table.TemplateSet) error
	// GetByUniqueKey get template set by unique key.
	GetByUniqueKey(kit *kit.Kit, bizID, templateSpaceID uint32, name string) (*table.TemplateSet, error)
	// GetByUniqueKeyForUpdate get template set by unique key for update which allow to update name.
	GetByUniqueKeyForUpdate(kit *kit.Kit, bizID, templateSpaceID, selfID uint32, name string) (
		*table.TemplateSet, error)
	// ListByIDs list template sets by template set ids.
	ListByIDs(kit *kit.Kit, ids []uint32) ([]*table.TemplateSet, error)
	// ListByIDsWithTx list template sets by template set ids with transaction.
	ListByIDsWithTx(kit *kit.Kit, tx *gen.QueryTx, ids []uint32) ([]*table.TemplateSet, error)
	// AddTmplsToTmplSetsWithTx add templates to template sets with transaction.
	AddTmplsToTmplSetsWithTx(kit *kit.Kit, tx *gen.QueryTx, tmplIDs []uint32, tmplSetIDs []uint32) error
	// DeleteTmplsFromTmplSetsWithTx delete templates from template sets with transaction.
	DeleteTmplsFromTmplSetsWithTx(kit *kit.Kit, tx *gen.QueryTx, tmplIDs []uint32, tmplSetIDs []uint32) error
	// AddTmplToTmplSetsWithTx add a template to template sets with transaction.
	AddTmplToTmplSetsWithTx(kit *kit.Kit, tx *gen.QueryTx, tmplID uint32, tmplSetIDs []uint32) error
	// DeleteTmplFromAllTmplSetsWithTx delete a template from all template sets with transaction.
	DeleteTmplFromAllTmplSetsWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, tmplID uint32) error
	// ListAppTmplSets list all the template sets of the app.
	ListAppTmplSets(kit *kit.Kit, bizID, appID uint32) ([]*table.TemplateSet, error)
	// ListAllTemplateIDs list all template ids of all template sets in one template space.
	ListAllTemplateIDs(kit *kit.Kit, bizID, templateSpaceID uint32) ([]uint32, error)
	// ListAllTmplSetsOfBiz list all template sets of one biz
	ListAllTmplSetsOfBiz(kit *kit.Kit, bizID, appID uint32) ([]*table.TemplateSet, error)
	// ValidateTmplNumber verify whether the current number of template set's templates has reached the maximum.
	ValidateTmplNumber(kt *kit.Kit, tx *gen.QueryTx, bizID, tmplSetID uint32) error
	// ValidateWillExceedMaxTmplCount 给定一个数 和当前数量相加, 判断是否超过最大限制
	ValidateWillExceedMaxTmplCount(kt *kit.Kit, tx *gen.QueryTx, bizID,
		tmplSetID uint32, number int) error
	// GetByTemplateSetByID get template set by id
	GetByTemplateSetByID(kit *kit.Kit, bizID, id uint32) (*table.TemplateSet, error)
	// BatchAddTmplsToTmplSetsWithTx 批量添加至某个套餐中
	BatchAddTmplsToTmplSetsWithTx(kit *kit.Kit, tx *gen.QueryTx, templateSet []*table.TemplateSet) error
	// ListByTemplateSpaceIdAndIds list template sets by template set ids and template_space_id.
	ListByTemplateSpaceIdAndIds(kit *kit.Kit, templateSpaceID uint32, ids []uint32) ([]*table.TemplateSet, error)
}

var _ TemplateSet = new(templateSetDao)

type templateSetDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// ListByTemplateSpaceIdAndIds list template sets by template set ids and template_space_id.
func (dao *templateSetDao) ListByTemplateSpaceIdAndIds(kit *kit.Kit, templateSpaceID uint32,
	ids []uint32) ([]*table.TemplateSet, error) {

	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)

	return q.Where(m.TemplateSpaceID.Eq(templateSpaceID), m.ID.In(ids...)).Find()
}

// BatchAddTmplsToTmplSetsWithTx 批量添加至某个套餐中
func (dao *templateSetDao) BatchAddTmplsToTmplSetsWithTx(kit *kit.Kit, tx *gen.QueryTx,
	templateSet []*table.TemplateSet) error {
	if len(templateSet) == 0 {
		return nil
	}

	return tx.TemplateSet.WithContext(kit.Ctx).Save(templateSet...)
}

// GetByTemplateSetByID get template set by id
func (dao *templateSetDao) GetByTemplateSetByID(kit *kit.Kit, bizID uint32, id uint32) (
	*table.TemplateSet, error) {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)

	return q.Where(m.BizID.Eq(bizID), m.ID.Eq(id)).Take()
}

// Create one template set instance.
func (dao *templateSetDao) Create(kit *kit.Kit, g *table.TemplateSet) (uint32, error) {
	if err := g.ValidateCreate(kit); err != nil {
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
	if err := g.ValidateUpdate(kit); err != nil {
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
		if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
			Select(m.Name, m.Memo, m.Reviser, m.Public, m.BoundApps).
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

// UpdateWithTx update one template set's info with transaction.
func (dao *templateSetDao) UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.TemplateSet) error {
	if err := g.ValidateUpdate(kit); err != nil {
		return err
	}

	if err := dao.validateAttachmentExist(kit, g.Attachment); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := tx.TemplateSet
	q := tx.TemplateSet.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)
	if err := ad.Do(tx.Query); err != nil {
		return err
	}

	if len(g.Spec.TemplateIDs) == 0 {
		g.Spec.TemplateIDs = []uint32{}
	}

	if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
		Select(m.Name, m.Memo, m.Reviser, m.TemplateIDs, m.Public, m.BoundApps).
		Updates(g); err != nil {
		return err
	}

	return nil
}

// List template sets with options.
func (dao *templateSetDao) List(kit *kit.Kit, bizID, templateSpaceID uint32, s search.Searcher, opt *types.BasePage) (
	[]*table.TemplateSet, int64, error) {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)

	var conds []rawgen.Condition
	// add search condition
	if s != nil {
		exprs := s.SearchExprs(dao.genQ)
		if len(exprs) > 0 {
			var do gen.ITemplateSetDo
			for i := range exprs {
				if i == 0 {
					do = q.Where(exprs[i])
				}
				do = do.Or(exprs[i])
			}
			conds = append(conds, do)
		}
	}

	d := q.Where(m.BizID.Eq(bizID), m.TemplateSpaceID.Eq(templateSpaceID)).Where(conds...).Order(m.Name)
	if opt.All {
		result, err := d.Find()
		if err != nil {
			return nil, 0, err
		}
		return result, int64(len(result)), err
	}

	return d.FindByPage(opt.Offset(), opt.LimitInt())
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

	// 删除操作, 获取当前记录做审计
	m := tx.TemplateSet
	q := tx.TemplateSet.WithContext(kit.Ctx)
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

// GetByUniqueKey get template set by unique key
func (dao *templateSetDao) GetByUniqueKey(kit *kit.Kit, bizID, templateSpaceID uint32, name string) (
	*table.TemplateSet, error) {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)

	tplSet, err := q.Where(m.BizID.Eq(bizID), m.TemplateSpaceID.Eq(templateSpaceID), m.Name.Eq(name)).Take()
	if err != nil {
		return nil, fmt.Errorf("get template space failed, err: %v", err)
	}

	return tplSet, nil
}

// GetByUniqueKeyForUpdate get template set by unique key for update which allow to update name.
func (dao *templateSetDao) GetByUniqueKeyForUpdate(kit *kit.Kit, bizID, templateSpaceID, selfID uint32,
	name string) (*table.TemplateSet, error) {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)

	tplSet, err := q.Where(m.BizID.Eq(bizID), m.TemplateSpaceID.Eq(templateSpaceID),
		m.ID.Neq(selfID), m.Name.Eq(name)).Take()
	if err != nil {
		return nil, fmt.Errorf("get template space failed, err: %v", err)
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

// ListByIDsWithTx list template sets by template set ids with transaction.
func (dao *templateSetDao) ListByIDsWithTx(kit *kit.Kit, tx *gen.QueryTx, ids []uint32) ([]*table.TemplateSet, error) {
	m := tx.TemplateSet
	q := tx.TemplateSet.WithContext(kit.Ctx)
	result, err := q.Where(m.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}

	return result, nil
}

// AddTmplsToTmplSetsWithTx add templates to template sets with transaction.
func (dao *templateSetDao) AddTmplsToTmplSetsWithTx(kit *kit.Kit, tx *gen.QueryTx, tmplIDs []uint32,
	tmplSetIDs []uint32) error {
	m := tx.TemplateSet
	q := tx.TemplateSet.WithContext(kit.Ctx)
	for _, tmplID := range tmplIDs {
		if _, err := q.Where(m.ID.In(tmplSetIDs...)).
			Not(rawgen.Cond(datatypes.JSONArrayQuery("template_ids").Contains(tmplID))...).
			Update(m.TemplateIDs, gorm.Expr("JSON_ARRAY_APPEND(template_ids, '$', ?)", tmplID)); err != nil {
			return err
		}
	}
	return nil
}

// DeleteTmplsFromTmplSetsWithTx delete templates from template sets with transaction.
func (dao *templateSetDao) DeleteTmplsFromTmplSetsWithTx(kit *kit.Kit, tx *gen.QueryTx, tmplIDs,
	tmplSetIDs []uint32) error {
	m := tx.TemplateSet
	q := tx.TemplateSet.WithContext(kit.Ctx)
	for _, tmplID := range tmplIDs {
		// subQuery get the array of template ids after delete the target template id, set it to '[]' if no records found
		subQuery := "COALESCE ((SELECT JSON_ARRAYAGG(oid) new_oids FROM " +
			"JSON_TABLE (template_ids, '$[*]' COLUMNS (oid BIGINT (1) UNSIGNED PATH '$')) AS t1 WHERE oid<> ?), '[]')"
		if _, err := q.Where(m.ID.In(tmplSetIDs...)).
			Where(rawgen.Cond(datatypes.JSONArrayQuery("template_ids").Contains(tmplID))...).
			Update(m.TemplateIDs, gorm.Expr(subQuery, tmplID)); err != nil {
			return err
		}
	}
	return nil
}

// AddTmplToTmplSetsWithTx add a template to template sets with transaction.
func (dao *templateSetDao) AddTmplToTmplSetsWithTx(
	kit *kit.Kit, tx *gen.QueryTx, tmplID uint32, tmplSetIDs []uint32) error {
	m := tx.TemplateSet
	q := tx.TemplateSet.WithContext(kit.Ctx)
	if _, err := q.Where(m.ID.In(tmplSetIDs...)).
		Not(rawgen.Cond(datatypes.JSONArrayQuery("template_ids").Contains(tmplID))...).
		Update(m.TemplateIDs, gorm.Expr("JSON_ARRAY_APPEND(template_ids, '$', ?)", tmplID)); err != nil {
		return err
	}
	return nil
}

// DeleteTmplFromAllTmplSetsWithTx delete a template from all template sets with transaction.
func (dao *templateSetDao) DeleteTmplFromAllTmplSetsWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, tmplID uint32) error {
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

// ListAppTmplSets list all the template sets of the app.
func (dao *templateSetDao) ListAppTmplSets(kit *kit.Kit, bizID, appID uint32) ([]*table.TemplateSet, error) {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)

	return q.Where(m.BizID.Eq(bizID)).
		Where(m.Public.Is(true)).
		Or(rawgen.Cond(datatypes.JSONArrayQuery("bound_apps").Contains(appID))...).
		Order(m.Name).
		Find()
}

// ListAllTemplateIDs list all template ids of all template sets in one template space.
func (dao *templateSetDao) ListAllTemplateIDs(kit *kit.Kit, bizID, templateSpaceID uint32) ([]uint32, error) {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx)

	var result []dtypes.Uint32Slice
	if err := q.Select(m.TemplateIDs).
		Where(m.BizID.Eq(bizID), m.TemplateSpaceID.Eq(templateSpaceID)).
		Pluck(m.TemplateIDs, &result); err != nil {
		return nil, err
	}

	idMap := make(map[uint32]struct{})
	for _, ids := range result {
		for _, id := range ids {
			idMap[id] = struct{}{}
		}
	}

	ids := make([]uint32, 0, len(idMap))
	for id := range idMap {
		ids = append(ids, id)
	}

	return ids, nil
}

// ListAllTmplSetsOfBiz list all template sets of one biz
func (dao *templateSetDao) ListAllTmplSetsOfBiz(kit *kit.Kit, bizID, appID uint32) ([]*table.TemplateSet, error) {
	m := dao.genQ.TemplateSet
	q := dao.genQ.TemplateSet.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID))
	// if appID > 0 , return the app's all visible template sets
	if appID > 0 {
		q = q.Where(m.Public.Is(true)).
			Or(rawgen.Cond(datatypes.JSONArrayQuery("bound_apps").Contains(appID))...)
		return q.Where(m.BizID.Eq(bizID)).Find()
	}
	return q.Find()
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
	// allow template ids to be empty
	if len(templateIDs) == 0 {
		return nil
	}

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

// ValidateTmplNumber verify whether the current number of template set's templates has reached the maximum.
func (dao *templateSetDao) ValidateTmplNumber(kt *kit.Kit, tx *gen.QueryTx, bizID, tmplSetID uint32) error {

	// get template count
	m := tx.TemplateSet
	tmplSet, err := m.WithContext(kt.Ctx).Where(m.BizID.Eq(bizID), m.ID.Eq(tmplSetID)).Take()
	if err != nil {
		return errf.New(errf.InvalidParameter,
			fmt.Sprintf("get template set %d's failed, err: %v", tmplSetID, err))
	}

	count := len(tmplSet.Spec.TemplateIDs)
	tmplSetTmplCnt := getTmplSetTmplCnt(bizID)
	if count > tmplSetTmplCnt {
		return errf.New(errf.InvalidParameter,
			i18n.T(kt, "the total number of template set %d's templates exceeded the limit %d", tmplSetID, tmplSetTmplCnt))
	}

	return nil
}

// ValidateWillExceedMaxTmplCount 给定一个数 和当前数量相加, 判断是否超过最大限制
func (dao *templateSetDao) ValidateWillExceedMaxTmplCount(kt *kit.Kit, tx *gen.QueryTx, bizID,
	tmplSetID uint32, number int) error {

	// get template count
	m := tx.TemplateSet
	tmplSet, err := m.WithContext(kt.Ctx).Where(m.BizID.Eq(bizID), m.ID.Eq(tmplSetID)).Take()
	if err != nil {
		return fmt.Errorf("get template set %d's failed, err: %v", tmplSetID, err)
	}

	count := len(tmplSet.Spec.TemplateIDs) + number
	tmplSetTmplCnt := getTmplSetTmplCnt(bizID)
	if count > tmplSetTmplCnt {
		return errf.New(errf.InvalidParameter,
			i18n.T(kt, "the total number of template set %d's templates exceeded the limit %d", tmplSetID, tmplSetTmplCnt))
	}

	return nil
}

func getTmplSetTmplCnt(bizID uint32) int {
	if resLimit, ok := cc.DataService().FeatureFlags.ResourceLimit.Spec[fmt.Sprintf("%d", bizID)]; ok {
		if resLimit.TmplSetTmplCnt > 0 {
			return int(resLimit.TmplSetTmplCnt)
		}
	}
	return int(cc.DataService().FeatureFlags.ResourceLimit.Default.TmplSetTmplCnt)
}
