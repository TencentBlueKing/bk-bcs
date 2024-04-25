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

// GetProxyHostResponse response
type GetProxyHostResponse struct {
	*BaseResponse
	Data []*ProxyHost `json:"data"`
}

// InstallJobResponse response
type InstallJobResponse struct {
	*BaseResponse
	Data *Job `json:"data"`
}

// ListCloudResponse response
type ListCloudResponse struct {
	*BaseResponse
	Data []*Cloud `json:"data"`
}

// GetBizProxyHostResponse response
type GetBizProxyHostResponse struct {
	*BaseResponse
	Data []*BizProxyHost `json:"data"`
}

// CreateCloudResponse return bk_cloud_id
type CreateCloudResponse struct {
	*BaseResponse
	Data CloudID `json:"data"`
}

// ListHostResponse get host
type ListHostResponse struct {
	*BaseResponse
	Data *ListHostData `json:"data"`
}

// ListHostData  host data
type ListHostData struct {
	Total int        `json:"total"` // 主机总数
	List  []HostInfo `json:"list"`  // 汇总后的主机信息
}

// GetJobDetailResponse get job detail
type GetJobDetailResponse struct {
	*BaseResponse
	Data *GetJobDetailData
}

// GetJobDetailData get job detail
type GetJobDetailData struct {
	JobID          int32          `json:"job_id"`
	CreatedBy      string         `json:"created_by"`
	JobType        string         `json:"job_type"`
	JobTypeDisplay string         `json:"job_type_display"`
	IPFilterList   []string       `json:"ip_filter_list"`
	Total          *int32         `json:"total,omitempty"`
	List           []*JobHostList `json:"list,omitempty"`
	Statistics     *JobStatistics `json:"statistics"`
	Status         string         `json:"status"`
	EndTime        string         `json:"end_time"`
	StartTime      string         `json:"start_time"`
	CostTime       string         `json:"cost_time"`
	Meta           *JobMeta       `json:"meta"`
}

// JobHostList get job host list
type JobHostList struct {
	FilterHost    bool   `json:"filter_host,omitempty"`
	BkHostID      int32  `json:"bk_host_id,omitempty"`
	IP            string `json:"ip,omitempty"`
	InnerIP       string `json:"inner_ip,omitempty"`
	InnerIPv6     string `json:"inner_ipv6,omitempty"`
	BkCloudID     int32  `json:"bk_cloud_id,omitempty"`
	BkCloudName   string `json:"bk_cloud_name,omitempty"`
	BkBizID       int32  `json:"bk_biz_id,omitempty"`
	BkBizName     string `json:"bk_biz_name,omitempty"`
	JobID         int32  `json:"job_id,omitempty"`
	Status        string `json:"status,omitempty"`
	StatusDisplay string `json:"status_display,omitempty"`
}

// JobStatistics job statistics
type JobStatistics struct {
	TotalCount   int32 `json:"total_count"`
	FailedCount  int32 `json:"failed_count"`
	IgnoredCount int32 `json:"ignored_count"`
	PendingCount int32 `json:"pending_count"`
	RunningCount int32 `json:"running_count"`
	SuccessCount int32 `json:"success_count"`
}

// JobMeta meta
type JobMeta struct {
	Type            string `json:"type"`
	StepType        string `json:"step_type"`
	OpType          string `json:"op_type"`
	OpTypeDisplay   string `json:"op_type_display"`
	StepTypeDisplay string `json:"step_type_display"`
	Name            string `json:"name,omitempty"`
	Category        string `json:"category,omitempty"`
	PluginName      string `json:"plugin_name,omitempty"`
}
