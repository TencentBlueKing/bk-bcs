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

	"github.com/TencentBlueKing/bkmonitor-kits/logger"
	"github.com/oklog/run"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api"
)

// APIServerCmd
func APIServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "Monitor api server",
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		runCmd(cmd, runAPIServer)
	}

	cmd.Flags().StringVar(&httpAddress, "http-address", "0.0.0.0:8089", "API listen http ip")

	return cmd
}

// runAPIServer apiserver 子服务
func runAPIServer(ctx context.Context, g *run.Group, opt *option) error {
	logger.Infow("listening for requests and metrics", "address", httpAddress)
	server, err := api.NewAPIServer(ctx, httpAddress)
	if err != nil {
		return errors.Wrap(err, "apiserver")
	}

	// 启动 apiserver
	g.Add(func() error {
		return server.Run()
	}, func(err error) {
		server.Close()
	})

	return nil
}
