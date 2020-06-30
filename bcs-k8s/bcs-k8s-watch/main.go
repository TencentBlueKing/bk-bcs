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
	"flag"
	"runtime"

	"github.com/spf13/pflag"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-watch/app"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-watch/pkg/util/basic"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// initialize logger.
	logConf := conf.LogConfig{
		LogDir:          "/logs",
		ToStdErr:        true,
		AlsoToStdErr:    true,
		StdErrThreshold: "0",
	}
	glog.InitLogs(logConf)
	defer glog.CloseLogs()

	// parse command line flags.
	var configFilePath string
	var pidFilePath string

	pflag.CommandLine.StringVar(&configFilePath, "config", "", "config file for data watch")
	pflag.CommandLine.StringVar(&pidFilePath, "pid", "", "pid file path where the pid is write to")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	basic.HandleVersionFlag(pflag.CommandLine)
	pflag.Parse()

	// pre-run.
	err := app.PrepareRun(configFilePath, pidFilePath)
	if err != nil {
		panic(err.Error())
	}
	glog.Info("bcs-k8s-watch starting...")

	// real-run.
	app.Run(configFilePath)
	glog.Info("bcs-k8s-watch running now.")
}
