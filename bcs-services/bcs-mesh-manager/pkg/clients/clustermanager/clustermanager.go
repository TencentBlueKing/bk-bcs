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

// Package clustermanager 获取clustermanager client
package clustermanager

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
)

// GetClient 获取clustermanager client
func GetClient() (clustermanager.ClusterManagerClient, func(), error) {
	clustermanagerClient, closeFunc, err := clustermanager.GetClient(common.ServiceDomain)
	if err != nil {
		return nil, nil, err
	}
	return clustermanagerClient, closeFunc, nil
}

// GetClusterInfo 获取集群信息
func GetClusterInfo(ctx context.Context, projectID, clusterID string) (*clustermanager.Cluster, error) {
	clustermanagerClient, closeFunc, err := GetClient()
	if err != nil {
		return nil, err
	}
	defer closeFunc()

	clusterInfo, err := clustermanagerClient.GetCluster(ctx, &clustermanager.GetClusterReq{
		ClusterID: clusterID,
		ProjectId: projectID,
	})
	if err != nil {
		return nil, err
	}
	return clusterInfo.Data, nil
}
