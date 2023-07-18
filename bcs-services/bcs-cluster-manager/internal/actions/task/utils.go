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

package task

import (
	"context"
	"errors"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// Passwd flag
var Passwd = []string{"password", "passwd"}

func strContains(ipList []string, ip string) bool {
	for i := range ipList {
		if strings.EqualFold(ipList[i], ip) {
			return true
		}
	}
	return false
}

func hiddenTaskPassword(task *proto.Task) {
	if task != nil && len(task.Steps) > 0 {
		for i := range task.Steps {
			for k := range task.Steps[i].Params {
				if utils.StringInSlice(k,
					[]string{cloudprovider.BkSopsTaskUrlKey.String(), cloudprovider.ShowSopsUrlKey.String()}) {
					continue
				}
				delete(task.Steps[i].Params, k)
			}
		}
	}

	if task != nil && len(task.CommonParams) > 0 {
		for k, v := range task.CommonParams {
			if utils.StringInSlice(strings.ToLower(k), Passwd) || utils.StringContainInSlice(v, Passwd) ||
				utils.StringInSlice(k, []string{cloudprovider.DynamicClusterKubeConfigKey.String()}) {
				delete(task.CommonParams, k)
			}
		}
	}
}

func updateNodeGroupStatus(model store.ClusterManagerModel, nodeGroupID, status string) error {
	group, err := model.GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("updateNodeGroupStatus[%s] GetNodeGroup failed: %v", nodeGroupID, err)
		return err
	}
	if group == nil {
		blog.Warnf("updateNodeGroupStatus[%s] not found nodegroup", nodeGroupID)
		return nil
	}

	group.Status = status
	err = model.UpdateNodeGroup(context.Background(), group)
	if err != nil {
		blog.Errorf("updateNodeGroupStatus[%s] UpdateNodeGroup failed: %v", nodeGroupID, err)
		return err
	}

	return nil
}

func updateClusterStatus(model store.ClusterManagerModel, clusterID, status string) error {
	cluster, err := model.GetCluster(context.Background(), clusterID)
	if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
		blog.Errorf("updateClusterStatus[%s] GetCluster failed: %v", clusterID, err)
		return err
	}
	if cluster == nil {
		blog.Warnf("updateClusterStatus[%s] not found cluster", clusterID)
		return nil
	}

	cluster.Status = status
	err = model.UpdateCluster(context.Background(), cluster)
	if err != nil {
		blog.Errorf("updateClusterStatus[%s] UpdateCluster failed: %v", clusterID, err)
		return err
	}

	return nil
}

func updateNodeStatus(model store.ClusterManagerModel, nodeList []string, status string) error {
	for i := range nodeList {
		node, err := model.GetNodeByIP(context.Background(), nodeList[i])
		if err != nil && !errors.Is(err, drivers.ErrTableRecordNotFound) {
			blog.Errorf("updateNodeStatus[%s] GetNodeByIP failed: %v", nodeList[i], err)
			return err
		}
		if node == nil {
			blog.Warnf("updateNodeStatus[%s] not found node", nodeList[i])
			continue
		}

		node.Status = status
		err = model.UpdateNode(context.Background(), node)
		if err != nil {
			blog.Errorf("updateNodeStatus[%s] UpdateNode failed: %v", nodeList[i], err)
			return err
		}
	}

	return nil
}
