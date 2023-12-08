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

package tasks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/resource/tresource"
)

// CreateNodePoolTask create nodePool
func CreateNodePoolTask(taskID, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateNodePoolTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateNodePoolTask[%s] run current step %s, system: %s, old state: %s, params %v",
		taskID, stepName, step.System, step.Status, step.Params)

	// extract valid parameter
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	poolProvider := step.Params[cloudprovider.PoolProvider.String()]
	resourcePoolID := step.Params[cloudprovider.PoolID.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CreateNodePoolTask[%s] GetClusterDependBasicInfo for NodeGroup %s to clean Node in task %s "+
			"step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = createNodeGroupAction(ctx, dependInfo, cloudprovider.ResourcePoolData{
		Provider:       poolProvider,
		ResourcePoolID: resourcePoolID,
	})
	if err != nil {
		blog.Errorf("CreateNodePoolTask[%s] createNodeGroupAction failed: %v", taskID, err.Error())
		retErr := fmt.Errorf("createNodeGroupAction failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateNodePoolTask[%s]: task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func createNodeGroupAction(ctx context.Context, data *cloudprovider.CloudDependBasicInfo,
	pool cloudprovider.ResourcePoolData) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	consumerID, err := createResourcePool(ctx, data, pool)
	if err != nil {
		blog.Errorf("createNodeGroupAction[%s] failed: %v", taskID, err)
		return err
	}

	err = cloudprovider.UpdateNodeGroupCloudAndModuleInfo(data.NodeGroup.NodeGroupID, consumerID,
		true, data.Cluster.BusinessID)
	if err != nil {
		blog.Errorf("createNodeGroupAction[%s] UpdateNodeGroupCloudAndModuleInfo failed: %v", taskID, err)
		return err
	}

	blog.Infof("createNodeGroupAction[%s] successful", taskID)
	return nil
}

func createResourcePool(ctx context.Context, data *cloudprovider.CloudDependBasicInfo,
	pool cloudprovider.ResourcePoolData) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	consumerID, err := tresource.GetResourceManagerClient().CreateResourcePool(ctx, resource.ResourcePoolInfo{
		Name:      data.NodeGroup.NodeGroupID,
		Provider:  pool.Provider,
		ClusterID: data.Cluster.ClusterID,
		RelativeDevicePool: func() []string {
			if pool.ResourcePoolID == "" {
				return nil
			}
			return strings.Split(pool.ResourcePoolID, ",")
		}(),
		PoolID:   []string{pool.ResourcePoolID},
		Operator: common.ClusterManager,
	})
	if err != nil {
		blog.Errorf("task[%s] createResourcePool failed: %v", taskID, err)
		return "", err
	}

	blog.Infof("task[%s] createResourcePool successful[%s]", taskID, consumerID)
	return consumerID, nil
}
