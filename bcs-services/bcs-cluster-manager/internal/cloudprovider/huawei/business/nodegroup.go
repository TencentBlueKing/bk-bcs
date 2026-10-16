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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
)

// GetNodeGroupMaxPod get max pod number for node group
func GetNodeGroupMaxPod(common cloudprovider.CommonOption, clusterId string) (int32, error) {
	client, err := api.NewCceClient(&common)
	if err != nil {
		return 0, err
	}

	nodeGroups, err := client.ListClusterNodeGroups(clusterId)
	if err != nil {
		return 0, err
	}

	if len(nodeGroups) == 0 {
		return 0, nil
	}

	nodeTemplate := nodeGroups[0].Spec.NodeTemplate
	if nodeTemplate.ExtendParam == nil || nodeTemplate.ExtendParam.MaxPods == nil {
		// 获取机型列表
		ecsClient, err := api.NewEcsClient(&common)
		if err != nil {
			return 0, err
		}

		az := ""
		if nodeTemplate.Az != "random" {
			az = nodeTemplate.Az
		}

		flavors, err := ecsClient.GetAllFlavors(az)
		if err != nil {
			return 0, err
		}

		for _, flavor := range *flavors {
			if flavor.Name == nodeTemplate.Flavor {
				ram := flavor.Ram / 1024
				switch {
				case ram < 8:
					return 20, nil
				case ram < 16:
					return 40, nil
				case ram < 32:
					return 60, nil
				case ram < 64:
					return 80, nil
				default:
					return 110, nil
				}
			}
		}
	}

	return *nodeGroups[0].Spec.NodeTemplate.ExtendParam.MaxPods, nil
}
