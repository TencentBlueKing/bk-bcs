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
 */

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	logger "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/klog/v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/app"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-webconsole/console/tracing"
)

func main() {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mgr := app.NewWebConsoleManager(ctx, nil)
	if err := mgr.Init(); err != nil {
		logger.Errorf("init webconsole error: %s", err)
		os.Exit(1)
	}

	//初始化 Tracer
	shutdown, err := tracing.InitTracing(config.G.Tracing)
	if err != nil {
		klog.Info(err.Error())
	}
	if shutdown != nil {
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				klog.Infof("failed to shutdown TracerProvider: %s", err.Error())
			}
		}()
	}

	if err := mgr.Run(); err != nil {
		logger.Errorf("run webconsole error: %s", err)
		os.Exit(1)
	}
}
