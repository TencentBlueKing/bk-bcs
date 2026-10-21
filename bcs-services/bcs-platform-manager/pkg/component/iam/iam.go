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
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
)

var (
	iamClientMu sync.Mutex
	iamClients  = make(map[string]iam.PermClient)
)

// GetIAMClient returns a cached IAM client for the tenant. Empty tenantID uses the default tenant.
func GetIAMClient(tenantID string) (iam.PermClient, error) {
	if tenantID == "" {
		tenantID = iam.DefaultTenantId
	}

	iamClientMu.Lock()
	defer iamClientMu.Unlock()
	if cli, ok := iamClients[tenantID]; ok {
		return cli, nil
	}

	opt := &iam.Options{
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
	}
	iam.ApplyV4Config(opt, config.G.IAM.EnableV4, config.G.IAM.V4GateWayHost)
	cli, err := iam.NewIamClient(opt)
	if err != nil {
		return nil, err
	}
	iamClients[tenantID] = cli
	return cli, nil
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

// GetNsScopePermClient get namespace scoped perm iam client
func GetNsScopePermClient(tenantID string) (*namespace.BCSNamespacePerm, error) {
	iamClient, err := GetIAMClient(tenantID)
	if err != nil {
		return nil, err
	}
	return namespace.NewBCSNamespacePermClient(iamClient), nil
}
