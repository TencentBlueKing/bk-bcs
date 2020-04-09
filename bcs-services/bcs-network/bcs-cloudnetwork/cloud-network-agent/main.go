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
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/cloud-network-agent/controller"
	"bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/cloud-network-agent/options"
	eniaws "bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/eni/aws"
	"bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/networkutil"
	"bk-bcs/bcs-services/bcs-network/bcs-cloudnetwork/pkg/nodenetwork"
)

func main() {
	op := options.New()
	options.Parse(op)

	blog.InitLogs(op.LogConfig)
	defer blog.CloseLogs()

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	nodeNetClient := nodenetwork.New(op.Kubeconfig, op.KubeResyncPeriod, op.KubeCacheSyncTimeout)
	if err := nodeNetClient.Init(); err != nil {
		blog.Fatalf("init node network client failed, err %s", err.Error())
	}

	netutil := new(networkutil.NetUtil)

	ifacesStr := strings.Replace(op.Ifaces, ";", ",", -1)
	ifaces := strings.Split(ifacesStr, ",")
	instanceIP, err := netutil.GetAvailableHostIP(ifaces)
	if err != nil {
		blog.Fatalf("get node ip failed, err %s", err.Error())
	}

	eniClient := eniaws.New(instanceIP)

	ctrl := controller.New(op, nodeNetClient, eniClient, netutil)
	if err := ctrl.GetNodeNetwork(); err != nil {
		blog.Fatalf("get node network failed, err %s", err.Error())
	}

	err = ctrl.Init()
	if err != nil {
		blog.Fatalf("init controller failed, err %s", err.Error())
	}

	wg.Add(1)
	go ctrl.Run(ctx, &wg)

	interupt := make(chan os.Signal, 10)
	signal.Notify(interupt, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM,
		syscall.SIGUSR1, syscall.SIGUSR2)
	for {
		select {
		case <-interupt:
			fmt.Printf("Get signal from system. Exit\n")
			cancel()
			wg.Wait()
			return
		}
	}
}
