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
	"strconv"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"

	"golang.org/x/oauth2"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// ComputeServiceClient compute service client
type ComputeServiceClient struct {
	gkeProjectID         string
	computeServiceClient *compute.Service
}

// NewComputeServiceClient create a client for google compute service
func NewComputeServiceClient(opt *cloudprovider.CommonOption) (*ComputeServiceClient, error) {
	if opt == nil || opt.Account == nil {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Account.ServiceAccountSecret) == 0 || opt.Account.GkeProjectID == "" {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	computeServiceClient, err := getComputeServiceClient(context.Background(), opt.Account.ServiceAccountSecret)
	if err != nil {
		return nil, cloudprovider.ErrCloudInitFailed
	}
	return &ComputeServiceClient{
		gkeProjectID:         opt.Account.GkeProjectID,
		computeServiceClient: computeServiceClient,
	}, nil
}

// ListRegions list regions
func (c *ComputeServiceClient) ListRegions(ctx context.Context) ([]*proto.RegionInfo, error) {
	if c.gkeProjectID == "" {
		return nil, fmt.Errorf("ListRegions failed: gkeProjectId is required")
	}

	regions, err := c.computeServiceClient.Regions.List(c.gkeProjectID).Context(ctx).Do()
	if err != nil {
		return nil, err
	}

	result := make([]*proto.RegionInfo, 0)
	for _, v := range regions.Items {
		if v.Name != "" && v.Description != "" {
			result = append(result, &proto.RegionInfo{
				Region:      v.Name,
				RegionName:  v.Description,
				RegionState: v.Status,
			})
		}
	}
	return result, nil
}

// ListZones list zones
func (c *ComputeServiceClient) ListZones(ctx context.Context) ([]*proto.ZoneInfo, error) {
	if c.gkeProjectID == "" {
		return nil, fmt.Errorf("ListZones failed: gkeProjectId is required")
	}

	zones, err := c.computeServiceClient.Zones.List(c.gkeProjectID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("ListZones failed: %v", err)
	}

	result := make([]*proto.ZoneInfo, 0)
	for _, v := range zones.Items {
		if v.Name != "" && v.Description != "" {
			result = append(result, &proto.ZoneInfo{
				ZoneID:    strconv.FormatUint(v.Id, 10),
				Zone:      v.Name,
				ZoneName:  v.Description,
				ZoneState: v.Status,
			})
		}
	}
	return result, nil
}

func getComputeServiceClient(ctx context.Context, credentialContent string) (*compute.Service, error) {
	ts, err := GetTokenSource(ctx, credentialContent)
	if err != nil {
		return nil, fmt.Errorf("getComputeServiceClient failed: %v", err)
	}

	service, err := compute.NewService(ctx, option.WithHTTPClient(oauth2.NewClient(ctx, ts)))
	if err != nil {
		return nil, fmt.Errorf("getComputeServiceClient failed: %v", err)
	}

	return service, nil
}
