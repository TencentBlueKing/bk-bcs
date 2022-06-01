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
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/actions"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// CreateCloudNodeGroupTask create cloud node group task
func CreateCloudNodeGroupTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	cloudID := step.Params["CloudID"]
	nodeGroupID := step.Params["NodeGroupID"]
	group, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: get nodegroup for %s failed", taskID, nodeGroupID)
		retErr := fmt.Errorf("get nodegroup information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get cloud and cluster info
	cloud, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), cloudID, group.ClusterID)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: get cloud/cluster for nodegroup %s in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: get credential for nodegroup %s in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption.Region = group.Region

	// create node group
	tkeCli, err := api.NewTkeClient(cmOption)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: get tke client for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}
	nodePool := api.CreateNodePoolInput{
		ClusterID:                &group.ClusterID,
		AutoScalingGroupPara:     generateAutoScalingGroupPara(group.AutoScaling),
		LaunchConfigurePara:      generateLaunchConfigurePara(group.LaunchTemplate),
		InstanceAdvancedSettings: generateInstanceAdvanceSettings(group.NodeTemplate),
		// 不开启腾讯云 CA 组件，因为需要部署 BCS 自己的 CA 组件
		EnableAutoscale: common.BoolPtr(false),
		Name:            &group.Name,
		Labels:          api.MapToLabels(group.Labels),
		Taints:          api.MapToTaints(group.Taints),
	}
	npID, err := tkeCli.CreateClusterNodePool(&nodePool)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool[%s] api in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call CreateClusterNodePool[%s] api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool successful", taskID)
	group.CloudNodeGroupID = npID

	// update nodegorup cloudNodeGroupID
	err = updateNodeGroupCloudNodeGroupID(nodeGroupID, npID)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: updateNodeGroupCloudNodeGroupID[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call CreateCloudNodeGroupTask updateNodeGroupCloudNodeGroupID[%s] api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool updateNodeGroupCloudNodeGroupID successful", taskID)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	state.Task.CommonParams["CloudNodeGroupID"] = npID
	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// CheckCloudNodeGroupStatusTask check cloud node group status task
func CheckCloudNodeGroupStatusTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	nodeGroupID := step.Params["NodeGroupID"]
	cloudID := step.Params["CloudID"]

	group, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: get nodegroup for %s failed", taskID, nodeGroupID)
		retErr := fmt.Errorf("get nodegroup information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	cloud, cluster, err := actions.GetCloudAndCluster(cloudprovider.GetStorageModel(), cloudID, group.ClusterID)
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: get cloud/cluster for nodegroup %s in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get dependency resource for cloudprovider operation
	cmOption, err := cloudprovider.GetCredential(&cloudprovider.CredentialData{
		Cloud:     cloud,
		AccountID: cluster.CloudAccountID,
	})
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: get credential for nodegroup %s in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud credential err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption.Region = group.Region

	// get qcloud client
	cli, err := api.NewTkeClient(cmOption)
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: get tke client for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// wait node group state to normal
	ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Minute)
	defer cancel()
	asgID := ""
	err = cloudprovider.LoopDoFunc(ctx, func() error {
		np, err := cli.DescribeClusterNodePoolDetail(group.ClusterID, group.CloudNodeGroupID)
		if err != nil {
			blog.Errorf("taskID[%s] DescribeClusterNodePoolDetail[%s/%s] failed: %v", taskID, group.ClusterID,
				group.CloudNodeGroupID, err)
			return nil
		}
		if np == nil {
			return nil
		}
		asgID = *np.AutoscalingGroupId
		switch {
		case *np.LifeState == api.NodeGroupLifeStateCreating:
			blog.Infof("taskID[%s] DescribeClusterNodePoolDetail[%s] still creating, status[%s]",
				taskID, group.CloudNodeGroupID, *np.LifeState)
			return nil
		case *np.LifeState == api.NodeGroupLifeStateNormal:
			return cloudprovider.EndLoop
		default:
			return nil
		}
	}, cloudprovider.LoopInterval(5*time.Second))
	if err != nil {
		blog.Errorf("taskID[%s] DescribeClusterNodePoolDetail failed: %v", taskID, err)
		return err
	}
	updateNodeGroupASGID(nodeGroupID, asgID)

	// wait all nodes to be ready
	err = cloudprovider.LoopDoFunc(ctx, func() error {
		np, err := cli.DescribeClusterNodePoolDetail(group.ClusterID, group.CloudNodeGroupID)
		if err != nil {
			blog.Errorf("taskID[%s] DescribeClusterNodePoolDetail[%s/%s] failed: %v", taskID, group.ClusterID,
				group.CloudNodeGroupID, err)
			return nil
		}
		if np == nil || np.NodeCountSummary == nil {
			return nil
		}
		if np.NodeCountSummary.ManuallyAdded == nil || np.NodeCountSummary.AutoscalingAdded == nil {
			return nil
		}
		allNormalNodesCount := *np.NodeCountSummary.ManuallyAdded.Normal + *np.NodeCountSummary.AutoscalingAdded.Normal
		switch {
		case *np.DesiredNodesNum == allNormalNodesCount:
			return cloudprovider.EndLoop
		default:
			return nil
		}
	}, cloudprovider.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("taskID[%s] DescribeClusterNodePoolDetail failed: %v", taskID, err)
		return err
	}
	return nil
}

// InstallAutoScalerTask install auto scaler task
func InstallAutoScalerTask(taskID string, stepName string) error {
	return nil
}

// UpdateCreateNodeGroupDBInfoTask update create node group db info task
func UpdateCreateNodeGroupDBInfoTask(taskID string, stepName string) error {
	start := time.Now()
	//get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	// step login started here
	nodeGroupID := step.Params["NodeGroupID"]

	np, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		blog.Errorf("UpdateCreateNodeGroupDBInfoTask[%s]: get cluster for %s failed", taskID, nodeGroupID)
		retErr := fmt.Errorf("get nodegroup information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	np.Status = icommon.StatusRunning

	err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), np)
	if err != nil {
		blog.Errorf("UpdateCreateNodeGroupDBInfoTask[%s]: update nodegroup status for %s failed", taskID, np.Status)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateCreateNodeGroupDBInfoTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func generateAutoScalingGroupPara(as *proto.AutoScalingGroup) *api.AutoScalingGroup {
	if as == nil {
		return nil
	}
	return &api.AutoScalingGroup{
		AutoScalingGroupName: common.StringPtr(as.AutoScalingName),
		MaxSize:              common.Uint64Ptr(uint64(as.MaxSize)),
		MinSize:              common.Uint64Ptr(uint64(as.MinSize)),
		// TODO 使用集群的 VPC
		VpcID:                 common.StringPtr(as.VpcID),
		DefaultCooldown:       common.Uint64Ptr(uint64(as.DefaultCooldown)),
		SubnetIds:             common.StringPtrs(as.SubnetIDs),
		DesiredCapacity:       common.Uint64Ptr(uint64(as.DesiredSize)),
		RetryPolicy:           common.StringPtr(as.RetryPolicy),
		ServiceSettings:       &api.ServiceSettings{ScalingMode: common.StringPtr(as.ScalingMode)},
		MultiZoneSubnetPolicy: common.StringPtr(as.MultiZoneSubnetPolicy),
	}
}

func generateLaunchConfigurePara(template *proto.LaunchConfiguration) *api.LaunchConfiguration {
	if template == nil {
		return nil
	}
	conf := &api.LaunchConfiguration{
		LaunchConfigurationName: &template.LaunchConfigureName,
		InstanceType:            &template.InstanceType,
		InstanceChargeType:      common.StringPtr("POSTPAID_BY_HOUR"),
		InternetAccessible: &api.InternetAccessible{
			InternetChargeType: common.StringPtr("TRAFFIC_POSTPAID_BY_HOUR"),
		},
		LoginSettings:    &api.LoginSettings{Password: template.InitLoginPassword},
		SecurityGroupIds: common.StringPtrs(template.SecurityGroupIDs),
	}
	if template.ImageInfo != nil {
		conf.ImageID = &template.ImageInfo.ImageID
	}
	if template.SystemDisk != nil {
		conf.SystemDisk = &api.SystemDisk{
			DiskType: &template.SystemDisk.DiskType}
		diskSize, err := strconv.Atoi(template.SystemDisk.DiskSize)
		if err != nil && diskSize > 0 {
			conf.SystemDisk.DiskSize = common.Uint64Ptr(uint64(diskSize))
		}
	}
	if template.DataDisks != nil {
		conf.DataDisks = make([]*api.DataDisk, 0)
		for _, v := range template.DataDisks {
			disk := &api.DataDisk{DiskType: v.DiskType}
			diskSize, err := strconv.Atoi(v.DiskSize)
			if err != nil && diskSize > 0 {
				disk.DiskSize = uint32(diskSize)
			}
			conf.DataDisks = append(conf.DataDisks, disk)
		}
	}
	if template.InternetAccess != nil {
		if template.InternetAccess.InternetChargeType != "" {
			conf.InternetAccessible.InternetChargeType = common.StringPtr(template.InternetAccess.InternetChargeType)
		}
		bandwidth, err := strconv.Atoi(template.InternetAccess.InternetMaxBandwidth)
		if err == nil && bandwidth > 0 {
			conf.InternetAccessible.InternetMaxBandwidthOut = common.Uint64Ptr(uint64(bandwidth))
		}
		conf.InternetAccessible.PublicIPAssigned = common.BoolPtr(template.InternetAccess.PublicIPAssigned)
	}
	if template.InstanceChargeType != "" {
		conf.InstanceChargeType = common.StringPtr(template.InstanceChargeType)
	}
	conf.EnhancedService = &api.EnhancedService{
		SecurityService: template.IsSecurityService,
		MonitorService:  template.IsMonitorService,
	}
	return conf
}

func generateInstanceAdvanceSettings(template *proto.NodeTemplate) *api.InstanceAdvancedSettings {
	if template == nil {
		return nil
	}
	result := &api.InstanceAdvancedSettings{
		MountTarget:     template.MountTarget,
		DockerGraphPath: template.DockerGraphPath,
		Unschedulable:   common.Int64Ptr(int64(template.UnSchedulable)),
		UserScript:      template.UserScript,
	}
	if template.Labels != nil {
		result.Labels = make([]*api.KeyValue, 0)
		for k, v := range template.Labels {
			result.Labels = append(result.Labels, &api.KeyValue{Name: k, Value: v})
		}
	}
	if template.DataDisks != nil {
		result.DataDisks = make([]api.DataDetailDisk, 0)
		for _, v := range template.DataDisks {
			diskSize, _ := strconv.Atoi(v.DiskSize)
			result.DataDisks = append(result.DataDisks, api.DataDetailDisk{
				DiskType:           v.DiskType,
				DiskSize:           int64(diskSize),
				MountTarget:        v.MountTarget,
				FileSystem:         v.FileSystem,
				AutoFormatAndMount: v.AutoFormatAndMount,
			})
		}
	}
	if template.ExtraArgs != nil {
		// parse kubelet extra args
		kubeletArgs := strings.Split(template.ExtraArgs["kubelet"], ";")
		if len(kubeletArgs) > 0 {
			result.ExtraArgs = &api.InstanceExtraArgs{Kubelet: kubeletArgs}
		}
	}
	return result
}
