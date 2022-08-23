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
 *
 */

package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2022-03-01/containerservice"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// ContainerServiceClient container service client
type ContainerServiceClient struct {
	resourceGroupName     string
	managedClustersClient containerservice.ManagedClustersClient
}

// NewContainerServiceClient create container service client
func NewContainerServiceClient(opt *cloudprovider.CommonOption) (*ContainerServiceClient, error) {
	if opt == nil || opt.Account == nil {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Account.SubscriptionID) == 0 || len(opt.Account.TenantID) == 0 ||
		len(opt.Account.ClientID) == 0 || len(opt.Account.ClientSecret) == 0 ||
		len(opt.Account.ResourceGroupName) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}

	// get Authorizer
	authorizer, err := getAuthorizer(opt.Account.TenantID, opt.Account.ClientID, opt.Account.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("get authorizer error: %v", err)
	}

	// new ManagedClustersClient
	managedClustersClient := containerservice.NewManagedClustersClient(opt.Account.SubscriptionID)
	managedClustersClient.Authorizer = authorizer
	return &ContainerServiceClient{
		managedClustersClient: managedClustersClient,
		resourceGroupName:     opt.Account.ResourceGroupName,
	}, nil
}

// ListCluster list clusters
func (cs *ContainerServiceClient) ListCluster(ctx context.Context, location string) ([]containerservice.ManagedCluster, error) {
	pager, err := cs.managedClustersClient.ListByResourceGroup(ctx, cs.resourceGroupName)
	if err != nil {
		return nil, err
	}
	result := make([]containerservice.ManagedCluster, 0)
	for pager.NotDone() {
		for _, v := range pager.Values() {
			if len(location) != 0 && *v.Location != location {
				continue
			}
			result = append(result, v)
		}
		if err := pager.NextWithContext(ctx); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func getClusterIDFromARM(arm string) string {
	s := strings.Split(arm, "/")
	return s[len(s)-1]
}

// GetCluster get cluster
func (cs *ContainerServiceClient) GetCluster(ctx context.Context, clusterName string) (containerservice.ManagedCluster, error) {
	return cs.managedClustersClient.Get(ctx, cs.resourceGroupName, getClusterIDFromARM(clusterName))
}

// GetClusterCredentials get cluster credentials
func (cs *ContainerServiceClient) GetClusterCredentials(ctx context.Context, clusterName string) (
	[]containerservice.CredentialResult, error) {
	resp, err := cs.managedClustersClient.ListClusterAdminCredentials(ctx, cs.resourceGroupName, getClusterIDFromARM(clusterName), "")
	if err != nil {
		return nil, err
	}
	if resp.Kubeconfigs == nil || len(*resp.Kubeconfigs) == 0 {
		return nil, fmt.Errorf("kubeconfigs is empty")
	}
	return *resp.Kubeconfigs, nil
}

// DeleteCluster delete cluster
func (cs *ContainerServiceClient) DeleteCluster(ctx context.Context, clusterName string) error {
	future, err := cs.managedClustersClient.Delete(ctx, cs.resourceGroupName, getClusterIDFromARM(clusterName))
	if err != nil {
		return err
	}
	err = future.WaitForCompletionRef(ctx, cs.managedClustersClient.Client)
	if err != nil {
		return err
	}
	_, err = future.Result(cs.managedClustersClient)
	return err
}
