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
	"encoding/base64"
	"fmt"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// HookRevision supplies all the hook revision related operations.
type HookRevision interface {
	// Create one hook revision instance.
	Create(kit *kit.Kit, hook *table.HookRevision) (uint32, error)
	// CreateWithTx create hook revision instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, hr *table.HookRevision) (uint32, error)
	// Get hook revision by id
	Get(kit *kit.Kit, bizID, hookID, id uint32) (*table.HookRevision, error)
	// GetByName get HookRevision by name
	GetByName(kit *kit.Kit, bizID, hookID uint32, name string) (*table.HookRevision, error)
	// List hooks with options.
	List(kit *kit.Kit, opt *types.ListHookRevisionsOption) ([]*table.HookRevision, int64, error)
	// ListWithRefer hook revisions with refer info.
	ListWithRefer(kit *kit.Kit, opt *types.ListHookRevisionsOption) (
		[]*types.ListHookRevisionsWithReferDetail, int64, error)
	// ListHookRevisionReferences list hook references.
	ListHookRevisionReferences(kit *kit.Kit, opt *types.ListHookRevisionReferencesOption) (
		[]*types.ListHookRevisionReferencesDetail, int64, error)
	// DeleteWithTx one hook revision instance with transaction.
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, hr *table.HookRevision) error
	// GetByPubState hook revision by State
	GetByPubState(kit *kit.Kit, opt *types.GetByPubStateOption) (*table.HookRevision, error)
	// DeleteByHookIDWithTx  delete revision revision with transaction
	DeleteByHookIDWithTx(kit *kit.Kit, tx *gen.QueryTx, hookID, bizID uint32) error
	// UpdatePubStateWithTx update hookRevision State instance with transaction.
	UpdatePubStateWithTx(kit *kit.Kit, tx *gen.QueryTx, hr *table.HookRevision) error
	// Update one HookRevision's info.
	Update(kit *kit.Kit, hr *table.HookRevision) error
}

var _ HookRevision = new(hookRevisionDao)

type hookRevisionDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Create one hook instance.
func (dao *hookRevisionDao) Create(kit *kit.Kit, hr *table.HookRevision) (uint32, error) {

	if err := hr.ValidateCreate(kit); err != nil {
		return 0, err
	}

	hr.Spec.Content = base64.StdEncoding.EncodeToString([]byte(hr.Spec.Content))

	// generate a HookRevision id and update to HookRevision.
	id, err := dao.idGen.One(kit, hr.TableName())
	if err != nil {
		return 0, err
	}
	hr.ID = id

	ad := dao.auditDao.DecoratorV2(kit, hr.Attachment.BizID).PrepareCreate(hr)

	// 多个使用事务处理
	createTx := func(tx *gen.Query) error {
		q := tx.HookRevision.WithContext(kit.Ctx)
		if e := q.Create(hr); e != nil {
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

	return hr.ID, nil

}

// CreateWithTx create one hookRevision instance with transaction.
func (dao *hookRevisionDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, hr *table.HookRevision) (uint32, error) {

	if err := hr.ValidateCreate(kit); err != nil {
		return 0, err
	}

	hr.Spec.Content = base64.StdEncoding.EncodeToString([]byte(hr.Spec.Content))

	// generate a HookRevision id and update to HookRevision.
	id, err := dao.idGen.One(kit, hr.TableName())
	if err != nil {
		return 0, err
	}
	hr.ID = id

	ad := dao.auditDao.DecoratorV2(kit, hr.Attachment.BizID).PrepareCreate(hr)

	err = tx.HookRevision.WithContext(kit.Ctx).Create(hr)
	if err != nil {
		return 0, err
	}
	err = ad.Do(tx.Query)
	if err != nil {
		return 0, err
	}

	return hr.ID, nil
}

// Get hookRevision by id
func (dao *hookRevisionDao) Get(kit *kit.Kit, bizID, hookID, id uint32) (*table.HookRevision, error) {

	m := dao.genQ.HookRevision
	q := dao.genQ.HookRevision.WithContext(kit.Ctx)

	hr, err := q.Where(m.BizID.Eq(bizID), m.HookID.Eq(hookID), m.ID.Eq(id)).Take()
	if err != nil {
		return nil, fmt.Errorf("get hookRevision failed, err: %v", err)
	}

	content, err := base64.StdEncoding.DecodeString(hr.Spec.Content)
	if err != nil {
		return nil, err
	}
	hr.Spec.Content = string(content)

	return hr, nil
}

// GetByName get HookRevision by name
func (dao *hookRevisionDao) GetByName(kit *kit.Kit, bizID, hookID uint32, name string) (*table.HookRevision, error) {
	m := dao.genQ.HookRevision
	q := dao.genQ.HookRevision.WithContext(kit.Ctx)

	hr, err := q.Where(m.BizID.Eq(bizID), m.HookID.Eq(hookID), m.Name.Eq(name)).Take()
	if err != nil {
		return nil, fmt.Errorf("get hookRevision failed, err: %v", err)
	}

	content, err := base64.StdEncoding.DecodeString(hr.Spec.Content)
	if err != nil {
		return nil, err
	}
	hr.Spec.Content = string(content)

	return hr, nil
}

// List hooks with options.
func (dao *hookRevisionDao) List(kit *kit.Kit,
	opt *types.ListHookRevisionsOption) ([]*table.HookRevision, int64, error) {

	m := dao.genQ.HookRevision
	q := dao.genQ.HookRevision.WithContext(kit.Ctx).Where(
		m.BizID.Eq(opt.BizID),
		m.HookID.Eq(opt.HookID)).Order(m.ID.Desc())

	if opt.SearchKey != "" {
		searchKey := "(?i)" + opt.SearchKey
		// Where 内嵌表示括号, 例如: q.Where(q.Where(a).Or(b)) => (a or b)
		// 参考: https://gorm.io/zh_CN/gen/query.html#Group-%E6%9D%A1%E4%BB%B6
		q = q.Where(q.Where(m.Name.Regexp(searchKey)).Or(m.Memo.Regexp(searchKey)).Or(m.Reviser.Regexp(searchKey)))
	}

	if opt.State != "" {
		q = q.Where(m.State.Eq(opt.State.String()))
	}

	var result []*table.HookRevision
	var count int64
	var err error

	if opt.Page.Start == 0 && opt.Page.Limit == 0 {
		result, err = q.Find()
		if err != nil {
			return nil, 0, err
		}
		count = int64(len(result))
	} else {
		result, count, err = q.FindByPage(opt.Page.Offset(), opt.Page.LimitInt())
		if err != nil {
			return nil, 0, err
		}
	}

	for i := range result {
		content, e := base64.StdEncoding.DecodeString(result[i].Spec.Content)
		if e != nil {
			return nil, 0, e
		}
		result[i].Spec.Content = string(content)
	}

	return result, count, err
}

// ListWithRefer hook revisions with refer info.
func (dao *hookRevisionDao) ListWithRefer(kit *kit.Kit, opt *types.ListHookRevisionsOption) (
	[]*types.ListHookRevisionsWithReferDetail, int64, error) {

	m := dao.genQ.HookRevision
	rh := dao.genQ.ReleasedHook
	q := dao.genQ.HookRevision.WithContext(kit.Ctx)

	if opt.SearchKey != "" {
		searchKey := "(?i)" + opt.SearchKey
		// Where 内嵌表示括号, 例如: q.Where(q.Where(a).Or(b)) => (a or b)
		// 参考: https://gorm.io/zh_CN/gen/query.html#Group-%E6%9D%A1%E4%BB%B6
		q = q.Where(m.BizID.Eq(opt.BizID), m.HookID.Eq(opt.HookID)).
			Where(q.Where(m.Name.Regexp(searchKey)).Or(m.Memo.Regexp(searchKey)).Or(m.Reviser.Regexp(searchKey)))
	} else {
		q = q.Where(m.BizID.Eq(opt.BizID), m.HookID.Eq(opt.HookID))
	}

	if opt.State != "" {
		q = q.Where(m.State.Eq(opt.State.String()))
	}

	details := make([]*types.ListHookRevisionsWithReferDetail, 0)
	var count int64
	var err error

	q = q.Select(m.ALL, rh.ID.Count().As("refer_count"), rh.ReleaseID.Min().Eq(0).As("refer_editing_release")).
		LeftJoin(rh, m.ID.EqCol(rh.HookRevisionID)).
		Order(m.ID.Desc()).
		Group(m.ID)

	if opt.Page.Start == 0 && opt.Page.Limit == 0 {
		if err = q.Scan(&details); err != nil {
			return nil, 0, err
		}
		count = int64(len(details))
	} else {
		count, err = q.ScanByPage(&details, opt.Page.Offset(), opt.Page.LimitInt())
		if err != nil {
			return nil, 0, err
		}
	}
	for i := range details {
		content, e := base64.StdEncoding.DecodeString(details[i].HookRevision.Spec.Content)
		if e != nil {
			return nil, 0, e
		}
		details[i].HookRevision.Spec.Content = string(content)
	}

	return details, count, err
}

// ListHookRevisionReferences list hook references.
func (dao *hookRevisionDao) ListHookRevisionReferences(kit *kit.Kit, opt *types.ListHookRevisionReferencesOption) (
	[]*types.ListHookRevisionReferencesDetail, int64, error) {

	rh := dao.genQ.ReleasedHook
	r := dao.genQ.Release
	a := dao.genQ.App

	details := make([]*types.ListHookRevisionReferencesDetail, 0)
	var count int64
	var err error

	query := rh.WithContext(kit.Ctx).
		Select(rh.HookRevisionID.As("revision_id"), rh.HookRevisionName.As("revision_name"),
			rh.HookType.As("hook_type"), a.ID.As("app_id"), a.Name.As("app_name"),
			r.ID.As("release_id"), r.Name.As("release_name"), r.Deprecated).
		LeftJoin(a, rh.AppID.EqCol(a.ID)).
		LeftJoin(r, rh.ReleaseID.EqCol(r.ID)).
		Where(rh.HookID.Eq(opt.HookID), rh.HookRevisionID.Eq(opt.HookRevisionsID), rh.BizID.Eq(opt.BizID))

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
			details[i].ReleaseName = "未命名版本"
		}
	}

	return details, count, err
}

// DeleteWithTx one hook revision instance with transaction.
func (dao *hookRevisionDao) DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, hr *table.HookRevision) error {

	// 参数校验
	if err := hr.ValidateDelete(); err != nil {
		return err
	}

	// 删除操作, 获取当前记录做审计
	m := tx.HookRevision
	q := tx.HookRevision.WithContext(kit.Ctx)
	oldOne, err := q.Where(m.ID.Eq(hr.ID), m.BizID.Eq(hr.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, hr.Attachment.BizID).PrepareDelete(oldOne)

	// 多个使用事务处理
	deleteTx := func(tx *gen.Query) error {
		q = tx.HookRevision.WithContext(kit.Ctx)
		if _, e := q.Where(m.BizID.Eq(hr.Attachment.BizID), m.ID.Eq(hr.ID)).Delete(hr); e != nil {
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

// DeleteByHookIDWithTx  delete revision revision with transaction
func (dao *hookRevisionDao) DeleteByHookIDWithTx(kit *kit.Kit, tx *gen.QueryTx, hookID, bizID uint32) error {

	m := tx.HookRevision
	q := tx.HookRevision.WithContext(kit.Ctx)

	if _, e := q.Where(m.BizID.Eq(bizID), m.HookID.Eq(hookID)).Delete(); e != nil {
		return e
	}

	return nil
}

// GetByPubState hook revision by State
func (dao *hookRevisionDao) GetByPubState(kit *kit.Kit,
	opt *types.GetByPubStateOption) (*table.HookRevision, error) {

	// 参数校验
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	m := dao.genQ.HookRevision
	q := dao.genQ.HookRevision.WithContext(kit.Ctx)

	revision, err := q.Where(
		m.BizID.Eq(opt.BizID),
		m.HookID.Eq(opt.HookID),
		m.State.Eq(opt.State.String()),
	).Take()
	if err != nil {
		return nil, err
	}

	content, err := base64.StdEncoding.DecodeString(revision.Spec.Content)
	if err != nil {
		return nil, err
	}
	revision.Spec.Content = string(content)

	return revision, nil

}

// UpdatePubStateWithTx update hookRevision State instance with transaction.
func (dao *hookRevisionDao) UpdatePubStateWithTx(kit *kit.Kit, tx *gen.QueryTx, hr *table.HookRevision) error {

	if err := hr.ValidatePublish(); err != nil {
		return err
	}

	q := tx.HookRevision.WithContext(kit.Ctx)
	m := tx.HookRevision

	oldOne, err := q.Where(m.HookID.Eq(hr.Attachment.HookID), m.BizID.Eq(hr.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, hr.Attachment.BizID).PrepareUpdate(hr, oldOne)

	if _, e := q.Where(m.ID.Eq(hr.ID), m.BizID.Eq(hr.Attachment.BizID)).
		Omit(m.UpdatedAt).
		Select(m.State, m.Reviser).Updates(hr); e != nil {
		return e
	}

	if e := ad.Do(tx.Query); e != nil {
		return e
	}

	return nil
}

// Update one HookRevision's info.
func (dao *hookRevisionDao) Update(kit *kit.Kit, hr *table.HookRevision) error {

	if err := hr.ValidateUpdate(kit); err != nil {
		return err
	}

	q := dao.genQ.HookRevision.WithContext(kit.Ctx)
	m := dao.genQ.HookRevision

	oldOne, err := q.Where(m.HookID.Eq(hr.Attachment.HookID), m.BizID.Eq(hr.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, hr.Attachment.BizID).PrepareUpdate(hr, oldOne)

	hr.Spec.Content = base64.StdEncoding.EncodeToString([]byte(hr.Spec.Content))

	// 多个使用事务处理
	updateTx := func(tx *gen.Query) error {
		q = tx.HookRevision.WithContext(kit.Ctx)
		if _, e := q.Where(m.BizID.Eq(hr.Attachment.BizID), m.ID.Eq(hr.ID)).
			Select(m.Name, m.Memo, m.Content, m.Reviser).
			Updates(hr); e != nil {
			return e
		}

		if e := ad.Do(tx); e != nil {
			return e
		}
		return nil
	}
	if e := dao.genQ.Transaction(updateTx); e != nil {
		return e
	}

	return nil
}
