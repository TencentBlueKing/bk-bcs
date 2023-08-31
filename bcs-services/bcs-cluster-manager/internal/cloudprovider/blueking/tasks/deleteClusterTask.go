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

package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// CleanClusterDBInfoTask clean cluster DB info
func CleanClusterDBInfoTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateCreateClusterDBInfoTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]

	// update cluster status deleting
	cluster, err := cloudprovider.UpdateClusterStatus(clusterID, common.StatusDeleting)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete cluster for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete cluster for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete cluster[%s] in DB successful", taskID, clusterID)

	// delete cluster nodes
	err = cloudprovider.GetStorageModel().DeleteNodesByClusterID(context.Background(), cluster.ClusterID)
	if err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: delete nodes for %s failed", taskID, clusterID)
		retErr := fmt.Errorf("delete node for %s failed, %s", clusterID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CleanClusterDBInfoTask[%s]: delete nodes for cluster[%s] in DB successful", taskID, clusterID)

	// sync clean cluster dependency(pass-cc/token/credential)
	utils.SyncDeletePassCCCluster(taskID, cluster)
	utils.DeleteBcsAgentToken(cluster.ClusterID)
	utils.DeleteClusterCredentialInfo(cluster.ClusterID)

	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CleanClusterDBInfoTask[%s]: task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}
