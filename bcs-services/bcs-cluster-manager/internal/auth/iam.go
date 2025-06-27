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

package auth

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
)

// GetProjectIamClient project iam client
func GetProjectIamClient(tenantId string) (*project.BCSProjectPerm, error) {
	iamClient, err := InitPermClient(tenantId)
	if err != nil {
		return nil, err
	}

	return project.NewBCSProjectPermClient(iamClient), nil
}

// GetClusterIamClient cluster iam client
func GetClusterIamClient(tenantId string) (*cluster.BCSClusterPerm, error) {
	iamClient, err := InitPermClient(tenantId)
	if err != nil {
		return nil, err
	}

	return cluster.NewBCSClusterPermClient(iamClient), nil
}

// GetCloudAccountIamClient cloud account client
func GetCloudAccountIamClient(tenantId string) (*cloudaccount.BCSCloudAccountPerm, error) {
	iamClient, err := InitPermClient(tenantId)
	if err != nil {
		return nil, err
	}

	return cloudaccount.NewBCSAccountPermClient(iamClient), nil
}

// InitPermClient init perm client
func InitPermClient(tenantId string) (iam.PermClient, error) {
	var err error
	iamClient, err := iam.NewIamClient(&iam.Options{
		SystemID:    options.GetGlobalCMOptions().IAM.SystemID,
		AppCode:     options.GetGlobalCMOptions().IAM.AppCode,
		AppSecret:   options.GetGlobalCMOptions().IAM.AppSecret,
		External:    options.GetGlobalCMOptions().IAM.External,
		GateWayHost: options.GetGlobalCMOptions().IAM.GatewayServer,
		IAMHost:     options.GetGlobalCMOptions().IAM.IAMServer,
		BkiIAMHost:  options.GetGlobalCMOptions().IAM.BkiIAMServer,
		Metric:      options.GetGlobalCMOptions().IAM.Metric,
		Debug:       options.GetGlobalCMOptions().IAM.Debug,
		TenantId:    tenantId,
	})

	if err != nil {
		return nil, err
	}

	return iamClient, nil
}
