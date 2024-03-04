/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package api

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/pkg/errors"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// ResourceGroupsClient defines the interface for azure subscription client
type ResourceGroupsClient struct {
	subscriptionID string
	groupClient    *armresources.ResourceGroupsClient
}

// NewResourceGroupsClient create azure api resource group client
func NewResourceGroupsClient(opt *cloudprovider.CommonOption) (*ResourceGroupsClient, error) {
	if opt == nil || opt.Account == nil {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Account.SubscriptionID) == 0 || len(opt.Account.TenantID) == 0 ||
		len(opt.Account.ClientID) == 0 || len(opt.Account.ClientSecret) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}

	// get Authorizer
	cred, err := getClientCredential(opt.Account.TenantID, opt.Account.ClientID, opt.Account.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("get authorizer error: %v", err)
	}
	clientFactory, err := armresources.NewClientFactory(opt.Account.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("create client factory error: %v", err)
	}

	return &ResourceGroupsClient{
		subscriptionID: opt.Account.SubscriptionID,
		groupClient:    clientFactory.NewResourceGroupsClient(),
	}, nil
}

// ListResourceGroups get azure resource groups list
func (group *ResourceGroupsClient) ListResourceGroups(ctx context.Context) ([]*proto.ResourceGroupInfo, error) {
	pager := group.groupClient.NewListPager(nil)
	result := make([]*armresources.ResourceGroup, 0)
	for pager.More() {
		next, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list locations")
		}
		result = append(result, next.Value...)
	}

	groupsInfo := make([]*proto.ResourceGroupInfo, 0)
	for _, g := range result {
		groupsInfo = append(groupsInfo, &proto.ResourceGroupInfo{
			Name:   *g.Name,
			Region: *g.Location,
			ProvisioningState: func() string {
				if g.Properties != nil && g.Properties.ProvisioningState != nil {
					return *g.Properties.ProvisioningState
				}
				return ""
			}(),
		})
	}

	return groupsInfo, nil
}
