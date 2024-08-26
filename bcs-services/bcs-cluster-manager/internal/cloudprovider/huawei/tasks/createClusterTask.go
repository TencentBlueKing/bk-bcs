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
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/business"
	providerutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
)

// CreateCCEClusterTask call huawei interface to create cluster
func CreateCCEClusterTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateClusterTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	nodeTemplateID := step.Params[cloudprovider.NodeTemplateIDKey.String()]
	operator := state.Task.CommonParams[cloudprovider.OperatorKey.String()]

	// get dependent basic info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:      clusterID,
		CloudID:        cloudID,
		NodeTemplateID: nodeTemplateID,
	})
	if err != nil {
		blog.Errorf("CreateClusterTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err) // nolint
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	req, err := api.GenerateCreateClusterRequest(ctx, dependInfo.Cluster, operator)
	if err != nil {
		blog.Errorf("createCluster[%s] generateCreateClusterRequest failed: %v", taskID, err)
		retErr := fmt.Errorf("generate create cluster request failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	// create cluster task
	clsId, jobId, err := createCluster(ctx, dependInfo, req, dependInfo.Cluster.SystemID)
	if err != nil {
		blog.Errorf("CreateClusterTask[%s] createCluster for cluster[%s] failed, %s",
			taskID, clusterID, err)
		retErr := fmt.Errorf("createCluster err, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)

		_ = cloudprovider.UpdateClusterErrMessage(clusterID, fmt.Sprintf("submit createCluster[%s] failed: %v",
			dependInfo.Cluster.GetClusterID(), err))
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	state.Task.CommonParams[cloudprovider.CloudSystemID.String()] = clsId
	state.Task.CommonParams[cloudprovider.CloudJobID.String()] = jobId

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func createCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	request *api.CreateClusterRequest, clsId string) (string, string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	client, err := api.NewCceClient(info.CmOption)
	if err != nil {
		return "", "", err
	}

	jobId := ""
	if clsId != "" {
		cluster, errGet := client.GetCceCluster(clsId)
		if errGet != nil {
			blog.Errorf("createCluster[%s] GetCluster[%s] failed, %s",
				taskID, info.Cluster.ClusterID, errGet)
			retErr := fmt.Errorf("call GetCluster[%s] api err, %s", info.Cluster.ClusterID, errGet)
			return "", "", retErr
		}
		// update cluster systemID
		info.Cluster.SystemID = *cluster.Metadata.Uid
	} else {
		rsp, err := client.CreateCluster(request)
		if err != nil {
			return "", "", err
		}

		if rsp.Metadata.Uid != nil {
			info.Cluster.SystemID = *rsp.Metadata.Uid

			err = cloudprovider.GetStorageModel().UpdateCluster(ctx, info.Cluster)
			if err != nil {
				blog.Errorf("createCluster[%s] updateClusterSystemID[%s] failed %s",
					taskID, info.Cluster.ClusterID, err)
				retErr := fmt.Errorf("call CreateCluster updateClusterSystemID[%s] api err: %s",
					info.Cluster.ClusterID, err)
				return "", "", retErr
			}

			blog.Infof("createCluster[%s] call CreateCluster updateClusterSystemID successful", taskID)
		} else if request.Spec.Charge.ChargeType == icommon.PREPAID && rsp.Status != nil && rsp.Status.JobID != nil {
			jobId = *rsp.Status.JobID
		}
	}

	return info.Cluster.SystemID, jobId, nil
}

// CheckCCEClusterStatusTask check cluster status
func CheckCCEClusterStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckClusterStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckClusterStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	systemID := state.Task.CommonParams[cloudprovider.CloudSystemID.String()]
	jobID := state.Task.CommonParams[cloudprovider.CloudJobID.String()]

	if systemID == "" && jobID == "" {
		blog.Errorf("CheckClusterStatusTask[%s]: cloud clusterID and cloud jobID is empty for cluster %s in task %s "+
			"step %s failed", taskID, clusterID, taskID, stepName)
		retErr := fmt.Errorf("cloud clusterID and cloud jobID is empty")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckClusterStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = checkClusterStatus(ctx, dependInfo, systemID, jobID)
	if err != nil {
		blog.Errorf("CheckClusterStatusTask[%s] checkClusterStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("checkClusterStatus[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)

		_ = cloudprovider.UpdateClusterErrMessage(clusterID, fmt.Sprintf("check cluster[%s] status failed: %v",
			dependInfo.Cluster.GetClusterID(), err))

		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckClusterStatusTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

func checkClusterStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, systemID, jobID string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// get cce client
	cli, err := api.NewCceClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] get client failed: %s", taskID, err)
		retErr := fmt.Errorf("get cloud client err, %s", err)
		return retErr
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// loop cluster status
	err = loop.LoopDoFunc(ctx, func() error {
		if systemID == "" && jobID != "" {
			rsp, errShow := cli.ShowJob(jobID)
			if errShow != nil {
				blog.Errorf("checkClusterStatus[%s] show job failed: %v", taskID, errShow)
				return nil
			}

			if rsp.Spec.ClusterUID == nil {
				blog.Errorf("checkClusterStatus[%s] show job clusterID is nil", taskID)
				return nil
			}

			systemID = *rsp.Spec.ClusterUID
			info.Cluster.SystemID = *rsp.Spec.ClusterUID

			err = cloudprovider.GetStorageModel().UpdateCluster(ctx, info.Cluster)
			if err != nil {
				blog.Errorf("checkClusterStatus[%s] updateClusterSystemID[%s] failed %s",
					taskID, info.Cluster.ClusterID, err)
				retErr := fmt.Errorf("call CreateCluster updateClusterSystemID[%s] api err: %s",
					info.Cluster.ClusterID, err)
				return retErr
			}

			blog.Infof("checkClusterStatus[%s] call CreateCluster updateClusterSystemID successful", taskID)
		}

		cluster, errGet := cli.GetCceCluster(systemID)
		if errGet != nil {
			blog.Errorf("checkClusterStatus[%s] GetCluster failed: %v", taskID, errGet)
			return nil
		}

		blog.Infof("checkClusterStatus[%s] cluster[%s] current status[%s]", taskID,
			info.Cluster.ClusterID, *cluster.Status.Phase)

		switch *cluster.Status.Phase {
		case api.Creating:
			blog.Infof("checkClusterStatus[%s] cluster[%s] creating", taskID, info.Cluster.ClusterID)
		case api.Available:
			return loop.EndLoop
		case api.Error:
			blog.Errorf("checkClusterStatus[%s] cluster[%s] error: %s", taskID, info.Cluster.ClusterID)
			return fmt.Errorf("checkClusterStatus[%s] status error, reason: %s",
				info.Cluster.ClusterID, *cluster.Status.Reason)
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
		return err
	}

	blog.Infof("checkClusterStatus[%s] cluster[%s] status ok", taskID, info.Cluster.ClusterID)

	return nil
}

// CreateCCENodeGroupTask create cce node group
func CreateCCENodeGroupTask(taskID string, stepName string) error { // nolint
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateCCENodeGroupTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateCCENodeGroupTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]

	// get dependent basic info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CreateCCENodeGroupTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err) // nolint
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cceCli, err := api.NewCceClient(dependInfo.CmOption)
	if err != nil {
		blog.Errorf("CreateCCENodeGroupTask[%s]: get cce client in task %s step %s failed, %s",
			taskID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud cce client err, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	ngs, err := cceCli.ListClusterNodeGroups(dependInfo.Cluster.SystemID)
	if err != nil {
		blog.Errorf("CreateCCENodeGroupTask[%s]: get cce all nodegroup in task %s step %s failed, %s",
			taskID, taskID, stepName, err)
		retErr := fmt.Errorf("get cce all nodegroup err, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	nodeGroup, errGet := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), nodeGroupID)
	if errGet != nil {
		blog.Errorf("CreateCCENodeGroupTask[%s]: GetNodeGroupByGroupID for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, errGet)
		retErr := fmt.Errorf("get nodegroup information failed, %s", errGet)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// 查找云是否已经存在该节点池
	nodeGroupName := strings.ToLower(nodeGroup.NodeGroupID)
	found := false
	cloudNodeGroupID := ""
	for _, ng := range ngs {
		if ng.Metadata.Name == nodeGroupName {
			cloudNodeGroupID = *ng.Metadata.Uid
			found = true
		}
	}

	if found {
		nodeGroup.CloudNodeGroupID = cloudNodeGroupID
		err = updateNodeGroupCloudNodeGroupID(nodeGroup.NodeGroupID, nodeGroup)
		if err != nil {
			blog.Errorf("CreateCCENodeGroupTask[%s]: updateNodeGroupCloudNodeGroupID[%s] in task %s step %s failed, %s",
				taskID, nodeGroup.NodeGroupID, taskID, stepName, err)
			retErr := fmt.Errorf("call CreateCCENodeGroupTask updateNodeGroupCloudNodeGroupID[%s] api err, %s",
				nodeGroup.NodeGroupID, err)
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}

		if err = state.UpdateStepSucc(start, stepName); err != nil {
			blog.Errorf("CheckCCENodeGroupStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
			return err
		}

		return nil
	}

	// cce nodePool名称以小写字母开头，由小写字母、数字、中划线(-)组成，长度范围1-50位，且不能以中划线(-)结尾
	nodeGroup.NodeGroupID = nodeGroupName
	req, err := api.GenerateCreateNodePoolRequest(nodeGroup, dependInfo.Cluster)
	if err != nil {
		blog.Errorf("CreateCCENodeGroupTask[%s]: generate create nodepool request[%s] in task %s step %s failed, %s",
			taskID, nodeGroup.NodeGroupID, taskID, stepName, err)
		retErr := fmt.Errorf("generate create nodepool request err, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	rsp, err := cceCli.CreateClusterNodePool(req)
	if err != nil {
		blog.Errorf("CreateCCENodeGroupTask[%s]: call CreateClusterNodePool[%s] api in task %s "+
			"step %s failed, %s, rsp: %+v", taskID, nodeGroup.NodeGroupID, taskID, stepName, err, rsp)
		retErr := fmt.Errorf("call CreateClusterNodePool[%s] api err, %s", nodeGroup.NodeGroupID, err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	blog.Infof("CreateCCENodeGroupTask[%s]: call CreateClusterNodePool successful", taskID)

	// 保存cce节点池id
	nodeGroup.CloudNodeGroupID = *rsp.Metadata.Uid
	// update nodegorup cloudNodeGroupID
	err = updateNodeGroupCloudNodeGroupID(nodeGroupID, nodeGroup)
	if err != nil {
		blog.Errorf("CreateCCENodeGroupTask[%s]: updateNodeGroupCloudNodeGroupID[%s] in task %s step %s failed, %s",
			taskID, nodeGroup.NodeGroupID, taskID, stepName, err)
		retErr := fmt.Errorf("call CreateCCENodeGroupTask updateNodeGroupCloudNodeGroupID[%s] api err, %s",
			nodeGroup.NodeGroupID, err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateCCENodeGroupTask[%s]: call CreateClusterNodePool updateNodeGroupCloudNodeGroupID successful",
		taskID)

	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCCENodeGroupStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// CheckCCENodeGroupsStatusTask check cce nodegroups status
func CheckCCENodeGroupsStatusTask(taskID string, stepName string) error { // nolint
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckCCENodeGroupStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckCCENodeGroupStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckCCENodeGroupsStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err) // nolint
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cceCli, err := api.NewCceClient(dependInfo.CmOption)
	if err != nil {
		blog.Errorf("CheckCCENodeGroupStatusTask[%s]: get cce client in task %s step %s failed, %s",
			taskID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud cce client err, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	nodeGroup, errGet := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), nodeGroupID)
	if errGet != nil {
		blog.Errorf("CheckCCENodeGroupStatusTask[%s]: GetNodeGroupByGroupID for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, errGet)
		retErr := fmt.Errorf("get nodegroup information failed, %s", errGet)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// loop nodegroups status
	err = loop.LoopDoFunc(ctx, func() error {
		nodePool, errLocal := cceCli.GetClusterNodePool(dependInfo.Cluster.SystemID, nodeGroup.CloudNodeGroupID)
		if errLocal != nil {
			blog.Errorf("taskID[%s] GetClusterNodePool[%s/%s] failed: %v", taskID, dependInfo.Cluster.SystemID,
				nodeGroup.CloudNodeGroupID, errLocal)
			return nil
		}

		switch {
		case nodePool.Status.Phase.Value() == api.NodePoolError:
			return fmt.Errorf("create nodegroup failed")
		case nodePool.Status.Phase.Value() == "":
			return loop.EndLoop
		default:
			blog.Infof("taskID[%s] GetClusterNodePool[%s] still creating, status[%s]",
				taskID, nodeGroup.CloudNodeGroupID, nodePool.Status.Phase.Value())
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("CheckCCENodeGroupStatusTask[%s]: GetClusterNodePool failed: %v", taskID, err)
		retErr := fmt.Errorf("GetClusterNodePool failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	if nodeGroup.AutoScaling.DesiredSize > 0 {
		_, err = cceCli.UpdateNodePoolDesiredNodes(dependInfo.Cluster.SystemID, nodeGroup.CloudNodeGroupID,
			int32(nodeGroup.AutoScaling.DesiredSize), false)
		if err != nil {
			blog.Errorf("updateNodeGroups[%s] desired nodes failed: %s", taskID, err)
			retErr := fmt.Errorf("desired nodes err, %s", err)
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
	}

	nodeGroup.Status = common.StatusRunning
	err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), nodeGroup)
	if err != nil {
		blog.Errorf("updateNodeGroups[%s] update nodegroup failed: %s", taskID, err)
		retErr := fmt.Errorf("update nodegroup failed err, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCCENodeGroupStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// CheckCCEClusterNodesStatusTask check cluster nodes status
func CheckCCEClusterNodesStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckCCEClusterNodesStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckCCEClusterNodesStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckCCEClusterNodesStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = checkClusterNodesStatus(ctx, dependInfo, nodeGroupID)
	if err != nil {
		blog.Errorf("CheckCCEClusterNodesStatusTask[%s] checkClusterStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("CheckCCEClusterNodesStatusTask[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	nodeIPs, err := updateNodeToDB(ctx, dependInfo, nodeGroupID)
	if err != nil {
		blog.Errorf("UpdateNodesToDBTask[%s] checkNodesGroupStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("UpdateNodesToDBTask[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	state.Task.CommonParams[cloudprovider.DynamicNodeIPListKey.String()] = strings.Join(nodeIPs, ",")
	state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(nodeIPs, ",")
	state.Task.CommonParams[cloudprovider.NodeNamesKey.String()] = strings.Join(nodeIPs, ",")
	state.Task.NodeIPList = nodeIPs

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCCEClusterNodesStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func checkClusterNodesStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, // nolint
	nodeGroupID string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), nodeGroupID)
	if err != nil {
		return fmt.Errorf("get nodegroup information failed, %s", err)
	}

	cceCli, err := api.NewCceClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterNodesStatus[%s] get cce client failed: %s", taskID, err)
		return fmt.Errorf("get cloud cce client err, %s", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	cloudNodeGroupID := nodeGroup.CloudNodeGroupID
	err = loop.LoopDoFunc(ctx, func() error {
		nodePool, errLocal := cceCli.GetClusterNodePool(info.Cluster.SystemID, cloudNodeGroupID)
		if errLocal != nil {
			blog.Errorf("taskID[%s] checkClusterNodesStatus[%s/%s] failed: %v", taskID, info.Cluster.SystemID,
				cloudNodeGroupID, errLocal)
			return nil
		}

		switch {
		case nodePool.Status.Phase.Value() == api.NodePoolError:
			return fmt.Errorf("node expansion failed")
		case nodePool.Status.Phase.Value() != "":
			blog.Infof("taskID[%s] checkClusterNodesStatus[%s] still creating, status[%s]",
				taskID, nodeGroupID, nodePool.Status.Phase.Value())
		case nodePool.Status.Phase.Value() == "":
			return loop.EndLoop
		default:
			return nil
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("checkClusterNodesStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
		return err
	}

	return nil
}

func updateNodeToDB(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	nodeGroupID string) ([]string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	cceCli, err := api.NewCceClient(info.CmOption)
	if err != nil {
		blog.Errorf("updateNodeToDB[%s] get cce client failed: %s", taskID, err)
		return nil, fmt.Errorf("get cloud cce client err, %s", err)
	}

	nodeIPs := make([]string, 0)
	nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), nodeGroupID)
	if err != nil {
		return nil, fmt.Errorf("updateNodeToDB GetNodeGroupByGroupID information failed, %s", err)
	}

	nodes, err := cceCli.ListClusterNodePoolNodes(info.Cluster.SystemID, nodeGroup.CloudNodeGroupID)
	if err != nil {
		blog.Errorf("checkClusterNodesStatus[%s] list nodes failed: %s", taskID, err)
		return nil, fmt.Errorf("list nodes err, %s", err)
	}

	for _, n := range nodes {
		node := &proto.Node{
			NodeID:       *n.Metadata.Uid,
			NodeName:     *n.Metadata.Name,
			NodeGroupID:  nodeGroup.NodeGroupID,
			InstanceType: n.Spec.Flavor,
			ClusterID:    info.Cluster.ClusterID,
			InnerIP:      *n.Status.PrivateIP,
			ZoneID:       n.Spec.Az,
		}

		if n.Status.Phase.Value() == api.NodeActive {
			node.Status = common.StatusRunning
		} else {
			node.Status = common.StatusAddNodesFailed
		}

		if n.Status.PrivateIPv6IP != nil {
			node.InnerIPv6 = *n.Status.PrivateIPv6IP
		}

		node.ZoneName = fmt.Sprintf("可用区%d", business.GetZoneNameByZoneId(info.Cluster.Region, n.Spec.Az))

		err = cloudprovider.GetStorageModel().CreateNode(context.Background(), node)
		if err != nil {
			return nil, fmt.Errorf("updateNodeToDB CreateNode[%s] failed, %v", node.NodeName, err)
		}

		nodeIPs = append(nodeIPs, node.InnerIP)
	}

	return nodeIPs, nil
}

// RegisterCCEClusterKubeConfigTask register cluster kubeconfig
func RegisterCCEClusterKubeConfigTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RegisterCceClusterKubeConfigTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("RegisterCceClusterKubeConfigTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("RegisterCceClusterKubeConfigTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// import cluster credential
	err = importClusterCredential(dependInfo)
	if err != nil {
		blog.Errorf("RegisterCceClusterKubeConfigTask[%s] importClusterCredential failed: %s", taskID, err)
		retErr := fmt.Errorf("importClusterCredential failed %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("RegisterCceClusterKubeConfigTask[%s] importClusterCredential success", taskID)

	// dynamic inject paras
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RegisterCceClusterKubeConfigTask[%s:%s] update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

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
		blog.Infof("UpdateCreateClusterDBInfoTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("UpdateCreateClusterDBInfoTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// handler logic
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err)
		retErr := fmt.Errorf("get cloud/project information failed, %s", err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update module name
	bkBizID, _ := strconv.Atoi(dependInfo.Cluster.GetBusinessID())
	if dependInfo.Cluster.GetClusterBasicSettings().GetModule().GetMasterModuleID() != "" {
		bkModuleID, _ := strconv.Atoi(dependInfo.Cluster.GetClusterBasicSettings().GetModule().GetMasterModuleID())
		dependInfo.Cluster.
			GetClusterBasicSettings().
			GetModule().MasterModuleName = cloudprovider.GetModuleName(bkBizID, bkModuleID)
	}
	if dependInfo.Cluster.GetClusterBasicSettings().GetModule().GetWorkerModuleID() != "" {
		bkModuleID, _ := strconv.Atoi(dependInfo.Cluster.GetClusterBasicSettings().GetModule().GetWorkerModuleID())
		dependInfo.Cluster.
			GetClusterBasicSettings().
			GetModule().WorkerModuleName = cloudprovider.GetModuleName(bkBizID, bkModuleID)
	}

	// delete passwd
	if dependInfo.Cluster.GetNodeSettings().GetMasterLogin() != nil {
		dependInfo.Cluster.GetNodeSettings().GetMasterLogin().InitLoginPassword = ""
		if dependInfo.Cluster.GetNodeSettings().GetMasterLogin().GetKeyPair() != nil {
			dependInfo.Cluster.GetNodeSettings().GetMasterLogin().GetKeyPair().KeySecret = ""
		}
	}
	if dependInfo.Cluster.GetNodeSettings().GetWorkerLogin() != nil {
		dependInfo.Cluster.GetNodeSettings().GetWorkerLogin().InitLoginPassword = ""
		if dependInfo.Cluster.GetNodeSettings().GetWorkerLogin().GetKeyPair() != nil {
			dependInfo.Cluster.GetNodeSettings().GetWorkerLogin().GetKeyPair().KeySecret = ""
		}
	}

	_ = cloudprovider.UpdateCluster(dependInfo.Cluster)

	// sync clusterData to pass-cc
	providerutils.SyncClusterInfoToPassCC(taskID, dependInfo.Cluster)

	// sync cluster perms
	providerutils.AuthClusterResourceCreatorPerm(ctx, dependInfo.Cluster.ClusterID,
		dependInfo.Cluster.ClusterName, dependInfo.Cluster.Creator)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}
