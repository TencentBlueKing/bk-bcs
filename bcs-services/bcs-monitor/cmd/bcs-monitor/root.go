/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云-监控平台 (Blueking - Monitor) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/TencentBlueKing/bkmonitor-kits/logger"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/prometheus/common/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	traclient "github.com/thanos-io/thanos/pkg/tracing/client"
)

var (
	// Used for flags.
	cfgFile  string
	logLevel string

	rootCmd = &cobra.Command{
		Use:   "bcs-monitor",
		Short: "a unified metrics server for bcs-monitor",
		Long:  `A unified metrics server for bcs-monitor`,
	}
)

// Execute 执行根命令, 公共参数在 context 中传递
func Execute(ctx context.Context) error {
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {

		// version 命令不需要初始化配置
		if cmd.Name() == VersionCmd().Name() {
			return nil
		}

		initConfig(cmd)

		return nil
	}
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	// cobra.OnInitialize(initConfig)

	// 不开启 自动排序
	cobra.EnableCommandSorting = false

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/bcs-monitor.yml)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log.level", "", "Log filtering level. (default info)")

	// 不开启 completion 子命令
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// rootCmd.SilenceErrors = true
	// rootCmd.SilenceUsage = true

	rootCmd.AddCommand(APIServerCmd())
	rootCmd.AddCommand(QueryCmd())
	rootCmd.AddCommand(StoreGWCmd())
	rootCmd.AddCommand(VersionCmd())

	// 自定义 help 函数, 需要主动关闭 runGroup
	defaultHelpFn := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		defaultHelpFn(cmd, args)

		cmdOpt, _ := getOption(cmd.Context())
		cmdOpt.cancel()
	})
}

func initConfig(cmd *cobra.Command) {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		cwd, err := os.Getwd()
		cobra.CheckErr(err)

		// Search config in home directory with name (without extension).
		viper.AddConfigPath("/etc")
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.AddConfigPath(filepath.Join(cwd, "etc"))

		viper.SetConfigName("bcs-monitor")
		viper.SetConfigType("yml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logger.Errorf("Parse config file error: %v", err)
		os.Exit(1)
	}

	if err := config.G.ReadFromViper(viper.GetViper()); err != nil {
		logger.Errorf("unmarshal viper config error :%s", err)
		os.Exit(1)
	}

	// init tracer
	ctx := cmd.Context()
	cmdOpt, _ := getOption(ctx)

	cmdOpt.tracer = traclient.NoopTracer()

	logger.Infof("Using config file:%s", viper.ConfigFileUsed())
}

// VersionCmd 展示版本号
func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show application version",
		Long:  `All software has versions. This is bcs-monitor's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version.Print("bcs-monitor"))

			cmdOpt, _ := getOption(cmd.Context())
			cmdOpt.cancel()
		},
	}
	return cmd
}
