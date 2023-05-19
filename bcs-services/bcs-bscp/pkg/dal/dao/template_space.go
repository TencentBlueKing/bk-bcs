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
)

// TemplateSpace supplies all the TemplateSpace related operations.
type TemplateSpace interface {
	// Create one TemplateSpace instance.
	Create(kit *kit.Kit, TemplateSpace *table.TemplateSpace) (uint32, error)
	// Update one TemplateSpace's info.
	Update(kit *kit.Kit, TemplateSpace *table.TemplateSpace) error
	// List TemplateSpaces with options.
	List(kit *kit.Kit, bizID uint32, offset, limit int) ([]*table.TemplateSpace, int64, error)
	// Delete one strategy instance.
	Delete(kit *kit.Kit, strategy *table.TemplateSpace) error
	// GetByName get templateSpace by name.
	GetByName(kit *kit.Kit, bizID uint32, name string) (*table.TemplateSpace, error)
}

var _ TemplateSpace = new(templateSpaceDao)

type templateSpaceDao struct {
	genM     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one TemplateSpace instance.
func (dao *templateSpaceDao) Create(kit *kit.Kit, g *table.TemplateSpace) (uint32, error) {
	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate a TemplateSpace id and update to TemplateSpace.
	id, err := dao.idGen.One(kit, table.TemplateSpaceTable)
	if err != nil {
		return 0, err
	}
	g.ID = id

	ab := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理
	err = dao.genM.Transaction(func(tx *gen.Query) error {
		q := tx.TemplateSpace.WithContext(kit.Ctx)
		if err := q.Create(g); err != nil {
			return err
		}

		if err := ab.Do(tx); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return g.ID, nil
}

// Update one TemplateSpace instance.
func (dao *templateSpaceDao) Update(kit *kit.Kit, g *table.TemplateSpace) error {
	if err := g.ValidateUpdate(); err != nil {
		return err
	}

	m := dao.genM.TemplateSpace
	q := dao.genM.TemplateSpace.WithContext(kit.Ctx)

	// 更新记录审计, 查询上一个记录数据
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ab := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(oldOne, g)

	// 多个使用事务处理
	err = dao.genM.Transaction(func(tx *gen.Query) error {
		if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).Select(m.Memo, m.Reviser).Updates(g); err != nil {
			return err
		}

		if err := ab.Do(tx); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// List TemplateSpaces with options.
func (dao *templateSpaceDao) List(kit *kit.Kit, bizID uint32, offset, limit int) ([]*table.TemplateSpace, int64, error) {
	m := dao.genM.TemplateSpace
	q := dao.genM.TemplateSpace.WithContext(kit.Ctx)

	result, count, err := q.Where(m.BizID.Eq(bizID)).FindByPage(offset, limit)
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

// Delete one TemplateSpace instance.
func (dao *templateSpaceDao) Delete(kit *kit.Kit, g *table.TemplateSpace) error {
	// 参数校验
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genM.TemplateSpace
	q := dao.genM.TemplateSpace.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.TemplateSpace.WithContext(kit.Ctx)
		if _, err := q.Delete(g); err != nil {
			return err
		}

		if err := ad.Do(tx); err != nil {
			return err
		}
		return nil
	}
	if err := dao.genM.Transaction(deleteTx); err != nil {
		return err
	}

	return nil
}

// GetByName get by name
func (dao *templateSpaceDao) GetByName(kit *kit.Kit, bizID uint32, name string) (*table.TemplateSpace, error) {
	m := dao.genM.TemplateSpace
	q := dao.genM.TemplateSpace.WithContext(kit.Ctx)

	q.Where(m.Reviser)

	tplSpace, err := q.Where(m.BizID.Eq(bizID), m.Name.Eq(name)).Take()
	if err != nil {
		return nil, fmt.Errorf("get templateSpace failed, err: %v", err)
	}

	return tplSpace, nil
}
