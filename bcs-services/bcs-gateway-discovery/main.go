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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-gateway-discovery/app"
)

//disovery now is designed for stage that bkbcs routes http/https traffics.
// so layer 4 traffic forwarding to bkbcs backends are temporarily out of our consideration.

func main() {
	//init command line option
	opt := app.NewServerOptions()
	//option parsing
	conf.Parse(opt)
	//init logger
	blog.InitLogs(opt.LogConfig)
	defer blog.CloseLogs()
	//create app, init
	svc := app.New()
	if err := svc.Init(opt); err != nil {
		fmt.Printf("bcs-gateway-discovery init failed, %s. service exit\n", err.Error())
		os.Exit(-1)
	}
	//system signal trap for server exit
	go func() {
		//system signal trap
		close := make(chan os.Signal, 10)
		signal.Notify(close, syscall.SIGINT, syscall.SIGTERM)
		<-close
		blog.Infof("bcs-gateway-dicovery catch exit signal, exit in 3 seconds...")
		fmt.Printf("bcs-gateway-dicovery catch exit signal, exit in 3 seconds...\n")
		svc.Stop()
		time.Sleep(time.Second * 3)
	}()
	//start running
	if err := svc.Run(); err != nil {
		fmt.Printf("bcs-gateway-discovery enter running loop failed, %s\n", err.Error())
		os.Exit(-1)
	}
}
