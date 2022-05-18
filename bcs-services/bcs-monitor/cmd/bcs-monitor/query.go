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

	"github.com/TencentBlueKing/bkmonitor-kits/logger/gokit"
	"github.com/oklog/run"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/query"
)

// QueryCmd
func QueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "PromQL compatible query api",
		Long:  `Query node exposing PromQL enabled Query API with data retrieved from multiple store-gw.`,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		runCmd(cmd, runQuery)
	}

	cmd.Flags().StringVar(&config.G.API.HTTP.Address, "http-address", config.G.API.HTTP.Address, "API listen http ip")
	cmd.Flags().StringVar(&config.G.API.GRPC.Address, "grpc-address", config.G.API.GRPC.Address, "API listen grpc ip")
	cmd.Flags().StringArrayVar(&config.G.API.StoreList, "store", config.G.API.StoreList, "the store list that api connect")

	// 设置配置命令行优先级高与配置文件
	viper.BindPFlag("query.http.address", cmd.Flag("http-address"))
	viper.BindPFlag("query.grpc.address", cmd.Flag("grpc-address"))
	viper.BindPFlag("query.store", cmd.Flag("store"))
	return cmd
}

func runQuery(ctx context.Context, g *run.Group, opt *option) error {
	var (
		reg       = opt.reg
		kitLogger = gokit.NewLogger(opt.logger)
		apiServer *query.API
		err       error
	)

	opt.logger.Info("starting bcs-monitor api node")
	apiServer, err = query.NewAPI(reg, opt.tracer, kitLogger, config.G.API, g)
	if err != nil {
		opt.logger.Errorf("New api error: %s", err)
		return err
	}

	g.Add(apiServer.RunGetStore, apiServer.ShutDownGetStore)
	g.Add(apiServer.RunHttp, apiServer.ShutDownHttp)
	g.Add(apiServer.RunGrpc, apiServer.ShutDownGrpc)

	return err
}
