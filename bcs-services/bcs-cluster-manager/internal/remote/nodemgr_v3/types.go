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

// Package nodemgr_v3 nodeman v3 接口类型定义
package nodemgr_v3

// ======================== 基础类型 ========================

// BaseResponse v3 基础响应
type BaseResponse struct {
	Code      int        `json:"code"`
	Message   string     `json:"message"`
	RequestID string     `json:"request_id"`
	Result    bool       `json:"result"`
	Error     *ErrorInfo `json:"error,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	System  string        `json:"system"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// V3Page v3 分页请求
type V3Page struct {
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

// NetworkAreaPage 管控区域分页请求
type NetworkAreaPage struct {
	Count bool   `json:"count"`
	Start uint32 `json:"start,omitempty"`
	Limit uint32 `json:"limit,omitempty"`
	Sort  string `json:"sort,omitempty"`
	Order string `json:"order,omitempty"`
}

// ======================== 常量枚举 ========================

// LoginMode 登录方式
type LoginMode string

const (
	// LoginModePassword 密码登录
	LoginModePassword LoginMode = "password"
	// LoginModeKeyFile 密钥文件登录
	LoginModeKeyFile LoginMode = "keyfile"
	// LoginModePasswordVault 密码库登录
	LoginModePasswordVault LoginMode = "password_vault"
)

// Addressing 寻址方式
type Addressing string

const (
	// AddressingStatic 静态寻址
	AddressingStatic Addressing = "static"
	// AddressingDynamic 动态寻址
	AddressingDynamic Addressing = "dynamic"
)

// OSType 操作系统类型
type OSType string

const (
	// LinuxOSType linux
	LinuxOSType OSType = "linux"
	// WindowsOSType windows
	WindowsOSType OSType = "windows"
	// DarwinOSType darwin
	DarwinOSType OSType = "darwin"
)

// InstallMethod 安装方式
type InstallMethod string

const (
	// InstallMethodSSH ssh 安装
	InstallMethodSSH InstallMethod = "ssh"
	// InstallMethodWMI wmi 安装
	InstallMethodWMI InstallMethod = "wmi"
)

// NodeRole 节点角色
type NodeRole string

const (
	// NodeRoleAgent agent
	NodeRoleAgent NodeRole = "agent"
	// NodeRoleProxy proxy
	NodeRoleProxy NodeRole = "proxy"
	// NodeRoleBlank 未分配角色
	NodeRoleBlank NodeRole = "blank"
)

// OperationState 操作状态
type OperationState string

const (
	// OperationStateInit 初始化
	OperationStateInit OperationState = "init"
	// OperationStateLaunched 已启动
	OperationStateLaunched OperationState = "launched"
	// OperationStateRunning 运行中
	OperationStateRunning OperationState = "running"
	// OperationStateSuccess 成功
	OperationStateSuccess OperationState = "success"
	// OperationStateFailed 失败
	OperationStateFailed OperationState = "failed"
	// OperationStateTimeout 超时
	OperationStateTimeout OperationState = "timeout"
	// OperationStateTerminated 已终止
	OperationStateTerminated OperationState = "terminated"
)

// WorkflowStatus 工作流状态
type WorkflowStatus string

const (
	// WorkflowStatusRunning 运行中
	WorkflowStatusRunning WorkflowStatus = "running"
	// WorkflowStatusSuccess 成功
	WorkflowStatusSuccess WorkflowStatus = "success"
	// WorkflowStatusFailed 失败
	WorkflowStatusFailed WorkflowStatus = "failed"
	// WorkflowStatusPartialFailed 部分失败
	WorkflowStatusPartialFailed WorkflowStatus = "partial_failed"
)

// WorkflowType 工作流类型
type WorkflowType string

const (
	// WorkflowTypeInstallAgent 安装 Agent
	WorkflowTypeInstallAgent WorkflowType = "install_agent"
	// WorkflowTypeInstallProxy 安装 Proxy
	WorkflowTypeInstallProxy WorkflowType = "install_proxy"
)

// ======================== Agent 安装 ========================

// AgentInstallRequest Agent 安装请求
type AgentInstallRequest struct {
	Host          []AgentInstallHost `json:"host"`
	TargetVersion []TargetVersion    `json:"target_version,omitempty"`
	IsManual      bool               `json:"is_manual,omitempty"`
}

// AgentInstallHost Agent 安装主机信息
type AgentInstallHost struct {
	BkAddressing             Addressing    `json:"bk_addressing"`
	BkBizID                  int           `json:"bk_biz_id,omitempty"`
	BkHostInnerIP            []string      `json:"bk_host_innerip"`
	BkHostInnerIPv6          []string      `json:"bk_host_innerip_v6,omitempty"`
	LoginIP                  string        `json:"login_ip"`
	LoginPort                int           `json:"login_port,omitempty"`
	LoginUser                string        `json:"login_user"`
	LoginMode                LoginMode     `json:"login_mode"`
	LoginPassword            string        `json:"login_password,omitempty"`
	LoginKeyFile             string        `json:"login_key_file,omitempty"`
	BkNetworkunitID          int           `json:"bk_networkunit_id,omitempty"`
	OsType                   OSType        `json:"os_type"`
	BkHostID                 int           `json:"bk_host_id,omitempty"`
	ReRegister               bool          `json:"re_register,omitempty"`
	InstallPreOrderedPlugins bool          `json:"install_pre_ordered_plugins,omitempty"`
	RenewGseTask             bool          `json:"renew_gse_task,omitempty"`
	RenewGseProc             bool          `json:"renew_gse_proc,omitempty"`
	InstallMethod            InstallMethod `json:"install_method,omitempty"`
}

// TargetVersion 目标版本
type TargetVersion struct {
	Version string `json:"version"`
	CpuArch string `json:"cpu_arch"`
	OsType  OSType `json:"os_type"`
}

// AgentInstallResponse Agent 安装响应
type AgentInstallResponse struct {
	BaseResponse
	Data AgentInstallRespData `json:"data"`
}

// AgentInstallRespData Agent 安装响应数据
type AgentInstallRespData struct {
	WorkflowID string `json:"workflow_id"`
}

// ======================== Proxy 安装 ========================

// ProxyInstallRequest Proxy 安装请求
type ProxyInstallRequest struct {
	Host          []ProxyInstallHost `json:"host"`
	TargetVersion []TargetVersion    `json:"target_version,omitempty"`
	IsManual      bool               `json:"is_manual,omitempty"`
	IsOffline     bool               `json:"is_offline,omitempty"`
}

// ProxyInstallHost Proxy 安装主机信息
type ProxyInstallHost struct {
	BkBizID                  int           `json:"bk_biz_id"`
	BkNetworkunitID          int           `json:"bk_networkunit_id"`
	BkHostID                 int           `json:"bk_host_id,omitempty"`
	BkAddressing             Addressing    `json:"bk_addressing"`
	BkHostInnerIP            []string      `json:"bk_host_innerip,omitempty"`
	BkHostInnerIPv6          []string      `json:"bk_host_innerip_v6,omitempty"`
	OsType                   OSType        `json:"os_type"`
	CpuArch                  string        `json:"cpu_arch,omitempty"`
	LoginIP                  string        `json:"login_ip"`
	LoginPort                int           `json:"login_port"`
	LoginUser                string        `json:"login_user"`
	LoginMode                LoginMode     `json:"login_mode"`
	LoginPassword            string        `json:"login_password,omitempty"`
	LoginKeyFile             string        `json:"login_key_file,omitempty"`
	ExportIP                 string        `json:"export_ip,omitempty"`
	ExportIPv6               string        `json:"export_ip_v6,omitempty"`
	AdvertiseIP              string        `json:"advertise_ip,omitempty"`
	AdvertiseIPv6            string        `json:"advertise_ip_v6,omitempty"`
	ReRegister               bool          `json:"re_register,omitempty"`
	InstallPreOrderedPlugins bool          `json:"install_pre_ordered_plugins,omitempty"`
	RenewGseTask             bool          `json:"renew_gse_task,omitempty"`
	RenewGseProc             bool          `json:"renew_gse_proc,omitempty"`
	InstallMethod            InstallMethod `json:"install_method,omitempty"`
	ProxyTags                []string      `json:"proxy_tags,omitempty"`
	ProxyInstallOriginUnitID int           `json:"proxy_install_origin_unit_id,omitempty"`
	CreditExpiredIntervalSec int           `json:"credit_expired_interval_sec,omitempty"`
	RelayDownloadPort        int           `json:"relay_download_port,omitempty"`
	RelayCallbackPort        int           `json:"relay_callback_port,omitempty"`
}

// ProxyInstallResponse Proxy 安装响应
type ProxyInstallResponse struct {
	BaseResponse
	Data ProxyInstallRespData `json:"data"`
}

// ProxyInstallRespData Proxy 安装响应数据
type ProxyInstallRespData struct {
	WorkflowID string `json:"workflow_id"`
}

// ======================== 工作流列表 ========================

// WorkflowListRequest 工作流列表请求
type WorkflowListRequest struct {
	Page                   *V3Page                `json:"page,omitempty"`
	OnlyCount              bool                   `json:"only_count,omitempty"`
	ExactIncludeConditions *WorkflowListCondition `json:"exact_include_conditions,omitempty"`
	FuzzyIncludeConditions map[string]interface{} `json:"fuzzy_include_conditions,omitempty"`
	OperateTimeRange       *OperateTimeRange      `json:"operate_time_range,omitempty"`
}

// WorkflowListCondition 工作流列表精确匹配条件
type WorkflowListCondition struct {
	WorkflowID      []string         `json:"workflow_id,omitempty"`
	Type            []WorkflowType   `json:"type,omitempty"`
	BkBizID         []int            `json:"bk_biz_id,omitempty"`
	Status          []WorkflowStatus `json:"status,omitempty"`
	Operator        []string         `json:"operator,omitempty"`
	BkHostInnerIP   []string         `json:"bk_host_innerip,omitempty"`
	BkHostInnerIPv6 []string         `json:"bk_host_innerip_v6,omitempty"`
	NodeRole        []NodeRole       `json:"node_role,omitempty"`
}

// OperateTimeRange 操作时间范围
type OperateTimeRange struct {
	StartTimeStampSec int64 `json:"start_timestamp_sec,omitempty"`
	EndTimeStampSec   int64 `json:"end_timestamp_sec,omitempty"`
}

// WorkflowListResponse 工作流列表响应
type WorkflowListResponse struct {
	BaseResponse
	Data WorkflowListData `json:"data"`
}

// WorkflowListData 工作流列表数据
type WorkflowListData struct {
	Total int64          `json:"total"`
	Items []WorkflowItem `json:"items"`
}

// WorkflowItem 工作流条目
type WorkflowItem struct {
	WorkflowID      string   `json:"workflow_id"`
	TriggerID       string   `json:"trigger_id"`
	Type            string   `json:"type"`
	BkBizID         []int    `json:"bk_biz_id"`
	BkNetworkareaID []int    `json:"bk_networkarea_id"`
	BkNetworkunitID []int    `json:"bk_networkunit_id"`
	Operator        string   `json:"operator"`
	OperateTime     int64    `json:"operate_time"`
	FinishTime      int64    `json:"finish_time"`
	Status          string   `json:"status"`
	BkBizName       []string `json:"bk_biz_name"`
	NodeRole        []string `json:"node_role"`
}

// ======================== 工作流操作列表 ========================

// WorkflowOperationListRequest 工作流操作列表请求
type WorkflowOperationListRequest struct {
	OnlyCount              bool                            `json:"only_count,omitempty"`
	Page                   *V3Page                         `json:"page,omitempty"`
	WorkflowID             string                          `json:"workflow_id"`
	ExactIncludeConditions *WorkflowOperationListCondition `json:"exact_include_conditions,omitempty"`
	FuzzyIncludeConditions map[string]interface{}          `json:"fuzzy_include_conditions,omitempty"`
}

// WorkflowOperationListCondition 工作流操作列表精确匹配条件
type WorkflowOperationListCondition struct {
	NodeVersion     []string         `json:"node_version,omitempty"`
	BkHostInnerIP   []string         `json:"bk_host_innerip,omitempty"`
	BkHostInnerIPv6 []string         `json:"bk_host_innerip_v6,omitempty"`
	BkBizID         []int            `json:"bk_biz_id,omitempty"`
	BkNetworkareaID []int            `json:"bk_networkarea_id,omitempty"`
	BkNetworkunitID []int            `json:"bk_networkunit_id,omitempty"`
	State           []OperationState `json:"state,omitempty"`
}

// WorkflowOperationListResponse 工作流操作列表响应
type WorkflowOperationListResponse struct {
	BaseResponse
	Data WorkflowOperationListData `json:"data"`
}

// WorkflowOperationListData 工作流操作列表数据
type WorkflowOperationListData struct {
	Total      int64       `json:"total"`
	Operations []Operation `json:"operations"`
}

// Operation 操作信息
type Operation struct {
	OperationID             string             `json:"operation_id"`
	InstanceIDs             []string           `json:"instance_ids"`
	Operator                string             `json:"operator"`
	CreateTime              int64              `json:"create_time"`
	NodeDeploymentInfo      NodeDeploymentInfo `json:"node_deployment_info"`
	LatestOperInstBriefData OperInstBriefData  `json:"latest_oper_inst_brief_data"`
}

// NodeDeploymentInfo 节点部署信息
type NodeDeploymentInfo struct {
	BkHostID            int      `json:"bk_host_id"`
	BkBizID             int      `json:"bk_biz_id"`
	BkHostInnerIPList   []string `json:"bk_host_innerip_list"`
	BkHostInnerIPv6List []string `json:"bk_host_innerip_v6_list"`
	BkNetworkareaID     int      `json:"bk_networkarea_id"`
	BkNetworkunitID     int      `json:"bk_networkunit_id"`
	NodeVersion         string   `json:"node_version"`
}

// OperInstBriefData 操作实例摘要数据
type OperInstBriefData struct {
	LifeCycle                 LifeCycle           `json:"life_cycle"`
	LatestActionInstBriefData ActionInstBriefData `json:"latest_action_inst_brief_data"`
}

// LifeCycle 生命周期
type LifeCycle struct {
	State      string `json:"state"`
	CreateTime int64  `json:"create_time"`
	StartTime  int64  `json:"start_time"`
	EndTime    int64  `json:"end_time"`
	StopTime   int64  `json:"stop_time"`
}

// ActionInstBriefData 动作实例摘要
type ActionInstBriefData struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// ======================== 工作流操作实例列表 ========================

// OperationInstanceListRequest 操作实例列表请求
type OperationInstanceListRequest struct {
	OnlyCount   bool     `json:"only_count,omitempty"`
	OperationID []string `json:"operation_id,omitempty"`
	OperInstID  []string `json:"oper_inst_id,omitempty"`
}

// OperationInstanceListResponse 操作实例列表响应
type OperationInstanceListResponse struct {
	BaseResponse
	Data OperationInstanceListData `json:"data"`
}

// OperationInstanceListData 操作实例列表数据
type OperationInstanceListData struct {
	Total        int64               `json:"total"`
	OperInstData []OperInstBriefItem `json:"oper_inst_data"`
}

// OperInstBriefItem 操作实例简要数据
type OperInstBriefItem struct {
	OperationID               string              `json:"operation_id"`
	OperInstID                string              `json:"oper_inst_id"`
	OperInstStatus            string              `json:"oper_inst_status"`
	OperationDefName          string              `json:"operation_def_name"`
	ParentOperationID         string              `json:"parent_operation_id"`
	ActionNames               []string            `json:"action_names"`
	LifeCycle                 LifeCycle           `json:"life_cycle"`
	LatestActionInstBriefData ActionInstBriefData `json:"latest_action_inst_brief_data"`
}

// ======================== 主机列表 ========================

// HostListRequest 主机列表请求
type HostListRequest struct {
	Page                   *V3Page                 `json:"page,omitempty"`
	OnlyCount              bool                    `json:"only_count,omitempty"`
	ExactIncludeConditions *HostListExactCondition `json:"exact_include_conditions,omitempty"`
	FuzzyIncludeConditions *HostListFuzzyCondition `json:"fuzzy_include_conditions,omitempty"`
}

// HostListExactCondition 主机列表精确匹配条件
type HostListExactCondition struct {
	BkHostID        []int    `json:"bk_host_id,omitempty"`
	BkBizID         []int    `json:"bk_biz_id,omitempty"`
	BkNetworkareaID []int    `json:"bk_networkarea_id,omitempty"`
	BkHostInnerIP   []string `json:"bk_host_innerip,omitempty"`
	BkHostInnerIPv6 []string `json:"bk_host_innerip_v6,omitempty"`
	BkSetID         []int    `json:"bk_set_id,omitempty"`
	BkModuleID      []int    `json:"bk_module_id,omitempty"`
	OsType          []string `json:"os_type,omitempty"`
	NodeRole        []string `json:"node_role,omitempty"`
	NodeStatus      []string `json:"node_status,omitempty"`
	NodeVersion     []string `json:"node_version,omitempty"`
	BkAgentID       []string `json:"bk_agent_id,omitempty"`
	BkNetworkunitID []int    `json:"bk_networkunit_id,omitempty"`
	NodeGeneration  []int    `json:"node_generation,omitempty"`
	Arch            []string `json:"arch,omitempty"`
	ProxyTags       []string `json:"proxy_tags,omitempty"`
}

// HostListFuzzyCondition 主机列表模糊匹配条件
type HostListFuzzyCondition struct {
	BkHostName      []string `json:"bk_host_name,omitempty"`
	DeptName        []string `json:"dept_name,omitempty"`
	BkHostInnerIP   []string `json:"bk_host_innerip,omitempty"`
	BkHostInnerIPv6 []string `json:"bk_host_innerip_v6,omitempty"`
	BkHostOuterIP   []string `json:"bk_host_outerip,omitempty"`
	BkHostOuterIPv6 []string `json:"bk_host_outerip_v6,omitempty"`
}

// HostListResponse 主机列表响应
type HostListResponse struct {
	BaseResponse
	Data HostListData `json:"data"`
}

// HostListData 主机列表数据
type HostListData struct {
	Total int64      `json:"total"`
	Items []HostItem `json:"items"`
}

// HostItem 主机条目
type HostItem struct {
	TenantID string    `json:"tenant_id"`
	BkHostID int       `json:"bk_host_id"`
	Info     HostInfo  `json:"info"`
	State    HostState `json:"state"`
}

// HostInfo 主机信息
type HostInfo struct {
	BkBizID             int      `json:"bk_biz_id"`
	BkNetworkareaID     int      `json:"bk_networkarea_id"`
	BkNetworkunitID     int      `json:"bk_networkunit_id"`
	BkHostName          string   `json:"bk_host_name"`
	DeptName            string   `json:"dept_name"`
	BkHostInnerIPList   []string `json:"bk_host_innerip_list"`
	BkHostInnerIPv6List []string `json:"bk_host_innerip_v6_list"`
	BkHostOuterIPList   []string `json:"bk_host_outerip_list"`
	BkHostOuterIPv6List []string `json:"bk_host_outerip_v6_list"`
	BkMac               string   `json:"bk_mac"`
	OsType              string   `json:"os_type"`
	CpuArch             string   `json:"cpu_arch"`
	LoginIP             string   `json:"login_ip"`
	LoginPort           int      `json:"login_port"`
	LoginUser           string   `json:"login_user"`
	LoginMode           string   `json:"login_mode"`
	LoginCreditValid    bool     `json:"login_credit_valid"`
	ExportIP            string   `json:"export_ip"`
	ExportIPv6          string   `json:"export_ip_v6"`
	AdvertiseIP         string   `json:"advertise_ip"`
	AdvertiseIPv6       string   `json:"advertise_ip_v6"`
	RelayCallbackPort   int      `json:"relay_callback_port"`
	RelayDownloadPort   int      `json:"relay_download_port"`
	BkAddressing        string   `json:"bk_addressing"`
}

// HostState 主机状态
type HostState struct {
	NodeRole       string   `json:"node_role"`
	NodeStatus     string   `json:"node_status"`
	NodeVersion    string   `json:"node_version"`
	BkAgentID      string   `json:"bk_agent_id"`
	NodeGeneration int      `json:"node_generation"`
	ProxyTags      []string `json:"proxy_tags"`
}

// ======================== 管控区域列表 ========================

// NetworkAreaListRequest 管控区域列表请求
type NetworkAreaListRequest struct {
	Page                   *NetworkAreaPage           `json:"page,omitempty"`
	OnlyCount              bool                       `json:"only_count,omitempty"`
	ExactIncludeConditions *NetworkAreaExactCondition `json:"exact_include_conditions,omitempty"`
	FuzzyIncludeConditions *NetworkAreaFuzzyCondition `json:"fuzzy_include_conditions,omitempty"`
}

// NetworkAreaExactCondition 管控区域精确匹配条件
type NetworkAreaExactCondition struct {
	BkNetworkareaID []int    `json:"bk_networkarea_id,omitempty"`
	CloudVendor     []string `json:"cloud_vendor,omitempty"`
}

// NetworkAreaFuzzyCondition 管控区域模糊匹配条件
type NetworkAreaFuzzyCondition struct {
	BkNetworkareaName []string `json:"bk_networkarea_name,omitempty"`
}

// NetworkAreaListResponse 管控区域列表响应
type NetworkAreaListResponse struct {
	BaseResponse
	Data NetworkAreaListData `json:"data"`
}

// NetworkAreaListData 管控区域列表数据
type NetworkAreaListData struct {
	Total int64             `json:"total"`
	Items []NetworkAreaItem `json:"items"`
}

// NetworkAreaItem 管控区域条目
type NetworkAreaItem struct {
	TenantID          string `json:"tenant_id"`
	BkNetworkareaID   int    `json:"bk_networkarea_id"`
	BkNetworkareaName string `json:"bk_networkarea_name"`
	CloudVendor       string `json:"cloud_vendor"`
}

// ======================== 管控单元列表 ========================

// NetworkUnitPage 管控单元列表分页配置
type NetworkUnitPage struct {
	Offset int32 `json:"offset,omitempty"`
	Limit  int32 `json:"limit,omitempty"`
}

// NetworkUnitExactCondition 管控单元精确匹配条件
type NetworkUnitExactCondition struct {
	BkNetworkunitID []int64 `json:"bk_networkunit_id,omitempty"`
	BkNetworkareaID []int64 `json:"bk_networkarea_id,omitempty"`
	IsDirect        []bool  `json:"is_direct,omitempty"`
	Generation      []int64 `json:"generation,omitempty"`
}

// NetworkUnitListRequest 管控单元列表请求
type NetworkUnitListRequest struct {
	Page                   *NetworkUnitPage           `json:"page,omitempty"`
	OnlyCount              bool                       `json:"only_count,omitempty"`
	ExactIncludeConditions *NetworkUnitExactCondition `json:"exact_include_conditions,omitempty"`
}

// NetworkUnitListResponse 管控单元列表响应
type NetworkUnitListResponse struct {
	BaseResponse
	Data NetworkUnitListData `json:"data"`
}

// NetworkUnitListData 管控单元列表数据
type NetworkUnitListData struct {
	Total int64             `json:"total"`
	Items []NetworkUnitItem `json:"items"`
}

// NetworkUnitLink 管控单元链路配置
type NetworkUnitLink struct {
	BkNetworkareaID int64 `json:"bk_networkarea_id"`
	BkNetworkunitID int64 `json:"bk_networkunit_id"`
	AccesspointID   int64 `json:"accesspoint_id"`
}

// NetworkUnitLinks 管控单元各链路配置
type NetworkUnitLinks struct {
	Cluster NetworkUnitLink `json:"cluster"`
	File    NetworkUnitLink `json:"file"`
	Data    NetworkUnitLink `json:"data"`
}

// NetworkUnitItem 管控单元条目
type NetworkUnitItem struct {
	TenantID           string              `json:"tenant_id"`
	BkNetworkunitID    int64               `json:"bk_networkunit_id"`
	BkNetworkunitName  string              `json:"bk_networkunit_name"`
	BkNetworkareaID    int64               `json:"bk_networkarea_id"`
	AccessPoints       []int64             `json:"accesspoints"`
	Links              NetworkUnitLinks    `json:"links"`
	IsDirect           bool                `json:"is_direct"`
	DirectEndpoints    map[string][]string `json:"direct_endpoints"`
	Generation         int64               `json:"generation"`
	CustomDeployConfig interface{}         `json:"custom_deploy_config"`
}

// ======================== 公钥获取 ========================

// PublicKeyGetResponse 公开密钥获取响应
type PublicKeyGetResponse struct {
	BaseResponse
	Data PublicKeyGetData `json:"data"`
}

// PublicKeyGetData 公开密钥数据
type PublicKeyGetData struct {
	PublicKey string `json:"public_key"`
}
