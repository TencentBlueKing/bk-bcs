package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/api"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
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

	cmd.Flags().StringVar(&config.G.API.HTTP.Address, "http-address", config.G.API.HTTP.Address, "API listen http ip")

	// 设置配置命令行优先级高与配置文件
	viper.BindPFlag("query.http.address", cmd.Flag("http-address"))
	return cmd
}

func runAPIServer(opt *option) error {
	var (
		// reg       = opt.reg
		// kitLogger = gokit.NewLogger(opt.logger)
		g         = opt.g
		apiServer *api.APIServer
		err       error
	)

	opt.logger.Info("starting bcs-monitor api node")
	apiServer, err = api.NewAPIServer(opt.ctx)
	if err != nil {
		opt.logger.Errorf("New api error: %s", err)
		return err
	}

	// 启动apiserver, 且支持
	g.Add(func() error {
		return apiServer.Run(":8089")
	}, func(err error) {
		apiServer.Close(opt.ctx)
	})

	return err
}
