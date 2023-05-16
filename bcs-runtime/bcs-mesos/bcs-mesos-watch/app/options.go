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

package app

import (
	"github.com/Tencent/bk-bcs/bcs-common/common/static"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-mesos-watch/types"
)

// DefaultConfig return default command line config
func DefaultConfig() *types.CmdConfig {
	return &types.CmdConfig{
		ClusterID:              "",
		ClusterInfo:            "127.0.0.1:2181/blueking",
		CAFile:                 "",
		CertFile:               "",
		KeyFile:                "",
		PassWord:               static.ClientCertPwd,
		RegDiscvSvr:            "",
		Address:                "127.0.0.1",
		ApplicationThreadNum:   100,
		TaskgroupThreadNum:     100,
		ExportserviceThreadNum: 100,
		DeploymentThreadNum:    100,
		ServerPassWord:         static.ServerCertPwd,
		IsExternal:             false,
	}
}
