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

package nodeman

// BaseResponse baseResp
type BaseResponse struct {
	Code      int    `json:"code"`
	Result    bool   `json:"result"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// Page page
type Page struct {
	Start int    `json:"start"`
	Limit int    `json:"limit"`
	Sort  string `json:"sort"`
}

// CloudListResponse cloud list response
type CloudListResponse struct {
	BaseResponse
	Data []CloudListData `json:"data"`
}

// CloudListData cloud list data
type CloudListData struct {
	BKCloudID   int    `json:"bk_cloud_id"`
	BKCloudName string `json:"bk_cloud_name"`
	APID        int    `json:"ap_id"`
	IsVisible   bool   `json:"is_visible"`
}

// JobType job type
type JobType string

// OSType os type
type OSType string

// AuthType auth type
type AuthType string

// JobStatus job status
type JobStatus string

// JobAgentStatus job install agent status
type JobAgentStatus string

// AgentStatus agent status
type AgentStatus string

const (
	// DefaultAPID ap id
	DefaultAPID = 1

	// InstallAgentJob install agent job
	InstallAgentJob JobType = "INSTALL_AGENT"
	// ReinstallAgentJob reinstall agent job
	ReinstallAgentJob JobType = "REINSTALL_AGENT"
	// InstallProxyJob install proxy job
	InstallProxyJob JobType = "INSTALL_PROXY"
	// ReinstallProxyJob reinstall proxy job
	ReinstallProxyJob JobType = "REINSTALL_PROXY"

	// LinuxOSType linux os type
	LinuxOSType OSType = "LINUX"

	// PasswordAuthType password auth type
	PasswordAuthType AuthType = "PASSWORD"
	// KeyAuthType key auth type
	KeyAuthType AuthType = "KEY"

	// RootAccount root account
	RootAccount = "root"

	// DefaultPort default port
	DefaultPort = 22
	// SpecialPort special port
	SpecialPort = 36000

	// JobSuccess job success
	JobSuccess JobStatus = "SUCCESS"
	// JobPartFailed job part failed
	JobPartFailed JobStatus = "PART_FAILED"
	// JobFailed job failed
	JobFailed JobStatus = "FAILED"
	// JobRunning job running
	JobRunning JobStatus = "RUNNING"

	// JobAgentPendingStatus job agent pending status
	JobAgentPendingStatus JobAgentStatus = "PENDING"
	// JobAgentRunningStatus job agent running status
	JobAgentRunningStatus JobAgentStatus = "RUNNING"
	// JobAgentSuccessStatus job agent success status
	JobAgentSuccessStatus JobAgentStatus = "SUCCESS"
	// JobAgentFailedStatus job agent failed status
	JobAgentFailedStatus JobAgentStatus = "FAILED"
	// JobAgentIgnoredStatus job agent ignored status
	JobAgentIgnoredStatus JobAgentStatus = "IGNORED"

	// AgentUnknownStatus agent unknown status
	AgentUnknownStatus AgentStatus = "RUNNING"
	// AgentTerminatedStatus agent terminated status
	AgentTerminatedStatus AgentStatus = "TERMINATED"
	// AgentRunning agent running status
	AgentRunning AgentStatus = "RUNNING"
	// AgentNotInstalled agent not installed
	AgentNotInstalled AgentStatus = "NOT_INSTALLED"
)

// JobInstallHost job install host
type JobInstallHost struct {
	BKCloudID          int      `json:"bk_cloud_id"`
	APID               int      `json:"ap_id"`
	BKBizID            int      `json:"bk_biz_id"`
	OSType             OSType   `json:"os_type"`
	InnerIP            string   `json:"inner_ip"`
	OuterIP            string   `json:"outer_ip"`
	LoginIP            string   `json:"login_ip"`
	DataIP             string   `json:"data_ip"`
	Account            string   `json:"account"`
	Port               int      `json:"port"`
	AuthType           AuthType `json:"auth_type"`
	Password           string   `json:"password"`
	Key                string   `json:"key"`
	ForceUpdateAgentId bool     `json:"force_update_agent_id"`
}

// JobInstallRequest job install request
type JobInstallRequest struct {
	JobType       JobType          `json:"job_type"`
	Hosts         []JobInstallHost `json:"hosts"`
	Retention     int              `json:"retention"`
	ReplaceHostID int              `json:"replace_host_id"`
}

// JobInstallResponse job install response
type JobInstallResponse struct {
	BaseResponse
	Data JobInstallData `json:"data"`
}

// JobInstallData job install data
type JobInstallData struct {
	JobID  int    `json:"job_id"`
	JobURL string `json:"job_url"`
}

// JobDetailsRequest job details request
type JobDetailsRequest struct {
	JobID    int `json:"job_id"`
	Page     int `json:"page"`
	PageSize int `json:"pagesize"`
}

// JobDetailsResponse job details details
type JobDetailsResponse struct {
	BaseResponse
	Data JobDetailsData `json:"data"`
}

// JobDetailsData job details data
type JobDetailsData struct {
	JobID          int                  `json:"job_id"`
	JobType        JobType              `json:"job_type"`
	JobTypeDisplay string               `json:"job_type_display"`
	Total          int                  `json:"total"`
	List           []JobHostDetail      `json:"list"`
	Status         JobStatus            `json:"status"`
	Statistics     JobDetailsStatistics `json:"statistics"`
}

// JobDetailsStatistics job statistics
type JobDetailsStatistics struct {
	TotalCount   int `json:"total_count"`
	FailedCount  int `json:"failed_count"`
	IgnoredCount int `json:"ignored_count"`
	PendingCount int `json:"pending_count"`
	RunningCount int `json:"running_count"`
	SuccessCount int `json:"success_count"`
}

// JobHostDetail job host detail
type JobHostDetail struct {
	InstanceID string         `json:"instance_id"`
	IP         string         `json:"ip"`
	InnerIP    string         `json:"inner_ip"`
	InnerIPv6  string         `json:"inner_ipv6"`
	BKHostID   int            `json:"bk_host_id"`
	BKCloudID  int            `json:"bk_cloud_id"`
	BKBizID    int            `json:"bk_biz_id"`
	Status     JobAgentStatus `json:"status"`
}

// ListHostsRequest list hosts request
type ListHostsRequest struct {
	BKBizIDs []int `json:"bk_biz_id"`
	Page     int   `json:"page"`
	PageSize int   `json:"pagesize"`
}

// ListHostsResponse list hosts response
type ListHostsResponse struct {
	BaseResponse
	Data ListHostsData `json:"data"`
}

// ListHostsData list hosts data
type ListHostsData struct {
	Total int        `json:"total"`
	List  []HostInfo `json:"list"`
}

// HostInfo host info
type HostInfo struct {
	Status     AgentStatus `json:"status"`
	Version    string      `json:"version"`
	BKCloudID  int         `json:"bk_cloud_id"`
	BKBizID    int         `json:"bk_biz_id"`
	BKHostID   int         `json:"bk_host_id"`
	BKHostName string      `json:"bk_host_name"`
	OSType     OSType      `json:"os_type"`
	InnerIP    string      `json:"inner_ip"`
	InnerIPv6  string      `json:"inner_ipv6"`
	OuterIP    string      `json:"outer_ip"`
	OuterIPv6  string      `json:"outer_ipv6"`
	Topology   []string    `json:"topology"`
}
