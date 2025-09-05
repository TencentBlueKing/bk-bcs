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

// Package cmd cmd start
package cmd

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/utils"
)

type contextKey int

const optionKey contextKey = iota

func cmdOption(cmd *cobra.Command) (context.Context, *run.Group, *option) {
	v, ok := cmd.Context().Value(optionKey).(*option)
	if !ok {
		panic("not cmd")
	}
	return cmd.Context(), v.g, v
}

// 命令基础参数
type option struct {
	g      *run.Group
	ctx    context.Context
	cancel func()
}

// CommandFunc 命令行函数
type CommandFunc func(context.Context, *run.Group, *option) error

// runCmd 启动命令
func runCmd(cmd *cobra.Command, cmdFunc CommandFunc) {
	if err := cmdFunc(cmdOption(cmd)); err != nil {
		blog.Fatalw("start server failed", "server", cmd.Name(), "err", err.Error())
	}
}

// runAPIServer 启动api服务
func runAPIServer(ctx context.Context, g *run.Group, opt *option) error {
	addrIPv6 := utils.GetIPv6AddrFromEnv()
	server, err := api.NewAPIServer(ctx, config.G.Base.BindAddress, config.G.Base.HttpPort, addrIPv6)
	if err != nil {
		return errors.Wrap(err, "apiserver")
	}

	sd, err := discovery.NewServiceDiscovery(ctx, platformManager, version.BcsVersion,
		config.G.Base.BindAddress, config.G.Base.HttpPort, addrIPv6)
	if err != nil {
		return err
	}

	// init storage
	storage.InitStorage()

	// 启动 apiserver
	g.Add(server.Run, func(err error) { _ = server.Close(); component.GetAuditClient().Close() })
	g.Add(sd.Run, func(error) {})

	return nil
}
