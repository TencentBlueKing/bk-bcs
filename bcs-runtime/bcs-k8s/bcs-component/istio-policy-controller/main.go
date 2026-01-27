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
// package xxx
package main

import (
	"flag"
	"os"
	"strconv"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/istio-policy-controller/controllers"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/istio-policy-controller/internal/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/istio-policy-controller/pkg/config"
	networkingv1 "istio.io/client-go/pkg/apis/networking/v1"
	"istio.io/client-go/pkg/clientset/versioned"

	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
	Config   string
)

func main() {
	opts := &option.ControllerOption{}

	flag.StringVar(&opts.Address, "address", "127.0.0.1", "address for controller")
	flag.IntVar(&opts.MetricPort, "metric_port", 8081, "metric port for controller")
	flag.StringVar(&opts.LogDir, "log_dir", "./logs", "If non-empty, write log files in this directory")
	flag.Uint64Var(&opts.LogMaxSize, "log_max_size", 500, "Max size (MB) per log file.")
	flag.IntVar(&opts.LogMaxNum, "log_max_num", 10, "Max num of log file.")
	flag.BoolVar(&opts.ToStdErr, "logtostderr", false, "log to standard error instead of files")
	flag.StringVar(&Config, "config", "./etc/config.yaml", "config file path")

	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	setupLog.Info("starting init config")
	// 初始化配置
	err := config.Init(Config)
	if err != nil {
		setupLog.Error(err, "unable to init config")
		os.Exit(1)
	}

	cfg, err := ctrl.GetConfig()
	if err != nil {
		setupLog.Error(err, "unable to get kubeconfig")
		os.Exit(1)
	}

	// 创建 Istio client
	istioClient, err := versioned.NewForConfig(cfg)
	if err != nil {
		setupLog.Error(err, "unable to create Istio client")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      opts.Address + ":" + strconv.Itoa(opts.MetricPort),
		LeaderElection:          true,
		LeaderElectionID:        "333fb49e.istioconroller.bkbcs.tencent.com",
		LeaderElectionNamespace: "default",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1) // nolint
	}

	// 注册 Istio networking v1 类型
	if err = networkingv1.AddToScheme(mgr.GetScheme()); err != nil {
		setupLog.Error(err, "unable to add Istio networking v1 to scheme")
		os.Exit(1)
	}

	if err = (&controllers.ServiceReconciler{
		Client:      mgr.GetClient(),
		Log:         ctrl.Log.WithName("controllers").WithName("service"),
		Option:      opts,
		IstioClient: istioClient,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create k8s Service controller", "controller", "Service")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
