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

// Package main xxx
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/version"
)

func main() {
	// 定义命令行参数
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "显示版本信息")
	flag.Parse()

	// 如果指定了 --version 参数，显示版本信息并退出
	if showVersion {
		fmt.Printf("BCS API Gateway Syncing\n")
		fmt.Printf("Version: %s\n", version.BcsVersion)
		fmt.Printf("GitHash: %s\n", version.BcsGitHash)
		fmt.Printf("BuildTime: %s\n", version.BcsBuildTime)
		fmt.Printf("GoVersion: %s\n", version.GoVersion)
		os.Exit(0)
	}

	// 如果没有指定 --version，显示使用说明
	fmt.Println("BCS API Gateway Syncing")
	fmt.Println("使用方法: ./bcs-api-gateway-syncing --version")
	os.Exit(1)
}
