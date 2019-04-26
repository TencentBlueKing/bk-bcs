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
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"bk-bcs/bcs-common/common/util"
	"bk-bcs/bcs-services/bcs-loadbalance/app"
	"bk-bcs/bcs-services/bcs-loadbalance/option"

	"bk-bcs/bcs-common/common/blog"

	"github.com/spf13/pflag"
)

const (
	EnvNameLBMetricPort   = "LB_METRICPORT"
	EnvNameLBProxyBinPath = "LB_PROXY_BINPATH"
	EnvNameLBProxyCfgPath = "LB_PROXY_CFGPATH"
)

var (
	zookeeper      string //zookeeper args
	watchpath      string //zookeeper service watch path
	group          string //bcs-loadbalance label for service join in
	proxy          string //proxy model
	bcszkaddr      string //bcs zookeeper address
	clusterid      string //cluster id to register path
	caFile         string //tls ca file path
	clientCertFile string //tls cert file
	clientKeyFile  string //tls key file
)

func init() {
	flags := pflag.CommandLine
	flags.StringVar(&zookeeper, "zk", "127.0.0.1:2381", "zookeeper links for data source")
	flags.StringVar(&watchpath, "zkpath", "", "service info path for watch, [required]")
	flags.StringVar(&group, "group", "external", "bcs loadbalance label for service join in")
	flags.StringVar(&proxy, "proxy", "haproxy", "proxy model, nginx or haproxy")
	flags.StringVar(&bcszkaddr, "bcszkaddr", "127.0.0.1:2181", "bcs zookeeper address")
	flags.StringVar(&clusterid, "clusterid", "", "loadbalance server mesos cluster id")
	flags.StringVar(&caFile, "ca_file", "", "tls ca file path")
	flags.StringVar(&clientCertFile, "client_cert_file", "", "tls cert file path")
	flags.StringVar(&clientKeyFile, "client_key_file", "", "tls key file path")
	util.InitFlags()
}

func getEnv(config *option.LBConfig) error {
	metricPort := os.Getenv(EnvNameLBMetricPort)
	if metricPort == "" {
		config.MetricPort = 59090
	} else {
		port, err := strconv.Atoi(metricPort)
		if err != nil {
			blog.Errorf("metricPort %s tran to in failed:%s", metricPort, err.Error())
			return err
		}
		config.MetricPort = uint(port)
	}
	return nil
}

func main() {
	if watchpath == "" {
		_, err := fmt.Fprintf(os.Stderr, "Starting %s failed: zkpath is required\n", os.Args[0])
		if err != nil {
			blog.Errorf("write failed message to std err failed, err %s", err.Error())
		}
		os.Exit(1)
	}
	//to compatible with 192.168.0.1:2181,192.168.0.2:2181
	//and 192.168.0.1:2181;192.168.0.2:2181
	zookeeper = strings.Replace(zookeeper, ";", ",", -1)
	bcszkaddr = strings.Replace(bcszkaddr, ";", ",", -1)

	config := option.NewDefaultConfig()
	config.Group = group
	config.Zookeeper = zookeeper
	config.WatchPath = watchpath
	config.Proxy = proxy
	config.BcsZkAddr = bcszkaddr
	config.ClusterID = clusterid
	config.CAFile = caFile
	config.ClientCertFile = clientCertFile
	config.ClientKeyFile = clientKeyFile
	config.BinPath = os.Getenv(EnvNameLBProxyBinPath)
	config.CfgPath = os.Getenv(EnvNameLBProxyCfgPath)
	if config.Proxy == option.ProxyHaproxy {
		if len(config.BinPath) == 0 {
			config.BinPath = option.ProxyHaproxyDefaultBinPath
		}
		if len(config.CfgPath) == 0 {
			config.CfgPath = option.ProxyHaproxyDefaultCfgPath
		}
	} else if config.Proxy == option.ProxyNginx {
		if len(config.BinPath) == 0 {
			config.BinPath = option.ProxyNginxDefaultBinPath
		}
		if len(config.CfgPath) == 0 {
			config.CfgPath = option.ProxyNginxDefaultCfgPath
		}
	} else {
		fmt.Printf("bcs-loadbalance unknown proxy %s", config.Proxy)
		os.Exit(1)
	}
	app.InitLogger(config)
	defer app.CloseLogger()

	if err := getEnv(config); err != nil {
		fmt.Printf("bcs-loadbalance starting get env error: %s\n", err.Error())
		blog.Errorf("bcs-loadbalance starting get env error: %s", err.Error())
		os.Exit(1)
	}
	processor := app.NewEventProcessor(config)
	interrupt := make(chan os.Signal, 10)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGUSR1, syscall.SIGUSR2)
	go processor.HandleSignal(interrupt)
	if err := processor.Start(); err != nil {
		processor.Stop()
		fmt.Printf("bcs-loadbalance starting error: %s\n", err.Error())
		blog.Errorf("bcs-loadbalance starting error: %s", err.Error())
		os.Exit(1)
	}
}
