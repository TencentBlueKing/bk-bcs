/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package paascc

import (
	"time"
)

// BaseResp base resp for paas cc
type BaseResp struct {
	RequestID string `json:"request_id"`
	Message   string `json:"message"`
	Result    bool   `json:"result"`
	Code      int    `json:"code"`
}

// Project project
type Project struct {
	ApprovalStatus int64     `json:"approval_status"`
	ApprovalTime   time.Time `json:"approval_time"`
	Approver       string    `json:"approver"`
	BgID           int64     `json:"bg_id"`
	BgName         string    `json:"bg_name"`
	Bgid           int64     `json:"bgid"`
	CcAppID        int64     `json:"cc_app_id"`
	CenterID       int64     `json:"center_id"`
	CenterName     string    `json:"center_name"`
	CreatedAt      time.Time `json:"created_at"`
	Creator        string    `json:"creator"`
	DataID         int64     `json:"data_id"`
	DeployType     string    `json:"deploy_type"`
	DeptID         int64     `json:"dept_id"`
	DeptName       string    `json:"dept_name"`
	Description    string    `json:"description"`
	EnglishName    string    `json:"english_name"`
	ID             int64     `json:"id"`
	IsOfflined     bool      `json:"is_offlined"`
	IsSecrecy      bool      `json:"is_secrecy"`
	Kind           int64     `json:"kind"`
	LogoAddr       string    `json:"logo_addr"`
	Name           string    `json:"name"`
	ProjectID      string    `json:"project_id"`
	ProjectName    string    `json:"project_name"`
	ProjectType    int64     `json:"project_type"`
	Remark         string    `json:"remark"`
	UpdatedAt      time.Time `json:"updated_at"`
	Updator        string    `json:"updator"`
	UseBk          bool      `json:"use_bk"`
}

// ListProjectsResult ListProjects result
type ListProjectsResult struct {
	BaseResp
	Data []*Project `json:"data"`
}

// GetProjectResult GetProject result
type GetProjectResult struct {
	BaseResp
	Data Project `json:"data"`
}

// Cluster cluster info
type Cluster struct {
	AreaID            int64     `json:"area_id"`
	Artifactory       string    `json:"artifactory"`
	CapacityUpdatedAt time.Time `json:"capacity_updated_at"`
	ClusterID         string    `json:"cluster_id"`
	ClusterNum        int64     `json:"cluster_num"`
	ConfigSvrCount    int64     `json:"config_svr_count"`
	CreatedAt         time.Time `json:"created_at"`
	Creator           string    `json:"creator"`
	Description       string    `json:"description"`
	Disabled          bool      `json:"disabled"`
	Environment       string    `json:"environment"`
	ExtraClusterID    string    `json:"extra_cluster_id"`
	IPResourceTotal   int64     `json:"ip_resource_total"`
	IPResourceUsed    int64     `json:"ip_resource_used"`
	MasterCount       int64     `json:"master_count"`
	Name              string    `json:"name"`
	NeedNat           bool      `json:"need_nat"`
	NodeCount         int64     `json:"node_count"`
	ProjectID         string    `json:"project_id"`
	RemainCPU         float64   `json:"remain_cpu"`
	RemainDisk        float64   `json:"remain_disk"`
	RemainMem         float64   `json:"remain_mem"`
	Status            string    `json:"status"`
	TotalCPU          float64   `json:"total_cpu"`
	TotalDisk         float64   `json:"total_disk"`
	TotalMem          float64   `json:"total_mem"`
	Type              string    `json:"type"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// ListProjectClustersResultData ListProjectClusters result data field
type ListProjectClustersResultData struct {
	Count   int64      `json:"count"`
	Results []*Cluster `json:"results"`
}

// ListProjectClustersResult ListProjectClusters result
type ListProjectClustersResult struct {
	BaseResp
	Data ListProjectClustersResultData `json:"data"`
}
