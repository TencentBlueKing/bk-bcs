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

	"gorm.io/datatypes"
	rawgen "gorm.io/gen"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	dtypes "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/types"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
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
	// Update one hook instance.
	Update(kit *kit.Kit, hook *table.Hook) error
	// UpdateWithTx update one hook instance with transaction.
	UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, hook *table.Hook) error
	// GetByID get hook only with id.
	GetByID(kit *kit.Kit, bizID, hookID uint32) (*table.Hook, error)
	// GetByName get hook by name
	GetByName(kit *kit.Kit, bizID uint32, name string) (*table.Hook, error)
	// FetchIDsExcluding 获取指定ID后排除的ID
	FetchIDsExcluding(kit *kit.Kit, bizID uint32, ids []uint32) ([]uint32, error)
	// CountNumberUnReferences 统计未引用的数量
	CountNumberUnReferences(kit *kit.Kit, bizID uint32, opt *types.ListHooksWithReferOption) (int64, error)
	// GetReferencedIDs 获取被引用的IDs
	GetReferencedIDs(kit *kit.Kit, bizID uint32) ([]uint32, error)
}

var _ Hook = new(hookDao)

type hookDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// UpdateWithTx update one hook instance with transaction.
func (dao *hookDao) UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.Hook) error {
	if err := g.ValidateUpdate(kit); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := tx.Hook
	q := tx.Hook.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).Updates(g); err != nil {
		return err
	}

	if err := ad.Do(tx.Query); err != nil {
		return err
	}

	return nil
}

// GetReferencedIDs 获取被引用的IDs
func (dao *hookDao) GetReferencedIDs(kit *kit.Kit, bizID uint32) ([]uint32, error) {

	h := dao.genQ.Hook
	rh := dao.genQ.ReleasedHook
	q := dao.genQ.Hook.WithContext(kit.Ctx).Where(h.BizID.Eq(bizID))

	var result []uint32
	err := q.Distinct(h.ID).
		LeftJoin(rh, h.ID.EqCol(rh.HookID)).
		Where(rh.HookID.IsNotNull()).
		Pluck(h.ID, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// CountNumberUnReferences 统计未引用的数量
func (dao *hookDao) CountNumberUnReferences(kit *kit.Kit, bizID uint32,
	opt *types.ListHooksWithReferOption) (int64, error) {

	h := dao.genQ.Hook
	rh := dao.genQ.ReleasedHook
	q := dao.genQ.Hook.WithContext(kit.Ctx).Where(h.BizID.Eq(bizID))
	if opt.Name != "" {
		q = q.Where(h.Name.Like("%" + opt.Name + "%"))
	}
	if opt.Tag != "" {
		q = q.Where(rawgen.Cond(datatypes.JSONArrayQuery("tags").Contains(opt.Tag))...)
	} else if opt.NotTag {
		// when the length of tags is 2, it must be '[]'
		// It could also be null
		q = q.Where(h.Tags.Length().Eq(2)).Or(h.Tags.Length().Eq(4))
	}

	return q.LeftJoin(rh, h.ID.EqCol(rh.HookID)).Where(rh.HookID.IsNull()).Count()
}

// FetchIDsExcluding 获取指定ID后排除的ID
func (dao *hookDao) FetchIDsExcluding(kit *kit.Kit, bizID uint32, ids []uint32) ([]uint32, error) {
	m := dao.genQ.Hook
	q := dao.genQ.Hook.WithContext(kit.Ctx)

	var result []uint32
	if err := q.Select(m.ID).
		Where(m.BizID.Eq(bizID), m.ID.NotIn(ids...)).
		Pluck(m.ID, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateWithTx create one hook instance with transaction.
func (dao *hookDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.Hook) (uint32, error) {
	if g == nil {
		return 0, errf.Errorf(errf.InvalidArgument, i18n.T(kit, "hook is nil"))
	}

	if err := g.ValidateCreate(kit); err != nil {
		return 0, err
	}

	// generate a hook id and update to hook.
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
	q := dao.genQ.Hook.WithContext(kit.Ctx).Where(h.BizID.Eq(opt.BizID)).Order(h.Name)

	if opt.Name != "" {
		q = q.Where(h.Name.Like("%" + opt.Name + "%"))
	}
	if opt.Tag != "" {
		q = q.Where(rawgen.Cond(datatypes.JSONArrayQuery("tags").Contains(opt.Tag))...)
	} else if opt.NotTag {
		// when the length of tags is 2, it must be '[]'
		// It could also be null
		q = q.Where(h.Tags.Length().Eq(2)).Or(h.Tags.Length().Eq(4))
	}

	if opt.SearchKey != "" {
		searchKey := "(?i)" + opt.SearchKey
		// Where 内嵌表示括号, 例如: q.Where(q.Where(a).Or(b)) => (a or b)
		// 参考: https://gorm.io/zh_CN/gen/query.html#Group-%E6%9D%A1%E4%BB%B6
		q = q.Where(q.Where(h.Name.Regexp(searchKey)).Or(h.Memo.Regexp(searchKey)).Or(
			h.Creator.Regexp(searchKey)).Or(h.Reviser.Regexp(searchKey)))
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

	query := rh.WithContext(kit.Ctx).
		Select(rh.ID.As("hook_revision_id"), rh.HookRevisionName.As("hook_revision_name"), rh.HookType.As("hook_type"),
			a.ID.As("app_id"), a.Name.As("app_name"), r.ID.As("release_id"), r.Name.As("release_name"), r.Deprecated).
		LeftJoin(a, rh.AppID.EqCol(a.ID)).
		LeftJoin(r, rh.ReleaseID.EqCol(r.ID)).
		Where(rh.HookID.Eq(opt.HookID), rh.BizID.Eq(opt.BizID))
	if opt.SearchKey != "" {
		searchKey := "(?i)" + opt.SearchKey
		// Where 内嵌表示括号, 例如: q.Where(q.Where(a).Or(b)) => (a or b)
		// 参考: https://gorm.io/zh_CN/gen/query.html#Group-%E6%9D%A1%E4%BB%B6
		query = query.Where(query.Where(
			a.Name.Regexp(searchKey)).Or(r.Name.Regexp(searchKey)).Or(rh.HookRevisionName.Regexp(searchKey)))
	}

	count, err = query.Order(rh.ID.Desc()).ScanByPage(&details, opt.Page.Offset(), opt.Page.LimitInt())

	for i := range details {
		if details[i].ReleaseID == 0 {
			details[i].ReleaseName = i18n.T(kit, "Unnamed Version")
		}
	}

	return details, count, err
}

// CountHookTag count hook tag
func (dao *hookDao) CountHookTag(kit *kit.Kit, bizID uint32) ([]*types.HookTagCount, error) {
	m := dao.genQ.Hook
	q := dao.genQ.Hook.WithContext(kit.Ctx)

	var allTags []dtypes.StringSlice
	if err := q.Select(m.Tags).Where(m.BizID.Eq(bizID)).
		// when the length of tags greater than 2, it must not be empty which means not be '[]'
		Where(m.Tags.Length().Gt(2)).
		Scan(&allTags); err != nil {
		return nil, err
	}
	if len(allTags) == 0 {
		return []*types.HookTagCount{}, nil
	}
	tagCnt := make(map[string]uint32)
	for _, tags := range allTags {
		for _, t := range tags {
			tagCnt[t]++
		}
	}

	counts := make([]*types.HookTagCount, 0)
	for t, cnt := range tagCnt {
		counts = append(counts, &types.HookTagCount{Tag: t, Counts: cnt})
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

// Update one hook instance.
func (dao *hookDao) Update(kit *kit.Kit, g *table.Hook) error {
	if err := g.ValidateUpdate(kit); err != nil {
		return err
	}

	// 更新操作, 获取当前记录做审计
	m := dao.genQ.Hook
	q := dao.genQ.Hook.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.Hook.WithContext(kit.Ctx)
		if _, err := q.Where(m.BizID.Eq(g.Attachment.BizID), m.ID.Eq(g.ID)).
			Select(m.Tags, m.Memo).Omit(m.Reviser, m.UpdatedAt).Updates(g); err != nil {
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
