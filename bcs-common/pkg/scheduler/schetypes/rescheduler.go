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

package types

//mesos address
type ReschedMesosAddress struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
}

// mesos info
type ReschedMesosInfo struct {
	Address  ReschedMesosAddress `json:"address"`
	Hostname string              `json:"hostname"`
	Id       string              `json:"id"`
	IP       int                 `json:"ip"`
	Pid      string              `json:"pid"`
	Port     int                 `json:"port"`
	Version  string              `json:"version"`
}

// mesos framework info
type Framework struct {
	ID string `json:"id"`
}
