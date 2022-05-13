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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/app"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider"
	_ "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/cloudprovider/manager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcsconf "github.com/Tencent/bk-bcs/bcs-common/common/conf"
	mconfig "github.com/asim/go-micro/v3/config"
	mfile "github.com/asim/go-micro/v3/config/source/file"
	mflag "github.com/asim/go-micro/v3/config/source/flag"
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
	flag.String("tunnel_peertoken",
		"Mx9vWfTZea4MEzc7SlvB8aFl0NhmYQvZzEomOYypDMKkev34Q9kIyh32RjXXCIcn", "peer token for tunnel")
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
	// mongo config
	flag.String("mongo_address", "127.0.0.1:27017", "mongo server address")
	flag.Uint("mongo_connecttimeout", 3, "mongo server connnect timeout")
	flag.String("mongo_database", "", "database in mongo for cluster manager")
	flag.String("mongo_username", "", "mongo username for cluster manager")
	flag.String("mongo_password", "", "mongo passsword for cluster manager")
	flag.Uint("mongo_maxpoolsize", 0, "mongo client connection pool max size, 0 means not set")
	flag.Uint("mongo_minpoolsize", 0, "mongo client connection pool min size, 0 means not set")
	// broker config
	flag.String("broker_address", "127.0.0.1:5672", "broker server for background taskserver")
	flag.String("broker_username", "", "broker username for background taskserver")
	flag.String("broker_password", "", "broker password for background taskserver")
	flag.String("broker_exchange", "clustermanager_task", "broker exchange queue for background taskserver")
	// config file path
	flag.String("conf", "", "config file path")
	flag.Parse()

	// parse cluster manager options
	opt := &options.ClusterManagerOptions{}
	conf, err := mconfig.NewConfig()
	if err != nil {
		blog.Fatalf("create config failed, err %s", err.Error())
	}
	err = conf.Load(
		mflag.NewSource(
			mflag.IncludeUnset(true),
		),
	)
	if err != nil {
		blog.Fatalf("load config flag source failed, err %s", err.Error())
	}
	if len(conf.Get("conf").String("")) > 0 {
		err = conf.Load(mfile.NewSource(mfile.WithPath(conf.Get("conf").String(""))))
		if err != nil {
			blog.Fatalf("load config file source failed, err %s", err.Error())
		}
	}
	err = conf.Scan(opt)
	if err != nil {
		blog.Fatalf("scan config failed, err %s", err.Error())
	}

	blog.InitLogs(bcsconf.LogConfig{
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

	clusterManager := app.NewClusterManager(opt)
	if err := clusterManager.Init(); err != nil {
		blog.Fatalf("init cluster manager failed, err %s", err.Error())
	}
	if err := clusterManager.Run(); err != nil {
		blog.Fatalf("run cluster manager failed, err %s", err.Error())
	}
}
