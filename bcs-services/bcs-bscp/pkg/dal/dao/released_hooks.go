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
	"errors"

	"gorm.io/gorm"

	"bscp.io/pkg/criteria/errf"
	"bscp.io/pkg/dal/gen"
	"bscp.io/pkg/dal/table"
	"bscp.io/pkg/kit"
)

// ReleasedHook supplies all the group related operations.
type ReleasedHook interface {
	// UpsertWithTx upserts the released hook with transaction.
	UpsertWithTx(kit *kit.Kit, tx *gen.QueryTx, rh *table.ReleasedHook) error
	// UpdateHookRevisionByReleaseIDWithTx updates the hook's revision info by release id with transaction.
	UpdateHookRevisionByReleaseIDWithTx(kit *kit.Kit, tx *gen.QueryTx,
		bizID, releaseID, hookID uint32, hookRevision *table.HookRevision) error
	// Get gets the released hook.
	Get(kit *kit.Kit, bizID, appID, releaseID uint32, tp table.HookType) (*table.ReleasedHook, error)
	// GetByReleaseID gets the pre hook and post hook by release id.
	GetByReleaseID(kit *kit.Kit, bizID, releaseID uint32) (*table.ReleasedHook, *table.ReleasedHook, error)
	// CreateWithTx creates a new released hook with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, releasedHook *table.ReleasedHook) (uint32, error)
	// ListAll list all released hooks in biz
	ListAll(kit *kit.Kit, bizID uint32) ([]*table.ReleasedHook, error)
	// DeleteByAppIDWithTx deletes the released hook by app id with transaction.
	DeleteByAppIDWithTx(kit *kit.Kit, tx *gen.QueryTx, appID, bizID uint32) error
	// DeleteByUniqueKeyWithTx deletes the released hook by unique key with transaction.
	DeleteByUniqueKeyWithTx(kit *kit.Kit, tx *gen.QueryTx, rh *table.ReleasedHook) error
	// DeleteByHookIDAndReleaseIDWithTx deletes the released hook by hook id and release id with transaction.
	DeleteByHookIDAndReleaseIDWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID, hookID, releaseID uint32) error
	// CountByHookIDAndReleaseID counts the released hook by hook id and release id.
	CountByHookIDAndReleaseID(kit *kit.Kit, bizID, hookID, releaseID uint32) (int64, error)
	// DeleteByHookRevisionIDAndReleaseIDWithTx deletes the released hook by revision id and release id with tx.
	DeleteByHookRevisionIDAndReleaseIDWithTx(kit *kit.Kit, tx *gen.QueryTx, bizID,
		hookID, hookRevisionID, releaseID uint32) error
	// CountByHookRevisionIDAndReleaseID counts the released hook by hook revision id and release id.
	CountByHookRevisionIDAndReleaseID(kit *kit.Kit, bizID, hookID, hookRevisionID, releaseID uint32) (int64, error)
}

var _ ReleasedHook = new(releasedHookDao)

type releasedHookDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// Get gets the released hook.
func (dao *releasedHookDao) Get(kit *kit.Kit, bizID, appID, releaseID uint32, tp table.HookType) (
	*table.ReleasedHook, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID is 0")
	}
	if appID == 0 {
		return nil, errf.New(errf.InvalidParameter, "appID is 0")
	}
	if err := tp.Validate(); err != nil {
		return nil, err
	}
	m := dao.genQ.ReleasedHook
	rh, err := m.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.ReleaseID.Eq(releaseID), m.HookType.Eq(tp.String())).Take()
	if err != nil {
		return nil, err
	}
	content, err := base64.StdEncoding.DecodeString(rh.Content)
	if err != nil {
		return nil, err
	}
	rh.Content = string(content)
	return rh, nil
}

// GetByReleaseID gets the pre hook and post hook by release id.
func (dao *releasedHookDao) GetByReleaseID(kit *kit.Kit, bizID, releaseID uint32) (
	*table.ReleasedHook, *table.ReleasedHook, error) {
	if bizID == 0 {
		return nil, nil, errf.New(errf.InvalidParameter, "bizID is 0")
	}
	m := dao.genQ.ReleasedHook
	pre, err := m.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.ReleaseID.Eq(releaseID),
		m.HookType.Eq(table.PreHook.String())).Take()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, err
	}
	if pre != nil {
		content, e := base64.StdEncoding.DecodeString(pre.Content)
		if e != nil {
			return nil, nil, e
		}
		pre.Content = string(content)
	}
	post, err := m.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.ReleaseID.Eq(releaseID),
		m.HookType.Eq(table.PostHook.String())).Take()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, err
	}
	if post != nil {
		content, e := base64.StdEncoding.DecodeString(post.Content)
		if e != nil {
			return nil, nil, e
		}
		post.Content = string(content)
	}

	return pre, post, nil
}

// UpsertWithTx upserts the released hook with transaction.
func (dao *releasedHookDao) UpsertWithTx(kit *kit.Kit, tx *gen.QueryTx, rh *table.ReleasedHook) error {
	if rh == nil {
		return errors.New("released hook is nil")
	}

	if err := rh.ValidateCreate(); err != nil {
		return err
	}

	// upsert the released hook.
	m := tx.ReleasedHook
	old, err := m.WithContext(kit.Ctx).Where(m.BizID.Eq(rh.BizID), m.AppID.Eq(rh.AppID),
		m.ReleaseID.Eq(rh.ReleaseID), m.HookType.Eq(rh.HookType.String())).Take()
	rh.Content = base64.StdEncoding.EncodeToString([]byte(rh.Content))
	// if old exists, update it.
	if err == nil {
		rh.ID = old.ID
		if _, e := m.WithContext(kit.Ctx).Updates(rh); e != nil {
			return e
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// if old not exists, create it.
		id, e := dao.idGen.One(kit, table.Name(rh.TableName()))
		if e != nil {
			return e
		}
		rh.ID = id
		return m.WithContext(kit.Ctx).Create(rh)
	}
	return err
}

// UpdateHookRevisionByReleaseIDWithTx updates the hook's revision info by release id with transaction.
func (dao *releasedHookDao) UpdateHookRevisionByReleaseIDWithTx(kit *kit.Kit, tx *gen.QueryTx,
	bizID, releaseID, hookID uint32, hookRevision *table.HookRevision) error {
	if bizID == 0 {
		return errf.New(errf.InvalidParameter, "bizID is 0")
	}
	if hookID == 0 {
		return errf.New(errf.InvalidParameter, "hookID is 0")
	}
	if hookRevision == nil {
		return errf.New(errf.InvalidParameter, "hookRevision is nil")
	}
	m := tx.ReleasedHook
	_, err := m.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.ReleaseID.Eq(releaseID), m.HookID.Eq(hookID)).
		Updates(table.ReleasedHook{
			HookRevisionID:   hookRevision.ID,
			HookRevisionName: hookRevision.Spec.Name,
			Content:          base64.StdEncoding.EncodeToString([]byte(hookRevision.Spec.Content)),
			Reviser:          hookRevision.Revision.Reviser,
		})
	return err
}

// CreateWithTx creates a new released hook with transaction.
func (dao *releasedHookDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, rh *table.ReleasedHook) (
	uint32, error) {
	if rh == nil {
		return 0, errors.New("released hook is nil")
	}

	if err := rh.ValidateCreate(); err != nil {
		return 0, err
	}

	rh.Content = base64.StdEncoding.EncodeToString([]byte(rh.Content))

	// generate an released hook id and update to released hook.
	id, err := dao.idGen.One(kit, table.Name(rh.TableName()))
	if err != nil {
		return 0, err
	}

	rh.ID = id

	if err := tx.ReleasedHook.WithContext(kit.Ctx).Create(rh); err != nil {
		return 0, err
	}

	return id, nil
}

// ListAll list all released hooks in biz
func (dao *releasedHookDao) ListAll(kit *kit.Kit, bizID uint32) ([]*table.ReleasedHook, error) {
	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID is 0")
	}

	m := dao.genQ.ReleasedHook
	rhs, err := m.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID)).Find()
	if err != nil {
		return nil, err
	}
	for i := range rhs {
		content, err := base64.StdEncoding.DecodeString(rhs[i].Content)
		if err != nil {
			return nil, err
		}
		rhs[i].Content = string(content)
	}
	return rhs, nil
}

// DeleteByAppIDWithTx deletes the released hook by app id with transaction.
func (dao *releasedHookDao) DeleteByAppIDWithTx(kit *kit.Kit, tx *gen.QueryTx, appID, bizID uint32) error {
	if bizID == 0 {
		return errf.New(errf.InvalidParameter, "bizID is 0")
	}
	if appID == 0 {
		return errf.New(errf.InvalidParameter, "appID is 0")
	}
	m := tx.ReleasedHook
	_, err := m.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID)).Delete()
	return err
}

// DeleteByUniqueKeyWithTx deletes the released hook by unique key with transaction.
func (dao *releasedHookDao) DeleteByUniqueKeyWithTx(kit *kit.Kit, tx *gen.QueryTx, rh *table.ReleasedHook) error {
	if rh == nil {
		return errors.New("released hook is nil")
	}
	if rh.AppID == 0 {
		return errf.New(errf.InvalidParameter, "appID is 0")
	}
	if rh.BizID == 0 {
		return errf.New(errf.InvalidParameter, "bizID is 0")
	}
	if err := rh.HookType.Validate(); err != nil {
		return err
	}
	m := tx.ReleasedHook
	_, err := m.WithContext(kit.Ctx).Where(m.BizID.Eq(rh.BizID), m.AppID.Eq(rh.AppID),
		m.ReleaseID.Eq(rh.ReleaseID), m.HookType.Eq(rh.HookType.String())).Delete()
	return err
}

// DeleteByHookIDAndReleaseIDWithTx deletes the released hook by hook id and release id with transaction.
func (dao *releasedHookDao) DeleteByHookIDAndReleaseIDWithTx(kit *kit.Kit, tx *gen.QueryTx,
	bizID, hookID, releaseID uint32) error {
	if bizID == 0 {
		return errf.New(errf.InvalidParameter, "bizID is 0")
	}
	if hookID == 0 {
		return errf.New(errf.InvalidParameter, "hook is 0")
	}
	m := tx.ReleasedHook
	_, err := m.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.HookID.Eq(hookID), m.ReleaseID.Eq(releaseID)).
		Delete()
	return err
}

// DeleteByHookRevisionIDAndReleaseIDWithTx deletes the released hook by hook revision id and release id with tx.
func (dao *releasedHookDao) DeleteByHookRevisionIDAndReleaseIDWithTx(kit *kit.Kit, tx *gen.QueryTx,
	bizID, hookID, hookRevisionID, releaseID uint32) error {
	if bizID == 0 {
		return errf.New(errf.InvalidParameter, "bizID is 0")
	}
	if hookID == 0 {
		return errf.New(errf.InvalidParameter, "hook is 0")
	}
	if hookRevisionID == 0 {
		return errf.New(errf.InvalidParameter, "hook revision is 0")
	}
	m := tx.ReleasedHook
	_, err := m.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.HookID.Eq(hookID), m.HookRevisionID.Eq(hookRevisionID), m.ReleaseID.Eq(releaseID)).
		Delete()
	return err
}

// CountByHookIDAndReleaseID counts the released hook by hook id and release id.
func (dao *releasedHookDao) CountByHookIDAndReleaseID(kit *kit.Kit, bizID, hookID, releaseID uint32) (int64, error) {
	if bizID == 0 {
		return 0, errf.New(errf.InvalidParameter, "bizID is 0")
	}
	if hookID == 0 {
		return 0, errf.New(errf.InvalidParameter, "hook is 0")
	}
	m := dao.genQ.ReleasedHook
	return m.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.HookID.Eq(hookID), m.ReleaseID.Eq(releaseID)).
		Count()
}

// CountByHookRevisionIDAndReleaseID counts the released hook by hook revision id and release id.
func (dao *releasedHookDao) CountByHookRevisionIDAndReleaseID(kit *kit.Kit,
	bizID, hookID, hookRevisionID, releaseID uint32) (int64, error) {
	if bizID == 0 {
		return 0, errf.New(errf.InvalidParameter, "bizID is 0")
	}
	if hookID == 0 {
		return 0, errf.New(errf.InvalidParameter, "hook is 0")
	}
	if hookRevisionID == 0 {
		return 0, errf.New(errf.InvalidParameter, "hook revision is 0")
	}
	m := dao.genQ.ReleasedHook
	return m.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.HookID.Eq(hookID), m.HookRevisionID.Eq(hookRevisionID), m.ReleaseID.Eq(releaseID)).
		Count()
}
