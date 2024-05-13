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

// Package business xxx
package business

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
)

// SubnetUsedIpCount 子网已使用IP数目统计
func SubnetUsedIpCount(ctx context.Context, opt *cloudprovider.ListNetworksOption, subnetID string) (uint32, error) {
	var usedIPCount uint32

	client, err := api.NewAksServiceImplWithCommonOption(&opt.CommonOption)
	if err != nil {
		return 0, fmt.Errorf("init AksService failed, %v", err)
	}

	interfaceList, err := client.ListNetworkNicAll(ctx)
	if err != nil {
		return 0, err
	}

	for _, nic := range interfaceList {
		if nic == nil || nic.Properties == nil {
			continue
		}

		for _, ipConfig := range nic.Properties.IPConfigurations {
			if ipConfig.Properties != nil && ipConfig.Properties.Subnet != nil &&
				*ipConfig.Properties.Subnet.ID == subnetID {
				usedIPCount++
			}
		}
	}

	return usedIPCount, nil
}
