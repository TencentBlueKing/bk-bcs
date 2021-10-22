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

	mapset "github.com/deckarep/golang-set"
)

type respGetCCToken struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		Identity     struct {
			Username string `json:"username"`
			UserType string `json:"user_type"`
		} `json:"identity"`
	} `json:"data"`
}

/************ common ***********/

type ccBaseResp struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type bcsBaseResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Result  bool   `json:"result"`
}

/************ common ***********/

/************ CC Cluster ***********/

type RespAllCluster struct {
	bcsBaseResp `json:",inline"`
	Data        []AllClusterData `json:"data"`
}

type AllClusterData struct {
	Code        string                      `json:"code"`
	ID          string                      `json:"id"`
	Name        string                      `json:"name"`
	ClusterList []AllClusterDataClusterList `json:"cluster_list"`
}

type RespAllMasterList struct {
	bcsBaseResp `json:",inline"`
	Data        []AllMasterListData `json:"data"`
}

type AllClusterDataClusterList struct {
	ID       string `json:"id"`
	IsPublic bool   `json:"is_public"`
	Name     string `json:"name"`
}

type AllMasterListData struct {
	ClusterId string `json:"cluster_id"`
	InnerIp   string `json:"inner_ip"`
	Status    string `json:"status"`
}

type respVersionConfig struct {
	bcsBaseResp `json:",inline"`
	Data        versionConfigData `json:"data"`
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

type respClustersInfo struct {
	bcsBaseResp `json:",inline"`
	Data        clustersInfoData `json:"data"`
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

/************ CC Cluster ***********/

/************ BCS Cluster ***********/

type searchClustersByIDData struct {
	ClusterID           string                                  `json:"clusterID"`   // required
	ClusterName         string                                  `json:"clusterName"` // required
	FederationClusterID string                                  `json:"federationClusterID"`
	Provider            string                                  `json:"provider"` // required
	Region              string                                  `json:"region"`   // required
	VpcID               string                                  `json:"vpcID"`
	ProjectID           string                                  `json:"projectID"`   // required
	BusinessID          string                                  `json:"businessID"`  // required
	Environment         string                                  `json:"environment"` // required
	EngineType          string                                  `json:"engineType"`  // required
	IsExclusive         bool                                    `json:"isExclusive"` // required
	ClusterType         string                                  `json:"clusterType"` // required
	Creator             string                                  `json:"creator"`     // required
	CreateTime          string                                  `json:"createTime"`
	UpdateTime          string                                  `json:"updateTime"`
	SystemID            string                                  `json:"systemID"`
	ManageType          string                                  `json:"manageType"`
	Master              map[string]searchClustersByIDDataMaster `json:"master"`
	Status              string                                  `json:"status"`
	Updater             string                                  `json:"updater"`
	NetworkType         string                                  `json:"networkType"`
}

type searchClustersByIDDataMaster struct {
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
}

type respFindCluster struct {
	bcsBaseResp `json:",inline"`
	Data        bcsRespFindCluster `json:"data"`
}

type ReqUpdateCluster struct {
	ClusterID            string                             `json:"clusterID"`
	ClusterName          string                             `json:"clusterName"`
	Provider             string                             `json:"provider"`
	Region               string                             `json:"region"`
	VpcID                string                             `json:"vpcID"`
	ProjectID            string                             `json:"projectID"`
	BusinessID           string                             `json:"businessID"`
	Environment          string                             `json:"environment"`
	EngineType           string                             `json:"engineType"`
	IsExclusive          bool                               `json:"isExclusive"`
	ClusterType          string                             `json:"clusterType"`
	FederationClusterID  string                             `json:"federationClusterID"`
	Updater              string                             `json:"updater"`
	Status               string                             `json:"status"`
	SystemID             string                             `json:"systemID"`
	ManageType           string                             `json:"manageType"`
	Master               []string                           `json:"master"`
	NetworkSettings      CreateClustersNetworkSettings      `json:"network_settings"`
	ClusterBasicSettings CreateClustersClusterBasicSettings `json:"cluster_basic_settings"`
	NetworkType          string                             `json:"networkType"`
}

/************ BCS Cluster ***********/

/************ node ***********/

type RespNodeList struct {
	bcsBaseResp `json:",inline"`
	Data        []NodeListData `json:"data"`
}

type NodeListData struct {
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

type ReqCreateNode struct {
	ClusterID         string   `json:"clusterID"`
	Nodes             []string `json:"nodes"`
	InitLoginPassword string   `json:"initLoginPassword"`
	NodeGroupID       string   `json:"nodeGroupID"`
	OnlyCreateInfo    bool     `json:"onlyCreateInfo"`
}

type ReqDeleteNode struct {
	ClusterID      string   `json:"clusterID"`
	Nodes          []string `json:"nodes"`
	DeleteMode     string   `json:"deleteMode"`     // 删除模式，RETAIN(移除集群，但是保留主机)，TERMINATE(只支持按量计费的机器)，默认是RETAIN
	IsForce        bool     `json:"isForce"`        // 不管节点处于任何状态都强制删除，例如可能刚初始化，NotReady等
	Operator       string   `json:"operator"`       // 操作者
	OnlyDeleteInfo bool     `json:"onlyDeleteInfo"` //默认为false。设置为true时，仅删除cluster-manager所记录的信息，不会触发任何自动化流程.
}

/************ node ***********/

/************ CC project ***********/

type respSearchProjectByID struct {
	bcsBaseResp `json:",inline"`
	Data        bcsProject `json:"data"`
}

/************ CC project ***********/

/************ BCS project ***********/

/************ BCS project end ***********/

/************ BCS node ***********/

type bcsRespNodeList struct {
	bcsBaseResp `json:",inline"`
	Data        []bcsNodeListData `json:"data"`
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

/************ BCS node end ***********/

// node

func diffNode(ccNodeIPS, bcsNodeIPS []string, clusterID string) (createNode *ReqCreateNode, deleteNode *ReqDeleteNode) {

	alreadySet := mapset.NewSet()
	for _, ip := range ccNodeIPS {
		alreadySet.Add(ip)
	}
	newSet := mapset.NewSet()
	for _, ip := range bcsNodeIPS {
		newSet.Add(ip)
	}

	toCreateSet := newSet.Difference(alreadySet)
	toDeleteSet := alreadySet.Difference(newSet)
	toCreateIt := toCreateSet.Iterator()
	toDeleteIt := toDeleteSet.Iterator()
	var toCreateArray, toDeleteArray []string
	for elem := range toCreateIt.C {
		toCreateArray = append(toCreateArray, elem.(string))
	}
	for elem := range toDeleteIt.C {
		toDeleteArray = append(toDeleteArray, elem.(string))
	}

	if len(toCreateArray) != 0 {
		createNode = &ReqCreateNode{
			ClusterID:         clusterID,
			Nodes:             toCreateArray,
			InitLoginPassword: "",
			NodeGroupID:       "",
			OnlyCreateInfo:    true,
		}
	}

	if len(toDeleteArray) != 0 {
		deleteNode = &ReqDeleteNode{
			ClusterID:      clusterID,
			Nodes:          toDeleteArray,
			DeleteMode:     "",
			IsForce:        false, // TODO 参数待确认
			Operator:       "",
			OnlyDeleteInfo: false, // TODO 参数待确认
		}
	}

	return createNode, deleteNode
}
