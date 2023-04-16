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
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/aws/api"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
)

const (
	launchTemplateNameFormat = "bcs-managed-lt-%s"
	launchTemplateTagKey     = "bcs-managed-template"
	launchTemplateTagValue   = "do-not-modify-or-delete"
	defaultStorageDeviceName = "/dev/xvda"
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

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	nodeGroupID := step.Params[cloudprovider.NodeGroupIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(clusterID, cloudID, nodeGroupID)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: GetClusterDependBasicInfo failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("CreateCloudNodeGroupTask GetClusterDependBasicInfo failed")
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
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	// set default value for nodegroup
	if group.AutoScaling != nil && group.AutoScaling.VpcID == "" {
		group.AutoScaling.VpcID = cluster.VpcID
	}

	ng, err := client.CreateNodegroup(generateCreateNodegroupInput(group, cluster, client))
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool[%s] api in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call CreateClusterNodePool[%s] api err, %s", nodeGroupID, err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool successful", taskID)
	group.CloudNodeGroupID = *ng.NodegroupName

	// update nodegorup cloudNodeGroupID
	err = cloudprovider.UpdateNodeGroupCloudNodeGroupID(nodeGroupID, group)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: updateNodeGroupCloudNodeGroupID[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call CreateCloudNodeGroupTask updateNodeGroupCloudNodeGroupID[%s] api err, %s", nodeGroupID,
			err.Error())
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

func generateCreateNodegroupInput(group *proto.NodeGroup, cluster *proto.Cluster,
	cli *api.AWSClientSet) *api.CreateNodegroupInput {
	if group.AutoScaling == nil || group.LaunchTemplate == nil || group.LaunchTemplate.SystemDisk == nil {
		return nil
	}
	sysDiskSize, _ := strconv.Atoi(group.LaunchTemplate.SystemDisk.DiskSize)
	nodeGroup := &api.CreateNodegroupInput{
		ClusterName:   &cluster.SystemID,
		NodegroupName: &group.NodeGroupID,
		ScalingConfig: &api.NodegroupScalingConfig{
			DesiredSize: aws.Int64(0),
			MaxSize:     aws.Int64(int64(group.AutoScaling.MaxSize)),
			MinSize:     aws.Int64(int64(group.AutoScaling.MinSize)),
		},
		DiskSize: aws.Int64(int64(sysDiskSize)),
		Tags:     aws.StringMap(group.Tags),
		Labels:   aws.StringMap(group.Labels),
	}
	nodeGroup.CapacityType = &group.LaunchTemplate.InstanceChargeType
	if nodeGroup.CapacityType != aws.String(eks.CapacityTypesOnDemand) &&
		nodeGroup.CapacityType != aws.String(eks.CapacityTypesSpot) {
		nodeGroup.CapacityType = aws.String(eks.CapacityTypesOnDemand)
	}
	if group.NodeTemplate != nil {
		nodeGroup.Taints = api.MapToTaints(group.NodeTemplate.Taints)
	}

	lt, err := createLaunchTemplate(cluster.SystemID, cli.EC2Client)
	if err != nil {
		blog.Errorf("create launch template failed, %v", err)
		return nil
	}
	nodeGroup.LaunchTemplate, err = createNewLaunchTemplateVersion(*lt.LaunchTemplateId, nodeGroup, group, cli.EC2Client)
	if err != nil {
		blog.Errorf("createNewLaunchTemplateVersion failed, %v", err)
		return nil
	}

	role, err := cli.GetRole(&iam.GetRoleInput{RoleName: aws.String(group.NodeRole)})
	if err != nil {
		blog.Errorf("GetRole failed, %v", err)
		return nil
	}
	nodeGroup.NodeRole = role.Arn

	return nodeGroup
}

func createLaunchTemplate(clusterName string, cli *api.EC2Client) (*api.LaunchTemplate, error) {
	// Create first version of the launch template as default version. It will not be used for any node group.
	lt, err := cli.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
		LaunchTemplateNames: []*string{aws.String(fmt.Sprintf(launchTemplateNameFormat, clusterName))}})
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			launchTemplateCreateInput := &api.CreateLaunchTemplateInput{
				LaunchTemplateName: aws.String(fmt.Sprintf(launchTemplateNameFormat, clusterName)),
				LaunchTemplateData: &api.RequestLaunchTemplateData{
					UserData: aws.String("bcs managed lt user data"),
				},
				TagSpecifications: []*api.TagSpecification{
					{
						ResourceType: aws.String(api.ResourceTypeLaunchTemplate),
						Tags: []*api.Tag{
							{
								Key:   aws.String(launchTemplateTagKey),
								Value: aws.String(launchTemplateTagValue),
							},
						},
					},
				},
			}
			output, err := cli.CreateLaunchTemplate(launchTemplateCreateInput)
			if err != nil {
				return nil, err
			}
			return &api.LaunchTemplate{
				LaunchTemplateName:  output.LaunchTemplateName,
				LaunchTemplateId:    output.LaunchTemplateId,
				LatestVersionNumber: output.LatestVersionNumber,
			}, nil
		}
		return nil, err
	}

	return &api.LaunchTemplate{
		LaunchTemplateName:  lt[0].LaunchTemplateName,
		LaunchTemplateId:    lt[0].LaunchTemplateId,
		LatestVersionNumber: lt[0].LatestVersionNumber,
	}, nil
}

func createNewLaunchTemplateVersion(ltID string, input *api.CreateNodegroupInput, group *proto.NodeGroup,
	cli *api.EC2Client) (*api.LaunchTemplateSpecification, error) {
	ltData, err := buildLaunchTemplateData(input, group, cli)
	if err != nil {
		return nil, err
	}

	launchTemplateVersionInput := &ec2.CreateLaunchTemplateVersionInput{
		LaunchTemplateData: ltData,
		LaunchTemplateId:   aws.String(ltID),
	}

	output, err := cli.CreateLaunchTemplateVersion(launchTemplateVersionInput)
	if err != nil {
		return nil, err
	}
	version := strconv.Itoa(int(*output.VersionNumber))

	return &api.LaunchTemplateSpecification{
		Id:      output.LaunchTemplateName,
		Name:    output.LaunchTemplateId,
		Version: aws.String(version),
	}, nil
}

func buildLaunchTemplateData(input *api.CreateNodegroupInput, group *proto.NodeGroup, cli *api.EC2Client) (
	*ec2.RequestLaunchTemplateData, error) {
	var imageID *string
	if group.LaunchTemplate.ImageInfo != nil {
		imageID = aws.String(group.LaunchTemplate.ImageInfo.ImageID)
	}

	deviceName := aws.String(defaultStorageDeviceName)
	if rootDeviceName, err := getImageRootDeviceName([]*string{imageID}, cli); err != nil {
		return nil, err
	} else if rootDeviceName != nil {
		deviceName = rootDeviceName
	}

	userdata := group.LaunchTemplate.UserData
	if userdata != "" {
		userdata = api.DefaultUserDataHeader + userdata + api.DefaultUserDataTail
		userdata = base64.StdEncoding.EncodeToString([]byte(userdata))
	}

	launchTemplateData := &ec2.RequestLaunchTemplateData{
		ImageId:          imageID,
		KeyName:          aws.String(group.LaunchTemplate.SshKey),
		UserData:         aws.String(userdata),
		InstanceType:     aws.String(group.LaunchTemplate.InstanceType),
		SecurityGroupIds: aws.StringSlice(group.LaunchTemplate.SecurityGroupIDs),
		BlockDeviceMappings: []*ec2.LaunchTemplateBlockDeviceMappingRequest{
			{
				DeviceName: deviceName,
				Ebs: &ec2.LaunchTemplateEbsBlockDeviceRequest{
					VolumeSize:          input.DiskSize,
					DeleteOnTermination: aws.Bool(true),
					VolumeType:          aws.String(group.LaunchTemplate.SystemDisk.DiskType),
				},
			},
		},
		TagSpecifications: api.CreateTagSpecs(aws.StringMap(group.Tags)),
	}

	if len(group.LaunchTemplate.DataDisks) != 0 {
		for k, v := range group.LaunchTemplate.DataDisks {
			if k >= len(api.DeviceName) {
				return nil, fmt.Errorf("data disks counts can't larger than %d", len(api.DeviceName))
			}
			size, err := strconv.Atoi(v.DiskSize)
			if err != nil {
				return nil, err
			}
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
	launchTemplateData.InstanceType = aws.String(group.LaunchTemplate.InstanceType)

	return launchTemplateData, nil
}

func getImageRootDeviceName(imageID []*string, cli *api.EC2Client) (*string, error) {
	describeOutput, err := cli.DescribeImages(&ec2.DescribeImagesInput{ImageIds: imageID})
	if err != nil {
		return nil, err
	}
	return describeOutput[0].RootDeviceName, nil
}

// CheckCloudNodeGroupStatusTask check cloud node group status task
func CheckCloudNodeGroupStatusTask(taskID string, stepName string) error {
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

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(clusterID, cloudID, nodeGroupID)
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: getClusterDependBasicInfo failed: %v", taskID, err)
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption := dependInfo.CmOption
	cluster := dependInfo.Cluster
	group := dependInfo.NodeGroup

	// get eks client
	client, err := api.NewAWSClientSet(cmOption)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: get aws clientSet for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get aws client set err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}

	// wait node group state to normal
	ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Minute)
	defer cancel()
	var asgName, ltName, ltVersion *string
	err = cloudprovider.LoopDoFunc(ctx, func() error {
		ng, err := client.DescribeNodegroup(&group.CloudNodeGroupID, &cluster.SystemID)
		if err != nil {
			blog.Errorf("taskID[%s] DescribeClusterNodePoolDetail[%s/%s] failed: %v", taskID, cluster.SystemID,
				group.CloudNodeGroupID, err)
			return nil
		}
		if ng == nil {
			return nil
		}

		if ng.Resources != nil && ng.Resources.AutoScalingGroups != nil {
			asgName = ng.Resources.AutoScalingGroups[0].Name
		}
		if ng.LaunchTemplate != nil {
			ltName = ng.LaunchTemplate.Name
			ltVersion = ng.LaunchTemplate.Version
		}
		switch {
		case *ng.Status == api.NodeGroupStatusCreating:
			blog.Infof("taskID[%s] DescribeNodegroup[%s] still creating, status[%s]",
				taskID, group.CloudNodeGroupID, *ng.Status)
			return nil
		case *ng.Status == api.NodeGroupStatusActive:
			return cloudprovider.EndLoop
		default:
			return nil
		}
	}, cloudprovider.LoopInterval(5*time.Second))
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: DescribeNodegroup failed: %v", taskID, err)
		retErr := fmt.Errorf("DescribeNodegroup failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	asgInfo, ltvInfo, err := getAsgAndLtv(client, asgName, ltName, ltVersion)
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: getAsgAndLtv failed: %v", taskID, err)
		retErr := fmt.Errorf("getAsgAndLtv failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), generateNodeGroupFromAsgAndLtv(group,
		asgInfo[0], ltvInfo[0]))
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

func getAsgAndLtv(client *api.AWSClientSet, asgName, ltName, ltVersion *string) ([]*autoscaling.Group,
	[]*ec2.LaunchTemplateVersion, error) {
	// get asg info
	asgInfo, err := client.DescribeAutoScalingGroups(&autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{asgName}})
	if err != nil {
		return nil, nil, err
	}
	// get launchTemplateVersion
	ltvInfo, err := client.DescribeLaunchTemplateVersions(&ec2.DescribeLaunchTemplateVersionsInput{
		LaunchTemplateName: ltName, Versions: []*string{ltVersion}})
	if err != nil {
		return nil, nil, err
	}

	return asgInfo, ltvInfo, nil
}

func generateNodeGroupFromAsgAndLtv(group *proto.NodeGroup, asg *autoscaling.Group,
	ltv *ec2.LaunchTemplateVersion) *proto.NodeGroup {
	group = generateNodeGroupFromAsg(group, asg)
	return generateNodeGroupFromLtv(group, ltv)
}

func generateNodeGroupFromAsg(group *proto.NodeGroup, asg *autoscaling.Group) *proto.NodeGroup {
	if asg.AutoScalingGroupName != nil {
		group.AutoScaling.AutoScalingName = *asg.AutoScalingGroupName
		group.AutoScaling.AutoScalingID = *asg.AutoScalingGroupName
	}
	if asg.MaxSize != nil {
		group.AutoScaling.MinSize = uint32(*asg.MaxSize)
	}
	if asg.MinSize != nil {
		group.AutoScaling.MinSize = uint32(*asg.MinSize)
	}
	if asg.DesiredCapacity != nil {
		group.AutoScaling.DesiredSize = uint32(*asg.DesiredCapacity)
	}
	if asg.DefaultCooldown != nil {
		group.AutoScaling.DefaultCooldown = uint32(*asg.DefaultCooldown)
	}
	if asg.VPCZoneIdentifier != nil {
		subnetIDs := strings.Split(*asg.VPCZoneIdentifier, ",")
		group.AutoScaling.SubnetIDs = subnetIDs
	}

	return group
}

func generateNodeGroupFromLtv(group *proto.NodeGroup, ltv *ec2.LaunchTemplateVersion) *proto.NodeGroup {
	if ltv.LaunchTemplateId != nil {
		group.LaunchTemplate.LaunchConfigurationID = *ltv.LaunchTemplateId
	}
	if ltv.LaunchTemplateName != nil {
		group.LaunchTemplate.LaunchConfigureName = *ltv.LaunchTemplateName
	}
	if ltv.LaunchTemplateData != nil {
		group.LaunchTemplate.InstanceType = *ltv.LaunchTemplateData.InstanceType
		if ltv.LaunchTemplateData.SecurityGroupIds != nil {
			group.LaunchTemplate.SecurityGroupIDs = make([]string, 0)
			for _, v := range ltv.LaunchTemplateData.SecurityGroupIds {
				group.LaunchTemplate.SecurityGroupIDs = append(group.LaunchTemplate.SecurityGroupIDs, *v)
			}
		}
		group.LaunchTemplate.ImageInfo = &proto.ImageInfo{ImageID: *ltv.LaunchTemplateData.ImageId}
		group.LaunchTemplate.UserData = *ltv.LaunchTemplateData.UserData
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
