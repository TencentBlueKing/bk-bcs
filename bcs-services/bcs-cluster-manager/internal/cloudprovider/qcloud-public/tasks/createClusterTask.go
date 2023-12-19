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
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud-public/business"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/qcloud/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/template"
	providerutils "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/utils"
	icommon "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/loop"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/utils"
)

// as far as possible to keep every operation unit simple

// generateClusterCIDRInfo cidr info
func generateClusterCIDRInfo(ctx context.Context,
	info *cloudprovider.CloudDependBasicInfo) (*api.ClusterCIDRSettings, error) {
	cidrInfo := &api.ClusterCIDRSettings{
		// 用于分配集群容器和服务 IP 的 CIDR，不得与 VPC CIDR 冲突，也不得与同 VPC 内其他集群 CIDR 冲突。
		// 且网段范围必须在内网网段内，例如:10.1.0.0/14, 192.168.0.1/18,172.16.0.0/16。
		ClusterCIDR:          info.Cluster.NetworkSettings.ClusterIPv4CIDR,
		MaxNodePodNum:        uint64(info.Cluster.NetworkSettings.MaxNodePodNum),
		MaxClusterServiceNum: uint64(info.Cluster.NetworkSettings.MaxServiceNum),
		ServiceCIDR:          info.Cluster.NetworkSettings.ServiceIPv4CIDR,
	}

	// vpc-cni模式下设置
	if info.Cluster.GetClusterAdvanceSettings().GetNetworkType() == icommon.VpcCni {
		if info.Cluster.GetNetworkSettings().GetClaimExpiredSeconds() > 0 {
			cidrInfo.ClaimExpiredSeconds = info.Cluster.GetNetworkSettings().GetClaimExpiredSeconds()
		}

		subnetIds := make([]string, 0)
		if len(info.Cluster.GetNetworkSettings().GetSubnetSource().GetNew()) > 0 {
			// 各个可用区自动分配指定数量的子网
			ids, err := business.AllocateClusterVpcCniSubnets(ctx, info.Cluster.ClusterID, info.Cluster.VpcID,
				info.Cluster.GetNetworkSettings().GetSubnetSource().GetNew(), info.CmOption)
			if err != nil {
				return nil, err
			}

			subnetIds = append(subnetIds, ids...)
		}

		if len(info.Cluster.GetNetworkSettings().GetSubnetSource().GetExisted().GetIds()) > 0 {
			subnetIds = append(subnetIds, info.Cluster.GetNetworkSettings().GetSubnetSource().GetExisted().GetIds()...)
		}

		// update cluster subnetIds
		info.Cluster.NetworkSettings.EniSubnetIDs = subnetIds
		cidrInfo.EniSubnetIds = subnetIds
	}

	return cidrInfo, nil
}

// generateClusterBasicInfo cluster basic info
func generateClusterBasicInfo(cluster *proto.Cluster) *api.ClusterBasicSettings {
	basicInfo := &api.ClusterBasicSettings{
		// 集群操作系统，支持设置公共镜像(字段传相应镜像Name)和自定义镜像(字段传相应镜像ID)
		// 详情参考：https://cloud.tencent.com/document/product/457/68289
		ClusterOS:          cluster.ClusterBasicSettings.OS,
		ClusterVersion:     cluster.ClusterBasicSettings.Version,
		ClusterName:        cluster.ClusterID,
		ClusterDescription: cluster.GetDescription(),
		VpcID:              cluster.VpcID,
		ProjectID: func() int64 {
			extra := cluster.GetExtraInfo()
			id, ok := extra[icommon.CloudProjectId]
			if ok {
				projectId, _ := strconv.Atoi(id)
				return int64(projectId)
			}

			return 0
		}(),
		TagSpecification: func() []*api.TagSpecification {
			// build qcloud tag info
			if len(cluster.ClusterBasicSettings.ClusterTags) > 0 {
				var (
					cloudClusterTags = make([]*api.TagSpecification, 0)
					tags             = make([]*api.Tag, 0)
				)

				for k, v := range cluster.ClusterBasicSettings.ClusterTags {
					tags = append(tags, &api.Tag{
						Key:   common.StringPtr(k),
						Value: common.StringPtr(v),
					})
				}
				cloudClusterTags = append(cloudClusterTags, &api.TagSpecification{
					ResourceType: icommon.TagClusterResourceKey,
					Tags:         tags,
				})

				return cloudClusterTags
			}

			return nil
		}(),
		// 当选择Cilium Overlay网络插件时，TKE会从该子网获取2个IP用来创建内网负载均衡
		SubnetID: cluster.ClusterBasicSettings.SubnetID,
		// 托管集群等级 & 是否自动变更集群等级
		ClusterLevel:              cluster.ClusterBasicSettings.ClusterLevel,
		IsAutoUpgradeClusterLevel: cluster.ClusterBasicSettings.IsAutoUpgradeClusterLevel,
	}

	return basicInfo
}

// generateClusterAdvancedInfo cluster advanced info
func generateClusterAdvancedInfo(cluster *proto.Cluster) *api.ClusterAdvancedSettings {
	advancedInfo := &api.ClusterAdvancedSettings{
		IPVS:             cluster.ClusterAdvanceSettings.IPVS,
		ContainerRuntime: cluster.ClusterAdvanceSettings.ContainerRuntime,
		// 集群网络类型（包括GR(全局路由)和VPC-CNI两种模式，默认为GR
		NetworkType: cluster.ClusterAdvanceSettings.NetworkType,
		ExtraArgs:   &api.ClusterExtraArgs{},
		IsNonStaticIpMode: func() bool {
			return !cluster.GetNetworkSettings().GetIsStaticIpMode()
		}(),
		VpcCniType:         api.TKERouteEni,
		RuntimeVersion:     cluster.ClusterAdvanceSettings.RuntimeVersion,
		DeletionProtection: cluster.ClusterAdvanceSettings.DeletionProtection,
		AuditEnabled:       cluster.ClusterAdvanceSettings.AuditEnabled,
	}

	// cluster control component extraArgs
	if len(cluster.ClusterAdvanceSettings.ExtraArgs) > 0 {
		if apiserver, ok := cluster.ClusterAdvanceSettings.ExtraArgs[icommon.KubeAPIServer]; ok {
			paras := strings.Split(apiserver, ";")
			advancedInfo.ExtraArgs.KubeAPIServer = common.StringPtrs(paras)
		}

		if controller, ok := cluster.ClusterAdvanceSettings.ExtraArgs[icommon.KubeController]; ok {
			paras := strings.Split(controller, ";")
			advancedInfo.ExtraArgs.KubeControllerManager = common.StringPtrs(paras)
		}

		if scheduler, ok := cluster.ClusterAdvanceSettings.ExtraArgs[icommon.KubeScheduler]; ok {
			paras := strings.Split(scheduler, ";")
			advancedInfo.ExtraArgs.KubeScheduler = common.StringPtrs(paras)
		}

		if etcd, ok := cluster.ClusterAdvanceSettings.ExtraArgs[icommon.Etcd]; ok {
			paras := strings.Split(etcd, ";")
			advancedInfo.ExtraArgs.Etcd = common.StringPtrs(paras)
		}
	}

	return advancedInfo
}

// 独立集群创建集群请求

// handleClusterMasterNodes handle cluster master nodes
func handleClusterMasterNodes(ctx context.Context, req *api.CreateClusterRequest,
	info *cloudprovider.CloudDependBasicInfo, instanceIDs []string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// filter nodes data disks
	filterDisk, err := business.FilterNodesByDataDisk(instanceIDs, info.CmOption)
	if err != nil {
		blog.Errorf("createClusterReq[%s] FilterNodesByDataDisk[%s] failed: %+v",
			taskID, info.Cluster.ClusterID, err)
		return err
	}

	blog.Infof("createClusterReq FilterNodesByDataDisk result[%+v]", filterDisk)

	if req.ExistedInstancesForNode == nil {
		req.ExistedInstancesForNode = make([]*api.ExistedInstancesForNode, 0)
	}

	// single disk & many disk
	if len(filterDisk.SingleDiskInstance) > 0 {
		req.ExistedInstancesForNode = append(req.ExistedInstancesForNode,
			generateMasterExistedInstance(api.MASTER_ETCD.String(), filterDisk.SingleDiskInstance, false, info.Cluster))
	}
	if len(filterDisk.ManyDiskInstance) > 0 {
		req.ExistedInstancesForNode = append(req.ExistedInstancesForNode,
			generateMasterExistedInstance(api.MASTER_ETCD.String(), filterDisk.ManyDiskInstance, true, info.Cluster))
	}

	return nil
}

// generateMasterExistedInstance cluster master setting
func generateMasterExistedInstance(role string, instanceIDs []string, manyDisk bool,
	cls *proto.Cluster) *api.ExistedInstancesForNode {
	existedInstance := &api.ExistedInstancesForNode{
		NodeRole: role,
		ExistedInstancesPara: &api.ExistedInstancesPara{
			InstanceIDs: instanceIDs,
			LoginSettings: func() *api.LoginSettings {
				return &api.LoginSettings{
					Password: func() string {
						if len(cls.GetNodeSettings().GetMasterLogin().InitLoginPassword) > 0 {
							return cls.GetNodeSettings().GetMasterLogin().GetInitLoginPassword()
						}
						return ""
					}(),
					KeyIds: func() []string {
						if len(cls.GetNodeSettings().GetMasterLogin().GetKeyPair().GetKeyID()) > 0 {
							return strings.Split(cls.GetNodeSettings().GetMasterLogin().GetKeyPair().GetKeyID(), ",")
						}

						return nil
					}(),
				}
			}(),
			InstanceAdvancedSettings: business.GenerateInstanceAdvanceInfo(cls,
				&business.NodeAdvancedOptions{NodeScheduler: true}),
			SecurityGroupIds: func() []*string {
				if len(cls.GetNodeSettings().GetMasterSecurityGroups()) == 0 {
					return nil
				}

				return common.StringPtrs(cls.GetNodeSettings().GetMasterSecurityGroups())
			}(),
		},
	}

	// instance advanced setting override
	existedInstance.InstanceAdvancedSettingsOverride = business.GenerateInstanceAdvanceInfo(cls,
		&business.NodeAdvancedOptions{NodeScheduler: true})
	if manyDisk {
		existedInstance.InstanceAdvancedSettingsOverride.DataDisks =
			[]api.DataDetailDisk{api.GetDefaultDataDisk(api.Ext4)}
	}

	return existedInstance
}

// handleClusterWorkerNodes handle cluster worker nodes
func handleClusterWorkerNodes(ctx context.Context, req *api.CreateClusterRequest,
	info *cloudprovider.CloudDependBasicInfo, instanceIDs []string, operator string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	if len(instanceIDs) == 0 {
		return nil
	}

	if req.ExistedInstancesForNode == nil {
		req.ExistedInstancesForNode = make([]*api.ExistedInstancesForNode, 0)
	}

	// filter nodes data disks
	filterDisk, err := business.FilterNodesByDataDisk(instanceIDs, info.CmOption)
	if err != nil {
		blog.Errorf("handleClusterWorkerNodes[%s] FilterNodesByDataDisk[%s] failed: %+v",
			taskID, info.Cluster.ClusterID, err)
		retErr := fmt.Errorf("call FilterNodesByDataDisk[%s] api err, %s", info.Cluster.ClusterID, err.Error())
		return retErr
	}

	blog.Infof("handleClusterWorkerNodes[%s] FilterNodesByDataDisk result[%+v]", taskID, filterDisk)

	// single disk
	if len(filterDisk.SingleDiskInstance) > 0 {
		req.ExistedInstancesForNode = append(req.ExistedInstancesForNode,
			generateWorkerExistedInstance(info, filterDisk.SingleDiskInstance, filterDisk.SingleDiskInstanceIP,
				false, operator))
	}
	// many disk
	if len(filterDisk.ManyDiskInstance) > 0 {
		req.ExistedInstancesForNode = append(req.ExistedInstancesForNode,
			generateWorkerExistedInstance(info, filterDisk.ManyDiskInstance, filterDisk.ManyDiskInstanceIP,
				true, operator))
	}

	return nil
}

// generateWorkerExistedInstance cluster worker setting
func generateWorkerExistedInstance(info *cloudprovider.CloudDependBasicInfo, nodeIDs, nodeIPs []string,
	manyDisk bool, operator string) *api.ExistedInstancesForNode {
	existedInstance := &api.ExistedInstancesForNode{
		NodeRole: api.WORKER.String(),
		ExistedInstancesPara: &api.ExistedInstancesPara{
			InstanceIDs: nodeIDs,
			LoginSettings: func() *api.LoginSettings {
				return &api.LoginSettings{
					Password: func() string {
						if len(info.Cluster.GetNodeSettings().GetWorkerLogin().InitLoginPassword) > 0 {
							return info.Cluster.GetNodeSettings().GetWorkerLogin().InitLoginPassword
						}
						return ""
					}(),
					KeyIds: func() []string {
						if len(info.Cluster.GetNodeSettings().GetWorkerLogin().GetKeyPair().GetKeyID()) > 0 {
							return strings.Split(info.Cluster.GetNodeSettings().GetWorkerLogin().GetKeyPair().GetKeyID(), ",")
						}

						return nil
					}(),
				}
			}(),
			InstanceAdvancedSettings: business.GenerateClsAdvancedInsSettingFromNT(info, template.RenderVars{
				Cluster:  info.Cluster,
				IPList:   strings.Join(nodeIPs, ","),
				Operator: operator,
				Render:   true,
			}, &business.NodeAdvancedOptions{NodeScheduler: true}),
			SecurityGroupIds: func() []*string {
				if len(info.Cluster.GetNodeSettings().GetWorkerSecurityGroups()) == 0 {
					return nil
				}

				return common.StringPtrs(info.Cluster.GetNodeSettings().GetWorkerSecurityGroups())
			}(),
		},
	}

	// instance advanced setting
	existedInstance.InstanceAdvancedSettingsOverride = business.GenerateClsAdvancedInsSettingFromNT(info,
		template.RenderVars{
			Cluster:  info.Cluster,
			IPList:   strings.Join(nodeIPs, ","),
			Operator: operator,
			Render:   true,
		}, &business.NodeAdvancedOptions{NodeScheduler: true})

	if manyDisk {
		existedInstance.InstanceAdvancedSettingsOverride.DataDisks =
			[]api.DataDetailDisk{api.GetDefaultDataDisk(api.Ext4)}
	}

	return existedInstance
}

// disksToCVMDisks transfer cvm disk
func disksToCVMDisks(disks []*proto.CloudDataDisk) []*cvm.DataDisk {
	if len(disks) == 0 {
		return nil
	}

	cvmDisks := make([]*cvm.DataDisk, 0)
	for i := range disks {
		size, _ := utils.StringToInt(disks[i].DiskSize)

		cvmDisks = append(cvmDisks, &cvm.DataDisk{
			DiskSize: common.Int64Ptr(int64(size)),
			DiskType: common.StringPtr(disks[i].DiskType),
		})
	}

	return cvmDisks
}

// generateNewRunInstance run instances by instance template
// nolint
func generateNewRunInstance(info *cloudprovider.CloudDependBasicInfo, role string,
	templates []*proto.InstanceTemplateConfig, operator string) *api.RunInstancesForNode {
	runInstance := &api.RunInstancesForNode{
		NodeRole: role,
	}

	// create instance template
	for i := range templates {
		createInsRequest := cvm.NewRunInstancesRequest()

		// 实例计费类型: 默认值 POSTPAID_BY_HOUR
		createInsRequest.InstanceChargeType = func() *string {
			if templates[i].GetInstanceChargeType() == "" {
				return common.StringPtr(api.POSTPAIDBYHOUR)
			}

			return common.StringPtr(templates[i].GetInstanceChargeType())
		}()

		createInsRequest.InstanceChargePrepaid = func() *cvm.InstanceChargePrepaid {
			if templates[i].GetCharge() == nil || templates[i].GetCharge().GetPeriod() == 0 ||
				templates[i].GetCharge().GetRenewFlag() == "" {
				return nil
			}
			return &cvm.InstanceChargePrepaid{
				Period:    common.Int64Ptr(int64(templates[i].GetCharge().GetPeriod())),
				RenewFlag: common.StringPtr(templates[i].GetCharge().GetRenewFlag()),
			}
		}()

		createInsRequest.Placement = &cvm.Placement{
			Zone: common.StringPtr(templates[i].GetZone()),
			ProjectId: func() *int64 {
				extra := info.Cluster.GetExtraInfo()
				id, ok := extra[icommon.CloudProjectId]
				if ok {
					projectId, _ := strconv.Atoi(id)
					return common.Int64Ptr(int64(projectId))
				}

				return nil
			}(),
		}

		createInsRequest.InstanceType = common.StringPtr(templates[i].GetInstanceType())
		// createInsRequest.ImageId = nil

		systemDiskSize, _ := utils.StringToInt(templates[i].GetSystemDisk().GetDiskSize())
		createInsRequest.SystemDisk = &cvm.SystemDisk{
			DiskType: common.StringPtr(templates[i].GetSystemDisk().GetDiskType()),
			DiskSize: common.Int64Ptr(int64(systemDiskSize)),
		}
		createInsRequest.DataDisks = disksToCVMDisks(templates[i].GetCloudDataDisks())

		createInsRequest.VirtualPrivateCloud = &cvm.VirtualPrivateCloud{
			VpcId: func() *string {
				if len(templates[i].GetVpcID()) != 0 {
					return common.StringPtr(templates[i].VpcID)
				}

				return common.StringPtr(info.Cluster.GetVpcID())
			}(),
			SubnetId: common.StringPtr(templates[i].SubnetID),
		}

		// 公网带宽相关信息设置
		createInsRequest.InternetAccessible = func() *cvm.InternetAccessible {
			if templates[i].GetInternetAccess() != nil {

				internet := &cvm.InternetAccessible{
					InternetChargeType: common.StringPtr(api.InternetChargeTypeTrafficPostpaidByHour),
				}

				if templates[i].GetInternetAccess().GetPublicIPAssigned() {
					internet.PublicIpAssigned = common.BoolPtr(true)

					bw, _ := strconv.Atoi(templates[i].GetInternetAccess().GetInternetMaxBandwidth())
					internet.InternetMaxBandwidthOut = common.Int64Ptr(int64(bw))
				}
				if templates[i].GetInternetAccess().GetInternetChargeType() != "" {
					internet.InternetChargeType = common.StringPtr(templates[i].GetInternetAccess().GetInternetChargeType())
				}
				if templates[i].GetInternetAccess().GetBandwidthPackageId() != "" {
					internet.BandwidthPackageId = common.StringPtr(templates[i].GetInternetAccess().GetBandwidthPackageId())
				}

				return internet
			}

			return nil
		}()
		// createInsRequest.EnhancedService = nil
		createInsRequest.InstanceCount = common.Int64Ptr(int64(templates[i].GetApplyNum()))
		createInsRequest.LoginSettings = &cvm.LoginSettings{
			Password: func() *string {
				switch role {
				case api.MASTER_ETCD.String():
					if len(info.Cluster.GetNodeSettings().GetMasterLogin().GetInitLoginPassword()) > 0 {
						return common.StringPtr(info.Cluster.GetNodeSettings().GetMasterLogin().GetInitLoginPassword())
					}
					return nil
				case api.WORKER.String():
					if len(info.Cluster.GetNodeSettings().GetWorkerLogin().GetInitLoginPassword()) > 0 {
						return common.StringPtr(info.Cluster.GetNodeSettings().GetWorkerLogin().GetInitLoginPassword())
					}
					return nil
				default:
				}
				return nil
			}(),
			KeyIds: func() []*string {
				switch role {
				case api.MASTER_ETCD.String():
					if len(info.Cluster.GetNodeSettings().GetMasterLogin().GetKeyPair().GetKeyID()) > 0 {
						keyIds := strings.Split(info.Cluster.GetNodeSettings().
							GetMasterLogin().GetKeyPair().GetKeyID(), ",")
						return common.StringPtrs(keyIds)
					}
					return nil
				case api.WORKER.String():
					if len(info.Cluster.GetNodeSettings().GetWorkerLogin().GetKeyPair().GetKeyID()) > 0 {
						keyIds := strings.Split(info.Cluster.GetNodeSettings().
							GetWorkerLogin().GetKeyPair().GetKeyID(), ",")
						return common.StringPtrs(keyIds)
					}
					return nil
				default:
				}
				return nil
			}(),
		}
		createInsRequest.SecurityGroupIds = common.StringPtrs(templates[i].GetSecurityGroupIDs())

		requestStr := createInsRequest.ToJsonString()
		runInstance.RunInstancesPara = append(runInstance.RunInstancesPara, common.StringPtr(requestStr))

		runInstance.InstanceAdvancedSettingsOverrides = append(runInstance.InstanceAdvancedSettingsOverrides,
			business.GenerateClsAdvancedInsSettingFromNT(info, template.RenderVars{
				Cluster:  info.Cluster,
				IPList:   "",
				Operator: operator,
				Render:   true,
			}, &business.NodeAdvancedOptions{
				NodeScheduler: true,
				Disks:         templates[i].GetCloudDataDisks(),
			}))
	}

	return runInstance
}

// generateNewRunInstance run instances by instance template
func generateNewInstanceForDisk(templates []*proto.InstanceTemplateConfig) []*api.InstanceDataDiskMountSetting {
	diskMounts := make([]*api.InstanceDataDiskMountSetting, 0)

	for i := range templates {
		diskMounts = append(diskMounts, &api.InstanceDataDiskMountSetting{
			InstanceType: common.StringPtr(templates[i].InstanceType),
			DataDisks: func() []*api.DataDetailDisk {
				localDisk := make([]*api.DataDetailDisk, 0)
				for cnt := range templates[i].CloudDataDisks {
					size, _ := strconv.Atoi(templates[i].CloudDataDisks[cnt].DiskSize)
					localDisk = append(localDisk, &api.DataDetailDisk{
						DiskType:           templates[i].CloudDataDisks[cnt].DiskType,
						DiskSize:           int64(size),
						FileSystem:         templates[i].CloudDataDisks[cnt].FileSystem,
						MountTarget:        templates[i].CloudDataDisks[cnt].MountTarget,
						AutoFormatAndMount: templates[i].CloudDataDisks[cnt].AutoFormatAndMount,
					})
				}

				return localDisk
			}(),
			Zone: common.StringPtr(templates[i].Zone),
		})
	}

	return diskMounts
}

// generateCreateClusterRequest 独立集群 or 托管集群
// nolint
func generateCreateClusterRequest(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	masterIps, workerIps []string, operator string) (*api.CreateClusterRequest, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// cluster create request
	req := &api.CreateClusterRequest{
		// 新增节点 or 使用已有节点
		AddNodeMode: info.Cluster.AutoGenerateMasterNodes,

		Region:      info.Cluster.Region,
		ClusterType: info.Cluster.ManageType,

		ClusterBasic:    generateClusterBasicInfo(info.Cluster),
		ClusterAdvanced: generateClusterAdvancedInfo(info.Cluster),

		InstanceAdvanced: business.GenerateClsAdvancedInsSettingFromNT(info, template.RenderVars{
			Cluster:  info.Cluster,
			IPList:   "",
			Operator: operator,
			Render:   true,
		}, &business.NodeAdvancedOptions{NodeScheduler: true}),

		ExistedInstancesForNode:       make([]*api.ExistedInstancesForNode, 0),
		RunInstancesForNode:           make([]*api.RunInstancesForNode, 0),
		InstanceDataDiskMountSettings: nil,
	}

	// network info
	cidr, err := generateClusterCIDRInfo(ctx, info)
	if err != nil {
		return nil, err
	}
	req.ClusterCIDR = cidr

	switch info.Cluster.ManageType {
	case icommon.ClusterManageTypeIndependent:
		// 新增节点模式
		if req.AddNodeMode {
			masterNodesTpl, workerNodesTpl := business.GetMasterNodeTemplateConfig(info.Cluster.Template)

			// master && worker nodes
			req.RunInstancesForNode = append(req.RunInstancesForNode,
				generateNewRunInstance(info, api.MASTER_ETCD.String(), masterNodesTpl, operator))

			req.RunInstancesForNode = append(req.RunInstancesForNode,
				generateNewRunInstance(info, api.WORKER.String(), workerNodesTpl, operator))

			if req.InstanceDataDiskMountSettings == nil {
				req.InstanceDataDiskMountSettings = make([]*api.InstanceDataDiskMountSetting, 0)
			}

			req.InstanceDataDiskMountSettings = generateNewInstanceForDisk(workerNodesTpl)

			return req, nil
		}
		// 使用已有节点模式
		masterIds, err := trans2InsIdByInsIp(masterIps, info.CmOption)
		if err != nil {
			blog.Errorf("generateIndependentClusterRequest[%s] trans2InsIdByInsIp masterIps failed: %v", taskID, err)
			return nil, err
		}
		err = handleClusterMasterNodes(ctx, req, info, masterIds)
		if err != nil {
			blog.Errorf("createTkeCluster[%s] handleClusterMasterNodes for cluster[%s] failed: %v",
				taskID, info.Cluster.ClusterID, err)
			return nil, err
		}

		// 节点:  使用已存在节点 / 也可能是通过自动生产
		if len(workerIps) > 0 {
			workerIds, err := trans2InsIdByInsIp(workerIps, info.CmOption)
			if err != nil {
				blog.Errorf("generateIndependentClusterRequest[%s] trans2InsIdByInsIp workerIps failed: %v", taskID, err)
				return nil, err
			}

			err = handleClusterWorkerNodes(ctx, req, info, workerIds, operator)
			if err != nil {
				blog.Errorf("createTkeCluster[%s] handleClusterWorkerNodes for cluster[%s] failed: %v",
					taskID, info.Cluster.ClusterID, err)
				return nil, err
			}
		}
	case icommon.ClusterManageTypeManaged:
		// 新增节点模式
		if req.AddNodeMode {
			_, workerNodesTpl := business.GetMasterNodeTemplateConfig(info.Cluster.Template)

			// worker nodes
			req.RunInstancesForNode = append(req.RunInstancesForNode,
				generateNewRunInstance(info, api.WORKER.String(), workerNodesTpl, operator))

			if req.InstanceDataDiskMountSettings == nil {
				req.InstanceDataDiskMountSettings = make([]*api.InstanceDataDiskMountSetting, 0)
			}

			req.InstanceDataDiskMountSettings = generateNewInstanceForDisk(workerNodesTpl)

			return req, nil
		}

		// 节点:  使用已存在节点
		if len(workerIps) > 0 {
			workerIds, err := trans2InsIdByInsIp(workerIps, info.CmOption)
			if err != nil {
				blog.Errorf("generateIndependentClusterRequest[%s] trans2InsIdByInsIp workerIps failed: %v", taskID, err)
				return nil, err
			}

			err = handleClusterWorkerNodes(ctx, req, info, workerIds, operator)
			if err != nil {
				blog.Errorf("createTkeCluster[%s] handleClusterWorkerNodes for cluster[%s] failed: %v",
					taskID, info.Cluster.ClusterID, err)
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("generateCreateClusterRequest[%s] not supported clusterType[%s]",
			taskID, info.Cluster.ManageType)
	}

	// handle default addon parameters
	// req.Addons = handleTkeDefaultExtensionAddons(ctx, info.CmOption)

	return req, nil
}

func trans2InsIdByInsIp(ips []string, opt *cloudprovider.CommonOption) ([]string, error) {
	nodes, err := business.ListNodesByIP(ips, &cloudprovider.ListNodesOption{
		Common: opt,
	})
	if err != nil {
		return nil, err
	}

	instanceIds := make([]string, 0)
	for i := range nodes {
		instanceIds = append(instanceIds, nodes[i].NodeID)
	}
	return instanceIds, nil
}

// createCluster check cluster if exist, create cluster when not exist
func createCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo,
	request *api.CreateClusterRequest, clsId string) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	tkeCli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("createCluster[%s]: get tke client for cluster[%s] failed, %s",
			taskID, info.Cluster.ClusterID, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		return "", retErr
	}

	if clsId != "" {
		// check clusterId if exist in qcloud
		tkeCluster, errGet := tkeCli.GetTKECluster(clsId)
		if errGet != nil {
			blog.Errorf("createCluster[%s] GetTKECluster[%s] failed, %s",
				taskID, info.Cluster.ClusterID, errGet.Error())
			retErr := fmt.Errorf("call GetTKECluster[%s] api err, %s", info.Cluster.ClusterID, errGet.Error())
			return "", retErr
		}
		return *tkeCluster.ClusterId, nil
	}

	resp, errCreate := tkeCli.CreateTKECluster(request)
	if errCreate != nil {
		blog.Errorf("createCluster[%s] call CreateTKECluster[%s] failed, %s",
			taskID, info.Cluster.ClusterID, errCreate.Error())
		retErr := fmt.Errorf("call CreateTKECluster[%s] api err, %s", info.Cluster.ClusterID, errCreate.Error())
		return "", retErr
	}
	blog.Infof("createCluster[%s] CreateTKECluster[%s] successful", taskID, info.Cluster.ClusterID)

	return resp.ClusterID, nil
}

// createTkeCluster create tke cluster
func createTkeCluster(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, // nolint
	masterIps []string, workerIps []string, operator string) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// 分别操作 独立集群 和 托管集群
	req, err := generateCreateClusterRequest(ctx, info, masterIps, workerIps, operator)
	if err != nil {
		blog.Errorf("createTkeCluster[%s] generateCreateClusterRequest failed: %v", taskID, err)
		return "", err
	}

	// handle default addon parameters
	req.Addons = handleTkeDefaultExtensionAddons(ctx, info.CmOption)

	systemId, err := createCluster(ctx, info, req, info.Cluster.SystemID)
	if err != nil {
		blog.Errorf("createTkeCluster[%s] call createCluster[%s] failed, %s",
			taskID, info.Cluster.ClusterID, err.Error())
		retErr := fmt.Errorf("call CreateTKECluster[%s] api err, %s", info.Cluster.ClusterID, err.Error())
		return "", retErr
	}

	blog.Infof("createTkeCluster[%s] CreateTKECluster[%s] successful", taskID, info.Cluster.ClusterID)

	// update cluster systemID
	info.Cluster.SystemID = systemId

	err = cloudprovider.GetStorageModel().UpdateCluster(ctx, info.Cluster)
	if err != nil {
		blog.Errorf("createTkeCluster[%s] updateClusterSystemID[%s] failed %s",
			taskID, info.Cluster.ClusterID, err.Error())
		retErr := fmt.Errorf("call CreateTKECluster updateClusterSystemID[%s] api err: %s",
			info.Cluster.ClusterID, err.Error())
		return "", retErr
	}
	blog.Infof("createTkeCluster[%s] call CreateTKECluster updateClusterSystemID successful", taskID)

	return systemId, nil
}

// CreateTkeClusterTask call qcloud interface to create cluster
func CreateTkeClusterTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CreateTkeClusterTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CreateTkeClusterTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	// independent cluster use existed nodes
	masterNodes := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params,
		cloudprovider.MasterNodeIPsKey.String(), ",")
	workerNodes := cloudprovider.ParseNodeIpOrIdFromCommonMap(step.Params,
		cloudprovider.WorkerNodeIPsKey.String(), ",")

	nodeTemplateID := step.Params[cloudprovider.NodeTemplateIDKey.String()]
	operator := state.Task.CommonParams[cloudprovider.OperatorKey.String()]

	// get dependent basic info
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID:      clusterID,
		CloudID:        cloudID,
		NodeTemplateID: nodeTemplateID,
	})
	if err != nil {
		blog.Errorf("CreateTkeClusterTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error()) // nolint
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// create cluster task
	clsId, err := createTkeCluster(ctx, dependInfo, masterNodes, workerNodes, operator)
	if err != nil {
		blog.Errorf("CreateTkeClusterTask[%s] createTkeCluster for cluster[%s] failed, %s",
			taskID, clusterID, err.Error())
		retErr := fmt.Errorf("createTkeCluster err, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	state.Task.CommonParams[cloudprovider.CloudSystemID.String()] = clsId

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CreateTkeClusterTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}
	return nil
}

// checkClusterStatus check cluster status
func checkClusterStatus(ctx context.Context, info *cloudprovider.CloudDependBasicInfo, systemID string) error {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	// get qcloud client
	cli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] get tke client failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("get cloud tke client err, %s", err.Error())
		return retErr
	}

	var (
		abnormal = false
	)

	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// loop cluster status
	err = loop.LoopDoFunc(ctx, func() error {
		cluster, errGet := cli.GetTKECluster(systemID)
		if errGet != nil {
			blog.Errorf("checkClusterStatus[%s] GetTKECluster failed: %v", taskID, errGet)
			return nil
		}

		blog.Infof("checkClusterStatus[%s] cluster[%s] current status[%s]", taskID,
			info.Cluster.ClusterID, *cluster.ClusterStatus)

		switch *cluster.ClusterStatus {
		case api.ClusterStatusRunning:
			return loop.EndLoop
		case api.ClusterStatusAbnormal:
			abnormal = true
			return loop.EndLoop
		}

		return nil
	}, loop.LoopInterval(10*time.Second))
	if err != nil {
		blog.Errorf("checkClusterStatus[%s] cluster[%s] failed: %v", taskID, info.Cluster.ClusterID, err)
		return err
	}

	if abnormal {
		blog.Errorf("checkClusterStatus[%s] GetTKECluster[%s] failed: abnormal", taskID, info.Cluster.ClusterID)
		retErr := fmt.Errorf("cluster[%s] status abnormal", info.Cluster.ClusterID)
		return retErr
	}

	return nil
}

// CheckTkeClusterStatusTask check cluster create status
func CheckTkeClusterStatusTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckTkeClusterStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckTkeClusterStatusTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	systemID := state.Task.CommonParams[cloudprovider.CloudSystemID.String()]

	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckTkeClusterStatusTask[%s]: GetClusterDependBasicInfo for cluster %s in task %s "+
			"step %s failed, %s", taskID, clusterID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// check cluster status
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)
	err = checkClusterStatus(ctx, dependInfo, systemID)
	if err != nil {
		blog.Errorf("CheckTkeClusterStatusTask[%s] checkClusterStatus[%s] failed: %v",
			taskID, clusterID, err)
		retErr := fmt.Errorf("checkClusterStatus[%s] timeout|abnormal", clusterID)
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckTkeClusterStatusTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}

// CheckCreateClusterNodeStatusTask check cluster node status
func CheckCreateClusterNodeStatusTask(taskID string, stepName string) error {
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("CheckCreateClusterNodeStatusTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("CheckCreateClusterNodeStatusTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// extract valid info
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	// handler logic
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("CheckCreateClusterNodeStatusTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// check cluster all nodes status
	addSuccessNodes, addFailureNodes, err := business.CheckClusterAllInstanceStatus(ctx, dependInfo)
	if err != nil {
		blog.Errorf("CheckCreateClusterNodeStatusTask[%s] CheckClusterInstanceStatus failed, %s",
			taskID, err.Error())
		retErr := fmt.Errorf("CheckClusterInstanceStatus failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("CheckCreateClusterNodeStatusTask[%s] addSuccessNodes[%v] addFailureNodes[%v]",
		taskID, addSuccessNodes, addFailureNodes)

	// update response information to task common params
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	if len(addFailureNodes) > 0 {
		state.Task.CommonParams[cloudprovider.FailedClusterNodeIDsKey.String()] = strings.Join(addFailureNodes, ",")
	}
	if len(addSuccessNodes) == 0 {
		blog.Errorf("CheckCreateClusterNodeStatusTask[%s] nodes init failed", taskID)
		retErr := fmt.Errorf("节点初始化失败, 请联系管理员")
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	state.Task.CommonParams[cloudprovider.SuccessClusterNodeIDsKey.String()] = strings.Join(addSuccessNodes, ",")

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("CheckCreateClusterNodeStatusTask[%s] task %s %s update to storage fatal", taskID, taskID, stepName)
		return err
	}

	return nil
}

// RegisterTkeClusterKubeConfigTask register cluster kubeconfig
func RegisterTkeClusterKubeConfigTask(taskID string, stepName string) error { // nolint
	start := time.Now()

	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("RegisterTkeClusterKubeConfigTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("RegisterTkeClusterKubeConfigTask[%s] task %s run current step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// inject taskID
	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]
	nodeIpList := cloudprovider.ParseNodeIpOrIdFromCommonMap(state.Task.CommonParams,
		cloudprovider.NodeIPsKey.String(), ",")

	// update connect cluster status when task retry
	connect, ok := step.Params[cloudprovider.ConnectClusterKey.String()]
	if ok && connect == icommon.True {
		step.Params[cloudprovider.ConnectClusterKey.String()] = icommon.False
	}

	// handler logic
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("RegisterTkeClusterKubeConfigTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// wait for cluster stable

	var (
		subnet = dependInfo.Cluster.GetClusterAdvanceSettings().GetClusterConnectSetting().GetSubnetId()
	)

	// inter connect && subnet empty
	if !dependInfo.Cluster.GetClusterAdvanceSettings().GetClusterConnectSetting().GetIsExtranet() && subnet == "" {
		subnet, err = getRandomSubnetFromNodes(ctx, dependInfo, nodeIpList)
		if err != nil {
			blog.Errorf("RegisterTkeClusterKubeConfigTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
				taskID, taskID, stepName, err.Error())
			retErr := fmt.Errorf("getRandomSubnetFromNodes failed, %s", err.Error())
			_ = state.UpdateStepFailure(start, stepName, retErr)
			return retErr
		}
	}

	// open tke internal kubeconfig
	err = registerTKEClusterEndpoint(ctx, dependInfo, api.ClusterEndpointConfig{
		IsExtranet: dependInfo.Cluster.GetClusterAdvanceSettings().GetClusterConnectSetting().GetIsExtranet(),
		SubnetId:   subnet,
		SecurityGroup: func() string {
			if !dependInfo.Cluster.GetClusterAdvanceSettings().GetClusterConnectSetting().GetIsExtranet() {
				return ""
			}

			return dependInfo.Cluster.GetClusterAdvanceSettings().GetClusterConnectSetting().GetSecurityGroup()
		}(),
		ExtensiveParameters: func() string {
			if !dependInfo.Cluster.GetClusterAdvanceSettings().GetClusterConnectSetting().GetIsExtranet() {
				return ""
			}

			bandWidth, _ := strconv.Atoi(
				dependInfo.Cluster.GetClusterAdvanceSettings().GetClusterConnectSetting().
					GetInternet().GetInternetMaxBandwidth())

			internet := &business.InternetConnect{
				InternetAccessible: struct {
					InternetChargeType      string `json:"InternetChargeType"`
					InternetMaxBandwidthOut int    `json:"InternetMaxBandwidthOut"`
				}{
					InternetChargeType: dependInfo.Cluster.GetClusterAdvanceSettings().
						GetClusterConnectSetting().GetInternet().GetInternetChargeType(),
					InternetMaxBandwidthOut: bandWidth,
				},
			}
			internetBytes, _ := json.Marshal(internet)

			blog.Infof("RegisterTkeClusterKubeConfigTask[%s] internet[%s]", taskID, string(internetBytes))
			return string(internetBytes)
		}(),
	})
	if err != nil {
		blog.Errorf("RegisterTkeClusterKubeConfigTask[%s] registerTKEClusterEndpoint failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("registerTKEClusterEndpoint failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("RegisterTkeClusterKubeConfigTask[%s] registerTKEClusterEndpoint success", taskID)

	// 开启admin权限, 并生成kubeconfig
	kube, err := openClusterAdminKubeConfig(ctx, dependInfo)
	if err != nil {
		blog.Errorf("RegisterTkeClusterKubeConfigTask[%s] registerTKEClusterEndpoint failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("registerTKEClusterEndpoint failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("RegisterTkeClusterKubeConfigTask[%s] openClusterAdminKubeConfig[%s] success", taskID, kube)

	// check cluster connection
	err = providerutils.CheckClusterConnect(ctx, kube)
	if err != nil {
		blog.Errorf("RegisterTkeClusterKubeConfigTask[%s] checkClusterConnect "+
			"by kubeConfig failed: %v", taskID, err)
		retErr := fmt.Errorf("checkClusterConnect %v", err)
		step.Params[cloudprovider.ConnectClusterKey.String()] = icommon.True
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// import cluster credential
	err = importClusterCredential(ctx, dependInfo,
		dependInfo.Cluster.GetClusterAdvanceSettings().GetClusterConnectSetting().GetIsExtranet(),
		false, "", kube)
	if err != nil {
		blog.Errorf("RegisterTkeClusterKubeConfigTask[%s] importClusterCredential failed: %s", taskID, err.Error())
		retErr := fmt.Errorf("importClusterCredential failed %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}
	blog.Infof("RegisterTkeClusterKubeConfigTask[%s] importClusterCredential success", taskID)

	// dynamic inject paras
	if state.Task.CommonParams == nil {
		state.Task.CommonParams = make(map[string]string)
	}
	state.Task.CommonParams[cloudprovider.DynamicClusterKubeConfigKey.String()] = kube

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("RegisterTkeClusterKubeConfigTask[%s:%s] update to storage fatal", taskID, stepName)
		return err
	}

	return nil
}

// getRandomSubnetFromNodes get random subnet from nodes
func getRandomSubnetFromNodes(
	ctx context.Context, info *cloudprovider.CloudDependBasicInfo, nodeIps []string) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	cvmClient, err := api.GetCVMClient(info.CmOption)
	if err != nil {
		blog.Errorf("getRandomSubnetFromNodes[%s] GetCVMClient failed: %v", taskID, err)
		return "", err
	}

	insList, err := cvmClient.GetInstancesByIp(nodeIps)
	if err != nil {
		blog.Errorf("getRandomSubnetFromNodes[%s] GetInstancesByIp failed: %v", taskID, err)
		return "", err
	}

	// filter vpc subnets
	var (
		subnetMap  = make(map[string]struct{}, 0)
		subnetList = make([]string, 0)
	)
	for i := range insList {
		_, ok := subnetMap[*insList[i].VirtualPrivateCloud.SubnetId]
		if !ok {
			subnetMap[*insList[i].VirtualPrivateCloud.SubnetId] = struct{}{}
			subnetList = append(subnetList, *insList[i].VirtualPrivateCloud.SubnetId)
		}
	}
	blog.Infof("getRandomSubnetFromNodes[%s] success[%+v]", taskID, subnetList)

	rand.Seed(time.Now().Unix())                       // nolint
	return subnetList[rand.Intn(len(subnetList))], nil // nolint
}

// openClusterAdminKubeConfig open account cluster admin perm
func openClusterAdminKubeConfig(ctx context.Context, info *cloudprovider.CloudDependBasicInfo) (string, error) {
	taskID := cloudprovider.GetTaskIDFromContext(ctx)

	tkeCli, err := api.NewTkeClient(info.CmOption)
	if err != nil {
		blog.Errorf("openClusterAdminKubeConfig[%s] NewTkeClient failed: %v", taskID, err)
		return "", err
	}

	// get qcloud account clusterAdminRole
	err = tkeCli.AcquireClusterAdminRole(info.Cluster.SystemID)
	if err != nil {
		blog.Errorf("openClusterAdminKubeConfig[%s] AcquireClusterAdminRole failed: %v", taskID, err)
		return "", err
	}

	kube, err := tkeCli.GetTKEClusterKubeConfig(info.Cluster.SystemID,
		info.Cluster.GetClusterAdvanceSettings().GetClusterConnectSetting().GetIsExtranet())
	if err != nil {
		blog.Errorf("openClusterAdminKubeConfig[%s] GetTKEClusterKubeConfig failed: %v", taskID, err)
		return "", err
	}

	return kube, nil
}

// UpdateCreateClusterDBInfoTask update cluster DB info
func UpdateCreateClusterDBInfoTask(taskID string, stepName string) error {
	start := time.Now()
	// get task and task current step
	state, step, err := cloudprovider.GetTaskStateAndCurrentStep(taskID, stepName)
	if err != nil {
		return err
	}
	// previous step successful when retry task
	if step == nil {
		blog.Infof("UpdateCreateClusterDBInfoTask[%s]: current step[%s] successful and skip", taskID, stepName)
		return nil
	}
	blog.Infof("UpdateCreateClusterDBInfoTask[%s]: task %s run step %s, system: %s, old state: %s, params %v",
		taskID, taskID, stepName, step.System, step.Status, step.Params)

	// step login started here
	clusterID := step.Params[cloudprovider.ClusterIDKey.String()]
	cloudID := step.Params[cloudprovider.CloudIDKey.String()]

	ctx := cloudprovider.WithTaskIDForContext(context.Background(), taskID)

	// handler logic
	dependInfo, err := cloudprovider.GetClusterDependBasicInfo(cloudprovider.GetBasicInfoReq{
		ClusterID: clusterID,
		CloudID:   cloudID,
	})
	if err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s] GetClusterDependBasicInfo in task %s step %s failed, %s",
			taskID, taskID, stepName, err.Error())
		retErr := fmt.Errorf("get cloud/project information failed, %s", err.Error())
		_ = state.UpdateStepFailure(start, stepName, retErr)
		return retErr
	}

	// 后面会去除密码
	bkBizID, _ := strconv.Atoi(dependInfo.Cluster.GetBusinessID())
	if dependInfo.Cluster.GetClusterBasicSettings().GetModule().GetMasterModuleID() != "" {
		bkModuleID, _ := strconv.Atoi(dependInfo.Cluster.GetClusterBasicSettings().GetModule().GetMasterModuleID())
		dependInfo.Cluster.
			GetClusterBasicSettings().
			GetModule().MasterModuleName = cloudprovider.GetModuleName(bkBizID, bkModuleID)
	}
	if dependInfo.Cluster.GetClusterBasicSettings().GetModule().GetWorkerModuleID() != "" {
		bkModuleID, _ := strconv.Atoi(dependInfo.Cluster.GetClusterBasicSettings().GetModule().GetWorkerModuleID())
		dependInfo.Cluster.
			GetClusterBasicSettings().
			GetModule().WorkerModuleName = cloudprovider.GetModuleName(bkBizID, bkModuleID)
	}
	_ = cloudprovider.UpdateCluster(dependInfo.Cluster)

	// sync clusterData to pass-cc
	providerutils.SyncClusterInfoToPassCC(taskID, dependInfo.Cluster)

	// sync cluster perms
	providerutils.AuthClusterResourceCreatorPerm(ctx, dependInfo.Cluster.ClusterID,
		dependInfo.Cluster.ClusterName, dependInfo.Cluster.Creator)

	// update step
	if err = state.UpdateStepSucc(start, stepName); err != nil {
		blog.Errorf("UpdateCreateClusterDBInfoTask[%s] task %s %s update to storage fatal",
			taskID, taskID, stepName)
		return err
	}

	return nil
}
