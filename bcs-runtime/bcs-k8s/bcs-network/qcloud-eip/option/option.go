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

package option

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
)

const (
	// DefaultConfigFilePath default config file path for init, release, recover operation
	// except for CmdADD CmdDel
	DefaultConfigFilePath = "/data/bcs/bcs-cni/conf/qcloud-eip.conf"
	// DefaultConfigLogDir default log dir
	DefaultConfigLogDir = "/data/bcs/bcs-cni/logs"
	// WorkModeCNI default work mode, process work as cni plugin
	WorkModeCNI = "cni"
	// WorkModeInit init network interface, including apply eni, write netservice
	WorkModeInit = "init"
	// WorkModeRecover recover network interface and route table
	WorkModeRecover = "recover"
	// WorkModeClean (Deprecated) clean network interface
	WorkModeClean = "clean"
	// WorkModeRelease release eni
	WorkModeRelease = "release"
)

// Option option for qcloud-eip
type Option struct {
	conf.LogConfig
	conf.FileConfig
	WorkMode string `json:"work-mode" value:"cni" usage:"working mode: init, recover, cni, release, clean"`
	EniNum   int    `json:"eni-num" value:"1" usage:"eni num: applied eni num when init cvm"`
	IPNum    int    `json:"ip-num" value:"0" usage:"ip num: applied ip num for each eni, default 0(means getting max ips for eni)"`
}
