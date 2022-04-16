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
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/proxy"
)

var (
	// Used for flags.
	cfgFile     string
	bindAddress string
	clusterId   string
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
		if err := Run(bindAddress, clusterId); err != nil {
			fmt.Printf("lei")
			os.Exit(1)
		}
	}

	flags := cmd.Flags()
	flags.StringVar(&bindAddress, "bind-address", "0.0.0.0:8088", "The IP address on which to listen for the --secure-port port.")
	flags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/config.yml)")
	flags.StringVar(&clusterId, "cluster-id", "", "cluster member")
	return cmd
}

func Run(bindAddress, clusterId string) error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	handler, err := proxy.NewHandler(clusterId)
	if err != nil {
		zap.L().Fatal("create proxy handler failed", zap.Error(err))
	}

	ln, err := net.Listen("tcp4", bindAddress)
	if err != nil {
		return err
	}
	defer ln.Close()

	r := mux.NewRouter()

	r.Handle("/{uri:.*}", handler)

	srv := &http.Server{
		Handler: r,
	}
	fmt.Println("lei", bindAddress)
	return srv.Serve(ln)
}
