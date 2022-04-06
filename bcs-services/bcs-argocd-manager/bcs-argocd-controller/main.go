/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/conf"
	instancecontroller "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-controller/controllers/argocdinstance"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/bcs-argocd-controller/options"
	clientset "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/clientset/versioned"
	informers "github.com/Tencent/bk-bcs/bcs-services/bcs-argocd-manager/pkg/client/informers/externalversions"

	microCfg "go-micro.dev/v4/config"
	microFile "go-micro.dev/v4/config/source/file"
	microFlg "go-micro.dev/v4/config/source/flag"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// controller config
	flag.Bool("debug", false, "run in debug mode")

	// log config
	flag.String("bcslog_dir", "./logs", "If non-empty, write log files in this directory")
	flag.Uint64("bcslog_maxsize", 500, "Max size (MB) per log file.")
	flag.Int("bcslog_maxnum", 10, "Max num of log file. The oldest will be removed if there is a extra file created.")
	flag.Bool("bcslog_tostderr", false, "log to standard error instead of files")
	flag.Bool("bcslog_alsotostderr", true, "log to standard error as well as files")
	flag.Int("bcslog_v", 0, "log level for V logs")
	flag.String("bcslog_stderrthreshold", "2", "logs at or above this threshold go to stderr")
	flag.String("bcslog_vmodule", "", "comma-separated list of pattern=N settings for file-filtered logging")
	flag.String("bcslog_backtraceat", "", "when logging hits line file:N, emit a stack trace")

	// plugin config
	flag.String("plugin_serverimage_registry", "", "plugin sidecar server image registry")
	flag.String("plugin_serverimage_repository", "", "plugin sidecar server image repository")
	flag.String("plugin_serverimage_pullpolicy", "", "plugin sidecar server image pullpolicy")
	flag.String("plugin_serverimage_tag", "", "plugin sidecar server image tag")
	flag.String("plugin_clientimage_registry", "", "plugin sidecar client image registry")
	flag.String("plugin_clientimage_repository", "", "plugin sidecar client image repository")
	flag.String("plugin_clientimage_pullpolicy", "", "plugin sidecar client image pullpolicy")
	flag.String("plugin_clientimage_tag", "", "plugin sidecar client image tag")

	// kubeconfig
	flag.String("masterurl", "", "url of the k8s master")
	flag.String("kubeconfig", "", "kubeconfig path")

	// config file path
	flag.String("conf", "", "config file path")
	flag.Parse()

	opt := &options.ArgocdControllerOptions{}
	config, err := microCfg.NewConfig()
	if err != nil {
		blog.Fatalf("create config failed, %s", err.Error())
	}

	if err = config.Load(
		microFlg.NewSource(
			microFlg.IncludeUnset(true),
		),
	); err != nil {
		blog.Fatalf("load config from flag failed, %s", err.Error())
	}

	if len(config.Get("conf").String("")) > 0 {
		err = config.Load(microFile.NewSource(microFile.WithPath(config.Get("conf").String(""))))
		if err != nil {
			blog.Fatalf("load config from file failed, err %s", err.Error())
		}
	}

	if err = config.Scan(opt); err != nil {
		blog.Fatalf("scan config failed, %s", err.Error())
	}

	blog.InitLogs(conf.LogConfig{
		LogDir:          opt.BcsLog.LogDir,
		LogMaxSize:      opt.BcsLog.LogMaxSize,
		LogMaxNum:       opt.BcsLog.LogMaxNum,
		ToStdErr:        opt.BcsLog.ToStdErr,
		AlsoToStdErr:    opt.BcsLog.AlsoToStdErr,
		Verbosity:       opt.BcsLog.Verbosity,
		StdErrThreshold: opt.BcsLog.StdErrThreshold,
		VModule:         opt.BcsLog.VModule,
		TraceLocation:   opt.BcsLog.TraceLocation,
	})
	stopCh := make(chan struct{})

	blog.Infof("server image: %s", opt.Plugin.ServerImage.Repository)
	kubeConfig, err := clientcmd.BuildConfigFromFlags(opt.MasterURL, opt.KubeConfig)
	if err != nil {
		blog.Fatalf("build kube config failed, err %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		blog.Fatalf("create kubernetes client failed, err %s", err.Error())
	}
	tkexClient, err := clientset.NewForConfig(kubeConfig)
	if err != nil {
		blog.Fatalf("create tkex client failed, err %s", err.Error())
	}
	kubeInformerFactory := kubeinformers.NewSharedInformerFactory(kubeClient, time.Second*30)
	tkexInformerFactory := informers.NewSharedInformerFactory(tkexClient, time.Second*30)

	// init helm action client
	flags := genericclioptions.NewConfigFlags(false)
	if opt.KubeConfig != "" {
		flags.KubeConfig = &opt.KubeConfig
	}
	actionConfig := new(action.Configuration)
	clientGetter := genericclioptions.NewConfigFlags(false)
	clientGetter.KubeConfig = flags.KubeConfig
	if err := actionConfig.Init(clientGetter, "", "", blog.Info); err != nil {
		blog.Fatalf("init helm action config failed: %v", err)

	}
	controller := instancecontroller.NewController(kubeConfig, opt.Plugin, kubeClient, tkexClient,
		tkexInformerFactory.Tkex().V1alpha1().ArgocdInstances(),
		kubeInformerFactory.Core().V1().Namespaces(),
		kubeInformerFactory.Core().V1().Services())

	kubeInformerFactory.Start(stopCh)
	tkexInformerFactory.Start(stopCh)

	if err = controller.Run(1, stopCh); err != nil {
		blog.Fatalf("Error running controller: %s", err.Error())
	}
}
