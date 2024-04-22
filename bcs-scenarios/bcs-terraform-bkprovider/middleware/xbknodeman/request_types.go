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

package xbknodeman

// GetProxyHostRequest get proxy host
type GetProxyHostRequest struct {
	BkCloudId int64 `json:"bk_cloud_id"`
}

// InstallJobRequest install job
type InstallJobRequest struct {
	JobType string         `json:"job_type"`
	Hosts   []*InstallHost `json:"hosts"`
}

// ListCloudRequest list cloud
type ListCloudRequest struct {
	WithDefaultArea *bool `json:"with_default_area,omitempty"`
}

// GetBizProxyHostRequest get biz proxy host
type GetBizProxyHostRequest struct {
}

// CreateCloudRequest create cloud
type CreateCloudRequest struct {
	BkCloudName string `json:"bk_cloud_name"`
	Isp         string `json:"isp"`
	ApID        int64  `json:"ap_id"`
}

// DeleteCloudRequest delete cloud
type DeleteCloudRequest struct {
	BkCloudID int64 `json:"bk_cloud_id"`
}

// UpdateCloudRequest update cloud
type UpdateCloudRequest struct {
	BkCloudID   int64  `json:"bk_cloud_id"`
	BkCloudName string `json:"bk_cloud_name"`
	Isp         string `json:"isp"`
	ApID        int64  `json:"ap_id"`
}

// ListHostRequest list host
type ListHostRequest struct {
	Page       int64       `json:"page,omitempty"`
	PageSize   int64       `json:"page_size,omitempty"`
	Conditions []Condition `json:"conditions,omitempty"`
}

// Condition  搜索条件
type Condition struct {
	// 可选值inner_ip | node_type(AGENT\PROXY\PAGENT) | os_type(
	// LINUX\WINDOWS\AIX\SOLARIS) | status | bk_cloud_id | query(IP、操作系统、Agent状态、Agent版本、云区域模糊搜索)
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// GetJobDetailRequest get job detail
type GetJobDetailRequest struct {
	JobID      int64       `json:"job_id"`
	Page       int64       `json:"page,omitempty"`
	PageSize   int64       `json:"pagesize,omitempty"`
	Conditions []Condition `json:"conditions,omitempty"`
}
