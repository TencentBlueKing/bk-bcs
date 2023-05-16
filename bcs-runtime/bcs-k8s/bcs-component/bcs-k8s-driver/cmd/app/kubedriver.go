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

package app

import (
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-driver/kubedriver"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-k8s-driver/kubedriver/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/pkg/util/basic"
	"runtime"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/json"
)

// Run xxx
func Run() error {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Get options from commandline input
	serverOptions := options.NewKubeDriverServerOptions()
	serverOptions.BindFlagSet(pflag.CommandLine)
	basic.HandleVersionFlag(pflag.CommandLine)
	pflag.Parse()

	logConf := conf.LogConfig{
		LogDir:          "logs",
		ToStdErr:        true,
		AlsoToStdErr:    true,
		StdErrThreshold: "0",
	}

	blog.InitLogs(logConf)
	defer blog.CloseLogs()

	by, _ := json.Marshal(serverOptions)
	fmt.Println(string(by))

	if err := kubedriver.StartServer(serverOptions); err != nil {
		blog.Fatalf("Unable to start k8s-driver, error: %s", err)
		return err
	}
	return nil
}
