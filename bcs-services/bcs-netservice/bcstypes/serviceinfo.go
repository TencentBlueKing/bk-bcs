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

package bcstypes

const (
	//BCS_SERV_BASEPATH base path for discovery
	BCS_SERV_BASEPATH = "/bcs/services/endpoints"
	//BCS_MODULE_NETSERVICE module name
	BCS_MODULE_NETSERVICE = "netservice"
)

//ServerInfo base server information
//todo(DeveloperJim): need to move back to bcs-common
type ServerInfo struct {
	IP         string `json:"ip"`
	Port       uint   `json:"port"`
	MetricPort uint   `json:"metric_port"`
	HostName   string `json:"hostname"`
	Scheme     string `json:"scheme"` //http, https
	Version    string `json:"version"`
	Cluster    string `json:"cluster"`
	Pid        int    `json:"pid"`
}

//NetServiceInfo for bcs-netservice
type NetServiceInfo struct {
	ServerInfo
}
