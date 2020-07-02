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
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-executor/process-executor/driver"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-executor/process-executor/executor"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-executor/process-executor/types"
)

func Run() {
	cxt := context.Background()
	bcsExecutor := executor.NewExecutor(cxt)
	driver, err := driver.NewExecutorDriver(cxt, bcsExecutor)
	if err != nil {
		blog.Errorf("new xecutorDriver error %s, and exit", err.Error())
		os.Exit(1)
	}

	driver.Start()
	interupt := make(chan os.Signal, 10)
	signal.Notify(interupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	go handleSysSignal(interupt, bcsExecutor)
	go handleExecutorFinish(bcsExecutor)
}

func handleSysSignal(signalChan <-chan os.Signal, executor executor.Executor) {
	select {
	case <-signalChan:
		blog.Infof("Get signal from system. Executor was killed, ready to Stop")
		executor.Shutdown()
		return
	}
}

func handleExecutorFinish(bcsExecutor executor.Executor) {
	for {
		time.Sleep(time.Second)

		if bcsExecutor.GetExecutorStatus() == types.ExecutorStatusFinish {
			blog.Infof("executor status %s, and exit", bcsExecutor.GetExecutorStatus())
			os.Exit(1)
		}
	}
}
