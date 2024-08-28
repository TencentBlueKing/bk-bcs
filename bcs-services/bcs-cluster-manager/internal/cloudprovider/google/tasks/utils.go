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
	container "google.golang.org/api/container/v1"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/google/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
)

// updateNodeGroupCloudNodeGroupID set nodegroup cloudNodeGroupID
func updateNodeGroupCloudNodeGroupID(nodeGroupID string, newGroup *cmproto.NodeGroup) error {
	group, err := cloudprovider.GetStorageModel().GetNodeGroup(context.Background(), nodeGroupID)
	if err != nil {
		return err
	}

	group.CloudNodeGroupID = newGroup.CloudNodeGroupID
	group.Region = newGroup.Region
	err = cloudprovider.GetStorageModel().UpdateNodeGroup(context.Background(), group)
	if err != nil {
		return err
	}

	return nil
}

func checkOperationStatus(computeCli *api.ComputeServiceClient, url, taskID string, d time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return loop.LoopDoFunc(ctx, func() error {
		o, err := api.GetOperation(computeCli, url)
		if err != nil {
			blog.Warnf("Error[%s] while getting operation %s on %s: %v", taskID, o.Name, o.TargetLink, err)
			return nil
		}
		blog.Infof("Operation[%s] [%s] %s status: %s", taskID, url, o.Name, o.Status)
		if o.Status == "DONE" {
			if o.Error != nil {
				errBytes, err := o.Error.MarshalJSON()
				if err != nil {
					errBytes = []byte(fmt.Sprintf("operation failed, but error couldn't be recovered: %v", err))
				}
				return fmt.Errorf("error while getting operation %s on %s: %s", o.Name, o.TargetLink, errBytes)
			}
			return loop.EndLoop
		}
		blog.Infof("taskID[%s] operation %s still running", taskID, o.SelfLink)

		return nil
	}, loop.LoopInterval(d))
}

func checkGKEOperationStatus(containerCli *api.ContainerServiceClient, operation *container.Operation,
	taskID string, d time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	return loop.LoopDoFunc(ctx, func() error {

		o, err := containerCli.GetGKEOperation(context.Background(), operation.Name)
		if err != nil {
			return err
		}
		if o.Status == "DONE" {
			if o.Error != nil {
				return fmt.Errorf("operation error: %d, %v", o.Error.Code, o.Error.Details)
			}
			return loop.EndLoop
		}
		blog.Infof("taskID[%s] operation[%s] %s still running", taskID, o.Status, o.SelfLink)
		return nil
	}, loop.LoopInterval(d))
}

// GenerateCreateNodePoolInput generate create node pool input
func GenerateCreateNodePoolInput(group *proto.NodeGroup, cluster *proto.Cluster) *api.CreateNodePoolRequest {
	if group.NodeTemplate.MaxPodsPerNode == 0 {
		group.NodeTemplate.MaxPodsPerNode = 110
	}
	return &api.CreateNodePoolRequest{
		NodePool: &api.NodePool{
			// gke nodePool名称中不允许有大写字母
			Name:             group.CloudNodeGroupID,
			Config:           generateNodeConfig(group),
			InitialNodeCount: int64(group.AutoScaling.DesiredSize),
			Locations:        group.AutoScaling.Zones,
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

// generateNodeConfig generate node config
func generateNodeConfig(nodeGroup *proto.NodeGroup) *api.NodeConfig {
	if nodeGroup.LaunchTemplate == nil {
		return nil
	}
	template := nodeGroup.LaunchTemplate
	diskSize, _ := strconv.Atoi(template.SystemDisk.DiskSize)
	conf := &api.NodeConfig{
		MachineType: template.InstanceType,
		Labels:      nodeGroup.NodeTemplate.Labels,
		Taints:      api.MapTaints(nodeGroup.NodeTemplate.Taints),
		DiskSizeGb:  int64(diskSize),
		DiskType:    template.SystemDisk.DiskType,
	}
	if template.ImageInfo != nil {
		conf.ImageType = template.ImageInfo.ImageName
	}
	return conf
}

// generateNodeManagement generate node management
func generateNodeManagement(nodeGroup *proto.NodeGroup, cluster *proto.Cluster) *api.NodeManagement {
	if nodeGroup.AutoScaling == nil {
		return nil
	}
	nm := &api.NodeManagement{}
	nm.AutoUpgrade = nodeGroup.AutoScaling.AutoUpgrade
	nm.AutoRepair = nodeGroup.AutoScaling.ReplaceUnhealthy
	if cluster.ExtraInfo != nil {
		if cluster.ExtraInfo[api.GKEClusterReleaseChannel] != "" {
			// when releaseChannel is set, autoUpgrade and autoRepair must be true
			nm.AutoUpgrade = true
			nm.AutoRepair = true
		}
	}
	return nm
}
