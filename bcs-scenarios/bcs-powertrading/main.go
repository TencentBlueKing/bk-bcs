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
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"

	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-powertrading/cmd"
)

func main() {
	opts := cmd.DefaultOptions()

	conf.Parse(opts)
	blog.InitLogs(opts.LogConfig)
	defer blog.CloseLogs()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		interrupt := make(chan os.Signal, 10)
		signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)
		for s := range interrupt {
			blog.Infof("Received signal %v from system. Exit!", s)
			cancel()
			return
		}
	}()
	// config option verification
	if err := opts.Complete(); err != nil {
		fmt.Fprintf(os.Stderr, "server option complete failed, %s\n", err.Error())
		return
	}
	if err := opts.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "server option validate failed, %s\n", err.Error())
		return
	}
	blog.Infof("opts:%v", opts)
	s := cmd.NewServer(opts)
	if err := s.Init(ctx); err != nil {
		blog.Fatalf("init server failed: %s", err.Error())
	}
	if err := s.Run(); err != nil {
		blog.Fatalf("run server failed: %s", err.Error())
	}
}
