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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/business"
)

func allocateClusterSubnets(ctx context.Context,
	subnet *proto.SubnetSource, clusterID, vpcID string, opt *cloudprovider.CommonOption) ([]string, error) {
	taskId := cloudprovider.GetTaskIDFromContext(ctx)

	if subnet == nil || (len(subnet.GetNew()) == 0 && len(subnet.Existed.GetIds()) == 0) {
		return nil, fmt.Errorf("allocateClusterSubnets subnet data empty")
	}

	subnetIds := make([]string, 0)
	if len(subnet.GetExisted().GetIds()) > 0 {
		subnetIds = append(subnetIds, subnet.GetExisted().GetIds()...)
	}

	allocateSubnets, err := business.AllocateClusterVpcCniSubnets(ctx,
		clusterID, vpcID, subnet.GetNew(), opt)
	if err != nil {
		blog.Errorf("allocateClusterSubnets AllocateClusterVpcCniSubnets[%s] failed: %v", taskId, err)
		return nil, err
	}
	subnetIds = append(subnetIds, allocateSubnets...)

	return subnetIds, err
}

// AllocateClusterSubnetTask allocate cluster subnet
func AllocateClusterSubnetTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("AllocateClusterSubnetTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("AllocateClusterSubnetTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	subnetInfo := step.Params[cloudprovider.SubnetInfoKey.String()]
	cloudID := state.Task.CommonParams[cloudprovider.CloudIDKey.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("AllocateClusterSubnetTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	var subnetSource *proto.SubnetSource
	if len(subnetInfo) > 0 {
		_ = json.Unmarshal([]byte(subnetInfo), &subnetSource)
	} else {
		subnetSource = dependInfo.Cluster.GetNetworkSettings().GetSubnetSource()
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	subnetIDs, err := allocateClusterSubnets(ctx, subnetSource, dependInfo.Cluster.ClusterID,
		dependInfo.Cluster.VpcID, dependInfo.CmOption)
	if err != nil {
		blog.Errorf("AllocateClusterSubnetTask[%s] failed: %v",
			taskID, err.Error())
		retErr := fmt.Errorf("AllocateClusterSubnetTask[%s] abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	if len(subnetIDs) == 0 {
		blog.Errorf("AllocateClusterSubnetTask[%s] failed: subnetIDs empty", taskID)
		retErr := fmt.Errorf("AllocateClusterSubnetTask[%s] subnetIds empty", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	state.Task.CommonParams[cloudprovider.SubnetIDKey.String()] = strings.Join(subnetIDs, ",")

	// update cluster subnets
	dependInfo.Cluster.NetworkSettings.EniSubnetIDs = subnetIDs
	err = cloudprovider.GetStorageModel().UpdateCluster(ctx, dependInfo.Cluster)
	if err != nil {
		blog.Errorf("AllocateClusterSubnetTask[%s] update cluster failed: %v",
			dependInfo.Cluster.ClusterID, err.Error())
		return err
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("AllocateClusterSubnetTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// OpenClusterVpcCniTask open cluster vpc-cni networkMode
func OpenClusterVpcCniTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("OpenClusterVpcCniTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("OpenClusterVpcCniTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := state.Task.CommonParams[cloudprovider.CloudIDKey.String()]
	systemID := state.Task.CommonParams[cloudprovider.CloudSystemID.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("OpenClusterVpcCniTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// 调用api打开vpc-cni模式
	err = enableTkeClusterVpcCni(ctx, systemID, dependInfo.Cluster.NetworkSettings.EniSubnetIDs,
		dependInfo.Cluster, dependInfo.CmOption)
	if err != nil {
		blog.Errorf("openTkeClusterVpcCni[%s] enableTkeClusterVpcCni failed: %v",
			dependInfo.Cluster.ClusterID, err.Error())
		retErr := fmt.Errorf("openClusterVpcCniTask[%s] abnormal, %s", clusterID, err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("OpenClusterVpcCniTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func disableTkeClusterVpcCni(systemID string, opt *cloudprovider.CommonOption) error {
	client, err := api.NewTkeClient(opt)
	if err != nil {
		return err
	}

	err = client.CloseVpcCniMode(systemID)
	if err != nil {
		return err
	}

	return nil
}

// CloseClusterVpcCniTask close cluster vpc-cni networkMode
func CloseClusterVpcCniTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CloseClusterVpcCniTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CloseClusterVpcCniTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := state.Task.CommonParams[cloudprovider.CloudIDKey.String()]
	systemID := state.Task.CommonParams[cloudprovider.CloudSystemID.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CloseClusterVpcCniTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	err = disableTkeClusterVpcCni(systemID, dependInfo.CmOption)
	if err != nil {
		blog.Errorf("CloseClusterVpcCniTask[%s] disableTkeClusterVpcCni failed: %v",
			taskID, err)
		retErr := fmt.Errorf("CloseClusterVpcCniTask[%s] abnormal %s", clusterID, err)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	blog.Errorf("CloseClusterVpcCniTask[%s] disableTkeClusterVpcCni successful", taskID)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CloseClusterVpcCniTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}
