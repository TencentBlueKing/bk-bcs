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
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// UserGroupPrivilege supplies all the user privileges related operations.
type UserGroupPrivilege interface {
	// CreateWithTx create one user privileges instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, data *table.UserGroupPrivilege) (uint32, error)
	// UpdateWithTx Update one user privileges instance with transaction.
	UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, id uint32, data *table.UserGroupPrivilege) error
	// GetUserGroupPrivilegeByGid 通过用户组名获取用户组数据
	GetUserGroupPrivilegeByUserGroup(kit *kit.Kit, tx *gen.QueryTx, bizID, appID,
		templateSpaceID uint32, name string) (*table.UserGroupPrivilege, error)
	// GetUserGroupPrivilegeByGid 通过gid获取用户组数据
	GetUserGroupPrivilegeByGid(kit *kit.Kit, tx *gen.QueryTx, bizID, appID, templateSpaceID, gid uint32) (
		*table.UserGroupPrivilege, error)
	// List 获取用户组数据列表
	List(kit *kit.Kit, bizID, appID, templateSpaceID uint32, name string,
		opt *types.BasePage) ([]*table.UserGroupPrivilege, int64, error)
	// GetUserGroupPrivilege 获取用户组权限数据
	GetUserGroupPrivilege(kit *kit.Kit, bizID, appID, templateSpaceID, id uint32) (*table.UserGroupPrivilege, error)
	// DeleteWithTx delete one user group privileges instance with transaction.
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, data *table.UserGroupPrivilege) error
	// ListUserPrivileges xxx
	ListUserGroupPrivileges(kit *kit.Kit, bizID, appID, templateSpaceID uint32) ([]types.UserGroupPrivilege, error)
	ListGroupPrivsBySpaceIDs(kit *kit.Kit, bizID uint32, templateSpaceID []uint32) ([]types.UserGroupPrivilege, error)
}

var _ UserGroupPrivilege = new(userGroupPrivilegeDao)

type userGroupPrivilegeDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// ListGroupPrivsBySpaceIDs implements UserGroupPrivilege.
func (dao *userGroupPrivilegeDao) ListGroupPrivsBySpaceIDs(kit *kit.Kit, bizID uint32, templateSpaceID []uint32) (
	[]types.UserGroupPrivilege, error) {
	m := dao.genQ.UserGroupPrivilege

	var userGroupPrivileges []types.UserGroupPrivilege

	err := dao.genQ.UserGroupPrivilege.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.TemplateSpaceID.In(templateSpaceID...)).
		Scan(&userGroupPrivileges)
	if err != nil {
		return nil, err
	}

	return userGroupPrivileges, nil
}

// ListUserGroupPrivileges implements UserGroupPrivilege.
func (dao *userGroupPrivilegeDao) ListUserGroupPrivileges(kit *kit.Kit, bizID uint32, appID uint32,
	templateSpaceID uint32) ([]types.UserGroupPrivilege, error) {
	m := dao.genQ.UserGroupPrivilege

	var userGroupPrivileges []types.UserGroupPrivilege

	err := dao.genQ.UserGroupPrivilege.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID),
		m.AppID.Eq(appID), m.TemplateSpaceID.Eq(templateSpaceID)).Scan(&userGroupPrivileges)
	if err != nil {
		return nil, err
	}

	return userGroupPrivileges, nil
}

// DeleteWithTx implements UserGroupPrivilege.
func (dao *userGroupPrivilegeDao) DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx,
	data *table.UserGroupPrivilege) error {
	m := tx.UserGroupPrivilege

	_, err := tx.UserGroupPrivilege.WithContext(kit.Ctx).
		Where(m.ID.Eq(data.ID), m.BizID.Eq(data.Attachment.BizID), m.AppID.Eq(data.Attachment.AppID),
			m.TemplateSpaceID.Eq(data.Attachment.TemplateSpaceID)).
		Delete(data)

	return err
}

// GetUserGroupPrivilege implements UserGroupPrivilege.
func (dao *userGroupPrivilegeDao) GetUserGroupPrivilege(kit *kit.Kit, bizID uint32, appID uint32,
	templateSpaceID uint32, id uint32) (*table.UserGroupPrivilege, error) {
	m := dao.genQ.UserGroupPrivilege

	return dao.genQ.UserGroupPrivilege.WithContext(kit.Ctx).Where(m.ID.Eq(id), m.BizID.Eq(bizID),
		m.AppID.Eq(appID), m.TemplateSpaceID.Eq(templateSpaceID)).Take()
}

// GetUserGroupPrivilegeByGid implements UserGroupPrivilege.
func (dao *userGroupPrivilegeDao) GetUserGroupPrivilegeByGid(kit *kit.Kit, tx *gen.QueryTx, bizID uint32,
	appID uint32, templateSpaceID uint32, gid uint32) (*table.UserGroupPrivilege, error) {
	m := dao.genQ.UserGroupPrivilege

	return tx.UserGroupPrivilege.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID),
		m.AppID.Eq(appID), m.TemplateSpaceID.Eq(templateSpaceID), m.Gid.Eq(gid)).Take()
}

// GetUserGroupPrivilegeByUserGroup implements UserGroupPrivilege.
func (dao *userGroupPrivilegeDao) GetUserGroupPrivilegeByUserGroup(kit *kit.Kit, tx *gen.QueryTx, bizID uint32,
	appID uint32, templateSpaceID uint32, name string) (*table.UserGroupPrivilege, error) {
	m := dao.genQ.UserGroupPrivilege

	return tx.UserGroupPrivilege.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID),
		m.AppID.Eq(appID), m.TemplateSpaceID.Eq(templateSpaceID), m.UserGroup.Eq(name)).Take()
}

// ListUserGroupPrivileges implements UserPrivilege.
func (dao *userGroupPrivilegeDao) List(kit *kit.Kit, bizID uint32, appID, templateSpaceID uint32,
	name string, opt *types.BasePage) ([]*table.UserGroupPrivilege, int64, error) {

	m := dao.genQ.UserGroupPrivilege
	q := dao.genQ.UserGroupPrivilege.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.TemplateSpaceID.Eq(templateSpaceID)).Or(m.AppID.Eq(0))

	if name != "" {
		q = q.Where(m.UserGroup.Like("%" + name + "%"))
	}

	if opt.All {
		result, err := q.Find()
		if err != nil {
			return nil, 0, err
		}
		return result, int64(len(result)), err
	}

	return q.FindByPage(opt.Offset(), opt.LimitInt())
}

// CreateWithTx implements UserPrivilege.
func (dao *userGroupPrivilegeDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx,
	data *table.UserGroupPrivilege) (uint32, error) {

	if data == nil {
		return 0, errf.ErrInvalidArgF(kit)
	}

	// generate an user privilege id and update to user privilege.
	id, err := dao.idGen.One(kit, table.UserPrivilegeTable)
	if err != nil {
		return 0, err
	}

	data.ID = id

	if err := tx.UserGroupPrivilege.WithContext(kit.Ctx).Create(data); err != nil {
		return 0, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "create user group permissions failed, err: %v", err))
	}

	return data.ID, nil
}

// UpdateWithTx implements UserPrivilege.
func (dao *userGroupPrivilegeDao) UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, id uint32,
	data *table.UserGroupPrivilege) error {

	m := dao.genQ.UserGroupPrivilege

	_, err := tx.UserGroupPrivilege.WithContext(kit.Ctx).Where(m.ID.Eq(id)).Updates(data)
	if err != nil {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "update user group permissions failed, err: %v", err))
	}

	return nil
}
