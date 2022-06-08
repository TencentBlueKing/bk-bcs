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

	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/TencentBlueKing/bkmonitor-kits/logger"
	"github.com/TencentBlueKing/bkmonitor-kits/logger/gokit"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/query"
)

var storeList []string

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

	cmd.Flags().StringVar(&httpAddress, "http-address", "0.0.0.0:10902", "API listen http ip")
	cmd.Flags().StringVar(&advertiseAddress, "advertise-address", "", "The IP address on which to advertise the server")
	cmd.Flags().StringArrayVar(&storeList, "store", []string{}, "the store list that api connect")

	return cmd
}

func runQuery(ctx context.Context, g *run.Group, opt *option) error {
	kitLogger := gokit.NewLogger(logger.StandardLogger())

	logger.Infow("listening for requests and metrics", "service", "query", "address", httpAddress)
	queryServer, err := query.NewQueryAPI(ctx, opt.reg, opt.tracer, kitLogger, httpAddress, storeList, g)
	if err != nil {
		return errors.Wrap(err, "query")
	}

	sdName := fmt.Sprintf("%s-%s", appName, "query")
	sd, err := discovery.NewServiceDiscovery(ctx, sdName, version.BcsVersion, httpAddress, advertiseAddress)
	if err != nil {
		return err
	}

	g.Add(queryServer.RunGetStore, queryServer.ShutDownGetStore)
	g.Add(queryServer.RunHttp, queryServer.ShutDownHttp)
	g.Add(sd.Run, func(error) {})

	return err
}
