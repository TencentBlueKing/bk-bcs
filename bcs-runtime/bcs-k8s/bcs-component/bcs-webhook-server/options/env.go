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

package options

// Env 传递参数给initContainer
type Env struct {
	NodeIp             string `json:"node_ip"`
	PodIp              string `json:"pod_ip"`
	CallType           string `json:"call_type"`
	ExternalSysType    string `json:"external_sys_type"`
	ExternalSysConfig  string `json:"external_sys_config"`
	InitContainerImage string `json:"init_container_image"`
	AppCode            string `json:"app_code"`
	AppSecret          string `json:"app_secret"`
	AppOperator        string `json:"app_operator"`
}
