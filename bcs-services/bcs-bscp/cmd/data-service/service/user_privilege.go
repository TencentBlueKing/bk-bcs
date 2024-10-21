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
	"fmt"
	"path"

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

// ListUserPrivileges 获取用户权限
func (s *Service) ListUserPrivileges(ctx context.Context, req *pbds.ListUserPrivilegesReq) (
	*pbds.ListUserPrivilegesResp, error) {
	kit := kit.FromGrpcContext(ctx)

	opt := &types.BasePage{Start: req.Start, Limit: uint(req.Limit), All: req.All}
	if err := opt.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	data, count, err := s.dao.UserPrivilege().List(kit, req.BizId, req.AppId, req.TemplateSpaceId, req.Name, opt)
	if err != nil {
		return nil, err
	}

	item := make([]*pbds.ListUserPrivilegesResp_Detail, 0, len(data))

	for _, v := range data {
		item = append(item, &pbds.ListUserPrivilegesResp_Detail{
			Id:            v.ID,
			Name:          v.Spec.User,
			PrivilegeType: string(v.Spec.PrivilegeType),
			ReadOnly:      v.Spec.ReadOnly,
			Pid:           v.Attachment.Uid,
		})
	}

	return &pbds.ListUserPrivilegesResp{Count: uint32(count), Details: item}, nil
}

// DeleteUserPrivilege 删除用户权限
func (s *Service) DeleteUserPrivilege(ctx context.Context, req *pbds.DeleteUserPrivilegesReq) (
	*pbbase.EmptyResp, error) {
	kit := kit.FromGrpcContext(ctx)

	// 获取删除的数据
	item, err := s.dao.UserPrivilege().GetUserPrivilege(kit, req.BizId, req.AppId, req.TemplateSpaceId, req.Id)
	if err != nil {
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "get user privilege failed, err: %v", err))
	}

	if item.Spec.ReadOnly {
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "this user has read-only privilege"))
	}

	tx := s.dao.GenQuery().Begin()

	if err = s.dao.UserPrivilege().DeleteWithTx(kit, tx, item); err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kit.Rid)
		}
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "delete user privilege failed, err: %v", err))
	}

	if req.TemplateSpaceId != 0 && req.AppId == 0 {
		// 删除模板版本中对应的用户权限
		err = s.dao.TemplateRevision().UpdateUserPrivilegesWithTx(kit, tx, req.BizId,
			req.TemplateSpaceId, item.Spec.User, "")
		if err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kit.Rid)
			}
			return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "delete user privilege failed, err: %v", err))
		}
	}

	if req.TemplateSpaceId == 0 && req.AppId != 0 {
		// 删除未命名版本中对应的用户权限
		err = s.dao.ConfigItem().UpdateUserPrivilegesWithTx(kit, tx, req.BizId,
			req.AppId, item.Spec.User, "")
		if err != nil {
			if rErr := tx.Rollback(); rErr != nil {
				logs.Errorf("transaction rollback failed, err: %v, rid: %s", rErr, kit.Rid)
			}
			return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "delete user privilege failed, err: %v", err))
		}
	}

	if err = tx.Commit(); err != nil {
		logs.Errorf("commit transaction failed, err: %v, rid: %s", err, kit.Rid)
		return nil, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "delete user privilege failed, err: %v", err))
	}

	return &pbbase.EmptyResp{}, nil
}

// batcheUpsertPermissionGroup 批量创建或更新权限组(包含用户和用户组)
func (s *Service) batcheUpsertPermissionGroup(kt *kit.Kit, tx *gen.QueryTx, bizID, appID, templateSpaceID uint32,
	userPrivileges []*table.UserPrivilege, userGroupPrivileges []*table.UserGroupPrivilege, isTemplate bool) error {

	// 处理用户权限
	err := s.batcheUpsertUserPrivileges(kt, tx, bizID, appID, templateSpaceID, userPrivileges, isTemplate)
	if err != nil {
		return err
	}

	// 处理用户组权限
	err = s.batcheUpsertUserGroupPrivileges(kt, tx, bizID, appID, templateSpaceID, userGroupPrivileges, isTemplate)
	if err != nil {
		return err
	}

	return nil
}

// upsertSinglePermissions 创建或更新单个权限(用户和用户组)
func (s *Service) upsertSinglePermissions(kit *kit.Kit, tx *gen.QueryTx, bizID, appID, templateSpaceID uint32,
	userPrivilege *table.UserPrivilege, userGroupPrivilege *table.UserGroupPrivilege, isTemplate bool) error {

	// 更新用户权限
	err := s.handleUserPrivilege(kit, tx, bizID, appID, templateSpaceID, userPrivilege, isTemplate)
	if err != nil {
		return err
	}

	// 更新用户组权限
	err = s.handleUserGroupPrivilege(kit, tx, bizID, appID, templateSpaceID, userGroupPrivilege, isTemplate)
	if err != nil {
		return err
	}

	return nil
}

// batcheUpsertUserPrivileges 批量创建或更新用户(只有用户)
func (s *Service) batcheUpsertUserPrivileges(kit *kit.Kit, tx *gen.QueryTx, bizID, appID, templateSpaceID uint32,
	userPrivileges []*table.UserPrivilege, isTemplate bool) error {
	// 创建两个 map 来追踪 uid -> user 和 user -> uid 的唯一性
	// 验证提交的权限数据存在唯一性
	users, uids := make(map[uint32]string), make(map[string]uint32)
	for _, v := range userPrivileges {
		// 不同的用户存在相同的uid
		if user, exists := users[v.Attachment.Uid]; exists && user != v.Spec.User {
			return errors.New(i18n.T(kit, "different user have the same UID, %s", v.Spec.User))
		}

		// 不同的uid存在相同的用户
		if uid, exists := uids[v.Spec.User]; exists && uid != v.Attachment.Uid {
			return errors.New(i18n.T(kit, "different UID have the same user, %s", v.Spec.User))
		}

		users[v.Attachment.Uid] = v.Spec.User
		uids[v.Spec.User] = v.Attachment.Uid
		if err := s.handleUserPrivilege(kit, tx, bizID, appID, templateSpaceID, v, isTemplate); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) handleSingleNonTemplateFilePermissions(kit *kit.Kit, bizID, appID uint32,
	data *table.ConfigItem) (*table.ConfigItem, error) {

	// 查询对应服务下的所有权限
	userPrivileges, err := s.dao.UserPrivilege().ListUserPrivileges(kit, bizID, appID, 0)
	if err != nil {
		return data, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "get permissions under the app failed, err: %v", err))
	}

	userPrivilegesMap := map[string]uint32{}
	for _, v := range userPrivileges {
		userPrivilegesMap[v.User] = v.Uid
	}

	userGroupPrivileges, err := s.dao.UserGroupPrivilege().ListUserGroupPrivileges(kit, bizID, appID, 0)
	if err != nil {
		return data, errf.Errorf(errf.DBOpFailed,
			i18n.T(kit, "get user group permissions under the app failed, err: %v", err))
	}

	userGroupPrivilegesMap := map[string]uint32{}
	for _, v := range userGroupPrivileges {
		userGroupPrivilegesMap[v.UserGroup] = v.Gid
	}

	data.Spec.Permission.Uid = userPrivilegesMap[data.Spec.Permission.User]
	data.Spec.Permission.Gid = userGroupPrivilegesMap[data.Spec.Permission.UserGroup]

	return data, nil
}

// 处理非模板文件权限
func (s *Service) handleNonTemplateFilePermissions(kit *kit.Kit, bizID, appID uint32,
	data []*table.ConfigItem) ([]*table.ConfigItem, error) {

	for _, v := range data {
		permissions, err := s.handleSingleNonTemplateFilePermissions(kit, bizID, appID, v)
		if err != nil {
			return nil, err
		}
		v.Spec.Permission.Uid = permissions.Spec.Permission.Uid
		v.Spec.Permission.Gid = permissions.Spec.Permission.Gid
	}

	return data, nil
}

// 获取并处理 UserPrivilege 数据
func (s *Service) handleUserPrivilege(kit *kit.Kit, tx *gen.QueryTx, bizID, appID, templateSpaceID uint32,
	data *table.UserPrivilege, isTemplate bool) error {

	if data.Spec.User == "root" && data.Attachment.Uid == 0 {
		return nil
	}

	// 验证 user 和 uid
	if err := validateUserAndUid(kit, data); err != nil {
		return err
	}

	existingUID, err := s.dao.UserPrivilege().GetUserPrivilegeByUid(kit, tx, bizID, appID,
		templateSpaceID, data.Attachment.Uid)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "create or update user privilege failed, err: %v", err))
	}

	existingUser, err := s.dao.UserPrivilege().GetUserPrivilegeByUser(kit, tx, bizID, appID, templateSpaceID,
		data.Spec.User)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "create or update user privilege failed, err: %v", err))
	}

	// uid 和 user 都不存在新增数据
	if existingUID == nil && existingUser == nil {
		data.Spec.PrivilegeType = table.PrivilegeTypeCustom
		data.Revision = &table.Revision{Creator: kit.User}
		if _, err := s.dao.UserPrivilege().CreateWithTx(kit, tx, data); err != nil {
			return errf.Errorf(errf.DBOpFailed, i18n.T(kit, "create or update user privilege failed, err: %v", err))
		}
		return nil
	}

	// 如果用户和 UID 相等，则无需处理
	if existingUID != nil && existingUser != nil && existingUID.Attachment.Uid == existingUser.Attachment.Uid &&
		existingUID.Spec.User == existingUser.Spec.User {
		return nil
	}

	// 如果用户或 UID 存在冲突
	if existingUID != nil && existingUser != nil && existingUID.Attachment.Uid != existingUser.Attachment.Uid &&
		existingUID.Spec.User != existingUser.Spec.User {
		return errf.Errorf(errf.DBOpFailed,
			i18n.T(kit, "the user or UID already exists, user: %s, uid: %d", data.Spec.User, data.Attachment.Uid))
	}

	return s.upsertUserPrivilege(kit, tx, bizID, appID, templateSpaceID, data, existingUID, existingUser, isTemplate)
}

// upsertUserPrivilege 更新用户权限数据
// 判断模板和非模板文件，更新对应的表数据
func (s *Service) upsertUserPrivilege(kit *kit.Kit, tx *gen.QueryTx, bizID, appID, templateSpaceID uint32,
	data *table.UserPrivilege, existingUID, existingUser *table.UserPrivilege, isTemplate bool) error {

	data.Revision = &table.Revision{Reviser: kit.User}

	if existingUID == nil {
		return s.dao.UserPrivilege().UpdateWithTx(kit, tx, existingUser.ID, data)
	}

	if existingUser == nil {
		if err := s.dao.UserPrivilege().UpdateWithTx(kit, tx, existingUID.ID, data); err != nil {
			return err
		}
		if isTemplate {
			return s.dao.TemplateRevision().UpdateUserPrivilegesWithTx(kit, tx, bizID,
				templateSpaceID, existingUID.Spec.User, data.Spec.User)
		}
		return s.dao.ConfigItem().UpdateUserPrivilegesWithTx(kit, tx, bizID, appID, existingUID.Spec.User, data.Spec.User)
	}

	return nil
}

// 校验 root 和 UID 逻辑
func validateUserAndUid(kit *kit.Kit, data *table.UserPrivilege) error {
	if data.Spec.User == "root" && data.Attachment.Uid != 0 {
		return errors.New(i18n.T(kit, "user is root but UID is not 0"))
	}
	if data.Spec.User != "root" && data.Attachment.Uid == 0 {
		return errors.New(i18n.T(kit, "UID is 0 but user is not root"))
	}
	return nil
}

// handleTplRevPerms 获取模板版本权限列表
func (s *Service) listTplRevPerms(kit *kit.Kit, bizID uint32,
	tmplRevisions []*table.TemplateRevision) ([]*table.TemplateRevision, error) {

	templateSpaceIDs := []uint32{}
	for _, v := range tmplRevisions {
		templateSpaceIDs = append(templateSpaceIDs, v.Attachment.TemplateSpaceID)
	}

	// 根据空间ID获取用户
	userPrivileges, err := s.dao.UserPrivilege().ListUserPrivsBySpaceIDs(kit, bizID, templateSpaceIDs)
	if err != nil {
		return tmplRevisions, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "get permissions under the app failed, err: %v", err))
	}

	// 初始化模板空间用户权限的 map
	templateSpaceUserPrivs := make(map[uint32]map[string]uint32)

	for _, v := range userPrivileges {
		// 检查并初始化内部 map
		if templateSpaceUserPrivs[v.TemplateSpaceID] == nil {
			templateSpaceUserPrivs[v.TemplateSpaceID] = make(map[string]uint32)
		}
		// 将用户和 UID 添加到内部 map 中
		templateSpaceUserPrivs[v.TemplateSpaceID][v.User] = v.Uid
	}

	// 根据空间ID获取用户组
	userGroupPrivileges, err := s.dao.UserGroupPrivilege().ListGroupPrivsBySpaceIDs(kit, bizID, templateSpaceIDs)
	if err != nil {
		return tmplRevisions, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "get permissions under the app failed, err: %v", err))
	}

	// 初始化模板空间用户权限的 map
	templateSpaceGroupPrivs := make(map[uint32]map[string]uint32)

	for _, v := range userGroupPrivileges {
		// 检查并初始化内部 map
		if templateSpaceGroupPrivs[v.TemplateSpaceID] == nil {
			templateSpaceGroupPrivs[v.TemplateSpaceID] = make(map[string]uint32)
		}
		// 将用户和 GID 添加到内部 map 中
		templateSpaceGroupPrivs[v.TemplateSpaceID][v.UserGroup] = v.Gid
	}

	for _, v := range tmplRevisions {
		v.Spec.Permission.Uid = templateSpaceUserPrivs[v.Attachment.TemplateSpaceID][v.Spec.Permission.User]
		v.Spec.Permission.Gid = templateSpaceGroupPrivs[v.Attachment.TemplateSpaceID][v.Spec.Permission.UserGroup]
	}

	return tmplRevisions, nil
}

// handleSingleTplRevPerm  获取模板版本权限
func (s *Service) getTplRevPerm(kit *kit.Kit, bizID uint32,
	tmplRevisions *table.TemplateRevision) (*table.TemplateRevision, error) {
	if tmplRevisions == nil {
		return nil, nil
	}

	// 根据空间ID获取用户
	userPrivileges, err := s.dao.UserPrivilege().ListUserPrivsBySpaceIDs(kit, bizID,
		[]uint32{tmplRevisions.Attachment.TemplateSpaceID})
	if err != nil {
		return tmplRevisions, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "get permissions under the app failed, err: %v", err))
	}

	// 根据空间ID获取用户组
	userGroupPrivileges, err := s.dao.UserGroupPrivilege().ListGroupPrivsBySpaceIDs(kit, bizID,
		[]uint32{tmplRevisions.Attachment.TemplateSpaceID})
	if err != nil {
		return tmplRevisions, errf.Errorf(errf.DBOpFailed, i18n.T(kit, "get permissions under the app failed, err: %v", err))
	}

	if len(userPrivileges) == 0 && len(userGroupPrivileges) == 0 {
		return tmplRevisions, nil
	}

	tmplRevisions.Spec.Permission.Uid = userPrivileges[0].Uid
	tmplRevisions.Spec.Permission.Gid = userGroupPrivileges[0].Gid

	return tmplRevisions, nil
}

// checkTplNonTplPerms 检测模板和非模板之间的权限
func checkTplNonTplPerms(kit *kit.Kit, configItem []*table.ConfigItem, tmplRevisions []*table.TemplateRevision) error {

	var items []types.FileGroupPrivilege

	collectItems := func(name, path string, uid uint32, user string, gid uint32, userGroup string,
		templateSpaceID uint32) {
		items = append(items, types.FileGroupPrivilege{
			Name:            name,
			Path:            path,
			TemplateSpaceID: templateSpaceID,
			Uid:             uid,
			User:            user,
			Gid:             gid,
			UserGroup:       userGroup,
		})
	}

	for _, v := range configItem {
		collectItems(v.Spec.Name, v.Spec.Path, v.Spec.Permission.Uid, v.Spec.Permission.User,
			v.Spec.Permission.Gid, v.Spec.Permission.UserGroup, 0)
	}
	for _, v := range tmplRevisions {
		collectItems(v.Spec.Name, v.Spec.Path, v.Spec.Permission.Uid, v.Spec.Permission.User,
			v.Spec.Permission.Gid, v.Spec.Permission.UserGroup, v.Attachment.TemplateSpaceID)
	}

	if len(items) > 0 {
		uidToItem, userToItem := make(map[uint32]types.FileGroupPrivilege), make(map[string]types.FileGroupPrivilege)
		gidToItem, userGroupToItem := make(map[uint32]types.FileGroupPrivilege), make(map[string]types.FileGroupPrivilege)

		for _, v := range items {
			// 检查 root 权限
			if err := checkRootUser(kit, v.User, v.Uid, v.UserGroup, v.Gid); err != nil {
				return err
			}

			// 检查 UID 和 User 唯一性
			if err := checkUIDUserUniqueness(kit, v, uidToItem, userToItem); err != nil {
				return err
			}

			// 检查 GID 和 UserGroup 唯一性
			if err := checkGIDUserGroupUniqueness(kit, v, gidToItem, userGroupToItem); err != nil {
				return err
			}
		}
	}

	return nil
}

// 封装检查 UID 和 User 唯一性的逻辑，并生成冲突提示
func checkUIDUserUniqueness(kit *kit.Kit, v types.FileGroupPrivilege, uidToItem map[uint32]types.FileGroupPrivilege,
	userToItem map[string]types.FileGroupPrivilege) error {
	privilegeType := "user"
	// 检查 UID 的唯一性
	if existingItem, exists := uidToItem[v.Uid]; exists && existingItem.User != v.User {
		// 生成冲突提示信息
		return createConflictError(kit, existingItem, v, privilegeType)
	}

	// 检查用户的唯一性
	if existingItem, exists := userToItem[v.User]; exists && existingItem.Uid != v.Uid {
		// 生成冲突提示信息
		return createConflictError(kit, existingItem, v, privilegeType)
	}

	// 如果没有冲突，记录当前项
	uidToItem[v.Uid] = v
	userToItem[v.User] = v

	return nil
}

// 封装检查 GID 和 UserGroup 唯一性的逻辑，并生成冲突提示
func checkGIDUserGroupUniqueness(kit *kit.Kit, v types.FileGroupPrivilege,
	gidToItem map[uint32]types.FileGroupPrivilege, userGroupToItem map[string]types.FileGroupPrivilege) error {

	privilegeType := "user group"
	// 检查 UID 的唯一性
	if existingItem, exists := gidToItem[v.Gid]; exists && existingItem.UserGroup != v.UserGroup {
		// 生成冲突提示信息
		return createConflictError(kit, existingItem, v, privilegeType)
	}

	// 检查用户的唯一性
	if existingItem, exists := userGroupToItem[v.UserGroup]; exists && existingItem.Gid != v.Gid {
		// 生成冲突提示信息
		return createConflictError(kit, existingItem, v, privilegeType)
	}
	// 如果没有冲突，记录当前项
	gidToItem[v.Gid] = v
	userGroupToItem[v.UserGroup] = v

	return nil
}

// 生成冲突提示的函数
func createConflictError(kit *kit.Kit, item1, item2 types.FileGroupPrivilege, privilegeType string) error {
	// 基础的冲突提示
	conflictMsg := fmt.Sprintf("file %s and file %s have %s permission conflicts",
		path.Join(item1.Path, item1.Name), path.Join(item2.Path, item2.Name), privilegeType)

	// 如果其中任一文件存在 TemplateSpaceID，则需要在提示中包含该信息
	// TODO 是否判断那个空间下的文件错误
	// if item1.TemplateSpaceID != 0 || item2.TemplateSpaceID != 0 {
	// 	return errors.New(i18n.T(kit, "TemplateSpaceID involved: %d, %d. "+conflictMsg,
	// 		item1.TemplateSpaceID, item2.TemplateSpaceID))
	// }

	// 返回标准冲突提示
	return errors.New(i18n.T(kit, conflictMsg))
}

// 检查 root 权限的逻辑
func checkRootUser(kit *kit.Kit, user string, uid uint32, userGroup string, gid uint32) error {
	// 定义一个检查函数
	check := func(name string, id uint32) error {
		if name == "root" && id == 0 {
			return nil
		}
		if name == "root" && id != 0 {
			return errors.New(i18n.T(kit, "%s is root but ID is not 0", name))
		}
		if name != "root" && id == 0 {
			return errors.New(i18n.T(kit, "ID is 0 but %s is not root", name))
		}
		return nil
	}

	// 检查 user 和 uid
	if err := check(user, uid); err != nil {
		return err
	}

	// 检查 userGroup 和 gid
	return check(userGroup, gid)
}
