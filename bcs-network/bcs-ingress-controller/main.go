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
	"flag"
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	clbv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/clb/v1"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/controllers"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/option"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = networkextensionv1.AddToScheme(scheme)
	_ = clbv1.AddToScheme(scheme)
}

func main() {

	opts := &option.ControllerOption{}

	var verbosity int
	flag.StringVar(&opts.Address, "address", "127.0.0.1", "address for controller")
	flag.IntVar(&opts.MetricPort, "metric_port", 8081, "metric port for controller")
	flag.IntVar(&opts.Port, "port", 8080, "por for controller")
	flag.StringVar(&opts.Cloud, "cloud", "tencentcloud", "cloud mode for bcs network controller")
	flag.StringVar(&opts.Cluster, "cluster", "", "clusterid for bcs cluster")

	flag.StringVar(&opts.LogDir, "log_dir", "./logs", "If non-empty, write log files in this directory")
	flag.Uint64Var(&opts.LogMaxSize, "log_max_size", 500, "Max size (MB) per log file.")
	flag.IntVar(&opts.LogMaxNum, "log_max_num", 10, "Max num of log file.")
	flag.BoolVar(&opts.ToStdErr, "logtostderr", false, "log to standard error instead of files")
	flag.BoolVar(&opts.AlsoToStdErr, "alsologtostderr", false, "log to standard error as well as files")

	flag.IntVar(&verbosity, "v", 0, "log level for V logs")
	flag.StringVar(&opts.StdErrThreshold, "stderrthreshold", "2", "logs at or above this threshold go to stderr")
	flag.StringVar(&opts.VModule, "vmodule", "", "comma-separated list of pattern=N settings for file-filtered logging")
	flag.StringVar(&opts.TraceLocation, "log_backtrace_at", "", "when logging hits line file:N, emit a stack trace")

	flag.Parse()

	opts.Verbosity = int32(verbosity)

	blog.InitLogs(opts.LogConfig)
	defer blog.CloseLogs()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      opts.Address + ":" + strconv.Itoa(opts.MetricPort),
		LeaderElection:          false,
		LeaderElectionID:        "33fb49e.cloudlbconroller.bkbcs.tencent.com",
		LeaderElectionNamespace: "bcs-system",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.IngressReconciler{
		Client:         mgr.GetClient(),
		Log:            ctrl.Log.WithName("controllers").WithName("Ingress"),
		Option:         opts,
		IngressEventer: mgr.GetEventRecorderFor("bcs-ingress-controller"),
		SvcFilter:      controllers.NewServiceFilter(mgr.GetClient()),
		PodFilter:      controllers.NewPodFilter(mgr.GetClient()),
		StsFilter:      controllers.NewStatefulSetFilter(mgr.GetClient()),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Ingress")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
