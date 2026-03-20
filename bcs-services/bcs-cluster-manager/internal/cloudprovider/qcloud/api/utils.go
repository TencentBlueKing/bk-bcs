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

package api

import (
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	gocache "github.com/patrickmn/go-cache"
	as "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/as/v20180419"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	tke "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tke/v20180525"
	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/cache"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

const (
	limit = 100
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

var (
	// ErrClusterNotFound cluster not found when delete cluster
	ErrClusterNotFound = errors.New(tke.FAILEDOPERATION_CLUSTERNOTFOUND)
)

// generateInstanceAdvancedSetting transfer InstanceAdvancedSettings to cloudInstanceAdvancedSettings
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
			UserScript: func() *string {
				if len(advancedSetting.UserScript) > 0 {
					return common.StringPtr(advancedSetting.UserScript)
				}

				return nil
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
				// 多盘数据盘挂载信息：需要设置购买多个数据盘; 添加已有节点时, 请确保填写的分区信息在节点上真实存在
				dataDisks := make([]*tke.DataDisk, 0)
				for i := range advancedSetting.DataDisks {
					dataDisks = append(dataDisks, &tke.DataDisk{
						DiskType:           common.StringPtr(advancedSetting.DataDisks[i].DiskType),
						DiskSize:           common.Int64Ptr(advancedSetting.DataDisks[i].DiskSize),
						FileSystem:         common.StringPtr(advancedSetting.DataDisks[i].FileSystem),
						MountTarget:        common.StringPtr(advancedSetting.DataDisks[i].MountTarget),
						DiskPartition:      common.StringPtr(advancedSetting.DataDisks[i].DiskPartition),
						AutoFormatAndMount: common.BoolPtr(advancedSetting.DataDisks[i].AutoFormatAndMount),
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
			Taints: func() []*tke.Taint {
				if len(advancedSetting.TaintList) > 0 {
					return generateTaint(advancedSetting.TaintList)
				}

				return nil
			}(),
			// base64 编码的用户脚本，在初始化节点之前执行，目前只对添加已有节点生效
			PreStartUserScript: func() *string {
				if len(advancedSetting.PreStartUserScript) > 0 {
					return common.StringPtr(advancedSetting.PreStartUserScript)
				}
				return nil
			}(),
			GPUArgs: func() *tke.GPUArgs {
				if advancedSetting != nil && advancedSetting.GPUArgs != nil {
					gpuArgs := &tke.GPUArgs{
						MIGEnable: common.BoolPtr(advancedSetting.GPUArgs.MIGEnable),
					}

					if advancedSetting.GPUArgs.Driver != nil {
						gpuArgs.Driver = &tke.DriverVersion{
							Version: common.StringPtr(advancedSetting.GPUArgs.Driver.Version),
							Name:    common.StringPtr(advancedSetting.GPUArgs.Driver.Name),
						}
					}

					if advancedSetting.GPUArgs.CUDA != nil {
						gpuArgs.CUDA = &tke.DriverVersion{
							Version: common.StringPtr(advancedSetting.GPUArgs.CUDA.Version),
							Name:    common.StringPtr(advancedSetting.GPUArgs.CUDA.Name),
						}
					}

					if advancedSetting.GPUArgs.CUDNN != nil {
						gpuArgs.CUDNN = &tke.CUDNN{
							Version: common.StringPtr(advancedSetting.GPUArgs.CUDNN.Version),
							Name:    common.StringPtr(advancedSetting.GPUArgs.CUDNN.Name),
							DevName: common.StringPtr(advancedSetting.GPUArgs.CUDNN.DevName),
							DocName: common.StringPtr(advancedSetting.GPUArgs.CUDNN.DocName),
						}
					}

					if advancedSetting.GPUArgs.CustomDriver != nil {
						gpuArgs.CustomDriver = &tke.CustomDriver{
							Address: common.StringPtr(advancedSetting.GPUArgs.CustomDriver.Address),
						}
					}

					return gpuArgs
				}
				return nil
			}(),
		}

		return advancedSet
	}

	return nil
}

// generateAddExistedInstancesReq existed instance request
func generateAddExistedInstancesReq(addReq *AddExistedInstanceReq) *tke.AddExistedInstancesRequest {
	// add existed instance to cluster request
	req := tke.NewAddExistedInstancesRequest()
	req.ClusterId = common.StringPtr(addReq.ClusterID)
	req.InstanceIds = common.StringPtrs(addReq.InstanceIDs)

	if addReq.LoginSetting != nil {
		req.LoginSettings = &tke.LoginSettings{
			Password: func() *string {
				if len(addReq.LoginSetting.Password) > 0 {
					return common.StringPtr(addReq.LoginSetting.Password)
				}
				return nil
			}(),
			KeyIds: func() []*string {
				if len(addReq.LoginSetting.KeyIds) > 0 {
					return common.StringPtrs(addReq.LoginSetting.KeyIds)
				}
				return nil
			}(),
		}
	}

	if addReq.EnhancedSetting != nil {
		req.EnhancedService = &tke.EnhancedService{}
		if addReq.EnhancedSetting.SecurityService != nil &&
			addReq.EnhancedSetting.SecurityService.Enabled != nil {
			req.EnhancedService.SecurityService = &tke.RunSecurityServiceEnabled{
				Enabled: common.BoolPtr(*addReq.EnhancedSetting.SecurityService.Enabled),
			}
		}
		if addReq.EnhancedSetting.MonitorService != nil &&
			addReq.EnhancedSetting.MonitorService.Enabled != nil {
			req.EnhancedService.MonitorService = &tke.RunMonitorServiceEnabled{
				Enabled: common.BoolPtr(*addReq.EnhancedSetting.MonitorService.Enabled),
			}
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

	if len(addReq.SkipValidateOptions) > 0 {
		req.SkipValidateOptions = common.StringPtrs(addReq.SkipValidateOptions)
	}

	if len(addReq.InstanceAdvancedSettingsOverrides) > 0 {
		req.InstanceAdvancedSettingsOverrides = make([]*tke.InstanceAdvancedSettings, 0)

		for i := range addReq.InstanceAdvancedSettingsOverrides {
			req.InstanceAdvancedSettingsOverrides = append(req.InstanceAdvancedSettingsOverrides,
				generateInstanceAdvancedSetting(addReq.InstanceAdvancedSettingsOverrides[i]))
		}
	}

	if len(addReq.ImageId) > 0 {
		req.ImageId = common.StringPtr(addReq.ImageId)
	}

	return req
}

// generateClusterRequestInfo create cluster request
// nolint
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
		EniSubnetIds:         common.StringPtrs(request.ClusterCIDR.EniSubnetIds),
		ClaimExpiredSeconds: func() *int64 {
			if request.ClusterCIDR.ClaimExpiredSeconds <= 0 {
				return common.Int64Ptr(300)
			}
			return common.Int64Ptr(int64(request.ClusterCIDR.ClaimExpiredSeconds))
		}(),
		ServiceCIDR: func() *string {
			if request.ClusterCIDR.ServiceCIDR != "" {
				return common.StringPtr(request.ClusterCIDR.ServiceCIDR)
			}
			return nil
		}(),
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

	// runInstances 新增节点 mode
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
				InstanceAdvancedSettingsOverrides: func() []*tke.InstanceAdvancedSettings {

					instanceSettings := make([]*tke.InstanceAdvancedSettings, 0)

					for cnt := range request.RunInstancesForNode[i].InstanceAdvancedSettingsOverrides {
						instanceSettings = append(instanceSettings,
							generateInstanceAdvancedSetting(request.RunInstancesForNode[i].InstanceAdvancedSettingsOverrides[cnt]))
					}
					return instanceSettings
				}(),
			})
		}

		if len(request.InstanceDataDiskMountSettings) > 0 {
			for i := range request.InstanceDataDiskMountSettings {
				if req.InstanceDataDiskMountSettings == nil {
					req.InstanceDataDiskMountSettings = make([]*tke.InstanceDataDiskMountSetting, 0)
				}

				req.InstanceDataDiskMountSettings = append(req.InstanceDataDiskMountSettings, &tke.InstanceDataDiskMountSetting{
					InstanceType: request.InstanceDataDiskMountSettings[i].InstanceType,
					DataDisks: func() []*tke.DataDisk {
						disks := make([]*tke.DataDisk, 0)
						for cnt := range request.InstanceDataDiskMountSettings[i].DataDisks {
							disks = append(disks, &tke.DataDisk{
								DiskType:           common.StringPtr(request.InstanceDataDiskMountSettings[i].DataDisks[cnt].DiskType),
								FileSystem:         common.StringPtr(request.InstanceDataDiskMountSettings[i].DataDisks[cnt].FileSystem),
								DiskSize:           common.Int64Ptr(request.InstanceDataDiskMountSettings[i].DataDisks[cnt].DiskSize),
								AutoFormatAndMount: common.BoolPtr(request.InstanceDataDiskMountSettings[i].DataDisks[cnt].AutoFormatAndMount),
								MountTarget:        common.StringPtr(request.InstanceDataDiskMountSettings[i].DataDisks[cnt].MountTarget),
							})
						}
						return disks
					}(),
					Zone: request.InstanceDataDiskMountSettings[i].Zone,
				})
			}
		}

		for i := range request.Addons {
			req.ExtensionAddons = append(req.ExtensionAddons, &tke.ExtensionAddon{
				AddonName:  common.StringPtr(request.Addons[i].AddonName),
				AddonParam: common.StringPtr(request.Addons[i].AddonParam),
			})
		}

		return req, nil
	}

	// existed instance to cluster
	if len(request.ExistedInstancesForNode) > 0 {
		for i := range request.ExistedInstancesForNode {
			if len(request.ExistedInstancesForNode[i].ExistedInstancesPara.InstanceIDs) == 0 {
				return nil, fmt.Errorf("CreateClusterRequest ExistedInstancesForNode instance is null")
			}

			existedInstanceNodes := &tke.ExistedInstancesForNode{
				NodeRole: common.StringPtr(request.ExistedInstancesForNode[i].NodeRole),
				ExistedInstancesPara: &tke.ExistedInstancesPara{
					InstanceIds:   common.StringPtrs(request.ExistedInstancesForNode[i].ExistedInstancesPara.InstanceIDs),
					LoginSettings: generateLoginSet(request.ExistedInstancesForNode[i].ExistedInstancesPara.LoginSettings),
					InstanceAdvancedSettings: generateInstanceAdvancedSetting(
						request.ExistedInstancesForNode[i].ExistedInstancesPara.InstanceAdvancedSettings),
				},
			}
			if request.ExistedInstancesForNode[i].InstanceAdvancedSettingsOverride != nil {
				existedInstanceNodes.InstanceAdvancedSettingsOverride =
					generateInstanceAdvancedSetting(request.ExistedInstancesForNode[i].InstanceAdvancedSettingsOverride)
			}

			req.ExistedInstancesForNode = append(req.ExistedInstancesForNode, existedInstanceNodes)
		}
	}

	for i := range request.Addons {
		req.ExtensionAddons = append(req.ExtensionAddons, &tke.ExtensionAddon{
			AddonName:  common.StringPtr(request.Addons[i].AddonName),
			AddonParam: common.StringPtr(request.Addons[i].AddonParam),
		})
	}

	return req, nil
}

// generateClusterBasic create cluster basic setting
func generateClusterBasic(basic *ClusterBasicSettings) *tke.ClusterBasicSettings {
	tkeClusterBasic := &tke.ClusterBasicSettings{
		ClusterOs:      common.StringPtr(basic.ClusterOS),
		ClusterVersion: common.StringPtr(basic.ClusterVersion),
		ClusterName:    common.StringPtr(basic.ClusterName),
		VpcId:          common.StringPtr(basic.VpcID),
		ProjectId:      common.Int64Ptr(basic.ProjectID),
		AutoUpgradeClusterLevel: &tke.AutoUpgradeClusterLevel{
			IsAutoUpgrade: common.BoolPtr(basic.IsAutoUpgradeClusterLevel),
		},
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
	if len(basic.SubnetID) > 0 {
		tkeClusterBasic.SubnetId = common.StringPtr(basic.SubnetID)
	}
	if len(basic.ClusterLevel) > 0 {
		tkeClusterBasic.ClusterLevel = common.StringPtr(basic.ClusterLevel)
	}

	return tkeClusterBasic
}

// generateClusterAdvancedSet create cluster advanced setting
func generateClusterAdvancedSet(request *ClusterAdvancedSettings) *tke.ClusterAdvancedSettings {
	if request == nil {
		return nil
	}

	clusterAdvance := &tke.ClusterAdvancedSettings{
		IPVS:               common.BoolPtr(request.IPVS),
		ContainerRuntime:   common.StringPtr(request.ContainerRuntime),
		RuntimeVersion:     common.StringPtr(request.RuntimeVersion),
		DeletionProtection: common.BoolPtr(request.DeletionProtection),
		AuditEnabled:       common.BoolPtr(request.AuditEnabled),
		IsNonStaticIpMode:  common.BoolPtr(request.IsNonStaticIpMode),
		VpcCniType:         common.StringPtr(request.VpcCniType),
	}
	if len(request.NetworkType) > 0 {
		clusterAdvance.NetworkType = common.StringPtr(request.NetworkType)
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

// generateInstanceAdvancedSet transfer input para to tke para
func generateInstanceAdvancedSet(request *InstanceAdvancedSettings) *tke.InstanceAdvancedSettings {
	if request == nil {
		return nil
	}

	instanceAdvance := &tke.InstanceAdvancedSettings{
		MountTarget:     common.StringPtr(request.MountTarget),
		DockerGraphPath: common.StringPtr(request.DockerGraphPath),
		Unschedulable:   request.Unschedulable,
		UserScript:      common.StringPtr(request.UserScript),
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

// generateEnhancedService enhanced service
func generateEnhancedService(service *EnhancedService) *tke.EnhancedService { // nolint
	if service == nil {
		return nil
	}

	svc := &tke.EnhancedService{}
	if service.SecurityService != nil && service.SecurityService.Enabled != nil {
		svc.SecurityService = &tke.RunSecurityServiceEnabled{Enabled: common.BoolPtr(*service.SecurityService.Enabled)}
	}
	if service.MonitorService != nil && service.MonitorService.Enabled != nil {
		svc.MonitorService = &tke.RunMonitorServiceEnabled{Enabled: common.BoolPtr(*service.MonitorService.Enabled)}
	}
	return svc
}

// generateLoginSet login setting
func generateLoginSet(settings *LoginSettings) *tke.LoginSettings {
	if settings == nil {
		return nil
	}

	if len(settings.KeyIds) == 0 && settings.Password == "" {
		return nil
	}

	return &tke.LoginSettings{
		Password: func() *string {
			if len(settings.Password) > 0 {
				return common.StringPtr(settings.Password)
			}
			return nil
		}(),
		KeyIds: func() []*string {
			if len(settings.KeyIds) > 0 {
				return common.StringPtrs(settings.KeyIds)
			}
			return nil
		}(),
	}
}

// generateClusterNodePool cluster node pool
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
	req.ContainerRuntime = nodePool.ContainerRuntime
	req.RuntimeVersion = nodePool.RuntimeVersion
	return req
}

func generateLabel(labels []*Label) []*tke.Label {
	if len(labels) == 0 {
		return nil
	}

	result := make([]*tke.Label, 0)
	for _, v := range labels {
		result = append(result, &tke.Label{Name: v.Name, Value: v.Value})
	}
	return result
}

func generateTaint(taints []*Taint) []*tke.Taint {
	if len(taints) == 0 {
		return nil
	}

	result := make([]*tke.Taint, 0)
	for _, v := range taints {
		result = append(result, &tke.Taint{Key: v.Key, Value: v.Value, Effect: v.Effect})
	}
	return result
}

func generateTag(tags []*Tag) []*tke.Tag {
	if len(tags) == 0 {
		return nil
	}

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

func convertModifyNodePool(nodePool *ModifyClusterNodePoolInput) *tke.ModifyClusterNodePoolRequest { // nolint
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
func MapToTaints(taints []*proto.Taint) []*Taint {
	result := make([]*Taint, 0)
	for _, v := range taints {
		key := v.Key
		value := v.Value
		effect := v.Effect
		result = append(result, &Taint{Key: &key, Value: &value, Effect: &effect})
	}
	return result
}

// MapToTags converts a map of string-string to a slice of Tag
func MapToTags(tags map[string]string) []*Tag {
	result := make([]*Tag, 0)
	for k, v := range tags {
		key := k
		value := v
		result = append(result, &Tag{Key: &key, Value: &value})
	}
	return result
}

// MapToCloudLabels converts a map of string-string to a slice of Label
func MapToCloudLabels(labels map[string]string) []*tke.Label {
	result := make([]*tke.Label, 0)
	for k, v := range labels {
		name := k
		value := v
		result = append(result, &tke.Label{Name: &name, Value: &value})
	}
	return result
}

// MapToCloudTaints converts a map of string-string to a slice of Taint
func MapToCloudTaints(taints []*proto.Taint) []*tke.Taint {
	result := make([]*tke.Taint, 0)
	for _, v := range taints {
		key := v.Key
		value := v.Value
		effect := v.Effect
		result = append(result, &tke.Taint{Key: &key, Value: &value, Effect: &effect})
	}
	return result
}

// MapToCloudTags converts a map of string-string to a slice of Tag
func MapToCloudTags(tags map[string]string) []*tke.Tag {
	result := make([]*tke.Tag, 0)
	for k, v := range tags {
		key := k
		value := v
		result = append(result, &tke.Tag{Key: &key, Value: &value})
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

const (
	cacheImageNameToImage = "cached_image_name"
)

func buildCacheName(keyPrefix string, region, name string) string {
	return fmt.Sprintf("%s_%v_%v", keyPrefix, region, name)
}

// setImageNameCacheData set image cache
func setImageNameCacheData(region, name string, imageData *cvm.Image) error {
	cacheName := buildCacheName(cacheImageNameToImage, region, name)

	var err error

	image, exist := cache.GetCache().Get(cacheName)
	if exist {
		blog.Infof("SetImageNameCacheData cacheName:%s, cache exist %+v", cacheName, image)
		err = cache.GetCache().Replace(cacheName, imageData, gocache.DefaultExpiration)
	} else {
		err = cache.GetCache().Add(cacheName, imageData, gocache.DefaultExpiration)
	}
	if err != nil {
		return err
	}

	return nil
}

// getImageNameCacheData get image name data
func getImageNameCacheData(region, name string) (*cvm.Image, bool) {
	cacheName := buildCacheName(cacheImageNameToImage, region, name)

	val, ok := cache.GetCache().Get(cacheName)
	if ok && val != nil {
		blog.Infof("GetImageNameCacheData cacheName:%s, cache exist %+v", cacheName, val)
		if image, ok1 := val.(*cvm.Image); ok1 {
			return image, true
		}
	}

	return nil, false
}
