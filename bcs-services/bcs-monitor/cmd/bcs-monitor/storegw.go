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

	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

// StoreGWCmd StoreGW 命令
func StoreGWCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storegw",
		Short: "Heterogeneous storage gateway",
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		runCmd(cmd, runStoreGW)
	}

	flags := cmd.Flags()
	flags.StringVar(&config.G.StoreGW.HTTP.Address, "http-address", config.G.StoreGW.HTTP.Address, "store gateway listen http ip, default localhost:10210")
	flags.StringVar(&config.G.StoreGW.GRPC.Address, "grpc-address", config.G.StoreGW.GRPC.Address, "store gateway listen grpc ip, default localhost:10211")

	// 设置配置命令行优先级高与配置文件
	viper.BindPFlag("store.http.address", cmd.Flag("http-address"))
	viper.BindPFlag("store.grpc.address", cmd.Flag("grpc-address"))

	return cmd
}

func runStoreGW(ctx context.Context, g *run.Group, opt *option) error {
	var (
		err error
	)

	return err
}
