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
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

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
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloudnode"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloudnode/native"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/conflicthandler"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/eventer"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/generator"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/httpsvr"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/ingresscache"
	internalmetric "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/nodecache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/portpoolcache"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/webhookserver"
	listenerctrl "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/listenercontroller"
	namespacectrl "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/namespacecontroller"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/nodecontroller"
	portbindingctrl "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/portbindingcontroller"
	portpoolctrl "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/portpoolcontroller"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
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
	opts.BindFromCommandLine()

	blog.InitLogs(opts.LogConfig)
	defer blog.CloseLogs()

	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))

	opts.SetFromEnv()

	// init port pool cache
	portPoolCache := portpoolcache.NewCache()
	go portPoolCache.Start()

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      "0", // "0"表示禁用默认的Metric Service， 需要使用自己的实现支持IPV6
		LeaderElection:          true,
		LeaderElectionID:        "33fb49e.cloudlbconroller.bkbcs.tencent.com",
		LeaderElectionNamespace: opts.ElectionNamespace,
		NewClient: func(cache cache.Cache, config *rest.Config, options client.Options) (client.Client, error) {
			// 调高对K8S client的QPS限制，优化大批量监听器时的处理效率
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

	// init event watcher
	k8sClient, err := initInClusterClient()
	if err != nil {
		blog.Fatalf("init in-cluster client failed: %v", err)
	}
	eventWatcher := eventer.NewKubeEventer(k8sClient)
	if err = eventWatcher.Init(); err != nil {
		blog.Fatalf("init event watcher failed: %v", err)
	}
	go eventWatcher.Start(context.Background())

	validater, lbClient, nodeClient := initClient(opts, mgr.GetClient(), eventWatcher)

	if len(opts.Region) == 0 {
		blog.Errorf("region cannot be empty")
		os.Exit(1)
	}

	// 用于异步处理listener删除
	listenerHelper := listenerctrl.NewListenerHelper(mgr.GetClient())
	// 缓存Ingress使用到的LB信息 （会在Check中定时刷新）
	lbIDCache := gocache.New(time.Duration(opts.LBCacheExpiration)*time.Minute, 120*time.Minute)
	lbNameCache := gocache.New(time.Duration(opts.LBCacheExpiration)*time.Minute, 120*time.Minute)
	ingressConverter, err := generator.NewIngressConverter(&generator.IngressConverterOpt{
		DefaultRegion:     opts.Region,
		IsTCPUDPPortReuse: opts.IsTCPUDPPortReuse,
		Cloud:             opts.Cloud,
	}, mgr.GetClient(), validater, lbClient, listenerHelper, lbIDCache, lbNameCache)
	if err != nil {
		blog.Errorf("create ingress converter failed, err %s", err.Error())
		os.Exit(1)
	}

	nodeCache := nodecache.NewNodeCache(mgr.GetClient(), nodeClient, opts.NodeExternalWorkerEnable,
		opts.NodeExternalIPConfigmap, opts.PodNamespace)
	// ingressCache 缓存ingress相关的service/workload信息，避免量大时影响ingress调谐时间
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

	nodeBindCache := portbindingctrl.NewNodePortBindingCache()
	portBindingReconciler := portbindingctrl.NewPortBindingReconciler(
		context.Background(), opts.PortBindingCheckInterval, mgr.GetClient(), portPoolCache,
		mgr.GetEventRecorderFor("bcs-ingress-controller"), opts.NodePortBindingNs, nodeBindCache)
	if err = portBindingReconciler.SetupWithManager(mgr); err != nil {
		blog.Errorf("unable to create port binding reconciler, err %s", err.Error())
		os.Exit(1)
	}

	namespaceReconciler := namespacectrl.NewNamespaceReconciler(context.Background(), mgr.GetClient(), nodeBindCache)
	if err = namespaceReconciler.SetupWithManager(mgr); err != nil {
		blog.Errorf("unable to create namespace reconciler, err %v", err)
		os.Exit(1)
	}

	if opts.NodeInfoExporterOpen {
		nodeReconciler := nodecontroller.NewNodeReconciler(context.Background(), mgr.GetClient(),
			mgr.GetEventRecorderFor("bcs-ingress-controller"), opts, nodeCache, nodeClient)
		if err = nodeReconciler.SetupWithManager(mgr); err != nil {
			blog.Errorf("unable to create node reconciler, err %s", err.Error())
			os.Exit(1)
		}
	}

	// conflictHandler 避免不同Ingress/PortPool之间出现端口冲突
	conflictHandler := conflicthandler.NewConflictHandler(opts.ConflictCheckOpen, opts.IsTCPUDPPortReuse, opts.Region,
		mgr.GetClient(), ingressConverter, mgr.GetEventRecorderFor("bcs-ingress-controller"))
	// init webhook server
	webhookServerOpts := &webhookserver.ServerOption{
		Addrs:          opts.PodIPs,
		Port:           opts.Port,
		ServerCertFile: opts.ServerCertFile,
		ServerKeyFile:  opts.ServerKeyFile,
	}
	webhookServer, err := webhookserver.NewHookServer(webhookServerOpts, mgr.GetClient(), lbClient, portPoolCache,
		eventWatcher, validater, ingressConverter, conflictHandler, opts.NodePortBindingNs,
		mgr.GetEventRecorderFor("bcs-ingress-controller"))
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

	err = initHttpServer(opts, mgr, nodeCache, nodeBindCache)
	if err != nil {
		blog.Errorf("init http server failed: %v", err.Error())
		os.Exit(1)
	}
	blog.Infof("starting http server")

	// 定时执行检查
	checkRunner := check.NewCheckRunner(context.Background())
	checkRunner.
		Register(check.NewPortBindChecker(mgr.GetClient(), mgr.GetEventRecorderFor("bcs-ingress-controller"))).
		Register(check.NewListenerChecker(mgr.GetClient(), listenerHelper)).
		Register(check.NewIngressChecker(mgr.GetClient(), lbClient, lbIDCache, lbNameCache, opts.LBCacheExpiration)).
		Start()
	blog.Infof("starting check runner")

	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		blog.Errorf("problem running manager, err %s", err.Error())
		os.Exit(1)
	}
}

// initInClusterClient return client from clsuter config
func initInClusterClient() (*kubernetes.Clientset, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrapf(err, "get in-cluster config failed")
	}
	cli, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "create in-cluster client failed")
	}
	return cli, nil
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

	internalmetric.ControllerInfo.WithLabelValues(op.ImageTag, op.Cloud).Set(1)
}

// initHttpServer init ingress controller http server
// httpServer提供
// 1. 集群内Ingress/PortPool/PortBinding/Listener等信息的查询
// 2. 维护节点信息，提供接口给Pod获取所在节点的信息
func initHttpServer(op *option.ControllerOption, mgr manager.Manager, nodeCache *nodecache.NodeCache,
	nodeBindCache *portbindingctrl.NodePortBindingCache) error {
	server := httpserver.NewHttpServer(op.HttpServerPort, op.Address, "")
	if op.Conf.ServCert.IsSSL {
		server.SetSsl(op.Conf.ServCert.CAFile, op.Conf.ServCert.CertFile, op.Conf.ServCert.KeyFile,
			op.Conf.ServCert.CertPasswd)
	}

	// server.SetInsecureServer(op.Conf.InsecureAddress, op.Conf.InsecurePort)
	server.SetInsecureServer(op.Address, op.HttpServerPort)
	ws := server.NewWebService("/ingresscontroller", nil)
	httpServerClient := &httpsvr.HttpServerClient{
		Mgr:               mgr,
		NodeCache:         nodeCache,
		Ops:               op,
		NodePortBindCache: nodeBindCache,
	}
	// aga supporter can only be init when use
	if op.Cloud == constant.CloudAWS {
		httpServerClient.AgaSupporter = aws.NewAgaSupporter()
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

// initClient 根据使用云厂商的不同，返回对应云厂商的实现
func initClient(opts *option.ControllerOption, cli client.Client, eventWatcher eventer.WatchEventInterface) (cloud.
	Validater, cloud.LoadBalance, cloudnode.NodeClient) {
	var validater cloud.Validater
	var lbClient cloud.LoadBalance
	var nodeClient cloudnode.NodeClient
	var err error
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
			// NameSpacedLB在处理监听器时，会使用对应命名空间下的Secret作为云密钥
			lbClient = namespacedlb.NewNamespacedLB(cli, eventWatcher,
				tencentcloud.NewClbWithSecret)
		}
		nodeClient = native.NewNativeNodeClient()

	case constant.CloudAWS:
		validater = aws.NewELbValidater()
		if !opts.IsNamespaceScope {
			lbClient, err = aws.NewElb()
			if err != nil {
				blog.Errorf("init cloud failed, err %s", err.Error())
				os.Exit(1)
			}
		} else {
			lbClient = namespacedlb.NewNamespacedLB(cli, eventWatcher, aws.NewElbWithSecret)
		}
		nodeClient = native.NewNativeNodeClient()

	case constant.CloudGCP:
		validater = gcp.NewGclbValidater()
		if !opts.IsNamespaceScope {
			lbClient, err = gcp.NewGclb(cli, eventWatcher)
			if err != nil {
				blog.Errorf("init cloud failed, err %s", err.Error())
				os.Exit(1)
			}
		} else {
			lbClient = namespacedlb.NewNamespacedLB(cli, eventWatcher,
				gcp.NewGclbWithSecret)
		}
		nodeClient = native.NewNativeNodeClient()

	case constant.CloudAzure:
		validater = azure.NewAlbValidater()
		if !opts.IsNamespaceScope {
			lbClient, err = azure.NewAlb()
			if err != nil {
				blog.Errorf("init cloud failed, err %s", err.Error())
				os.Exit(1)
			}
		} else {
			lbClient = namespacedlb.NewNamespacedLB(cli, eventWatcher, azure.NewAlbWithSecret)
		}
		nodeClient = native.NewNativeNodeClient()

	default:
		blog.Errorf("unknown cloud type '%s'", opts.Cloud)
		os.Exit(1)
	}
	return validater, lbClient, nodeClient
}
