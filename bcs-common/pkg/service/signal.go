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

package service

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

// SetupSignalContext catch exit signal for process graceful exit
func SetupSignalContext() context.Context {
	closeHandler := make(chan os.Signal, 2)
	ctx, cancel := context.WithCancel(context.Background())
	signal.Notify(closeHandler, syscall.SIGINT, syscall.SIGTERM)
	//waiting for signal
	go func() {
		<-closeHandler
		blog.Infof("service catch exit signal, exit in 3 seconds...")
		fmt.Printf("service catch exit signal, exit in 3 seconds...\n")
		cancel()
		time.Sleep(time.Second * 3)
		<-closeHandler
		fmt.Printf("service catch second exit signal, exit immediately...\n")
		os.Exit(-2)
	}()
	return ctx
}
