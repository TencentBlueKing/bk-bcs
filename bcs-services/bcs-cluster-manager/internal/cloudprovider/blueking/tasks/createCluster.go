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

// Package tasks xxx
package tasks

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tenant"
)

// UpdateCreateClusterDBInfoTask update cluster DB info
func UpdateCreateClusterDBInfoTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateCreateClusterDBInfoTask[%s]: current step[%s] successful and skip",
			taskID, stepName)
		return nil
	}

	blog.Infof("UpdateCreateClusterDBInfoTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeIps := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeIPsKey.String(), ",")

	// update nodes status
	_ = cloudprovider.UpdateNodeListStatus(true, nodeIps, icommon.StatusRunning)
	cluster, err := cloudprovider.UpdateClusterStatus(clusterID, icommon.StatusInitialization)
	if err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s]: update cluster systemID for %s failed",
			taskID, clusterID)
	}

	// sync clusterData to pass-cc
	if cluster != nil {
		utils.SyncClusterInfoToPassCC(taskID, cluster)

		// sync cluster perms
		ctx, err := tenant.WithTenantIdByResourceForContext(context.Background(), tenant.ResourceMetaData{
			ClusterId: clusterID,
		})
		if err != nil {
			blog.Errorf("UpdateEKSNodesToDBTask WithTenantIdByResourceForContext failed: %v", err)
		}
		utils.AuthClusterResourceCreatorPerm(ctx, cluster.ClusterID, cluster.ClusterName, cluster.Creator)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}
