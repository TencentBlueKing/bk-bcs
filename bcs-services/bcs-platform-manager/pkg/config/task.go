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

package config

// TaskConf : config for task
type TaskConf struct {
	BcsSubnetResourceCron string `yaml:"bcs_subnet_resource_cron"`
	VpcIPMonitorCron      string `yaml:"vpc_ip_monitor_cron"`
	RemainOverlayIPNum    int    `json:"remain_overlayip_num" yaml:"remain_overlayip_num"`
	RemainUnderlayIPNum   int    `json:"remain_underlayip_num" yaml:"remain_underlayip_num"`
	AllocateSubnetIPCnt   int    `json:"allocate_subnet_ip_cnt" yaml:"allocate_subnet_ip_cnt"`
}

// defaultTaskConf :
func defaultTaskConf() *TaskConf {
	// only for development
	return &TaskConf{
		BcsSubnetResourceCron: "*/60 * * * *",
		VpcIPMonitorCron:      "*/60 * * * *",
		RemainOverlayIPNum:    0,
		RemainUnderlayIPNum:   0,
		AllocateSubnetIPCnt:   0,
	}
}
