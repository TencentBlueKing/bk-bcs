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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-proxy/server"

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
	// server config
	flag.String("address", "127.0.0.1", "grpc server address")
	flag.String("insecureaddress", "127.0.0.1", "insecure server address")
	flag.Uint("port", 8081, "grpc server port")
	flag.Uint("httpport", 8080, "http server port")
	flag.Uint("httpinsecureport", 8079, "http insecure port if tls is enabled")
	flag.Uint("metricport", 8082, "metric port")
	flag.String("serverca", "", "tls ca file for server")
	flag.String("servercert", "", "tls cert file for server")
	flag.String("serverkey", "", "tls key file for server")
	// client config
	flag.String("clientca", "", "tls ca file for cluster manager as client")
	flag.String("clientcert", "", "tls cert file for cluster manager as client")
	flag.String("clientkey", "", "tls key file for cluster manager as client")
	// swagger config
	flag.String("swagger_dir", "", "swagger files for api docs")
	// tunnel config
	flag.String("tunnel_peertoken", "", "peer token for tunnel")
	flag.String("tunnel_managedclusterid", "", "managed cluster id for tunnel")
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
	// config file path
	flag.String("conf", "", "config file path")
	flag.Parse()

	opt := &options.ProxyOptions{}
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

	argocdProxy := server.NewArgocdProxy(opt)
	if err := argocdProxy.Init(); err != nil {
		blog.Fatalf("init bcs argocd proxy failed, %s", err.Error())
	}

	if err := argocdProxy.Run(); err != nil {
		blog.Fatalf("run bcs argocd proxy failed, %s", err.Error())
	}
}
