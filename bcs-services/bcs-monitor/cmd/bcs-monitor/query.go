package main

import (
	"github.com/TencentBlueKing/bkmonitor-kits/logger/gokit"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/query"
)

func QueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "PromQL compatible query api",
		Long:  `Query node exposing PromQL enabled Query API with data retrieved from multiple store-gw.`,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmdOpt, _ := getOption(cmd.Context())
		if err := runQuery(cmdOpt); err != nil {
			cmdOpt.logger.Fatalf("execute %s command failed: %s", cmd.Use, err)
		}

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

func runQuery(opt *option) error {
	var (
		reg       = opt.reg
		kitLogger = gokit.NewLogger(opt.logger)
		g         = opt.g
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
