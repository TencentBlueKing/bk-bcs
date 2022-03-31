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

package u1x21x202203082112

import (
	"time"
)

type project struct {
	ProjectID   string
	Name        string
	EnglishName string
	Creator     string
	ProjectType int
	UseBKRes    bool
	Description string
	IsOffline   bool
	Kind        string
	BusinessID  string
	DeployType  int
	BgID        string
	BgName      string
	DeptID      string
	DeptName    string
	CenterID    string
	CenterName  string
	IsSecret    bool
	Updater     string
	Credentials interface{}
}

type cluster struct {
	ClusterID            string
	ClusterName          string
	Provider             string
	Region               string
	VpcID                string
	ProjectID            string
	BusinessID           string
	Environment          string
	EngineType           string
	IsExclusive          bool
	ClusterType          string
	FederationClusterID  string
	Creator              string
	OnlyCreateInfo       bool
	CloudID              string
	ManageType           string
	SystemReinstall      bool
	InitLoginPassword    string
	NetworkType          string
	Master               []string
	Node                 []string
	AreaId               int
	NetworkSettings      createClustersNetworkSettings
	ClusterBasicSettings createClustersClusterBasicSettings
}

type node struct {
	ProjectID         string
	ClusterID         string
	InnerIP           string //nodeIP
	InitLoginPassword string
	NodeGroupID       string
	OnlyCreateInfo    bool
	Creator           string
}

//ccVersionConfigData :bcs cc request clusterConfig api data
type ccVersionConfigData struct {
	ClusterId string    `json:"cluster_id"`
	Configure string    `json:"configure"`
	CreatedAt time.Time `json:"created_at"`
	Creator   string    `json:"creator"`
	ID        int       `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ccversionConfigure :ccVersionConfigData.Creator
type ccversionConfigure struct {
	AreaID string `json:"area_id"`
	VpcID  string `json:"vpc_id"`
}

type reqDeleteNode struct {
	ClusterID string   `json:"clusterID"`
	Nodes     []string `json:"nodes"`
	// DeleteMode :删除模式，RETAIN(移除集群，但是保留主机)，TERMINATE(只支持按量计费的机器)，默认是RETAIN
	DeleteMode string `json:"deleteMode"`
	//IsForce :不管节点处于任何状态都强制删除，例如可能刚初始化，NotReady等
	IsForce  bool   `json:"isForce"`
	Operator string `json:"operator"` // 操作者
	//OnlyDeleteInfo :默认为false。设置为true时，仅删除cluster-manager所记录的信息，不会触发任何自动化流程.
	OnlyDeleteInfo bool `json:"onlyDeleteInfo"`
}

type bcsNodeListData struct {
	NodeID       string `json:"nodeID"`
	InnerIP      string `json:"innerIP"`
	InstanceType string `json:"instanceType"`
	CPU          int    `json:"CPU"`
	Mem          int    `json:"mem"`
	GPU          int    `json:"GPU"`
	Status       string `json:"status"`
	ZoneID       string `json:"zoneID"`
	NodeGroupID  string `json:"nodeGroupID"`
	ClusterID    string `json:"clusterID"`
	VPC          string `json:"VPC"`
	Region       string `json:"region"`
	Passwd       string `json:"passwd"`
	Zone         int    `json:"zone"`
}

type bcsClusterBase struct {
	ClusterID           string `json:"clusterID"`   // required
	ClusterName         string `json:"clusterName"` // required
	Provider            string `json:"provider"`    // required
	Region              string `json:"region"`      // required
	VpcID               string `json:"vpcID"`
	ProjectID           string `json:"projectID"`   // required
	BusinessID          string `json:"businessID"`  // required
	Environment         string `json:"environment"` // required
	EngineType          string `json:"engineType"`  // required
	IsExclusive         bool   `json:"isExclusive"` // required
	ClusterType         string `json:"clusterType"` // required
	FederationClusterID string `json:"federationClusterID"`
	Creator             string `json:"creator"` // required
	OnlyCreateInfo      bool   `json:"onlyCreateInfo"`
	CloudID             string `json:"cloudID"`
	ManageType          string `json:"manageType"`
	SystemReinstall     bool   `json:"systemReinstall"`
	InitLoginPassword   string `json:"initLoginPassword"`
	NetworkType         string `json:"networkType"`
}

type bcsReqUpdateCluster struct {
	bcsClusterBase
	NetworkSettings        createClustersNetworkSettings      `json:"networkSettings"` // TODO 待定
	ClusterBasicSettings   createClustersClusterBasicSettings `json:"clusterBasicSettings"`
	Updater                string                             `json:"updater"`
	Master                 []string                           `json:"master"`
	Node                   []string                           `json:"node"`
	Labels                 interface{}                        `json:"labels,omitempty"`
	BcsAddons              interface{}                        `json:"bcsAddons,omitempty"`
	ExtraAddons            interface{}                        `json:"extraAddons,omitempty"`
	ClusterAdvanceSettings interface{}                        `json:"clusterAdvanceSettings,omitempty"`
	NodeSettings           interface{}                        `json:"nodeSettings,omitempty"`
	// 创建集群是否使用已存在节点, 默认false, 即使用已经存在的节点, 从创建集群参数中获取
	AutoGenerateMasterNodes bool `json:"autoGenerateMasterNodes"`
	// 创建集群时 autoGenerateMasterNodes 为true, 系统自动生成master节点, 需要指定instances生成的配置信息,支持不同可用区实例"
	Instances interface{} `json:"instances,omitempty"`
	ExtraInfo interface{} `json:"ExtraInfo"`
	// 集群master节点的Instance id
	MasterInstanceID []string `json:"masterInstanceID"`
	//"集群状态，可能状态CREATING，RUNNING，DELETING，FALURE，INITIALIZATION，DELETED"
	Status string `json:"status"`
	// kubernetes集群在各云平台上资源ID
	SystemID string `json:"systemID"`
}

type bcsRespFindCluster struct {
	bcsClusterBase
	NetworkSettings        createClustersNetworkSettings      `json:"networkSettings"` // TODO 待定
	ClusterBasicSettings   createClustersClusterBasicSettings `json:"clusterBasicSettings"`
	Creator                string                             `json:"creator"` // required
	Updater                string                             `json:"updater"`
	Labels                 interface{}                        `json:"labels,omitempty"`
	BcsAddons              interface{}                        `json:"bcsAddons,omitempty"`
	ExtraAddons            interface{}                        `json:"extraAddons,omitempty"`
	ClusterAdvanceSettings interface{}                        `json:"clusterAdvanceSettings,omitempty"`
	NodeSettings           interface{}                        `json:"nodeSettings,omitempty"`
	// 创建集群是否使用已存在节点, 默认false, 即使用已经存在的节点, 从创建集群参数中获取
	AutoGenerateMasterNodes bool `json:"autoGenerateMasterNodes"`
	// 创建集群时 autoGenerateMasterNodes 为true, 系统自动生成master节点, 需要指定instances生成的配置信息,支持不同可用区实例"
	Instances interface{} `json:"instances,omitempty"`
	ExtraInfo interface{} `json:"ExtraInfo"`
	// 集群master节点的Instance id
	MasterInstanceID []string `json:"masterInstanceID"`
	//"集群状态，可能状态CREATING，RUNNING，DELETING，FALURE，INITIALIZATION，DELETED"
	Status string `json:"status"`
	// kubernetes集群在各云平台上资源ID
	SystemID string                              `json:"systemID"`
	Master   map[string]bcsRespFindClusterMaster `json:"master"`
}

type bcsRespFindClusterMaster struct {
	NodeID       string `json:"nodeID"`
	InnerIP      string `json:"innerIP"`
	InstanceType string `json:"instanceType"`
	CPU          int    `json:"CPU"`
	Mem          int    `json:"mem"`
	GPU          int    `json:"GPU"`
	Status       string `json:"status"`
	ZoneID       string `json:"zoneID"`
	NodeGroupID  string `json:"nodeGroupID"`
	ClusterID    string `json:"clusterID"`
	VPC          string `json:"VPC"`
	Region       string `json:"region"`
	Passwd       string `json:"passwd"`
	Zone         int    `json:"zone"`
}

type createClustersNetworkSettings struct {
	ClusterIPv4CIDR string `json:"clusterIPv4CIDR"`
	ServiceIPv4CIDR string `json:"serviceIPv4CIDR"`
	MaxNodePodNum   string `json:"maxNodePodNum"`
	MaxServiceNum   string `json:"maxServiceNum"`
}

type createClustersClusterBasicSettings struct {
	OS          string            `json:"OS"`
	Version     string            `json:"version"`
	ClusterTags map[string]string `json:"clusterTags"`
}

type ccProject struct {
	ApprovalStatus int       `json:"approval_status"`
	ApprovalTime   time.Time `json:"approval_time"`
	Approver       string    `json:"approver"`
	BgID           int       `json:"bg_id"`
	BgName         string    `json:"bg_name"`
	CcAppId        int       `json:"cc_app_id"`
	CenterID       int       `json:"center_id"`
	CenterName     string    `json:"center_name"`
	CreatedAt      time.Time `json:"created_at"`
	Creator        string    `json:"creator"`
	DataId         int       `json:"data_id"`
	DeployType     string    `json:"deploy_type"`
	DeptID         int       `json:"dept_id"`
	DeptName       string    `json:"dept_name"`
	Description    string    `json:"description"`
	EnglishName    string    `json:"english_name"`
	ID             int       `json:"id"`
	IsOfflined     bool      `json:"is_offlined"`
	IsSecrecy      bool      `json:"is_secrecy"`
	Kind           int       `json:"kind"`
	LogoAddr       string    `json:"logo_addr"`
	Name           string    `json:"name"`
	ProjectID      string    `json:"project_id"`
	ProjectName    string    `json:"project_name"`
	ProjectType    int       `json:"project_type"`
	Remark         string    `json:"remark"`
	UpdatedAt      time.Time `json:"updated_at"`
	Updator        string    `json:"updator"`
	UseBk          bool      `json:"use_bk"`
}

/***************************************************/
// cc返回的基础数据
type ccResp struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
	Result    bool        `json:"result"`
	Data      interface{} `json:"data"`
}

// cc 获取所有项目
type ccGetAllProject struct {
	Count   int         `json:"count"`
	Results []ccProject `json:"results"`
}

// cc获取所有cluster
type ccGetAllClusterData struct {
	Code        string             `json:"code"`
	ProjectID   string             `json:"id"` //ProjectID :api返回为id，=> ProjectID
	Name        string             `json:"name"`
	ClusterList []ccAllClusterList `json:"cluster_list"`
}

type ccAllClusterList struct {
	ClusterID     string `json:"id"` //ClusterID api返回为id，=> ClusterID
	IsPublic      bool   `json:"is_public"`
	ClusterName   string `json:"name"` //ClusterName api返回为name，=> ClusterName
	NamespaceList []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
}

// cc获取cluster信息
type ccGetClustersInfoData struct {
	ID                int         `json:"id"`
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	DeletedAt         interface{} `json:"deleted_at"`
	Extra             string      `json:"extra"`
	Name              string      `json:"name"`
	Creator           string      `json:"creator"`
	Description       string      `json:"description"`
	ProjectId         string      `json:"project_id"`
	RelatedProjects   string      `json:"related_projects"`
	ClusterID         string      `json:"cluster_id"`
	ClusterNum        int         `json:"cluster_num"`
	Status            string      `json:"status"`
	Disabled          bool        `json:"disabled"`
	Type              string      `json:"type"`
	Environment       string      `json:"environment"`
	AreaId            int         `json:"area_id"`
	ConfigSvrCount    int         `json:"config_svr_count"`
	MasterCount       int         `json:"master_count"`
	NodeCount         int         `json:"node_count"`
	IpResourceTotal   int         `json:"ip_resource_total"`
	IpResourceUsed    int         `json:"ip_resource_used"`
	Artifactory       string      `json:"artifactory"`
	TotalMem          int         `json:"total_mem"`
	RemainMem         int         `json:"remain_mem"`
	TotalCpu          int         `json:"total_cpu"`
	RemainCpu         int         `json:"remain_cpu"`
	TotalDisk         int         `json:"total_disk"`
	RemainDisk        int         `json:"remain_disk"`
	CapacityUpdatedAt time.Time   `json:"capacity_updated_at"`
	NotNeedNat        bool        `json:"not_need_nat"`
	ExtraClusterId    string      `json:"extra_cluster_id"`
	State             string      `json:"state"`
}

// cc获取所有节点
type ccGetAllNode struct {
	ID          int         `json:"id"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	DeletedAt   interface{} `json:"deleted_at"`
	Extra       string      `json:"extra"`
	Name        string      `json:"name"`
	Creator     string      `json:"creator"`
	Description string      `json:"description"`
	ProjectId   string      `json:"project_id"`
	ClusterId   string      `json:"cluster_id"`
	Status      string      `json:"status"`
	Kind        string      `json:"kind"`
	InnerIp     string      `json:"inner_ip"`
	OutterIp    string      `json:"outter_ip"`
	DeviceClass string      `json:"device_class"`
	Cpu         int         `json:"cpu"`
	Mem         int         `json:"mem"`
	Disk        int         `json:"disk"`
	IpResources int         `json:"ip_resources"`
	InstanceId  string      `json:"instance_id"`
}

// cc获取所有master
type ccGetAllMaster struct {
	ClusterId string `json:"cluster_id"`
	InnerIp   string `json:"inner_ip"`
	Status    string `json:"status"`
}

// 需要迁移的数据
type cmCreateProject struct {
	ProjectID   string                 `json:"projectID,omitempty"` //项目ID，长度为32位字符串
	Name        string                 `json:"name,omitempty"`
	EnglishName string                 `json:"englishName,omitempty"`
	Creator     string                 `json:"creator,omitempty"`
	ProjectType int                    `json:"projectType,omitempty"`
	UseBKRes    bool                   `json:"useBKRes,omitempty"`
	Description string                 `json:"description,omitempty"`
	IsOffline   bool                   `json:"isOffline,omitempty"`
	Kind        string                 `json:"kind,omitempty"`
	BusinessID  string                 `json:"businessID,omitempty"`
	DeployType  int                    `json:"deployType,omitempty"`
	BgID        string                 `json:"bgID,omitempty"`
	BgName      string                 `json:"bgName,omitempty"`
	DeptID      string                 `json:"deptID,omitempty"`
	DeptName    string                 `json:"deptName,omitempty"`
	CenterID    string                 `json:"centerID,omitempty"`
	CenterName  string                 `json:"centerName,omitempty"`
	IsSecret    bool                   `json:"isSecret,omitempty"`
	Credentials map[string]credentials `json:"credentials,omitempty"` // 用于记录账户信息
}

// 用于记录账号信息
type credentials struct {
	Key    string `json:"key"`
	Secret string `json:"secret"`
}

// 更新cm project
type cmUpdateProject struct {
	ProjectID   string                 `json:"projectID"`
	Name        string                 `json:"name"`
	Updater     string                 `json:"updater"`
	ProjectType int                    `json:"projectType"`
	UseBKRes    bool                   `json:"useBKRes"`
	Description string                 `json:"description"`
	IsOffline   bool                   `json:"isOffline"`
	Kind        string                 `json:"kind"`
	DeployType  int                    `json:"deployType"`
	BgID        string                 `json:"bgID"`
	BgName      string                 `json:"bgName"`
	DeptID      string                 `json:"deptID"`
	DeptName    string                 `json:"deptName"`
	CenterID    string                 `json:"centerID"`
	CenterName  string                 `json:"centerName"`
	IsSecret    bool                   `json:"isSecret"`
	BusinessID  string                 `json:"businessID"`
	Credentials map[string]credentials `json:"credentials,omitempty"` // 用于记录账户信息
}

// 查询cm project
type cmGetProject struct {
	ProjectID   string                 `json:"projectID"`
	Name        string                 `json:"name"`
	EnglishName string                 `json:"englishName"`
	Creator     string                 `json:"creator"`
	Updater     string                 `json:"updater"`
	ProjectType int                    `json:"projectType"`
	UseBKRes    bool                   `json:"useBKRes"`
	Description string                 `json:"description"`
	IsOffline   bool                   `json:"isOffline"`
	Kind        string                 `json:"kind"`
	BusinessID  string                 `json:"businessID"`
	DeployType  int                    `json:"deployType"`
	BgID        string                 `json:"bgID"`
	BgName      string                 `json:"bgName"`
	DeptID      string                 `json:"deptID"`
	DeptName    string                 `json:"deptName"`
	CenterID    string                 `json:"centerID"`
	CenterName  string                 `json:"centerName"`
	IsSecret    bool                   `json:"isSecret"`
	CreatTime   string                 `json:"creatTime"`
	UpdateTime  string                 `json:"updateTime"`
	Credentials map[string]credentials `json:"credentials"` // 用于记录账户信息
}

// cm 创建cluster
type cmCreateCluster struct {
	ClusterID               string                 `json:"clusterID,omitempty"`
	ClusterName             string                 `json:"clusterName,omitempty"`
	Provider                string                 `json:"provider,omitempty"`
	Region                  string                 `json:"region,omitempty"`
	VpcID                   string                 `json:"vpcID,omitempty"`
	ProjectID               string                 `json:"projectID,omitempty"`
	BusinessID              string                 `json:"businessID,omitempty"`
	Environment             string                 `json:"environment,omitempty"`
	EngineType              string                 `json:"engineType,omitempty"`
	IsExclusive             bool                   `json:"isExclusive,omitempty"`
	ClusterType             string                 `json:"clusterType,omitempty"`
	FederationClusterID     string                 `json:"federationClusterID,omitempty"`
	Labels                  map[string]string      `json:"labels,omitempty"`
	Creator                 string                 `json:"creator,omitempty"`
	OnlyCreateInfo          bool                   `json:"onlyCreateInfo,omitempty"`
	BcsAddons               map[string]bcsAddons   `json:"bcsAddons,omitempty"`
	ExtraAddons             map[string]extraAddons `json:"extraAddons,omitempty"`
	CloudID                 string                 `json:"cloudID,omitempty"`
	ManageType              string                 `json:"manageType,omitempty"`
	Master                  []string               `json:"master,omitempty,omitempty"`
	Nodes                   []string               `json:"nodes,omitempty"`
	NetworkSettings         netWorkSettings        `json:"networkSettings,omitempty"`
	ClusterBasicSettings    clusterBasicSettings   `json:"clusterBasicSettings,omitempty"`
	ClusterAdvanceSettings  clusterAdvanceSettings `json:"clusterAdvanceSettings,omitempty"`
	NodeSettings            nodeSettings           `json:"nodeSettings,omitempty"`
	SystemReinstall         bool                   `json:"systemReinstall,omitempty"`
	InitLoginPassword       string                 `json:"initLoginPassword,omitempty"`
	NetworkType             string                 `json:"networkType,omitempty"`
	AutoGenerateMasterNodes bool                   `json:"autoGenerateMasterNodes,omitempty"`
	Instances               []instances            `json:"instances,omitempty"`
	ExtraInfo               map[string]string      `json:"extraInfo,omitempty"`
	ModuleID                string                 `json:"moduleID,omitempty"`
	ExtraClusterID          string                 `json:"extraClusterID,omitempty"`
	IsCommonCluster         bool                   `json:"isCommonCluster,omitempty"`
	Description             string                 `json:"description,omitempty"`
	ClusterCategory         string                 `json:"clusterCategory,omitempty"`
	IsShared                bool                   `json:"is_shared,omitempty"`
}

type bcsAddons struct {
	System string            `json:"system"`
	Link   string            `json:"link"`
	Params map[string]string `json:"params"`
}

type extraAddons struct {
	System string            `json:"system"`
	Link   string            `json:"link"`
	Params map[string]string `json:"params"`
}

type netWorkSettings struct {
	ClusterIPv4CIDR     string       `json:"clusterIPv4CIDR,omitempty"`
	ServiceIPv4CIDR     string       `json:"serviceIPv4CIDR,omitempty"`
	MaxNodePodNum       uint32       `json:"maxNodePodNum,omitempty"`
	MaxServiceNum       uint32       `json:"maxServiceNum,omitempty"`
	EnableVPCCni        bool         `json:"enableVPCCni,omitempty"`
	EniSubnetIDs        []string     `json:"eniSubnetIDs,omitempty"`
	SubnetSource        subnetSource `json:"subnetSource,omitempty"`
	IsStaticIpMode      bool         `json:"isStaticIpMode,omitempty"`
	ClaimExpiredSeconds uint32       `json:"claimExpiredSeconds,omitempty"`
}

type subnetSource struct {
	New     newSubnet        `json:"new,omitempty"`
	Existed existedSubnetIDs `json:"existed,omitempty"`
}

type newSubnet struct {
	Mask uint32 `json:"mask,omitempty"`
	Zone string `json:"zone,omitempty"`
}

type existedSubnetIDs struct {
	Ids []string `json:"ids,omitempty"`
}

type clusterBasicSettings struct {
	OS          string            `json:"OS"`
	Version     string            `json:"version"`
	ClusterTags map[string]string `json:"clusterTags"`
	VersionName string            `json:"versionName"`
}

type clusterAdvanceSettings struct {
	IPVS             bool              `json:"IPVS"`
	ContainerRuntime string            `json:"containerRuntime"`
	RuntimeVersion   string            `json:"runtimeVersion"`
	ExtraArgs        map[string]string `json:"extraArgs"`
}

type nodeSettings struct {
	DockerGraphPath string            `json:"dockerGraphPath"`
	MountTarget     string            `json:"mountTarget"`
	UnSchedulable   int               `json:"unSchedulable"`
	Labels          map[string]string `json:"labels"`
	ExtraArgs       map[string]string `json:"extraArgs"`
}

type instances struct {
	Region             string `json:"region"`
	Zone               string `json:"zone"`
	VpcID              string `json:"vpcID"`
	SubnetID           string `json:"subnetID"`
	ApplyNum           int    `json:"applyNum"`
	CPU                int    `json:"CPU"`
	Mem                int    `json:"Mem"`
	GPU                int    `json:"GPU"`
	InstanceType       string `json:"instanceType"`
	InstanceChargeType string `json:"instanceChargeType"`
	SystemDisk         struct {
		DiskType string `json:"diskType"`
		DiskSize string `json:"diskSize"`
	} `json:"systemDisk"`
	DataDisks []struct {
		DiskType string `json:"diskType"`
		DiskSize string `json:"diskSize"`
	} `json:"dataDisks"`
	ImageInfo struct {
		ImageID   string `json:"imageID"`
		ImageName string `json:"imageName"`
	} `json:"imageInfo"`
	InitLoginPassword string   `json:"initLoginPassword"`
	SecurityGroupIDs  []string `json:"securityGroupIDs"`
	IsSecurityService bool     `json:"isSecurityService"`
	IsMonitorService  bool     `json:"isMonitorService"`
}

// 查询cm cluster
type cmGetCluster struct {
	ClusterID               string                 `json:"clusterID"`
	ClusterName             string                 `json:"clusterName"`
	FederationClusterID     string                 `json:"federationClusterID"`
	Provider                string                 `json:"provider"`
	Region                  string                 `json:"region"`
	VpcID                   string                 `json:"vpcID"`
	ProjectID               string                 `json:"projectID"`
	BusinessID              string                 `json:"businessID"`
	Environment             string                 `json:"environment"`
	EngineType              string                 `json:"engineType"`
	IsExclusive             bool                   `json:"isExclusive"`
	ClusterType             string                 `json:"clusterType"`
	Labels                  map[string]string      `json:"labels"`
	Creator                 string                 `json:"creator"`
	CreateTime              string                 `json:"createTime"`
	UpdateTime              string                 `json:"updateTime"`
	BcsAddons               bcsAddons              `json:"bcsAddons"`
	ExtraAddons             extraAddons            `json:"extraAddons"`
	SystemID                string                 `json:"systemID"`
	ManageType              string                 `json:"manageType"`
	Master                  map[string]master      `json:"master"`
	NetworkSettings         netWorkSettings        `json:"networkSettings"`
	ClusterBasicSettings    clusterBasicSettings   `json:"clusterBasicSettings"`
	ClusterAdvanceSettings  clusterAdvanceSettings `json:"clusterAdvanceSettings"`
	NodeSettings            nodeSettings           `json:"nodeSettings"`
	Status                  string                 `json:"status"`
	Updater                 string                 `json:"updater"`
	NetworkType             string                 `json:"networkType"`
	AutoGenerateMasterNodes bool                   `json:"autoGenerateMasterNodes"`
	Template                []template             `json:"template"`
	ExtraInfo               map[string]string      `json:"extraInfo"`
	ModuleID                string                 `json:"moduleID"`
	ExtraClusterID          string                 `json:"extraClusterID"`
	IsCommonCluster         bool                   `json:"isCommonCluster"`
	Description             string                 `json:"description"`
	ClusterCategory         string                 `json:"clusterCategory"`
	IsShared                bool                   `json:"is_shared"`
}

type template struct {
	Region             string `json:"region"`
	Zone               string `json:"zone"`
	VpcID              string `json:"vpcID"`
	SubnetID           string `json:"subnetID"`
	ApplyNum           int    `json:"applyNum"`
	CPU                int    `json:"CPU"`
	Mem                int    `json:"Mem"`
	GPU                int    `json:"GPU"`
	InstanceType       string `json:"instanceType"`
	InstanceChargeType string `json:"instanceChargeType"`
	SystemDisk         struct {
		DiskType string `json:"diskType"`
		DiskSize string `json:"diskSize"`
	} `json:"systemDisk"`
	DataDisks []struct {
		DiskType string `json:"diskType"`
		DiskSize string `json:"diskSize"`
	} `json:"dataDisks"`
	ImageInfo struct {
		ImageID   string `json:"imageID"`
		ImageName string `json:"imageName"`
	} `json:"imageInfo"`
	InitLoginPassword string   `json:"initLoginPassword"`
	SecurityGroupIDs  []string `json:"securityGroupIDs"`
	IsSecurityService bool     `json:"isSecurityService"`
	IsMonitorService  bool     `json:"isMonitorService"`
}

type master struct {
	NodeID       string `json:"nodeID"`
	InnerIP      string `json:"innerIP"`
	InstanceType string `json:"instanceType"`
	CPU          int    `json:"CPU"`
	Mem          int    `json:"mem"`
	GPU          int    `json:"GPU"`
	Status       string `json:"status"`
	ZoneID       string `json:"zoneID"`
	NodeGroupID  string `json:"nodeGroupID"`
	ClusterID    string `json:"clusterID"`
	VPC          string `json:"VPC"`
	Region       string `json:"region"`
	Passwd       string `json:"passwd"`
	Zone         int    `json:"zone"`
	DeviceID     string `json:"deviceID"`
}

// cm 创建node
type cmCreateNode struct {
	ClusterID         string   `json:"clusterID,omitempty"`
	Nodes             []string `json:"nodes,omitempty"`
	InitLoginPassword string   `json:"initLoginPassword,omitempty"`
	NodeGroupID       string   `json:"nodeGroupID,omitempty"`
	OnlyCreateInfo    bool     `json:"onlyCreateInfo,omitempty"`
	Operator          string   `json:"operator,omitempty"`
}
