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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/auth/basic"
	ecs "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	region "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/region"
)

// EcsClient as client
type EcsClient struct {
	*ecs.EcsClient
}

// NewEcsClient new ecs client
func NewEcsClient(opt *cloudprovider.CommonOption) (*EcsClient, error) {
	if opt == nil || opt.Account == nil || len(opt.Account.SecretID) == 0 || len(opt.Account.SecretKey) == 0 {
		return nil, cloudprovider.ErrCloudCredentialLost
	}
	if len(opt.Region) == 0 {
		return nil, cloudprovider.ErrCloudRegionLost
	}

	auth := basic.NewCredentialsBuilder().WithAk(opt.Account.SecretID).WithSk(opt.Account.SecretKey).
		WithProjectId(opt.Account.HwCCEProjectID).Build()

	client := ecs.NewEcsClient(
		ecs.EcsClientBuilder().WithRegion(region.ValueOf(opt.Region)).WithCredential(auth).Build(),
	)

	return &EcsClient{
		EcsClient: client,
	}, nil
}

// ListEcsDetails batch get ecs server detail
func (e *EcsClient) ListEcsDetails(serverIds []string) ([]*model.ServerDetail, error) {
	servers := make([]*model.ServerDetail, 0)
	for _, v := range serverIds {
		rsp, err := e.ShowServer(&model.ShowServerRequest{
			ServerId: v,
		})
		if err != nil {
			return nil, err
		}

		servers = append(servers, rsp.Server)
	}

	return servers, nil
}
