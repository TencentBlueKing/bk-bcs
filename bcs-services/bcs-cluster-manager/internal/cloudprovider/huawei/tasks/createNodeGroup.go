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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
)

// CreateCloudNodeGroupTask create cloud node group task
func CreateCloudNodeGroupTask(taskID string, stepName string) error {
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
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: getClusterDependBasicInfo failed: %v", taskID, err)
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption := dependInfo.CmOption
	cluster := dependInfo.Cluster
	group := dependInfo.NodeGroup

	cceCli, err := api.NewCceClient(cmOption)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: get cce client for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud cce client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	// cce nodePool名称以小写字母开头，由小写字母、数字、中划线(-)组成，长度范围1-50位，且不能以中划线(-)结尾
	group.NodeGroupID = strings.ToLower(group.NodeGroupID)
	req, err := api.GenerateCreateNodePoolRequest(group, cluster)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: generate create nodepool request[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("generate create nodepool request err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	rsp, err := cceCli.CreateClusterNodePool(req)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool[%s] api in task %s "+
			"step %s failed, %s, rsp: %+v", taskID, nodeGroupID, taskID, stepName, err.Error(), rsp)
		retErr := fmt.Errorf("call CreateClusterNodePool[%s] api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	blog.Infof("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool successful", taskID)

	// 保存cce节点池id
	group.CloudNodeGroupID = *rsp.Metadata.Uid
	// update nodegorup cloudNodeGroupID
	err = updateNodeGroupCloudNodeGroupID(nodeGroupID, group)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: updateNodeGroupCloudNodeGroupID[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call CreateCloudNodeGroupTask updateNodeGroupCloudNodeGroupID[%s] api err, %s",
			nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool updateNodeGroupCloudNodeGroupID successful",
		taskID)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	state.Task.CommonParams["CloudNodeGroupID"] = group.CloudNodeGroupID
	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// CheckCloudNodeGroupStatusTask check cloud node group status task
func CheckCloudNodeGroupStatusTask(taskID string, stepName string) error { // nolint
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
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: getClusterDependBasicInfo failed: %v", taskID, err)
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption := dependInfo.CmOption
	cluster := dependInfo.Cluster
	group := dependInfo.NodeGroup

	cceCli, err := api.NewCceClient(cmOption)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: get cce client for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud cce client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	// wait node group state to normal
	ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Minute)
	defer cancel()

	err = loop.LoopDoFunc(ctx, func() error {
		nodePool, errLocal := cceCli.GetClusterNodePool(cluster.SystemID, group.CloudNodeGroupID)
		if errLocal != nil {
			blog.Errorf("taskID[%s] GetClusterNodePool[%s/%s] failed: %v", taskID, cluster.SystemID,
				group.CloudNodeGroupID, errLocal)
			return nil
		}

		switch {
		case nodePool.Status.Phase.Value() != "":
			blog.Infof("taskID[%s] GetClusterNodePool[%s] still creating, status[%s]",
				taskID, group.CloudNodeGroupID, nodePool.Status.Phase.Value())
			return nil
		case nodePool.Status.Phase.Value() == "":
			return loop.EndLoop
		default:
			return nil
		}

	}, loop.LoopInterval(5*time.Second))
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: GetClusterNodePool failed: %v", taskID, err)
		retErr := fmt.Errorf("GetClusterNodePool failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	blog.Infof("CheckCloudNodeGroupStatusTask[%s]: call GetClusterNodePool successful", taskID)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}
