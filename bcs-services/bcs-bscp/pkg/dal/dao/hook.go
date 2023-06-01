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

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/types"
)

// Hook supplies all the hook related operations.
type Hook interface {
	// Create one hook instance.
	Create(kit *kit.Kit, hook *table.Hook, release *table.HookRelease) (uint32, error)
	// List hooks with options.
	List(kit *kit.Kit, opt *types.ListHooksOption) ([]*table.Hook, int64, error)
	// CountHookTag count hook tag
	CountHookTag(kit *kit.Kit, bizID uint32) ([]*types.HookTagCount, error)
	// Delete one strategy instance.
	Delete(kit *kit.Kit, strategy *table.Hook) error
	// GetByID get hook only with id.
	GetByID(kit *kit.Kit, bizID, hookID uint32) (*table.Hook, error)
	// GetByName get hook by name
	GetByName(kit *kit.Kit, bizID uint32, name string) (*table.Hook, error)
}

var _ Hook = new(hookDao)

type hookDao struct {
	genQ           *gen.Query
	idGen          IDGenInterface
	auditDao       AuditDao
	hookReleaseDao HookRelease
}

// Create one hook instance.
func (dao *hookDao) Create(kit *kit.Kit, g *table.Hook,
	release *table.HookRelease) (uint32, error) {

	if g == nil {
		return 0, errors.New("hook is nil")
	}

	if release == nil {
		return 0, errors.New("hook release is nil")
	}

	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	//generate a hook id and update to hook.
	id, err := dao.idGen.One(kit, table.HookTable)
	if err != nil {
		return 0, err
	}
	g.ID = id
	release.Attachment.HookID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.Hook.WithContext(kit.Ctx)
		if e := q.Create(g); e != nil {
			return e
		}

		_, err = dao.hookReleaseDao.CreateWithTx(kit, tx, release)
		if err != nil {
			return err
		}

		if e := ad.Do(tx); e != nil {
			return e
		}

		return nil
	}
	if e := dao.genQ.Transaction(createTx); e != nil {
		return 0, err
	}

	return g.ID, nil

}

// List hooks with options.
func (dao *hookDao) List(kit *kit.Kit, opt *types.ListHooksOption) ([]*table.Hook, int64, error) {

	m := dao.genQ.Hook
	q := dao.genQ.Hook.WithContext(kit.Ctx)

	q = q.Where(m.BizID.Eq(opt.BizID), m.Name.Like(fmt.Sprintf("%%%s%%", opt.Name)))
	if !opt.All {
		q = q.Where(m.Tag.Eq(opt.Tag))
	}

	result, count, err := q.FindByPage(opt.Page.Offset(), opt.Page.LimitInt())
	if err != nil {
		return nil, 0, err
	}

	return result, count, nil
}

// CountHookTag count hook tag
func (dao *hookDao) CountHookTag(kit *kit.Kit, bizID uint32) ([]*types.HookTagCount, error) {

	m := dao.genQ.Hook
	q := dao.genQ.Hook.WithContext(kit.Ctx)

	counts := make([]*types.HookTagCount, 0)
	err := q.Select(m.Tag, m.ID.Count().As("counts")).Where(m.BizID.Eq(bizID), m.Tag.Neq("")).Group(m.Tag).Scan(&counts)
	if err != nil {
		return nil, err
	}

	return counts, nil
}

// Delete one hook instance.
func (dao *hookDao) Delete(kit *kit.Kit, g *table.Hook) error {

	if g == nil {
		return errf.New(errf.InvalidParameter, "hook is nil")
	}

	if err := g.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	// 删除操作, 获取当前记录做审计
	m := dao.genQ.Hook
	q := dao.genQ.Hook.WithContext(kit.Ctx)

	hookRM := dao.genQ.HookRelease
	hookRQ := dao.genQ.HookRelease.WithContext(kit.Ctx)

	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	hookRelease := &table.HookRelease{
		Attachment: &table.HookReleaseAttachment{
			BizID:  g.Attachment.BizID,
			HookID: g.ID,
		},
	}

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.Hook.WithContext(kit.Ctx)
		if _, e := q.Where(m.BizID.Eq(g.Attachment.BizID)).Delete(g); e != nil {
			return e
		}

		hookRQ = tx.HookRelease.WithContext(kit.Ctx)
		if _, e := hookRQ.Where(hookRM.BizID.Eq(g.Attachment.BizID), hookRM.HookID.Eq(g.ID)).Delete(hookRelease); e != nil {
			return e
		}

		if err = dao.hookReleaseDao.DeleteByHookIDWithTx(kit, tx, hookRelease); err != nil {
			return err
		}

		if e := ad.Do(tx); e != nil {
			return e
		}
		return nil
	}
	err = dao.genQ.Transaction(deleteTx)
	if err != nil {
		return err
	}

	return nil
}

// GetByID get hook only with id.
func (dao *hookDao) GetByID(kit *kit.Kit, bizID, hookID uint32) (*table.Hook, error) {

	m := dao.genQ.Hook
	q := dao.genQ.Hook.WithContext(kit.Ctx)

	hook, err := q.Where(m.BizID.Eq(bizID), m.ID.Eq(hookID)).Take()
	if err != nil {
		return nil, err
	}

	return hook, nil
}

// GetByName get a Hook by name
func (dao *hookDao) GetByName(kit *kit.Kit, bizID uint32, name string) (*table.Hook, error) {
	m := dao.genQ.Hook
	q := dao.genQ.Hook.WithContext(kit.Ctx)

	hook, err := q.Where(m.BizID.Eq(bizID), m.Name.Eq(name)).Take()
	if err != nil {
		return nil, err
	}

	return hook, nil
}
