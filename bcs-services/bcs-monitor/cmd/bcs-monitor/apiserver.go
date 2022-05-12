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
