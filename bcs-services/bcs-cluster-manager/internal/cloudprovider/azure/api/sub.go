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

package api

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/pkg/errors"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

// SubClient defines the interface for azure subscription client
type SubClient struct {
	subscriptionID string
	subClient      *armsubscriptions.Client
}

// NewAMClient create azure api management client
func NewAMClient(opt *cloudprovider.CommonOption) (*SubClient, error) {
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
	clientFactory, err := armsubscriptions.NewClientFactory(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("create client factory error: %v", err)
	}

	return &SubClient{
		subscriptionID: opt.Account.SubscriptionID,
		subClient:      clientFactory.NewClient(),
	}, nil
}

// ListLocations list locations
func (sub *SubClient) ListLocations(ctx context.Context) ([]*proto.RegionInfo, error) {
	pager := sub.subClient.NewListLocationsPager(sub.subscriptionID, nil)
	result := make([]*armsubscriptions.Location, 0)
	for pager.More() {
		next, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list locations")
		}
		result = append(result, next.Value...)
	}

	regions := make([]*proto.RegionInfo, 0)
	for _, v := range result {
		if v.Name != nil && v.DisplayName != nil {
			regions = append(regions, &proto.RegionInfo{
				Region:      *v.Name,
				RegionName:  *v.DisplayName,
				RegionState: BCSRegionStateAvailable,
			})
		}
	}
	return regions, nil
}

// ListAvailabilityZones list availability zones
func (sub *SubClient) ListAvailabilityZones(ctx context.Context, location string) ([]*proto.ZoneInfo, error) {
	pager := sub.subClient.NewListLocationsPager(sub.subscriptionID, nil)
	result := make([]*armsubscriptions.Location, 0)
	for pager.More() {
		next, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list zones")
		}
		result = append(result, next.Value...)
	}

	zones := make([]*proto.ZoneInfo, 0)
	for _, v := range result {
		if *v.Name == location {
			for _, z := range v.AvailabilityZoneMappings {
				zones = append(zones, &proto.ZoneInfo{
					Zone:     *z.LogicalZone,
					ZoneName: *z.PhysicalZone,
				})
			}
			break
		}
	}

	return zones, nil
}
