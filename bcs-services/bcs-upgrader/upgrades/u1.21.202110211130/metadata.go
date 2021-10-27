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

package u1_21_202110211130

import (
	"time"
)

type allClusterData struct {
	Code        string                      `json:"code"`
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	ClusterList []allClusterDataClusterList `json:"cluster_list"`
}

type allClusterDataClusterList struct {
	ID       string `json:"id"`
	IsPublic bool   `json:"is_public"`
	Name     string `json:"name"`
}

type allMasterListData struct {
	ClusterId string `json:"cluster_id"`
	InnerIp   string `json:"inner_ip"`
	Status    string `json:"status"`
}

type versionConfigData struct {
	ClusterId string    `json:"cluster_id"`
	Configure string    `json:"configure"`
	CreatedAt time.Time `json:"created_at"`
	Creator   string    `json:"creator"`
	ID        int       `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

type versionConfigure struct {
	AreaID string `json:"area_id"`
	VpcID  string `json:"vpc_id"`
}

type clustersInfoData struct {
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

type nodeListData struct {
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

type reqCreateNode struct {
	ClusterID         string   `json:"clusterID"`
	Nodes             []string `json:"nodes"`
	InitLoginPassword string   `json:"initLoginPassword"`
	NodeGroupID       string   `json:"nodeGroupID"`
	OnlyCreateInfo    bool     `json:"onlyCreateInfo"`
}

type reqDeleteNode struct {
	ClusterID string   `json:"clusterID"`
	Nodes     []string `json:"nodes"`
	// 删除模式，RETAIN(移除集群，但是保留主机)，TERMINATE(只支持按量计费的机器)，默认是RETAIN
	DeleteMode string `json:"deleteMode"`
	// 不管节点处于任何状态都强制删除，例如可能刚初始化，NotReady等
	IsForce  bool   `json:"isForce"`
	Operator string `json:"operator"` // 操作者
	//默认为false。设置为true时，仅删除cluster-manager所记录的信息，不会触发任何自动化流程.
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

type bcsReqCreateCluster struct {
	bcsClusterBase
	Creator              string                             `json:"creator"` // required
	Master               []string                           `json:"master"`
	Node                 []string                           `json:"node"`
	NetworkSettings      createClustersNetworkSettings      `json:"networkSettings"` // TODO 待定
	ClusterBasicSettings createClustersClusterBasicSettings `json:"clusterBasicSettings"`
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

type respAllProjectData struct {
	Count   int         `json:"count"`
	Results []ccProject `json:"results"`
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

type bcsProject struct {
	ProjectID   string      `json:"projectID"`   // required
	Name        string      `json:"name"`        // required
	EnglishName string      `json:"englishName"` // required
	Creator     string      `json:"creator"`     // required
	ProjectType int         `json:"projectType"` // required
	UseBKRes    bool        `json:"useBKRes"`    // required
	Description string      `json:"description"` // required
	IsOffline   bool        `json:"isOffline"`
	Kind        string      `json:"kind"`
	BusinessID  string      `json:"businessID"` // required
	DeployType  int         `json:"deployType"` // required
	BgID        string      `json:"bgID"`
	BgName      string      `json:"bgName"`
	DeptID      string      `json:"deptID"`
	DeptName    string      `json:"deptName"`
	CenterID    string      `json:"centerID"`
	CenterName  string      `json:"centerName"`
	IsSecret    bool        `json:"isSecret"`
	Updater     string      `json:"updater"` // update/get
	Credentials interface{} `json:"credentials"`
}
