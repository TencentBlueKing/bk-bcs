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
	as "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/as/v20180419"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	cutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/encrypt"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/tenant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// CreateCloudNodeGroupTask create cloud node group task
func CreateCloudNodeGroupTask(taskID string, stepName string) error {
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start create cloud nodegroup")
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
		blog.Errorf("CheckCleanNodeGroupNodesStatusTask[%s]: GetClusterDependBasicInfo for nodegroup %s "+
			"in task %s step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error()) // nolint
		retErr := fmt.Errorf("get cloud/cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// create node group
	tkeCli, err := api.NewTkeClient(dependInfo.CmOption)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: get tke client for nodegroup[%s] in task %s "+
			"step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	// set default value for nodegroup
	if dependInfo.NodeGroup.AutoScaling != nil && dependInfo.NodeGroup.AutoScaling.VpcID == "" {
		dependInfo.NodeGroup.AutoScaling.VpcID = dependInfo.Cluster.VpcID
	}
	if dependInfo.NodeGroup.LaunchTemplate != nil {
		if dependInfo.NodeGroup.LaunchTemplate.InstanceChargeType == "" {
			dependInfo.NodeGroup.LaunchTemplate.InstanceChargeType = "POSTPAID_BY_HOUR"
		}
	}

	// create cloud nodePool
	npID, err := tkeCli.CreateClusterNodePool(generateCreateNodePoolInput(dependInfo.NodeGroup, dependInfo.Cluster))
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("create cluster nodepool failed [%s]", err))
		blog.Errorf("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool[%s] api in task %s "+
			"step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call CreateClusterNodePool[%s] api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool successful", taskID)
	dependInfo.NodeGroup.CloudNodeGroupID = npID

	// update nodegorup cloudNodeGroupID
	err = cloudprovider.UpdateNodeGroupCloudNodeGroupID(nodeGroupID, dependInfo.NodeGroup)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: updateNodeGroupCloudNodeGroupID[%s] in task %s "+
			"step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call CreateCloudNodeGroupTask updateNodeGroupCloudNodeGroupID[%s] "+
			"api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool "+
		"updateNodeGroupCloudNodeGroupID successful", taskID)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"create cloud nodegroup successful")

	state.Task.CommonParams["CloudNodeGroupID"] = npID
	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// generateCreateNodePoolInput nodePool request
func generateCreateNodePoolInput(group *proto.NodeGroup, cluster *proto.Cluster) *api.CreateNodePoolInput {
	nodePool := api.CreateNodePoolInput{
		ClusterID:                &cluster.SystemID,
		AutoScalingGroupPara:     generateAutoScalingGroupPara(group.AutoScaling),
		LaunchConfigurePara:      generateLaunchConfigurePara(group.LaunchTemplate, group.NodeTemplate),
		InstanceAdvancedSettings: generateInstanceAdvanceSettings(group.NodeTemplate),
		// 不开启腾讯云 CA 组件，因为需要部署 BCS 自己的 CA 组件
		EnableAutoscale: common.BoolPtr(false),
		Name:            &group.Name,
		Tags:            api.MapToTags(group.Tags),
	}

	// 节点池Os 当为自定义镜像时，传镜像id；否则为公共镜像的osName; 若为空复用集群级别
	// 示例值：ubuntu18.04.1x86_64
	if group.NodeTemplate != nil && group.NodeTemplate.NodeOS != "" {
		nodePool.NodePoolOs = &group.NodeTemplate.NodeOS
	}
	if group.NodeTemplate != nil {
		nodePool.Taints = api.MapToTaints(group.NodeTemplate.Taints)
		nodePool.Labels = api.MapToLabels(group.NodeTemplate.Labels)
		if group.NodeTemplate.Runtime != nil {
			nodePool.ContainerRuntime = &group.NodeTemplate.Runtime.ContainerRuntime
			nodePool.RuntimeVersion = &group.NodeTemplate.Runtime.RuntimeVersion
		}
	}
	if nodePool.AutoScalingGroupPara != nil && nodePool.AutoScalingGroupPara.VpcID == nil {
		nodePool.AutoScalingGroupPara.VpcID = &cluster.VpcID
	}
	return &nodePool
}

// CheckCloudNodeGroupStatusTask check cloud node group status task
func CheckCloudNodeGroupStatusTask(taskID string, stepName string) error { // nolint
	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"start check cloud nodegroup status")
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
		blog.Errorf("CheckCleanNodeGroupNodesStatusTask[%s]: GetClusterDependBasicInfo for nodegroup %s "+
			"in task %s step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/cluster information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get qcloud client
	tkeCli, err := api.NewTkeClient(dependInfo.CmOption)
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: get tke client for nodegroup[%s] "+
			"in task %s step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// get as client
	asCli, err := api.NewASClient(dependInfo.CmOption)
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: get as client for nodegroup[%s] "+
			"in task %s step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud as client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// wait node group state to normal
	ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Minute)
	defer cancel()

	var (
		asgID = ""
		ascID = ""
	)

	// cloud nodePool status check
	cloudNodeGroup := &tke.NodePool{}

	// loop cluster nodePool status
	err = loop.LoopDoFunc(ctx, func() error {
		np, errPool := tkeCli.DescribeClusterNodePoolDetail(dependInfo.Cluster.SystemID,
			dependInfo.NodeGroup.CloudNodeGroupID)
		if errPool != nil {
			blog.Errorf("taskID[%s] DescribeClusterNodePoolDetail[%s/%s] failed: %v",
				taskID, dependInfo.Cluster.SystemID, dependInfo.NodeGroup.CloudNodeGroupID, err)
			return nil
		}
		if np == nil {
			return nil
		}
		cloudNodeGroup = np
		asgID = *np.AutoscalingGroupId
		ascID = *np.LaunchConfigurationId
		switch {
		case *np.LifeState == api.NodeGroupLifeStateCreating:
			cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
				fmt.Sprintf("still creating, status [%s]", *np.LifeState))
			blog.Infof("taskID[%s] DescribeClusterNodePoolDetail[%s] still creating, status[%s]",
				taskID, dependInfo.NodeGroup.CloudNodeGroupID, *np.LifeState)
			return nil
		case *np.LifeState == api.NodeGroupLifeStateNormal:
			return loop.EndLoop
		default:
			return nil
		}
	}, loop.LoopInterval(5*time.Second))
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("describe cluster nodepool detail failed [%s]", err))
		blog.Errorf("taskID[%s] DescribeClusterNodePoolDetail failed: %v", taskID, err)
		return err
	}

	// get asg info
	asgArr, err := asCli.DescribeAutoScalingGroups(asgID)
	if err != nil {
		blog.Errorf("taskID[%s] DescribeAutoScalingGroups[%s] failed: %v", taskID, asgID, err)
		return err
	}

	// get launchConfiguration
	ascArr, err := asCli.DescribeLaunchConfigurations([]string{ascID})
	if err != nil {
		blog.Errorf("taskID[%s] DescribeLaunchConfigurations[%s] failed: %v", taskID, ascID, err)
		return err
	}

	// update nodeGroup
	err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(),
		generateNodeGroupFromAsgAndAsc(dependInfo.NodeGroup, cloudNodeGroup, asgArr, ascArr[0],
			dependInfo.Cluster.BusinessID))
	if err != nil {
		cloudprovider.GetStorageModel().CreateTaskStepLogError(context.Background(), taskID, stepName,
			fmt.Sprintf("update nodegroup failed [%s]", err))
		blog.Errorf("CreateCloudNodeGroupTask[%s]: updateNodeGroupCloudArgsID[%s] "+
			"in task %s step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call CreateCloudNodeGroupTask updateNodeGroupCloudArgsID[%s] "+
			"api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	cloudprovider.GetStorageModel().CreateTaskStepLogInfo(context.Background(), taskID, stepName,
		"check cloud nodegroup status successful")

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}
	return nil
}

// generateNodeGroupFromAsgAndAsc trans asg&asc to nodeGroup
func generateNodeGroupFromAsgAndAsc(group *proto.NodeGroup, cloudNodeGroup *tke.NodePool, asg *as.AutoScalingGroup,
	asc *as.LaunchConfiguration, bkBizIDString string) *proto.NodeGroup {
	group = generateNodeGroupFromAsg(group, cloudNodeGroup, asg)

	ctx, err := tenant.WithTenantIdByResourceForContext(context.Background(),
		tenant.ResourceMetaData{ProjectId: group.ProjectID})
	if err != nil {
		blog.Errorf("generateNodeGroupFromAsgAndAsc[%s] failed: %v", group.NodeGroupID, err)
	}

	if group.Area == nil {
		group.Area = &proto.CloudArea{}
	}
	cloudAreaName := cloudprovider.GetBKCloudName(ctx, int(group.Area.BkCloudID))
	group.Area.BkCloudName = cloudAreaName

	if group.NodeTemplate != nil && group.NodeTemplate.Module != nil &&
		len(group.NodeTemplate.Module.ScaleOutModuleID) != 0 {
		bkBizID, _ := strconv.Atoi(bkBizIDString)
		bkModuleID, _ := strconv.Atoi(group.NodeTemplate.Module.ScaleOutModuleID)
		group.NodeTemplate.Module.ScaleOutModuleName = cloudprovider.GetModuleName(ctx, bkBizID, bkModuleID)
	}
	return generateNodeGroupFromAsc(group, cloudNodeGroup, asc)
}

// generateNodeGroupFromAsg trans nodeGroup from asg
func generateNodeGroupFromAsg(group *proto.NodeGroup, cloudNodeGroup *tke.NodePool, // nolint
	asg *as.AutoScalingGroup) *proto.NodeGroup {
	// asg
	if asg.AutoScalingGroupId != nil {
		group.AutoScaling.AutoScalingID = *asg.AutoScalingGroupId
	}
	if asg.AutoScalingGroupName != nil {
		group.AutoScaling.AutoScalingName = *asg.AutoScalingGroupName
	}
	if asg.MaxSize != nil {
		group.AutoScaling.MaxSize = uint32(*asg.MaxSize)
	}
	if asg.MinSize != nil {
		group.AutoScaling.MinSize = uint32(*asg.MinSize)
	}
	if asg.DesiredCapacity != nil {
		group.AutoScaling.DesiredSize = uint32(*asg.DesiredCapacity)
	}
	if asg.VpcId != nil {
		group.AutoScaling.VpcID = *asg.VpcId
	}
	if asg.DefaultCooldown != nil {
		group.AutoScaling.DefaultCooldown = uint32(*asg.DefaultCooldown)
	}
	if asg.SubnetIdSet != nil {
		subnetIDs := make([]string, 0)
		for _, v := range asg.SubnetIdSet {
			subnetIDs = append(subnetIDs, *v)
		}
		group.AutoScaling.SubnetIDs = subnetIDs
	}
	if asg.RetryPolicy != nil {
		group.AutoScaling.RetryPolicy = *asg.RetryPolicy
	}
	if asg.MultiZoneSubnetPolicy != nil {
		group.AutoScaling.MultiZoneSubnetPolicy = *asg.MultiZoneSubnetPolicy
	}
	if asg.ServiceSettings != nil && asg.ServiceSettings.ReplaceMonitorUnhealthy != nil {
		group.AutoScaling.ReplaceUnhealthy = *asg.ServiceSettings.ReplaceMonitorUnhealthy
	}
	if asg.ServiceSettings != nil && asg.ServiceSettings.ScalingMode != nil {
		group.AutoScaling.ScalingMode = *asg.ServiceSettings.ScalingMode
	}

	return group
}

// generateNodeGroupFromAsc trans nodeGroup from asc
func generateNodeGroupFromAsc(group *proto.NodeGroup, cloudNodeGroup *tke.NodePool,
	asc *as.LaunchConfiguration) *proto.NodeGroup {
	// asc
	if asc.LaunchConfigurationId != nil {
		group.LaunchTemplate.LaunchConfigurationID = *asc.LaunchConfigurationId
	}
	if asc.LaunchConfigurationName != nil {
		group.LaunchTemplate.LaunchConfigureName = *asc.LaunchConfigurationName
	}
	if asc.ProjectId != nil {
		group.LaunchTemplate.ProjectID = fmt.Sprintf("%d", uint32(*asc.ProjectId))
	}
	if asc.InstanceType != nil {
		group.LaunchTemplate.InstanceType = *asc.InstanceType
	}
	if asc.InstanceChargeType != nil {
		group.LaunchTemplate.InstanceChargeType = *asc.InstanceChargeType
	}
	if asc.InternetAccessible != nil {
		group.LaunchTemplate.InternetAccess = generateInternetAccessible(asc)
	}
	if asc.SecurityGroupIds != nil {
		group.LaunchTemplate.SecurityGroupIDs = make([]string, 0)
		for _, v := range asc.SecurityGroupIds {
			group.LaunchTemplate.SecurityGroupIDs = append(group.LaunchTemplate.SecurityGroupIDs, *v)
		}
	}
	if asc.ImageId != nil {
		group.LaunchTemplate.ImageInfo = generateImageInfo(cloudNodeGroup, group, *asc.ImageId)
	}
	/*
		if asc.UserData != nil {
			group.LaunchTemplate.UserData = *asc.UserData
		}
	*/
	if asc.EnhancedService != nil {
		if asc.EnhancedService.MonitorService != nil && asc.EnhancedService.MonitorService.Enabled != nil {
			group.LaunchTemplate.IsMonitorService = *asc.EnhancedService.MonitorService.Enabled
		}
		if asc.EnhancedService.SecurityService != nil && asc.EnhancedService.SecurityService.Enabled != nil {
			group.LaunchTemplate.IsSecurityService = *asc.EnhancedService.SecurityService.Enabled
		}
	}
	if asc.ProjectId != nil {
		group.LaunchTemplate.ProjectID = fmt.Sprintf("%d", uint32(*asc.ProjectId))
	}
	return group
}

// generateInternetAccessible internet setting
func generateInternetAccessible(asc *as.LaunchConfiguration) *proto.InternetAccessible {
	internetAccess := &proto.InternetAccessible{}
	// internet bandwidth
	if asc.InternetAccessible.InternetMaxBandwidthOut != nil {
		internetAccess.InternetMaxBandwidth = strconv.Itoa(int(*asc.InternetAccessible.InternetMaxBandwidthOut))
	}
	// publicIP assign
	if asc.InternetAccessible.PublicIpAssigned != nil {
		internetAccess.PublicIPAssigned = *asc.InternetAccessible.PublicIpAssigned
	}
	// internet chargeType
	if asc.InternetAccessible.InternetChargeType != nil {
		internetAccess.InternetChargeType = *asc.InternetAccessible.InternetChargeType
	}
	if asc.InternetAccessible.BandwidthPackageId != nil {
		internetAccess.BandwidthPackageId = *asc.InternetAccessible.BandwidthPackageId
	}

	return internetAccess
}

// generateImageInfo image info
func generateImageInfo(cloudNodeGroup *tke.NodePool, group *proto.NodeGroup, imageID string) *proto.ImageInfo { // nolint
	imageInfo := &proto.ImageInfo{ImageID: imageID}
	if cloudNodeGroup != nil && cloudNodeGroup.NodePoolOs != nil {
		for _, v := range utils.ImageOsList {
			if v.ImageID == imageID {
				imageInfo.ImageName = v.Alias
				break
			}
		}
	}
	return imageInfo
}

// UpdateCreateNodeGroupDBInfoTask update create node group db info task
func UpdateCreateNodeGroupDBInfoTask(taskID string, stepName string) error {
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
		blog.Errorf("UpdateCreateNodeGroupDBInfoTask[%s]: update nodegroup status for %s failed",
			taskID, np.Status)
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateCreateNodeGroupDBInfoTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// generateAutoScalingGroupPara build asg paras
func generateAutoScalingGroupPara(as *proto.AutoScalingGroup) *api.AutoScalingGroup {
	if as == nil {
		return nil
	}

	// autoscaling group
	asg := &api.AutoScalingGroup{
		MaxSize:         common.Uint64Ptr(uint64(as.MaxSize)),
		MinSize:         common.Uint64Ptr(uint64(as.MinSize)),
		SubnetIds:       common.StringPtrs(as.SubnetIDs),
		DesiredCapacity: common.Uint64Ptr(uint64(as.DesiredSize)),
	}
	if as.AutoScalingName != "" {
		asg.AutoScalingGroupName = common.StringPtr(as.AutoScalingName)
	}
	if as.VpcID != "" {
		asg.VpcID = common.StringPtr(as.VpcID)
	}
	if as.DefaultCooldown != 0 {
		asg.DefaultCooldown = common.Uint64Ptr(uint64(as.DefaultCooldown))
	}
	if as.RetryPolicy != "" {
		asg.RetryPolicy = common.StringPtr(as.RetryPolicy)
	}
	if as.ScalingMode != "" {
		asg.ServiceSettings = &api.ServiceSettings{ScalingMode: common.StringPtr(as.ScalingMode)}
	}
	if as.MultiZoneSubnetPolicy != "" {
		asg.MultiZoneSubnetPolicy = common.StringPtr(as.MultiZoneSubnetPolicy)
	}
	return asg
}

// generateLaunchConfigurePara launch template paras
func generateLaunchConfigurePara(template *proto.LaunchConfiguration,
	nodeTemplate *proto.NodeTemplate) *api.LaunchConfiguration { // nolint
	if template == nil {
		return nil
	}
	// launch config
	conf := &api.LaunchConfiguration{
		LaunchConfigurationName: &template.LaunchConfigureName,
		InstanceType:            &template.InstanceType,
		InstanceChargeType:      &template.InstanceChargeType,
		LoginSettings: &api.LoginSettings{
			Password: func() string {
				if len(template.InitLoginPassword) == 0 {
					return ""
				}
				passwd, _ := encrypt.Decrypt(nil, template.InitLoginPassword)
				return passwd
			}(),
			KeyIds: func() []string {
				if template.GetKeyPair() == nil || template.GetKeyPair().GetKeyID() == "" {
					return nil
				}

				return []string{template.GetKeyPair().GetKeyID()}
			}(),
		},
		SecurityGroupIds: common.StringPtrs(template.SecurityGroupIDs),
	}
	// system disks
	if template.SystemDisk != nil {
		conf.SystemDisk = &api.SystemDisk{
			DiskType: &template.SystemDisk.DiskType}
		diskSize, _ := strconv.Atoi(template.SystemDisk.DiskSize)
		conf.SystemDisk.DiskSize = common.Uint64Ptr(uint64(diskSize))
	}
	// data disks
	if template.DataDisks != nil {
		conf.DataDisks = make([]*api.LaunchConfigureDataDisk, 0)
		for _, v := range template.DataDisks {
			diskType := v.DiskType
			disk := &api.LaunchConfigureDataDisk{DiskType: &diskType}
			diskSize, _ := strconv.Atoi(v.DiskSize)
			disk.DiskSize = common.Uint64Ptr(uint64(diskSize))
			conf.DataDisks = append(conf.DataDisks, disk)
		}
	}
	// internet access
	if template.InternetAccess != nil {
		bw, _ := strconv.Atoi(template.InternetAccess.InternetMaxBandwidth)
		conf.InternetAccessible = &api.InternetAccessible{
			PublicIPAssigned:        common.BoolPtr(template.InternetAccess.PublicIPAssigned),
			InternetMaxBandwidthOut: common.Uint64Ptr(uint64(bw)),
		}
		if !template.InternetAccess.PublicIPAssigned {
			conf.InternetAccessible.InternetMaxBandwidthOut = common.Uint64Ptr(0)
		}
		if template.InternetAccess.InternetChargeType != "" {
			conf.InternetAccessible.InternetChargeType = common.StringPtr(template.InternetAccess.InternetChargeType)
		}
		if template.InternetAccess.BandwidthPackageId != "" {
			conf.InternetAccessible.BandwidthPackageID = common.StringPtr(template.InternetAccess.BandwidthPackageId)
		}
	}
	// enhanced service
	conf.EnhancedService = &api.EnhancedService{
		SecurityService: &api.RunSecurityServiceEnabled{Enabled: common.BoolPtr(template.IsSecurityService)},
		MonitorService:  &api.RunMonitorServiceEnabled{Enabled: common.BoolPtr(template.IsMonitorService)},
	}
	return conf
}

// generateInstanceAdvanceSettings build instance advanced setting
func generateInstanceAdvanceSettings(template *proto.NodeTemplate) *api.InstanceAdvancedSettings {
	if template == nil {
		return nil
	}

	// instance advanced setting
	result := &api.InstanceAdvancedSettings{
		Unschedulable: common.Int64Ptr(int64(template.UnSchedulable)),
	}
	if template.MountTarget != "" {
		result.MountTarget = template.MountTarget
	}
	if template.DockerGraphPath != "" {
		result.DockerGraphPath = template.DockerGraphPath
	}
	if template.PreStartUserScript != "" {
		result.UserScript = template.PreStartUserScript
	}
	if template.Labels != nil {
		result.Labels = make([]*api.KeyValue, 0)
		for k, v := range template.Labels {
			result.Labels = append(result.Labels, &api.KeyValue{Name: k, Value: v})
		}
	}
	// data disks
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
	// extra args
	kubeletMap := cutils.GetKubeletParas(template)
	// parse kubelet extra args
	kubeletArgs := strings.Split(kubeletMap[icommon.Kubelet], ";")
	if len(kubeletArgs) > 0 {
		result.ExtraArgs = &api.InstanceExtraArgs{Kubelet: utils.FilterEmptyString(kubeletArgs)}
	}

	return result
}
