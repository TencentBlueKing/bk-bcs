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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/avast/retry-go"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/cce/v3/model"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// DeleteClusterInstance delete cluster instance
func DeleteClusterInstance(client *api.CceClient, clusterID string, nodes []model.Node) ([]string, error) {
	success := make([]string, 0)
	for _, node := range nodes {
		_, err := client.DeleteNode(clusterID, *node.Metadata.Uid, true)
		if err != nil {
			continue
		}

		success = append(success, *node.Metadata.Uid)
	}

	return success, nil
}

func deleteNode(client *api.CceClient, clusterID string, nodes model.Node) {} // nolint

// FilterClusterInstanceFromNodesIDs nodeIDs existInCluster or notExistInCluster
func FilterClusterInstanceFromNodesIDs(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	nodeIDs []string) ([]string, []string, error) {
	var (
		nodes             []model.Node
		existInCluster    = make([]string, 0)
		notExistInCluster = make([]string, 0)
	)

	client, err := api.NewCceClient(info.CmOption)
	if err != nil {
		return nil, nil, err
	}

	err = retry.Do(func() error {
		nodes, err = client.ListClusterNodes(info.Cluster.SystemID)
		if err != nil {
			return err
		}
		return nil
	}, retry.Attempts(3))

	for _, id := range nodeIDs {
		exit := false
		for _, node := range nodes {
			if *node.Metadata.Uid == id {
				exit = true
				existInCluster = append(existInCluster, id)
			}
		}
		if !exit {
			notExistInCluster = append(notExistInCluster, id)
		}
	}

	return existInCluster, notExistInCluster, nil
}

// CheckClusterDeletedNodes check if nodeIds are deleted in cluster
func CheckClusterDeletedNodes(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeIds []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// get qcloud client
	cli, err := api.NewCceClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterInstanceStatus[%s] failed, %s", taskID, err)
		return err
	}

	// wait node group state to normal
	timeCtx, cancel := context.WithTimeout(context.TODO(), 5*time.Minute)
	defer cancel()

	// wait all nodes to be ready
	err = loop.LoopDoFunc(timeCtx, func() error {
		instances, errQuery := cli.ListClusterNodes(info.Cluster.GetSystemID())
		if errQuery != nil {
			blog.Errorf("CheckClusterDeletedNodes[%s] QueryTkeClusterAllInstances failed: %v", taskID, errQuery)
			return nil
		}

		if len(instances) == 0 {
			return loop.EndLoop
		}

		clusterNodeIds := make([]string, 0)
		for i := range instances {
			clusterNodeIds = append(clusterNodeIds, *instances[i].Metadata.Uid)
		}

		for i := range nodeIds {
			if utils.StringInSlice(nodeIds[i], clusterNodeIds) {
				blog.Infof("CheckClusterDeletedNodes[%s] %s in cluster[%v]", taskID, nodeIds[i], clusterNodeIds)
				return nil
			}
		}

		return loop.EndLoop
	}, loop.LoopInterval(20*time.Second))
	// other error
	if err != nil {
		blog.Errorf("CheckClusterDeletedNodes[%s] failed: %v", taskID, err)
		return err
	}

	blog.Infof("CheckClusterDeletedNodes[%s] deleted nodes success[%v]", taskID, nodeIds)
	return nil
}

// DeleteTkeClusterByClusterId delete cluster by clusterId
func DeleteTkeClusterByClusterId(ctx context.Context, opt *cloudprovider.CommonOption,
	clsId string, deleteMode string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(clsId) == 0 {
		blog.Warnf("DeleteTkeClusterByClusterId[%s] clusterID empty", taskID)
		return nil
	}

	cli, err := api.NewCceClient(opt)
	if err != nil {
		blog.Errorf("DeleteTkeClusterByClusterId[%s] init tkeClient failed: %v", taskID, err)
		return err
	}

	err = cli.DeleteCceCluster(clsId)
	if err != nil && !strings.Contains(err.Error(), "Resource not found") {
		blog.Errorf("DeleteTkeClusterByClusterId[%s] deleteCluster failed: %v", taskID, err)
		return err
	}

	blog.Infof("DeleteTkeClusterByClusterId[%s] deleteCluster[%s] success", taskID, clsId)

	return nil
}
