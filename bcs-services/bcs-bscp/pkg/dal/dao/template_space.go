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
	"fmt"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/types"
)

// TemplateSpace supplies all the template space related operations.
type TemplateSpace interface {
	// Create one template space instance.
	Create(kit *kit.Kit, templateSpace *table.TemplateSpace) (uint32, error)
	// Update one template space's info.
	Update(kit *kit.Kit, templateSpace *table.TemplateSpace) error
	// List template spaces with options.
	List(kit *kit.Kit, bizID uint32, opt *types.BasePage) ([]*table.TemplateSpace, int64, error)
	// Delete one template space instance.
	Delete(kit *kit.Kit, templateSpace *table.TemplateSpace) error
	// GetByUniqueKey get template space by unique key.
	GetByUniqueKey(kit *kit.Kit, bizID uint32, name string) (*table.TemplateSpace, error)
	// GetAllBizs get all biz ids of template spaces.
	GetAllBizs(kit *kit.Kit) ([]uint32, error)
	// CreateDefault create default template space instance together with its default template set instance
	CreateDefault(kit *kit.Kit, bizID uint32) (uint32, error)
}

var _ TemplateSpace = new(templateSpaceDao)

type templateSpaceDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one template space instance.
// Every template space must have one default template set, so they should be created together.
func (dao *templateSpaceDao) Create(kit *kit.Kit, g *table.TemplateSpace) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	tmplSpaceID, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = tmplSpaceID

	sg := &table.TemplateSet{
		Spec: &table.TemplateSetSpec{
			Name:   "默认套餐",
			Memo:   "当前空间下的所有模版",
			Public: true,
		},
		Attachment: &table.TemplateSetAttachment{
			BizID:           g.Attachment.BizID,
			TemplateSpaceID: g.ID,
		},
		Revision: &table.Revision{
			Creator: g.Revision.Creator,
			Reviser: g.Revision.Reviser,
		},
	}
	tmplSetID, err := dao.idGen.One(kit, table.Name(sg.TableName()))
	if err != nil {
		return 0, err
	}
	sg.ID = tmplSetID

	tmplSpaceAD := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)
	tmplSetAD := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(sg)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		if err := tx.TemplateSpace.WithContext(kit.Ctx).Create(g); err != nil {
			return err
		}

		// 连带创建模版空间下的默认套餐
		if err := tx.TemplateSet.WithContext(kit.Ctx).Create(sg); err != nil {
			return err
		}

		if err := tmplSpaceAD.Do(tx); err != nil {
			return err
		}
		if err := tmplSetAD.Do(tx); err != nil {
			return err
		}

		return nil
	}
	if err := dao.genQ.Transaction(createTx); err != nil {
		return 0, err
	}

	return g.ID, nil
}

// Update one template space instance.
func (dao *templateSpaceDao) Update(kit *kit.Kit, g *table.TemplateSpace) error {
	if err := g.ValidateUpdate(); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := dao.genQ.TemplateSpace
	q := dao.genQ.TemplateSpace.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.TemplateSpace.WithContext(kit.Ctx)
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

// List template spaces with options.
func (dao *templateSpaceDao) List(kit *kit.Kit, bizID uint32, opt *types.BasePage) ([]*table.TemplateSpace, int64, error) {
	m := dao.genQ.TemplateSpace
	q := dao.genQ.TemplateSpace.WithContext(kit.Ctx)

	result, count, err := q.Where(m.BizID.Eq(bizID)).FindByPage(opt.Offset(), opt.LimitInt())
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

// Delete one template space instance.
func (dao *templateSpaceDao) Delete(kit *kit.Kit, g *table.TemplateSpace) error {
	// 参数校验
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.TemplateSpace
	q := dao.genQ.TemplateSpace.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.TemplateSpace.WithContext(kit.Ctx)
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

// GetAllBizs get all bizs of template spaces.
func (dao *templateSpaceDao) GetAllBizs(kit *kit.Kit) ([]uint32, error) {
	m := dao.genQ.TemplateSpace
	q := dao.genQ.TemplateSpace.WithContext(kit.Ctx)
	var bizIDs []uint32

	if err := q.Distinct(m.BizID).Pluck(m.ID, &bizIDs); err != nil {
		return nil, err
	}

	return bizIDs, nil
}

// CreateDefault create default template space instance together with its default template set instance
func (dao *templateSpaceDao) CreateDefault(kit *kit.Kit, bizID uint32) (uint32, error) {
	g := &table.TemplateSpace{
		ID: 0,
		Spec: &table.TemplateSpaceSpec{
			Name: "默认空间",
			Memo: "这是默认空间",
		},
		Attachment: &table.TemplateSpaceAttachment{
			BizID: kit.BizID,
		},
		Revision: &table.Revision{
			Creator: kit.User,
			Reviser: kit.User,
		},
	}
	tmplSpaceID, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = tmplSpaceID

	sg := &table.TemplateSet{
		Spec: &table.TemplateSetSpec{
			Name:   "默认套餐",
			Memo:   "当前空间下的所有模版",
			Public: true,
		},
		Attachment: &table.TemplateSetAttachment{
			BizID:           g.Attachment.BizID,
			TemplateSpaceID: g.ID,
		},
		Revision: &table.Revision{
			Creator: g.Revision.Creator,
			Reviser: g.Revision.Reviser,
		},
	}
	tmplSetID, err := dao.idGen.One(kit, table.Name(sg.TableName()))
	if err != nil {
		return 0, err
	}
	sg.ID = tmplSetID

	tmplSpaceAD := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)
	tmplSetAD := dao.auditDao.DecoratorV2(kit, sg.Attachment.BizID).PrepareCreate(sg)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		if err := tx.TemplateSpace.WithContext(kit.Ctx).Create(g); err != nil {
			return err
		}

		// 连带创建模版空间下的默认套餐
		if err := tx.TemplateSet.WithContext(kit.Ctx).Create(sg); err != nil {
			return err
		}

		if err := tmplSpaceAD.Do(tx); err != nil {
			return err
		}
		if err := tmplSetAD.Do(tx); err != nil {
			return err
		}

		return nil
	}
	if err := dao.genQ.Transaction(createTx); err != nil {
		return 0, err
	}

	return g.ID, nil
}

// GetByUniqueKey get template space by unique key
func (dao *templateSpaceDao) GetByUniqueKey(kit *kit.Kit, bizID uint32, name string) (*table.TemplateSpace, error) {
	m := dao.genQ.TemplateSpace
	q := dao.genQ.TemplateSpace.WithContext(kit.Ctx)

	templateSpace, err := q.Where(m.BizID.Eq(bizID), m.Name.Eq(name)).Take()
	if err != nil {
		return nil, fmt.Errorf("get template space failed, err: %v", err)
	}

	return templateSpace, nil
}
