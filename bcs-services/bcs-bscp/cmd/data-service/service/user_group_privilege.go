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

package service

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/criteria/errf"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/gen"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/dal/table"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/i18n"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/kit"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
	pbbase "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
	pbds "github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/data-service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/types"
)

// ListUserGroupPrivileges 获取用户组权限数据
func (s *Service) ListUserGroupPrivileges(ctx context.Context, req *pbds.ListUserPrivilegesReq) (
	*pbds.ListUserPrivilegesResp, error) {
	kit := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	data, count, err := s.dao.UserGroupPrivilege().List(kit, req.BizId, req.AppId, req.TemplateSpaceId, req.Name, opt)
	if err != nil {
		return nil, err
	}

	item := make([]*pbds.ListUserPrivilegesResp_Detail, 0, len(data))

	for _, v := range data {
		item = append(item, &pbds.ListUserPrivilegesResp_Detail{
			Id:            v.ID,
			Name:          v.Spec.UserGroup,
			PrivilegeType: string(v.Spec.PrivilegeType),
			ReadOnly:      v.Spec.ReadOnly,
			Pid:           v.Attachment.Gid,
		})
	}

	return &pbds.ListUserPrivilegesResp{Count: uint32(count), Details: item}, nil
}

// DeleteUserGroupPrivilege 删除用户组权限数据
func (s *Service) DeleteUserGroupPrivilege(ctx context.Context, req *pbds.DeleteUserPrivilegesReq) (
	*pbbase.EmptyResp, error) {
	kit := kit.FromGrpcContext(ctx)

	// 获取删除的数据
	item, err := s.dao.UserGroupPrivilege().GetUserGroupPrivilege(kit, req.BizId, req.AppId, req.TemplateSpaceId, req.Id)
	if err != nil {
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "get user group privilege failed, err: %v", err))
	}

	if item.Spec.ReadOnly {
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "this user group has read-only privilege"))
	}

	tx := s.dao.GenQuery().Begin()

	if err = s.dao.UserGroupPrivilege().DeleteWithTx(kit, tx, item); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kit.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "delete user group privilege failed, err: %v", err))
	}

	if req.TemplateSpaceId != 0 && req.AppId == 0 {
		// 删除模板版本中对应的用户权限
		err = s.dao.TemplateRevision().UpdateGroupPrivilegesWithTx(kit, tx, req.BizId,
			req.TemplateSpaceId, item.Spec.UserGroup, "")
		if err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kit.Rid)
			}
			return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "delete user privilege failed, err: %v", err))
		}
	}

	if req.TemplateSpaceId == 0 && req.AppId != 0 {
		// 删除未命名版本中对应的用户权限
		err = s.dao.ConfigItem().UpdateUserGroupPrivilegesWithTx(kit, tx, req.BizId,
			req.AppId, item.Spec.UserGroup, "")
		if err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kit.Rid)
			}
			return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "delete user group privilege failed, err: %v", err))
		}
	}

	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kit.Rid)
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "delete user group privilege failed, err: %v", err))
	}

	return &pbbase.EmptyResp{}, nil
}

// batcheUpsertUserGroupPrivileges 批量创建或更新用户组权限
func (s *Service) batcheUpsertUserGroupPrivileges(kit *kit.Kit, tx *gen.QueryTx, bizID, appID, templateSpaceID uint32,
	userGroupPrivileges []*table.UserGroupPrivilege, isTemplate bool) error {

	// 创建两个 map 来追踪 gid -> user group 和 user group -> gid 的唯一性
	// 验证提交的权限数据存在唯一性
	userGroups, gids := make(map[uint32]string), make(map[string]uint32)
	for _, v := range userGroupPrivileges {
		if user, exists := userGroups[v.Attachment.Gid]; exists && user != v.Spec.UserGroup {
			return errors.New(i18n.T(kit, "different user group have the same GID, %s", v.Spec.UserGroup))
		}

		if uid, exists := gids[v.Spec.UserGroup]; exists && uid != v.Attachment.Gid {
			return errors.New(i18n.T(kit, "different GID have the same user group, %s", v.Spec.UserGroup))
		}

		userGroups[v.Attachment.Gid] = v.Spec.UserGroup
		gids[v.Spec.UserGroup] = v.Attachment.Gid
		if err := s.handleUserGroupPrivilege(kit, tx, bizID, appID, templateSpaceID, v, isTemplate); err != nil {
			return err
		}
	}

	return nil
}

// 获取并处理 UserPrivilege 数据
// nolint:goconst
func (s *Service) handleUserGroupPrivilege(kit *kit.Kit, tx *gen.QueryTx, bizID, appID, templateSpaceID uint32,
	data *table.UserGroupPrivilege, isTemplate bool) error {

	// 如果是 root 且为 0 不处理
	if data.Spec.UserGroup == "root" && data.Attachment.Gid == 0 {
		return nil
	}

	// 验证 root 和 uid
	if err := validateUserGroupAndGID(kit, data); err != nil {
		return err
	}

	existingGID, err := s.dao.UserGroupPrivilege().GetUserGroupPrivilegeByGid(kit, tx, bizID, appID,
		templateSpaceID, data.Attachment.Gid)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "create or update user group privilege failed, err: %v", err))
	}

	existingUserGroup, err := s.dao.UserGroupPrivilege().GetUserGroupPrivilegeByUserGroup(kit, tx, bizID, appID,
		templateSpaceID, data.Spec.UserGroup)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "create or update user group privilege failed, err: %v", err))
	}

	// gid 和 user group 都不存在新增数据
	if existingGID == nil && existingUserGroup == nil {
		data.Spec.PrivilegeType = table.PrivilegeTypeCustom
		data.Revision = &table.Revision{Creator: kit.User}
		if _, err := s.dao.UserGroupPrivilege().CreateWithTx(kit, tx, data); err != nil {
			return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "create or update user group privilege failed, err: %v", err))
		}
		return nil
	}

	// gid 和 user group 都存在，且两者数据都相等的情况是不需要处理的
	if existingGID != nil && existingUserGroup != nil && existingGID.Attachment.Gid == existingUserGroup.Attachment.Gid &&
		existingGID.Spec.UserGroup == existingUserGroup.Spec.UserGroup {
		return nil
	}

	// gid 和 user group 都存在，但不相等
	if existingGID != nil && existingUserGroup != nil && existingGID.Attachment.Gid != existingUserGroup.Attachment.Gid &&
		existingGID.Spec.UserGroup != existingUserGroup.Spec.UserGroup {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "the user group or GID already exists, userGroup: %s, gid: %d",
			data.Spec.UserGroup, data.Attachment.Gid))
	}

	return s.upsertUserGroupPrivilege(kit, tx, bizID, appID, templateSpaceID, data,
		existingGID, existingUserGroup, isTemplate)
}

// upsertUserGroupPrivilege 创建或更新用户组的数据
// 判断模板和非模板文件，更新对应的表数据
func (s *Service) upsertUserGroupPrivilege(kit *kit.Kit, tx *gen.QueryTx, bizID, appID, templateSpaceID uint32,
	data *table.UserGroupPrivilege, existingGID, existingUserGroup *table.UserGroupPrivilege, isTemplate bool) error {

	data.Revision = &table.Revision{Reviser: kit.User}

	if existingGID == nil {
		return s.dao.UserGroupPrivilege().UpdateWithTx(kit, tx, existingUserGroup.ID, data)
	}

	if existingUserGroup == nil {
		if err := s.dao.UserGroupPrivilege().UpdateWithTx(kit, tx, existingGID.ID, data); err != nil {
			return err
		}
		if isTemplate {
			return s.dao.TemplateRevision().UpdateGroupPrivilegesWithTx(kit, tx, bizID, templateSpaceID,
				existingGID.Spec.UserGroup, data.Spec.UserGroup)
		}
		return s.dao.ConfigItem().UpdateUserGroupPrivilegesWithTx(kit, tx, bizID, appID, existingGID.Spec.UserGroup,
			data.Spec.UserGroup)
	}

	return nil
}

// 校验 user group 和 GID 逻辑
func validateUserGroupAndGID(kit *kit.Kit, userGroupPrivilege *table.UserGroupPrivilege) error {

	// 检查 user group 是否为 root 且 gid 是否为 0
	if userGroupPrivilege.Spec.UserGroup == "root" && userGroupPrivilege.Attachment.Gid != 0 {
		return errors.New(i18n.T(kit, "user group is root but GID is not 0"))
	}

	// 检查 gid 是否为 0 且 user group 是否为 root
	if userGroupPrivilege.Spec.UserGroup != "root" && userGroupPrivilege.Attachment.Gid == 0 {
		return errors.New(i18n.T(kit, "GID is 0 but user group is not root"))
	}

	return nil
}
