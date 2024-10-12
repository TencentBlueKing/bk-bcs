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
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpserver"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/ipv6server"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/emicklei/go-restful"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/bcs-ingress-inspector/option"
	portbindingctrl "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/bcs-ingress-inspector/portbindingcontroller"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/httpsvr"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/portbindingcontroller"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)
	_ = networkextensionv1.AddToScheme(scheme)
}

// nolint  funlen
func main() {
	opts := &option.ControllerOption{}
	opts.BindFromCommandLine()

	blog.InitLogs(opts.LogConfig)
	defer blog.CloseLogs()

	ctrl.SetLogger(zap.New(zap.UseDevMode(false)))

	opts.SetFromEnv()
	ctx, cancel := context.WithCancel(context.Background())
	go StartSignalHandler(cancel, 3)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: "0", // "0"表示禁用默认的Metric Service， 需要使用自己的实现支持IPV6
		LeaderElection:     false,
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

	nodeBindCache := portbindingcontroller.NewNodePortBindingCache(mgr.GetClient())
	portBindingReconciler := portbindingctrl.NewPortBindingReconciler(
		ctx, mgr.GetClient(), mgr.GetEventRecorderFor("bcs-ingress-controller"), nodeBindCache)
	if err = portBindingReconciler.SetupWithManager(mgr); err != nil {
		blog.Errorf("unable to create port binding reconciler, err %s", err.Error())
		os.Exit(1)
	}

	err = initHttpServer(opts, mgr, nodeBindCache)
	if err != nil {
		blog.Errorf("init http server failed: %v", err.Error())
		os.Exit(1)
	}
	blog.Infof("starting http server")

	if err = mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		blog.Errorf("problem running manager, err %s", err.Error())
		os.Exit(1)
	}
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
// httpServer提供
// 1. 集群内Ingress/PortPool/PortBinding/Listener等信息的查询
// 2. 维护节点信息，提供接口给Pod获取所在节点的信息
func initHttpServer(op *option.ControllerOption, mgr manager.Manager,
	nodeBindCache *portbindingcontroller.NodePortBindingCache) error {
	server := httpserver.NewHttpServer(op.HttpServerPort, op.Address, "")
	if op.Conf.ServCert.IsSSL {
		server.SetSsl(op.Conf.ServCert.CAFile, op.Conf.ServCert.CertFile, op.Conf.ServCert.KeyFile,
			op.Conf.ServCert.CertPasswd)
	}

	// server.SetInsecureServer(op.Conf.InsecureAddress, op.Conf.InsecurePort)
	server.SetInsecureServer(op.Address, op.HttpServerPort)
	ws := server.NewWebService("/ingresscontroller", []restful.FilterFunction{globalLoggingFilter})
	httpServerClient := &httpsvr.HttpServerClient{
		Mgr: mgr,
		// NodeCache: nodeCache,
		// Ops:               op,
		NodePortBindCache: nodeBindCache,
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

// StartSignalHandler trap system signal for exit
func StartSignalHandler(stop context.CancelFunc, gracefulExit int) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-ch
	blog.Infof("server received stop signal.")
	// trap system signal, stop
	stop()
	tick := time.NewTicker(time.Second * time.Duration(gracefulExit))
	select {
	case <-ch:
		// double kill, just terminate immediately
		os.Exit(0)
	case <-tick.C:
		// timeout
		return
	}
}

// 全局日志过滤器
func globalLoggingFilter(req *restful.Request, resp *restful.Response, chain *restful.FilterChain) {
	// 打印请求信息
	blog.Infof("Received request: %s %s from %s, contentLength:%d", req.Request.Method, req.Request.URL,
		req.Request.RemoteAddr, req.Request.ContentLength)

	// 继续处理请求
	chain.ProcessFilter(req, resp)
}
