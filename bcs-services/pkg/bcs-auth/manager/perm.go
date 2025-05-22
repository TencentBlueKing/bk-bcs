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

package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	tiam "github.com/TencentBlueKing/iam-go-sdk"
	blog "k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
)

// PermManager manage active grant user permission
type PermManager struct {
	iamClient iam.PermClient
}

// NewBCSPermManagerClient init perm manager client
func NewBCSPermManagerClient(cli iam.PermClient) *PermManager {
	return &PermManager{iamClient: cli}
}

// CreateProjectGradeManager create project grade manager
func (pm *PermManager) CreateProjectGradeManager(ctx context.Context, users []string, info *GradeManagerInfo) (uint64,
	error) {
	if pm == nil {
		return 0, utils.ErrServerNotInited
	}

	if len(users) == 0 {
		return 0, fmt.Errorf("CreateProjectGradeManager failed: users empty")
	}

	if err := info.Project.validate(); err != nil {
		return 0, err
	}

	authScopes := info.Project.BuildScopePerm()
	req := iam.GradeManagerRequest{
		System:              iam.SystemIDBKBCS,
		Name:                info.Name,
		Description:         info.Desc,
		Members:             users,
		AuthorizationScopes: authScopes,
		SubjectScopes:       []tiam.Subject{iam.GlobalSubjectUser},
		TenantId:            info.Project.getTenantId(),
	}
	managerID, err := pm.iamClient.CreateGradeManagers(ctx, req)
	if err != nil {
		blog.Errorf("CreateProjectGradeManager failed: %v", err)
		return 0, err
	}

	return managerID, err
}

// CreateProjectUserGroup create manager or viewer user group for project level
func (pm *PermManager) CreateProjectUserGroup(ctx context.Context, gradeManagerID uint64, info UserGroupInfo) error {
	if pm == nil {
		return utils.ErrServerNotInited
	}

	// step1: create project manager userGroup
	// 用户组名称保持全局唯一
	groupID, err := pm.createUserGroup(ctx, gradeManagerID, info)
	if err != nil {
		blog.Errorf("CreateProjectManagerUserGroup createUserGroup failed: %v", err)
		return err
	}
	// step2: add userGroup members
	err = pm.addUserGroupMember(ctx, info.Project.getTenantId(), groupID, info.Users)
	if err != nil {
		blog.Errorf("CreateProjectManagerUserGroup addUserGroupMember failed: %v", err)
		return err
	}
	// step3: build userGroup AuthorizationScope
	pm.createProjectUserGroupPolicy(ctx, info.Project.getTenantId(), groupID, info.Policy, info.Project)

	return nil
}

// CreateClusterGradeManager cluster level grade manager
func (pm *PermManager) CreateClusterGradeManager(ctx context.Context, users []string, cls *Cluster) (uint64, error) {
	if pm == nil {
		return 0, utils.ErrServerNotInited
	}

	if len(users) == 0 {
		return 0, fmt.Errorf("CreateClusterGradeManager failed: users empty")
	}

	if err := cls.validate(); err != nil {
		return 0, err
	}

	authScopes := cls.BuildScopePerm()

	req := iam.GradeManagerRequest{
		System:              iam.SystemIDBKBCS,
		Name:                fmt.Sprintf("集群[%s]分级管理员", cls.ClusterName),
		Description:         "容器管理平台集群维度默认分级管理员",
		Members:             users,
		AuthorizationScopes: authScopes,
		SubjectScopes:       []tiam.Subject{iam.GlobalSubjectUser},
		TenantId:            cls.getTenantId(),
	}
	managerID, err := pm.iamClient.CreateGradeManagers(ctx, req)
	if err != nil {
		blog.Errorf("CreateClusterGradeManager failed: %v", err)
		return 0, err
	}

	return managerID, err
}

// CreateClusterUserGroup create manager or viewer user group for project level
func (pm *PermManager) CreateClusterUserGroup(ctx context.Context, gradeManagerID uint64, info UserGroupInfo) error {
	if pm == nil {
		return utils.ErrServerNotInited
	}

	// step1: create project manager userGroup
	// 用户组名称保持全局唯一
	groupID, err := pm.createUserGroup(ctx, gradeManagerID, info)
	if err != nil {
		blog.Errorf("CreateProjectManagerUserGroup createUserGroup failed: %v", err)
		return err
	}
	// step2: add userGroup members
	err = pm.addUserGroupMember(ctx, info.Project.getTenantId(), groupID, info.Users)
	if err != nil {
		blog.Errorf("CreateProjectManagerUserGroup addUserGroupMember failed: %v", err)
		return err
	}
	// step3: build userGroup AuthorizationScope
	pm.createProjectUserGroupPolicy(ctx, info.Project.getTenantId(), groupID, info.Policy, info.Project)

	return nil
}

// createUserGroup create single userGroup
func (pm *PermManager) createUserGroup(ctx context.Context, gradeManagerID uint64, info UserGroupInfo) (uint64, error) {
	if pm == nil {
		return 0, utils.ErrServerNotInited
	}

	// 用户组名称保持全局唯一
	userGroups := make([]iam.UserGroup, 0)
	userGroups = append(userGroups, iam.UserGroup{
		Name:        info.Name,
		Description: info.Desc,
	})
	groups, err := pm.iamClient.CreateUserGroup(ctx, gradeManagerID, iam.CreateUserGroupRequest{
		Groups:   userGroups,
		TenantId: info.Project.getTenantId(),
	})
	if err != nil {
		blog.Errorf("createUserGroup gradeManager[%v] failed: %v", gradeManagerID, err)
		return 0, err
	}
	blog.Infof("createUserGroup gradeManager[%v] successful", gradeManagerID)

	return groups[0], nil
}

// addUserGroupMember xxx
// createUserGroup create single userGroup
func (pm *PermManager) addUserGroupMember(ctx context.Context, tenantId string,
	userGroupID uint64, members []string) error {
	if pm == nil {
		return utils.ErrServerNotInited
	}

	if tenantId == "" {
		tenantId = utils.DefaultTenantId
	}

	subjects := make([]tiam.Subject, 0)
	for i := range members {
		subjects = append(subjects, tiam.Subject{Type: iam.User.String(), ID: members[i]})
	}

	err := pm.iamClient.AddUserGroupMembers(ctx, tenantId, userGroupID, iam.AddGroupMemberRequest{
		Members:  subjects,
		// default ExpiredAt half year
		ExpiredAt: int(time.Now().Add(time.Hour * 24 * 30 * 6).Unix()),
	})
	if err != nil {
		blog.Errorf("addUserGroupMember userGroup[%v] failed: %v", userGroupID, err)
		return err
	}
	blog.Infof("addUserGroupMember userGroup[%v] successful", userGroupID)

	return nil
}

func (pm *PermManager) createProjectUserGroupPolicy(ctx context.Context, tenantId string, groupID uint64,
	pType PolicyType, p *Project) {
	scopeFuncs := make([]AuthorizationScopeFunc, 0)

	switch pType {
	case Manager:
		scopeFuncs = append(scopeFuncs, p.buildProjectCreateScopePerm, p.buildProjectOtherScope,
			p.buildClusterCreateScope, p.buildClusterOtherScope, p.buildClusterScopedScope,
			p.buildNamespaceCreateListScope, p.buildNamespaceOtherScope, p.buildNamespaceScopedScope,
			p.buildTemplateSetCreateScope, p.buildTemplateSetOtherScope,
			p.buildCloudAccountScope)
	case Viewer:
		scopeFuncs = append(scopeFuncs, p.buildProjectViewScope, p.buildClusterViewScope, p.buildClusterScopedViewScope,
			p.buildNamespaceListScope, p.buildNamespaceViewScope, p.buildNamespaceScopedViewScope,
			p.buildTemplateSetViewScope)
	default:
	}

	var (
		wg = sync.WaitGroup{}
	)
	for i := range scopeFuncs {
		wg.Add(1)
		go func(scope iam.AuthorizationScope) {
			defer wg.Done()
			err := pm.iamClient.CreateUserGroupPolicies(ctx, tenantId, groupID, scope)
			if err != nil {
				blog.Errorf("createProjectUserGroupPolicy CreateUserGroupPolicies failed: %v", err)
				return
			}
		}(scopeFuncs[i]())
	}
	wg.Wait()

	blog.Infof("createProjectUserGroupPolicy[%v] successful", groupID)
}

// nolint
func (pm *PermManager) createClusterUserGroupPolicy(ctx context.Context, groupID uint64, pType PolicyType, c *Cluster) {
	scopeFuncs := make([]AuthorizationScopeFunc, 0)

	switch pType {
	case Manager:
		scopeFuncs = append(scopeFuncs, c.buildClusterCreateScope, c.buildClusterOtherScope, c.buildClusterScopedScope,
			c.buildNamespaceCreateListScope, c.buildNamespaceOtherScope, c.buildNamespaceScopedScope)
	case Viewer:
		scopeFuncs = append(scopeFuncs, c.buildClusterViewScope, c.buildClusterScopedViewScope,
			c.buildNamespaceViewScope, c.buildNamespaceListScope, c.buildNamespaceScopedViewScope)
	default:
	}

	var (
		wg = sync.WaitGroup{}
	)
	for i := range scopeFuncs {
		wg.Add(1)
		go func(scope iam.AuthorizationScope) {
			defer wg.Done()
			err := pm.iamClient.CreateUserGroupPolicies(ctx, c.getTenantId(), groupID, scope)
			if err != nil {
				blog.Errorf("createClusterUserGroupPolicy CreateUserGroupPolicies failed: %v", err)
				return
			}
		}(scopeFuncs[i]())
	}
	wg.Wait()

	blog.Infof("createClusterUserGroupPolicy[%v] successful", groupID)
}
