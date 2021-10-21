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

package u1_21_202109291130

import (
	"encoding/json"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

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
	Creator              string `json:"creator"` // required
	Master               []string
	Node                 []string
	NetworkSettings      CreateClustersNetworkSettings      `json:"networkSettings"` // TODO 待定
	ClusterBasicSettings CreateClustersClusterBasicSettings `json:"clusterBasicSettings"`
}

type bcsReqUpdateCluster struct {
	bcsClusterBase
	NetworkSettings        CreateClustersNetworkSettings      `json:"networkSettings"` // TODO 待定
	ClusterBasicSettings   CreateClustersClusterBasicSettings `json:"clusterBasicSettings"`
	Updater                string                             `json:"updater"`
	Master                 []string
	Node                   []string
	Labels                 interface{} `json:"labels,omitempty"`
	BcsAddons              interface{} `json:"bcsAddons,omitempty"`
	ExtraAddons            interface{} `json:"extraAddons,omitempty"`
	ClusterAdvanceSettings interface{} `json:"clusterAdvanceSettings,omitempty"`
	NodeSettings           interface{} `json:"nodeSettings,omitempty"`
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
	NetworkSettings        CreateClustersNetworkSettings      `json:"networkSettings"` // TODO 待定
	ClusterBasicSettings   CreateClustersClusterBasicSettings `json:"clusterBasicSettings"`
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
	SystemID string                     `json:"systemID"`
	Master   []bcsRespFindClusterMaster `json:"master"`
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

type BCSCluster struct {
	ClusterID            string                             `json:"clusterID"`   // required
	ClusterName          string                             `json:"clusterName"` // required
	Provider             string                             `json:"provider"`    // required
	Region               string                             `json:"region"`      // required
	VpcID                string                             `json:"vpcID"`
	ProjectID            string                             `json:"projectID"`   // required
	BusinessID           string                             `json:"businessID"`  // required
	Environment          string                             `json:"environment"` // required
	EngineType           string                             `json:"engineType"`  // required
	IsExclusive          bool                               `json:"isExclusive"` // required
	ClusterType          string                             `json:"clusterType"` // required
	FederationClusterID  string                             `json:"federationClusterID"`
	Creator              string                             `json:"creator"` // required
	OnlyCreateInfo       bool                               `json:"onlyCreateInfo"`
	CloudID              string                             `json:"cloudID"`
	ManageType           string                             `json:"manageType"`
	Master               []string                           `json:"master"`
	Nodes                []string                           `json:"nodes"`
	SystemReinstall      bool                               `json:"systemReinstall"`
	InitLoginPassword    string                             `json:"initLoginPassword"`
	NetworkType          string                             `json:"networkType"`
	NetworkSettings      CreateClustersNetworkSettings      `json:"networkSettings"` // TODO 待定
	ClusterBasicSettings CreateClustersClusterBasicSettings `json:"clusterBasicSettings"`

	Labels                  interface{} `json:"labels,omitempty"`
	BcsAddons               interface{} `json:"bcsAddons,omitempty"`
	ExtraAddons             interface{} `json:"extraAddons,omitempty"`
	ClusterAdvanceSettings  interface{} `json:"clusterAdvanceSettings,omitempty"`
	NodeSettings            interface{} `json:"nodeSettings,omitempty"`
	AutoGenerateMasterNodes bool        `json:"autoGenerateMasterNodes"` // 创建集群是否使用已存在节点, 默认false, 即使用已经存在的节点, 从创建集群参数中获取
	Instances               interface{} `json:"instances,omitempty"`     // 创建集群时 autoGenerateMasterNodes 为true, 系统自动生成master节点, 需要指定instances生成的配置信息,支持不同可用区实例"
	ExtraInfo               interface{} `json:"ExtraInfo"`
	MasterInstanceID        []string    `json:"masterInstanceID"` // 集群master节点的Instance id

	Updater  string `json:"updater"`
	Status   string `json:"status"`   //"集群状态，可能状态CREATING，RUNNING，DELETING，FALURE，INITIALIZATION，DELETED"
	SystemID string `json:"systemID"` // kubernetes集群在各云平台上资源ID
}

type CreateClustersNetworkSettings struct {
	ClusterIPv4CIDR string `json:"clusterIPv4CIDR"`
	ServiceIPv4CIDR string `json:"serviceIPv4CIDR"`
	MaxNodePodNum   string `json:"maxNodePodNum"`
	MaxServiceNum   string `json:"maxServiceNum"`
}

type CreateClustersClusterBasicSettings struct {
	OS          string            `json:"OS"`
	Version     string            `json:"version"`
	ClusterTags map[string]string `json:"clusterTags"`
}

type create struct {
	ClusterID           string      `json:"clusterID"`
	ClusterName         string      `json:"clusterName"`
	Provider            string      `json:"provider"`
	Region              string      `json:"region"`
	VpcID               string      `json:"vpcID"`
	ProjectID           string      `json:"projectID"`
	BusinessID          string      `json:"businessID"`
	Environment         string      `json:"environment"`
	EngineType          string      `json:"engineType"`
	IsExclusive         bool        `json:"isExclusive"`
	ClusterType         string      `json:"clusterType"`
	FederationClusterID string      `json:"federationClusterID"`
	Labels              interface{} `json:"labels,omitempty"`
	Creator             string      `json:"creator"`
	OnlyCreateInfo      bool        `json:"onlyCreateInfo"`
	BcsAddons           struct {
		AdditionalProp1 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp1"`
		AdditionalProp2 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp2"`
		AdditionalProp3 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp3"`
	} `json:"bcsAddons,omitempty"`
	ExtraAddons struct {
		AdditionalProp1 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp1"`
		AdditionalProp2 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp2"`
		AdditionalProp3 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp3"`
	} `json:"extraAddons,omitempty"`
	CloudID         string   `json:"cloudID"`
	ManageType      string   `json:"manageType"`
	Master          []string `json:"master"`
	Nodes           []string `json:"nodes"`
	NetworkSettings struct {
		ClusterIPv4CIDR string `json:"clusterIPv4CIDR"`
		ServiceIPv4CIDR string `json:"serviceIPv4CIDR"`
		MaxNodePodNum   string `json:"maxNodePodNum"`
		MaxServiceNum   string `json:"maxServiceNum"`
	} `json:"networkSettings,omitempty"`
	ClusterBasicSettings struct {
		OS          string `json:"OS"`
		Version     string `json:"version"`
		ClusterTags struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"clusterTags"`
		VersionName string `json:"versionName"`
	} `json:"clusterBasicSettings,omitempty"`
	ClusterAdvanceSettings struct {
		IPVS             bool   `json:"IPVS"`
		ContainerRuntime string `json:"containerRuntime"`
		RuntimeVersion   string `json:"runtimeVersion"`
		ExtraArgs        struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"extraArgs"`
	} `json:"clusterAdvanceSettings,omitempty"`
	NodeSettings struct {
		DockerGraphPath string `json:"dockerGraphPath"`
		MountTarget     string `json:"mountTarget"`
		UnSchedulable   int    `json:"unSchedulable"`
		Labels          struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"labels"`
		ExtraArgs struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"extraArgs"`
	} `json:"nodeSettings,omitempty"`
	SystemReinstall         bool   `json:"systemReinstall"`
	InitLoginPassword       string `json:"initLoginPassword"`
	NetworkType             string `json:"networkType"`
	AutoGenerateMasterNodes bool   `json:"autoGenerateMasterNodes"`
	Instances               []struct {
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
	} `json:"instances,omitempty"`
	ExtraInfo struct {
		AdditionalProp1 string `json:"additionalProp1"`
		AdditionalProp2 string `json:"additionalProp2"`
		AdditionalProp3 string `json:"additionalProp3"`
	} `json:"ExtraInfo"`
	MasterInstanceID []string `json:"masterInstanceID"`

	Updater  string `json:"updater"`
	Status   string `json:"status"`   //"集群状态，可能状态CREATING，RUNNING，DELETING，FALURE，INITIALIZATION，DELETED"
	SystemID string `json:"systemID"` // kubernetes集群在各云平台上资源ID

}

type get struct {
	ClusterID           string `json:"clusterID"`
	ClusterName         string `json:"clusterName"`
	FederationClusterID string `json:"federationClusterID"`
	Provider            string `json:"provider"`
	Region              string `json:"region"`
	VpcID               string `json:"vpcID"`
	ProjectID           string `json:"projectID"`
	BusinessID          string `json:"businessID"`
	Environment         string `json:"environment"`
	EngineType          string `json:"engineType"`
	IsExclusive         bool   `json:"isExclusive"`
	ClusterType         string `json:"clusterType"`
	Labels              struct {
		AdditionalProp1 string `json:"additionalProp1"`
		AdditionalProp2 string `json:"additionalProp2"`
		AdditionalProp3 string `json:"additionalProp3"`
	} `json:"labels"`
	Creator    string `json:"creator"`
	CreateTime string `json:"createTime"`
	UpdateTime string `json:"updateTime"`
	BcsAddons  struct {
		AdditionalProp1 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp1"`
		AdditionalProp2 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp2"`
		AdditionalProp3 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp3"`
	} `json:"bcsAddons"`
	ExtraAddons struct {
		AdditionalProp1 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp1"`
		AdditionalProp2 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp2"`
		AdditionalProp3 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp3"`
	} `json:"extraAddons"`
	SystemID   string `json:"systemID"`
	ManageType string `json:"manageType"`
	Master     struct {
		AdditionalProp1 struct {
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
		} `json:"additionalProp1"`
		AdditionalProp2 struct {
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
		} `json:"additionalProp2"`
		AdditionalProp3 struct {
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
		} `json:"additionalProp3"`
	} `json:"master"`
	NetworkSettings struct {
		ClusterIPv4CIDR string `json:"clusterIPv4CIDR"`
		ServiceIPv4CIDR string `json:"serviceIPv4CIDR"`
		MaxNodePodNum   string `json:"maxNodePodNum"`
		MaxServiceNum   string `json:"maxServiceNum"`
	} `json:"networkSettings"`
	ClusterBasicSettings struct {
		OS          string `json:"OS"`
		Version     string `json:"version"`
		ClusterTags struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"clusterTags"`
		VersionName string `json:"versionName"`
	} `json:"clusterBasicSettings"`
	ClusterAdvanceSettings struct {
		IPVS             bool   `json:"IPVS"`
		ContainerRuntime string `json:"containerRuntime"`
		RuntimeVersion   string `json:"runtimeVersion"`
		ExtraArgs        struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"extraArgs"`
	} `json:"clusterAdvanceSettings"`
	NodeSettings struct {
		DockerGraphPath string `json:"dockerGraphPath"`
		MountTarget     string `json:"mountTarget"`
		UnSchedulable   int    `json:"unSchedulable"`
		Labels          struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"labels"`
		ExtraArgs struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"extraArgs"`
	} `json:"nodeSettings"`
	Status                  string `json:"status"`
	Updater                 string `json:"updater"`
	NetworkType             string `json:"networkType"`
	AutoGenerateMasterNodes bool   `json:"autoGenerateMasterNodes"`
	Template                []struct {
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
	} `json:"template"`
	ExtraInfo struct {
		AdditionalProp1 string `json:"additionalProp1"`
		AdditionalProp2 string `json:"additionalProp2"`
		AdditionalProp3 string `json:"additionalProp3"`
	} `json:"ExtraInfo"`
	MasterInstanceID []string `json:"masterInstanceID"`
}

type update struct {
	ClusterID           string `json:"clusterID"`
	ClusterName         string `json:"clusterName"`
	Provider            string `json:"provider"`
	Region              string `json:"region"`
	VpcID               string `json:"vpcID"`
	ProjectID           string `json:"projectID"`
	BusinessID          string `json:"businessID"`
	Environment         string `json:"environment"`
	EngineType          string `json:"engineType"`
	IsExclusive         bool   `json:"isExclusive"`
	ClusterType         string `json:"clusterType"`
	FederationClusterID string `json:"federationClusterID"`
	Labels              struct {
		AdditionalProp1 string `json:"additionalProp1"`
		AdditionalProp2 string `json:"additionalProp2"`
		AdditionalProp3 string `json:"additionalProp3"`
	} `json:"labels"`
	Updater   string `json:"updater"`
	Status    string `json:"status"`
	BcsAddons struct {
		AdditionalProp1 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp1"`
		AdditionalProp2 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp2"`
		AdditionalProp3 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp3"`
	} `json:"bcsAddons"`
	ExtraAddons struct {
		AdditionalProp1 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp1"`
		AdditionalProp2 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp2"`
		AdditionalProp3 struct {
			System string `json:"system"`
			Link   string `json:"link"`
			Params struct {
				AdditionalProp1 string `json:"additionalProp1"`
				AdditionalProp2 string `json:"additionalProp2"`
				AdditionalProp3 string `json:"additionalProp3"`
			} `json:"params"`
		} `json:"additionalProp3"`
	} `json:"extraAddons"`
	SystemID        string   `json:"systemID"`
	ManageType      string   `json:"manageType"`
	Master          []string `json:"master"`
	NetworkSettings struct {
		ClusterIPv4CIDR string `json:"clusterIPv4CIDR"`
		ServiceIPv4CIDR string `json:"serviceIPv4CIDR"`
		MaxNodePodNum   string `json:"maxNodePodNum"`
		MaxServiceNum   string `json:"maxServiceNum"`
	} `json:"networkSettings"`
	ClusterBasicSettings struct {
		OS          string `json:"OS"`
		Version     string `json:"version"`
		ClusterTags struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"clusterTags"`
		VersionName string `json:"versionName"`
	} `json:"clusterBasicSettings"`
	ClusterAdvanceSettings struct {
		IPVS             bool   `json:"IPVS"`
		ContainerRuntime string `json:"containerRuntime"`
		RuntimeVersion   string `json:"runtimeVersion"`
		ExtraArgs        struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"extraArgs"`
	} `json:"clusterAdvanceSettings"`
	NodeSettings struct {
		DockerGraphPath string `json:"dockerGraphPath"`
		MountTarget     string `json:"mountTarget"`
		UnSchedulable   int    `json:"unSchedulable"`
		Labels          struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"labels"`
		ExtraArgs struct {
			AdditionalProp1 string `json:"additionalProp1"`
			AdditionalProp2 string `json:"additionalProp2"`
			AdditionalProp3 string `json:"additionalProp3"`
		} `json:"extraArgs"`
	} `json:"nodeSettings"`
	NetworkType string `json:"networkType"`
	ExtraInfo   struct {
		AdditionalProp1 string `json:"additionalProp1"`
		AdditionalProp2 string `json:"additionalProp2"`
		AdditionalProp3 string `json:"additionalProp3"`
	} `json:"ExtraInfo"`
}

// 比较cluster
func diffCluster(ccData bcsReqCreateCluster, bcsData bcsRespFindCluster) (bool, *bcsReqUpdateCluster, error) {

	// 对比基础数据
	if ccData.bcsClusterBase != bcsData.bcsClusterBase {
		clusters := &bcsReqUpdateCluster{
			bcsClusterBase:          ccData.bcsClusterBase,
			NetworkSettings:         ccData.NetworkSettings,
			ClusterBasicSettings:    ccData.ClusterBasicSettings,
			Updater:                 bcsData.Updater,
			Master:                  ccData.Master,
			Node:                    ccData.Node,
			Labels:                  bcsData.Labels,
			BcsAddons:               bcsData.BcsAddons,
			ExtraAddons:             bcsData.ExtraAddons,
			ClusterAdvanceSettings:  bcsData.ClusterAdvanceSettings,
			NodeSettings:            bcsData.NodeSettings,
			AutoGenerateMasterNodes: bcsData.AutoGenerateMasterNodes,
			Instances:               bcsData.Instances,
			ExtraInfo:               bcsData.ExtraInfo,
			MasterInstanceID:        bcsData.MasterInstanceID,
			Status:                  bcsData.Status,
			SystemID:                bcsData.SystemID,
		}
		return true, clusters, nil
	}

	if len(bcsData.Master) == len(ccData.Master) {
		return false, nil, nil
	}

	// 对比master
	bcsMasterIPMap := make(map[string]string)
	for _, master := range bcsData.Master {
		bcsMasterIPMap[master.InnerIP] = master.InnerIP
	}

	for _, ip := range ccData.Master {
		if _, ok := bcsMasterIPMap[ip]; !ok {
			clusters := &bcsReqUpdateCluster{
				bcsClusterBase:          ccData.bcsClusterBase,
				NetworkSettings:         ccData.NetworkSettings,
				ClusterBasicSettings:    ccData.ClusterBasicSettings,
				Updater:                 bcsData.Updater,
				Master:                  ccData.Master,
				Node:                    ccData.Node,
				Labels:                  bcsData.Labels,
				BcsAddons:               bcsData.BcsAddons,
				ExtraAddons:             bcsData.ExtraAddons,
				ClusterAdvanceSettings:  bcsData.ClusterAdvanceSettings,
				NodeSettings:            bcsData.NodeSettings,
				AutoGenerateMasterNodes: bcsData.AutoGenerateMasterNodes,
				Instances:               bcsData.Instances,
				ExtraInfo:               bcsData.ExtraInfo,
				MasterInstanceID:        bcsData.MasterInstanceID,
				Status:                  bcsData.Status,
				SystemID:                bcsData.SystemID,
			}
			return true, clusters, nil
		}
	}

	return false, nil, nil
}

func genCluster(projectID, clusterID, ccAppID string) (*bcsReqCreateCluster, error) {
	ccCluster, err := clusterInfo(projectID, clusterID)
	if err != nil {
		blog.Errorf("get cc cluster(%s) data failed, err: %v", clusterID, err)
		return nil, err
	}

	masterList, err := allMasterList()
	if err != nil {
		blog.Errorf("get cc cluster(%s) master List failed, err: %v", clusterID, err)
		return nil, err
	}
	masterIP := make([]string, 0)
	for _, data := range masterList {
		if data.ClusterId == clusterID {
			masterIP = append(masterIP, data.InnerIp)
		}
	}

	nodeList, err := allNodeList()
	if err != nil {
		blog.Errorf("get cc cluster(%s) node failed, err: %v", clusterID, err)
		return nil, err
	}
	nodeIP := make([]string, 0)
	for _, data := range nodeList {
		if data.ClusterId == clusterID {
			nodeIP = append(nodeIP, data.InnerIp)
		}
	}

	configVersion, err := versionConfig(clusterID)
	if err != nil {
		blog.Errorf("get cc cluster(%s) config version failed, err: %v", clusterID, err)
		return nil, err
	}

	versionConfigure := new(versionConfigure)
	err = json.Unmarshal([]byte(configVersion.Configure), versionConfigure)
	if err != nil {
		blog.Errorf("config version deJson failed, err: %v", clusterID, err)
		return nil, err
	}

	cluster := &bcsReqCreateCluster{
		bcsClusterBase: bcsClusterBase{
			ClusterID:           ccCluster.ClusterID,
			ClusterName:         ccCluster.Name,
			Provider:            "bcs",
			Region:              "", // TODO 待定
			VpcID:               versionConfigure.VpcID,
			ProjectID:           ccCluster.ProjectId,
			BusinessID:          ccAppID,
			Environment:         ccCluster.Environment,
			EngineType:          "k8s",
			IsExclusive:         false,
			ClusterType:         "single",
			FederationClusterID: "",
		},
		Creator: ccCluster.Creator,
		Master:  masterIP,
		Node:    nodeIP,
		NetworkSettings: CreateClustersNetworkSettings{
			ClusterIPv4CIDR: "",
			ServiceIPv4CIDR: "",
			MaxNodePodNum:   "",
			MaxServiceNum:   "",
		},
		ClusterBasicSettings: CreateClustersClusterBasicSettings{
			OS:          "",
			Version:     "1.12.3", // 默认版本
			ClusterTags: map[string]string{},
		},
	}

	return cluster, nil
}
