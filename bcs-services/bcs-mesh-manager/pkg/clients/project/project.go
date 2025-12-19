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

// Package project 获取project client
package project

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
)

// GetClient 获取project manager client
func GetClient() (*bcsproject.ProjectClient, func(), error) {
	projectClient, closeFunc, err := bcsproject.GetClient(common.ProjectManagerServiceName)
	if err != nil {
		return nil, nil, err
	}
	return projectClient, closeFunc, nil
}

// GetProjectByCode 根据项目代码获取项目信息
func GetProjectByCode(ctx context.Context, projectCode string) (*bcsproject.Project, error) {
	projectClient, closeFunc, err := GetClient()
	if err != nil {
		blog.Errorf("GetProjectByCode: failed to get project client, error: %s", err.Error())
		return nil, fmt.Errorf("failed to get project client: %s", err.Error())
	}
	defer closeFunc()

	projectInfo, err := projectClient.Project.GetProject(ctx, &bcsproject.GetProjectRequest{
		ProjectIDOrCode: projectCode,
	})
	if err != nil {
		blog.Errorf("GetProjectByCode: GetProject RPC call failed for project code %s, error: %s",
			projectCode, err.Error())
		return nil, err
	}

	if projectInfo.Data == nil {
		blog.Errorf("GetProjectByCode: project data is nil for project code: %s", projectCode)
		return nil, fmt.Errorf("project data is nil for project code: %s", projectCode)
	}

	return projectInfo.Data, nil
}
