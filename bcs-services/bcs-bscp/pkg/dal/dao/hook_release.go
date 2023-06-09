/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "as IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

package dao

import (
	"fmt"

	"gorm.io/gorm"

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/types"
)

// HookRelease supplies all the hook release related operations.
type HookRelease interface {
	// Create one hook release instance.
	Create(kit *kit.Kit, hook *table.HookRelease) (uint32, error)
	// CreateWithTx create hook release instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.HookRelease) (uint32, error)
	// Get hook release by id
	Get(kit *kit.Kit, bizID, hookID, id uint32) (*table.HookRelease, error)
	// GetByName get HookRelease by name
	GetByName(kit *kit.Kit, bizID, hookID uint32, name string) (*table.HookRelease, error)
	// List hooks with options.
	List(kit *kit.Kit, opt *types.ListHookReleasesOption) ([]*table.HookRelease, int64, error)
	// Delete one strategy instance.
	Delete(kit *kit.Kit, g *table.HookRelease) error
	// GetByPubState hook release by PubState
	GetByPubState(kit *kit.Kit, opt *types.GetByPubStateOption) (*table.HookRelease, error)
	// DeleteByHookIDWithTx  delete release revision with transaction
	DeleteByHookIDWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.HookRelease) error
	// PublishNumPlusOneWithTx PublishNum +1 revision with transaction
	PublishNumPlusOneWithTx(kit *kit.Kit, tx *gen.Query) error
	// UpdatePubStateWithTx update hookRelease PubState instance with transaction.
	UpdatePubStateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.HookRelease) error
}

var _ HookRelease = new(hookReleaseDao)

type hookReleaseDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one hook instance.
func (dao *hookReleaseDao) Create(kit *kit.Kit, g *table.HookRelease) (uint32, error) {

	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate a HookRelease id and update to HookRelease.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.HookRelease.WithContext(kit.Ctx)
		if e := q.Create(g); e != nil {
			return e
		}

		if e := ad.Do(tx); e != nil {
			return e
		}

		return nil
	}
	if e := dao.genQ.Transaction(createTx); e != nil {
		return 0, e
	}

	return g.ID, nil

}

// CreateWithTx create one hookRelease instance with transaction.
func (dao *hookReleaseDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.HookRelease) (uint32, error) {

	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	// generate a HookRelease id and update to HookRelease.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	err = tx.HookRelease.WithContext(kit.Ctx).Create(g)
	if err != nil {
		return 0, err
	}
	err = ad.Do(tx.Query)
	if err != nil {
		return 0, err
	}

	return g.ID, nil
}

// Get hookRelease by id
func (dao *hookReleaseDao) Get(kit *kit.Kit, bizID, hookID, id uint32) (*table.HookRelease, error) {

	m := dao.genQ.HookRelease
	q := dao.genQ.HookRelease.WithContext(kit.Ctx)

	tplSpace, err := q.Where(m.BizID.Eq(bizID), m.HookID.Eq(hookID), m.ID.Eq(id)).Take()
	if err != nil {
		return nil, fmt.Errorf("get hookRelease failed, err: %v", err)
	}

	return tplSpace, nil
}

// GetByName get HookRelease by name
func (dao *hookReleaseDao) GetByName(kit *kit.Kit, bizID, hookID uint32, name string) (*table.HookRelease, error) {
	m := dao.genQ.HookRelease
	q := dao.genQ.HookRelease.WithContext(kit.Ctx)

	tplSpace, err := q.Where(m.BizID.Eq(bizID), m.HookID.Eq(hookID), m.Name.Eq(name)).Take()
	if err != nil {
		return nil, fmt.Errorf("get hookRelease failed, err: %v", err)
	}

	return tplSpace, nil
}

// List hooks with options.
func (dao *hookReleaseDao) List(kit *kit.Kit,
	opt *types.ListHookReleasesOption) ([]*table.HookRelease, int64, error) {

	m := dao.genQ.HookRelease
	q := dao.genQ.HookRelease.WithContext(kit.Ctx).Where(
		m.BizID.Eq(opt.BizID),
		m.HookID.Eq(opt.HookID))

	if opt.SearchKey != "" {
		q = q.Where(m.Name.Like(fmt.Sprintf("%%%s%%", opt.SearchKey)))
	}

	if opt.Page.Start == 0 && opt.Page.Limit == 0 {
		result, err := q.Find()
		if err != nil {
			return nil, 0, err
		}

		return result, int64(len(result)), err

	} else {
		result, count, err := q.FindByPage(opt.Page.Offset(), opt.Page.LimitInt())
		if err != nil {
			return nil, 0, err
		}

		return result, count, err
	}

}

// Delete one strategy instance.
func (dao *hookReleaseDao) Delete(kit *kit.Kit, g *table.HookRelease) error {

	// 参数校验
	if err := g.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.HookRelease
	q := dao.genQ.HookRelease.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.HookRelease.WithContext(kit.Ctx)
		if _, e := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).Delete(g); e != nil {
			return e
		}

		if e := ad.Do(tx); e != nil {
			return e
		}
		return nil
	}
	if e := dao.genQ.Transaction(deleteTx); e != nil {
		return e
	}

	return nil
}

// DeleteByHookIDWithTx  delete release revision with transaction
func (dao *hookReleaseDao) DeleteByHookIDWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.HookRelease) error {

	// 参数校验
	if err := g.ValidateDeleteByHookID(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := tx.HookRelease
	q := tx.HookRelease.WithContext(kit.Ctx)

	oldOne, err := q.Where(m.HookID.Eq(g.Attachment.HookID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	if _, e := q.Where(m.BizID.Eq(g.Attachment.BizID), m.HookID.Eq(g.Attachment.HookID)).Delete(g); e != nil {
		return e
	}

	if e := ad.Do(tx.Query); e != nil {
		return err
	}

	return nil
}

// GetByPubState hook release by PubState
func (dao *hookReleaseDao) GetByPubState(kit *kit.Kit,
	opt *types.GetByPubStateOption) (*table.HookRelease, error) {

	// 参数校验
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	m := dao.genQ.HookRelease
	q := dao.genQ.HookRelease.WithContext(kit.Ctx)

	release, err := q.Where(
		m.BizID.Eq(opt.BizID),
		m.HookID.Eq(opt.HookID),
		m.PubState.Eq(opt.State.String()),
	).Take()
	if err != nil {
		return nil, err
	}

	return release, nil

}

// PublishNumPlusOneWithTx PublishNum +1 revision with transaction
func (dao *hookReleaseDao) PublishNumPlusOneWithTx(kit *kit.Kit, tx *gen.Query) error {

	m := tx.HookRelease
	_, err := tx.WithContext(kit.Ctx).Hook.Update(m.PublishNum, gorm.Expr("publish_num + ?", 1))
	return err
}

// UpdatePubStateWithTx update hookRelease PubState instance with transaction.
func (dao *hookReleaseDao) UpdatePubStateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.HookRelease) error {

	if err := g.ValidatePublish(); err != nil {
		return err
	}

	q := tx.HookRelease.WithContext(kit.Ctx)
	m := tx.HookRelease

	oldOne, err := q.Where(m.HookID.Eq(g.Attachment.HookID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	if _, e := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Select(m.PubState, m.Reviser).Updates(g); e != nil {
		return e
	}

	if e := ad.Do(tx.Query); e != nil {
		return e
	}

	return nil
}
