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
	"encoding/base64"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/api"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
)

const (
	defaultStorageDeviceName = "/dev/xvda"
)

// CreateCloudNodeGroupTask create cloud node group task
func CreateCloudNodeGroupTask(taskID string, stepName string) error { // nolint
	start := time.Now()
	// get task information and validate
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	if step == nil {
		return nil
	}

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	// GetClusterDependBasicInfo get cluster/cloud/nodeGroup depend info, nodeGroup may be nil.
	// only get metadata, try not to change it
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:   clusterID,
		CloudID:     cloudID,
		NodeGroupID: nodeGroupID,
	})
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CreateCloudNodeGroupTask GetClusterDependBasicInfo failed")
		// UpdateStepFailure update step failure
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption := dependInfo.CmOption
	cluster := dependInfo.Cluster
	group := dependInfo.NodeGroup

	// step login started here
	client, err := api.NewAWSClientSet(cmOption)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: get aws clientSet for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get aws client set err, %s", err.Error())
		// UpdateStepFailure update step failure
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	// set default value for nodegroup
	if group.AutoScaling != nil && group.AutoScaling.VpcID == "" {
		group.AutoScaling.VpcID = cluster.VpcID
	}

	input, err := generateCreateNodegroupInput(group, cluster, client)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: call generateCreateNodegroupInput[%s] api in task %s "+
			"step %s failed, %s", taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call CreateClusterNodePool[%s] api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	ng, err := client.CreateNodegroup(input)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool[%s] api in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call CreateClusterNodePool[%s] api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool successful", taskID)
	group.CloudNodeGroupID = *ng.NodegroupName

	// update nodegorup
	err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), group)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: updateNodeGroupCloudNodeGroupID[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call CreateCloudNodeGroupTask updateNodeGroupCloudNodeGroupID[%s] api err, %s", nodeGroupID,
			err.Error())
		// UpdateStepFailure update step failure
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateCloudNodeGroupTask[%s]: call CreateNodegroup updateNodeGroupCloudNodeGroupID successful",
		taskID)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	state.Task.CommonParams["CloudNodeGroupID"] = *ng.NodegroupName
	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// create launch template
func createLaunchTemplate(cluster *proto.Cluster, group *proto.NodeGroup, cli *api.EC2Client) (
	*api.LaunchTemplate, error) {
	ltData, _ := buildLaunchTemplateData(group)
	launchTemplateCreateInput := &ec2.CreateLaunchTemplateInput{
		LaunchTemplateName: aws.String(fmt.Sprintf("eks-lt-%s-%s", cluster.SystemID, group.NodeGroupID)),
		LaunchTemplateData: ltData,
		TagSpecifications: []*ec2.TagSpecification{
			{
				ResourceType: aws.String(api.ResourceTypeLaunchTemplate),
				Tags: []*ec2.Tag{
					{
						Key:   aws.String("eks:cluster-name"),
						Value: aws.String(cluster.SystemID),
					},
					{
						Key:   aws.String("eks:nodegroup-name"),
						Value: aws.String(group.NodeGroupID),
					},
				},
			},
		},
	}
	// CreateLaunchTemplate creates a LaunchTemplate
	output, creErr := cli.CreateLaunchTemplate(launchTemplateCreateInput)
	if creErr != nil {
		return nil, creErr
	}
	return &api.LaunchTemplate{
		LaunchTemplateName:  output.LaunchTemplateName,
		LaunchTemplateId:    output.LaunchTemplateId,
		LatestVersionNumber: output.LatestVersionNumber,
	}, nil
}

// generate create node group input
func generateCreateNodegroupInput(group *proto.NodeGroup, cluster *proto.Cluster,
	cli *api.AWSClientSet) (*api.CreateNodegroupInput, error) {
	if group.AutoScaling == nil || group.LaunchTemplate == nil || group.LaunchTemplate.SystemDisk == nil {
		return nil, fmt.Errorf("generateCreateNodegroupInput AutoScaling|LaunchTemplate|SystemDisk is empty ")
	}
	nodeGroup := &api.CreateNodegroupInput{
		AmiType:       aws.String(group.LaunchTemplate.ImageInfo.ImageName),
		ClusterName:   &cluster.SystemID,
		NodegroupName: &group.NodeGroupID,
		ScalingConfig: &api.NodegroupScalingConfig{
			DesiredSize: aws.Int64(int64(group.AutoScaling.DesiredSize)),
			MaxSize:     aws.Int64(int64(group.AutoScaling.MaxSize)),
			MinSize:     aws.Int64(int64(group.AutoScaling.MinSize)),
		},
		Subnets: aws.StringSlice(group.AutoScaling.SubnetIDs),
	}
	nodeGroup.CapacityType = &group.LaunchTemplate.InstanceChargeType
	if nodeGroup.CapacityType != aws.String(eks.CapacityTypesOnDemand) &&
		nodeGroup.CapacityType != aws.String(eks.CapacityTypesSpot) {
		nodeGroup.CapacityType = aws.String(eks.CapacityTypesOnDemand)
	}
	if len(group.Labels) != 0 {
		nodeGroup.Labels = aws.StringMap(group.Labels)
	}
	if len(group.Tags) != 0 {
		nodeGroup.Tags = aws.StringMap(group.Tags)
	}
	if group.NodeTemplate != nil {
		nodeGroup.Taints = api.MapToTaints(group.NodeTemplate.Taints)
	}

	lt, err := createLaunchTemplate(cluster, group, cli.EC2Client)
	if err != nil {
		blog.Errorf("create launch template failed, %v", err)
		return nil, fmt.Errorf("generateCreateNodegroupInput createLaunchTemplate failed, %v", err)
	}
	nodeGroup.LaunchTemplate = &api.LaunchTemplateSpecification{
		Id: lt.LaunchTemplateId, Version: aws.String(strconv.Itoa(int(*lt.LatestVersionNumber)))}

	role, err := cli.GetRole(&iam.GetRoleInput{RoleName: aws.String(group.AutoScaling.GetServiceRole())})
	if err != nil {
		blog.Errorf("GetRole failed, %v", err)
		return nil, fmt.Errorf("generateCreateNodegroupInput GetRole failed, %v", err)
	}
	nodeGroup.NodeRole = role.Arn

	return nodeGroup, nil
}

// getEKSOptimizedImages get eks optimized images
func getEKSOptimizedImages(cli *api.EC2Client, cluster *proto.Cluster) (*ec2.Image, error) { // nolint
	imageList, err := cli.DescribeImages(&ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{Name: aws.String("name"), Values: []*string{aws.String(fmt.Sprintf("amazon-eks-node-al*%s*",
				cluster.ClusterBasicSettings.Version))}},
			{Name: aws.String("architecture"), Values: []*string{aws.String("x86_64")}},
			{Name: aws.String("state"), Values: []*string{aws.String("available")}},
		},
		Owners: aws.StringSlice([]string{"amazon"}),
	})
	if err != nil {
		return nil, err
	}
	// sort imageList by name
	sort.Slice(imageList, func(i, j int) bool {
		return *imageList[i].Name > *imageList[j].Name
	})

	if len(imageList) == 0 {
		return nil, fmt.Errorf("got empty image list")
	}

	return imageList[0], nil
}

// build Launch Template Data
func buildLaunchTemplateData(group *proto.NodeGroup) (*ec2.RequestLaunchTemplateData, error) {
	sysDiskSize, _ := strconv.Atoi(group.LaunchTemplate.SystemDisk.DiskSize)
	launchTemplateData := &ec2.RequestLaunchTemplateData{
		BlockDeviceMappings: []*ec2.LaunchTemplateBlockDeviceMappingRequest{
			{
				DeviceName: aws.String(defaultStorageDeviceName),
				Ebs: &ec2.LaunchTemplateEbsBlockDeviceRequest{
					VolumeSize:          aws.Int64(int64(sysDiskSize)),
					DeleteOnTermination: aws.Bool(true),
					VolumeType:          aws.String(group.LaunchTemplate.SystemDisk.DiskType),
				},
			},
		},
		InstanceType: aws.String(group.LaunchTemplate.InstanceType),
		KeyName:      aws.String(group.LaunchTemplate.GetKeyPair().GetKeyID()),
		NetworkInterfaces: []*ec2.LaunchTemplateInstanceNetworkInterfaceSpecificationRequest{
			{
				AssociatePublicIpAddress: aws.Bool(group.LaunchTemplate.InternetAccess.PublicIPAssigned),
				DeviceIndex:              aws.Int64(0),
				Groups:                   aws.StringSlice(group.LaunchTemplate.SecurityGroupIDs),
			},
		},
	}

	if len(group.LaunchTemplate.DataDisks) != 0 {
		for k, v := range group.LaunchTemplate.DataDisks {
			if k >= len(api.DeviceName) {
				return nil, fmt.Errorf("data disks counts can't larger than %d", len(api.DeviceName))
			}
			size, _ := strconv.Atoi(v.DiskSize)
			launchTemplateData.BlockDeviceMappings = append(launchTemplateData.BlockDeviceMappings,
				&ec2.LaunchTemplateBlockDeviceMappingRequest{
					DeviceName: aws.String(api.DeviceName[k]),
					Ebs: &ec2.LaunchTemplateEbsBlockDeviceRequest{
						VolumeSize:          aws.Int64(int64(size)),
						DeleteOnTermination: aws.Bool(true),
						VolumeType:          aws.String(v.DiskType),
					},
				})
		}
	}

	script := group.NodeTemplate.PreStartUserScript
	if script != "" {
		scriptByte, _ := base64.StdEncoding.DecodeString(script)
		script = api.DefaultUserDataHeader + string(scriptByte) + api.DefaultUserDataTail
		script = base64.StdEncoding.EncodeToString([]byte(script))
		launchTemplateData.UserData = aws.String(script)
	}

	return launchTemplateData, nil
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
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

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

	asgInfo, ltvInfo, err := checkNodegroupStatus(ctx, dependInfo)
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: getClusterDependBasicInfo failed: %v", taskID, err)
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(),
		generateNodeGroupFromAsgAndLtv(dependInfo.NodeGroup, asgInfo, ltvInfo))
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: updateNodeGroupCloudArgsID[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call CreateCloudNodeGroupTask updateNodeGroupCloudArgsID[%s] api err, %s", nodeGroupID,
			err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func checkNodegroupStatus(rootCtx context.Context, dependInfo *cloudprovider.CloudDependBasicInfo) (*autoscaling.Group,
	*ec2.LaunchTemplateVersion, error) { // nolint
	taskID := cloudprovider.GetTaskIDFromContext(rootCtx)
	cmOption := dependInfo.CmOption
	cluster := dependInfo.Cluster
	group := dependInfo.NodeGroup
	// get eks clientSet
	client, err := api.NewAWSClientSet(cmOption)
	if err != nil {
		blog.Errorf("taskID[%s] checkNodegroupStatus get aws clientSet failed, %s", taskID, err.Error())
		return nil, nil, err
	}
	// wait node group state to normal
	ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Minute)
	defer cancel()
	var asgName, ltName, ltVersion *string
	err = loop.LoopDoFunc(ctx, func() error {
		ng, desErr := client.DescribeNodegroup(&group.CloudNodeGroupID, &cluster.SystemID)
		if desErr != nil {
			blog.Errorf("taskID[%s] DescribeNodegroup[%s/%s] failed: %v", taskID, cluster.SystemID,
				group.CloudNodeGroupID, desErr)
			return nil
		}

		if ng.Resources != nil && ng.Resources.AutoScalingGroups != nil {
			asgName = ng.Resources.AutoScalingGroups[0].Name
			asg, errGet := client.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
				AutoScalingGroupNames: []*string{asgName},
			})
			if errGet != nil {
				return errGet
			}
			ltName = asg[0].MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.LaunchTemplateName
			ltVersion = asg[0].MixedInstancesPolicy.LaunchTemplate.LaunchTemplateSpecification.Version
		}

		switch *ng.Status {
		case api.NodeGroupStatusCreating:
			blog.Infof("taskID[%s] DescribeNodegroup[%s] still creating, status[%s]",
				taskID, group.CloudNodeGroupID, *ng.Status)
			return nil
		case api.NodeGroupStatusCreateFailed:
			return fmt.Errorf("NodeGroup[%s] create failed, status[%s]", group.NodeGroupID, *ng.Status)
		case api.NodeGroupStatusActive:
			return loop.EndLoop
		default:
			return nil
		}
	}, loop.LoopInterval(5*time.Second))
	if err != nil {
		blog.Errorf("checkNodegroupStatus[%s]: failed: %v", taskID, err)
		retErr := fmt.Errorf("checkNodegroupStatus failed, %s", err.Error())
		return nil, nil, retErr
	}

	asgInfo, err := client.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{asgName}})
	if err != nil {
		return nil, nil, err
	}

	ltvInfo, err := getAsgAndLtv(client, ltName, ltVersion)
	if err != nil {
		blog.Errorf("checkNodegroupStatus[%s]: getAsgAndLtv failed: %v", taskID, err)
		retErr := fmt.Errorf("checkNodegroupStatus getAsgAndLtv failed, %s", err.Error())
		return nil, nil, retErr
	}

	return asgInfo[0], ltvInfo, nil
}

// get Asg And Ltv
func getAsgAndLtv(client *api.AWSClientSet, ltName, ltVersion *string) (*ec2.LaunchTemplateVersion, error) {
	// get launchTemplateVersion
	ltvInfo, err := client.DescribeLaunchTemplateVersions(&ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateName: ltName, Versions: []*string{ltVersion}})
	if err != nil {
		return nil, err
	}

	return ltvInfo[0], nil
}

// generate Node Group From Asg And Ltv
func generateNodeGroupFromAsgAndLtv(group *proto.NodeGroup, asg *autoscaling.Group,
	ltv *ec2.LaunchTemplateVersion) *proto.NodeGroup {
	group = generateNodeGroupFromAsg(group, asg)
	return generateNodeGroupFromLtv(group, ltv)
}

// generate Node Group From Asg
func generateNodeGroupFromAsg(group *proto.NodeGroup, asg *autoscaling.Group) *proto.NodeGroup {
	if asg.AutoScalingGroupName != nil {
		group.AutoScaling.AutoScalingName = *asg.AutoScalingGroupName
		group.AutoScaling.AutoScalingID = *asg.AutoScalingGroupARN
	}
	if asg.AvailabilityZones != nil {
		for _, z := range asg.AvailabilityZones {
			group.AutoScaling.Zones = append(group.AutoScaling.Zones, *z)
		}
	}
	if asg.DesiredCapacity != nil {
		group.AutoScaling.DesiredSize = uint32(*asg.DesiredCapacity)
	}
	if asg.DefaultCooldown != nil {
		group.AutoScaling.DefaultCooldown = uint32(*asg.DefaultCooldown)
	}

	return group
}

// generate Node Group From Ltv
func generateNodeGroupFromLtv(group *proto.NodeGroup, ltv *ec2.LaunchTemplateVersion) *proto.NodeGroup {
	if ltv.LaunchTemplateId != nil {
		group.LaunchTemplate.LaunchConfigurationID = *ltv.LaunchTemplateId
	}
	if ltv.LaunchTemplateName != nil {
		group.LaunchTemplate.LaunchConfigureName = *ltv.LaunchTemplateName
	}
	if ltv.LaunchTemplateData != nil {
		if ltv.LaunchTemplateData.Monitoring != nil {
			group.LaunchTemplate.IsMonitorService = *ltv.LaunchTemplateData.Monitoring.Enabled
		}
	}

	return group
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
		blog.Errorf("UpdateCreateNodeGroupDBInfoTask[%s]: update nodegroup status for %s failed", taskID, np.Status)
	}

	// update step
	if err := state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateCreateNodeGroupDBInfoTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}
