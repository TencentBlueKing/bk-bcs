/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package version xxx
package version

import (
	"fmt"
	"os"
	"runtime"
)

var (
	// Version 版本号
	Version = ""
	// GitCommit CommitID
	GitCommit = ""
	// BuildTime 二进制构建事件
	BuildTime = ""
	// GoVersion Go 版本号
	GoVersion = runtime.Version()
)

// GetVersion 获取版本信息
func GetVersion() string {
	return fmt.Sprintf(
		"\nVersion  : %s\nGitCommit: %s\nBuildTime: %s\nGoVersion: %s\n",
		Version, GitCommit, BuildTime, GoVersion,
	)
}

// ShowVersionAndExit 打印版本信息并退出
func ShowVersionAndExit() {
	fmt.Printf("%s", GetVersion()) // nolint:forbidigo
	os.Exit(0)
}
