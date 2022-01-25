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

package passcc

const (
	// State state
	State = "bcs_new"
	// Status status
	Status = "normal"
)

const (
	defaultEsbURL         = "http://9.140.129.207:8081"
	defaultWebhookImage   = "xxx.com:8090/public/bcs/k8s/bcs-webhook-server:1.2.0"
	defaultPrivilegeImage = "xxx.com:8090/public/bcs/k8s/gcs-privilege:1.0.0"
)

// ClusterParamsRequest xxx
type ClusterParamsRequest struct {
	ClusterID          string           `json:"cluster_id"`
	ClusterName        string           `json:"name"`
	ClusterDescription string           `json:"description"`
	AreaID             int              `json:"area_id"`
	VpcID              string           `json:"vpc_id"`
	Env                string           `json:"environment"`
	MasterIPs          []ManagerMasters `json:"master_ips"`
	NeedNAT            bool             `json:"need_nat"`
	Version            string           `json:"version"`
	NetworkType        string           `json:"network_type"`
	Coes               string           `json:"coes"`
	KubeProxyMode      string           `json:"kube_proxy_mode"`
	Creator            string           `json:"creator"`
	Type               string           `json:"type"`
	ExtraClusterID     string           `json:"extra_cluster_id"`
	State              string           `json:"state"`
	Status             string           `json:"status"`
}

// ManagerMasters masterIP
type ManagerMasters struct {
	InnerIP string `json:"inner_ip"`
}

// CreateClusterConfParams xxx
type CreateClusterConfParams struct {
	Creator   string `json:"creator"`
	ClusterID string `json:"cluster_id"`
	Configure string `json:"configure"`
}

// ClusterSnapShootInfo snapInfo
type ClusterSnapShootInfo struct {
	Regions                 string              `json:"regions"`
	ClusterID               string              `json:"cluster_id"`
	MasterIPList            []string            `json:"master_ip_list"`
	VpcID                   string              `json:"vpc_id"`
	SystemDataID            uint32              `json:"bcs_system_data_id"`
	ClusterCIDRSettings     ClusterCIDRInfo     `json:"ClusterCIDRSettings"`
	ClusterType             string              `json:"ClusterType"`
	ClusterBasicSettings    ClusterBasicInfo    `json:"ClusterBasicSettings"`
	ClusterAdvancedSettings ClusterAdvancedInfo `json:"ClusterAdvancedSettings"`
	NetWorkType             string              `json:"network_type"`
	EsbURL                  string              `json:"esb_url"`
	WebhookImage            string              `json:"bcs_webhook_image"`
	PrivilegeImage          string              `json:"gcs_privilege_image"`
	VersionName             string              `json:"version_name"`
	Version                 string              `json:"version"`
	ClusterVersion          string              `json:"ClusterVersion"`
	ControlIP               string              `json:"control_ip"`
	MasterIPs               []string            `json:"master_ips"`
	Env                     string              `json:"environment"`
	ProjectName             string              `json:"product_name"`
	ProjectCode             string              `json:"project_code"`
	AreaName                string              `json:"area_name"`
	ExtraClusterID          string              `json:"extra_cluster_id"`
}

// ClusterCIDRInfo cidrInfo
type ClusterCIDRInfo struct {
	ClusterCIDR          string `json:"ClusterCIDR"`
	MaxNodePodNum        uint32 `json:"MaxNodePodNum"`
	MaxClusterServiceNum uint32 `json:"MaxClusterServiceNum"`
}

// ClusterBasicInfo basicInfo
type ClusterBasicInfo struct {
	ClusterOS      string `json:"ClusterOs"`
	ClusterVersion string `json:"ClusterVersion"`
	ClusterName    string `json:"ClusterName"`
}

// ClusterAdvancedInfo advancedInfo
type ClusterAdvancedInfo struct {
	IPVS bool `json:"IPVS"`
}

var testAreaCode = map[string]int{
	"ap-guangzhou": 7,
	"ap-shanghai":  8,
	"ap-shenzhen":  9,
	"ap-nanjing":   10,
}

var prodAreaCode = map[string]int{
	"ap-chongqing":     24,
	"ap-guangzhou":     17,
	"ap-nanjing":       21,
	"ap-seoul":         25,
	"ap-shanghai":      18,
	"ap-shenyang-ec":   27,
	"ap-shenzhen":      22,
	"ap-singapore":     19,
	"ap-tianjin":       20,
	"ap-tokyo":         28,
	"ap-xian-ec":       26,
	"na-siliconvalley": 23,
}

// DeleteClusterRequest deleteCluster
type DeleteClusterRequest struct {
	ProjectID string
	ClusterID string
}

// CommonResp common resp
type CommonResp struct {
	Code      uint   `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}
