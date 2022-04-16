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

package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/proxy"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var (
	// Used for flags.
	cfgFile     string
	bindAddress string
)

const (
	cmdName = "bcs-unified-apiserver"
)

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		cobra.CheckErr(err)

		cwd, err := os.Getwd()
		cobra.CheckErr(err)

		// Search config in home directory with name (without extension).
		viper.AddConfigPath("/etc")
		viper.AddConfigPath(".")
		viper.AddConfigPath(home)
		viper.AddConfigPath(filepath.Join(cwd, "etc"))

		viper.SetConfigName("prime")
		viper.SetConfigType("yml")
	}

	viper.AutomaticEnv()

	zapProd, _ := zap.NewProduction()
	defer zapProd.Sync() // flushes buffer, if any
	logger := zapProd.Sugar()

	if err := viper.ReadInConfig(); err != nil {
		logger.Errorf("Parse config file error: %v", err)
		os.Exit(1)
	}

	out, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		logger.Errorf("Marshal config file error: %v", err)
		os.Exit(1)
	}

	if err := config.G.ReadFrom(out); err != nil {
		logger.Errorf("ReadFrom config file error: %v", err)
		os.Exit(1)
	}

	logger.Infof("Using config file:%s", viper.ConfigFileUsed())
}

func NewUnifiedAPIServer(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdName,
		Short: "BCS Unified APIServer",
		Long:  `BCS Unified APIServer for isolated, shared and federated cluster`,
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		handler, err := proxy.NewHandler("")
		if err != nil {
			zap.L().Fatal("create proxy handler failed", zap.Error(err))
		}

		httpServer := httpserver.NewHttpServer(
			8088,
			"0.0.0.0",
			"",
		)

		router := httpServer.GetRouter()
		router.Handle("/{uri:.*}", handler)
		if err := httpServer.ListenAndServeMux(false); err != nil {
			fmt.Println(err)
			blog.Errorf("http listen and serve failed, err %s", err.Error())
			os.Exit(1)
		}
		ch := make(chan int)
		<-ch
	}

	flags := cmd.Flags()
	flags.String("bind-address", bindAddress, "The IP address on which to listen for the --secure-port port.")
	flags.String("config", cfgFile, "config file (default is $HOME/config.yml)")
	return cmd
}
