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
	"crypto/tls"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/version"
	registry "github.com/Tencent/bk-bcs/bcs-common/pkg/registry"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/tracing"
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

	// 初始化 Tracer
	shutdown, errorInitTracing := tracing.InitTracing(&op.TracingConf)
	if errorInitTracing != nil {
		blog.Info(errorInitTracing.Error())
	}
	if shutdown != nil {
		defer func() {
			if err := shutdown(context.Background()); err != nil {
				blog.Infof("failed to shutdown TracerProvider: %s", err.Error())
			}
		}()
	}

	// etcd registry
	etcdRegistry, err := turnOnEtcdRegistry(op)
	if err != nil {
		blog.Errorf("turnOnEtcdRegistry failed: %v", err.Error())
		// nolint
		os.Exit(1)
	}
	defer func() {
		if etcdRegistry != nil {
			// waiting for api gateway to close all connections
			blog.Infof("deregister etcd registry")
			err = etcdRegistry.Deregister()
			if err != nil {
				blog.Errorf("deregister etcd registry failed: %v", err.Error())
				os.Exit(1)
			}
			time.Sleep(time.Second * 10)
			blog.Infof("deregister etcd registry done")
		}
		component.GetAuditClient().Close()
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

	go func() {
		blog.Info("", http.ListenAndServe(":6060", nil))
	}()

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	blog.Infof("Got OS shutdown signal, shutting down bcs-user-manager server gracefully...")

}

// turnOnEtcdRegistry xxx
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

	ipv4 := opt.Address
	ipv6 := opt.IPv6Address
	port := strconv.Itoa(int(opt.Port))

	// service inject metadata to discovery center
	metadata := make(map[string]string)
	metadata["httpport"] = strconv.Itoa(int(opt.Port))

	// 适配单栈环境（ipv6注册地址不能是本地回环地址）
	if v := net.ParseIP(ipv6); v != nil && !v.IsLoopback() {
		metadata[types.IPV6] = net.JoinHostPort(ipv6, port)
	}

	// init go-micro registry
	eOption := &registry.Options{
		Name:         userManager,
		Version:      version.BcsVersion,
		RegistryAddr: strings.Split(opt.Etcd.Address, ","),
		RegAddr:      net.JoinHostPort(ipv4, port),
		Config:       tlsCfg,
		Meta:         metadata,
	}
	etcdRegistry := registry.NewEtcdRegistry(eOption)
	if err := etcdRegistry.Register(); err != nil {
		blog.Errorf("etcd registry feature turn on but register failed, %s", err.Error())
		return nil, err
	}

	return etcdRegistry, nil
}
