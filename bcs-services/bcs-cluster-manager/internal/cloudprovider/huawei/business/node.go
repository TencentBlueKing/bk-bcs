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
	"strings"

	modelv2 "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/clusterops"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
)

// GetCloudZones get cloud zones
func GetCloudZones(common cloudprovider.CommonOption) ([]modelv2.NovaAvailabilityZone, error) {
	client, err := api.NewEcsClient(&common)
	if err != nil {
		return nil, err
	}

	return client.ListAvailabilityZones()
}

// GetRuntimeInfo get runtime info
func GetRuntimeInfo(clusterID string) (map[string][]string, error) {
	k8sOperator := clusterops.NewK8SOperator(options.GetGlobalCMOptions(), cloudprovider.GetStorageModel())
	nodes, err := k8sOperator.ListClusterNodes(context.Background(), clusterID)
	if err != nil {
		return nil, err
	}

	runtimeInfo := make(map[string][]string)
	for _, node := range nodes {
		runtime := strings.Split(node.Status.NodeInfo.ContainerRuntimeVersion, "://")
		if len(runtime) > 1 {
			runtimeVersion := strings.Split(runtime[1], "-")
			if len(runtimeVersion) > 1 {
				runtimeInfo[runtime[0]] = append(runtimeInfo[runtime[0]], runtimeVersion[0])
			}

		}
	}

	return runtimeInfo, err
}
