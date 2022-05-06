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
	"github.com/thanos-io/thanos/pkg/tracing/client"

	"github.com/TencentBlueKing/bkmonitor-kits/logger"
)

type contextKey int

const optionKey contextKey = iota

func getOption(ctx context.Context) (*option, bool) {
	v, ok := ctx.Value(optionKey).(*option)
	return v, ok
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
	metrics.MustRegister(
		version.NewCollector("bcs_monitor"),
		prometheus.NewGoCollector(),
		prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}),
	)

	prometheus.DefaultRegisterer = metrics

	// 日志配置
	// loggerOpt := logger.Options{}
	// err := config.UnmarshalKey("logging", &loggerOpt)
	// if err != nil {
	// 	logger.Fatal("logging config not valid")
	// }

	// loggerOpt.Level = "debug"
	// loggerOpt.Stdout = true

	var g run.Group

	cmdOpt := &option{
		g:      &g,
		reg:    metrics,
		tracer: client.NoopTracer(),
		logger: logger.StandardLogger(),
	}

	ctx := context.WithValue(context.Background(), optionKey, cmdOpt)

	// Create a signal channel to dispatch reload events to sub-commands.
	reloadCh := make(chan struct{}, 1)

	// Listen for termination signals.
	{
		cancel := make(chan struct{})
		g.Add(func() error {
			return interrupt(cancel)
		}, func(error) {
			close(cancel)
		})
	}

	// Listen for reload signals.
	{
		cancel := make(chan struct{})
		g.Add(func() error {
			return reload(cancel, reloadCh)
		}, func(error) {
			close(cancel)
		})
	}

	// 主动停止，调用opt.cancel()
	ctx, cancel := context.WithCancel(ctx)
	{
		g.Add(func() error {
			<-ctx.Done()
			return ctx.Err()
		}, func(error) {
			cancel()
		})
		cmdOpt.ctx = ctx
		cmdOpt.cancel = cancel
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

func interrupt(cancel <-chan struct{}) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case s := <-c:
		logger.Infow("caught signal. Exiting.", "signal", s)
		return nil
	case <-cancel:
		return errors.New("canceled")
	}
}

func reload(cancel <-chan struct{}, r chan<- struct{}) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	for {
		select {
		case s := <-c:
			logger.Infow("caught signal. Reloading.", "signal", s)
			select {
			case r <- struct{}{}:
				logger.Info("reload dispatched.")
			default:
			}
		case <-cancel:
			return errors.New("canceled")
		}
	}
}
