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
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// UserPrivilege supplies all the user privileges related operations.
type UserPrivilege interface {
	// CreateWithTx create one user privileges instance with transaction.
	CreateWithTx(kit *kit.Kit, tx *gen.QueryTx, userPrivilege *table.UserPrivilege) (uint32, error)
	// UpdateWithTx Update one user privileges instance with transaction.
	UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, id uint32, userPrivilege *table.UserPrivilege) error
	// List user privileges with options.
	List(kit *kit.Kit, bizID, appID, templateSpaceID uint32, name string,
		opt *types.BasePage) ([]*table.UserPrivilege, int64, error)
	GetUserPrivilegeByUser(kit *kit.Kit, tx *gen.QueryTx, bizID, appID, templateSpaceID uint32, name string) (
		*table.UserPrivilege, error)
	GetUserPrivilegeByUid(kit *kit.Kit, tx *gen.QueryTx, bizID, appID, templateSpaceID, uid uint32) (
		*table.UserPrivilege, error)
	GetUserPrivilege(kit *kit.Kit, bizID, appID, templateSpaceID, id uint32) (*table.UserPrivilege, error)
	// DeleteWithTx delete one user privileges instance with transaction.
	DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, data *table.UserPrivilege) error
	// ListUserPrivileges xxx
	ListUserPrivileges(kit *kit.Kit, bizID, appID, templateSpaceID uint32) ([]types.UserPrivilege, error)
	// ListUserPrivsBySpaceIDs xxx
	ListUserPrivsBySpaceIDs(kit *kit.Kit, bizID uint32, templateSpaceIDs []uint32) ([]types.UserPrivilege, error)
}

var _ UserPrivilege = new(userPrivilegeDao)

type userPrivilegeDao struct {
	genQ     *gen.Query
	idGen    IDGenInterface
	auditDao AuditDao
}

// ListUserPrivsBySpaceIDs implements UserPrivilege.
func (dao *userPrivilegeDao) ListUserPrivsBySpaceIDs(kit *kit.Kit, bizID uint32,
	templateSpaceIDs []uint32) ([]types.UserPrivilege, error) {
	m := dao.genQ.UserPrivilege
	var userPrivileges []types.UserPrivilege

	err := dao.genQ.UserPrivilege.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID),
		m.TemplateSpaceID.In(templateSpaceIDs...)).Scan(&userPrivileges)
	if err != nil {
		return nil, err
	}

	return userPrivileges, nil
}

// ListUserPrivileges implements UserPrivilege.
func (dao *userPrivilegeDao) ListUserPrivileges(kit *kit.Kit, bizID uint32, appID uint32,
	templateSpaceID uint32) ([]types.UserPrivilege, error) {
	m := dao.genQ.UserPrivilege

	var userPrivileges []types.UserPrivilege

	err := dao.genQ.UserPrivilege.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID),
		m.AppID.Eq(appID), m.TemplateSpaceID.Eq(templateSpaceID)).Scan(&userPrivileges)
	if err != nil {
		return nil, err
	}

	return userPrivileges, nil
}

// DeleteWithTx implements UserPrivilege.
func (dao *userPrivilegeDao) DeleteWithTx(kit *kit.Kit, tx *gen.QueryTx, data *table.UserPrivilege) error {

	m := tx.UserPrivilege

	_, err := tx.UserPrivilege.WithContext(kit.Ctx).
		Where(m.ID.Eq(data.ID), m.BizID.Eq(data.Attachment.BizID), m.AppID.Eq(data.Attachment.AppID),
			m.TemplateSpaceID.Eq(data.Attachment.TemplateSpaceID)).
		Delete(data)

	return err
}

// GetUserPrivilege implements UserPrivilege.
func (dao *userPrivilegeDao) GetUserPrivilege(kit *kit.Kit, bizID uint32, appID uint32,
	templateSpaceID, id uint32) (*table.UserPrivilege, error) {
	m := dao.genQ.UserPrivilege

	return dao.genQ.UserPrivilege.WithContext(kit.Ctx).Where(m.ID.Eq(id), m.BizID.Eq(bizID),
		m.AppID.Eq(appID), m.TemplateSpaceID.Eq(templateSpaceID)).Take()
}

// GetUserPrivilegeByUid implements UserPrivilege.
func (dao *userPrivilegeDao) GetUserPrivilegeByUid(kit *kit.Kit, tx *gen.QueryTx, bizID uint32, appID uint32,
	templateSpaceID uint32, uid uint32) (*table.UserPrivilege, error) {
	m := dao.genQ.UserPrivilege

	return tx.UserPrivilege.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID), m.AppID.Eq(appID),
		m.TemplateSpaceID.Eq(templateSpaceID), m.Uid.Eq(uid)).Take()
}

// GetUserPrivilegeByUser implements UserPrivilege.
func (dao *userPrivilegeDao) GetUserPrivilegeByUser(kit *kit.Kit, tx *gen.QueryTx, bizID uint32, appID uint32,
	templateSpaceID uint32, name string) (*table.UserPrivilege, error) {

	m := dao.genQ.UserPrivilege

	return tx.UserPrivilege.WithContext(kit.Ctx).Where(m.BizID.Eq(bizID),
		m.AppID.Eq(appID), m.TemplateSpaceID.Eq(templateSpaceID), m.User.Eq(name)).Take()
}

// ListUserGroupPrivileges implements UserPrivilege.
func (dao *userPrivilegeDao) List(kit *kit.Kit, bizID uint32, appID, templateSpaceID uint32,
	name string, opt *types.BasePage) ([]*table.UserPrivilege, int64, error) {

	m := dao.genQ.UserPrivilege
	q := dao.genQ.UserPrivilege.WithContext(kit.Ctx).
		Where(m.BizID.Eq(bizID), m.AppID.Eq(appID), m.TemplateSpaceID.Eq(templateSpaceID)).
		Or(m.AppID.Eq(0))

	if name != "" {
		q = q.Where(m.User.Like("%" + name + "%"))
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
func (dao *userPrivilegeDao) CreateWithTx(kit *kit.Kit, tx *gen.QueryTx,
	userPrivilege *table.UserPrivilege) (uint32, error) {

	if userPrivilege == nil {
		return 0, errf.ErrInvalidArgF(kit)
	}

	// generate an user privilege id and update to user privilege.
	id, err := dao.idGen.One(kit, table.UserPrivilegeTable)
	if err != nil {
		return 0, err
	}

	userPrivilege.ID = id

	if err := tx.UserPrivilege.WithContext(kit.Ctx).Create(userPrivilege); err != nil {
		return 0, err
	}

	return userPrivilege.ID, nil
}

// UpdateWithTx implements UserPrivilege.
func (dao *userPrivilegeDao) UpdateWithTx(kit *kit.Kit, tx *gen.QueryTx, id uint32,
	userPrivilege *table.UserPrivilege) error {

	m := dao.genQ.UserPrivilege

	_, err := tx.UserPrivilege.WithContext(kit.Ctx).Where(m.ID.Eq(id)).Updates(userPrivilege)
	if err != nil {
		return err
	}

	return nil
}
