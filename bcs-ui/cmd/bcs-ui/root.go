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
	"net"
	"os"
	"path/filepath"

	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/automaxprocs/maxprocs"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-ui/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-ui/pkg/web"
)

var (
	// Used for flags.
	cfgFile       string
	httpAddress   string
	appName       = "bcs-ui"
	podIPsEnv     = "POD_IPs"        // 双栈监听环境变量
	ipv6Interface = "IPV6_INTERFACE" // ipv6本地网关地址

	rootCmd = &cobra.Command{
		Use:   appName,
		Short: "bcs ui server",
	}
)

// Execute 执行
func Execute() {
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		// Running in container with limits but with empty/wrong value of GOMAXPROCS env var could lead to throttling by cpu
		// maxprocs will automate adjustment by using cgroups info about cpu limit if it set as value for runtime.GOMAXPROCS.
		if _, err := maxprocs.Set(maxprocs.Logger(func(template string, args ...interface{}) { klog.Infof(template, args) })); err != nil {
			klog.InfoS("Failed to set GOMAXPROCS automatically", "err", err)
		}

		initConfig()

		addrIPv6 := getIPv6AddrFromEnv(httpAddress)
		svr, err := web.NewWebServer(context.Background(), httpAddress, addrIPv6)
		if err != nil {
			os.Exit(1)
		}

		klog.InfoS("listening for requests and metrics", "address", httpAddress)

		svr.Run()

	}
	rootCmd.Version = printVersion()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// cobra.OnInitialize(initConfig)

	// 不开启 自动排序
	cobra.EnableCommandSorting = false

	// 不开启 completion 子命令
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/bcs-ui.yml)")
	rootCmd.PersistentFlags().StringVar(&httpAddress, "http-address", "127.0.0.1:8080", `listen http address`)

	// rootCmd.SilenceErrors = true
	// rootCmd.SilenceUsage = true
	rootCmd.Version = printVersion()

	// rootCmd.SetVersionTemplate(`{{printf "%s" .Version}}`)
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

		viper.SetConfigName("bcs-ui")
		viper.SetConfigType("yml")
	}

	if err := viper.ReadInConfig(); err != nil {
		cobra.CheckErr(err)
	}

	if err := config.G.ReadFromViper(viper.GetViper()); err != nil {
		cobra.CheckErr(err)
	}

	// 日志配置已经Ready, 后面都需要使用日志
	klog.Infof("Using config file:%s", viper.ConfigFileUsed())
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
