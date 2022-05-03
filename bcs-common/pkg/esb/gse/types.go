/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gse

// BaseReq base request for gse request
type BaseReq struct {
	BkAppCode   string `json:"bk_app_code"`
	BkAppSecret string `json:"bk_app_secret"`
	AccessToken string `json:"access_token,omitempty"`
	BkTicket    string `json:"bk_ticket,omitempty"`
	BkUsername  string `json:"bk_username,omitempty"`
}

// BaseResp base response from esb
type BaseResp struct {
	Code    int64  `json:"code"`
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

// GetAgentStatusReqHost input host info type for get_agent_status
type GetAgentStatusReqHost struct {
	IP        string `json:"ip"`
	BkCloudID int    `json:"bk_cloud_id"`
}

// GetAgentStatusReq request struct of get_agent_status
type GetAgentStatusReq struct {
	BaseReq           `json:",inline"`
	BkSupplierAccount string                  `json:"bk_supplier_account"`
	Hosts             []GetAgentStatusReqHost `json:"hosts"`
}

// AgentStatusData data map for agent status
type AgentStatusData struct {
	IP           string `json:"ip"`
	BkCloudID    int    `json:"bk_cloud_id"`
	BkAgentAlive int    `json:"bk_agent_alive"`
}

// GetAgentStatusResp response struct of get_agent_status
type GetAgentStatusResp struct {
	BaseResp `json:",inline"`
	Data     map[string]AgentStatusData `json:"data,omitempty"`
}
