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

	"github.com/pkg/errors"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
)

/*
	节点池
*/

// CreatePool 创建节点池.
func (aks *AksServiceImpl) CreatePool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (*proto.NodeGroup,
	error) {
	pool := new(armcontainerservice.AgentPool)
	if err := aks.NodeGroupToAgentPool(info.NodeGroup, pool); err != nil {
		return nil, errors.Wrapf(err, "bcs nodeGroup to azure agentPool failed")
	}
	return aks.CreatePoolWithName(ctx, pool, info.Cluster.SystemID, info.NodeGroup.CloudNodeGroupID, info.NodeGroup)
}

// CreatePoolWithName 从名称创建节点池.
// pool - 代理节点池.
// resourceName - K8S名称(Cluster.SystemID).
func (aks *AksServiceImpl) CreatePoolWithName(ctx context.Context, pool *armcontainerservice.AgentPool,
	resourceName, poolName string, group *proto.NodeGroup) (*proto.NodeGroup, error) {
	pool, err := aks.CreatePoolAndReturn(ctx, pool, resourceName, poolName)
	if err != nil {
		return nil, errors.Wrapf(err, "call CreatePoolAndReturn failed")
	}
	if group == nil {
		group = new(proto.NodeGroup)
	}
	if err = aks.AgentPoolToNodeGroup(pool, group); err != nil {
		return group, errors.Wrapf(err, "call AgentPoolToNodeGroup failed")
	}
	return group, nil
}

// CreatePoolAndReturn 从名称创建节点池.
// pool - 代理节点池.
// resourceName - K8S名称(Cluster.SystemID).
// poolName - 节点池名称(NodeGroup.CloudNodeGroupID).
func (aks *AksServiceImpl) CreatePoolAndReturn(ctx context.Context, pool *armcontainerservice.AgentPool,
	resourceName, poolName string) (*armcontainerservice.AgentPool, error) {
	// 创建节点池
	poller, err := aks.poolClient.BeginCreateOrUpdate(ctx, aks.resourcesGroup, resourceName, poolName, *pool,
		nil)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to finish the request,resourcesGroupName:%s,resourceName:%s,agentPoolName:%s",
			aks.resourcesGroup, resourceName, poolName)
	}
	// 每5秒钟轮询一次
	resp, err := poller.PollUntilDone(ctx, pollFrequency5)
	if err != nil {
		return nil, errors.Wrapf(err,
			"failed to pull the result,resourcesGroupName:%s,resourceName:%s,agentPoolName:%s",
			aks.resourcesGroup, resourceName, poolName)
	}
	return &resp.AgentPool, nil
}

// DeletePool 删除节点池.
func (aks *AksServiceImpl) DeletePool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) error {
	return aks.DeletePoolWithName(ctx, info.Cluster.SystemID, info.NodeGroup.CloudNodeGroupID)
}

// DeletePoolWithName 从名称删除节点池.
// resourceName - K8S名称(Cluster.SystemID).
// poolName - 节点池名称(NodeGroup.CloudNodeGroupID).
func (aks *AksServiceImpl) DeletePoolWithName(ctx context.Context, resourceName, poolName string) error {
	poller, err := aks.poolClient.BeginDelete(ctx, aks.resourcesGroup, resourceName, poolName, nil)
	if err != nil {
		return errors.Wrapf(err, "failed to finish the request,resourcesGroupName:%s,resourceName:%s,poolName:%s",
			aks.resourcesGroup, resourceName, poolName)
	}
	if _, err = poller.PollUntilDone(ctx, pollFrequency5); err != nil {
		return errors.Wrapf(err, "failed to pull the result,resourcesGroupName:%s,resourceName:%s,poolName:%s",
			aks.resourcesGroup, resourceName, poolName)
	}
	return nil
}

// UpdatePool 修改节点池(覆盖).
func (aks *AksServiceImpl) UpdatePool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (*proto.NodeGroup,
	error) {
	cluster := info.Cluster
	group := info.NodeGroup
	// 拉取云上节点池
	pool, err := aks.GetPoolAndReturn(ctx, cluster.SystemID, group.CloudNodeGroupID)
	if err != nil {
		return nil, errors.Wrapf(err, "call GetPoolAndReturn failed")
	}
	// 覆盖
	if err = aks.NodeGroupToAgentPool(group, pool); err != nil {
		return nil, errors.Wrapf(err, "bcs nodeGroup to azure agentPool failed")
	}
	// 修改
	if _, err = aks.UpdatePoolAndReturn(ctx, pool, cluster.SystemID, group.CloudNodeGroupID); err != nil {
		return nil, errors.Wrapf(err, "call UpdatePoolAndReturn failed")
	}
	return group, nil
}

// UpdatePoolAndReturn 从名称修改节点池.
// pool - 代理节点池.
// resourceName - K8S名称(Cluster.SystemID).
// poolName - 节点池名称(NodeGroup.CloudNodeGroupID).
func (aks *AksServiceImpl) UpdatePoolAndReturn(ctx context.Context, pool *armcontainerservice.AgentPool,
	resourceName, poolName string) (*armcontainerservice.AgentPool, error) {
	poller, err := aks.poolClient.BeginCreateOrUpdate(ctx, aks.resourcesGroup, resourceName, poolName, *pool, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request,resourcesGroupName:%s,resourceName:%s,poolName:%s",
			aks.resourcesGroup, resourceName, poolName)
	}
	resp, err := poller.PollUntilDone(ctx, pollFrequency2)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to pull the result,resourcesGroupName:%s,resourceName:%s,poolName:%s",
			aks.resourcesGroup, resourceName, poolName)
	}
	return &resp.AgentPool, nil
}

// GetPool 获取节点池.
func (aks *AksServiceImpl) GetPool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (*proto.NodeGroup, error) {
	return aks.GetPoolWithName(ctx, info.Cluster.SystemID, info.NodeGroup.CloudNodeGroupID, info.NodeGroup)
}

// GetPoolWithName 从名称获取节点池.
// resourceName - K8S名称(Cluster.SystemID).
// poolName - 节点池名称(NodeGroup.CloudNodeGroupID).
func (aks *AksServiceImpl) GetPoolWithName(ctx context.Context, resourceName, poolName string, group *proto.NodeGroup) (
	*proto.NodeGroup, error) {
	pool, err := aks.GetPoolAndReturn(ctx, resourceName, poolName)
	if err != nil {
		return nil, errors.Wrapf(err, "call GetPoolAndReturn failed")
	}
	if group == nil {
		group = new(proto.NodeGroup)
	}
	if err = aks.AgentPoolToNodeGroup(pool, group); err != nil {
		return group, errors.Wrapf(err, "call AgentPoolToNodeGroup falied")
	}
	return group, nil
}

// GetPoolAndReturn 从名称获取节点池.
// resourceName - K8S名称(Cluster.SystemID).
// poolName - 节点池名称(NodeGroup.CloudNodeGroupID).
func (aks *AksServiceImpl) GetPoolAndReturn(ctx context.Context, resourceName, poolName string) (
	*armcontainerservice.AgentPool, error) {
	resp, err := aks.poolClient.Get(ctx, aks.resourcesGroup, resourceName, poolName, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to finish the request,resourcesGroupName:%s,resourceName:%s,poolName:%s",
			aks.resourcesGroup, resourceName, poolName)
	}
	return &resp.AgentPool, nil
}

// ListPool 获取节点池列表.
func (aks *AksServiceImpl) ListPool(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) ([]*proto.NodeGroup,
	error) {
	return aks.ListPoolWithName(ctx, info.Cluster.SystemID)
}

// ListPoolWithName 从名称获取节点池列表.
// resourceName - K8S名称(Cluster.SystemID).
func (aks *AksServiceImpl) ListPoolWithName(ctx context.Context, resourceName string) ([]*proto.NodeGroup, error) {
	pools, err := aks.ListPoolAndReturn(ctx, resourceName)
	if err != nil {
		return nil, errors.Wrapf(err, "call ListPoolAndReturn failed")
	}
	resp := make([]*proto.NodeGroup, len(pools))
	for i, pool := range pools {
		resp[i] = new(proto.NodeGroup)
		if err = aks.AgentPoolToNodeGroup(pool, resp[i]); err != nil {
			return nil, errors.Wrapf(err, "bcs nodeGroup to azure agentPool failed")
		}
	}
	return resp, nil
}

// ListPoolAndReturn 从名称获取节点池列表.
// resourceName - K8S名称(Cluster.SystemID).
func (aks *AksServiceImpl) ListPoolAndReturn(ctx context.Context, resourceName string) (
	[]*armcontainerservice.AgentPool, error) {
	resp := make([]*armcontainerservice.AgentPool, 0)
	pager := aks.poolClient.NewListPager(aks.resourcesGroup, resourceName, nil)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to advance page,resourcesGroupName:%s,resourceName:%s",
				aks.resourcesGroup, resourceName)
		}
		resp = append(resp, nextResult.Value...)
	}
	return resp, nil
}
