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
	"errors"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/util"
	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/automaxprocs/maxprocs"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"

	"bscp.io/cmd/ui/service"
	"bscp.io/pkg/config"
	"bscp.io/pkg/version"
)

var (
	// Used for flags.
	cfgFile       string
	httpAddress   string
	appName       = "bscp-ui"
	podIPsEnv     = "POD_IPs"        // 双栈监听环境变量
	ipv6Interface = "IPV6_INTERFACE" // ipv6本地网关地址
	outConfInfo   bool

	rootCmd = &cobra.Command{
		Use:   appName,
		Short: "bscp ui server",
		Run: func(c *cobra.Command, args []string) {
			// 输出初始化配置
			if outConfInfo {
				encoder := yaml.NewEncoder(os.Stdout)
				encoder.SetIndent(2)
				if err := encoder.Encode(config.G); err != nil {
					klog.ErrorS(err, "output init confinfo failed")
					os.Exit(1)
				}
				os.Exit(0)
			}

			if err := RunCmd(); err != nil && !errors.Is(err, context.Canceled) {
				klog.ErrorS(err, "run cmd failed")
				os.Exit(1)
			}
		},
	}
)

// RunCmd run cli cmd
func RunCmd() error {
	// Running in container with limits but with empty/wrong value of GOMAXPROCS env var could lead to throttling by cpu
	// maxprocs will automate adjustment by using cgroups info about cpu limit if it set as value for runtime.GOMAXPROCS.
	if _, err := maxprocs.Set(maxprocs.Logger(func(template string, args ...interface{}) { klog.Infof(template, args) })); err != nil {
		klog.InfoS("Failed to set GOMAXPROCS automatically", "err", err)
	}

	var g run.Group

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g.Add(func() error {
		<-ctx.Done()
		return ctx.Err()
	}, func(error) {
		stop()
	})

	addrIPv6 := getIPv6AddrFromEnv(httpAddress)
	svr, err := service.NewWebServer(ctx, httpAddress, addrIPv6)
	if err != nil {
		klog.Errorf("init web server err: %s, exited", err)
		os.Exit(1)
	}

	klog.InfoS("listening for requests and metrics", "address", httpAddress)

	g.Add(svr.Run, func(err error) { svr.Close() })

	return g.Run()
}

func init() {
	cobra.OnInitialize(initConfig)

	// 不开启 completion 子命令
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.Flags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.Flags().StringVar(&httpAddress, "http-address", "127.0.0.1:8080", `listen http address`)
	rootCmd.Flags().BoolVarP(&outConfInfo, "confinfo", "o", false, "print init confinfo to stdout")

	// 添加版本
	rootCmd.SetVersionTemplate(`{{println .Version}}`)
	rootCmd.Version = version.FormatVersion("", version.Row)
}

func initConfig() {
	// 过滤不需要配置的子命令
	cmd, _, _ := rootCmd.Find(os.Args[1:])
	if cmd.Name() == "help" || cmd.Name() == "version" || outConfInfo {
		return
	}

	if cfgFile == "" {
		klog.Errorf("config file path is required")
		os.Exit(1)
	}

	viper.SetConfigType("yaml")
	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err != nil {
		cobra.CheckErr(err)
	}

	if err := config.G.ReadFromViper(viper.GetViper()); err != nil {
		cobra.CheckErr(err)
	}

	// 日志配置已经Ready, 后面都需要使用日志
	klog.Infof("Using config file:%s", viper.ConfigFileUsed())
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
	ipv6Interface = os.Getenv(ipv6Interface)
	if ipv6Interface != "" {
		ipv6 = ipv6 + "%" + ipv6Interface
	}

	return net.JoinHostPort(ipv6, listenPort)
}
