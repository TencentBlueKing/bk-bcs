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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/cache"
	common "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
)

var (
	// ErrNotInited err server not init
	ErrNotInited = errors.New("server not init")

	// ClusterStatusRunning cluster status running
	ClusterStatusRunning = "RUNNING"
	// CacheKeyClusterPrefix cluster Prefix
	CacheKeyClusterPrefix = "CLUSTER_%s"
)

// GetCluster get cluster by clusterID
func GetCluster(ctx context.Context, clusterID string) (*clustermanager.Cluster, error) {
	// 1. if hit, get from cache
	c := cache.GetCache()
	if cluster, exists := c.Get(fmt.Sprintf(CacheKeyClusterPrefix, clusterID)); exists {
		return cluster.(*clustermanager.Cluster), nil
	}
	cli, closeCon, err := clustermanager.GetClient(common.ServiceDomain)
	if err != nil {
		logging.Error("get cluster manager client failed, err: %s", err.Error())
		return nil, err
	}
	defer closeCon()
	req := &clustermanager.GetClusterReq{
		ClusterID: clusterID,
	}
	resp, err := cli.GetCluster(ctx, req)
	if err != nil {
		logging.Error("get cluster from cluster manager failed, err: %s", err.Error())
		return nil, err
	}
	if resp.GetCode() != 0 {
		logging.Error("get cluster from cluster manager failed, msg: %s", resp.GetMessage())
		return nil, errors.New(resp.GetMessage())
	}
	_ = c.Add(fmt.Sprintf(CacheKeyClusterPrefix, clusterID), resp.GetData(), 5*time.Minute)
	return resp.GetData(), nil
}

// ListClusters list clusters by projectID
func ListClusters(ctx context.Context, projectID string) ([]*clustermanager.Cluster, error) {
	cli, closeCon, err := clustermanager.GetClient(common.ServiceDomain)
	if err != nil {
		logging.Error("get cluster manager client failed, err: %s", err.Error())
		return nil, err
	}
	defer closeCon()
	req := &clustermanager.ListClusterReq{
		ProjectID: projectID,
		Status:    ClusterStatusRunning,
	}
	resp, err := cli.ListCluster(ctx, req)
	if err != nil {
		logging.Error("list clusters from cluster manager failed, err: %s", err.Error())
		return nil, err
	}
	if resp.GetCode() != 0 {
		logging.Error("list clusters from cluster manager failed, msg: %s", resp.GetMessage())
		return nil, errors.New(resp.GetMessage())
	}
	return resp.GetData(), nil
}

// GetResourceUsage get project resource usage
func GetResourceUsage(ctx context.Context, projectID, provider string) (
	[]*clustermanager.ProjectAutoscalerQuota, error) {
	cli, closeCon, err := clustermanager.GetClient(common.ServiceDomain)
	if err != nil {
		logging.Error("get cluster manager client failed, err: %s", err.Error())
		return nil, err
	}
	defer closeCon()
	req := &clustermanager.GetProjectResourceQuotaUsageRequest{
		ProjectID:  projectID,
		ProviderID: provider,
	}
	resp, err := cli.GetProjectResourceQuotaUsage(ctx, req)
	if err != nil {
		logging.Error("get project resource usage from cluster manager failed, err: %s", err.Error())
		return nil, err
	}
	if resp.GetCode() != 0 {
		logging.Error("get project resource usage from cluster manager failed, msg: %s", resp.GetMessage())
		return nil, errors.New(resp.GetMessage())
	}

	data, err := resp.GetData().MarshalJSON()
	if err != nil {
		return nil, err
	}
	logging.Info("get project resource usage from cluster manager, data: %s", string(data))

	var pqs []*clustermanager.ProjectAutoscalerQuota

	if err = json.Unmarshal(data, &pqs); err != nil {
		logging.Error("unmarshal error: %s", err.Error())
		return nil, err
	}

	for _, pq := range pqs {
		pqRaw, _ := json.Marshal(pq)
		logging.Info("pq: %s", string(pqRaw))
	}

	return pqs, nil
}

// GetNodeGroup get node group
func GetNodeGroup(ctx context.Context, nodeGroupID string) (*clustermanager.NodeGroup, error) {
	cli, closeCon, err := clustermanager.GetClient(common.ServiceDomain)
	if err != nil {
		logging.Error("get cluster manager client failed, err: %s", err.Error())
		return nil, err
	}
	defer closeCon()
	req := &clustermanager.GetNodeGroupRequest{
		NodeGroupID: nodeGroupID,
	}

	resp, err := cli.GetNodeGroup(ctx, req)
	if err != nil {
		logging.Error("get project resource usage from cluster manager failed, err: %s", err.Error())
		return nil, err
	}

	if resp.GetCode() != 0 {
		logging.Error("get project resource usage from cluster manager failed, msg: %s", resp.GetMessage())
		return nil, errors.New(resp.GetMessage())
	}

	return resp.GetData(), nil
}
