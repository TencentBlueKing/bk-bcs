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
	"fmt"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// Group supplies all the group related operations.
type Group interface {
	// ListAppValidGroups list app valid groups.
	ListAppValidGroups(kit *kit.Kit, bizID, appID uint32) ([]*table.Group, error)
	// CreateWithTx Create one group instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, group *table.Group) (uint32, error)
	// UpdateWithTx Update one group instance with transaction.
	UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, group *table.Group) error
	// Get group by id.
	Get(kit *kit.Kit, id, bizID uint32) (*table.Group, error)
	// GetByName get group by name.
	GetByName(kit *kit.Kit, bizID uint32, name string) (*table.Group, error)
	// ListAll list all the groups in biz.
	ListAll(kit *kit.Kit, bizID uint32) ([]*table.Group, error)
	// DeleteWithTx delete one group instance with transaction.
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, group *table.Group) error
	// ListAppGroups list all the groups of the app.
	ListAppGroups(kit *kit.Kit, bizID, appID uint32) ([]*table.Group, error)
	// ListGroupReleasedApps list all the released apps of the group.
	ListGroupReleasedApps(kit *kit.Kit, opts *types.ListGroupReleasedAppsOption) (
		*types.ListGroupReleasedAppsDetails, error)
}

var _ Group = new(groupDao)

type groupDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
	lock     LockDao
}

// ListAppValidGroups list app valid groups.
func (dao *groupDao) ListAppValidGroups(kit *kit.Kit, bizID, appID uint32) (
	[]*table.Group, error) {

	if bizID == 0 || appID == 0 {
		return nil, fmt.Errorf("bizID or appID can not be 0")
	}

	g := dao.genQ.Group
	gq := dao.genQ.Group.WithContext(kit.Ctx)
	gab := dao.genQ.GroupAppBind
	subQuery := gab.WithContext(kit.Ctx).Select(gab.GroupID).Where(gab.BizID.Eq(bizID), gab.AppID.Eq(appID))
	return gq.Where(g.BizID.Eq(bizID)).Where(gq.Where(g.Public.Is(true)).Or(gq.Columns(g.ID).In(subQuery))).Find()
}

// CreateWithTx Create one group instance with transaction.
func (dao *groupDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.Group) (uint32, error) {

	if g == nil {
		return 0, errf.New(errf.InvalidParameter, "group is nil")
	}

	if err := g.ValidateCreate(kit); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	// generate a group id and update to group.
	id, err := dao.idGen.One(kit, table.GroupTable)
	if err != nil {
		return 0, err
	}

	g.ID = id
	if err = tx.Query.Group.WithContext(kit.Ctx).Create(g); err != nil {
		return 0, err
	}

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareCreate(g)
	if err = ad.Do(tx.Query); err != nil {
		return 0, fmt.Errorf("audit create group failed, err: %v", err)
	}

	return id, nil
}

// UpdateWithTx Update one group instance with transaction.
func (dao *groupDao) UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.Group) error {

	if g == nil {
		return errf.New(errf.InvalidParameter, "group is nil")
	}

	if err := g.ValidateUpdate(kit); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	m := tx.Group

	oldOne, err := m.WithContext(kit.Ctx).Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}
	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareUpdate(g, oldOne)

	_, err = m.WithContext(kit.Ctx).
		Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).
		Select(m.Name, m.Public, m.Selector, m.UID, m.Reviser).
		Updates(g)
	if err != nil {
		return err
	}

	if err = ad.Do(tx.Query); err != nil {
		return fmt.Errorf("audit update group failed, err: %v", err)
	}
	return nil
}

// Get group by id.
func (dao *groupDao) Get(kit *kit.Kit, id, bizID uint32) (*table.Group, error) {

	if bizID == 0 || id == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID or id is 0")
	}

	if id == 0 {
		return nil, errf.New(errf.InvalidParameter, "group id can not be 0")
	}
	m := dao.genQ.Group
	return m.WithContext(kit.Ctx).Where(m.ID.Eq(id), m.BizID.Eq(bizID)).Take()
}

// GetByName get group by name.
func (dao *groupDao) GetByName(kit *kit.Kit, bizID uint32, name string) (*table.Group, error) {

	if bizID == 0 || name == "" {
		return nil, errf.New(errf.InvalidParameter, "biz id or name is empty")
	}

	m := dao.genQ.Group
	return m.WithContext(kit.Ctx).Where(m.Name.Eq(name), m.BizID.Eq(bizID)).Take()
}

// ListAll list all the groups in biz.
func (dao *groupDao) ListAll(kit *kit.Kit, bizID uint32) ([]*table.Group, error) {

	if bizID == 0 {
		return nil, errf.New(errf.InvalidParameter, "biz id is 0")
	}
	m := dao.genQ.Group
	return m.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID)).Find()

}

// DeleteWithTx delete group with transaction.
func (dao *groupDao) DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, g *table.Group) error {

	if g == nil {
		return errf.New(errf.InvalidParameter, "group is nil")
	}

	if err := g.ValidateDelete(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	m := tx.Group
	oldOne, err := m.WithContext(kit.Ctx).Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Take()
	if err != nil {
		return err
	}

	ad := dao.auditDao.DecoratorV2(kit, g.Attachment.BizID).PrepareDelete(oldOne)

	if _, err = m.WithContext(kit.Ctx).Where(m.ID.Eq(g.ID), m.BizID.Eq(g.Attachment.BizID)).Delete(); err != nil {
		return err
	}
	if err = ad.Do(tx.Query); err != nil {
		return err
	}
	return nil
}

// ListAppGroups list groups by app id.
func (dao *groupDao) ListAppGroups(kit *kit.Kit, bizID, appID uint32) ([]*table.Group, error) {

	if bizID == 0 || appID == 0 {
		return nil, errf.New(errf.InvalidParameter, "bizID or appID is 0")
	}
	gabM := dao.genQ.GroupAppBind
	gabQ := dao.genQ.GroupAppBind.WithContext(kit.Ctx)

	groupM := dao.genQ.Group
	groupQ := dao.genQ.Group.WithContext(kit.Ctx)

	subQuery := gabQ.Select(gabM.GroupID).Where(gabM.BizID.Eq(bizID), gabM.AppID.Eq(appID))
	return groupQ.
		Where(groupM.BizID.Eq(bizID)).Where(
		groupQ.Where(groupQ.Columns(groupM.ID).In(subQuery)).Or(groupM.Public.Is(true))).
		Find()
}

// ListGroupReleasedApps list group released apps and their latest release info.
func (dao *groupDao) ListGroupReleasedApps(kit *kit.Kit, opts *types.ListGroupReleasedAppsOption) (
	*types.ListGroupReleasedAppsDetails, error) {
	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list group released apps options null")
	}
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	a := dao.genQ.App
	aq := a.WithContext(kit.Ctx)
	r := dao.genQ.Release
	g := dao.genQ.ReleasedGroup

	list := make([]*types.ListGroupReleasedAppsData, 0)

	var count int64
	var err error

	if opts.SearchKey == "" {
		count, err = a.WithContext(kit.Ctx).
			Select(a.ID.As("app_id"), a.Name.As("app_name"), r.ID.As("release_id"), r.Name.As("release_name"), g.Edited).
			Join(r, a.ID.EqCol(r.AppID)).Join(g, r.ID.EqCol(g.ReleaseID), a.ID.EqCol(g.AppID)).
			Where(g.GroupID.Eq(opts.GroupID), a.BizID.Eq(opts.BizID), r.BizID.Eq(opts.BizID), g.BizID.Eq(opts.BizID)).
			ScanByPage(&list, int(opts.Start), int(opts.Limit))
	} else {
		count, err = a.WithContext(kit.Ctx).
			Select(a.ID.As("app_id"), a.Name.As("app_name"), r.ID.As("release_id"), r.Name.As("release_name"), g.Edited).
			Join(r, a.ID.EqCol(r.AppID)).Join(g, r.ID.EqCol(g.ReleaseID), a.ID.EqCol(g.AppID)).
			Where(g.GroupID.Eq(opts.GroupID), a.BizID.Eq(opts.BizID), r.BizID.Eq(opts.BizID), g.BizID.Eq(opts.BizID)).
			Where(aq.Where(a.Name.Regexp("(?i)"+opts.SearchKey)).Or(r.Name.Regexp("(?i)"+opts.SearchKey))).
			ScanByPage(&list, int(opts.Start), int(opts.Limit))
	}

	if err != nil {
		return nil, err
	}

	return &types.ListGroupReleasedAppsDetails{
		Count:   uint32(count),
		Details: list,
	}, nil
}
