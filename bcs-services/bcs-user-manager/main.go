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
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/registry"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/job/notify"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/options"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	op := &options.UserManagerOptions{}
	conf.Parse(op)

	blog.InitLogs(op.LogConfig)
	defer blog.CloseLogs()

	app.Run(op)

	// etcd registry
	etcdRegistry, err := turnOnEtcdRegistry(op)
	if err != nil {
		blog.Errorf("turnOnEtcdRegistry failed: %v", err.Error())
		os.Exit(1)
	}
	defer func() {
		if etcdRegistry != nil {
			// waiting for api gateway to close all connections
			time.Sleep(time.Second * 5)
			_ = etcdRegistry.Deregister()
		}
	}()

	// sync expired token and notify
	if op.TokenNotify.Feature {
		blog.Info("start token notify job")
		tokenNotify, err := notify.NewTokenNotify(op)
		if err != nil {
			blog.Fatalf("new token notify failed, %s", err.Error())
		}
		go tokenNotify.Run()
		defer tokenNotify.Stop()
	}

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	blog.Infof("Got OS shutdown signal, shutting down bcs-user-manager server gracefully...")

	return
}

// register user-manager service to etcd
func turnOnEtcdRegistry(opt *options.UserManagerOptions) (registry.Registry, error) {
	if !opt.Etcd.Feature {
		return nil, nil
	}

	const (
		userManager = "usermanager.bkbcs.tencent.com"
	)

	var tlsCfg *tls.Config
	if !opt.InsecureEtcd {
		var err error
		tlsCfg, err = opt.Etcd.GetTLSConfig()
		if err != nil {
			blog.Errorf("turn on etcd registry feature but configuration not correct, %s", err.Error())
			os.Exit(1)
		}
	}

	// init go-micro registry
	eOption := &registry.Options{
		Name:         userManager,
		Version:      version.BcsVersion,
		RegistryAddr: strings.Split(opt.Etcd.Address, ","),
		RegAddr:      fmt.Sprintf("%s:%d", opt.Address, opt.Port),
		Config:       tlsCfg,
	}
	etcdRegistry := registry.NewEtcdRegistry(eOption)
	if err := etcdRegistry.Register(); err != nil {
		blog.Errorf("etcd registry feature turn on but register failed, %s", err.Error())
		return nil, err
	}

	return etcdRegistry, nil
}
