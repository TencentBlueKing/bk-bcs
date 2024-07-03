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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
)

// CreateClusterTask call qcloud interface to create cluster
func CreateClusterTask(taskID string, stepName string) error {
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
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error()) // nolint
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	req, err := api.GenerateCreateClusterRequest(ctx, dependInfo.Cluster, operator)
	if err != nil {
		blog.Errorf("createCluster[%s] generateCreateClusterRequest failed: %v", taskID, err)
		return err
	}

	// create cluster task
	clsId, err := createCluster(ctx, dependInfo, req, dependInfo.Cluster.SystemID)
	if err != nil {
		blog.Errorf("CreateClusterTask[%s] createCluster for cluster[%s] failed, %s",
			taskID, clusterID, err.Error())
		retErr := fmt.Errorf("createCluster err, %s", err.Error())
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

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func createCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	request *api.CreateClusterRequest, clsId string) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	client, err := api.NewCceClient(info.CmOption)
	if err != nil {
		return "", err
	}

	if clsId != "" {
		cluster, errGet := client.GetCceCluster(clsId)
		if errGet != nil {
			blog.Errorf("createCluster[%s] GetCluster[%s] failed, %s",
				taskID, info.Cluster.ClusterID, errGet.Error())
			retErr := fmt.Errorf("call GetCluster[%s] api err, %s", info.Cluster.ClusterID, errGet.Error())
			return "", retErr
		}
		// update cluster systemID
		info.Cluster.SystemID = *cluster.Metadata.Uid
	} else {
		rsp, err := client.CreateCluster(request)
		if err != nil {
			return "", err
		}

		info.Cluster.SystemID = *rsp.Metadata.Uid
	}

	err = cloudprovider.GetStorageModel().UpdateCluster(ctx, info.Cluster)
	if err != nil {
		blog.Errorf("createCluster[%s] updateClusterSystemID[%s] failed %s",
			taskID, info.Cluster.ClusterID, err.Error())
		retErr := fmt.Errorf("call CreateCluster updateClusterSystemID[%s] api err: %s",
			info.Cluster.ClusterID, err.Error())
		return "", retErr
	}
	blog.Infof("createCluster[%s] call CreateCluster updateClusterSystemID successful", taskID)

	return info.Cluster.SystemID, nil
}

// CheckClusterStatusTask check cluster status
func CheckClusterStatusTask(taskID string, stepName string) error {
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

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckClusterStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = checkClusterStatus(ctx, dependInfo, systemID)
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

func checkClusterStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, systemID string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// get qcloud client
	cli, err := api.NewCceClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] get client failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("get cloud client err, %s", err.Error())
		return retErr
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// loop cluster status
	err = loop.LoopDoFunc(ctx, func() error {
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
				info.Cluster.ClusterID, cluster.Status.Reason)
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

// RegisterCceClusterKubeConfigTask register cluster kubeconfig
func RegisterCceClusterKubeConfigTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RegisterClusterKubeConfigTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("RegisterClusterKubeConfigTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("RegisterTkeClusterKubeConfigTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// import cluster credential
	err = importClusterCredential(dependInfo)
	if err != nil {
		blog.Errorf("RegisterTkeClusterKubeConfigTask[%s] importClusterCredential failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("importClusterCredential failed %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("RegisterTkeClusterKubeConfigTask[%s] importClusterCredential success", taskID)

	// dynamic inject paras
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RegisterTkeClusterKubeConfigTask[%s:%s] update to storage fatal", taskID, stepName)
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
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
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

// CreateCCENodeGroupTask create cce node group
func CreateCCENodeGroupTask(taskID string, stepName string) error {
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
	nodeGroupIDs := step.Params[cloudprovider.NodeGroupIDKey.String()]

	// get dependent basic info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CreateCCENodeGroupTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error()) // nolint
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cceCli, err := api.NewCceClient(dependInfo.CmOption)
	if err != nil {
		blog.Errorf("CreateCCENodeGroupTask[%s]: get cce client in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud cce client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	ngs, err := cceCli.ListClusterNodeGroups(dependInfo.Cluster.SystemID)
	if err != nil {
		blog.Errorf("CreateCCENodeGroupTask[%s]: get cce all nodegroup in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cce all nodegroup err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	nodeGroups := make([]*proto.NodeGroup, 0)
	for _, ngID := range strings.Split(nodeGroupIDs, ",") {
		nodeGroup, errGet := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if errGet != nil {
			blog.Errorf("CreateCCENodeGroupTask[%s]: GetNodeGroupByGroupID for cluster %s in task %s "+
				"step %s failed, %s", taskID, clusterID, taskID, stepName, errGet.Error())
			retErr := fmt.Errorf("get nodegroup information failed, %s", errGet.Error())
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

		if !found {
			nodeGroups = append(nodeGroups, nodeGroup)
		} else if found && nodeGroup.CloudNodeGroupID == "" {
			nodeGroup.CloudNodeGroupID = cloudNodeGroupID
			err = updateNodeGroupCloudNodeGroupID(nodeGroup.NodeGroupID, nodeGroup)
			if err != nil {
				blog.Errorf("CreateCCENodeGroupTask[%s]: updateNodeGroupCloudNodeGroupID[%s] in task %s step %s failed, %s",
					taskID, nodeGroup.NodeGroupID, taskID, stepName, err.Error())
				retErr := fmt.Errorf("call CreateCCENodeGroupTask updateNodeGroupCloudNodeGroupID[%s] api err, %s",
					nodeGroup.NodeGroupID, err.Error())
				_ = state.UpdateStepFailure(start, stepName, retErr)
				return retErr
			}
		}
	}

	for _, group := range nodeGroups {
		// cce nodePool名称以小写字母开头，由小写字母、数字、中划线(-)组成，长度范围1-50位，且不能以中划线(-)结尾
		group.NodeGroupID = strings.ToLower(group.NodeGroupID)
		req, err := api.GenerateCreateNodePoolRequest(group, dependInfo.Cluster)
		if err != nil {
			blog.Errorf("CreateCCENodeGroupTask[%s]: generate create nodepool request[%s] in task %s step %s failed, %s",
				taskID, group.NodeGroupID, taskID, stepName, err.Error())
			retErr := fmt.Errorf("generate create nodepool request err, %s", err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}

		rsp, err := cceCli.CreateClusterNodePool(req)
		if err != nil {
			blog.Errorf("CreateCCENodeGroupTask[%s]: call CreateClusterNodePool[%s] api in task %s "+
				"step %s failed, %s, rsp: %+v", taskID, group.NodeGroupID, taskID, stepName, err.Error(), rsp)
			retErr := fmt.Errorf("call CreateClusterNodePool[%s] api err, %s", group.NodeGroupID, err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}

		blog.Infof("CreateCCENodeGroupTask[%s]: call CreateClusterNodePool successful", taskID)

		// 保存cce节点池id
		group.CloudNodeGroupID = *rsp.Metadata.Uid
		// update nodegorup cloudNodeGroupID
		err = updateNodeGroupCloudNodeGroupID(group.NodeGroupID, group)
		if err != nil {
			blog.Errorf("CreateCCENodeGroupTask[%s]: updateNodeGroupCloudNodeGroupID[%s] in task %s step %s failed, %s",
				taskID, group.NodeGroupID, taskID, stepName, err.Error())
			retErr := fmt.Errorf("call CreateCCENodeGroupTask updateNodeGroupCloudNodeGroupID[%s] api err, %s",
				group.NodeGroupID, err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
		blog.Infof("CreateCCENodeGroupTask[%s]: call CreateClusterNodePool updateNodeGroupCloudNodeGroupID successful",
			taskID)

		time.Sleep(time.Second)
	}

	return nil
}

// CheckCCENodeGroupsStatusTask check cce nodegroups status
func CheckCCENodeGroupsStatusTask(taskID string, stepName string) error {
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
	nodeGroupIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params, cloudprovider.NodeGroupIDKey.String(), ",")
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})

	cceCli, err := api.NewCceClient(dependInfo.CmOption)
	if err != nil {
		blog.Errorf("CheckCCENodeGroupStatusTask[%s]: get cce client in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud cce client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	nodeGroups := make([]*proto.NodeGroup, 0)
	for _, ngID := range nodeGroupIDs {
		nodeGroup, errGet := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if errGet != nil {
			blog.Errorf("CheckCCENodeGroupStatusTask[%s]: GetNodeGroupByGroupID for cluster %s in task %s "+
				"step %s failed, %s", taskID, clusterID, taskID, stepName, errGet.Error())
			retErr := fmt.Errorf("get nodegroup information failed, %s", errGet.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}

		nodeGroups = append(nodeGroups, nodeGroup)
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	statusMap := make(map[string]bool)
	// loop nodegroups status
	err = loop.LoopDoFunc(ctx, func() error {
		for _, group := range nodeGroups {
			nodePool, errLocal := cceCli.GetClusterNodePool(dependInfo.Cluster.SystemID, group.CloudNodeGroupID)
			if errLocal != nil {
				blog.Errorf("taskID[%s] GetClusterNodePool[%s/%s] failed: %v", taskID, dependInfo.Cluster.SystemID,
					group.CloudNodeGroupID, errLocal)
				return nil
			}

			switch {
			case nodePool.Status.Phase.Value() == api.NodePoolError:
				statusMap[group.NodeGroupID] = false
			case nodePool.Status.Phase.Value() == "":
				statusMap[group.NodeGroupID] = true
			default:
				blog.Infof("taskID[%s] GetClusterNodePool[%s] still creating, status[%s]",
					taskID, group.CloudNodeGroupID, nodePool.Status.Phase.Value())
			}

			if len(statusMap) == len(nodeGroups) {
				return loop.EndLoop
			}
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("CheckCCENodeGroupStatusTask[%s]: GetClusterNodePool failed: %v", taskID, err)
		retErr := fmt.Errorf("GetClusterNodePool failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	success := make([]string, 0)
	Failure := make([]string, 0)
	for id, ok := range statusMap {
		if ok {
			success = append(success, id)
		} else {
			Failure = append(Failure, id)
		}
	}

	blog.Infof("CheckCCENodeGroupStatusTask[%s] success[%v] failure[%v]", taskID, success, Failure)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	if len(Failure) > 0 {
		state.Task.CommonParams[cloudprovider.FailedNodeGroupIDsKey.String()] = strings.Join(Failure, ",")
	}

	if len(success) == 0 {
		blog.Errorf("CheckCCENodeGroupStatusTask[%s] nodegroups init failed", taskID)
		retErr := fmt.Errorf("节点池初始化失败, 请联系管理员")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	state.Task.CommonParams[cloudprovider.SuccessNodeGroupIDsKey.String()] = strings.Join(success, ",")

	ctx = cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = updateNodeGroups(ctx, dependInfo, Failure, success)
	if err != nil {
		blog.Errorf("UpdateCCENodesGroupToDBTask[%s] updateNodeGroups[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("UpdateCCENodesGroupToDBTask[%s] timeout|abnormal", clusterID)
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

func updateNodeGroups(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	addFailedNodeGroupIDs, addSuccessNodeGroupIDs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(addFailedNodeGroupIDs) > 0 {
		for _, ngID := range addFailedNodeGroupIDs {
			err := cloudprovider.UpdateNodeGroupStatus(ngID, common.StatusCreateNodeGroupFailed)
			if err != nil {
				return fmt.Errorf("updateNodeGroups updateNodeGroupStatus[%s] failed, %v", ngID, err)
			}
		}
	}

	for _, ngID := range addSuccessNodeGroupIDs {
		nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return fmt.Errorf("updateNodeGroups GetNodeGroupByGroupID failed, %s", err.Error())
		}

		if nodeGroup.AutoScaling.DesiredSize > 0 {
			cli, err := api.NewCceClient(info.CmOption)
			if err != nil {
				blog.Errorf("updateNodeGroups[%s] get cce client failed: %s", taskID, err.Error())
				return fmt.Errorf("get cloud aks client err, %s", err.Error())
			}

			_, err = cli.UpdateNodePoolDesiredNodes(info.Cluster.SystemID, nodeGroup.CloudNodeGroupID,
				int32(nodeGroup.AutoScaling.DesiredSize), false)
			if err != nil {
				blog.Errorf("updateNodeGroups[%s] desired nodes failed: %s", taskID, err.Error())
				return fmt.Errorf("desired nodes err, %s", err.Error())
			}

			nodeGroup.Status = common.StatusRunning
			err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), nodeGroup)
			if err != nil {
				return fmt.Errorf("updateNodeGroups UpdateNodeGroup[%s] failed, %s",
					nodeGroup.NodeGroupID, err.Error())
			}
		}
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
	nodeGroupIDs := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.SuccessNodeGroupIDsKey.String(), ",")

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckCCEClusterNodesStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	addSuccessNodes, addFailureNodes, err := checkClusterNodesStatus(ctx, dependInfo, nodeGroupIDs)
	if err != nil {
		blog.Errorf("CheckCCEClusterNodesStatusTask[%s] checkClusterStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("checkClusterStatus[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(addFailureNodes) > 0 {
		state.Task.CommonParams[cloudprovider.FailedClusterNodeIDsKey.String()] = strings.Join(addFailureNodes, ",")
	}
	if len(addSuccessNodes) == 0 {
		blog.Errorf("CheckCCEClusterNodesStatusTask[%s] nodes init failed", taskID)
		retErr := fmt.Errorf("节点初始化失败, 请联系管理员")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(addSuccessNodes, ",")

	err = updateNodeToDB(ctx, dependInfo, nodeGroupIDs)
	if err != nil {
		blog.Errorf("UpdateNodesToDBTask[%s] checkNodesGroupStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("UpdateNodesToDBTask[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	state.Task.CommonParams[cloudprovider.NodeIPsKey.String()] = strings.Join(addSuccessNodes, ",")

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCCEClusterNodesStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func checkClusterNodesStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, // nolint
	nodeGroupIDs []string) ([]string, []string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	cloudNodeGroupIDs := make([]string, 0)
	for _, ngID := range nodeGroupIDs {
		nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return nil, nil, fmt.Errorf("get nodegroup information failed, %s", err.Error())
		}
		cloudNodeGroupIDs = append(cloudNodeGroupIDs, nodeGroup.CloudNodeGroupID)
	}

	cceCli, err := api.NewCceClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterNodesStatus[%s] get cce client failed: %s", taskID, err.Error())
		return nil, nil, fmt.Errorf("get cloud cce client err, %s", err.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	err = loop.LoopDoFunc(ctx, func() error {
		index := 0
		for _, id := range nodeGroupIDs {
			nodePool, errLocal := cceCli.GetClusterNodePool(info.Cluster.SystemID, id)
			if errLocal != nil {
				blog.Errorf("taskID[%s] checkClusterNodesStatus[%s/%s] failed: %v", taskID, info.Cluster.SystemID,
					id, errLocal)
				return nil
			}

			switch {
			case nodePool.Status.Phase.Value() == api.NodePoolError:
				index++
			case nodePool.Status.Phase.Value() != "":
				blog.Infof("taskID[%s] checkClusterNodesStatus[%s] still creating, status[%s]",
					taskID, id, nodePool.Status.Phase.Value())
			case nodePool.Status.Phase.Value() == "":
				index++
			default:
				return nil
			}
		}

		if index == len(nodeGroupIDs) {
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(10*time.Second))

	success, failure := make([]string, 0), make([]string, 0)
	for _, id := range nodeGroupIDs {
		nodes, err := cceCli.ListClusterNodePoolNodes(info.Cluster.SystemID, id)
		if err != nil {
			blog.Errorf("checkClusterNodesStatus[%s] list nodes failed: %s", taskID, err.Error())
			return nil, nil, fmt.Errorf("list nodes err, %s", err.Error())
		}

		for _, node := range nodes {
			if node.Status.Phase.Value() == api.NodeActive {
				success = append(success, *node.Metadata.Uid)
			} else {
				failure = append(failure, *node.Metadata.Uid)
			}
		}
	}

	return success, failure, nil
}

func updateNodeToDB(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeGroupIDs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)
	cceCli, err := api.NewCceClient(info.CmOption)
	if err != nil {
		blog.Errorf("updateNodeToDB[%s] get cce client failed: %s", taskID, err.Error())
		return fmt.Errorf("get cloud cce client err, %s", err.Error())
	}

	for _, ngID := range nodeGroupIDs {
		nodeGroup, err := actions.GetNodeGroupByGroupID(cloudprovider.GetStorageModel(), ngID)
		if err != nil {
			return fmt.Errorf("updateNodeToDB GetNodeGroupByGroupID information failed, %s", err.Error())
		}

		nodes, err := cceCli.ListClusterNodePoolNodes(info.Cluster.SystemID, nodeGroup.CloudNodeGroupID)
		if err != nil {
			blog.Errorf("checkClusterNodesStatus[%s] list nodes failed: %s", taskID, err.Error())
			return fmt.Errorf("list nodes err, %s", err.Error())
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
				return fmt.Errorf("updateNodeToDB CreateNode[%s] failed, %v", node.NodeName, err)
			}
		}
	}

	return nil
}
