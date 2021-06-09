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
	"runtime"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/app"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/options"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/ipscheduler/v1"
	"github.com/Tencent/bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/ipscheduler/v2"
)

const (
	ipSchedulerV1Type = "IpSchedulerV1"
	ipSchedulerV2Type = "IpSchedulerV2"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	op := options.NewServerOption()
	if err := options.Parse(op); err != nil {
		fmt.Printf("parse options failed: %v\n", err)
		os.Exit(1)
	}

	blog.InitLogs(op.LogConfig)
	defer blog.CloseLogs()

	conf := app.ParseConfig(op)
	app.RunPrometheusMetricsServer(conf)
	app.Run(conf)
	//pid
	if err := common.SavePid(op.ProcessConfig); err != nil {
		blog.Error("fail to save pid: err:%s", err.Error())
	}

	switch conf.CustomSchedulerType {
	case ipSchedulerV1Type:
		v1.DefaultIpScheduler = v1.NewIpScheduler(conf)
		v1.DefaultIpScheduler.UpdateNetPoolsPeriodically()
	case ipSchedulerV2Type:
		defaultIpScheduler, err := v2.NewIpScheduler(conf)
		if err != nil {
			blog.Errorf("failed to build IpSchedulerV2: %s", err.Error())
			os.Exit(1)
		}
		v2.DefaultIpScheduler = defaultIpScheduler
	}

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	blog.Infof("Got OS shutdown signal, shutting down bcs-cc-agent server gracefully...")

	return
}
