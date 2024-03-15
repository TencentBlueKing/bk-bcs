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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/query"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

var (
	strictStoreList []string
	storeList       []string
	httpSDURLs      []string
)

// QueryCmd xxx
func QueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "PromQL compatible query api",
		Long:  `Query node exposing PromQL enabled Query API with data retrieved from multiple store-gw.`,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		runCmd(cmd, runQuery)
	}

	cmd.Flags().StringArrayVar(&storeList, "store", []string{},
		"Addresses of statically configured store endpoints")
	cmd.Flags().StringArrayVar(&strictStoreList, "store-strict", []string{},
		"Addresses of statically configured store endpoints always used, even if the health check fails")
	cmd.Flags().StringArrayVar(&httpSDURLs, "store.http-sd-url", []string{},
		"HTTP-based service discovery provides store endpoints")

	return cmd
}

func runQuery(ctx context.Context, g *run.Group, opt *option) error {
	logKit := blog.LogKit

	blog.Infow("listening for requests and metrics", "service", "query", "address", bindAddress)
	addrIPv6 := utils.GetIPv6AddrFromEnv()
	queryServer, err := query.NewQueryAPI(ctx, opt.reg, opt.tracer, logKit, bindAddress, httpPort, addrIPv6,
		strictStoreList, storeList, httpSDURLs, g)
	if err != nil {
		return errors.Wrap(err, "query")
	}

	sdName := fmt.Sprintf("%s-%s", appName, "query")
	sd, err := discovery.NewServiceDiscovery(ctx, sdName, version.BcsVersion, bindAddress, httpPort, addrIPv6)
	if err != nil {
		return err
	}

	g.Add(queryServer.Run, queryServer.Close)
	g.Add(sd.Run, func(error) {})

	return err
}
