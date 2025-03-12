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
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-federation-manager/cmd"
)

func main() {
	opts := cmd.NewFederationManagerOptions()
	if err := cmd.Parse(opts); err != nil {
		fmt.Fprintf(os.Stderr, "set config file failed, err %s\n", err.Error())
		os.Exit(1)
	}

	blog.InitLogs(opts.LogConfig)
	defer blog.CloseLogs()

	server := cmd.NewServer(opts)
	if err := server.Init(); err != nil {
		blog.Fatalf("init federation manager failed, err %s", err.Error())
	}
	if err := server.Run(); err != nil {
		blog.Fatalf("run federation manager failed, %s", err.Error())
	}
}
