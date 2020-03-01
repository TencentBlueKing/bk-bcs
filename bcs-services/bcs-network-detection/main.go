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
	"os"
	"runtime"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/conf"
	"bk-bcs/bcs-common/common/license"
	"bk-bcs/bcs-services/bcs-network-detection/app"
	"bk-bcs/bcs-services/bcs-network-detection/app/options"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	op := options.NewOption()
	conf.Parse(op)

	blog.InitLogs(op.LogConfig)
	defer blog.CloseLogs()
	blog.Info("init logs success")
	license.CheckLicense(op.LicenseServerConfig)

	err := app.Run(op)
	if err != nil {
		blog.Errorf(err.Error())
		os.Exit(1)
	}

	ch := make(chan bool)
	<-ch
}
