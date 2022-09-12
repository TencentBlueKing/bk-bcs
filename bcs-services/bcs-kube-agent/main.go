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
	"fmt"
	"os"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-kube-agent/app"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	versionFlag        bool
	kubeconfig         string
	periodSync         int
	listenAddr         string
	bkeAddress         string
	clusterId          string
	insecureSkipVerify bool
	// ExternalProxyAddresses 外网跨云部署时，需要上报的api-server的公网或代理地址
	ExternalProxyAddresses string
	// 是否使用 websocket 进行注册
	useWebsocket  bool
	websocketPath string
	reportPath    string
	logConf       conf.LogConfig
)

var rootCmd = &cobra.Command{
	Use:   "bcs-kube-agent",
	Short: "bcs-kube-agent is the binary of bke agent",
	Long:  "bcs-kube-agent is the binary of bke agent to collect cluster info and report to bke",
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			version.ShowVersion()
			os.Exit(0)
		}

		logConf.ToStdErr = true
		blog.InitLogs(logConf)
		defer blog.CloseLogs()

		if err := app.Run(); err != nil {
			blog.Fatal(err)
		}
	},
}

var (
	cfgFile string
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().BoolVar(&versionFlag, "version", false, "display version info")
	rootCmd.PersistentFlags().StringVar(
		&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	rootCmd.PersistentFlags().IntVar(
		&periodSync, "periodsync", 60, "How often to sync message to kube-server, default is 30 seconds")
	rootCmd.PersistentFlags().StringVar(
		&listenAddr, "listen-addr", "0.0.0.0:10254", "The address on which the HTTP server will listen to")
	rootCmd.PersistentFlags().StringVar(
		&bkeAddress, "bke-address", "", "the bke address")
	rootCmd.PersistentFlags().StringVar(
		&reportPath, "report-path", "/bcsapi/v4/clustermanager/v1/clusters/%s/credentials", "the bke report url")
	rootCmd.PersistentFlags().StringVar(
		&clusterId, "cluster-id", "", "cluster which the agent run in")
	rootCmd.PersistentFlags().BoolVar(
		&insecureSkipVerify, "insecureSkipVerify", false, "verifies the server's certificate chain and host name")
	rootCmd.PersistentFlags().StringVar(
		&ExternalProxyAddresses, "external-proxy-addresses",
		"", "external proxy addresses of apiserver, separated by semicolon")
	rootCmd.PersistentFlags().BoolVar(
		&useWebsocket, "use-websocket", false, "whether use websocket to register to bcs-api")
	rootCmd.PersistentFlags().StringVar(
		&websocketPath, "websocket-path", "/bcsapi/v4/clustermanager/v1/websocket/connect",
		"path of the bke address for kubeagent websocket tunnel to register")
	rootCmd.PersistentFlags().Int32Var(&logConf.Verbosity, "verbosity", 3, "verbosity for log")

	// these three flag support direct flag and viper config at the same time, the direct flag could cover the viper config.
	viper.BindPFlag("agent.kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconfig"))
	viper.BindPFlag("agent.periodSync", rootCmd.PersistentFlags().Lookup("periodsync"))
	viper.BindPFlag("agent.listenAddr", rootCmd.PersistentFlags().Lookup("listen-addr"))
	viper.BindPFlag("bke.serverAddress", rootCmd.PersistentFlags().Lookup("bke-address"))
	viper.BindPFlag("cluster.id", rootCmd.PersistentFlags().Lookup("cluster-id"))
	viper.BindPFlag("agent.insecureSkipVerify", rootCmd.PersistentFlags().Lookup("insecureSkipVerify"))
	viper.BindPFlag("agent.external-proxy-addresses", rootCmd.PersistentFlags().Lookup("external-proxy-addresses"))
	viper.BindPFlag("agent.use-websocket", rootCmd.PersistentFlags().Lookup("use-websocket"))
	viper.BindPFlag("bke.websocket-path", rootCmd.PersistentFlags().Lookup("websocket-path"))
	viper.BindPFlag("bke.report-path", rootCmd.PersistentFlags().Lookup("report-path"))
}

func initConfig() {
	viper.SetConfigType("yaml")
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println("Can't read config:", err)
			os.Exit(1)
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
