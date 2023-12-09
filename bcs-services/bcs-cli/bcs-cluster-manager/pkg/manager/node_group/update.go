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

package nodegroup

import (
	"errors"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cli/bcs-cluster-manager/pkg/manager/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
)

// Update 更新节点池,更新成功返回nil
func (c *NodeGroupMgr) Update(req types.UpdateNodeGroupReq) error { // nolint
	timeRange := make([]*clustermanager.TimeRange, 0)

	// nodeGroup time range
	for _, v := range req.AutoScaling.TimeRanges {
		timeRange = append(timeRange, &clustermanager.TimeRange{
			Name:       v.Name,
			Schedule:   v.Schedule,
			Zone:       v.Zone,
			DesiredNum: v.DesiredNum,
		})
	}

	// nodeGroup node disk mount
	launchDataDisk := make([]*clustermanager.DataDisk, 0)
	for _, v := range req.LaunchTemplate.DataDisks {
		launchDataDisk = append(launchDataDisk, &clustermanager.DataDisk{
			DiskType: v.DiskType,
			DiskSize: v.DiskSize,
		})
	}

	// nodeGroup taint
	taint := make([]*clustermanager.Taint, 0)
	for _, v := range req.NodeTemplate.Taints {
		taint = append(taint, &clustermanager.Taint{
			Key:    v.Key,
			Value:  v.Value,
			Effect: v.Effect,
		})
	}

	// nodeGroup node dataDisks
	nodeDataDisk := make([]*clustermanager.DataDisk, 0)
	for _, v := range req.NodeTemplate.DataDisks {
		nodeDataDisk = append(nodeDataDisk, &clustermanager.DataDisk{
			DiskType: v.DiskType,
			DiskSize: v.DiskSize,
		})
	}

	// nodeGroup plugins
	outPlugins := make(map[string]*clustermanager.BKOpsPlugin)
	for k, v := range req.NodeTemplate.BcsScaleOutAddons.Plugins {
		outPlugins[k] = &clustermanager.BKOpsPlugin{
			System: v.System,
			Link:   v.Link,
			Params: v.Params,
		}
	}

	inPlugins := make(map[string]*clustermanager.BKOpsPlugin)
	for k, v := range req.NodeTemplate.BcsScaleInAddons.Plugins {
		inPlugins[k] = &clustermanager.BKOpsPlugin{
			System: v.System,
			Link:   v.Link,
			Params: v.Params,
		}
	}

	outExtraPlugins := make(map[string]*clustermanager.BKOpsPlugin)
	for k, v := range req.NodeTemplate.ScaleOutExtraAddons.Plugins {
		outExtraPlugins[k] = &clustermanager.BKOpsPlugin{
			System: v.System,
			Link:   v.Link,
			Params: v.Params,
		}
	}

	inExtraPlugins := make(map[string]*clustermanager.BKOpsPlugin)
	for k, v := range req.NodeTemplate.ScaleInExtraAddons.Plugins {
		inExtraPlugins[k] = &clustermanager.BKOpsPlugin{
			System: v.System,
			Link:   v.Link,
			Params: v.Params,
		}
	}

	// 构建UpdateNodeGroupRequest请求并更新节点池
	resp, err := c.client.UpdateNodeGroup(c.ctx, &clustermanager.UpdateNodeGroupRequest{
		NodeGroupID: req.NodeGroupID,
		ClusterID:   req.ClusterID,
		Name:        req.Name,
		Region:      req.Region,
		//EnableAutoscale: req.EnableAutoscale,
		AutoScaling: &clustermanager.AutoScalingGroup{
			AutoScalingID:         req.AutoScaling.AutoScalingID,
			AutoScalingName:       req.AutoScaling.AutoScalingName,
			MinSize:               req.AutoScaling.MinSize,
			MaxSize:               req.AutoScaling.MaxSize,
			DesiredSize:           req.AutoScaling.DesiredSize,
			VpcID:                 req.AutoScaling.VpcID,
			DefaultCooldown:       req.AutoScaling.DefaultCooldown,
			SubnetIDs:             req.AutoScaling.SubnetIDs,
			Zones:                 req.AutoScaling.Zones,
			RetryPolicy:           req.AutoScaling.RetryPolicy,
			MultiZoneSubnetPolicy: req.AutoScaling.MultiZoneSubnetPolicy,
			ReplaceUnhealthy:      req.AutoScaling.ReplaceUnhealthy,
			ScalingMode:           req.AutoScaling.ScalingMode,
			TimeRanges:            timeRange,
		},
		LaunchTemplate: &clustermanager.LaunchConfiguration{
			LaunchConfigurationID: req.LaunchTemplate.LaunchConfigurationID,
			LaunchConfigureName:   req.LaunchTemplate.LaunchConfigureName,
			ProjectID:             req.LaunchTemplate.ProjectID,
			CPU:                   req.LaunchTemplate.CPU,
			Mem:                   req.LaunchTemplate.Mem,
			GPU:                   req.LaunchTemplate.GPU,
			InstanceType:          req.LaunchTemplate.InstanceType,
			InstanceChargeType:    req.LaunchTemplate.InstanceChargeType,
			SystemDisk: &clustermanager.DataDisk{
				DiskType: req.LaunchTemplate.SystemDisk.DiskType,
				DiskSize: req.LaunchTemplate.SystemDisk.DiskSize,
			},
			DataDisks: launchDataDisk,
			InternetAccess: &clustermanager.InternetAccessible{
				InternetChargeType:   req.LaunchTemplate.InternetAccess.InternetChargeType,
				InternetMaxBandwidth: req.LaunchTemplate.InternetAccess.InternetMaxBandwidth,
				PublicIPAssigned:     req.LaunchTemplate.InternetAccess.PublicIPAssigned,
			},
			InitLoginPassword: req.LaunchTemplate.InitLoginPassword,
			SecurityGroupIDs:  req.LaunchTemplate.SecurityGroupIDs,
			ImageInfo: &clustermanager.ImageInfo{
				ImageID:   req.LaunchTemplate.ImageInfo.ImageID,
				ImageName: req.LaunchTemplate.ImageInfo.ImageName,
			},
			IsSecurityService: req.LaunchTemplate.IsSecurityService,
			IsMonitorService:  req.LaunchTemplate.IsMonitorService,
			UserData:          req.LaunchTemplate.UserData,
		},
		NodeTemplate: &clustermanager.NodeTemplate{
			NodeTemplateID:  req.NodeTemplate.NodeTemplateID,
			Name:            req.NodeTemplate.Name,
			ProjectID:       req.NodeTemplate.ProjectID,
			Labels:          req.NodeTemplate.Labels,
			Taints:          taint,
			DockerGraphPath: req.NodeTemplate.DockerGraphPath,
			MountTarget:     req.NodeTemplate.MountTarget,
			UserScript:      req.NodeTemplate.UserScript,
			UnSchedulable:   req.NodeTemplate.UnSchedulable,
			//DataDisks:          nodeDataDisk,
			ExtraArgs:          req.NodeTemplate.ExtraArgs,
			PreStartUserScript: req.NodeTemplate.PreStartUserScript,
			BcsScaleOutAddons: &clustermanager.Action{
				PreActions:  req.NodeTemplate.BcsScaleOutAddons.PreActions,
				PostActions: req.NodeTemplate.BcsScaleOutAddons.PostActions,
				Plugins:     outPlugins,
			},
			BcsScaleInAddons: &clustermanager.Action{
				PreActions:  req.NodeTemplate.BcsScaleInAddons.PreActions,
				PostActions: req.NodeTemplate.BcsScaleInAddons.PostActions,
				Plugins:     inPlugins,
			},
			ScaleOutExtraAddons: &clustermanager.Action{
				PreActions:  req.NodeTemplate.ScaleOutExtraAddons.PreActions,
				PostActions: req.NodeTemplate.ScaleOutExtraAddons.PostActions,
				Plugins:     outExtraPlugins,
			},
			ScaleInExtraAddons: &clustermanager.Action{
				PreActions:  req.NodeTemplate.ScaleInExtraAddons.PreActions,
				PostActions: req.NodeTemplate.ScaleInExtraAddons.PostActions,
				Plugins:     inExtraPlugins,
			},
			NodeOS: req.NodeTemplate.NodeOS,
			//ModuleID:   req.NodeTemplate.ModuleID,
			Creator:    req.NodeTemplate.Creator,
			Updater:    req.NodeTemplate.Updater,
			CreateTime: req.NodeTemplate.CreateTime,
			UpdateTime: req.NodeTemplate.UpdateTime,
			Desc:       req.NodeTemplate.Desc,
			Runtime: &clustermanager.RunTimeInfo{
				ContainerRuntime: req.NodeTemplate.Runtime.ContainerRuntime,
				RuntimeVersion:   req.NodeTemplate.Runtime.RuntimeVersion,
			},
			Module: &clustermanager.ModuleInfo{
				ScaleOutModuleID: req.NodeTemplate.Module.ScaleOutModuleID,
				ScaleInModuleID:  req.NodeTemplate.Module.ScaleInModuleID,
			},
		},
		Labels:  req.Labels,
		Taints:  req.Taints,
		Tags:    req.Tags,
		NodeOS:  req.NodeOS,
		Updater: req.Updater,
	})

	if err != nil {
		return err
	}

	if resp != nil && resp.Code != 0 {
		return errors.New(resp.Message)
	}

	return nil
}
