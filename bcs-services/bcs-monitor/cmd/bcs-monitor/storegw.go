package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/config"
)

func StoreGWCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storegw",
		Short: "Heterogeneous storage gateway",
		Long:  `Store node giving access to blocks in a bucket provider. Now supported GCS, S3, Azure, Swift, Tencent COS and Aliyun OSS.`,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		cmdOpt, _ := getOption(cmd.Context())
		if err := runStoreGW(cmdOpt); err != nil {
			cmdOpt.logger.Fatalf("execute %s command failed: %s", cmd.Use, err)
		}

	}

	flags := cmd.Flags()
	flags.StringVar(&config.G.StoreGW.HTTP.Address, "http-address", config.G.StoreGW.HTTP.Address, "store gateway listen http ip, default localhost:10210")
	flags.StringVar(&config.G.StoreGW.GRPC.Address, "grpc-address", config.G.StoreGW.GRPC.Address, "store gateway listen grpc ip, default localhost:10211")

	// 设置配置命令行优先级高与配置文件
	viper.BindPFlag("store.http.address", cmd.Flag("http-address"))
	viper.BindPFlag("store.grpc.address", cmd.Flag("grpc-address"))

	return cmd
}

func runStoreGW(opt *option) error {
	var (
		err error
	)

	return err
}
