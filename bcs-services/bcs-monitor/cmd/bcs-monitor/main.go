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
	"os"
	"os/signal"
	"syscall"

	"github.com/oklog/run"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	"github.com/spf13/cobra"
	"github.com/thanos-io/thanos/pkg/tracing/client"

	"github.com/TencentBlueKing/bkmonitor-kits/logger"
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

// CmdFunc 命令行函数
type CmdFunc func(context.Context, *run.Group, *option) error

// runCmd 启动命令
func runCmd(cmd *cobra.Command, cmdFunc CmdFunc) {
	if err := cmdFunc(cmdOption(cmd)); err != nil {
		logger.Fatalw("start server failed", "server", cmd.Name(), "err", err.Error())
	}
}

// 停止命令
func stopCmd(cmd *cobra.Command) {
	_, _, opt := cmdOption(cmd)
	opt.cancel()
}

// 命令基础参数
type option struct {
	g      *run.Group
	reg    *prometheus.Registry
	tracer opentracing.Tracer
	logger logger.Logger
	ctx    context.Context
	cancel func()
}

func main() {
	// metrics 配置
	metrics := prometheus.NewRegistry()
	metrics.MustRegister(version.NewCollector("bcs_monitor"))

	prometheus.DefaultRegisterer = metrics

	var g run.Group

	cmdOpt := &option{
		g:      &g,
		reg:    metrics,
		tracer: client.NoopTracer(),
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

	if err := Execute(ctx); err != nil {
		os.Exit(1)
	}

	if err := g.Run(); err != nil && err != ctx.Err() {
		// Use %+v for github.com/pkg/errors error to print with stack.
		logger.Errorw("err", fmt.Sprintf("%+v", errors.Wrap(err, "run command failed")))
		os.Exit(1)
	}

	if ctx.Err() == nil {
		logger.Info("exiting")
	}

}
