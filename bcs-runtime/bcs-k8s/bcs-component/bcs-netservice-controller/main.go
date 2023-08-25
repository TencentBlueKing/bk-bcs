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
	"fmt"
	"os"
	"strconv"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	netservicev1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/api/v1"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/controllers"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/internal/httpsvr"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-netservice-controller/internal/option"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(netservicev1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	opts := &option.ControllerOption{}

	var verbosity int
	flag.StringVar(&opts.Address, "address", "127.0.0.1", "address for controller")
	flag.IntVar(&opts.ProbePort, "probe_port", 8082, "probe port for controller")
	flag.IntVar(&opts.MetricPort, "metric_port", 8081, "metric port for controller")
	flag.IntVar(&opts.Port, "port", 8080, "port for controller")

	flag.BoolVar(&opts.EnableLeaderElect, "leader-elect", true, "enable leader elect for controller")

	flag.StringVar(&opts.LogDir, "log_dir", "./logs", "If non-empty, write log files in this directory")
	flag.Uint64Var(&opts.LogMaxSize, "log_max_size", 500, "Max size (MB) per log file.")
	flag.IntVar(&opts.LogMaxNum, "log_max_num", 10, "Max num of log file.")
	flag.BoolVar(&opts.ToStdErr, "logtostderr", false, "log to standard error instead of files")
	flag.BoolVar(&opts.AlsoToStdErr, "alsologtostderr", false, "log to standard error as well as files")

	flag.IntVar(&verbosity, "v", 0, "log level for V logs")
	flag.StringVar(&opts.StdErrThreshold, "stderrthreshold", "2", "logs at or above this threshold go to stderr")
	flag.StringVar(&opts.VModule, "vmodule", "", "comma-separated list of pattern=N settings for file-filtered logging")
	flag.StringVar(&opts.TraceLocation, "log_backtrace_at", "", "when logging hits line file:N, emit a stack trace")

	flag.UintVar(&opts.HttpServerPort, "http_svr_port", 8088, "port for controller http server")

	flag.Parse()
	opts.Verbosity = int32(verbosity)

	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))

	blog.InitLogs(opts.LogConfig)
	defer blog.CloseLogs()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      opts.Address + ":" + strconv.Itoa(opts.MetricPort),
		Port:                    opts.Port,
		HealthProbeBindAddress:  opts.Address + ":" + strconv.Itoa(opts.ProbePort),
		LeaderElection:          opts.EnableLeaderElect,
		LeaderElectionID:        "ca387ddc.netservice.bkbcs.tencent.com",
		LeaderElectionNamespace: "bcs-system",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.BCSNetPoolReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		IPFilter: controllers.NewIPFilter(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "BCSNetPool")
		os.Exit(1)
	}
	if err = (&netservicev1.BCSNetPool{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "BCSNetPool")
		os.Exit(1)
	}
	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	if err := initHttpServer(opts, mgr.GetClient()); err != nil {
		blog.Errorf("init http server failed, err %v", err)
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// initHttpServer init netservice controller http server
// httpServer提供
// 1. 申请IP地址接口
// 2. 释放IP地址接口
func initHttpServer(op *option.ControllerOption, client client.Client) error {
	server := httpserver.NewHttpServer(op.HttpServerPort, op.Address, "")
	if op.Conf.ServCert.IsSSL {
		server.SetSsl(op.Conf.ServCert.CAFile, op.Conf.ServCert.CertFile, op.Conf.ServCert.KeyFile,
			op.Conf.ServCert.CertPasswd)
	}

	server.SetInsecureServer(op.Address, op.HttpServerPort)
	ws := server.NewWebService("/netservicecontroller", nil)
	httpServerClient := &httpsvr.HttpServerClient{
		K8SClient: client,
	}
	httpsvr.InitRouters(ws, httpServerClient)

	router := server.GetRouter()
	webContainer := server.GetWebContainer()
	router.Handle("/netservicecontroller/{sub_path:.*}", webContainer)
	blog.Infof("Starting http server on %s:%d", op.Address, op.HttpServerPort)
	if err := server.ListenAndServeMux(op.Conf.VerifyClientTLS); err != nil {
		return fmt.Errorf("http ListenAndServe error %s", err.Error())
	}
	return nil
}
