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
	"runtime"
	"strings"
	"syscall"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/options"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	op := &options.UserManagerOptions{}
	conf.Parse(op)

	blog.InitLogs(op.LogConfig)
	defer blog.CloseLogs()

	app.Run(op)
	//etcd register
	if op.Etcd.Feature {
		tlsCfg, err := op.Etcd.GetTLSConfig()
		if err != nil {
			blog.Errorf("turn on etcd registry feature but configuration not correct, %s", err.Error())
			os.Exit(1)
		}
		// init go-micro registry
		eoption := &registry.Options{
			Name:         "usermanager.bkbcs.tencent.com",
			Version:      version.BcsVersion,
			RegistryAddr: strings.Split(op.Etcd.Address, ","),
			RegAddr:      fmt.Sprintf("%s:%d", op.Address, op.Port),
			Config:       tlsCfg,
		}
		etcdRegistry := registry.NewEtcdRegistry(eoption)
		if err := etcdRegistry.Register(); err != nil {
			blog.Errorf("etcd registry feature turn on but register failed, %s", err.Error())
			os.Exit(1)
		}
		defer func() {
			//when exit, clean registered information
			if op.Etcd.Feature {
				etcdRegistry.Deregister()
			}
		}()
	}

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	blog.Infof("Got OS shutdown signal, shutting down bcs-user-manager server gracefully...")

	return
}
