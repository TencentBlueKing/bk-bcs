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

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/mitchellh/go-homedir"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/rest/tracing"
)

var (
	// Used for flags.
	cfgFile         string
	appName         = "bcs-platform-manager"
	platformManager = "bcsplatformmanager"

	rootCmd = &cobra.Command{
		Use:   appName,
		Short: "A management platform for bcs",
		Run: func(cmd *cobra.Command, args []string) {
			runCmd(cmd, runAPIServer)
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is etc/bcs-platform-manager.yml)")

	// 自定义 help 函数, 需要主动关闭 runGroup
	defaultHelpFn := rootCmd.HelpFunc()
	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		defaultHelpFn(cmd, args)

		stopCmd(cmd)
	})
}

// Execute :xxx
func Execute() {
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		initConfig()
	}
	var g run.Group
	cmdOpt := &option{
		g: &g,
	}
	ctx := context.WithValue(context.Background(), optionKey, cmdOpt)

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	{
		g.Add(func() error {
			<-ctx.Done()
			return ctx.Err()
		}, func(error) {
			stop()
		})
		cmdOpt.ctx = ctx
		cmdOpt.cancel = stop
	}
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		blog.Info("start bcs platform-manager service error, %v", err)
		stop()
		os.Exit(1) // nolint
	}

	// 初始化 Tracer
	shutdown, errorInitTracing := tracing.InitTracing(config.G.TracingConf)
	if errorInitTracing != nil {
		blog.Info(errorInitTracing.Error())
	}
	if shutdown != nil {
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				blog.Infof("failed to shutdown TracerProvider: %s", err.Error())
			}
		}()
	}

	if err := g.Run(); err != nil && err != ctx.Err() {
		// Use %+v for github.com/pkg/errors error to print with stack.
		blog.Errorw("err", fmt.Sprintf("%+v", errors.Wrap(err, "run command failed")))
		stop()
		os.Exit(1)
	}
}

// stopCmd 停止命令
func stopCmd(cmd *cobra.Command) {
	_, _, opt := cmdOption(cmd)
	opt.cancel()
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

		viper.SetConfigName("bcs-platform-manager")
		viper.SetConfigType("yml")
	}

	if err := viper.ReadInConfig(); err != nil {
		cobra.CheckErr(err)
	}

	if err := config.G.ReadFromViper(viper.GetViper()); err != nil {
		cobra.CheckErr(err)
	}

	// blog初始化
	blog.InitLogs(*config.G.Logging)

	// 日志配置已经Ready, 后面都需要使用日志
	blog.Infof("Using config file:%s", viper.ConfigFileUsed())
}
