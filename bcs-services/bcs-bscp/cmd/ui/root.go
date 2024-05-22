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
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/automaxprocs/maxprocs"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"

	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/cmd/ui/service"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/cc"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/config"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/tools"
	"github.com/TencentBlueKing/bk-bcs/bcs-services/bcs-bscp/pkg/version"
)

const (
	appName = "bscp-ui"
)

var (
	// Used for flags.
	cfgFile     string
	bindAddr    string
	port        int
	outConfInfo bool

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
	if _, err := maxprocs.Set(maxprocs.Logger(func(template string, args ...interface{}) {
		klog.Infof(template, args)
	})); err != nil {
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

	envIP, envIPs := tools.GetIPsFromEnv()
	exposeIPs := []string{bindAddr}
	exposeIPs = append(exposeIPs, envIP)
	exposeIPs = append(exposeIPs, envIPs...)
	httpAddresses := make([]string, 0, len(exposeIPs))
	for _, exposeIP := range exposeIPs {
		if exposeIP == "" {
			continue
		}
		httpAddresses = append(httpAddresses, tools.GetListenAddr(exposeIP, port))
	}
	httpAddress := tools.GetListenAddr(bindAddr, port)
	svr, err := service.NewWebServer(ctx, httpAddress, httpAddresses)
	if err != nil {
		klog.Errorf("init web server err: %s, exited", err)
		os.Exit(1) //nolint:gocritic
	}

	klog.InfoS("listening for requests and metrics", "address", bindAddr)

	g.Add(svr.Run, func(err error) {
		_ = svr.Close()
	})

	return g.Run()
}

func init() {
	cc.InitService(cc.UIName)
	cobra.OnInitialize(initConfig)

	// 不开启 completion 子命令
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.Flags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.Flags().StringVar(&bindAddr, "bind-address", "127.0.0.1", "the IP address on which to listen")
	rootCmd.Flags().IntVar(&port, "port", 8080, "http/metrics port")
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
