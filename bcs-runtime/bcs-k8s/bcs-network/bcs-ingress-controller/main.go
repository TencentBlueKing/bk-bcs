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
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/ipv6server"
	clbv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/clb/v1"
	ingressctrl "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/ingresscontroller"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/check"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/aws"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/azure"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/gcp"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/namespacedlb"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/tencentcloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloudcollector"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/conflicthandler"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/eventer"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/generator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/httpsvr"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/ingresscache"
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

	flag.IntVar(&opts.KubernetesQPS, "kubernetes_qps", 100, "the qps of k8s client request")
	flag.IntVar(&opts.KubernetesBurst, "kubernetes_burst", 200, "the burst of k8s client request")

	flag.BoolVar(&opts.ConflictCheckOpen, "conflict_check_open", true, "if false, "+
		"skip all conflict checking about ingress and port pool")

	flag.UintVar(&opts.HttpServerPort, "http_svr_port", 8082, "port for ingress controller http server")

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

	podIPs := os.Getenv(constant.EnvNamePodIPs)
	if len(podIPs) == 0 {
		blog.Errorf("empty pod ip")
		podIPs = opts.Address
	}
	blog.Infof("pod ips: %s", podIPs)
	opts.PodIPs = strings.Split(podIPs, ",")

	// init port pool cache
	portPoolCache := portpoolcache.NewCache()
	go portPoolCache.Start()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      "0",
		LeaderElection:          true,
		LeaderElectionID:        "33fb49e.cloudlbconroller.bkbcs.tencent.com",
		LeaderElectionNamespace: opts.ElectionNamespace,
		NewClient: func(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
			config.QPS = float32(opts.KubernetesQPS)
			config.Burst = opts.KubernetesBurst
			// Create the Client for Write operations.
			c, err := client.New(config, options)
			if err != nil {
				return nil, err
			}

			return &client.DelegatingClient{
				Reader: &client.DelegatingReader{
					CacheReader:  cache,
					ClientReader: c,
				},
				Writer:       c,
				StatusClient: c,
			}, nil
		},
	})
	if err != nil {
		blog.Errorf("unable to start manager, err %s", err.Error())
		os.Exit(1)
	}
	runPrometheusMetrics(opts)

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
	case constant.CloudAzure:
		validater = azure.NewAlbValidater()
		if !opts.IsNamespaceScope {
			lbClient, err = azure.NewAlb()
			if err != nil {
				blog.Errorf("init cloud failed, err %s", err.Error())
				os.Exit(1)
			}
		} else {
			lbClient = namespacedlb.NewNamespacedLB(mgr.GetClient(), azure.NewAlbWithSecret)
		}
	default:
		blog.Errorf("unknown cloud type '%s'", opts.Cloud)
		os.Exit(1)
	}

	if len(opts.Region) == 0 {
		blog.Errorf("region cannot be empty")
		os.Exit(1)
	}

	ingressConverter, err := generator.NewIngressConverter(&generator.IngressConverterOpt{
		DefaultRegion:     opts.Region,
		IsTCPUDPPortReuse: opts.IsTCPUDPPortReuse,
		Cloud:             opts.Cloud,
	}, mgr.GetClient(), validater, lbClient)
	if err != nil {
		blog.Errorf("create ingress converter failed, err %s", err.Error())
		os.Exit(1)
	}
	ingressCache := ingresscache.NewDefaultCache()
	if err = (&ingressctrl.IngressReconciler{
		Ctx:              context.Background(),
		Client:           mgr.GetClient(),
		Log:              ctrl.Log.WithName("controllers").WithName("Ingress"),
		Option:           opts,
		IngressEventer:   mgr.GetEventRecorderFor("bcs-ingress-controller"),
		EpsFIlter:        ingressctrl.NewEndpointsFilter(mgr.GetClient(), ingressCache),
		PodFilter:        ingressctrl.NewPodFilter(mgr.GetClient(), ingressCache),
		IngressConverter: ingressConverter,
		Cache:            ingressCache,
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
		context.Background(), opts.PortBindingCheckInterval, mgr.GetClient(), portPoolCache, mgr.GetEventRecorderFor("bcs-ingress-controller"))
	if err = portBindingReconciler.SetupWithManager(mgr); err != nil {
		blog.Errorf("unable to create port binding reconciler, err %s", err.Error())
		os.Exit(1)
	}

	// init event watcher
	k8sClient, err := initInClusterClient()
	if err != nil {
		blog.Fatalf("init in-cluster client failed: %v", err)
	}
	eventClient := eventer.NewKubeEventer(k8sClient)
	if err = eventClient.Init(); err != nil {
		blog.Fatalf("init event watcher failed: %v", err)
	}
	go eventClient.Start(context.Background())

	conflictHandler := conflicthandler.NewConflictHandler(opts.ConflictCheckOpen, opts.IsTCPUDPPortReuse, opts.Region,
		mgr.GetClient(), ingressConverter)
	// init webhook server
	webhookServerOpts := &webhookserver.ServerOption{
		Addrs:          opts.PodIPs,
		Port:           opts.Port,
		ServerCertFile: opts.ServerCertFile,
		ServerKeyFile:  opts.ServerKeyFile,
	}
	webhookServer, err := webhookserver.NewHookServer(webhookServerOpts, mgr.GetClient(), lbClient, portPoolCache,
		eventClient, validater, ingressConverter, conflictHandler)
	if err != nil {
		blog.Errorf("create hook server failed, err %s", err.Error())
		os.Exit(1)
	}
	mgr.Add(webhookServer)

	// init cloud loadbalance backend status collector
	collector := cloudcollector.NewCloudCollector(lbClient, mgr.GetClient())
	go collector.Start()
	metrics.Registry.MustRegister(collector)

	blog.Infof("starting manager")

	err = initHttpServer(opts, mgr)
	if err != nil {
		blog.Errorf("init http server failed: %v", err.Error())
		os.Exit(1)
	}
	blog.Infof("starting http server")

	checkRunner := check.NewCheckRunner(context.Background())
	checkRunner.
		Register(check.NewPortBindChecker(mgr.GetClient(), mgr.GetEventRecorderFor("bcs-ingress-controller"))).
		Register(check.NewListenerChecker(mgr.GetClient())).
		Start()
	blog.Infof("starting check runner")

	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		blog.Errorf("problem running manager, err %s", err.Error())
		os.Exit(1)
	}
}

func initInClusterClient() (*kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "get in-cluster config failed")
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "create in-cluster client failed")
	}
	return client, nil
}

// runPrometheusMetrics starting prometheus metrics handler
func runPrometheusMetrics(op *option.ControllerOption) {
	http.Handle("/metrics", promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{}))
	// ipv4 ipv6
	ipv6Server := ipv6server.NewIPv6Server(op.PodIPs, strconv.Itoa(op.MetricPort), "", nil)
	// 启动server，同时监听v4、v6地址
	go func() {
		if err := ipv6Server.ListenAndServe(); err != nil {
			blog.Errorf("metric server listen err: %v", err)
		}
	}()
}

// initHttpServer init ingress controller http server
func initHttpServer(op *option.ControllerOption, mgr manager.Manager) error {
	server := httpserver.NewHttpServer(op.HttpServerPort, op.Address, "")
	if op.Conf.ServCert.IsSSL {
		server.SetSsl(op.Conf.ServCert.CAFile, op.Conf.ServCert.CertFile, op.Conf.ServCert.KeyFile,
			op.Conf.ServCert.CertPasswd)
	}

	server.SetInsecureServer(op.Conf.InsecureAddress, op.Conf.InsecurePort)
	ws := server.NewWebService("/ingresscontroller", nil)
	httpServerClient := &httpsvr.HttpServerClient{
		Mgr: mgr,
	}
	httpsvr.InitRouters(ws, httpServerClient)

	router := server.GetRouter()
	webContainer := server.GetWebContainer()
	router.Handle("/ingresscontroller/{sub_path:.*}", webContainer)
	if err := server.ListenAndServeMux(op.Conf.VerifyClientTLS); err != nil {
		return fmt.Errorf("http ListenAndServe error %s", err.Error())
	}
	return nil

}
