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

// Package iam client
package iam

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

// GetIAMClient get iam client
func GetIAMClient(tenantID string) (iam.PermClient, error) {
	return iam.NewIamClient(&iam.Options{
		SystemID:    config.G.Base.SystemID,
		AppCode:     config.G.Base.AppCode,
		AppSecret:   config.G.Base.AppSecret,
		External:    config.G.IAM.External,
		GateWayHost: config.G.IAM.GatewayServer,
		IAMHost:     config.G.IAM.IAMServer,
		BkiIAMHost:  config.G.IAM.BkIAMServer,
		Metric:      config.G.IAM.Metric,
		Debug:       config.G.IAM.Debug,
		TenantId:    tenantID,
	})
}

// GetProjectPermClient get project perm iam client
func GetProjectPermClient(tenantID string) (*project.BCSProjectPerm, error) {
	iamClient, err := GetIAMClient(tenantID)
	if err != nil {
		return nil, err
	}
	return project.NewBCSProjectPermClient(iamClient), nil
}

// GetClusterPermClient get cluster perm iam client
func GetClusterPermClient(tenantID string) (*cluster.BCSClusterPerm, error) {
	iamClient, err := GetIAMClient(tenantID)
	if err != nil {
		return nil, err
	}
	return cluster.NewBCSClusterPermClient(iamClient), nil
}
