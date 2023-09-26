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

package gse

import "fmt"

const (
	// DefaultBKSupplierAccount is default bk_supplier_account
	DefaultBKSupplierAccount = "0"
	// DefaultBKCloudID is default bk_cloud_id
	DefaultBKCloudID = 0

	// DefaultBkCloudName is default bk_cloud_name
	DefaultBkCloudName = "default area"

	limit = 200
)

// HostAgentStatus host agent status
type HostAgentStatus struct {
	Host
	Alive int
}

// Host info
type Host struct {
	IP        string `json:"ip"`
	BKCloudID int    `json:"bk_cloud_id"` // 云区域ID
	AgentID   string `json:"agentID,omitempty"`
}

// BKAgent agent info
type BKAgent struct {
	IP           string `json:"ip"`
	BKCloudID    int    `json:"bk_cloud_id"`    // 云区域ID
	BKAgentAlive int    `json:"bk_agent_alive"` // agent在线状态，0为不在线，1为在线
}

// Alive return bk_agent_alive
func (agent *BKAgent) Alive() bool {
	return agent.BKAgentAlive == 1
}

// BKAgentV2 agent info
type BKAgentV2 struct {
	BkAgentID    string `json:"bk_agent_id"` // 格式: '{cloud_id}:{ip}' 或 '{agent_id}'
	BKCloudID    int    `json:"bk_cloud_id"` // 云区域ID
	Version      string `json:"version"`
	RunMode      int    `json:"run_mode"`    // Agent运行模式, 0:proxy 1:agent
	BKAgentAlive int    `json:"status_code"` // agent在线状态 -1:未知 0:初始安装 1:启动中 2:运行中 3:有损状态 4:繁忙状态 5:升级中 6:停止中 7:解除安装
}

// Alive return bk_agent_alive
func (agent *BKAgentV2) Alive() bool {
	return agent.BKAgentAlive == 2
}

// BaseResp base response
type BaseResp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Result    bool   `json:"result"`
	RequestID string `json:"request_id"`
}

// GetAgentStatusReq get agent status req
type GetAgentStatusReq struct {
	BKSupplierAccount string `json:"bk_supplier_account,omitempty"` // 开发商账号
	Hosts             []Host `json:"hosts"`                         // 主机列表
}

// GetAgentStatusReqV2 get agent status req
type GetAgentStatusReqV2 struct {
	AgentIDList []string `json:"agent_id_list"` // 主机列表
}

// GetAgentStatusResp get agent status resp
type GetAgentStatusResp struct {
	BaseResp
	Data map[string]BKAgent `json:"data"` // key 格式: bk_cloud_id:ip
}

// GetAgentStatusRespV2 get agent status resp
type GetAgentStatusRespV2 struct {
	BaseResp
	Data []BKAgentV2 `json:"data"`
}

// BKAgentKey agent key
func BKAgentKey(bkCloudID int, ip string) string {
	return fmt.Sprintf("%d:%s", bkCloudID, ip)
}
