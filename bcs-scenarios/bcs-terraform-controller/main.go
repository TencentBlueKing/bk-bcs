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
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them. nolint
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/go-git/go-git/v5/plumbing"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// +kubebuilder:scaffold:imports nolint
	tfv1 "github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/controllers"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/option"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/server"
	"github.com/Tencent/bk-bcs/bcs-scenarios/bcs-terraform-controller/pkg/utils"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)
var (
	// verbosity
	verbosity int
	// probeAddr addr
	probeAddr string
	// metricsAddr m addr
	metricsAddr string
	// enableLeaderElection election
	enableLeaderElection bool
	// opts global config
	opts = &option.ControllerOption{}
)

// 端口变更！参考deployment.yaml
func init() {
	utilruntime.Must(tfv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	// git.Repository的ResolveRevision()方法描述与实际实现不同，缺少origin的情况的实现，故在这里加上origin条件
	plumbing.RefRevParseRules = append(plumbing.RefRevParseRules, "refs/remotes/origin/%s")

	// metrics config
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8081", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8082", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager. "+
		"Enabling this will ensure there is only one active controller manager.")
	// log config
	flag.IntVar(&verbosity, "v", 3, "log level for V logs")
	flag.StringVar(&opts.LogDir, "log_dir", "/data/bcs/logs", "If non-empty, write log files in this directory")
	flag.Uint64Var(&opts.LogMaxSize, "log_max_size", 500, "Max size (MB) per log file.")
	flag.IntVar(&opts.LogMaxNum, "log_max_num", 10, "Max num of log file.")
	flag.BoolVar(&opts.ToStdErr, "logtostderr", false, "log to standard error instead of files")
	flag.BoolVar(&opts.AlsoToStdErr, "alsologtostderr", true, "log to standard error as well as files")
	opts.Verbosity = int32(verbosity)
	// consul config
	flag.StringVar(&opts.ConsulScheme, "consul_scheme", "", "tf cli backend consul scheme")
	flag.StringVar(&opts.ConsulAddress, "consul_address", "", "tf cli backend consul address")
	flag.StringVar(&opts.ConsulPath, "consul_path", "", "tf cli backend consul path")
	// consul config
	flag.StringVar(&opts.GitopsHost, "gitops_host", "", "gitops host")
	flag.StringVar(&opts.GitopsUsername, "gitops_username", "", "gitops username")
	flag.StringVar(&opts.GitopsPassword, "gitops_password", "", "gitops password")
	// vault
	flag.StringVar(&opts.VaultCaPath, "vault_ca_path", "/data/bcs/cert/vault/vaultca", "vault private ca path")
	flag.Parse()

	blog.InitLogs(opts.LogConfig)

	blog.Infof("controller config: %s", utils.ToJsonString(opts))
	if err := option.CheckControllerOption(opts); err != nil {
		blog.Fatalf("check controllerOption failed, err: %s", err)
		return
	}
	// +kubebuilder:scaffold:scheme
}

func main() {
	defer blog.CloseLogs()
	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "333fb49e.bkbcs.tencent.com",
		// LeaderElectionReleaseOnCancel defines if the leader should step down voluntarily
		// when the Manager ends. This requires the binary to immediately end when the
		// Manager is stopped, otherwise, this setting is unsafe. Setting this significantly
		// speeds up voluntary leader transitions as the new leader don't have to wait
		// LeaseDuration time first.
		//
		// In the default scaffold provided, the program ends immediately after
		// the manager stops, so would be fine to enable this option. However,
		// if you are doing or is intended to do any operation such as perform cleanups
		// after the manager stops then its usage might be unsafe.
		// LeaderElectionReleaseOnCancel: true,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1) // nolint
	}

	svr := server.NewHandler("0.0.0.0", "8080", mgr.GetClient())
	go func() {
		if err = svr.Init(); err != nil {
			blog.Fatalf("http server init failed, err: %s", err)
			return
		}
		if err = svr.Run(); err != nil {
			blog.Fatalf("http server run terminated, err: %s", err)
			return
		}
	}()

	if err = (&controllers.TerraformReconciler{
		Config: opts,
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Terraform")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	if err = mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err = mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
