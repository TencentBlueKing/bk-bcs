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
	"runtime"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/options"
	"github.com/cch123/gogctuner"
)

const (
	// InCgroup default running in container
	InCgroup = true
	// DefaultGCTarget default GC target set to 70 percent
	DefaultGCTarget = 70.0
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	// tune gc
	go gogctuner.NewTuner(InCgroup, DefaultGCTarget)
}

func main() {
	watchConfig := options.NewWatchOptions()
	conf.Parse(watchConfig)

	// init logger
	glog.InitLogs(watchConfig.LogConfig)
	defer glog.CloseLogs()

	glog.Info("bcs-k8s-watch starting...")
	// real-run.
	app.Run(watchConfig)
	glog.Info("bcs-k8s-watch running now.")
}