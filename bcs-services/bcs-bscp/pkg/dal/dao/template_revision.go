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
	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/search"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// TemplateRevision supplies all the template revision related operations.
type TemplateRevision interface {
	// Create one template revision instance.
	Create(kit *kit.Kit, templateRevision *table.TemplateRevision) (uint32, error)
	// CreateWithTx create one template revision instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, template *table.TemplateRevision) (uint32, error)
	// List templates with options.
	List(kit *kit.Kit, bizID, templateID uint32, s search.Searcher, opt *types.BasePage) ([]*table.TemplateRevision,
		int64, error)
	// Delete one template revision instance.
	Delete(kit *kit.Kit, templateRevision *table.TemplateRevision) error
	// GetByUniqueKey get template revision by unique key.
	GetByUniqueKey(kit *kit.Kit, bizID, templateID uint32, revisionName string) (*table.TemplateRevision, error)
	// ListByIDs list template revisions by template revision ids.
	ListByIDs(kit *kit.Kit, ids []uint32) ([]*table.TemplateRevision, error)
	// ListByIDsWithTx list template revisions by template revision ids with transaction.
	ListByIDsWithTx(kit *kit.Kit, tx *gen.QueryTx, ids []uint32) ([]*table.TemplateRevision, error)
	// ListByTemplateIDs list template revisions by template ids.
	ListByTemplateIDs(kit *kit.Kit, bizID uint32, templateIDs []uint32) ([]*table.TemplateRevision, error)
	// ListByTemplateIDsWithTx list template revisions by template ids with transaction.
	ListByTemplateIDsWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID uint32, templateIDs []uint32) (
		[]*table.TemplateRevision, error)
	// DeleteForTmplWithTx delete template revision for one template with transaction.
	DeleteForTmplWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, templateID uint32) error
	// BatchCreateWithTx batch create template revisions instances with transaction.
	BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx, revisions []*table.TemplateRevision) error
	// ListLatestRevisionsGroupByTemplateIds Lists the latest version groups by template ids
	ListLatestRevisionsGroupByTemplateIds(kit *kit.Kit, templateIDs []uint32) ([]*table.TemplateRevision, error)
	// GetLatestTemplateRevision get latest template revision.
	GetLatestTemplateRevision(kit *kit.Kit, bizID, templateID uint32) (*table.TemplateRevision, error)
	// GetTemplateRevisionById get template revision by id.
	GetTemplateRevisionById(kit *kit.Kit, bizID, id uint32) (*table.TemplateRevision, error)
	// ListLatestGroupByTemplateIdsWithTx Lists the latest version groups by template ids with transaction.
	ListLatestGroupByTemplateIdsWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID uint32,
		templateIDs []uint32) ([]*table.TemplateRevision, error)
}

var _ TemplateRevision = new(templateRevisionDao)

type templateRevisionDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// ListLatestGroupByTemplateIdsWithTx Lists the latest version groups by template ids with transaction.
func (dao *templateRevisionDao) ListLatestGroupByTemplateIdsWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID uint32,
	templateIDs []uint32) ([]*table.TemplateRevision, error) {

	m := dao.genQ.TemplateRevision
	q := tx.TemplateRevision.WithContext(kit.Ctx)
	// 根据templateIDs获取一列最大 templateRevisionIDs
	// 再通过最大 templateRevisionIDs 获取 templateRevision 数据
	var templateRevisionIDs []struct{ Id uint32 }
	if err := q.Select(m.ID.Max().As("id")).
		Where(m.BizID.Eq(bizID), m.TemplateID.In(templateIDs...)).
		Group(m.TemplateID).
		Scan(&templateRevisionIDs); err != nil {
		return nil, err
	}
	ids := []uint32{}
	for _, item := range templateRevisionIDs {
		ids = append(ids, item.Id)
	}
	find, err := q.Where(m.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	return find, nil
}

// GetTemplateRevisionById get template revision by id.
func (dao *templateRevisionDao) GetTemplateRevisionById(kit *kit.Kit, bizID uint32, id uint32) (
	*table.TemplateRevision, error) {
	m := dao.genQ.TemplateRevision
	q := dao.genQ.TemplateRevision.WithContext(kit.Ctx)
	return q.Where(m.BizID.Eq(bizID), m.ID.Eq(id)).Take()
}

// GetLatestTemplateRevision get latest template revision.
func (dao *templateRevisionDao) GetLatestTemplateRevision(kit *kit.Kit, bizID uint32, templateID uint32) (
	*table.TemplateRevision, error) {
	m := dao.genQ.TemplateRevision
	q := dao.genQ.TemplateRevision.WithContext(kit.Ctx)
	return q.Where(m.BizID.Eq(bizID), m.TemplateID.Eq(templateID)).Order(m.ID.Desc()).Take()
}

// Create one template revision instance.
func (dao *templateRevisionDao) Create(kit *kit.Kit, g *table.TemplateRevision) (uint32, error) {
	if err := g.ValidateCreate(kit); err != nil {
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
	if err := g.ValidateCreate(kit); err != nil {
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
	if s != nil {
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
	}

	d := q.Where(m.BizID.Eq(bizID), m.TemplateID.Eq(templateID)).Where(conds...).Order(m.ID.Desc())
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

	return q.Where(m.BizID.Eq(bizID), m.TemplateID.Eq(templateID),
		m.RevisionName.Eq(revisionName)).Take()
}

// ListByIDs list template revisions by template revision ids.
func (dao *templateRevisionDao) ListByIDs(kit *kit.Kit, ids []uint32) ([]*table.TemplateRevision, error) {
	m := dao.genQ.TemplateRevision
	q := dao.genQ.TemplateRevision.WithContext(kit.Ctx)
	return q.Where(m.ID.In(ids...)).Find()
}

// ListByIDsWithTx list template revisions by template revision ids with transaction.
func (dao *templateRevisionDao) ListByIDsWithTx(kit *kit.Kit, tx *gen.QueryTx, ids []uint32) (
	[]*table.TemplateRevision, error) {
	m := tx.TemplateRevision
	q := tx.TemplateRevision.WithContext(kit.Ctx)
	return q.Where(m.ID.In(ids...)).Find()
}

// ListByTemplateIDs list template revisions by template ids.
func (dao *templateRevisionDao) ListByTemplateIDs(kit *kit.Kit, bizID uint32, templateIDs []uint32) (
	[]*table.TemplateRevision, error) {
	m := dao.genQ.TemplateRevision
	q := dao.genQ.TemplateRevision.WithContext(kit.Ctx)
	return q.Where(m.BizID.Eq(bizID), m.TemplateID.In(templateIDs...)).Order(m.ID.Desc()).Find()
}

// ListByTemplateIDsWithTx list template revisions by template ids with transaction.
func (dao *templateRevisionDao) ListByTemplateIDsWithTx(
	kit *kit.Kit, tx *gen.QueryTx, bizID uint32, templateIDs []uint32) ([]*table.TemplateRevision, error) {
	m := tx.TemplateRevision
	q := tx.TemplateRevision.WithContext(kit.Ctx)
	return q.Where(m.BizID.Eq(bizID), m.TemplateID.In(templateIDs...)).Order(m.ID.Desc()).Find()
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

// BatchCreateWithTx batch create template revision instances with transaction.
func (dao *templateRevisionDao) BatchCreateWithTx(kit *kit.Kit, tx *gen.QueryTx,
	revisions []*table.TemplateRevision) error {
	if len(revisions) == 0 {
		return nil
	}
	ids, err := dao.idGen.Batch(kit, table.TemplateRevisionsTable, len(revisions))
	if err != nil {
		return err
	}
	for i, item := range revisions {
		if err := item.ValidateCreate(kit); err != nil {
			return err
		}
		item.ID = ids[i]
	}
	return tx.Query.TemplateRevision.WithContext(kit.Ctx).CreateInBatches(revisions, 200)
}

// ListLatestRevisionsGroupByTemplateIds Lists the latest version groups by template ids
func (dao *templateRevisionDao) ListLatestRevisionsGroupByTemplateIds(kit *kit.Kit,
	templateIDs []uint32) ([]*table.TemplateRevision, error) {
	m := dao.genQ.TemplateRevision
	// 根据templateIDs获取一列最大 templateRevisionIDs
	// 再通过最大 templateRevisionIDs 获取 templateRevision 数据
	var templateRevisionIDs []struct{ Id uint32 }
	if err := m.WithContext(kit.Ctx).Select(m.ID.Max().As("id")).Where(m.TemplateID.In(templateIDs...)).
		Group(m.TemplateID).
		Scan(&templateRevisionIDs); err != nil {
		return nil, err
	}
	ids := []uint32{}
	for _, item := range templateRevisionIDs {
		ids = append(ids, item.Id)
	}
	find, err := m.WithContext(kit.Ctx).Where(m.ID.In(ids...)).Find()
	if err != nil {
		return nil, err
	}
	return find, nil
}
