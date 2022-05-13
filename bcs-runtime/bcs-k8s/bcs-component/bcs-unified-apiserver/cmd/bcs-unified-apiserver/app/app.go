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

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/cluster"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-unified-apiserver/pkg/config"
)

var (
	// Used for flags.
	cfgFile     string
	bindAddress string
	tlsCertFile string
	tlsKeyFile  string
)

const (
	cmdName = "bcs-unified-apiserver"
)

func initConfig() error {
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

		viper.SetConfigName("bcs-unified-apiserver")
		viper.SetConfigType("yml")
	}

	zapProd, _ := zap.NewProduction()
	defer zapProd.Sync() // flushes buffer, if any
	logger := zapProd.Sugar()

	if err := viper.ReadInConfig(); err != nil {
		logger.Errorf("Parse config file error: %v", err)
		return err
	}

	out, err := yaml.Marshal(viper.AllSettings())
	if err != nil {
		logger.Errorf("Marshal config file error: %v", err)
		return err
	}

	if err := config.G.ReadFrom(out); err != nil {
		logger.Errorf("ReadFrom config file error: %v", err)
		return err
	}

	logger.Infof("Using config file:%s", viper.ConfigFileUsed())
	return nil
}

// NewUnifiedAPIServer APIServer命令行
func NewUnifiedAPIServer(ctx context.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cmdName,
		Short: "BCS Unified APIServer",
		Long:  `BCS Unified APIServer for isolated, shared and federated cluster`,
	}

	cmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err := initConfig(); err != nil {
			return err
		}

		return nil
	}

	cmd.Run = func(cmd *cobra.Command, args []string) {
		if err := Run(bindAddress); err != nil {
			os.Exit(1)
		}
	}

	flags := cmd.Flags()
	flags.StringVar(&bindAddress, "bind-address", "0.0.0.0:8088", "The IP address on which to listen for the --secure-port port.")
	flags.StringVar(&tlsCertFile, "tls-cert-file", "", "TLS Certificate for https server")
	flags.StringVar(&tlsKeyFile, "tls-key-file", "", "TLS Key for the https server")
	flags.StringVar(&cfgFile, "config", "", "config file (dfefault is $HOME/config.yml)")

	return cmd
}

// Run 运行服务
func Run(bindAddress string) error {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	sugar := logger.Sugar()

	clusterHandler, err := cluster.NewHandler()
	if err != nil {
		zap.L().Fatal("create proxy handler failed", zap.Error(err))
	}

	ln, err := net.Listen("tcp4", bindAddress)
	if err != nil {
		return err
	}
	defer ln.Close()

	r := mux.NewRouter()

	r.Handle("/{uri:.*}", clusterHandler)

	srv := &http.Server{
		Handler: r,
	}

	if tlsCertFile != "" && tlsKeyFile != "" {
		sugar.Infof("start serve https://%s", bindAddress)
		return srv.ServeTLS(ln, tlsCertFile, tlsKeyFile)
	}

	sugar.Infof("start serve http://%s", bindAddress)
	return srv.Serve(ln)
}
