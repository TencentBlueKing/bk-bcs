/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
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
	"os"
	"os/signal"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/options"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/server"
)

func main() {
	opt := options.New()
	options.Parse(opt)
	blog.V(3).Infof("options: %+v", opt)
	blog.InitLogs(opt.LogConfig)

	svr := server.New(opt)
	if err := svr.Init(); err != nil {
		blog.Fatal(err.Error())
	}

	stopChan := make(chan struct{})
	go func() {
		interrupt := make(chan os.Signal, 10)
		signal.Notify(interrupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM,
			syscall.SIGUSR1, syscall.SIGUSR2)
		for s := range interrupt {
			blog.Infof("Received signal %v from system. Exit!", s)

			// Stop data informer
			svr.Stop()

			// Close networkPolicy controller
			close(stopChan)
			return
		}
	}()

	if err := svr.Run(stopChan); err != nil {
		os.Exit(1)
	}
}
