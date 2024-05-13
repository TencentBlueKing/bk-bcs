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

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config/watch"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/tracing"
)

var (
	// Used for flags.
	cfgFile      string
	logLevel     string
	certCfgFiles []string
	bindAddress  string
	httpPort     string
	appName      = "bcs-monitor"

	rootCmd = &cobra.Command{
		Use:   appName,
		Short: "A unified metrics and log server for bcs-monitor",
	}
)

// Execute 执行根命令, 公共参数在 context 中传递
func Execute(ctx context.Context) error {
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// version 命令不需要初始化配置
		if cmd.Name() == VersionCmd().Name() {
			return
		}

		initConfig()
		// 用于tracing init时不同模块的区分
		tracing.SetServiceName(cmd.Use)
	}
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	// cobra.OnInitialize(initConfig)

	// 不开启 自动排序
	cobra.EnableCommandSorting = false

	rootCmd.PersistentFlags().StringVar(&bindAddress, "bind-address", "0.0.0.0", "The IP address on which to listen")
	rootCmd.PersistentFlags().StringVar(&httpPort, "http-port", "10902", "API listen http/metrics port")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/bcs-monitor.yml)")
	rootCmd.PersistentFlags().StringArrayVar(&certCfgFiles, "credential-config", []string{}, "Credential config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log.level", "", "Log filtering level.")
	rootCmd.Flags().SortFlags = false
	rootCmd.PersistentFlags().SortFlags = false

	// 不开启 completion 子命令
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// rootCmd.SilenceErrors = true
	// rootCmd.SilenceUsage = true

	rootCmd.AddCommand(APIServerCmd())
	rootCmd.AddCommand(QueryCmd())
	rootCmd.AddCommand(StoreGWCmd())
	// 打印配置文件
	rootCmd.AddCommand(ConfigViewCmd())
	rootCmd.AddCommand(MigrateCmd())
	rootCmd.AddCommand(VersionCmd())
	rootCmd.Version = printVersion()
	rootCmd.SetVersionTemplate(`{{printf "%s" .Version}}`)

	// 自定义 help 函数, 需要主动关闭 runGroup
	defaultHelpFn := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		defaultHelpFn(cmd, args)

		stopCmd(cmd)
	})
}

func initConfig() {
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

	if err := viper.ReadInConfig(); err != nil {
		cobra.CheckErr(err)
	}

	if err := config.G.ReadFromViper(viper.GetViper()); err != nil {
		cobra.CheckErr(err)
	}

	// blog初始化
	blog.InitLogs(conf.LogConfig{
		LogDir:          config.G.Logging.LogDir,
		LogMaxSize:      config.G.Logging.LogMaxSize,
		LogMaxNum:       config.G.Logging.LogMaxNum,
		ToStdErr:        config.G.Logging.ToStdErr,
		AlsoToStdErr:    config.G.Logging.AlsoToStdErr,
		Verbosity:       config.G.Logging.Verbosity,
		StdErrThreshold: config.G.Logging.StdErrThreshold,
		VModule:         config.G.Logging.VModule,
		TraceLocation:   config.G.Logging.TraceLocation,
	})

	// 日志配置已经Ready, 后面都需要使用日志
	blog.Infof("Using config file:%s", viper.ConfigFileUsed())

	// watch 凭证文件
	if err := watch.MultiCredWatch(certCfgFiles); err != nil {
		blog.Fatal(err.Error())
	}
}

// VersionCmd 展示版本号
func VersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show application version",
		Long:  `All software has versions. This is bcs-monitor's`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(printVersion())
			stopCmd(cmd)
		},
	}
	return cmd
}

func printVersion() string {
	v := appName + ", " + version.GetVersion()
	return v
}

// ConfigViewCmd 打印配置文件
func ConfigViewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show application config",
		Long:  `Print the bcs-monitor configuration file`,
		Run: func(cmd *cobra.Command, args []string) {
			if cfgFile != "" {
				content, err := os.ReadFile(cfgFile)
				if err != nil {
					panic(err)
				}
				fmt.Println(string(content))
			}
			stopCmd(cmd)
		},
	}
	return cmd
}
