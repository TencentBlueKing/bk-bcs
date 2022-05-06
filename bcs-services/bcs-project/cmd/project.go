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

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/logging"
)

var configPath string

var rootCmd = &cobra.Command{
	Use:   "bcs-project",
	Short: "Bcs project service",
	Long:  "Bcs project service manage the project info, and provide crud apis",
	Run:   start,
}

// start 启动服务
func start(cmd *cobra.Command, args []string) {
	logging.Info("bcs project service start...")
	// 加载配置
	config, err := config.LoadConfig(configPath)
	if err != nil {
		panic(fmt.Errorf("load project config failed: %v", err))
	}
	// 初始化logging
	logging.InitLogger(&config.Log)
	logger := logging.GetLogger()
	defer logger.Sync()

	logging.Info("config file path: %s", configPath)

	// 启动服务
	projectSvc := newProjectSvc(config)
	if err := projectSvc.Init(); err != nil {
		logging.Error("init project service failed, err %s", err.Error())
	}
	if err := projectSvc.Run(); err != nil {
		logging.Error("run project service failed, err %s", err.Error())
	}
}

// Execute 执行命令
func Execute() {
	rootCmd.Flags().StringVarP(
		&configPath, "config", "c", "", "path of project service config files",
	)
	if err := rootCmd.Execute(); err != nil {
		logging.Info("start bcs project service error, %v", err)
		os.Exit(1)
	}
}
