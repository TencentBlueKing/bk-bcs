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

// Package main cmd start
package main

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/component/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

// APIServerCmd :
func APIServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Monitor api server",
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		runCmd(cmd, runAPIServer)
	}

	return cmd
}

// runAPIServer apiserver 子服务
func runAPIServer(ctx context.Context, g *run.Group, opt *option) error {
	addrIPv6 := utils.GetIPv6AddrFromEnv()
	server, err := api.NewAPIServer(ctx, bindAddress, httpPort, addrIPv6)
	if err != nil {
		return errors.Wrap(err, "apiserver")
	}

	sdName := fmt.Sprintf("%s-%s", appName, "api")
	sd, err := discovery.NewServiceDiscovery(ctx, sdName, version.BcsVersion, bindAddress, httpPort, addrIPv6)
	if err != nil {
		return err
	}

	// init storage
	storage.InitStorage()

	// 启动 apiserver
	g.Add(server.Run, func(err error) { _ = server.Close(); component.GetAuditClient().Close() })
	g.Add(sd.Run, func(error) {})

	bcs.CacheListClusters()

	return nil
}
