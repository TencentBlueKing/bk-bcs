/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"flag"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-server/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-server/server"

	microCfg "go-micro.dev/v4/config"
	microFile "go-micro.dev/v4/config/source/file"
	microFlg "go-micro.dev/v4/config/source/flag"
)

func main() {
	// etcd option
	flag.String("etcd_endpoints", "", "endpoints of etcd")
	flag.String("etcd_cert", "", "cert file of etcd")
	flag.String("etcd_key", "", "key file for etcd")
	flag.String("etcd_ca", "", "ca file for etcd")

	// log config
	flag.String("bcslog_dir", "./logs", "If non-empty, write log files in this directory")
	flag.Uint64("bcslog_maxsize", 500, "Max size (MB) per log file.")
	flag.Int("bcslog_maxnum", 10, "Max num of log file. The oldest will be removed if there is a extra file created.")
	flag.Bool("bcslog_tostderr", false, "log to standard error instead of files")
	flag.Bool("bcslog_alsotostderr", false, "log to standard error as well as files")
	flag.Int("bcslog_v", 0, "log level for V logs")
	flag.String("bcslog_stderrthreshold", "2", "logs at or above this threshold go to stderr")
	flag.String("bcslog_vmodule", "", "comma-separated list of pattern=N settings for file-filtered logging")
	flag.String("bcslog_backtraceat", "", "when logging hits line file:N, emit a stack trace")

	// swagger config
	flag.String("swagger_dir", "", "swagger files for api docs")

	// server config
	flag.String("address", "127.0.0.1", "grpc server address")
	flag.String("insecureaddress", "127.0.0.1", "insecure server address")
	flag.Uint("port", 8081, "grpc server port")
	flag.Uint("httpport", 8080, "http server port")
	flag.Uint("metricport", 8082, "metric port")
	flag.String("serverca", "", "tls ca file for server")
	flag.String("servercert", "", "tls cert file for server")
	flag.String("serverkey", "", "tls key file for server")
	flag.String("clientca", "", "tls ca file for client")
	flag.String("clientcert", "", "tls cert file for client")
	flag.String("clientkey", "", "tls key file for client")

	// kubeconfig path
	flag.String("masterurl", "", "url of k8s master")
	flag.String("kubeconfig", "", "kubeconfig path")

	// tunnel option
	flag.String("tunnel_agentid", "fake-cluster", "id for this proxy agent")
	flag.String("tunnel_proxyaddress", "r", "target proxy address")

	// config file path
	flag.String("conf", "", "config file path")
	flag.Parse()

	opt := &options.ArgocdServerOptions{}
	config, err := microCfg.NewConfig()
	if err != nil {
		blog.Fatalf("create config failed, %s", err.Error())
	}

	if err = config.Load(
		microFlg.NewSource(
			microFlg.IncludeUnset(true),
		),
	); err != nil {
		blog.Fatalf("load config from flag failed, %s", err.Error())
	}

	if len(config.Get("conf").String("")) > 0 {
		err = config.Load(microFile.NewSource(microFile.WithPath(config.Get("conf").String(""))))
		if err != nil {
			blog.Fatalf("load config from file failed, err %s", err.Error())
		}
	}

	if err = config.Scan(opt); err != nil {
		blog.Fatalf("scan config failed, %s", err.Error())
	}

	blog.InitLogs(conf.LogConfig{
		LogDir:          opt.BcsLog.LogDir,
		LogMaxSize:      opt.BcsLog.LogMaxSize,
		LogMaxNum:       opt.BcsLog.LogMaxNum,
		ToStdErr:        opt.BcsLog.ToStdErr,
		AlsoToStdErr:    opt.BcsLog.AlsoToStdErr,
		Verbosity:       opt.BcsLog.Verbosity,
		StdErrThreshold: opt.BcsLog.StdErrThreshold,
		VModule:         opt.BcsLog.VModule,
		TraceLocation:   opt.BcsLog.TraceLocation,
	})

	argocdServer := server.NewArgocdServer(opt)
	if err := argocdServer.Init(); err != nil {
		blog.Fatalf("init bcs argocd server failed, %s", err.Error())
	}

	if err := argocdServer.Run(); err != nil {
		blog.Fatalf("run bcs argocd server failed, %s", err.Error())
	}
}
