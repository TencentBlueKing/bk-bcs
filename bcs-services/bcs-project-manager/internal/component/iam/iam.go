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

// Package iam xxx
package iam

import (
	"context"
	"fmt"

	bcsIAM "github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/manager"
	authutils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"github.com/parnurzeal/gorequest"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bkuser"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
)

var (
	grantActionPath = "/api/v1/open/authorization/resource_creator_action/"
	timeout         = 10
)

// GrantProjectCreatorActions grant create action perm for project
func GrantProjectCreatorActions(ctx context.Context, username string, projectID string, projectName string) error {
	iamConf := config.GlobalConf.IAM
	// 使用网关访问
	reqURL := fmt.Sprintf("%s%s", iamConf.GatewayHost, grantActionPath)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "POST",
		Data: map[string]interface{}{
			"creator": username,
			"system":  bcsIAM.SystemIDBKBCS,
			"type":    "project",
			"id":      projectID,
			"name":    projectName,
		},
	}

	// auth headers
	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("GrantProjectCreatorActions get auth header failed, %s", err.Error())
		return errorx.NewRequestIAMErr(err.Error())
	}

	// 请求API
	proxy := ""
	_, err = component.Request(req, timeout, proxy, headers)
	if err != nil {
		logging.Error("grant creator actions for project failed, %s", err.Error())
		return errorx.NewRequestIAMErr(err.Error())
	}
	return nil
}

// GrantNamespaceCreatorActions grant create action perm for namespace
func GrantNamespaceCreatorActions(ctx context.Context, username, clusterID, namespace string) error {
	iamConf := config.GlobalConf.IAM
	// 使用网关访问
	reqURL := fmt.Sprintf("%s%s", iamConf.GatewayHost, grantActionPath)
	id := authutils.CalcIAMNsID(clusterID, namespace)
	req := gorequest.SuperAgent{
		Url:    reqURL,
		Method: "POST",
		Data: map[string]interface{}{
			"creator": username,
			"system":  bcsIAM.SystemIDBKBCS,
			"type":    "namespace",
			"id":      id,
			"name":    namespace,
		},
	}

	// auth headers
	headers, err := bkuser.GetAuthHeader(ctx)
	if err != nil {
		logging.Error("GrantProjectCreatorActions get auth header failed, %s", err.Error())
		return errorx.NewRequestIAMErr(err.Error())
	}

	// 请求API
	proxy := ""
	_, err = component.Request(req, timeout, proxy, headers)
	if err != nil {
		logging.Error("grant creator actions for namespace failed, %s", err.Error())
		return errorx.NewRequestIAMErr(err.Error())
	}
	return nil
}

func projectGradeManageName(name string) string {
	return fmt.Sprintf("项目[%s]分级管理员", name)
}

func projectGradeManageDesc(name string) string {
	return fmt.Sprintf("蓝鲸容器服务平台（TKEx-IEG）下项目[%s]的分级管理员角色，可为该项目下所有资源进行授权", name)
}

func projectManagerUserGroupName(name string) string {
	return fmt.Sprintf("项目[%s]管理权限用户组", name)
}

func projectManagerUserGroupDesc(name string) string {
	return fmt.Sprintf("可管理项目[%s]下所有资源信息，同时具备查看与操作权限", name)
}

func projectViewerUserGroupName(name string) string {
	return fmt.Sprintf("项目[%s]查看权限用户组", name)
}

func projectViewerUserGroupDesc(name string) string {
	return fmt.Sprintf("仅可查看项目[%s]下所有资源信息，无操作权限", name)
}

// CreateProjectPermManager create perm manager for project
func CreateProjectPermManager(projectID, projectName string, users []string) error {
	gradeID, err := auth.PermManagerClient.CreateProjectGradeManager(context.Background(), users,
		&manager.GradeManagerInfo{
			Name: projectGradeManageName(projectName),
			Desc: projectGradeManageDesc(projectName),
			Project: &manager.Project{
				ProjectID:   projectID,
				ProjectCode: projectName,
				Name:        projectName,
			},
		})
	if err != nil {
		logging.Error("CreateProjectGradeManager CreateProjectGradeManager failed: %v", err)
		return err
	}

	err = auth.PermManagerClient.CreateProjectUserGroup(context.Background(), gradeID, manager.UserGroupInfo{
		Name:  projectManagerUserGroupName(projectName),
		Desc:  projectManagerUserGroupDesc(projectName),
		Users: users,
		Project: &manager.Project{
			ProjectID:   projectID,
			ProjectCode: projectName,
			Name:        projectName,
		},
		Policy: manager.Manager,
	})
	if err != nil {
		logging.Error("CreateProjectGradeManager CreateProjectUserGroup[manager] failed: %v", err)
		return err
	}

	err = auth.PermManagerClient.CreateProjectUserGroup(context.Background(), gradeID, manager.UserGroupInfo{
		Name:  projectViewerUserGroupName(projectName),
		Desc:  projectViewerUserGroupDesc(projectName),
		Users: users,
		Project: &manager.Project{
			ProjectID:   projectID,
			ProjectCode: projectName,
			Name:        projectName,
		},
		Policy: manager.Viewer,
	})
	if err != nil {
		logging.Error("CreateProjectGradeManager CreateProjectUserGroup[manager] failed: %v", err)
		return err
	}

	logging.Info("CreateProjectGradeManager[%s] successful", projectID)
	return nil
}
