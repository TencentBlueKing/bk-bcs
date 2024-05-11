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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/huawei/api"
	providerutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
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

	var (
		abnormal = false
	)

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
			return loop.EndLoop
		case api.Available:
			abnormal = true
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
		return err
	}

	if abnormal {
		blog.Errorf("checkClusterStatus[%s] GetCluster[%s] failed: abnormal", taskID, info.Cluster.ClusterID)
		retErr := fmt.Errorf("cluster[%s] status abnormal", info.Cluster.ClusterID)
		return retErr
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
