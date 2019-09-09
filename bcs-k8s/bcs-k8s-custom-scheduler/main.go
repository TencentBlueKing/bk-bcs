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
	"runtime"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/app"
	"bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/options"
	//"bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/ipscheduler"
	"bk-bcs/bcs-k8s/bcs-k8s-custom-scheduler/pkg/ipscheduler"
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

	app.Run(op)

	ipscheduler.DefaultIpScheduler = ipscheduler.NewIpscheduler()
	ipscheduler.DefaultIpScheduler.UpdateNetPoolsPeriodically()

}
