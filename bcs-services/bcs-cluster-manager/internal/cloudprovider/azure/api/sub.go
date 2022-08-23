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

	"github.com/Azure/azure-sdk-for-go/services/subscription/mgmt/2020-09-01/subscription"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// SubClient defines the interface for azure subscription client
type SubClient struct {
	resourceGroupName string
	subscriptionID    string
	subClient         subscription.SubscriptionsClient
}

// NewAMClient create azure api management client
func NewAMClient(opt *cloudprovider.CommonOption) (*SubClient, error) {
	if opt == nil || opt.Account == nil {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Account.SubscriptionID) == 0 || len(opt.Account.TenantID) == 0 ||
		len(opt.Account.ClientID) == 0 || len(opt.Account.ClientSecret) == 0 || len(opt.Account.ResourceGroupName) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}

	// get Authorizer
	authorizer, err := getAuthorizer(opt.Account.TenantID, opt.Account.ClientID, opt.Account.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("get authorizer error: %v", err)
	}

	// new subscriptions client
	subClient := subscription.NewSubscriptionsClient()
	subClient.Authorizer = authorizer
	return &SubClient{
		resourceGroupName: opt.Account.ResourceGroupName,
		subscriptionID:    opt.Account.SubscriptionID,
		subClient:         subClient,
	}, nil
}

// ListLocations list locations
func (sub *SubClient) ListLocations(ctx context.Context) ([]*proto.RegionInfo, error) {
	locations, err := sub.subClient.ListLocations(ctx, sub.subscriptionID)
	if err != nil {
		return nil, fmt.Errorf("list locations error: %v", err)
	}

	if locations.Value == nil {
		return nil, nil
	}

	result := make([]*proto.RegionInfo, 0)
	for _, v := range *locations.Value {
		if v.Name != nil && v.DisplayName != nil {
			result = append(result, &proto.RegionInfo{
				Region:     *v.Name,
				RegionName: *v.DisplayName,
			})
		}
	}
	return result, nil
}
