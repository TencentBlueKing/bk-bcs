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

// Package project project operate
package project

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	"github.com/golang/protobuf/ptypes/wrappers"

	projectrmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/projectmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/utils"
)

// QuotaAction project action interface
type QuotaAction interface { // nolint
	CreateProjectQuota(ctx context.Context, req *types.CreateProjectQuotaReq) (bool, error)
	GetProjectQuota(ctx context.Context, quotaId string) (*types.GetProjectQuotaData, error)
	UpdateProjectQuota(ctx context.Context, req *types.UpdateProjectQuotaReq) (bool, error)
	ScaleUpProjectQuota(ctx context.Context, req *types.ScaleUpProjectQuotaReq) (bool, error)
	ScaleDownProjectQuota(ctx context.Context, req *types.ScaleDownProjectQuotaReq) (bool, error)
	DeleteProjectQuota(ctx context.Context, req *types.DeleteProjectQuotaReq) (bool, error)
	ListProjectQuotasV2(ctx context.Context, req *types.ListProjectQuotasV2Req) (*types.ListProjectQuotasData, error)
	GetProjectQuotasStatistics(ctx context.Context, req *types.GetProjectQuotasStatisticsReq) (
		*types.ProjectQuotasStatisticsData, error)
}

// Action action for project quota
type Action struct{}

// NewQuotaAction new project quota action
func NewQuotaAction() QuotaAction {
	return &Action{}
}

// CreateProjectQuota create project quota
func (a *Action) CreateProjectQuota(ctx context.Context, req *types.CreateProjectQuotaReq) (bool, error) { // nolint
	result, err := projectrmgr.CreateProjectQuota(ctx, &bcsproject.CreateProjectQuotaRequest{
		QuotaName:              req.QuotaName,
		ProjectID:              req.ProjectID,
		ProjectCode:            req.ProjectCode,
		ClusterId:              req.ClusterId,
		ClusterName:            req.ClusterName,
		NameSpace:              req.NameSpace,
		BusinessID:             req.BusinessID,
		BusinessName:           req.BusinessName,
		Description:            req.Description,
		QuotaType:              req.QuotaType,
		Provider:               req.Provider,
		Quota:                  conver2ProjectQuotaResource(req.Quota),
		Labels:                 req.Labels,
		Annotations:            req.Annotations,
		QuotaAttr:              conver2ProjectQuotaAttr(req.QuotaAttr),
		QuotaSharedEnabled:     &wrappers.BoolValue{Value: req.QuotaSharedEnabled},
		QuotaSharedProjectList: conver2ProjectQuotaSharedProject(req.QuotaSharedProjectList),
		SkipItsmApproval:       &wrappers.BoolValue{Value: req.SkipItsmApproval},
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}

// GetProjectQuota 获取项目资源额度
func (a *Action) GetProjectQuota(ctx context.Context, quotaId string) (*types.GetProjectQuotaData, error) {
	rsp, err := projectrmgr.GetProjectQuota(ctx, quotaId)
	if err != nil {
		return nil, utils.SystemError(err)
	}

	return &types.GetProjectQuotaData{
		Data: conver2TypeQuota(rsp.Data),
		Task: rsp.Task,
	}, nil
}

// UpdateProjectQuota 更新项目资源额度
func (a *Action) UpdateProjectQuota(ctx context.Context, req *types.UpdateProjectQuotaReq) (bool, error) {
	result, err := projectrmgr.UpdateProjectQuota(ctx, &bcsproject.UpdateProjectQuotaRequest{
		QuotaId:            req.QuotaId,
		Name:               req.Name,
		Quota:              conver2ProjectQuotaResource(req.Quota),
		Updater:            req.Updater,
		Labels:             req.Labels,
		Annotations:        req.Annotations,
		QuotaAttr:          conver2ProjectQuotaAttr(req.QuotaAttr),
		QuotaSharedEnabled: &wrappers.BoolValue{Value: req.QuotaSharedEnabled},
		QuotaSharedProjectList: &bcsproject.QuotaSharedProjectList{
			Values: conver2ProjectQuotaSharedProject(req.QuotaSharedProjectList),
		},
	})
	if err != nil {
		return false, utils.SystemError(err)
	}
	return result, nil
}

// ScaleUpProjectQuota 上升项目资源额度
func (a *Action) ScaleUpProjectQuota(ctx context.Context, req *types.ScaleUpProjectQuotaReq) (bool, error) {
	result, err := projectrmgr.ScaleUpProjectQuota(ctx, &bcsproject.ScaleUpProjectQuotaRequest{
		QuotaId:          req.QuotaId,
		Quota:            conver2ProjectQuotaResource(req.Quota),
		Updater:          req.Updater,
		SkipItsmApproval: &wrappers.BoolValue{Value: req.SkipItsmApproval},
	})
	if err != nil {
		return false, utils.SystemError(err)
	}
	return result, nil
}

// ScaleDownProjectQuota 下降项目资源额度
func (a *Action) ScaleDownProjectQuota(ctx context.Context, req *types.ScaleDownProjectQuotaReq) (bool, error) {
	result, err := projectrmgr.ScaleDownProjectQuota(ctx, &bcsproject.ScaleDownProjectQuotaRequest{
		QuotaId:          req.QuotaId,
		Quota:            conver2ProjectQuotaResource(req.Quota),
		Updater:          req.Updater,
		SkipItsmApproval: &wrappers.BoolValue{Value: req.SkipItsmApproval},
	})
	if err != nil {
		return false, utils.SystemError(err)
	}
	return result, nil
}

// DeleteProjectQuota 删除项目资源额度
func (a *Action) DeleteProjectQuota(ctx context.Context, req *types.DeleteProjectQuotaReq) (bool, error) {
	result, err := projectrmgr.DeleteProjectQuota(ctx, &bcsproject.DeleteProjectQuotaRequest{
		QuotaId:          req.QuotaId,
		OnlyDeleteInfo:   req.OnlyDeleteInfo,
		SkipItsmApproval: &wrappers.BoolValue{Value: req.SkipItsmApproval},
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}

// ListProjectQuotasV2 获取项目资源额度列表
func (a *Action) ListProjectQuotasV2(ctx context.Context, req *types.ListProjectQuotasV2Req) ( // nolint
	*types.ListProjectQuotasData, error) {
	data, err := projectrmgr.ListProjectQuotasV2(ctx, &bcsproject.ListProjectQuotasV2Request{
		QuotaId:         req.QuotaId,
		QuotaName:       req.QuotaName,
		ProjectIDOrCode: req.ProjectIDOrCode,
		BusinessID:      req.BusinessID,
		QuotaType:       req.QuotaType,
		Provider:        req.Provider,
		Page:            req.Page,
		Limit:           req.Limit,
	})
	if err != nil {
		return nil, utils.SystemError(err)
	}

	results := make([]*types.ProjectQuotaData, 0)
	for _, item := range data.Results {
		results = append(results, &types.ProjectQuotaData{
			QuotaId:    item.QuotaId,
			QuotaName:  item.QuotaName,
			IsDeleted:  item.IsDeleted,
			QuotaType:  item.QuotaType,
			Quota:      conver2TypeQuotaResource(item.Quota),
			Status:     item.Status,
			UpdateTime: item.UpdateTime,
			Updater:    item.Updater,
		})
	}

	return &types.ListProjectQuotasData{
		Total:   data.Total,
		Results: results,
	}, nil
}

// GetProjectQuotasUsage 获取项目资源额度使用情况
func (a *Action) GetProjectQuotasUsage(ctx context.Context, quotaId string) ( // nolint
	*types.GetProjectQuotasUsageData, error) {
	data, err := projectrmgr.GetProjectQuotasUsage(ctx, quotaId)
	if err != nil {
		return nil, utils.SystemError(err)
	}

	return &types.GetProjectQuotasUsageData{
		Quota:        conver2TypeQuota(data.Quota),
		Region:       data.Region,
		InstanceType: data.InstanceType,
		QuotaUsage: &types.ZoneResourceUsage{
			Zone:  data.QuotaUsage.Zone,
			Quota: data.QuotaUsage.Quota,
			Used:  data.QuotaUsage.Used,
		},
		Cpu: data.Cpu,
		Mem: data.Mem,
		Gpu: data.Gpu,
	}, nil
}

// GetProjectQuotasStatistics 获取项目资源额度统计信息
func (a *Action) GetProjectQuotasStatistics(ctx context.Context, req *types.GetProjectQuotasStatisticsReq) (
	*types.ProjectQuotasStatisticsData, error) {
	data, err := projectrmgr.GetProjectQuotasStatistics(ctx, &bcsproject.GetProjectQuotasStatisticsRequest{
		ProjectIDOrCode: req.ProjectIDOrCode,
		QuotaType:       req.QuotaType,
		IsContainShared: req.IsContainShared,
	})
	if err != nil {
		return nil, utils.SystemError(err)
	}

	return &types.ProjectQuotasStatisticsData{
		Cpu: func() *types.QuotaResourceData {
			if data.Cpu == nil {
				return &types.QuotaResourceData{}
			}
			return &types.QuotaResourceData{
				UsedNum:      data.Cpu.UsedNum,
				AvailableNum: data.Cpu.AvailableNum,
				TotalNum:     data.Cpu.TotalNum,
				UseRate:      data.Cpu.UseRate,
			}
		}(),
		Mem: func() *types.QuotaResourceData {
			if data.Mem == nil {
				return &types.QuotaResourceData{}
			}
			return &types.QuotaResourceData{
				UsedNum:      data.Mem.UsedNum,
				AvailableNum: data.Mem.AvailableNum,
				TotalNum:     data.Mem.TotalNum,
				UseRate:      data.Mem.UseRate,
			}
		}(),
		Gpu: func() *types.QuotaResourceData {
			if data.Gpu == nil {
				return &types.QuotaResourceData{}
			}
			return &types.QuotaResourceData{
				UsedNum:      data.Gpu.UsedNum,
				AvailableNum: data.Gpu.AvailableNum,
				TotalNum:     data.Gpu.TotalNum,
				UseRate:      data.Gpu.UseRate,
			}
		}(),
	}, nil
}

func conver2TypeQuota(quota *bcsproject.ProjectQuota) *types.ProjectQuota {
	return &types.ProjectQuota{
		QuotaId:      quota.QuotaId,
		QuotaName:    quota.QuotaName,
		ProjectID:    quota.ProjectID,
		ProjectCode:  quota.ProjectCode,
		ClusterId:    quota.ClusterId,
		ClusterName:  quota.ClusterName,
		NameSpace:    quota.NameSpace,
		BusinessID:   quota.BusinessID,
		BusinessName: quota.BusinessName,
		Description:  quota.Description,
		IsDeleted:    quota.IsDeleted,
		QuotaType:    quota.QuotaType,
		Quota:        conver2TypeQuotaResource(quota.Quota),
		Status:       quota.Status,
		Message:      quota.Message,
		CreateTime:   quota.CreateTime,
		UpdateTime:   quota.UpdateTime,
		Creator:      quota.Creator,
		Updater:      quota.Updater,
		Provider:     quota.Provider,
		NodeGroups: func() []*types.NodeGroupQuota {
			nodeGroups := make([]*types.NodeGroupQuota, 0)
			for _, item := range quota.NodeGroups {
				nodeGroups = append(nodeGroups, &types.NodeGroupQuota{
					ClusterId:   item.ClusterId,
					NodeGroupId: item.NodeGroupId,
					QuotaNum:    item.QuotaNum,
					QuotaUsed:   item.QuotaUsed,
				})
			}
			return nodeGroups
		}(),
		Labels:      quota.Labels,
		Annotations: quota.Annotations,
		QuotaAttr: func() *types.QuotaAttr {
			if quota.QuotaAttr == nil {
				return &types.QuotaAttr{}
			}
			return &types.QuotaAttr{
				SourceBkBizIDs:           quota.QuotaAttr.SourceBkBizIDs,
				SourceBkBizNames:         quota.QuotaAttr.SourceBkBizNames,
				ComputeType:              quota.QuotaAttr.ComputeType,
				PurchaseDurationType:     quota.QuotaAttr.PurchaseDurationType,
				PurchaseDurationSettings: quota.QuotaAttr.PurchaseDurationSettings,
				StartTime:                quota.QuotaAttr.StartTime,
				EndTime:                  quota.QuotaAttr.EndTime,
			}
		}(),
		QuotaSharedEnabled: quota.QuotaSharedEnabled,
		QuotaSharedProjectList: func() []*types.QuotaSharedProject {
			projects := make([]*types.QuotaSharedProject, 0)
			for _, item := range quota.QuotaSharedProjectList {
				projects = append(projects, &types.QuotaSharedProject{
					ProjectID:      item.ProjectID,
					ProjectCode:    item.ProjectCode,
					ProjectName:    item.ProjectName,
					ShareStrategy:  item.ShareStrategy,
					UsageLimit:     &types.QuotaLimit{QuotaNum: item.UsageLimit.QuotaNum},
					UsedAmount:     &types.QuotaLimit{QuotaNum: item.UsedAmount.QuotaNum},
					ShareStartTime: item.ShareStartTime,
					ShareEndTime:   item.ShareEndTime,
					Status:         item.Status,
				})
			}
			return projects
		}(),
	}
}

func conver2TypeQuotaResource(quota *bcsproject.QuotaResource) *types.QuotaResource {
	if quota == nil {
		return &types.QuotaResource{}
	}

	return &types.QuotaResource{
		ZoneResources: func() *types.InstanceTypeConfig {
			if quota.ZoneResources == nil {
				return &types.InstanceTypeConfig{}
			}
			return &types.InstanceTypeConfig{
				Region:       quota.ZoneResources.Region,
				InstanceType: quota.ZoneResources.InstanceType,
				Cpu:          quota.ZoneResources.Cpu,
				Mem:          quota.ZoneResources.Mem,
				Gpu:          quota.ZoneResources.Gpu,
				ZoneId:       quota.ZoneResources.ZoneId,
				ZoneName:     quota.ZoneResources.ZoneName,
				QuotaNum:     quota.ZoneResources.QuotaNum,
				QuotaUsed:    quota.ZoneResources.QuotaUsed,
				SystemDisk: func() *types.DataDisk {
					if quota.ZoneResources.SystemDisk == nil {
						return &types.DataDisk{}
					}
					return &types.DataDisk{
						DiskType: quota.ZoneResources.SystemDisk.DiskType,
						DiskSize: quota.ZoneResources.SystemDisk.DiskSize,
					}
				}(),
				DataDisks: func() []*types.DataDisk {
					disks := make([]*types.DataDisk, 0)
					for _, item := range quota.ZoneResources.DataDisks {
						disks = append(disks, &types.DataDisk{
							DiskType: item.DiskType,
							DiskSize: item.DiskSize,
						})
					}
					return disks
				}(),
			}
		}(),
		Cpu: func() *types.DeviceInfo {
			if quota.Cpu == nil {
				return &types.DeviceInfo{}
			}
			return &types.DeviceInfo{
				DeviceType:      quota.Cpu.DeviceType,
				DeviceQuota:     quota.Cpu.DeviceQuota,
				DeviceQuotaUsed: quota.Cpu.DeviceQuotaUsed,
				Attributes:      quota.Cpu.Attributes,
			}
		}(),
		Mem: func() *types.DeviceInfo {
			if quota.Mem == nil {
				return &types.DeviceInfo{}
			}
			return &types.DeviceInfo{
				DeviceType:      quota.Mem.DeviceType,
				DeviceQuota:     quota.Mem.DeviceQuota,
				DeviceQuotaUsed: quota.Mem.DeviceQuotaUsed,
				Attributes:      quota.Mem.Attributes,
			}
		}(),
		Gpu: func() *types.DeviceInfo {
			if quota.Gpu == nil {
				return &types.DeviceInfo{}
			}
			return &types.DeviceInfo{
				DeviceType:      quota.Gpu.DeviceType,
				DeviceQuota:     quota.Gpu.DeviceQuota,
				DeviceQuotaUsed: quota.Gpu.DeviceQuotaUsed,
				Attributes:      quota.Gpu.Attributes,
			}
		}(),
	}
}

func conver2ProjectQuotaAttr(attr *types.QuotaAttr) *bcsproject.QuotaAttr {
	if attr == nil {
		return &bcsproject.QuotaAttr{}
	}
	return &bcsproject.QuotaAttr{
		SourceBkBizIDs:           attr.SourceBkBizIDs,
		SourceBkBizNames:         attr.SourceBkBizNames,
		ComputeType:              attr.ComputeType,
		PurchaseDurationType:     attr.PurchaseDurationType,
		PurchaseDurationSettings: attr.PurchaseDurationSettings,
		StartTime:                attr.StartTime,
		EndTime:                  attr.EndTime,
	}
}

func conver2ProjectQuotaSharedProject(list []*types.QuotaSharedProject) []*bcsproject.QuotaSharedProject {
	projects := make([]*bcsproject.QuotaSharedProject, 0)
	for _, item := range list {
		projects = append(projects, &bcsproject.QuotaSharedProject{
			ProjectID:      item.ProjectID,
			ProjectCode:    item.ProjectCode,
			ProjectName:    item.ProjectName,
			ShareStrategy:  item.ShareStrategy,
			UsageLimit:     &bcsproject.QuotaLimit{QuotaNum: item.UsageLimit.QuotaNum},
			UsedAmount:     &bcsproject.QuotaLimit{QuotaNum: item.UsedAmount.QuotaNum},
			ShareStartTime: item.ShareStartTime,
			ShareEndTime:   item.ShareEndTime,
			Status:         item.Status,
		})
	}

	return projects
}

func conver2ProjectQuotaResource(quota *types.QuotaResource) *bcsproject.QuotaResource {
	if quota == nil {
		return &bcsproject.QuotaResource{}
	}

	return &bcsproject.QuotaResource{
		ZoneResources: func() *bcsproject.InstanceTypeConfig {
			if quota.ZoneResources == nil {
				return &bcsproject.InstanceTypeConfig{}
			}
			return &bcsproject.InstanceTypeConfig{
				Region:       quota.ZoneResources.Region,
				InstanceType: quota.ZoneResources.InstanceType,
				Cpu:          quota.ZoneResources.Cpu,
				Mem:          quota.ZoneResources.Mem,
				Gpu:          quota.ZoneResources.Gpu,
				ZoneId:       quota.ZoneResources.ZoneId,
				ZoneName:     quota.ZoneResources.ZoneName,
				QuotaNum:     quota.ZoneResources.QuotaNum,
				QuotaUsed:    quota.ZoneResources.QuotaUsed,
				SystemDisk: func() *bcsproject.DataDisk {
					if quota.ZoneResources.SystemDisk == nil {
						return &bcsproject.DataDisk{}
					}
					return &bcsproject.DataDisk{
						DiskType: quota.ZoneResources.SystemDisk.DiskType,
						DiskSize: quota.ZoneResources.SystemDisk.DiskSize,
					}
				}(),
				DataDisks: func() []*bcsproject.DataDisk {
					disks := make([]*bcsproject.DataDisk, 0)
					for _, item := range quota.ZoneResources.DataDisks {
						disks = append(disks, &bcsproject.DataDisk{
							DiskType: item.DiskType,
							DiskSize: item.DiskSize,
						})
					}
					return disks
				}(),
			}
		}(),
		Cpu: func() *bcsproject.DeviceInfo {
			if quota.Cpu == nil {
				return &bcsproject.DeviceInfo{}
			}
			return &bcsproject.DeviceInfo{
				DeviceType:      quota.Cpu.DeviceType,
				DeviceQuota:     quota.Cpu.DeviceQuota,
				DeviceQuotaUsed: quota.Cpu.DeviceQuotaUsed,
				Attributes:      quota.Cpu.Attributes,
			}
		}(),
		Mem: func() *bcsproject.DeviceInfo {
			if quota.Mem == nil {
				return &bcsproject.DeviceInfo{}
			}
			return &bcsproject.DeviceInfo{
				DeviceType:      quota.Mem.DeviceType,
				DeviceQuota:     quota.Mem.DeviceQuota,
				DeviceQuotaUsed: quota.Mem.DeviceQuotaUsed,
				Attributes:      quota.Mem.Attributes,
			}
		}(),
		Gpu: func() *bcsproject.DeviceInfo {
			if quota.Gpu == nil {
				return &bcsproject.DeviceInfo{}
			}
			return &bcsproject.DeviceInfo{
				DeviceType:      quota.Gpu.DeviceType,
				DeviceQuota:     quota.Gpu.DeviceQuota,
				DeviceQuotaUsed: quota.Gpu.DeviceQuotaUsed,
				Attributes:      quota.Gpu.Attributes,
			}
		}(),
	}
}
