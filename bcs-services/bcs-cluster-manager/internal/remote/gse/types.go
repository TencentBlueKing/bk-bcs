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
)

// Host host info
type Host struct {
	IP        string `json:"ip"`
	BKCloudID int    `json:"bk_cloud_id"` // 云区域ID
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

// BaseResp base response
type BaseResp struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Result    bool   `json:"result"`
	RequestID string `json:"request_id"`
}

// GetAgentStatusReq get agent status req
type GetAgentStatusReq struct {
	BKSupplierAccount string `json:"bk_supplier_account"` // 开发商账号
	Hosts             []Host `json:"hosts"`               // 主机列表
}

// GetAgentStatusResp get agent status resp
type GetAgentStatusResp struct {
	BaseResp
	Data map[string]BKAgent `json:"data"` // key 格式: bk_cloud_id:ip
}

// BKAgentKey agent key
func BKAgentKey(bkCloudID int, ip string) string {
	return fmt.Sprintf("%d:%s", bkCloudID, ip)
}
