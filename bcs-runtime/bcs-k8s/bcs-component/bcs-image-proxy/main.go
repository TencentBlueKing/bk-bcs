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

// Package main xxx
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/options"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/server"
	"github.com/Tencent/bk-bcs/bcs-component/bcs-image-proxy/pkg/utils"
)

func main() {
	op := options.Parse()
	defer blog.CloseLogs()
	blog.Infof("parsed options: %s", string(utils.ToJson(op)))

	quitCh := make(chan struct{})
	svr := server.NewImageProxyServer()
	if err := svr.Init(); err != nil {
		blog.Fatalf("init image proxy server failed, err %s", err.Error())
	}

	go func() {
		defer close(quitCh)
		if err := svr.Run(); err != nil {
			blog.Fatalf("failed to start server: %s", err.Error())
		}
	}()

	interrupt := make(chan os.Signal, 10)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)
	sig := <-interrupt
	blog.Infof("Received signal %v from system. Exit!", sig)
	svr.Shutdown()
	<-quitCh
	blog.Infof("server exited")
}
