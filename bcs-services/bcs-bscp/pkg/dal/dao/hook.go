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

	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
	"bscp.io/pkg/types"
)

// Hook supplies all the hook related operations.
type Hook interface {
	// CreateWithTx create one hook instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, hook *table.Hook) (uint32, error)
	// ListWithRefer hooks with refer info.
	ListWithRefer(kit *kit.Kit, opt *types.ListHooksWithReferOption) ([]*types.ListHooksWithReferDetail, int64, error)
	// ListHookReferences list hook references.
	ListHookReferences(kit *kit.Kit, opt *types.ListHookReferencesOption) (
		[]*types.ListHookReferencesDetail, int64, error)
	// CountHookTag count hook tag
	CountHookTag(kit *kit.Kit, bizID uint32) ([]*types.HookTagCount, error)
	// DeleteWithTx delete hook instance with transaction.
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.Hook) error
	// GetByID get hook only with id.
	GetByID(kit *kit.Kit, bizID, hookID uint32) (*table.Hook, error)
	// GetByName get hook by name
	GetByName(kit *kit.Kit, bizID uint32, name string) (*table.Hook, error)
}

var _ Hook = new(hookDao)

type hookDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// CreateWithTx create one hook instance with transaction.
func (dao *hookDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.Hook) (uint32, error) {
	if g == nil {
		return 0, errors.New("hook is nil")
	}

	if err := g.ValidateCreate(); err != nil {
		return 0, err
	}

	//generate a hook id and update to hook.
	id, err := dao.idGen.One(kit, table.Name(g.TableName()))
	if err != nil {
		return 0, err
	}
	g.ID = id

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)

	// 多个使用事务处理

	q := tx.Hook.WithContext(kit.Ctx)
	if e := q.Create(g); e != nil {
		return 0, e
	}

	if e := ad.Do(tx.Query); e != nil {
		return 0, e
	}

	return g.ID, nil
}

// ListWithRefer hooks with options.
func (dao *hookDao) ListWithRefer(kit *kit.Kit, opt *types.ListHooksWithReferOption) (
	[]*types.ListHooksWithReferDetail, int64, error) {

	h := dao.genQ.Hook
	hr := dao.genQ.HookRevision
	rh := dao.genQ.ReleasedHook
	q := dao.genQ.Hook.WithContext(kit.Ctx).Where(h.BizID.Eq(opt.BizID)).Order(h.ID.Desc())

	if opt.Name != "" {
		q = q.Where(h.Name.Like(fmt.Sprintf("%%%s%%", opt.Name)))
	}
	if opt.Tag != "" {
		q = q.Where(h.Tag.Eq(opt.Tag))
	} else {
		if opt.NotTag {
			q = q.Where(h.Tag.Eq(""))
		}
	}

	details := make([]*types.ListHooksWithReferDetail, 0)

	q = q.Select(h.ALL, rh.ID.Count().As("refer_count"), rh.ReleaseID.Min().Eq(0).As("refer_editing_release"),
		hr.ID.Max().As("published_revision_id")).
		LeftJoin(rh, h.ID.EqCol(rh.HookID)).
		LeftJoin(hr, h.ID.EqCol(hr.HookID), hr.State.Eq(table.HookRevisionStatusDeployed.String())).
		Group(h.ID)

	if opt.Page.Start == 0 && opt.Page.Limit == 0 {
		if err := q.Scan(&details); err != nil {
			return nil, 0, err
		}
		return details, int64(len(details)), nil
	}

	count, err := q.ScanByPage(&details, opt.Page.Offset(), opt.Page.LimitInt())
	if err != nil {
		return nil, 0, err
	}

	return details, count, err
}

// ListHookReferences list hook references.
func (dao *hookDao) ListHookReferences(kit *kit.Kit, opt *types.ListHookReferencesOption) (
	[]*types.ListHookReferencesDetail, int64, error) {

	rh := dao.genQ.ReleasedHook
	r := dao.genQ.Release
	a := dao.genQ.App

	details := make([]*types.ListHookReferencesDetail, 0)
	var count int64
	var err error

	count, err = rh.WithContext(kit.Ctx).
		Select(rh.ID.As("hook_revision_id"), rh.HookRevisionName.As("hook_revision_name"), rh.HookType.As("hook_type"),
			a.ID.As("app_id"), a.Name.As("app_name"), r.ID.As("release_id"), r.Name.As("release_name")).
		LeftJoin(a, rh.AppID.EqCol(a.ID)).
		LeftJoin(r, rh.ReleaseID.EqCol(r.ID)).
		Where(rh.HookID.Eq(opt.HookID), rh.BizID.Eq(opt.BizID)).
		Order(rh.ID.Desc()).
		ScanByPage(&details, opt.Page.Offset(), opt.Page.LimitInt())

	for i := range details {
		if details[i].ReleaseID == 0 {
			details[i].ReleaseName = "未命名版本"
		}
	}

	return details, count, err
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

// DeleteWithTx one hook instance.
func (dao *hookDao) DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.Hook) error {

	if g == nil {
		return errors.New("hook is nil")
	}

	if err := g.ValidateDelete(); err != nil {
		return err
	}

	m := tx.Hook
	q := tx.Hook.WithContext(kit.Ctx)

	oldOne, err := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	_, err = q.Where(m.BizID.Eq(g.Attachment.BizID)).Delete(g)
	if err != nil {
		return err
	}

	if e := ad.Do(tx.Query); e != nil {
		return e
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
