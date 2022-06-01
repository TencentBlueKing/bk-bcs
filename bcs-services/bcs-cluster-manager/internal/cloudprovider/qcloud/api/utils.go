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

package api

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
	as "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/as/v20180419"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

const (
	limit = 100

	maxFilterValues = 5
)

// FilterKey tke interface filterKey
type FilterKey string

// String xxx
func (f FilterKey) String() string {
	return string(f)
}

var (
	// Zone zone
	Zone FilterKey = "zone"
	// InstanceFamily instance-family
	InstanceFamily FilterKey = "instance-family"
)

func generateInstanceAdvancedSetting(advancedSetting *InstanceAdvancedSettings) *tke.InstanceAdvancedSettings {
	if advancedSetting != nil {
		advancedSet := &tke.InstanceAdvancedSettings{
			MountTarget: func() *string {
				if len(advancedSetting.MountTarget) == 0 {
					return common.StringPtr(MountTarget)
				}
				return common.StringPtr(advancedSetting.MountTarget)
			}(),
			DockerGraphPath: func() *string {
				if len(advancedSetting.DockerGraphPath) == 0 {
					return common.StringPtr(DockerGraphPath)
				}
				return common.StringPtr(advancedSetting.DockerGraphPath)
			}(),
			Unschedulable: func() *int64 {
				if advancedSetting.Unschedulable != nil {
					return advancedSetting.Unschedulable
				}

				return common.Int64Ptr(1)
			}(),
			Labels: func() []*tke.Label {
				if len(advancedSetting.Labels) == 0 {
					return nil
				}

				labels := make([]*tke.Label, 0)
				for i := range advancedSetting.Labels {
					labels = append(labels, &tke.Label{
						Name:  common.StringPtr(advancedSetting.Labels[i].Name),
						Value: common.StringPtr(advancedSetting.Labels[i].Value),
					})
				}
				return labels
			}(),
			DataDisks: func() []*tke.DataDisk {
				if len(advancedSetting.DataDisks) == 0 {
					return nil
				}

				dataDisks := make([]*tke.DataDisk, 0)
				for i := range advancedSetting.DataDisks {
					dataDisks = append(dataDisks, &tke.DataDisk{
						DiskType:    common.StringPtr(advancedSetting.DataDisks[i].DiskType),
						DiskSize:    common.Int64Ptr(advancedSetting.DataDisks[i].DiskSize),
						MountTarget: common.StringPtr(advancedSetting.DataDisks[i].MountTarget),
					})
				}
				return dataDisks
			}(),
			ExtraArgs: func() *tke.InstanceExtraArgs {
				if advancedSetting.ExtraArgs != nil && len(advancedSetting.ExtraArgs.Kubelet) > 0 {
					return &tke.InstanceExtraArgs{Kubelet: common.StringPtrs(advancedSetting.ExtraArgs.Kubelet)}
				}
				return nil
			}(),
			UserScript: func() *string {
				if len(advancedSetting.UserScript) > 0 {
					return common.StringPtr(advancedSetting.UserScript)
				}

				return nil
			}(),
		}

		return advancedSet
	}

	return nil
}

func generateAddExistedInstancesReq(addReq *AddExistedInstanceReq) *tke.AddExistedInstancesRequest {
	// add existed instance to cluster request
	req := tke.NewAddExistedInstancesRequest()
	req.ClusterId = common.StringPtr(addReq.ClusterID)
	req.InstanceIds = common.StringPtrs(addReq.InstanceIDs)

	if addReq.LoginSetting != nil {
		req.LoginSettings = &tke.LoginSettings{
			Password: common.StringPtr(addReq.LoginSetting.Password),
		}
	}

	if addReq.EnhancedSetting != nil {
		req.EnhancedService = &tke.EnhancedService{
			SecurityService: &tke.RunSecurityServiceEnabled{
				Enabled: common.BoolPtr(addReq.EnhancedSetting.SecurityService),
			},
			MonitorService: &tke.RunMonitorServiceEnabled{
				Enabled: common.BoolPtr(addReq.EnhancedSetting.MonitorService),
			},
		}
	}

	if len(addReq.SecurityGroupIds) > 0 {
		req.SecurityGroupIds = common.StringPtrs(addReq.SecurityGroupIds)
	}

	if addReq.NodePool != nil {
		req.NodePool = &tke.NodePoolOption{
			AddToNodePool:                    common.BoolPtr(addReq.NodePool.AddToNodePool),
			NodePoolId:                       common.StringPtr(addReq.NodePool.NodePoolID),
			InheritConfigurationFromNodePool: common.BoolPtr(addReq.NodePool.InheritConfigurationFromNodePool),
		}
	}

	if addReq.AdvancedSetting != nil {
		req.InstanceAdvancedSettings = generateInstanceAdvancedSetting(addReq.AdvancedSetting)
	}

	return req
}

func generateClusterRequestInfo(request *CreateClusterRequest) (*tke.CreateClusterRequest, error) {
	if request.Region == "" || request.ClusterType == "" {
		return nil, fmt.Errorf("CreateClusterRequest invalid region or clusterType info")
	}

	if request.ClusterCIDR == nil {
		return nil, fmt.Errorf("CreateClusterRequest ClusterCIDR info is null")
	}

	req := tke.NewCreateClusterRequest()
	req.ClusterType = common.StringPtr(request.ClusterType)

	// cluster CIDR
	req.ClusterCIDRSettings = &tke.ClusterCIDRSettings{
		ClusterCIDR:          common.StringPtr(request.ClusterCIDR.ClusterCIDR),
		MaxNodePodNum:        common.Uint64Ptr(request.ClusterCIDR.MaxNodePodNum),
		MaxClusterServiceNum: common.Uint64Ptr(request.ClusterCIDR.MaxClusterServiceNum),
	}
	if request.ClusterCIDR.ServiceCIDR != "" {
		req.ClusterCIDRSettings.ServiceCIDR = common.StringPtr(request.ClusterCIDR.ServiceCIDR)
	}

	// cluster Basic config
	if request.ClusterBasic != nil {
		req.ClusterBasicSettings = generateClusterBasic(request.ClusterBasic)
	}

	// cluster advanced config info
	if request.ClusterAdvanced != nil {
		req.ClusterAdvancedSettings = generateClusterAdvancedSet(request.ClusterAdvanced)
	}

	// cluster instance config info
	if request.InstanceAdvanced != nil {
		req.InstanceAdvancedSettings = generateInstanceAdvancedSetting(request.InstanceAdvanced)
	}

	// runInstances mode
	if request.AddNodeMode {
		if len(request.RunInstancesForNode) == 0 {
			return nil, fmt.Errorf("CreateClusterRequest RunInstancesForNode is null")
		}

		for i := range request.RunInstancesForNode {
			if len(request.RunInstancesForNode[i].RunInstancesPara) == 0 {
				return nil, fmt.Errorf("CreateClusterRequest RunInstancesPara is null")
			}
			req.RunInstancesForNode = append(req.RunInstancesForNode, &tke.RunInstancesForNode{
				NodeRole:         common.StringPtr(request.RunInstancesForNode[i].NodeRole),
				RunInstancesPara: request.RunInstancesForNode[i].RunInstancesPara,
			})
		}

		return req, nil
	}

	// existed instance to cluster
	if len(request.ExistedInstancesForNode) == 0 {
		return nil, fmt.Errorf("CreateClusterRequest ExistedInstancesForNode is null")
	}

	for i := range request.ExistedInstancesForNode {
		if len(request.ExistedInstancesForNode[i].ExistedInstancesPara.InstanceIDs) == 0 {
			return nil, fmt.Errorf("CreateClusterRequest ExistedInstancesForNode instance is null")
		}

		req.ExistedInstancesForNode = append(req.ExistedInstancesForNode, &tke.ExistedInstancesForNode{
			NodeRole: common.StringPtr(request.ExistedInstancesForNode[i].NodeRole),
			ExistedInstancesPara: &tke.ExistedInstancesPara{
				InstanceIds: common.StringPtrs(request.ExistedInstancesForNode[i].ExistedInstancesPara.InstanceIDs),
				//InstanceAdvancedSettings: generateInstanceAdvancedSet(request.InstanceAdvanced),
				//EnhancedService:          generateEnhancedService(request.ExistedInstancesForNode[i].ExistedInstancesPara.EnhancedService),
				LoginSettings: generateLoginSet(request.ExistedInstancesForNode[i].ExistedInstancesPara.LoginSettings),
				//SecurityGroupIds:         request.ExistedInstancesForNode[i].ExistedInstancesPara.SecurityGroupIds,
			},
		})
	}

	return req, nil
}

func generateClusterBasic(basic *ClusterBasicSettings) *tke.ClusterBasicSettings {
	tkeClusterBasic := &tke.ClusterBasicSettings{
		ClusterOs:      common.StringPtr(basic.ClusterOS),
		ClusterVersion: common.StringPtr(basic.ClusterVersion),
		ClusterName:    common.StringPtr(basic.ClusterName),
		VpcId:          common.StringPtr(basic.VpcID),
	}

	tags := make([]*tke.TagSpecification, 0)
	for i := range basic.TagSpecification {
		tags = append(tags, &tke.TagSpecification{
			ResourceType: common.StringPtr(basic.TagSpecification[i].ResourceType),
			Tags: func() []*tke.Tag {
				tag := make([]*tke.Tag, 0)
				for _, t := range basic.TagSpecification[i].Tags {
					tag = append(tag, &tke.Tag{
						Key:   t.Key,
						Value: t.Value,
					})
				}

				return tag
			}(),
		})
	}
	if len(tags) > 0 {
		tkeClusterBasic.TagSpecification = tags
	}

	return tkeClusterBasic
}

func generateClusterAdvancedSet(request *ClusterAdvancedSettings) *tke.ClusterAdvancedSettings {
	if request == nil {
		return nil
	}

	clusterAdvance := &tke.ClusterAdvancedSettings{
		IPVS:             common.BoolPtr(request.IPVS),
		ContainerRuntime: common.StringPtr(request.ContainerRuntime),
		RuntimeVersion:   common.StringPtr(request.RuntimeVersion),
	}

	if clusterAdvance.ExtraArgs == nil {
		clusterAdvance.ExtraArgs = &tke.ClusterExtraArgs{}
	}

	if len(request.ExtraArgs.KubeAPIServer) > 0 {
		clusterAdvance.ExtraArgs.KubeAPIServer = request.ExtraArgs.KubeAPIServer
	}
	if len(request.ExtraArgs.KubeControllerManager) > 0 {
		clusterAdvance.ExtraArgs.KubeControllerManager = request.ExtraArgs.KubeControllerManager
	}
	if len(request.ExtraArgs.KubeScheduler) > 0 {
		clusterAdvance.ExtraArgs.KubeScheduler = request.ExtraArgs.KubeScheduler
	}
	if len(request.ExtraArgs.Etcd) > 0 {
		clusterAdvance.ExtraArgs.Etcd = request.ExtraArgs.Etcd
	}

	return clusterAdvance
}

// transfer input para to tke para
func generateInstanceAdvancedSet(request *InstanceAdvancedSettings) *tke.InstanceAdvancedSettings {
	if request == nil {
		return nil
	}

	instanceAdvance := &tke.InstanceAdvancedSettings{
		MountTarget:     common.StringPtr(request.MountTarget),
		DockerGraphPath: common.StringPtr(request.DockerGraphPath),
		Unschedulable:   request.Unschedulable,
		Labels: func() []*tke.Label {
			if len(request.Labels) == 0 {
				return nil
			}

			labels := make([]*tke.Label, 0)
			for i := range request.Labels {
				labels = append(labels, &tke.Label{
					Name:  common.StringPtr(request.Labels[i].Name),
					Value: common.StringPtr(request.Labels[i].Value),
				})
			}
			return labels
		}(),
		DataDisks: func() []*tke.DataDisk {
			if len(request.DataDisks) == 0 {
				return nil
			}

			disks := make([]*tke.DataDisk, 0)
			for i := range request.DataDisks {
				disks = append(disks, &tke.DataDisk{
					DiskType:           common.StringPtr(request.DataDisks[i].DiskType),
					FileSystem:         common.StringPtr(request.DataDisks[i].FileSystem),
					DiskSize:           common.Int64Ptr(request.DataDisks[i].DiskSize),
					AutoFormatAndMount: common.BoolPtr(request.DataDisks[i].AutoFormatAndMount),
					MountTarget:        common.StringPtr(request.DataDisks[i].MountTarget),
				})
			}
			return disks
		}(),
		ExtraArgs: func() *tke.InstanceExtraArgs {
			if request.ExtraArgs == nil || len(request.ExtraArgs.Kubelet) == 0 {
				return nil
			}

			return &tke.InstanceExtraArgs{Kubelet: common.StringPtrs(request.ExtraArgs.Kubelet)}
		}(),
	}

	return instanceAdvance
}

func generateEnhancedService(service *EnhancedService) *tke.EnhancedService {
	if service == nil {
		return nil
	}

	return &tke.EnhancedService{
		SecurityService: &tke.RunSecurityServiceEnabled{Enabled: common.BoolPtr(service.SecurityService)},
		MonitorService:  &tke.RunMonitorServiceEnabled{Enabled: common.BoolPtr(service.MonitorService)},
	}
}

func generateLoginSet(settings *LoginSettings) *tke.LoginSettings {
	if settings == nil {
		return nil
	}

	return &tke.LoginSettings{
		Password: common.StringPtr(settings.Password),
	}
}

func generateClusterNodePool(nodePool *CreateNodePoolInput) *tke.CreateClusterNodePoolRequest {
	if nodePool == nil {
		return nil
	}
	req := tke.NewCreateClusterNodePoolRequest()
	req.ClusterId = nodePool.ClusterID
	asg := utils.ToJSONString(nodePool.AutoScalingGroupPara)
	req.AutoScalingGroupPara = &asg
	lunchConf := utils.ToJSONString(nodePool.LaunchConfigurePara)
	req.LaunchConfigurePara = &lunchConf
	req.InstanceAdvancedSettings = generateInstanceAdvancedSet(nodePool.InstanceAdvancedSettings)
	req.EnableAutoscale = nodePool.EnableAutoscale
	req.Name = nodePool.Name
	req.Labels = generateLabel(nodePool.Labels)
	req.Taints = generateTaint(nodePool.Taints)
	req.NodePoolOs = nodePool.NodePoolOs
	req.OsCustomizeType = nodePool.OsCustomizeType
	req.Tags = generateTag(nodePool.Tags)
	return req
}

func generateLabel(labels []*Label) []*tke.Label {
	result := make([]*tke.Label, 0)
	for _, v := range labels {
		result = append(result, &tke.Label{Name: v.Name, Value: v.Value})
	}
	return result
}

func generateTaint(taints []*Taint) []*tke.Taint {
	result := make([]*tke.Taint, 0)
	for _, v := range taints {
		result = append(result, &tke.Taint{Key: v.Key, Value: v.Value, Effect: v.Effect})
	}
	return result
}

func generateTag(tags []*Tag) []*tke.Tag {
	result := make([]*tke.Tag, 0)
	for _, v := range tags {
		result = append(result, &tke.Tag{Key: v.Key, Value: v.Value})
	}
	return result
}

func convertSubnet(subnet []*vpc.Subnet) []*Subnet {
	result := make([]*Subnet, 0)
	for _, v := range subnet {
		result = append(result, &Subnet{
			VpcID:                   v.VpcId,
			SubnetID:                v.SubnetId,
			SubnetName:              v.SubnetName,
			CidrBlock:               v.CidrBlock,
			IsDefault:               v.IsDefault,
			EnableBroadcast:         v.EnableBroadcast,
			Zone:                    v.Zone,
			RouteTableID:            v.RouteTableId,
			CreatedTime:             v.CreatedTime,
			AvailableIPAddressCount: v.AvailableIpAddressCount,
			Ipv6CidrBlock:           v.Ipv6CidrBlock,
			NetworkACLID:            v.NetworkAclId,
			IsRemoteVpcSnat:         v.IsRemoteVpcSnat,
			TotalIPAddressCount:     v.TotalIpAddressCount,
			CdcID:                   v.CdcId,
			IsCdcSubnet:             v.IsCdcSubnet,
		})
	}
	return result
}

func convertModifyNodePool(nodePool *ModifyClusterNodePoolInput) *tke.ModifyClusterNodePoolRequest {
	req := tke.NewModifyClusterNodePoolRequest()
	req.ClusterId = nodePool.ClusterID
	req.NodePoolId = nodePool.NodePoolID
	req.Name = nodePool.Name
	req.MaxNodesNum = nodePool.MaxNodesNum
	req.MinNodesNum = nodePool.MinNodesNum
	req.Labels = generateLabel(nodePool.Labels)
	req.Taints = generateTaint(nodePool.Taints)
	req.EnableAutoscale = nodePool.EnableAutoscale
	req.OsName = nodePool.OsName
	req.OsCustomizeType = nodePool.OsCustomizeType
	req.ExtraArgs = func(args *InstanceExtraArgs) *tke.InstanceExtraArgs {
		if args == nil {
			return nil
		}
		return &tke.InstanceExtraArgs{Kubelet: common.StringPtrs(args.Kubelet)}
	}(nodePool.ExtraArgs)
	req.Tags = generateTag(nodePool.Tags)
	req.Unschedulable = nodePool.Unschedulable
	return req
}

// MapToLabels converts a map of string-string to a slice of Label
func MapToLabels(labels map[string]string) []*Label {
	result := make([]*Label, 0)
	for k, v := range labels {
		name := k
		value := v
		result = append(result, &Label{Name: &name, Value: &value})
	}
	return result
}

// MapToTaints converts a map of string-string to a slice of Taint
func MapToTaints(taints map[string]string) []*Taint {
	result := make([]*Taint, 0)
	for k, v := range taints {
		key := k
		value := v
		result = append(result, &Taint{Key: &key, Value: &value})
	}
	return result
}

func convertASGInstance(ins *as.Instance) *AutoScalingInstances {
	return &AutoScalingInstances{
		InstanceID:              ins.InstanceId,
		AutoScalingGroupID:      ins.AutoScalingGroupId,
		LaunchConfigurationID:   ins.LaunchConfigurationId,
		LaunchConfigurationName: ins.LaunchConfigurationName,
		LifeCycleState:          ins.LifeCycleState,
		HealthStatus:            ins.HealthStatus,
		ProtectedFromScaleIn:    ins.ProtectedFromScaleIn,
		Zone:                    ins.Zone,
		CreationType:            ins.CreationType,
		AddTime:                 ins.AddTime,
		InstanceType:            ins.InstanceType,
		VersionNumber:           ins.VersionNumber,
		AutoScalingGroupName:    ins.AutoScalingGroupName,
	}
}
