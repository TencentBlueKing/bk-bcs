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
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	clbv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/clb/v1"
	ingressctrl "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/ingresscontroller"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/aws"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/gcp"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/namespacedlb"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/tencentcloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloudcollector"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/generator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/webhookserver"
	listenerctrl "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/listenercontroller"
	portbindingctrl "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/portbindingcontroller"
	portpoolctrl "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/portpoolcontroller"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = networkextensionv1.AddToScheme(scheme)
	_ = clbv1.AddToScheme(scheme)
}

func main() {

	opts := &option.ControllerOption{}
	var verbosity int
	var checkIntervalStr string
	flag.StringVar(&opts.Address, "address", "127.0.0.1", "address for controller")
	flag.IntVar(&opts.MetricPort, "metric_port", 8081, "metric port for controller")
	flag.IntVar(&opts.Port, "port", 8080, "por for controller")
	flag.StringVar(&opts.Cloud, "cloud", "tencentcloud", "cloud mode for controller")
	flag.StringVar(&opts.Region, "region", "", "default cloud region for controller")
	flag.StringVar(&opts.ElectionNamespace, "election_namespace", "bcs-system", "namespace for leader election")
	flag.BoolVar(&opts.IsNamespaceScope, "is_namespace_scope", false,
		"if the ingress can only be associated with the service and workload in the same namespace")
	flag.StringVar(&checkIntervalStr, "portbinding_check_interval", "3m",
		"check interval of port binding, golang time format")

	flag.StringVar(&opts.LogDir, "log_dir", "./logs", "If non-empty, write log files in this directory")
	flag.Uint64Var(&opts.LogMaxSize, "log_max_size", 500, "Max size (MB) per log file.")
	flag.IntVar(&opts.LogMaxNum, "log_max_num", 10, "Max num of log file.")
	flag.BoolVar(&opts.ToStdErr, "logtostderr", false, "log to standard error instead of files")
	flag.BoolVar(&opts.AlsoToStdErr, "alsologtostderr", false, "log to standard error as well as files")

	flag.IntVar(&verbosity, "v", 0, "log level for V logs")
	flag.StringVar(&opts.StdErrThreshold, "stderrthreshold", "2", "logs at or above this threshold go to stderr")
	flag.StringVar(&opts.VModule, "vmodule", "", "comma-separated list of pattern=N settings for file-filtered logging")
	flag.StringVar(&opts.TraceLocation, "log_backtrace_at", "", "when logging hits line file:N, emit a stack trace")

	flag.StringVar(&opts.ServerCertFile, "server_cert_file", "", "server cert file for webhook server")
	flag.StringVar(&opts.ServerKeyFile, "server_key_file", "", "server key file for webhook server")

	flag.Parse()

	opts.Verbosity = int32(verbosity)
	checkInterval, err := time.ParseDuration(checkIntervalStr)
	if err != nil {
		fmt.Printf("check interval %s invalid", checkIntervalStr)
		os.Exit(1)
	}
	opts.PortBindingCheckInterval = checkInterval

	blog.InitLogs(opts.LogConfig)
	defer blog.CloseLogs()

	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))

	// get env var name for tcp and udp port reuse
	isTCPUDPPortReuseStr := os.Getenv(constant.EnvNameIsTCPUDPPortReuse)
	if len(isTCPUDPPortReuseStr) != 0 {
		blog.Infof("env option %s is %s", constant.EnvNameIsTCPUDPPortReuse, isTCPUDPPortReuseStr)
		isTCPUDPPortReuse, err := strconv.ParseBool(isTCPUDPPortReuseStr)
		if err != nil {
			blog.Errorf("parse bool string %s failed, err %s", isTCPUDPPortReuseStr, err.Error())
			os.Exit(1)
		}
		if isTCPUDPPortReuse {
			opts.IsTCPUDPPortReuse = isTCPUDPPortReuse
		}
	}

	// get env var name for bulk mode
	isBulkModeStr := os.Getenv(constant.EnvNameIsBulkMode)
	if len(isBulkModeStr) != 0 {
		blog.Infof("env option %s is %s", constant.EnvNameIsBulkMode, isBulkModeStr)
		isBulkMode, err := strconv.ParseBool(isBulkModeStr)
		if err != nil {
			blog.Errorf("parse bool string %s failed, err %s", isBulkModeStr, err.Error())
			os.Exit(1)
		}
		if isBulkMode {
			opts.IsBulkMode = isBulkMode
		}
	}

	// init port pool cache
	portPoolCache := portpoolcache.NewCache()
	go portPoolCache.Start()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      opts.Address + ":" + strconv.Itoa(opts.MetricPort),
		LeaderElection:          true,
		LeaderElectionID:        "33fb49e.cloudlbconroller.bkbcs.tencent.com",
		LeaderElectionNamespace: opts.ElectionNamespace,
	})
	if err != nil {
		blog.Errorf("unable to start manager, err %s", err.Error())
		os.Exit(1)
	}

	var validater cloud.Validater
	var lbClient cloud.LoadBalance
	switch opts.Cloud {
	case constant.CloudTencent:
		validater = tencentcloud.NewClbValidater()
		if !opts.IsNamespaceScope {
			lbClient, err = tencentcloud.NewClb()
			if err != nil {
				blog.Errorf("init cloud failed, err %s", err.Error())
				os.Exit(1)
			}
		} else {
			lbClient = namespacedlb.NewNamespacedLB(mgr.GetClient(), tencentcloud.NewClbWithSecret)
		}

	case constant.CloudAWS:
		validater = aws.NewELbValidater()
		if !opts.IsNamespaceScope {
			lbClient, err = aws.NewElb()
			if err != nil {
				blog.Errorf("init cloud failed, err %s", err.Error())
				os.Exit(1)
			}
		} else {
			lbClient = namespacedlb.NewNamespacedLB(mgr.GetClient(), aws.NewElbWithSecret)
		}

	case constant.CloudGCP:
		validater = gcp.NewGclbValidater()
		if !opts.IsNamespaceScope {
			lbClient, err = gcp.NewGclb(mgr.GetClient())
			if err != nil {
				blog.Errorf("init cloud failed, err %s", err.Error())
				os.Exit(1)
			}
		} else {
			lbClient = namespacedlb.NewNamespacedLB(mgr.GetClient(), gcp.NewGclbWithSecret)
		}
	}

	if len(opts.Region) == 0 {
		blog.Errorf("region cannot be empty")
		os.Exit(1)
	}

	ingressConverter, err := generator.NewIngressConverter(&generator.IngressConverterOpt{
		DefaultRegion:     opts.Region,
		IsTCPUDPPortReuse: opts.IsTCPUDPPortReuse,
	}, mgr.GetClient(), validater, lbClient)
	if err != nil {
		blog.Errorf("create ingress converter failed, err %s", err.Error())
		os.Exit(1)
	}
	if err = (&ingressctrl.IngressReconciler{
		Ctx:              context.Background(),
		Client:           mgr.GetClient(),
		Log:              ctrl.Log.WithName("controllers").WithName("Ingress"),
		Option:           opts,
		IngressEventer:   mgr.GetEventRecorderFor("bcs-ingress-controller"),
		SvcFilter:        ingressctrl.NewServiceFilter(mgr.GetClient()),
		EpsFilter:        ingressctrl.NewEndpointsFilter(mgr.GetClient()),
		PodFilter:        ingressctrl.NewPodFilter(mgr.GetClient()),
		StsFilter:        ingressctrl.NewStatefulSetFilter(mgr.GetClient()),
		IngressConverter: ingressConverter,
	}).SetupWithManager(mgr); err != nil {
		blog.Errorf("unable to create ingress reconciler, err %s", err.Error())
		os.Exit(1)
	}

	listenerReconciler := listenerctrl.NewListenerReconciler()
	listenerReconciler.Ctx = context.Background()
	listenerReconciler.Client = mgr.GetClient()
	listenerReconciler.CloudLb = lbClient
	listenerReconciler.Option = opts
	listenerReconciler.ListenerEventer = mgr.GetEventRecorderFor("bcs-ingress-controller")
	if err = listenerReconciler.SetupWithManager(mgr); err != nil {
		blog.Errorf("unable to create listener reconciler, err %s", err.Error())
		os.Exit(1)
	}

	portPoolReconciler := portpoolctrl.NewPortPoolReconciler(context.Background(), opts, lbClient,
		mgr.GetClient(), mgr.GetEventRecorderFor("bcs-ingress-controller"), portPoolCache)
	if err = portPoolReconciler.SetupWithManager(mgr); err != nil {
		blog.Errorf("unable to create port pool reconciler, err %s", err.Error())
		os.Exit(1)
	}

	portBindingReconciler := portbindingctrl.NewPortBindingReconciler(
		context.Background(), opts.PortBindingCheckInterval, mgr.GetClient(), portPoolCache)
	if err = portBindingReconciler.SetupWithManager(mgr); err != nil {
		blog.Errorf("unable to create port binding reconciler, err %s", err.Error())
		os.Exit(1)
	}

	// init webhook server
	webhookServerOpts := &webhookserver.ServerOption{
		Addr:           opts.Address,
		Port:           opts.Port,
		ServerCertFile: opts.ServerCertFile,
		ServerKeyFile:  opts.ServerKeyFile,
	}
	webhookServer, err := webhookserver.NewHookServer(webhookServerOpts, mgr.GetClient(), portPoolCache)
	if err != nil {
		blog.Errorf("create hook server failed, err %s", err.Error())
		os.Exit(1)
	}
	go webhookServer.Start()
	blog.Infof("webhook server started")

	// init cloud loadbalance backend status collector
	collector := cloudcollector.NewCloudCollector(lbClient, mgr.GetClient())
	go collector.Start()
	metrics.Registry.MustRegister(collector)

	blog.Infof("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		blog.Errorf("problem running manager, err %s", err.Error())
		os.Exit(1)
	}
}
