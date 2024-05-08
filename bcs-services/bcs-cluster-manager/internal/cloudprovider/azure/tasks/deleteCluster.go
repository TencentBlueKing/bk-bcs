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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/azure/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// DeleteAKSClusterTask delete cluster task
func DeleteAKSClusterTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("DeleteAKSClusterTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("DeleteAKSClusterTask GetClusterDependBasicInfo failed")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get azure client
	cli, err := api.NewAksServiceImplWithCommonOption(dependInfo.CmOption)
	if err != nil {
		blog.Errorf("DeleteAKSClusterTask[%s]: get azure client for cluster[%s] in BasicInfo %s step %s failed, %s",
			taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get azure client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	resourceGroup, ok := dependInfo.Cluster.ExtraInfo[icommon.ClusterResourceGroup]
	if !ok {
		blog.Errorf("DeleteAKSClusterTask[%s] get cluster[%s] resourceGroup failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster[%s] resourceGroup failed", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if dependInfo.Cluster.SystemID != "" {
		err = cli.DeleteClusterWithName(context.Background(), resourceGroup, dependInfo.Cluster.SystemID)
		if err != nil {
			blog.Errorf("DeleteAKSClusterTask[%s]: call azure DeleteAKSCluster failed: %v",
				taskID, err)
			retErr := fmt.Errorf("call azure DeleteAKSCluster failed: %s", err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		_ = cloudprovider.UpdateClusterSystemID(clusterID, "")
		blog.Infof("DeleteAKSClusterTask[%s]: DeleteAKSCluster[%s] successful",
			taskID, dependInfo.Cluster.SystemID)
	} else {
		blog.Infof("DeleteAKSClusterTask[%s]: DeleteAKSCluster skip current step because SystemID empty", taskID)
	}

	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("DeleteAKSClusterTask[%s]: %s update to storage fatal", taskID, stepName)
		return err
	}
	return nil
}

// CleanClusterDBInfoTask clean cluster DB info
func CleanClusterDBInfoTask(taskID string, stepName string) error {
	// delete node && nodeGroup && cluster
	// get relative nodes by clusterID
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cluster, err := cloudprovider.GetStorageModel().GetCluster(context.Background(), clusterID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: get cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("get cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// delete nodes
	err = cloudprovider.GetStorageModel().DeleteNodesByClusterID(context.Background(), cluster.ClusterID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete nodes for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete node for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete nodes for cluster[%s] in DB successful", taskID, clusterID)

	// delete nodeGroup
	err = cloudprovider.GetStorageModel().DeleteNodeGroupByClusterID(context.Background(), cluster.ClusterID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete nodeGroups for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete nodeGroups for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete nodeGroups for cluster[%s] in DB successful", taskID, clusterID)

	// delete cluster
	cluster.Status = icommon.StatusDeleting
	err = cloudprovider.GetStorageModel().UpdateCluster(context.Background(), cluster)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete cluster for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete cluster[%s] in DB successful", taskID, clusterID)

	utils.SyncDeletePassCCCluster(taskID, cluster)
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: %s update to storage fatal", taskID, stepName)
		return err
	}
	return nil
}
