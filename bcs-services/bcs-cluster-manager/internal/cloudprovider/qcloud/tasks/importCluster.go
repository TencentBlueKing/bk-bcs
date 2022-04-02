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
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
)

// ImportClusterNodesTask call tkeInterface or kubeConfig import cluster nodes
func ImportClusterNodesTask(taskID string, stepName string) error {
	start := time.Now()
	// get task information and validate
	task, err := cloudprovider.GetStorageModel().GetTask(context.Background(), taskID)
	if err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: task %s get detail task information from storage failed, %s. " +
			"task retry", taskID, taskID, err.Error())
		return err
	}

	state := &cloudprovider.TaskState{Task: task, JobResult: cloudprovider.NewJobSyncResult(task)}
	if state.IsTerminated() {
		blog.Errorf("ImportClusterNodesTask[%s]: task %s is terminated, step %s skip", taskID, taskID, stepName)
		return fmt.Errorf("task %s terminated", taskID)
	}
	step, err := state.IsReadyToStep(stepName)
	if err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: task %s not turn to run step %s, err %s", taskID, taskID, stepName, err.Error())
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("ImportClusterNodesTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("ImportClusterNodesTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params["ClusterID"]
	cloudID := step.Params["CloudID"]

	basicInfo, err := cloudprovider.GetClusterDependBasicInfo(clusterID, cloudID)
	if err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: getClusterDependBasicInfo failed: %v", taskID, err)
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// import cluster instances
	err = importClusterInstances(basicInfo)
	if err != nil {
		blog.Errorf("ImportClusterNodesTask[%s]: importClusterInstances failed: %v", taskID, err)
		retErr := fmt.Errorf("importClusterInstances failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cloudprovider.UpdateClusterStatus(clusterID, icommon.StatusRunning)

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateClusterShieldAlarmTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func importClusterInstances(data *cloudprovider.CloudDependBasicInfo) error {
	masterIPs, nodeIPs, err := getClusterInstancesByClusterID(data)
	if err != nil {
		return err
	}

	// import cluster
	masterNodes := make(map[string]*proto.Node)
	nodes, err := transInstanceIPToNodes(masterIPs, &cloudprovider.ListNodesOption{
		Common:       data.CmOption,
		ClusterVPCID: data.Cluster.VpcID,
	})
	if err != nil {
		return nil
	}
	for _, node := range nodes {
		node.Status = icommon.StatusRunning
		masterNodes[node.InnerIP] = node
	}
	data.Cluster.Master = masterNodes

	err = importClusterNodesToCM(context.Background(), nodeIPs, &cloudprovider.ListNodesOption{
		Common:       data.CmOption,
		ClusterVPCID: data.Cluster.VpcID,
	})
	if err != nil {
		return err
	}

	return nil
}

func getClusterInstancesByClusterID(data *cloudprovider.CloudDependBasicInfo) ([]string, []string, error) {
	tkeCli, err := api.NewTkeClient(data.CmOption)
	if err != nil {
		return nil, nil, err
	}

	instancesList, err := tkeCli.QueryTkeClusterAllInstances(data.Cluster.SystemID)
	if err != nil {
		return nil, nil, err
	}

	var (
		masterIPs, nodeIPs = make([]string, 0), make([]string, 0)
	)
	for _, ins := range instancesList {
		switch ins.InstanceRole {
		case api.MASTER_ETCD.String():
			masterIPs = append(masterIPs, ins.InstanceIP)
		case api.WORKER.String():
			nodeIPs = append(nodeIPs, ins.InstanceIP)
		default:
			continue
		}
	}

	return masterIPs, nodeIPs, nil
}
