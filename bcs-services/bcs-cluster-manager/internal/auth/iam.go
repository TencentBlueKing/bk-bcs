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
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
)

var (
	iamClientMu      sync.Mutex
	iamClients       = make(map[string]iam.PermClient)
	iamClientFactory = defaultIAMClientFactory
)

func defaultIAMClientFactory(tenantID string) (iam.PermClient, error) {
	iamOpt := options.GetGlobalCMOptions().IAM
	opt := &iam.Options{
		SystemID:    iamOpt.SystemID,
		AppCode:     iamOpt.AppCode,
		AppSecret:   iamOpt.AppSecret,
		External:    iamOpt.External,
		GateWayHost: iamOpt.GatewayServer,
		IAMHost:     iamOpt.IAMServer,
		BkiIAMHost:  iamOpt.BkiIAMServer,
		Metric:      iamOpt.Metric,
		Debug:       iamOpt.Debug,
		TenantId:    tenantID,
	}
	iam.ApplyV4Config(opt, iamOpt.EnableV4, iamOpt.V4GateWayHost)
	return iam.NewIamClient(opt)
}

func setIAMClientFactory(factory func(tenantID string) (iam.PermClient, error)) {
	iamClientMu.Lock()
	defer iamClientMu.Unlock()
	if factory == nil {
		iamClientFactory = defaultIAMClientFactory
	} else {
		iamClientFactory = factory
	}
	iamClients = make(map[string]iam.PermClient)
}

func resetIAMClientCache() {
	setIAMClientFactory(nil)
}

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

// InitPermClient returns a cached IAM client for the tenant. Empty tenantID uses the default tenant.
func InitPermClient(tenantId string) (iam.PermClient, error) {
	if tenantId == "" {
		tenantId = iam.DefaultTenantId
	}

	iamClientMu.Lock()
	defer iamClientMu.Unlock()
	if cli, ok := iamClients[tenantId]; ok {
		return cli, nil
	}

	iamClient, err := iamClientFactory(tenantId)
	if err != nil {
		return nil, err
	}
	iamClients[tenantId] = iamClient
	return iamClient, nil
}
