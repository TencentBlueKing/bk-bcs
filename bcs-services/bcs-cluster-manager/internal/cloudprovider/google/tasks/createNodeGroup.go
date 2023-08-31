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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/container/v1"
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
		blog.Errorf("CreateCloudNodeGroupTask[%s]: getClusterDependBasicInfo failed: %v", taskID, err)
		retErr := fmt.Errorf("getClusterDependBasicInfo failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	cmOption := dependInfo.CmOption
	cluster := dependInfo.Cluster
	group := dependInfo.NodeGroup

	// create node group
	gkeCli, err := api.NewContainerServiceClient(cmOption)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: get gke client for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud gke client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return err
	}
	// gke nodePool名称中不允许有大写字母
	group.CloudNodeGroupID = strings.ToLower(group.NodeGroupID)

	operation, err := gkeCli.CreateClusterNodePool(context.Background(),
		generateCreateNodePoolInput(group, cluster), cluster.SystemID)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool[%s] api in task %s "+
			"step %s failed, %s, operation ID: %s", taskID, nodeGroupID, taskID, stepName, err.Error(), operation)
		retErr := fmt.Errorf("call CreateClusterNodePool[%s] api err, %s, operation ID: %s",
			nodeGroupID, err.Error(), operation)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool successful", taskID)

	// update nodegorup cloudNodeGroupID
	err = updateNodeGroupCloudNodeGroupID(nodeGroupID, group)
	if err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s]: updateNodeGroupCloudNodeGroupID[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("call CreateCloudNodeGroupTask updateNodeGroupCloudNodeGroupID[%s] api err, %s", nodeGroupID,
			err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CreateCloudNodeGroupTask[%s]: call CreateClusterNodePool updateNodeGroupCloudNodeGroupID successful",
		taskID)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}

	state.Task.CommonParams["CloudNodeGroupID"] = group.CloudNodeGroupID
	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateCloudNodeGroupTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func generateCreateNodePoolInput(group *proto.NodeGroup, cluster *proto.Cluster) *api.CreateNodePoolRequest {
	if group.NodeTemplate.MaxPodsPerNode == 0 {
		group.NodeTemplate.MaxPodsPerNode = 110
	}
	return &api.CreateNodePoolRequest{
		NodePool: &api.NodePool{
			Name:             group.CloudNodeGroupID,
			Config:           generateNodeConfig(group),
			InitialNodeCount: 0,
			MaxPodsConstraint: &api.MaxPodsConstraint{
				MaxPodsPerNode: int64(group.NodeTemplate.MaxPodsPerNode),
			},
			Autoscaling: &api.NodePoolAutoscaling{
				// 不开启谷歌云 CA 组件，因为需要部署 BCS 自己的 CA 组件
				Enabled: false,
			},
			Management: generateNodeManagement(group, cluster),
		},
	}
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
	cmOption := dependInfo.CmOption
	cluster := dependInfo.Cluster
	group := dependInfo.NodeGroup

	// get google cloud client
	client, err := api.NewGCPClientSet(cmOption)
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: get gcp client for nodegroup[%s] in task %s step %s failed, %s",
			taskID, nodeGroupID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud as client err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	containerCli := client.ContainerServiceClient
	computeCli := client.ComputeServiceClient

	// wait node group state to normal
	ctx, cancel := context.WithTimeout(context.TODO(), 20*time.Minute)
	defer cancel()

	cloudNodeGroup := &container.NodePool{}

	err = loop.LoopDoFunc(ctx, func() error {
		np, errPool := containerCli.GetClusterNodePool(context.Background(), cluster.SystemID, group.CloudNodeGroupID)
		if errPool != nil {
			blog.Errorf("taskID[%s] GetClusterNodePool[%s/%s] failed: %v", taskID, cluster.SystemID,
				group.CloudNodeGroupID, errPool)
			return nil
		}
		if np == nil {
			return nil
		}
		cloudNodeGroup = np
		switch {
		case np.Status == api.NodeGroupStatusProvisioning:
			blog.Infof("taskID[%s] GetClusterNodePool[%s] still creating, status[%s]",
				taskID, group.CloudNodeGroupID, np.Status)
			return nil
		case np.Status == api.NodeGroupStatusRunning:
			return loop.EndLoop
		default:
			return nil
		}
	}, loop.LoopInterval(5*time.Second))
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: GetClusterNodePool failed: %v", taskID, err)
		retErr := fmt.Errorf("GetClusterNodePool failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	newIt, igm, err := getIgmAndIt(computeCli, cloudNodeGroup, group, cluster, taskID)
	if err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s]: getIgmAndIt failed: %v", taskID, err)
		retErr := fmt.Errorf("getIgmAndIt failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), generateNodeGroupFromIgmAndIt(group,
		igm, newIt, cmOption))
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
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCloudNodeGroupStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

func getIgmAndIt(computeCli *api.ComputeServiceClient, cloudNodeGroup *container.NodePool, group *proto.NodeGroup,
	cluster *proto.Cluster, taskID string) (*compute.InstanceTemplate, *compute.InstanceGroupManager, error) {
	// get instanceGroupManager
	igm, err := api.GetInstanceGroupManager(computeCli, cloudNodeGroup.InstanceGroupUrls[0])
	if err != nil {
		blog.Errorf("taskID[%s] GetInstanceGroupManager failed: %v", taskID, err)
		return nil, nil, err
	}

	// get instanceTemplate info
	it, err := api.GetInstanceTemplate(computeCli, igm.InstanceTemplate)
	if err != nil {
		blog.Errorf("taskID[%s] GetInstanceGroupManager failed: %v", taskID, err)
		return nil, nil, err
	}
	newIt := it

	err = newItFromBaseIt(newIt, it, group, cluster, computeCli, taskID)
	if err != nil {
		return nil, nil, err
	}

	err = patchIgm(newIt, igm, computeCli, group, taskID)
	if err != nil {
		return nil, nil, err
	}

	if newIt.Name != it.Name {
		// 如果使用了新模版,则删除旧模版
		_, err = computeCli.DeleteInstanceTemplate(context.Background(), it.Name)
		if err != nil {
			return nil, nil, err
		}
	}

	return newIt, igm, nil
}

func patchIgm(newIt *compute.InstanceTemplate, igm *compute.InstanceGroupManager, computeCli *api.ComputeServiceClient,
	group *proto.NodeGroup, taskID string) error {
	newIgm := &compute.InstanceGroupManager{
		InstanceTemplate: newIt.SelfLink,
		BaseInstanceName: newIt.Name,
		UpdatePolicy:     api.GenerateUpdatePolicy(group),
	}
	o, err := api.PatchInstanceGroupManager(computeCli, igm.SelfLink, newIgm)
	if err != nil {
		blog.Errorf("taskID[%s] GetInstanceGroupManager failed: %v, operation ID: %s", taskID, err, o.SelfLink)
		return err
	}
	// 检查操作是否成功
	err = checkOperationStatus(computeCli, o.SelfLink, taskID, 3*time.Second)
	if err != nil {
		return fmt.Errorf("CheckCloudNodeGroupStatusTask[%s] GetOperation failed: %v", taskID, err)
	}

	return nil
}

func newItFromBaseIt(newIt, it *compute.InstanceTemplate, group *proto.NodeGroup, cluster *proto.Cluster,
	computeCli *api.ComputeServiceClient, taskID string) error {
	if len(group.LaunchTemplate.DataDisks) != 0 {
		newIt.Name = strings.Join([]string{"gke", cluster.SystemID, group.CloudNodeGroupID, utils.RandomHexString(8)}, "-")
		dataDisks := make([]*compute.AttachedDisk, 0)
		bootDisk := it.Properties.Disks[0]
		for _, d := range group.LaunchTemplate.DataDisks {
			diskSize, _ := strconv.Atoi(d.DiskSize)
			dataDisks = append(dataDisks, &compute.AttachedDisk{
				Type:             d.DiskType,
				DiskSizeGb:       int64(diskSize),
				Mode:             "READ_WRITE",
				Boot:             false,
				AutoDelete:       true,
				InitializeParams: bootDisk.InitializeParams,
			})
		}
		newIt.Properties.Disks = append(newIt.Properties.Disks, dataDisks...)
		o, err := computeCli.CreateInstanceTemplate(context.Background(), newIt)
		if err != nil {
			blog.Errorf("taskID[%s] CreateInstanceTemplate failed: %v, operation ID: %s", taskID, err, o.SelfLink)
			return err
		}
		// 检查实例模版是否创建成功
		err = checkOperationStatus(computeCli, o.SelfLink, taskID, 3*time.Second)
		if err != nil {
			return fmt.Errorf("CheckCloudNodeGroupStatusTask[%s] GetOperation failed: %v", taskID, err)
		}
	}

	return nil
}

func generateNodeGroupFromIgmAndIt(group *proto.NodeGroup, igm *compute.InstanceGroupManager,
	it *compute.InstanceTemplate, opt *cloudprovider.CommonOption) *proto.NodeGroup {
	group = generateNodeGroupFromIgm(group, igm)
	return generateNodeGroupFromIt(group, it, opt)
}

func generateNodeGroupFromIgm(group *proto.NodeGroup, igm *compute.InstanceGroupManager) *proto.NodeGroup {
	group.AutoScaling.AutoScalingID = igm.SelfLink
	group.AutoScaling.AutoScalingName = igm.Name
	group.AutoScaling.DesiredSize = uint32(igm.TargetSize)
	return group
}

func generateNodeGroupFromIt(group *proto.NodeGroup, it *compute.InstanceTemplate,
	opt *cloudprovider.CommonOption) *proto.NodeGroup {
	group.LaunchTemplate.LaunchConfigurationID = it.SelfLink
	group.LaunchTemplate.LaunchConfigureName = it.Name
	group.LaunchTemplate.ProjectID = opt.Account.GkeProjectID
	if it.Properties != nil {
		prop := it.Properties
		group.LaunchTemplate.InstanceType = prop.MachineType
		if prop.NetworkInterfaces != nil {
			if group.AutoScaling == nil {
				group.AutoScaling = &proto.AutoScalingGroup{}
			}
			group.AutoScaling.VpcID = prop.NetworkInterfaces[0].Network
			group.AutoScaling.SubnetIDs = append(group.AutoScaling.SubnetIDs, prop.NetworkInterfaces[0].Subnetwork)
		}
		if prop.Disks != nil {
			group.LaunchTemplate.ImageInfo = &proto.ImageInfo{
				ImageName: group.NodeOS,
			}
			group.LaunchTemplate.SystemDisk = &proto.DataDisk{
				DiskType: prop.Disks[0].InitializeParams.DiskType,
				DiskSize: strconv.FormatInt(prop.Disks[0].InitializeParams.DiskSizeGb, 10),
			}
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
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateCreateNodeGroupDBInfoTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

func generateNodeConfig(nodeGroup *proto.NodeGroup) *api.NodeConfig {
	if nodeGroup.LaunchTemplate == nil {
		return nil
	}
	template := nodeGroup.LaunchTemplate
	diskSize, _ := strconv.Atoi(template.SystemDisk.DiskSize)
	conf := &api.NodeConfig{
		MachineType: template.InstanceType,
		Labels:      nodeGroup.Labels,
		Taints:      api.MapTaints(nodeGroup.NodeTemplate.Taints),
		DiskSizeGb:  int64(diskSize),
		DiskType:    template.SystemDisk.DiskType,
	}
	if template.ImageInfo != nil {
		conf.ImageType = template.ImageInfo.ImageName
	}
	return conf
}

func generateNodeManagement(nodeGroup *proto.NodeGroup, cluster *proto.Cluster) *api.NodeManagement {
	if nodeGroup.AutoScaling == nil {
		return nil
	}
	nm := &api.NodeManagement{}
	nm.AutoUpgrade = nodeGroup.AutoScaling.AutoUpgrade
	nm.AutoRepair = nodeGroup.AutoScaling.ReplaceUnhealthy
	if cluster.ExtraInfo != nil {
		if cluster.ExtraInfo["releaseChannel"] != "" {
			// when releaseChannel is set, autoUpgrade and autoRepair must be true
			nm.AutoUpgrade = true
			nm.AutoRepair = true
		}
	}
	return nm
}
