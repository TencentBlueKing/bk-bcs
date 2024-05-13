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
	"os"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/oklog/run"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/common/version"
	"github.com/spf13/cobra"
	"github.com/thanos-io/thanos/pkg/tracing/client"
	"go.uber.org/automaxprocs/maxprocs"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/tracing"
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
		blog.Fatalw("start server failed", "server", cmd.Name(), "err", err.Error())
	}
}

// stopCmd 停止命令
func stopCmd(cmd *cobra.Command) {
	_, _, opt := cmdOption(cmd)
	opt.cancel()
}

// 命令基础参数
type option struct {
	g      *run.Group
	reg    *prometheus.Registry
	tracer opentracing.Tracer
	logger blog.GlogKit // nolint
	ctx    context.Context
	cancel func()
}

// isPrintVersion 是否是--version, -v命令
func isPrintVersion() bool {
	if len(os.Args) < 2 {
		return false
	}
	arg := os.Args[1]
	for _, name := range []string{"version", "v"} {
		if arg == "-"+name || arg == "--"+name {
			return true
		}
	}
	return false
}

func main() {
	// metrics 配置
	metrics := prometheus.NewRegistry()
	metrics.MustRegister(
		version.NewCollector("bcs_monitor"),
		collectors.NewGoCollector(
			collectors.WithGoCollectorRuntimeMetrics(collectors.GoRuntimeMetricsRule{Matcher: regexp.MustCompile("/.*")}),
		),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

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
		stop()
		os.Exit(1) // nolint
	}

	if isPrintVersion() {
		stop()
		os.Exit(0) // nolint
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

	// Running in container with limits but with empty/wrong value of GOMAXPROCS env var could lead to throttling by cpu
	// maxprocs will automate adjustment by using cgroups info about cpu limit if it set as value for runtime.GOMAXPROCS.
	if _, err := maxprocs.Set(maxprocs.Logger(func(template string, args ...interface{}) {
		blog.Infof(template, args)
	})); err != nil {
		blog.Warnw("Failed to set GOMAXPROCS automatically", "err", err)
	}
	if err := g.Run(); err != nil && err != ctx.Err() {
		// Use %+v for github.com/pkg/errors error to print with stack.
		blog.Errorw("err", fmt.Sprintf("%+v", errors.Wrap(err, "run command failed")))
		stop()
		os.Exit(1)
	}

	if ctx.Err() == nil {
		blog.Info("exiting")
	}

}
