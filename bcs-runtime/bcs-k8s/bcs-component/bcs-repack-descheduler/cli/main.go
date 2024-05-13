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

package main

import (
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/cli/command"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/cli/command/calc"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-repack-descheduler/cli/command/httpserver"
)

func init() {
	blog.InitLogs(conf.LogConfig{
		LogDir:          "./",
		LogMaxSize:      500,
		LogMaxNum:       10,
		ToStdErr:        false,
		AlsoToStdErr:    false,
		Verbosity:       6,
		StdErrThreshold: "2",
	})
}

func main() {
	defer blog.CloseLogs()
	cmd := NewCommand()

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// NewCommand create command
// nolint
func NewCommand() *cobra.Command {
	var root = &cobra.Command{
		Use:   "bcs-descheduler-cli",
		Short: "bcs-descheduler-cli used to check the logic with descheduler",
	}
	root.AddCommand(calc.NewCalcCmd())
	root.AddCommand(httpserver.NewHTTPCmd())
	root.CompletionOptions.DisableDefaultCmd = true
	root.PersistentFlags().StringVarP(&command.ConfigFile, "config", "f", "./config.json",
		`配置文件地址，配置文件中的内容:
  - bkDataUrl 蓝鲸数据模型地址
  - bkDataAppCode 蓝鲸数据模型调用账户
  - bkDataAppSecret 蓝鲸数据模型调用 Secret
  - bkDataToken 蓝鲸数据模型调用 Token
`)
	root.MarkPersistentFlagRequired("config")
	return root
}
