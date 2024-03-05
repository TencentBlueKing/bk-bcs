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
	"net/http"
	"os"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-node-external-worker/exporter"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-node-external-worker/httpsvr"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-node-external-worker/options"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
}

func main() {
	opts := options.Options{}
	opts.BindFromCommandLine()
	blog.InitLogs(opts.LogConfig)
	defer blog.CloseLogs()

	opts.SetFromEnv()

	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0", // "0"表示禁用默认的Metric Service， 需要使用自己的实现支持IPV6
		LeaderElection:     false,
	})
	if err != nil {
		blog.Fatalf("start ctrl manager failed, err: %s", err.Error())
		os.Exit(1) // nolint
	}

	blog.Infof("node-external-worker start...")

	httpServer := &httpsvr.HttpServerClient{Ops: opts}
	if err = httpServer.Init(); err != nil {
		blog.Fatalf("init http sever failed, err: %s", err.Error())
		os.Exit(1) // nolint
	}

	nodeExporter := exporter.NodeExporter{
		Ctx:        context.Background(),
		K8sClient:  mgr.GetClient(),
		Opts:       opts,
		HttpClient: http.Client{Timeout: time.Second * 2},
		HttpSvr:    httpServer,
	}
	go nodeExporter.Watch()

	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		blog.Errorf("problem running manager, err %s", err.Error())
		os.Exit(1) // nolint
	}
}
