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

// Package types pod types
package types

// CreateProjectQuotaReq xxx
type CreateProjectQuotaReq struct {
	QuotaName              string                `json:"quotaName"`
	ProjectID              string                `json:"projectID"`
	ProjectCode            string                `json:"projectCode"`
	ClusterId              string                `json:"clusterId"`
	ClusterName            string                `json:"clusterName"`
	NameSpace              string                `json:"nameSpace"`
	BusinessID             string                `json:"businessID"`
	BusinessName           string                `json:"businessName"`
	Description            string                `json:"description"`
	QuotaType              string                `json:"quotaType"`
	Provider               string                `json:"provider"`
	Quota                  *QuotaResource        `json:"quota"`
	Labels                 map[string]string     `json:"labels"`
	Annotations            map[string]string     `json:"annotations"`
	QuotaAttr              *QuotaAttr            `json:"quotaAttr"`
	QuotaSharedEnabled     bool                  `json:"quotaSharedEnabled"`
	QuotaSharedProjectList []*QuotaSharedProject `json:"quotaSharedProjectList"`
	SkipItsmApproval       bool                  `json:"skipItsmApproval"`
}

// QuotaResource xxx
type QuotaResource struct {
	ZoneResources *InstanceTypeConfig `json:"zoneResources"`
	Cpu           *DeviceInfo         `json:"cpu"`
	Mem           *DeviceInfo         `json:"mem"`
	Gpu           *DeviceInfo         `json:"gpu"`
}

// InstanceTypeConfig 整机资源配置
type InstanceTypeConfig struct {
	Region       string      `json:"region,"`
	InstanceType string      `json:"instanceType,"`
	Cpu          uint32      `json:"cpu,"`
	Mem          uint32      `json:"mem,"`
	Gpu          uint32      `json:"gpu,"`
	ZoneId       string      `json:"zoneId,"`
	ZoneName     string      `json:"zoneName,"`
	QuotaNum     uint32      `json:"quotaNum,"`
	QuotaUsed    uint32      `json:"quotaUsed,"`
	SystemDisk   *DataDisk   `json:"systemDisk,"`
	DataDisks    []*DataDisk `json:"dataDisks,"`
}

// QuotaAttr xxx
type QuotaAttr struct {
	SourceBkBizIDs           string `json:"sourceBkBizIDs"`
	SourceBkBizNames         string `json:"SourceBkBizNames"`
	ComputeType              string `json:"computeType"`
	PurchaseDurationType     string `json:"purchaseDurationType"`
	PurchaseDurationSettings string `json:"purchaseDurationSettings"`
	StartTime                string `json:"startTime"`
	EndTime                  string `json:"endTime"`
}

// QuotaSharedProject xxx
type QuotaSharedProject struct {
	ProjectID      string      `json:"projectID"`
	ProjectCode    string      `json:"projectCode"`
	ProjectName    string      `json:"projectName"`
	ShareStrategy  string      `json:"shareStrategy"`
	UsageLimit     *QuotaLimit `json:"usageLimit"`
	UsedAmount     *QuotaLimit `json:"usedAmount"`
	ShareStartTime string      `json:"shareStartTime"`
	ShareEndTime   string      `json:"shareEndTime"`
	Status         string      `json:"status"`
}

// QuotaLimit xxx
type QuotaLimit struct {
	QuotaNum int64 `json:"quotaNum"`
}

// DeviceInfo device data (cpu/mem/gpu)
type DeviceInfo struct {
	DeviceType      string            `json:"deviceType"`
	DeviceQuota     string            `json:"deviceQuota"`
	DeviceQuotaUsed string            `json:"deviceQuotaUsed"`
	Attributes      map[string]string `json:"attributes"`
}

// ProjectQuota xxx
type ProjectQuota struct {
	QuotaId                string                `json:"quotaId" in:"path=quotaId"`
	QuotaName              string                `json:"quotaName"`
	ProjectID              string                `json:"projectID"`
	ProjectCode            string                `json:"projectCode"`
	ClusterId              string                `json:"clusterId"`
	ClusterName            string                `json:"clusterName"`
	NameSpace              string                `json:"nameSpace"`
	BusinessID             string                `json:"businessID"`
	BusinessName           string                `json:"businessName"`
	Description            string                `json:"description"`
	IsDeleted              bool                  `json:"isDeleted"`
	QuotaType              string                `json:"quotaType"`
	Quota                  *QuotaResource        `json:"quota"`
	Status                 string                `json:"status"`
	Message                string                `json:"message"`
	CreateTime             string                `json:"createTime"`
	UpdateTime             string                `json:"updateTime"`
	Creator                string                `json:"creator"`
	Updater                string                `json:"updater"`
	Provider               string                `json:"provider"`
	NodeGroups             []*NodeGroup          `json:"nodeGroups"`
	Labels                 map[string]string     `json:"labels"`
	Annotations            map[string]string     `json:"annotations"`
	QuotaAttr              *QuotaAttr            `json:"quotaAttr"`
	QuotaSharedEnabled     bool                  `json:"quotaSharedEnabled"`
	QuotaSharedProjectList []*QuotaSharedProject `json:"quotaSharedProjectList"`
}

// UpdateProjectQuotaReq xxx
type UpdateProjectQuotaReq struct {
	QuotaId                string                `json:"quotaId" in:"path=quotaId"`
	Name                   string                `json:"name"`
	Quota                  *QuotaResource        `json:"quota"`
	Updater                string                `json:"updater"`
	Labels                 map[string]string     `json:"labels"`
	Annotations            map[string]string     `json:"annotations"`
	QuotaAttr              *QuotaAttr            `json:"quotaAttr"`
	QuotaSharedEnabled     bool                  `json:"quotaSharedEnabled"`
	QuotaSharedProjectList []*QuotaSharedProject `json:"quotaSharedProjectList"`
}

// ScaleUpProjectQuotaReq xxx
type ScaleUpProjectQuotaReq struct {
	QuotaId          string         `json:"quotaId" in:"path=quotaId"`
	Quota            *QuotaResource `json:"quota"`
	Updater          string         `json:"updater"`
	SkipItsmApproval bool           `json:"skipItsmApproval"`
}

// ScaleDownProjectQuotaReq xxx
type ScaleDownProjectQuotaReq struct {
	QuotaId          string         `json:"quotaId" in:"path=quotaId"`
	Quota            *QuotaResource `json:"quota"`
	Updater          string         `json:"updater"`
	SkipItsmApproval bool           `json:"skipItsmApproval"`
}

// DeleteProjectQuotaReq xxx
type DeleteProjectQuotaReq struct {
	QuotaId          string `json:"quotaId" in:"path=quotaId"`
	OnlyDeleteInfo   bool   `json:"onlyDeleteInfo" in:"query=onlyDeleteInfo"`
	SkipItsmApproval bool   `json:"skipItsmApproval" in:"query=skipItsmApproval"`
}

// ListProjectQuotasV2Req xxx
type ListProjectQuotasV2Req struct {
	QuotaId         string `json:"quotaId" in:"query=quotaId"`
	QuotaName       string `json:"quotaName" in:"query=quotaName"`
	ProjectIDOrCode string `json:"projectIDOrCode" in:"query=projectIDOrCode"`
	BusinessID      string `json:"businessID" in:"query=businessID"`
	QuotaType       string `json:"quotaType" in:"query=quotaType"`
	Provider        string `json:"provider" in:"query=provider"`
	Page            uint32 `json:"page" in:"query=page"`
	Limit           uint32 `json:"limit" in:"query=limit"`
}

// ListProjectQuotasData xxx
type ListProjectQuotasData struct {
	Total   uint32          `json:"total"`
	Results []*ProjectQuota `json:"results"`
}

// GetProjectQuotasUsageData xxx
type GetProjectQuotasUsageData struct {
	Quota        *ProjectQuota      `json:"quota"`
	Region       string             `json:"region"`
	InstanceType string             `json:"instanceType"`
	QuotaUsage   *ZoneResourceUsage `json:"quotaUsage"`
	Cpu          uint32             `json:"cpu"`
	Mem          uint32             `json:"mem"`
	Gpu          uint32             `json:"gpu"`
}

// ZoneResourceUsage xxx
type ZoneResourceUsage struct {
	Zone  string `json:"zone"`
	Quota uint32 `json:"quota"`
	Used  uint32 `json:"used"`
}

// GetProjectQuotasStatisticsReq xxx
type GetProjectQuotasStatisticsReq struct {
	ProjectIDOrCode string `json:"projectIDOrCode" in:"query=projectIDOrCode"`
	QuotaType       string `json:"quotaType" in:"query=quotaType"`
	IsContainShared bool   `json:"isContainShared" in:"query=isContainShared"`
}

// ProjectQuotasStatisticsData xxx
type ProjectQuotasStatisticsData struct {
	Cpu *QuotaResourceData `json:"cpu"`
	Mem *QuotaResourceData `json:"mem"`
	Gpu *QuotaResourceData `json:"gpu"`
}

// QuotaResourceData xxx
type QuotaResourceData struct {
	UsedNum      uint32  `json:"usedNum"`
	AvailableNum uint32  `json:"availableNum"`
	TotalNum     uint32  `json:"totalNum"`
	UseRate      float32 `json:"useRate"`
}

// GetProjectQuotaReq xxx
type GetProjectQuotaReq struct {
	QuotaId string `json:"quotaId" in:"path=quotaId"`
}
