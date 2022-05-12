/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云-监控平台 (Blueking - Monitor) available.
 * Copyright (C) 2017-2021 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package main

import (
	"github.com/spf13/cobra"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api"
)

func APIServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "apiserver",
		Short: "BCS Monitor api server",
		Long:  `BCS Monitor api server.`,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmdOpt, _ := getOption(cmd.Context())
		if err := runAPIServer(cmdOpt); err != nil {
			cmdOpt.logger.Fatalf("execute %s command failed: %s", cmd.Use, err)
		}
	}

	cmd.Flags().StringVar(&httpAddress, "http-address", "0.0.0.0:8089", "API listen http ip")

	return cmd
}

func runAPIServer(opt *option) error {
	var (
		g         = opt.g
		apiServer *api.APIServer
		err       error
	)

	opt.logger.Infow("listening for requests and metrics", "address", httpAddress)
	apiServer, err = api.NewAPIServer(opt.ctx)
	if err != nil {
		opt.logger.Errorf("New api error: %s", err)
		return err
	}

	// 启动apiserver, 且支持
	g.Add(func() error {
		return apiServer.Run(httpAddress)
	}, func(err error) {
		apiServer.Close(opt.ctx)
	})

	return err
}
