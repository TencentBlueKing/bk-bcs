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
	"context"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-log-manager/app/options"
)

func main() {
	var stopCh = make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	runtime.GOMAXPROCS(runtime.NumCPU())
	op := options.NewLogManagerOption()
	conf.Parse(op)

	blog.InitLogs(op.LogConfig)
	defer blog.CloseLogs()
	blog.Info("init logs success")

	blog.Infof("Config loaded as: %+v", op)
	err := app.Run(ctx, stopCh, op)
	if err != nil {
		blog.Errorf(err.Error())
		terminateProcess(cancel, stopCh)
	}

	notifySignal(cancel, stopCh)
}

func notifySignal(cancel context.CancelFunc, stopCh chan struct{}) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1)
	for {
		select {
		case sig, ok := <-c:
			if !ok {
				blog.Errorf("signal channel has been closed unexpected!")
				terminateProcess(cancel, stopCh)
				os.Exit(255)
			}
			blog.Errorf("Received signal: %+v", sig)
			terminateProcess(cancel, stopCh)
			switch sig {
			case syscall.SIGUSR1:
				os.Exit(255)
			default:
				os.Exit(0)
			}
		}
	}
}

func terminateProcess(cancel context.CancelFunc, stopCh chan struct{}) {
	blog.Error("Wait process stop running...")
	cancel()
	close(stopCh)
	t := time.NewTimer(time.Second * 3)
	<-t.C
	blog.Error("Process stopped.")
}
