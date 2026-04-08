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

// ListClusterReq list cluster request
type ListClusterReq struct {
	ClusterID   string `json:"clusterID" in:"query=clusterID"`
	ClusterName string `json:"clusterName" in:"query=clusterName"`
	ProjectID   string `json:"projectID" in:"query=projectID"`
	BusinessID  string `json:"businessID" in:"query=businessID"`
	Environment string `json:"environment" in:"query=environment"`
	Status      string `json:"status" in:"query=status"`
	Provider    string `json:"provider" in:"query=provider"`
	VpcID       string `json:"vpcID" in:"query=vpcID"`
	SystemID    string `json:"systemID" in:"query=systemID"`
	Creator     string `json:"creator" in:"query=creator"`
	ManageType  string `json:"manageType" in:"query=manageType"`
	All         bool   `json:"all" in:"query=all"`
	Page        uint32 `json:"page" in:"query=page"`
	Limit       uint32 `json:"limit" in:"query=limit"`
	Sort        string `json:"sort" in:"query=sort"`
	Order       string `json:"order" in:"query=order"`
}

// ListClusterResp list cluster response
type ListClusterResp struct {
	Total   uint32             `json:"total"`
	Results []*ListClusterData `json:"results"`
}

// ListClusterData list cluster data
type ListClusterData struct {
	ClusterID     string `json:"clusterID"`
	ClusterName   string `json:"clusterName"`
	Provider      string `json:"provider"`
	Status        string `json:"status"`
	Environment   string `json:"environment"`
	ManageType    string `json:"manageType"`
	ProjectID     string `json:"projectID"`
	ProjectName   string `json:"projectName"`
	ProjectCode   string `json:"projectCode"`
	BusinessID    string `json:"businessID"`
	BusinessName  string `json:"businessName"`
	Creator       string `json:"creator"`
	BizMaintainer string `json:"bizMaintainer"`
	Link          string `json:"link"`
}

// GetClusterReq get cluster request
type GetClusterReq struct {
	ClusterID string `json:"clusterID" in:"path=clusterID"`
	ProjectID string `json:"projectID" in:"query=projectID"`
}

// GetClusterResp get cluster response
type GetClusterResp struct {
	ClusterID            string               `json:"clusterID"`
	ClusterName          string               `json:"clusterName"`
	Provider             string               `json:"provider"`
	Region               string               `json:"region"`
	VpcID                string               `json:"vpcID"`
	ProjectID            string               `json:"projectID"`
	BusinessID           string               `json:"businessID"`
	Environment          string               `json:"environment"`
	EngineType           string               `json:"engineType"`
	ClusterType          string               `json:"clusterType"`
	Label                map[string]string    `json:"label"`
	Creator              string               `json:"creator"`
	CreateTime           string               `json:"createTime"`
	UpdateTime           string               `json:"updateTime"`
	SystemID             string               `json:"systemID"`
	ManageType           string               `json:"manageType"`
	ClusterBasicSettings *ClusterBasicSetting `json:"clusterBasicSettings"`
	Status               string               `json:"status"`
	Updater              string               `json:"updater"`
	NetworkType          string               `json:"networkType"`
	ModuleID             string               `json:"moduleID"`
	IsCommonCluster      bool                 `json:"isCommonCluster"`
	Description          string               `json:"description"`
	ClusterCategory      string               `json:"clusterCategory"`
	IsShared             bool                 `json:"isShared"`
	IsMixed              bool                 `json:"isMixed"`
	CloudAccountID       string               `json:"cloudAccountID"`
	Link                 string               `json:"link"`
}

// GetClusterOverviewReq get cluster overview request
type GetClusterOverviewReq struct {
	ClusterID   string `json:"clusterID" in:"path=clusterID"`
	ProjectCode string `json:"projectCode" in:"query=projectCode"`
}

// GetClusterOverviewResp 集群概览接口
type GetClusterOverviewResp struct {
	CPUUsage    *Usage     `json:"cpu_usage"`
	DiskUsage   *UsageByte `json:"disk_usage"`
	MemoryUsage *UsageByte `json:"memory_usage"`
	DiskIOUsage *Usage     `json:"diskio_usage"`
	PodUsage    *Usage     `json:"pod_usage"`
}

// Usage 使用量
type Usage struct {
	Used    string `json:"used"`
	Request string `json:"request"`
	Total   string `json:"total"`
}

// UsageByte 使用量, bytes单位
type UsageByte struct {
	UsedByte    string `json:"used_bytes"`
	RequestByte string `json:"request_bytes"`
	TotalByte   string `json:"total_bytes"`
}

// UpdateClusterBasicInfoReq update cluster basic info request
type UpdateClusterBasicInfoReq struct {
	ClusterID            string               `json:"clusterID" in:"path=clusterID"`
	ClusterName          string               `json:"clusterName"`
	Status               string               `json:"status"`
	ProjectID            string               `json:"projectID"`
	IsShared             bool                 `json:"isShared"`
	IsMixed              bool                 `json:"isMixed"`
	BusinessID           string               `json:"businessID"`
	Description          string               `json:"description"`
	Environment          string               `json:"environment"`
	Labels2              *MapStruct           `json:"labels2"`
	SharedRanges         *SharedClusterRanges `json:"sharedRanges"`
	ClusterBasicSettings *ClusterBasicSetting `json:"clusterBasicSettings"`
}

// UpdateClusterNetworkConfigReq update cluster network config request
type UpdateClusterNetworkConfigReq struct {
	ClusterID       string          `json:"clusterID" in:"path=clusterID"`
	NetworkSettings *NetworkSetting `json:"networkSettings"`
}

// UpdateClusterControlPlaneConfigReq update cluster basic info request
type UpdateClusterControlPlaneConfigReq struct {
	ClusterID string   `json:"clusterID" in:"path=clusterID"`
	Master    []string `json:"master"`
}

// MapStruct map structure
type MapStruct struct {
	Values map[string]string `json:"values"`
}

// SharedClusterRanges shared cluster ranges
type SharedClusterRanges struct {
	Bizs             []string `json:"bizs"`
	ProjectIdOrCodes []string `json:"projectIdOrCodes"`
}

// UpdateClusterOperatorReq update cluster operator request
type UpdateClusterOperatorReq struct {
	ClusterID string `json:"clusterID" in:"path=clusterID"`
	Creator   string `json:"creator"`
	Updater   string `json:"updater"`
}

// UpdateClusterProjectBusinessReq update cluster projectID or businessID request
type UpdateClusterProjectBusinessReq struct {
	ClusterID  string `json:"clusterID" in:"path=clusterID"`
	ProjectID  string `json:"projectID"`
	BusinessID string `json:"businessID"`
}

// ClusterBasicSettings cluster basic setting
type ClusterBasicSetting struct {
	OS                        string            `json:"OS"`
	Version                   string            `json:"version"`
	ClusterTags               map[string]string `json:"clusterTags"`
	VersionName               string            `json:"versionName"`
	SubnetID                  string            `json:"subnetID"`
	ClusterLevel              string            `json:"clusterLevel"`
	IsAutoUpgradeClusterLevel bool              `json:"isAutoUpgradeClusterLevel"`
	Area                      *CloudArea        `json:"area"`
	Module                    *ClusterModule    `json:"module"`
	UpgradePolicy             *UpgradePolicy    `json:"upgradePolicy"`
}

// ClusterAdvanceSetting cluster advance setting
type ClusterAdvanceSetting struct {
	IPVS               bool              `json:"IPVS"`
	ContainerRuntime   string            `json:"containerRuntime"`
	RuntimeVersion     string            `json:"runtimeVersion"`
	ExtraArgs          map[string]string `json:"extraArgs"`
	NetworkType        string            `json:"networkType"`
	DeletionProtection bool              `json:"deletionProtection"`
	AuditEnabled       bool              `json:"auditEnabled"`
	EnableHa           bool              `json:"enableHa"`
}

// ClusterModule cluster module
type ClusterModule struct {
	MasterModuleID   string `json:"masterModuleID"`
	MasterModuleName string `json:"masterModuleName"`
	WorkerModuleID   string `json:"workerModuleID"`
	WorkerModuleName string `json:"workerModuleName"`
}

// UpgradePolicy upgrade policy
type UpgradePolicy struct {
	SupportType string `json:"supportType"`
}

// GetClusterBasicInfoReq get cluster basic info request
type GetClusterBasicInfoReq struct {
	ClusterID string `json:"clusterID" in:"path=clusterID"`
	ProjectID string `json:"projectID" in:"query=projectID"`
}

// GetClusterBasicInfoResp get cluster basic info response
type GetClusterBasicInfoResp struct {
	ClusterID              string                 `json:"clusterID"`
	ClusterName            string                 `json:"clusterName"`
	Status                 string                 `json:"status"`
	Description            string                 `json:"description"`
	Provider               string                 `json:"provider"`
	ProjectID              string                 `json:"projectID"`
	ProjectName            string                 `json:"projectName"`
	BusinessID             string                 `json:"businessID"`
	BusinessName           string                 `json:"businessName"`
	ManageType             string                 `json:"manageType"`
	IsShared               bool                   `json:"isShared"`
	IsMixed                bool                   `json:"isMixed"`
	Labels                 map[string]string      `json:"labels"`
	ClusterBasicSettings   *ClusterBasicSetting   `json:"clusterBasicSettings"`
	ClusterAdvanceSettings *ClusterAdvanceSetting `json:"clusterAdvanceSettings"`
	ContainerRuntime       string                 `json:"containerRuntime"`
	RuntimeVersion         string                 `json:"runtimeVersion"`
	Region                 string                 `json:"region"`
	Environment            string                 `json:"environment"`
	SharedRanges           *SharedClusterRanges   `json:"sharedRanges"`
	ClusterCategory        string                 `json:"clusterCategory"`
	Creator                string                 `json:"creator"`
	CreateTime             string                 `json:"createTime"`
	UpdateTime             string                 `json:"updateTime"`
}

// GetClusterNetworkConfigReq get cluster network config request
type GetClusterNetworkConfigReq struct {
	ClusterID string `json:"clusterID" in:"path=clusterID"`
}

// GetClusterNetworkConfigResp get cluster network config response
type GetClusterNetworkConfigResp struct {
	NetworkSettings        *NetworkSetting        `json:"networkSettings"`
	ClusterAdvanceSettings *ClusterAdvanceSetting `json:"clusterAdvanceSettings"`
	Region                 string                 `json:"region"`
	VpcID                  string                 `json:"vpcID"`
	NetworkType            string                 `json:"networkType"`
	Subnets                []*Subnet              `json:"subnets"`
}

// GetClusterControlPlaneConfigReq get cluster control plane config request
type GetClusterControlPlaneConfigReq struct {
	ClusterID string `json:"clusterID" in:"path=clusterID"`
}

// GetClusterControlPlaneConfigResp get cluster network config response
type GetClusterControlPlaneConfigResp struct {
	ManageType                string           `json:"manageType"`
	ClusterLevel              string           `json:"clusterLevel"`
	SecurityGroup             string           `json:"securityGroup"`
	IsAutoUpgradeClusterLevel bool             `json:"isAutoUpgradeClusterLevel"`
	Module                    *ClusterModule   `json:"module"`
	Master                    map[string]*Node `json:"master"`
}

// AddClusterCidrReq add subnet to cluster request
type AddClusterCidrReq struct {
	ClusterID string   `json:"clusterID" in:"path=clusterID"`
	Cidrs     []string `json:"cidrs"`
	Operator  string   `json:"operator"`
}

// AddSubnetToClusterReq add subnet to cluster request
type AddSubnetToClusterReq struct {
	ClusterID  string       `json:"clusterID" in:"path=clusterID"`
	NewSubnets []*NewSubnet `json:"newSubnets"`
	Operator   string       `json:"operator"`
}

// NewSubnet new subnet
type NewSubnet struct {
	Mask  uint32 `json:"mask"`
	Zone  string `json:"zone"`
	IpCnt uint32 `json:"ipCnt"`
}

// Node node
type Node struct {
	NodeID         string `json:"nodeID"`
	InnerIP        string `json:"innerIP"`
	InstanceType   string `json:"instanceType"`
	CPU            uint32 `json:"CPU"`
	Mem            uint32 `json:"mem"`
	GPU            uint32 `json:"GPU"`
	Status         string `json:"status"`
	ZoneID         string `json:"zoneID"`
	NodeGroupID    string `json:"nodeGroupID"`
	ClusterID      string `json:"clusterID"`
	VPC            string `json:"VPC"`
	Region         string `json:"region"`
	Passwd         string `json:"passwd"`
	Zone           uint32 `json:"zone"`
	DeviceID       string `json:"deviceID"`
	NodeTemplateID string `json:"nodeTemplateID"`
	NodeType       string `json:"nodeType"`
	NodeName       string `json:"nodeName"`
	InnerIPv6      string `json:"innerIPv6"`
	ZoneName       string `json:"zoneName"`
	TaskID         string `json:"taskID"`
	FailedReason   string `json:"failedReason"`
	ChargeType     string `json:"chargeType"`
	DataDiskNum    uint32 `json:"dataDiskNum"`
	IsGpuNode      bool   `json:"isGpuNode"`
}

// NetworkSetting network setting
type NetworkSetting struct {
	ClusterIPv4CIDR     string        `json:"clusterIPv4CIDR"`
	ServiceIPv4CIDR     string        `json:"serviceIPv4CIDR"`
	MaxNodePodNum       uint32        `json:"maxNodePodNum"`
	MaxServiceNum       uint32        `json:"maxServiceNum"`
	EnableVPCCni        bool          `json:"enableVPCCni"`
	EniSubnetIDs        []string      `json:"eniSubnetIDs"`
	SubnetSource        *SubnetSource `json:"subnetSource"`
	IsStaticIpMode      bool          `json:"isStaticIpMode"`
	ClaimExpiredSeconds uint32        `json:"claimExpiredSeconds"`
	MultiClusterCIDR    []string      `json:"multiClusterCIDR"`
	CidrStep            uint32        `json:"cidrStep"`
	ClusterIpType       string        `json:"clusterIpType"`
	ClusterIPv6CIDR     string        `json:"clusterIPv6CIDR"`
	ServiceIPv6CIDR     string        `json:"serviceIPv6CIDR"`
	Status              string        `json:"status"`
	NetworkMode         string        `json:"networkMode"`
}

// Subnet subnet
type SubnetSource struct {
	New     []*NewSubnet      `json:"new"`
	Existed *ExistedSubnetIDs `json:"existed"`
}

// ExistedSubnetIDs existed subnet ids
type ExistedSubnetIDs struct {
	Ids []string `json:"ids"`
}
