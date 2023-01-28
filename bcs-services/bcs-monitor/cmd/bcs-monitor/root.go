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
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/TencentBlueKing/bkmonitor-kits/logger"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config/watch"
)

var (
	// Used for flags.
	cfgFile          string
	logLevel         string
	certCfgFiles     []string
	httpAddress      string
	advertiseAddress string
	appName          = "bcs-monitor"
	podIPsEnv        = "POD_IPs"        // 双栈监听环境变量
	ipv6Interface    = "IPV6_INTERFACE" // ipv6本地网关地址

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
	}
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	// cobra.OnInitialize(initConfig)

	// 不开启 自动排序
	cobra.EnableCommandSorting = false

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/bcs-monitor.yml)")
	rootCmd.PersistentFlags().StringArrayVar(&certCfgFiles, "credential-config", []string{}, "Credential config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log.level", "", "Log filtering level.")

	// 不开启 completion 子命令
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// rootCmd.SilenceErrors = true
	// rootCmd.SilenceUsage = true

	rootCmd.AddCommand(APIServerCmd())
	rootCmd.AddCommand(QueryCmd())
	rootCmd.AddCommand(StoreGWCmd())
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

	// 命令行日志级别
	if err := config.G.Logging.SetByCmd(logLevel); err != nil {
		cobra.CheckErr(err)
	}

	// 日志配置已经Ready, 后面都需要使用日志
	logger.Infof("Using config file:%s", viper.ConfigFileUsed())

	// watch 凭证文件
	if err := watch.MultiCredWatch(certCfgFiles); err != nil {
		logger.Fatal(err.Error())
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

// getIPv6AddrFromEnv 解析ipv6
func getIPv6AddrFromEnv(ipv4 string) string {
	host, listenPort, _ := net.SplitHostPort(ipv4)
	// ipv4 已经绑定了0.0.0.0 ipv6也不启动
	if ip := net.ParseIP(host); ip != nil && ip.IsUnspecified() {
		return ""
	}

	if listenPort == "" {
		return ""
	}

	podIPs := os.Getenv(podIPsEnv)
	if podIPs == "" {
		return ""
	}

	ipv6 := util.GetIPv6Address(podIPs)
	if ipv6 == "" {
		return ""
	}

	// 在实际中，ipv6不能是回环地址
	if v := net.ParseIP(ipv6); v == nil || v.IsLoopback() {
		return ""
	}

	// local link ipv6 需要带上 interface， 格式如::%eth0
	ipv6Interface := os.Getenv(ipv6Interface)
	if ipv6Interface != "" {
		ipv6 = ipv6 + "%" + ipv6Interface
	}

	return net.JoinHostPort(ipv6, listenPort)
}
