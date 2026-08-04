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
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/eks"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
)

// DeleteCloudNodeGroupTask delete cloud node group task
func DeleteCloudNodeGroupTask(taskID string, stepName string) error {
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
		blog.Errorf("DeleteCloudNodeGroupTask[%s]: getClusterDependBasicInfo failed: %v", taskID, err)
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption := dependInfo.CmOption
	cluster := dependInfo.Cluster
	group := dependInfo.NodeGroup

	// create node group
	eksCli, err := api.NewAWSClientSet(cmOption)
	if err != nil {
		blog.Errorf("DeleteCloudNodeGroupTask[%s]: get eks client for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get eks client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}
	found := true
	asgName := ""
	if group.CloudNodeGroupID != "" {
		ng, desErr := eksCli.DescribeNodegroup(&group.CloudNodeGroupID, &cluster.SystemID)
		if desErr != nil {
			if !strings.Contains(desErr.Error(), "ResourceNotFoundException") {
				blog.Errorf(
					"DeleteCloudNodeGroupTask[%s]: call DescribeClusterNodePoolDetail[%s] api in task %s step %s failed, %s",
					taskID, nodeGroupID, taskID, stepName, err)
				retErr := fmt.Errorf("call DescribeClusterNodePoolDetail[%s] api err, %s", nodeGroupID, err)
				_ = state.UpdateStepFailure(start, stepName, retErr)
				return retErr
			}

			blog.Warnf("DeleteCloudNodeGroupTask[%s]: nodegroup[%s/%s] in task %s step %s not found, skip delete",
				taskID, nodeGroupID, group.CloudNodeGroupID, stepName, stepName)
			found = false
		}
		asgName = *ng.Resources.AutoScalingGroups[0].Name
	}
	if found && group.CloudNodeGroupID != "" {
		ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
		asgInfo, err := eksCli.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{&asgName}})
		if err != nil {
			blog.Errorf("DeleteCloudNodeGroupTask[%s]: call DescribeAutoScalingGroups[%s] api in task %s step %s failed, %s",
				taskID, nodeGroupID, taskID, stepName, err.Error())
			retErr := fmt.Errorf("call DescribeAutoScalingGroups[%s] api err, %s", nodeGroupID, err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}

		err = resumeAWSProcesses(ctx, dependInfo, asgInfo[0])
		if err != nil {
			blog.Errorf("DeleteCloudNodeGroupTask[%s]: call resumeAWSProcesses[%s] api in task %s step %s failed, %s",
				taskID, nodeGroupID, taskID, stepName, err.Error())
			retErr := fmt.Errorf("call resumeAWSProcesses[%s] api err, %s", nodeGroupID, err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}

		_, err = eksCli.DeleteNodegroup(&eks.DeleteNodegroupInput{
			NodegroupName: &group.CloudNodeGroupID,
			ClusterName:   &cluster.SystemID})
		if err != nil {
			blog.Errorf("DeleteCloudNodeGroupTask[%s]: call DeleteNodegroup[%s] api in task %s step %s failed, %s",
				taskID, nodeGroupID, taskID, stepName, err.Error())
			retErr := fmt.Errorf("call DeleteNodegroup[%s] api err, %s", nodeGroupID, err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
	}
	blog.Infof("DeleteCloudNodeGroupTask[%s]: call DeleteNodegroup successful", taskID)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("DeleteCloudNodeGroupTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func resumeAWSProcesses(ctx context.Context, dependInfo *cloudprovider.CloudDependBasicInfo,
	asInfo *autoscaling.Group) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	client, err := api.NewAutoScalingClient(dependInfo.CmOption)
	if err != nil {
		blog.Errorf("taskID[%s] resumeAWSProcesses get aws clientSet failed, %s", taskID, err.Error())
		return err
	}

	pNames := []string{"AZRebalance", "HealthCheck", "ReplaceUnhealthy"}

	err = client.ResumeProcesses(asInfo.AutoScalingGroupName, pNames)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Minute)
	defer cancel()
	err = loop.LoopDoFunc(ctx, func() error {
		asgs, dErr := client.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
			AutoScalingGroupNames: []*string{asInfo.AutoScalingGroupName},
		})
		if dErr != nil {
			blog.Errorf("taskID[%s] DescribeAutoScalingGroups failed, %s", taskID, dErr.Error())
			return nil
		}

		if len(asgs) == 0 {
			blog.Errorf("taskID[%s] get autoscaling group info empty", taskID)
			return nil
		}
		if len(asgs[0].SuspendedProcesses) > 0 {
			blog.Errorf("taskID[%s] resume autoscaling group processes is not empty", taskID)
			return nil
		}

		if len(asgs[0].SuspendedProcesses) == 0 {
			blog.Infof("resume autoscaling group all processes successful")
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(5*time.Second))
	if err != nil {
		blog.Errorf("resumeAWSProcesses[%s]: failed: %v", taskID, err)
		retErr := fmt.Errorf("resumeAWSProcesses failed, %s", err.Error())
		return retErr
	}

	return nil
}
