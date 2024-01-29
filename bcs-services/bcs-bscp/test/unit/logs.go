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

package unit

import (
	"os"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/logs"
)

// DefaultTestLogDir default test log save dir.
const DefaultTestLogDir = "./log"

// InitTestLogOptions init logs, if log dir not exist, will to create dir.
func InitTestLogOptions(logDir string) {
	// ignore error, if log dir create failed, logs will use tmp dir to save log file.
	os.MkdirAll(logDir, os.ModePerm)

	logs.InitLogger(
		logs.LogConfig{
			LogDir:             logDir,
			LogLineMaxSize:     5,
			LogMaxSize:         500,
			LogMaxNum:          1,
			RestartNoScrolling: true,
			ToStdErr:           false,
			AlsoToStdErr:       false,
			Verbosity:          5,
			StdErrThreshold:    "2",
		},
	)
}
