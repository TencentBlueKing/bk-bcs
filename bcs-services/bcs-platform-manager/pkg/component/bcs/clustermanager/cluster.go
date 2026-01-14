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

// Package clustermanager xxx
package clustermanager

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/storage"
)

// GetCluster get cluster from cluster manager
func GetCluster(ctx context.Context, clusterID, ProjectID string) (*clustermanager.Cluster, error) {
	if cacheResult, ok := storage.LocalCache.Slot.Get(getClusterCacheKey(clusterID)); ok {
		return cacheResult.(*clustermanager.Cluster), nil
	}

	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	p, err := cli.GetCluster(ctx, &clustermanager.GetClusterReq{
		ClusterID: clusterID,
		ProjectId: ProjectID,
	})
	if err != nil {
		return nil, fmt.Errorf("GetCluster error: %s", err)
	}

	if p.Code != 0 {
		return nil, fmt.Errorf("GetCluster error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	storage.LocalCache.Slot.Set(getClusterCacheKey(clusterID), p.Data, time.Hour*8)

	return p.Data, nil
}

// ListCluster list cluster from cluster manager
/*func ListCluster(ctx context.Context, req *clustermanager.ListClusterV2Req) (
	*clustermanager.ClusterBasicInfoData, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return nil, err
	}

	defer Close(close)

	p, err := cli.ListClusterV2(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ListCluster error: %s", err)
	}

	if p.Code != 0 {
		return nil, fmt.Errorf("ListCluster error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	return p.Data, nil
}*/

// UpdateCluster update cluster from cluster manager
func UpdateCluster(ctx context.Context, req *clustermanager.UpdateClusterReq) (bool, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.UpdateCluster(ctx, req)
	if err != nil {
		return false, fmt.Errorf("UpdateCluster error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("UpdateCluster error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	storage.LocalCache.Slot.Delete(getClusterCacheKey(req.ClusterID))

	return p.Result, nil
}

// AddClusterCidr add cidr to cluster
/*func AddClusterCidr(ctx context.Context, req *clustermanager.AddClusterCidrReq) (bool, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.AddClusterCidr(ctx, req)
	if err != nil {
		return false, fmt.Errorf("AddClusterCidr error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("AddClusterCidr error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	storage.LocalCache.Slot.Delete(getClusterCacheKey(req.ClusterID))

	return p.Result, nil
}*/

// AddSubnetToCluster add subnet to cluster
func AddSubnetToCluster(ctx context.Context, req *clustermanager.AddSubnetToClusterReq) (bool, error) {
	cli, close, err := clustermanager.GetClient(config.ServiceDomain)
	if err != nil {
		return false, err
	}

	defer Close(close)

	p, err := cli.AddSubnetToCluster(ctx, req)
	if err != nil {
		return false, fmt.Errorf("AddSubnetToCluster error: %s", err)
	}

	if p.Code != 0 {
		return false, fmt.Errorf("AddSubnetToCluster error, code: %d, message: %s", p.Code, p.GetMessage())
	}

	storage.LocalCache.Slot.Delete(getClusterCacheKey(req.ClusterID))

	return p.Result, nil
}

func getClusterCacheKey(clusterID string) string {
	return fmt.Sprintf("bcs.Cluster.%s", clusterID)
}
