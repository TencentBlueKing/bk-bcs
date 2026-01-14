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

// Package cluster cluster operate
package cluster

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"
	"google.golang.org/protobuf/types/known/wrapperspb"

	clustermgr "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/clustermanager"
	projectrmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/projectmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/utils"
)

// ClusterAction cluster action interface
type ClusterAction interface { // nolint
	ListCluster(ctx context.Context, req *types.ListClusterReq) (*types.ListClusterResp, error)
	GetCluster(ctx context.Context, req *types.GetClusterReq) (*types.GetClusterResp, error)
	GetClusterOverview(ctx context.Context, req *types.GetClusterOverviewReq) (*types.GetClusterOverviewResp, error)
	GetClusterBasicInfo(ctx context.Context, req *types.GetClusterBasicInfoReq) (*types.GetClusterBasicInfoResp, error)
	GetClusterNetworkConfig(ctx context.Context, req *types.GetClusterNetworkConfigReq) (
		*types.GetClusterNetworkConfigResp, error)
	GetClusterControlPlaneConfig(ctx context.Context, req *types.GetClusterControlPlaneConfigReq) (
		*types.GetClusterControlPlaneConfigResp, error)
	UpdateClusterBasicInfo(ctx context.Context, req *types.UpdateClusterBasicInfoReq) (bool, error)
	UpdateClusterNetworkConfig(ctx context.Context, req *types.UpdateClusterNetworkConfigReq) (bool, error)
	UpdateClusterControlPlaneConfig(ctx context.Context, req *types.UpdateClusterControlPlaneConfigReq) (bool, error)
	UpdateClusterOperator(ctx context.Context, req *types.UpdateClusterOperatorReq) (bool, error)
	UpdateClusterProjectBusiness(ctx context.Context, req *types.UpdateClusterProjectBusinessReq) (bool, error)
	AddClusterCidr(ctx context.Context, req *types.AddClusterCidrReq) (bool, error)
	AddSubnetToCluster(ctx context.Context, req *types.AddSubnetToClusterReq) (bool, error)
}

// Action action for cloud vpc
type Action struct{}

// NewClusterAction new cluster action
func NewClusterAction() ClusterAction {
	return &Action{}
}

// ListCluster list cluster
func (a *Action) ListCluster(ctx context.Context, req *types.ListClusterReq) (*types.ListClusterResp, error) {
	/*projects, err := projectrmgr.ListAllProject(ctx)
	if err != nil {
		return nil, utils.SystemError(err)
	}

	businesses, err := cmdb.GetCmdbClient().GetBusiness()
	if err != nil {
		return nil, utils.SystemError(err)
	}

	clusterData, err := clustermgr.ListCluster(ctx, &clustermanager.ListClusterV2Req{
		ClusterID:   req.ClusterID,
		ClusterName: req.ClusterName,
		ProjectID:   req.ProjectID,
		BusinessID:  req.BusinessID,
		Environment: req.Environment,
		Status:      req.Status,
		Provider:    req.Provider,
		VpcID:       req.VpcID,
		SystemID:    req.SystemID,
		Creator:     req.Creator,
		ManageType:  req.ManageType,
		All:         req.All,
		Page:        req.Page,
		Limit:       req.Limit,
		Sort:        req.Sort,
		Order:       req.Order,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*types.ListClusterData, 0)
	for _, cluster := range clusterData.Results {
		projectCode := ""
		projectName := ""

		for _, project := range projects {
			if project.ProjectID == cluster.ProjectID {
				projectCode = project.ProjectCode
				projectName = project.Name
				break
			}
		}
		result = append(result, &types.ListClusterData{
			ClusterID:   cluster.ClusterID,
			ClusterName: cluster.ClusterName,
			Provider:    cluster.Provider,
			ProjectID:   cluster.ProjectID,
			ProjectName: projectName,
			ProjectCode: projectCode,
			BusinessID:  cluster.BusinessID,
			BusinessName: func() string {
				for _, business := range *businesses {
					if fmt.Sprint(business.BkBizID) == cluster.BusinessID {
						return business.BkBizName
					}
				}
				return ""
			}(),
			Environment:   cluster.Environment,
			Creator:       cluster.Creator,
			ManageType:    cluster.ManageType,
			Status:        cluster.Status,
			BizMaintainer: cluster.BizMaintainer,
			Link: func() string {
				return fmt.Sprintf("%s/bcs/projects/%s/clusters?clusterId=%s",
					config.G.BCS.Server, projectCode, cluster.ClusterID)
			}(),
		})
	}

	return &types.ListClusterResp{
		Total:   uint32(clusterData.Total),
		Results: result,
	}, nil*/
	return nil, nil
}

// GetCluster get cluster info
func (a *Action) GetCluster(ctx context.Context, req *types.GetClusterReq) (*types.GetClusterResp, error) {
	cluster, err := clustermgr.GetCluster(ctx, req.ClusterID, req.ProjectID)
	if err != nil {
		return nil, utils.SystemError(err)
	}

	resp := &types.GetClusterResp{
		ClusterID:       cluster.ClusterID,
		ClusterName:     cluster.ClusterName,
		Provider:        cluster.Provider,
		Region:          cluster.Region,
		VpcID:           cluster.VpcID,
		ProjectID:       cluster.ProjectID,
		BusinessID:      cluster.BusinessID,
		Environment:     cluster.Environment,
		EngineType:      cluster.EngineType,
		ClusterType:     cluster.ClusterType,
		Creator:         cluster.Creator,
		CreateTime:      cluster.CreateTime,
		UpdateTime:      cluster.UpdateTime,
		ManageType:      cluster.ManageType,
		Status:          cluster.Status,
		Updater:         cluster.Updater,
		Description:     cluster.Description,
		ClusterCategory: cluster.ClusterCategory,
		Label:           cluster.Labels,
		SystemID:        cluster.SystemID,
		NetworkType:     cluster.NetworkType,
		ModuleID:        cluster.ModuleID,
		IsCommonCluster: cluster.IsCommonCluster,
		IsShared:        cluster.IsShared,
		IsMixed:         cluster.IsMixed,
		CloudAccountID:  cluster.CloudAccountID,
	}

	if cluster.ClusterBasicSettings != nil {
		resp.ClusterBasicSettings = &types.ClusterBasicSetting{
			OS:                        cluster.ClusterBasicSettings.OS,
			Version:                   cluster.ClusterBasicSettings.Version,
			ClusterTags:               cluster.ClusterBasicSettings.ClusterTags,
			VersionName:               cluster.ClusterBasicSettings.VersionName,
			SubnetID:                  cluster.ClusterBasicSettings.SubnetID,
			ClusterLevel:              cluster.ClusterBasicSettings.ClusterLevel,
			IsAutoUpgradeClusterLevel: cluster.ClusterBasicSettings.IsAutoUpgradeClusterLevel,
		}
		if cluster.ClusterBasicSettings.Area != nil {
			resp.ClusterBasicSettings.Area = &types.CloudArea{
				BkCloudID:   cluster.ClusterBasicSettings.Area.BkCloudID,
				BkCloudName: cluster.ClusterBasicSettings.Area.BkCloudName,
			}
		}
		if cluster.ClusterBasicSettings.Module != nil {
			resp.ClusterBasicSettings.Module = &types.ClusterModule{
				MasterModuleID:   cluster.ClusterBasicSettings.Module.MasterModuleID,
				MasterModuleName: cluster.ClusterBasicSettings.Module.MasterModuleName,
				WorkerModuleID:   cluster.ClusterBasicSettings.Module.WorkerModuleID,
				WorkerModuleName: cluster.ClusterBasicSettings.Module.WorkerModuleName,
			}
		}
		if cluster.ClusterBasicSettings.UpgradePolicy != nil {
			resp.ClusterBasicSettings.UpgradePolicy = &types.UpgradePolicy{
				SupportType: cluster.ClusterBasicSettings.UpgradePolicy.SupportType,
			}
		}
	}

	return resp, nil
}

// GetClusterOverview get cluster overview
func (a *Action) GetClusterOverview(ctx context.Context, req *types.GetClusterOverviewReq) (
	*types.GetClusterOverviewResp, error) {
	// url := fmt.Sprintf("%s/bcsapi/v4/monitor/api/metrics/projects/%s/clusters/%s/overview",
	// 	config.G.BCS.Server, req.ProjectCode, req.ClusterID)
	// resp, err := component.GetClient().R().
	// 	SetContext(ctx).
	// 	SetHeaders(GetLaneIDByCtx(ctx)).
	// 	SetAuthToken(config.G.BCS.Token).
	// 	Get(url)

	// if err != nil {
	// 	blog.Errorf("list clusters error, %s", err.Error())
	// 	return nil, err
	// }

	// var result *types.GetClusterOverviewResp
	// if err = component.UnmarshalBKResult(resp, result); err != nil {
	// 	blog.Errorf("unmarshal clusters error, %s", err.Error())
	// 	return nil, err
	// }

	return nil, nil
}

// GetClusterBasicInfo get cluster basic info
func (a *Action) GetClusterBasicInfo(ctx context.Context, req *types.GetClusterBasicInfoReq) (
	*types.GetClusterBasicInfoResp, error) {
	cluster, err := clustermgr.GetCluster(ctx, req.ClusterID, req.ProjectID)
	if err != nil {
		return nil, utils.SystemError(err)
	}

	projects, err := projectrmgr.ListAllProject(ctx)
	if err != nil {
		return nil, utils.SystemError(err)
	}

	businesses, err := cmdb.GetCmdbClient().GetBusiness()
	if err != nil {
		return nil, utils.SystemError(err)
	}

	result := &types.GetClusterBasicInfoResp{
		ClusterID:   cluster.ClusterID,
		ClusterName: cluster.ClusterName,
		Status:      cluster.Status,
		ProjectID:   cluster.ProjectID,
		ProjectName: func() string {
			for _, project := range projects {
				if project.ProjectID == cluster.ProjectID {
					return project.Name
				}
			}
			return ""
		}(),
		BusinessID: cluster.BusinessID,
		BusinessName: func() string {
			for _, business := range *businesses {
				if fmt.Sprint(business.BkBizID) == cluster.BusinessID {
					return business.BkBizName
				}
			}
			return ""
		}(),
		Description:     cluster.Description,
		Provider:        cluster.Provider,
		ManageType:      cluster.ManageType,
		IsShared:        cluster.IsShared,
		IsMixed:         cluster.IsMixed,
		Labels:          cluster.Labels,
		Region:          cluster.Region,
		Environment:     cluster.Environment,
		ClusterCategory: cluster.ClusterCategory,
		Creator:         cluster.Creator,
		CreateTime:      cluster.CreateTime,
		UpdateTime:      cluster.UpdateTime,
	}

	if cluster.ClusterBasicSettings != nil {
		result.ClusterBasicSettings = &types.ClusterBasicSetting{
			Version: cluster.ClusterBasicSettings.Version,
		}
		if cluster.ClusterBasicSettings.Area != nil {
			result.ClusterBasicSettings.Area = &types.CloudArea{
				BkCloudID:   cluster.ClusterBasicSettings.Area.BkCloudID,
				BkCloudName: cluster.ClusterBasicSettings.Area.BkCloudName,
			}
		}
	}
	if cluster.ClusterAdvanceSettings != nil {
		result.ClusterAdvanceSettings = &types.ClusterAdvanceSetting{
			ContainerRuntime: cluster.ClusterAdvanceSettings.ContainerRuntime,
			RuntimeVersion:   cluster.ClusterAdvanceSettings.RuntimeVersion,
		}
	}
	if cluster.SharedRanges != nil {
		result.SharedRanges = &types.SharedClusterRanges{
			Bizs:             cluster.SharedRanges.Bizs,
			ProjectIdOrCodes: cluster.SharedRanges.ProjectIdOrCodes,
		}
	}

	return result, nil
}

// GetClusterNetworkConfig get cluster network config
func (a *Action) GetClusterNetworkConfig(ctx context.Context, req *types.GetClusterNetworkConfigReq) (
	*types.GetClusterNetworkConfigResp, error) {
	cluster, err := clustermgr.GetCluster(ctx, req.ClusterID, "")
	if err != nil {
		return nil, utils.SystemError(err)
	}

	result := &types.GetClusterNetworkConfigResp{
		Region:      cluster.Region,
		VpcID:       cluster.VpcID,
		NetworkType: cluster.NetworkType,
		Subnets:     make([]*types.Subnet, 0),
		NetworkSettings: &types.NetworkSetting{
			ClusterIPv4CIDR: cluster.NetworkSettings.ClusterIPv4CIDR,
			ServiceIPv4CIDR: cluster.NetworkSettings.ServiceIPv4CIDR,
			MaxNodePodNum:   cluster.NetworkSettings.MaxNodePodNum,
			MaxServiceNum:   cluster.NetworkSettings.MaxServiceNum,
			CidrStep:        cluster.NetworkSettings.CidrStep,
			EnableVPCCni:    cluster.NetworkSettings.EnableVPCCni,
			IsStaticIpMode:  cluster.NetworkSettings.IsStaticIpMode,
			NetworkMode:     cluster.NetworkSettings.NetworkMode,
		},
	}

	if cluster.NetworkSettings.EnableVPCCni {
		subnets, err := clustermgr.ListCloudSubnets(ctx, &clustermanager.ListCloudSubnetsRequest{
			VpcID:     cluster.VpcID,
			Region:    cluster.Region,
			AccountID: cluster.CloudAccountID,
		})
		if err != nil {
			return nil, utils.SystemError(err)
		}

		for _, id := range cluster.NetworkSettings.SubnetSource.Existed.Ids {
			for _, subnet := range subnets {
				if id == subnet.SubnetID {
					result.Subnets = append(result.Subnets, &types.Subnet{
						VpcID:                   subnet.VpcID,
						SubnetID:                subnet.SubnetID,
						SubnetName:              subnet.SubnetName,
						CidrRange:               subnet.CidrRange,
						Zone:                    subnet.Zone,
						AvailableIPAddressCount: subnet.AvailableIPAddressCount,
						ZoneName:                subnet.ZoneName,
					})
				}
			}
		}
	}

	if cluster.ClusterAdvanceSettings != nil {
		result.ClusterAdvanceSettings = &types.ClusterAdvanceSetting{
			IPVS:        cluster.ClusterAdvanceSettings.IPVS,
			NetworkType: cluster.ClusterAdvanceSettings.NetworkType,
		}
	}

	return result, nil
}

// GetClusterControlPlaneConfig get cluster control plane config
func (a *Action) GetClusterControlPlaneConfig(ctx context.Context, req *types.GetClusterControlPlaneConfigReq) (
	*types.GetClusterControlPlaneConfigResp, error) {
	cluster, err := clustermgr.GetCluster(ctx, req.ClusterID, "")
	if err != nil {
		return nil, utils.SystemError(err)
	}

	result := &types.GetClusterControlPlaneConfigResp{
		ManageType:                cluster.ManageType,
		ClusterLevel:              cluster.ClusterBasicSettings.ClusterLevel,
		IsAutoUpgradeClusterLevel: cluster.ClusterBasicSettings.IsAutoUpgradeClusterLevel,
		Module: &types.ClusterModule{
			MasterModuleID:   cluster.ClusterBasicSettings.Module.MasterModuleID,
			MasterModuleName: cluster.ClusterBasicSettings.Module.MasterModuleName,
			WorkerModuleID:   cluster.ClusterBasicSettings.Module.WorkerModuleID,
			WorkerModuleName: cluster.ClusterBasicSettings.Module.WorkerModuleName,
		},
	}

	if cluster.ClusterAdvanceSettings != nil {
		if cluster.ClusterAdvanceSettings.ClusterConnectSetting != nil {
			result.SecurityGroup = cluster.ClusterAdvanceSettings.ClusterConnectSetting.SecurityGroup
		}
	}
	if len(cluster.Master) > 0 {
		result.Master = make(map[string]*types.Node)
		for k, node := range cluster.Master {
			result.Master[k] = &types.Node{
				NodeID:         node.NodeID,
				InnerIP:        node.InnerIP,
				InstanceType:   node.InstanceType,
				CPU:            node.CPU,
				Mem:            node.Mem,
				GPU:            node.GPU,
				Status:         node.Status,
				ZoneID:         node.ZoneID,
				NodeGroupID:    node.NodeGroupID,
				ClusterID:      node.ClusterID,
				VPC:            node.VPC,
				Region:         node.Region,
				Passwd:         node.Passwd,
				Zone:           node.Zone,
				DeviceID:       node.DeviceID,
				NodeTemplateID: node.NodeTemplateID,
				NodeType:       node.NodeType,
				NodeName:       node.NodeName,
				InnerIPv6:      node.InnerIPv6,
				ZoneName:       node.ZoneName,
				TaskID:         node.TaskID,
				FailedReason:   node.FailedReason,
				ChargeType:     node.ChargeType,
				DataDiskNum:    node.DataDiskNum,
				IsGpuNode:      node.IsGpuNode,
			}
		}
	}

	return result, nil
}

// UpdateClusterBasicInfo update cluster basic info
func (a *Action) UpdateClusterBasicInfo(ctx context.Context, req *types.UpdateClusterBasicInfoReq) (bool, error) {
	cluster, err := clustermgr.GetCluster(ctx, req.ClusterID, "")
	if err != nil {
		return false, utils.SystemError(err)
	}

	updateReq := &clustermanager.UpdateClusterReq{
		ClusterID: req.ClusterID,
	}

	if req.ClusterName != "" {
		updateReq.ClusterName = req.ClusterName
	}
	if req.Status != "" {
		updateReq.Status = req.Status
	}
	if req.ProjectID != "" {
		updateReq.ProjectID = req.ProjectID
	}
	if req.BusinessID != "" {
		updateReq.BusinessID = req.BusinessID
	}
	if cluster.IsShared != req.IsShared {
		updateReq.IsShared = &wrapperspb.BoolValue{Value: req.IsShared}
	}
	if cluster.IsMixed != req.IsMixed {
		updateReq.IsMixed = &wrapperspb.BoolValue{Value: req.IsMixed}
	}
	if req.Description != "" {
		updateReq.Description = &wrapperspb.StringValue{Value: req.Description}
	}
	if req.Environment != "" {
		updateReq.Environment = req.Environment
	}
	if req.Labels2 != nil {
		updateReq.Labels2 = &clustermanager.MapStruct{Values: req.Labels2.Values}
	}
	if req.SharedRanges != nil {
		updateReq.SharedRanges = &clustermanager.SharedClusterRanges{
			Bizs:             req.SharedRanges.Bizs,
			ProjectIdOrCodes: req.SharedRanges.ProjectIdOrCodes,
		}
	}
	if req.ClusterBasicSettings != nil {
		updateReq.ClusterBasicSettings = &clustermanager.ClusterBasicSetting{
			OS:                        cluster.ClusterBasicSettings.OS,
			Version:                   cluster.ClusterBasicSettings.Version,
			ClusterTags:               cluster.ClusterBasicSettings.ClusterTags,
			VersionName:               cluster.ClusterBasicSettings.VersionName,
			SubnetID:                  cluster.ClusterBasicSettings.SubnetID,
			ClusterLevel:              cluster.ClusterBasicSettings.ClusterLevel,
			IsAutoUpgradeClusterLevel: cluster.ClusterBasicSettings.IsAutoUpgradeClusterLevel,
			Area: func() *clustermanager.CloudArea {
				if req.ClusterBasicSettings.Area != nil {
					return &clustermanager.CloudArea{
						BkCloudID:   req.ClusterBasicSettings.Area.BkCloudID,
						BkCloudName: req.ClusterBasicSettings.Area.BkCloudName,
					}
				}
				return nil
			}(),
			Module: func() *clustermanager.ClusterModule {
				if cluster.ClusterBasicSettings.Module != nil {
					return &clustermanager.ClusterModule{
						MasterModuleID:   cluster.ClusterBasicSettings.Module.MasterModuleID,
						MasterModuleName: cluster.ClusterBasicSettings.Module.MasterModuleName,
						WorkerModuleID:   cluster.ClusterBasicSettings.Module.WorkerModuleID,
						WorkerModuleName: cluster.ClusterBasicSettings.Module.WorkerModuleName,
					}
				}
				return nil
			}(),
			UpgradePolicy: func() *clustermanager.UpgradePolicy {
				if cluster.ClusterBasicSettings.UpgradePolicy != nil {
					return &clustermanager.UpgradePolicy{
						SupportType: cluster.ClusterBasicSettings.UpgradePolicy.SupportType,
					}
				}
				return nil
			}(),
		}
	}

	result, err := clustermgr.UpdateCluster(ctx, updateReq)
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}

// UpdateClusterNetworkConfig update cluster network config
func (a *Action) UpdateClusterNetworkConfig(ctx context.Context, req *types.UpdateClusterNetworkConfigReq) (bool, error) {
	result, err := clustermgr.UpdateCluster(ctx, &clustermanager.UpdateClusterReq{
		ClusterID: req.ClusterID,
		NetworkSettings: func() *clustermanager.NetworkSetting {
			if req.NetworkSettings != nil {
				return &clustermanager.NetworkSetting{
					ClusterIPv4CIDR: req.NetworkSettings.ClusterIPv4CIDR,
					ServiceIPv4CIDR: req.NetworkSettings.ServiceIPv4CIDR,
					MaxNodePodNum:   req.NetworkSettings.MaxNodePodNum,
					MaxServiceNum:   req.NetworkSettings.MaxServiceNum,
					EnableVPCCni:    req.NetworkSettings.EnableVPCCni,
					EniSubnetIDs:    req.NetworkSettings.EniSubnetIDs,
					SubnetSource: func() *clustermanager.SubnetSource {
						if req.NetworkSettings.SubnetSource != nil {
							newSubnets := make([]*clustermanager.NewSubnet, 0)
							for _, subnet := range req.NetworkSettings.SubnetSource.New {
								newSubnets = append(newSubnets, &clustermanager.NewSubnet{
									Zone:  subnet.Zone,
									Mask:  subnet.Mask,
									IpCnt: subnet.IpCnt,
								})
							}
							return &clustermanager.SubnetSource{
								New: newSubnets,
								Existed: func() *clustermanager.ExistedSubnetIDs {
									if req.NetworkSettings.SubnetSource.Existed != nil {
										return &clustermanager.ExistedSubnetIDs{
											Ids: req.NetworkSettings.SubnetSource.Existed.Ids,
										}
									}
									return nil
								}(),
							}
						}
						return nil
					}(),
					IsStaticIpMode:      req.NetworkSettings.IsStaticIpMode,
					ClaimExpiredSeconds: req.NetworkSettings.ClaimExpiredSeconds,
					MultiClusterCIDR:    req.NetworkSettings.MultiClusterCIDR,
					CidrStep:            req.NetworkSettings.CidrStep,
					ClusterIpType:       req.NetworkSettings.ClusterIpType,
					ClusterIPv6CIDR:     req.NetworkSettings.ClusterIPv6CIDR,
					ServiceIPv6CIDR:     req.NetworkSettings.ServiceIPv6CIDR,
					Status:              req.NetworkSettings.Status,
					NetworkMode:         req.NetworkSettings.NetworkMode,
				}
			}
			return nil
		}(),
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}

// UpdateClusterControlPlaneConfig update cluster control plane config
func (a *Action) UpdateClusterControlPlaneConfig(ctx context.Context, req *types.UpdateClusterControlPlaneConfigReq) (bool, error) {
	result, err := clustermgr.UpdateCluster(ctx, &clustermanager.UpdateClusterReq{
		ClusterID: req.ClusterID,
		Master:    req.Master,
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}

// UpdateClusterOperator update cluster operator
func (a *Action) UpdateClusterOperator(ctx context.Context, req *types.UpdateClusterOperatorReq) (bool, error) {
	result, err := clustermgr.UpdateCluster(ctx, &clustermanager.UpdateClusterReq{
		ClusterID: req.ClusterID,
		Creator:   req.Creator,
		Updater:   req.Updater,
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}

// UpdateClusterProjectBusiness update cluster project business
func (a *Action) UpdateClusterProjectBusiness(ctx context.Context,
	req *types.UpdateClusterProjectBusinessReq) (bool, error) {
	result, err := clustermgr.UpdateCluster(ctx, &clustermanager.UpdateClusterReq{
		ClusterID:  req.ClusterID,
		ProjectID:  req.ProjectID,
		BusinessID: req.BusinessID,
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}

func (a *Action) AddClusterCidr(ctx context.Context, req *types.AddClusterCidrReq) (bool, error) {
	/*result, err := clustermgr.AddClusterCidr(ctx, &clustermanager.AddClusterCidrReq{
		ClusterID: req.ClusterID,
		Cidrs:     req.Cidrs,
		Operator:  req.Operator,
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil*/
	return false, nil
}

// AddSubnetToCluster add subnet to cluster
func (a *Action) AddSubnetToCluster(ctx context.Context, req *types.AddSubnetToClusterReq) (bool, error) {
	cluster, err := clustermgr.GetCluster(ctx, req.ClusterID, "")
	if err != nil {
		return false, utils.SystemError(err)
	}

	newSubnets := make([]*clustermanager.NewSubnet, 0)
	for _, subnet := range req.NewSubnets {
		newSubnets = append(newSubnets, &clustermanager.NewSubnet{
			Zone:  subnet.Zone,
			Mask:  subnet.Mask,
			IpCnt: subnet.IpCnt,
		})
	}

	result, err := clustermgr.AddSubnetToCluster(ctx, &clustermanager.AddSubnetToClusterReq{
		ClusterID: req.ClusterID,
		Subnet: &clustermanager.SubnetSource{
			Existed: &clustermanager.ExistedSubnetIDs{
				Ids: cluster.NetworkSettings.EniSubnetIDs,
			},
			New: newSubnets,
		},
		Operator: req.Operator,
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}
