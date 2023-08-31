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
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/cmd/webhook/options"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-gitops-manager/pkg/webhook"
)

func main() {
	op := new(options.GitopsWebhookOptions)
	conf.Parse(op)
	blog.InitLogs(op.LogConfig)
	defer blog.CloseLogs()

	ctx, cancel := context.WithCancel(context.Background())
	srv := webhook.NewServer(ctx, op)
	if err := srv.Init(); err != nil {
		blog.Fatal(err)
	}

	go func() {
		interrupt := make(chan os.Signal, 10)
		signal.Notify(interrupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)
		for s := range interrupt {
			blog.Infof("Received signal %v from system. Exit!", s)
			cancel()
			return
		}
	}()

	if err := srv.Run(); err != nil {
		os.Exit(1)
	}
}
