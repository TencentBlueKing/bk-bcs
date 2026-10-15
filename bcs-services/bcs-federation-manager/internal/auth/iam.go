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
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/auth/iam"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cloudaccount"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/cluster"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/namespace"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/project"
)

var (
	iamClientMu      sync.Mutex
	iamClients       = make(map[string]iam.PermClient)
	iamClientFactory func(tenantID string) (iam.PermClient, error)

	// ProjectIamClient project iam client
	ProjectIamClient func(tenantID string) (*project.BCSProjectPerm, error)
	// ClusterIamClient cluster iam client
	ClusterIamClient func(tenantID string) (*cluster.BCSClusterPerm, error)
	// NamespaceIamClient namespace iam client
	NamespaceIamClient func(tenantID string) (*namespace.BCSNamespacePerm, error)
	// CloudAccountIamClient cloudaccount iam client
	CloudAccountIamClient func(tenantID string) (*cloudaccount.BCSCloudAccountPerm, error)
)

// SetIAMClientFactory registers the factory used to create tenant-scoped IAM clients.
func SetIAMClientFactory(factory func(tenantID string) (iam.PermClient, error)) {
	iamClientMu.Lock()
	defer iamClientMu.Unlock()
	iamClientFactory = factory
	iamClients = make(map[string]iam.PermClient)
}

// IAMClient returns a cached IAM client for the tenant. Empty tenantID uses the default tenant.
func IAMClient(tenantID string) (iam.PermClient, error) {
	if tenantID == "" {
		tenantID = iam.DefaultTenantId
	}

	iamClientMu.Lock()
	defer iamClientMu.Unlock()
	if cli, ok := iamClients[tenantID]; ok {
		return cli, nil
	}
	if iamClientFactory == nil {
		return nil, fmt.Errorf("iam client factory is not initialized")
	}
	cli, err := iamClientFactory(tenantID)
	if err != nil {
		return nil, err
	}
	iamClients[tenantID] = cli
	return cli, nil
}

// InitPermClient binds perm helpers to the tenant-scoped IAM client cache.
func InitPermClient() {
	ProjectIamClient = func(tenantID string) (*project.BCSProjectPerm, error) {
		cli, err := IAMClient(tenantID)
		if err != nil {
			return nil, err
		}
		return project.NewBCSProjectPermClient(cli), nil
	}
	ClusterIamClient = func(tenantID string) (*cluster.BCSClusterPerm, error) {
		cli, err := IAMClient(tenantID)
		if err != nil {
			return nil, err
		}
		return cluster.NewBCSClusterPermClient(cli), nil
	}
	NamespaceIamClient = func(tenantID string) (*namespace.BCSNamespacePerm, error) {
		cli, err := IAMClient(tenantID)
		if err != nil {
			return nil, err
		}
		return namespace.NewBCSNamespacePermClient(cli), nil
	}
	CloudAccountIamClient = func(tenantID string) (*cloudaccount.BCSCloudAccountPerm, error) {
		cli, err := IAMClient(tenantID)
		if err != nil {
			return nil, err
		}
		return cloudaccount.NewBCSAccountPermClient(cli), nil
	}
}
