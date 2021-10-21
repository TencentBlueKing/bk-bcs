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
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/app"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-loadbalance/option"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
)

func main() {
	var err error
	config := option.NewConfig()
	if err = config.Parse(); err != nil {
		fmt.Printf("parse config failed, err %s", err.Error())
		blog.Errorf("parse config failed, err %s", err.Error())
		os.Exit(1)
	}
	app.InitLogger(config)
	defer app.CloseLogger()
	processor := app.NewEventProcessor(config)
	if err = processor.Start(); err != nil {
		processor.Stop()
		fmt.Printf("bcs-loadbalance starting error %s\n", err.Error())
		blog.Errorf("bcs-loadbalance starting error %s", err.Error())
		os.Exit(1)
	}
	interrupt := make(chan os.Signal, 10)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGUSR1, syscall.SIGUSR2)
	processor.HandleSignal(interrupt)
}
