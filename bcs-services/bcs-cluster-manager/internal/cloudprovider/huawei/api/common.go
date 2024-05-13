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

// Package api xxx
package api

import (
	"fmt"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/global"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// getProjectAuth get project level auth
func getProjectAuth(ak, sk, projectId string) (*basic.Credentials, error) {
	return basic.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		WithProjectId(projectId).
		SafeBuild()
}

// getGlobalAuth get global level auth
func getGlobalAuth(ak, sk string) (*global.Credentials, error) {
	return global.NewCredentialsBuilder().
		WithAk(ak).
		WithSk(sk).
		SafeBuild()
}

// GetProjectIDByRegion get project ID by region (项目名称和regionId相同)
func GetProjectIDByRegion(opt *cloudprovider.CommonOption) (string, error) {
	client, err := NewIamClient(opt)
	if err != nil {
		return "", err
	}

	projects, err := client.ListProjects(opt.Region)
	if err != nil {
		return "", err
	}

	if len(projects) == 0 {
		return "", fmt.Errorf("project not found")
	}

	if len(projects) > 1 {
		return "", fmt.Errorf("the number of project is greater than one")
	}

	return projects[0].Id, nil
}
